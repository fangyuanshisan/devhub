package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	pluginregistry "devhub-gin-backend/internal/plugins"

	"devhub-gin-backend/internal/domain"
)

// DryRunPluginPackage scans a local plugin package directory, validates its files,
// loads manifest.json, and reuses existing manifest validation/dry-run logic.
//
// It never installs the plugin, never writes plugin/config/migration tables, and never executes code/SQL.
func (s *Service) DryRunPluginPackage(inputPath string) (domain.PluginPackageDryRunResult, error) {
	abs, clean, err := pluginregistry.NormalizePluginPackagePath(inputPath)
	if err != nil {
		return domain.PluginPackageDryRunResult{}, err
	}

	scan, err := pluginregistry.ScanPluginPackage(abs)
	if err != nil {
		return domain.PluginPackageDryRunResult{}, err
	}

	manifestPath := filepath.Join(abs, "manifest.json")
	readmePath := filepath.Join(abs, "README.md")
	configExamplePath := filepath.Join(abs, "config.example.json")
	checksumPath := filepath.Join(abs, "checksums.json")
	signaturePath := filepath.Join(abs, "signature.json")
	publisherPath := filepath.Join(abs, "publisher.json")

	info := domain.PluginPackageInfo{
		Path:               clean,
		DirName:            filepath.Base(abs),
		ManifestFound:      fileExists(manifestPath),
		ReadmeFound:        fileExists(readmePath),
		ConfigExampleFound: fileExists(configExamplePath),
		ChecksumFound:      fileExists(checksumPath),
		SignatureFound:     fileExists(signaturePath),
		PublisherFound:     fileExists(publisherPath),
	}

	// If manifest is missing, return an explicit error code (blocked).
	if !info.ManifestFound {
		return domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_package_manifest_missing", "插件包缺少 manifest.json").
			WithStatus(400).
			WithDetail("path", clean).
			WithSuggestion("请在插件包根目录放置 manifest.json（唯一主声明文件）。")
	}

	manifestRaw, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		return domain.PluginPackageDryRunResult{}, fmt.Errorf("读取 manifest.json 失败：%w", readErr)
	}

	manifest, checksum, decodeErr := pluginregistry.DecodePluginManifestJSON(manifestRaw)
	if decodeErr != nil {
		return domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_package_manifest_invalid", "manifest.json 不是合法的插件声明").
			WithStatus(400).
			WithDetail("path", clean).
			WithDetail("reason", strings.TrimSpace(decodeErr.Error())).
			WithSuggestion("请修复 manifest.json 后重试。")
	}
	_ = checksum
	info.Code = manifest.Code
	info.Name = manifest.Name
	info.Version = manifest.Version

	status := "ok"
	blockedCode := ""
	blockedReasons := []string{}
	warnings := []string{}
	errorsList := []string{}

	// Directory name suggestion.
	if info.DirName != "" && info.Code != "" && info.DirName != info.Code {
		warnings = append(warnings, fmt.Sprintf("目录名建议与 manifest.code 保持一致：dir=%s code=%s", info.DirName, info.Code))
		status = "warning"
	}

	// Scan-derived status.
	if len(scan.DangerousFiles) > 0 || len(scan.Errors) > 0 {
		status = "blocked"
		blockedCode = scanBlockedCode(scan)
		if blockedCode != "" {
			blockedReasons = append(blockedReasons, blockedCode)
		}
		errorsList = append(errorsList, scan.Errors...)
	}
	if len(scan.UnknownFiles) > 0 || len(scan.Warnings) > 0 {
		if status != "blocked" {
			status = "warning"
		}
		if len(scan.UnknownFiles) > 0 {
			warnings = append(warnings, fmt.Sprintf("发现 %d 个未知文件（unknown_files）", len(scan.UnknownFiles)))
		}
		warnings = append(warnings, scan.Warnings...)
	}
	if hasDeprecatedRootSchemaSQL(scan) {
		if status != "blocked" {
			status = "warning"
		}
		warnings = append(warnings, "根目录 001_schema.sql 已废弃，请迁移到 migrations/001_init.sql。DevHub 插件包标准迁移入口仅为 migrations/，dry-run 和安装流程不会执行根目录 001_schema.sql。")
	}
	migrationPlan, planWarnings := buildPackageMigrationPlan(abs, scan)
	if len(planWarnings) > 0 {
		if status != "blocked" {
			status = "warning"
		}
		warnings = append(warnings, planWarnings...)
	}

	// checksums.json verification (optional, but mismatch/invalid blocks dry-run).
	checksumResult, chkErr := pluginregistry.VerifyPluginPackageChecksums(abs, scan)
	if chkErr != nil {
		if apiErr, ok := chkErr.(*domain.APIError); ok {
			status = "blocked"
			if blockedCode == "" {
				blockedCode = apiErr.Code
			}
			blockedReasons = append(blockedReasons, apiErr.Code)
			errorsList = append(errorsList, fmt.Sprintf("[%s] %s", apiErr.Code, apiErr.Message))
			if apiErr.Suggestion != "" {
				warnings = append(warnings, fmt.Sprintf("建议：%s", apiErr.Suggestion))
			}
			if checksumResult.Status == "" {
				checksumResult.Status = "failed"
			}
			if checksumResult.Algorithm == "" {
				checksumResult.Algorithm = "sha256"
			}
			checksumResult.Errors = append(checksumResult.Errors, apiErr.Message)
		} else {
			return domain.PluginPackageDryRunResult{}, chkErr
		}
	} else {
		if len(checksumResult.Warnings) > 0 {
			warnings = append(warnings, checksumResult.Warnings...)
			if status == "ok" {
				status = "warning"
			}
		}
	}

	// signature/publisher verification (optional, but invalid/unsupported/failed blocks dry-run).
	signatureResult, sigErr := pluginregistry.VerifyPluginPackageSignatureWithTrusted(abs, scan, checksumResult, s.trustedPublishersConfig())
	if sigErr != nil {
		if apiErr, ok := sigErr.(*domain.APIError); ok && apiErr != nil {
			status = "blocked"
			if blockedCode == "" {
				blockedCode = apiErr.Code
			}
			blockedReasons = append(blockedReasons, apiErr.Code)
			errorsList = append(errorsList, fmt.Sprintf("[%s] %s", apiErr.Code, apiErr.Message))
			if apiErr.Suggestion != "" {
				warnings = append(warnings, fmt.Sprintf("建议：%s", apiErr.Suggestion))
			}
		} else {
			return domain.PluginPackageDryRunResult{}, sigErr
		}
	}

	// config.example.json sanity (optional, warning only).
	if info.ConfigExampleFound {
		raw, cfgErr := os.ReadFile(configExamplePath)
		if cfgErr != nil {
			warnings = append(warnings, fmt.Sprintf("读取 config.example.json 失败：%s", cfgErr.Error()))
			if status == "ok" {
				status = "warning"
			}
		} else {
			var v any
			if err := json.Unmarshal(raw, &v); err != nil {
				warnings = append(warnings, fmt.Sprintf("config.example.json 不是合法 JSON：%s", err.Error()))
				if status == "ok" {
					status = "warning"
				}
			}
		}
	} else {
		warnings = append(warnings, "建议提供 config.example.json（示例配置，不应包含真实 secret）")
		if status == "ok" {
			status = "warning"
		}
	}

	// README missing is a warning.
	if !info.ReadmeFound {
		warnings = append(warnings, "建议提供 README.md（插件说明）")
		if status == "ok" {
			status = "warning"
		}
	}

	existing := make([]domain.Plugin, 0, len(s.repo.Plugins()))
	for _, item := range s.repo.Plugins() {
		if strings.TrimSpace(item.Code) == strings.TrimSpace(manifest.Code) {
			continue
		}
		existing = append(existing, item)
	}
	validation := pluginregistry.ValidatePluginManifest(manifest, existing, currentCoreVersion())
	validation.Checksum = checksum
	manifestValidation := domain.PluginPackageManifestValidation{
		Valid:    validation.Valid,
		Errors:   append([]string(nil), validation.Errors...),
		Warnings: append([]string(nil), validation.Warnings...),
	}
	if !manifestValidation.Valid {
		status = "blocked"
		blockedCode = "plugin_package_manifest_invalid"
		blockedReasons = append(blockedReasons, blockedCode)
		errorsList = append(errorsList, "manifest 校验失败")
	}

	// In v1.4, "install dry-run" shares the same payload as manifest validation result.
	installDryRun := validation

	risk := pluginregistry.BuildPluginPackageRiskReport(info, scan, checksumResult, signatureResult, manifestValidation, installDryRun, blockedCode)
	if risk.Level == "blocked" {
		status = "blocked"
	}
	if (risk.Level == "medium" || risk.Level == "high") && status == "ok" {
		status = "warning"
	}

	generatedAt := timeNow()
	expiresAt := generatedAt.Add(pluginPackageInstallDryRunTTL)
	result := domain.PluginPackageDryRunResult{
		Package:            info,
		FileScan:           scan,
		Checksum:           checksumResult,
		Signature:          signatureResult,
		ManifestValidation: manifestValidation,
		InstallDryRun:      installDryRun,
		MigrationPlan:      migrationPlan,
		RiskReport:         risk,
		Status:             status,
		BlockedCode:        blockedCode,
		BlockedReasons:     uniqueStrings(blockedReasons),
		Warnings:           uniqueStrings(warnings),
		Errors:             uniqueStrings(errorsList),
		GeneratedAt:        generatedAt.Format("2006-01-02 15:04:05"),
		ExpiresAt:          expiresAt.Format("2006-01-02 15:04:05"),
	}
	result.DryRunID = s.signPluginPackageInstallDryRun(result, generatedAt, expiresAt)
	return result, nil
}

const pluginPackageInstallDryRunTTL = 30 * time.Minute

type pluginPackageInstallDryRunToken struct {
	Path              string   `json:"path"`
	PluginCode        string   `json:"plugin_code"`
	Version           string   `json:"version"`
	ManifestChecksum  string   `json:"manifest_checksum"`
	ChecksumStatus    string   `json:"checksum_status"`
	MigrationPlanHash string   `json:"migration_plan_hash"`
	Status            string   `json:"status"`
	RiskLevel         string   `json:"risk_level"`
	GeneratedAtUnix   int64    `json:"generated_at_unix"`
	ExpiresAtUnix     int64    `json:"expires_at_unix"`
	Warnings          []string `json:"warnings,omitempty"`
}

func (s *Service) signPluginPackageInstallDryRun(dry domain.PluginPackageDryRunResult, generatedAt, expiresAt time.Time) string {
	if s == nil || strings.TrimSpace(s.packageDryRunSecret) == "" {
		return ""
	}
	token := pluginPackageInstallDryRunToken{
		Path:              strings.TrimSpace(dry.Package.Path),
		PluginCode:        strings.TrimSpace(dry.Package.Code),
		Version:           strings.TrimSpace(dry.Package.Version),
		ManifestChecksum:  strings.TrimSpace(dry.InstallDryRun.Checksum),
		ChecksumStatus:    strings.TrimSpace(dry.Checksum.Status),
		MigrationPlanHash: hashPluginPackageMigrationPlan(dry.MigrationPlan),
		Status:            strings.TrimSpace(dry.Status),
		RiskLevel:         strings.TrimSpace(dry.RiskReport.Level),
		GeneratedAtUnix:   generatedAt.Unix(),
		ExpiresAtUnix:     expiresAt.Unix(),
		Warnings:          append([]string(nil), dry.Warnings...),
	}
	raw, err := json.Marshal(token)
	if err != nil {
		return ""
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(s.packageDryRunSecret))
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig
}

func (s *Service) verifyPluginPackageInstallDryRunID(dry domain.PluginPackageDryRunResult, dryRunID string) error {
	dryRunID = strings.TrimSpace(dryRunID)
	if dryRunID == "" {
		return domain.NewPluginError("plugin_package_install_dry_run_required", "请先执行安装 dry-run").
			WithStatus(400).
			WithDetail("path", dry.Package.Path).
			WithSuggestion("安装只能基于本地仓库包的当前 dry-run 计划执行，请先点击“执行安装 dry-run”。")
	}
	if strings.TrimSpace(s.packageDryRunSecret) == "" {
		return domain.NewPluginError("plugin_package_install_dry_run_invalid", "安装 dry-run 凭证不可用").
			WithStatus(400).
			WithSuggestion("请重新执行安装 dry-run 后再安装。")
	}
	parts := strings.Split(dryRunID, ".")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return domain.NewPluginError("plugin_package_install_dry_run_invalid", "dry-run 计划无效，请重新执行").
			WithStatus(400).
			WithSuggestion("请重新执行安装 dry-run，确认当前计划后再安装。")
	}
	mac := hmac.New(sha256.New, []byte(s.packageDryRunSecret))
	_, _ = mac.Write([]byte(parts[0]))
	expected := mac.Sum(nil)
	actual, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(actual, expected) {
		return domain.NewPluginError("plugin_package_install_dry_run_invalid", "dry-run 计划签名无效，请重新执行").
			WithStatus(400).
			WithSuggestion("请重新执行安装 dry-run，确认当前计划后再安装。")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return domain.NewPluginError("plugin_package_install_dry_run_invalid", "dry-run 计划无法解析，请重新执行").
			WithStatus(400)
	}
	var token pluginPackageInstallDryRunToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return domain.NewPluginError("plugin_package_install_dry_run_invalid", "dry-run 计划无法解析，请重新执行").
			WithStatus(400)
	}
	if token.ExpiresAtUnix <= 0 || timeNow().Unix() > token.ExpiresAtUnix {
		return domain.NewPluginError("plugin_package_install_dry_run_expired", "dry-run 已过期，请重新执行").
			WithStatus(400).
			WithDetail("path", dry.Package.Path).
			WithSuggestion("安装前请基于当前本地仓库包重新执行 dry-run。")
	}
	current := pluginPackageInstallDryRunToken{
		Path:              strings.TrimSpace(dry.Package.Path),
		PluginCode:        strings.TrimSpace(dry.Package.Code),
		Version:           strings.TrimSpace(dry.Package.Version),
		ManifestChecksum:  strings.TrimSpace(dry.InstallDryRun.Checksum),
		ChecksumStatus:    strings.TrimSpace(dry.Checksum.Status),
		MigrationPlanHash: hashPluginPackageMigrationPlan(dry.MigrationPlan),
		Status:            strings.TrimSpace(dry.Status),
		RiskLevel:         strings.TrimSpace(dry.RiskReport.Level),
	}
	if token.Path != current.Path || token.PluginCode != current.PluginCode || token.Version != current.Version || token.ManifestChecksum != current.ManifestChecksum || token.ChecksumStatus != current.ChecksumStatus || token.MigrationPlanHash != current.MigrationPlanHash || token.Status != current.Status || token.RiskLevel != current.RiskLevel {
		return domain.NewPluginError("plugin_package_install_dry_run_mismatch", "dry-run 计划与当前插件包不一致，请重新执行").
			WithStatus(400).
			WithDetail("path", dry.Package.Path).
			WithDetail("plugin_code", dry.Package.Code).
			WithDetail("version", dry.Package.Version).
			WithSuggestion("插件包内容、校验和、manifest 或 migration plan 已变化，请重新执行安装 dry-run。")
	}
	return nil
}

func hashPluginPackageMigrationPlan(plan []domain.PluginPackageMigrationPlanItem) string {
	type stableItem struct {
		Path        string `json:"path"`
		Name        string `json:"name"`
		SHA256      string `json:"sha256,omitempty"`
		Source      string `json:"source"`
		WillExecute bool   `json:"will_execute"`
	}
	items := make([]stableItem, 0, len(plan))
	for _, item := range plan {
		items = append(items, stableItem{
			Path:        strings.TrimSpace(item.Path),
			Name:        strings.TrimSpace(item.Name),
			SHA256:      strings.TrimSpace(item.SHA256),
			Source:      strings.TrimSpace(item.Source),
			WillExecute: item.WillExecute,
		})
	}
	raw, _ := json.Marshal(items)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func (s *Service) signPluginPackageCleanupToken(scope string, req domain.PluginPackageCleanupRequest, items []domain.PluginPackageCleanupItem) string {
	if s == nil || strings.TrimSpace(s.packageDryRunSecret) == "" {
		return ""
	}
	payload := map[string]any{
		"scope":                        strings.TrimSpace(scope),
		"cleanup_scope":                normalizePluginPackageCleanupScope(req.Scope),
		"statuses":                     normalizedCleanupStatuses(req.Statuses),
		"prefixes":                     normalizedCleanupPrefixes(req.Prefixes),
		"older_than_days":              req.OlderThanDays,
		"include_blocked":              req.IncludeBlocked,
		"include_invalid":              req.IncludeInvalid,
		"include_promoted_uninstalled": req.IncludePromotedUninstalled,
		"include_warning_uninstalled":  req.IncludeWarningUninstalled,
		"include_dry_run_failed":       req.IncludeDryRunFailed,
		"items":                        cleanupTokenItems(items),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	body := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(s.packageDryRunSecret))
	_, _ = mac.Write([]byte(body))
	return body + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) verifyPluginPackageCleanupToken(scope string, req domain.PluginPackageCleanupRequest, items []domain.PluginPackageCleanupItem) error {
	token := strings.TrimSpace(req.ConfirmToken)
	if token == "" {
		return domain.NewPluginError("plugin_package_cleanup_confirm_required", "批量清理必须先 dry-run 并携带确认 token").
			WithStatus(400).
			WithSuggestion("请先提交 dry_run=true 获取 confirm_token，再带 confirm_token 执行清理。")
	}
	expected := s.signPluginPackageCleanupToken(scope, domain.PluginPackageCleanupRequest{
		Scope:                      req.Scope,
		Statuses:                   req.Statuses,
		Prefixes:                   req.Prefixes,
		OlderThanDays:              req.OlderThanDays,
		IncludeBlocked:             req.IncludeBlocked,
		IncludeInvalid:             req.IncludeInvalid,
		IncludePromotedUninstalled: req.IncludePromotedUninstalled,
		IncludeWarningUninstalled:  req.IncludeWarningUninstalled,
		IncludeDryRunFailed:        req.IncludeDryRunFailed,
	}, items)
	if expected == "" || token != expected {
		if !s.cleanupTokenAllowsCurrentItems(scope, req, token, items) {
			return domain.NewPluginError("plugin_package_cleanup_confirm_invalid", "cleanup 确认 token 无效或已不匹配当前清理计划").
				WithStatus(400).
				WithSuggestion("请重新 dry-run，确认最新清理列表后再执行。")
		}
	}
	return nil
}

func (s *Service) cleanupTokenAllowsCurrentItems(scope string, req domain.PluginPackageCleanupRequest, token string, items []domain.PluginPackageCleanupItem) bool {
	if s == nil || strings.TrimSpace(s.packageDryRunSecret) == "" {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.packageDryRunSecret))
	_, _ = mac.Write([]byte(parts[0]))
	got, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(got, mac.Sum(nil)) {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	var payload struct {
		Scope                      string   `json:"scope"`
		CleanupScope               string   `json:"cleanup_scope"`
		Statuses                   []string `json:"statuses"`
		Prefixes                   []string `json:"prefixes"`
		OlderThanDays              int      `json:"older_than_days"`
		IncludeBlocked             bool     `json:"include_blocked"`
		IncludeInvalid             bool     `json:"include_invalid"`
		IncludePromotedUninstalled bool     `json:"include_promoted_uninstalled"`
		IncludeWarningUninstalled  bool     `json:"include_warning_uninstalled"`
		IncludeDryRunFailed        bool     `json:"include_dry_run_failed"`
		Items                      []string `json:"items"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false
	}
	if payload.Scope != strings.TrimSpace(scope) ||
		payload.CleanupScope != normalizePluginPackageCleanupScope(req.Scope) ||
		payload.OlderThanDays != req.OlderThanDays ||
		payload.IncludeBlocked != req.IncludeBlocked ||
		payload.IncludeInvalid != req.IncludeInvalid ||
		payload.IncludePromotedUninstalled != req.IncludePromotedUninstalled ||
		payload.IncludeWarningUninstalled != req.IncludeWarningUninstalled ||
		payload.IncludeDryRunFailed != req.IncludeDryRunFailed ||
		!stringSlicesEqual(payload.Statuses, normalizedCleanupStatuses(req.Statuses)) ||
		!stringSlicesEqual(payload.Prefixes, normalizedCleanupPrefixes(req.Prefixes)) {
		return false
	}
	allowed := map[string]bool{}
	for _, item := range payload.Items {
		allowed[item] = true
	}
	for _, item := range cleanupTokenItems(items) {
		if !allowed[item] {
			return false
		}
	}
	return true
}

func normalizedCleanupStatuses(statuses []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, status := range statuses {
		status = strings.TrimSpace(strings.ToLower(status))
		if status == "" || seen[status] {
			continue
		}
		seen[status] = true
		out = append(out, status)
	}
	sort.Strings(out)
	return out
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func cleanupTokenItems(items []domain.PluginPackageCleanupItem) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !item.CanDelete {
			continue
		}
		out = append(out, strings.Join([]string{item.Kind, item.ID, item.Path, item.Status, item.PluginCode, item.Version}, "|"))
	}
	sort.Strings(out)
	return out
}

func hasDeprecatedRootSchemaSQL(scan domain.PluginPackageFileScan) bool {
	for _, item := range append(append([]domain.PluginPackageFileEntry{}, scan.UnknownFiles...), scan.AllowedFiles...) {
		if strings.EqualFold(strings.TrimSpace(item.Path), "001_schema.sql") {
			return true
		}
	}
	return false
}

func buildPackageMigrationPlan(packageDir string, scan domain.PluginPackageFileScan) ([]domain.PluginPackageMigrationPlanItem, []string) {
	files := []domain.PluginPackageFileEntry{}
	for _, item := range scan.AllowedFiles {
		path := filepath.ToSlash(strings.TrimSpace(item.Path))
		if strings.HasPrefix(path, "migrations/") && strings.HasSuffix(strings.ToLower(path), ".sql") {
			files = append(files, item)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return filepath.ToSlash(files[i].Path) < filepath.ToSlash(files[j].Path)
	})

	plan := make([]domain.PluginPackageMigrationPlanItem, 0, len(files))
	warnings := []string{}
	for _, item := range files {
		path := filepath.ToSlash(strings.TrimSpace(item.Path))
		name := strings.TrimSuffix(strings.TrimPrefix(path, "migrations/"), filepath.Ext(path))
		full := filepath.Join(packageDir, filepath.FromSlash(path))
		raw, err := os.ReadFile(full)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("读取 migration 计划文件失败：%s", path))
			continue
		}
		sum := sha256.Sum256(raw)
		plan = append(plan, domain.PluginPackageMigrationPlanItem{
			Path:        path,
			Name:        name,
			Size:        item.Size,
			SHA256:      hex.EncodeToString(sum[:]),
			Source:      "migrations",
			WillExecute: false,
		})
	}
	return plan, warnings
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func uniqueStrings(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func scanBlockedCode(scan domain.PluginPackageFileScan) string {
	for _, item := range scan.DangerousFiles {
		switch strings.TrimSpace(item.Rule) {
		case "file_too_large", "manifest_too_large":
			return "plugin_package_file_too_large"
		}
	}
	for _, msg := range scan.Errors {
		if strings.Contains(msg, "文件数量超过限制") {
			return "plugin_package_too_many_files"
		}
		if strings.Contains(msg, "插件包总大小超过限制") {
			return "plugin_package_too_large"
		}
		if strings.Contains(msg, "单文件超过限制") {
			return "plugin_package_file_too_large"
		}
	}
	if len(scan.DangerousFiles) > 0 {
		return "plugin_package_dangerous_file"
	}
	if len(scan.Errors) > 0 {
		return "plugin_package_dry_run_blocked"
	}
	return ""
}
