package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginEnablePrecheckOperator struct {
	ID   int64
	Name string
}

type enableIssue struct {
	Code    string         `json:"code"`
	Section string         `json:"section,omitempty"`
	Path    string         `json:"path,omitempty"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type enableEvaluation struct {
	status    string
	canEnable bool
	errors    []enableIssue
	warnings  []enableIssue

	fileIntegrityResult map[string]any
	manifestResult      map[string]any
	dependencyResult    map[string]any
	configResult        map[string]any
	migrationResult     map[string]any
	permissionResult    map[string]any
	menuResult          map[string]any
	routeResult         map[string]any
	hookResult          map[string]any
	contentTypeResult   map[string]any
	runtimeResult       map[string]any
	summary             map[string]any
}

func (s *Service) latestEnablePrecheckChainForPlugin(pluginCode, version string) (domain.PluginPackagePrecheckRecord, domain.PluginPackageCompatCheckRecord, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	version = strings.TrimSpace(version)

	prechecks, _, err := s.repo.PluginPackagePrechecks(domain.PluginPackagePrecheckFilter{
		Status:     domain.PluginPackagePrecheckStatusPassed,
		PluginCode: pluginCode,
		Page:       1,
		PageSize:   100,
	})
	if err != nil {
		return domain.PluginPackagePrecheckRecord{}, domain.PluginPackageCompatCheckRecord{}, err
	}
	var pickedPrecheck domain.PluginPackagePrecheckRecord
	for _, it := range prechecks {
		if version != "" && strings.TrimSpace(it.Version) != version {
			continue
		}
		if it.ID > pickedPrecheck.ID {
			pickedPrecheck = it
		}
	}
	if pickedPrecheck.ID <= 0 {
		return domain.PluginPackagePrecheckRecord{}, domain.PluginPackageCompatCheckRecord{}, domain.NewPluginError("plugin_enable_precheck_precheck_missing", "缺少预检通过记录，不能执行启用前检查").
			WithStatus(400).
			WithDetail("plugin_code", pluginCode).
			WithDetail("version", version).
			WithSuggestion("请先完成 staging 包的解压安全检查与 manifest 预校验。")
	}

	compatChecks, _, err := s.repo.PluginPackageCompatChecks(domain.PluginPackageCompatCheckFilter{
		PluginCode:        pluginCode,
		PackagePrecheckID: pickedPrecheck.ID,
		Page:              1,
		PageSize:          100,
	})
	if err != nil {
		return domain.PluginPackagePrecheckRecord{}, domain.PluginPackageCompatCheckRecord{}, err
	}
	var pickedCompat domain.PluginPackageCompatCheckRecord
	for _, it := range compatChecks {
		if it.ID > pickedCompat.ID {
			pickedCompat = it
		}
	}
	if pickedCompat.ID <= 0 {
		return domain.PluginPackagePrecheckRecord{}, domain.PluginPackageCompatCheckRecord{}, domain.NewPluginError("plugin_enable_precheck_compat_missing", "缺少兼容性检查记录，不能执行启用前检查").
			WithStatus(400).
			WithDetail("plugin_code", pluginCode).
			WithDetail("version", version).
			WithDetail("package_precheck_id", pickedPrecheck.ID).
			WithSuggestion("请先对预检通过的包执行 compat-check。")
	}
	if !pickedCompat.CanInstall {
		return domain.PluginPackagePrecheckRecord{}, domain.PluginPackageCompatCheckRecord{}, domain.NewPluginError("plugin_enable_precheck_compat_blocked", "兼容性检查未通过，不能执行启用前检查").
			WithStatus(400).
			WithDetail("plugin_code", pluginCode).
			WithDetail("version", version).
			WithDetail("compat_check_id", pickedCompat.ID).
			WithDetail("compat_status", pickedCompat.Status).
			WithSuggestion("请先修复 compat-check 的 blockers 后重试。")
	}
	return pickedPrecheck, pickedCompat, nil
}

func (s *Service) RunPluginEnablePrecheckAs(operator PluginEnablePrecheckOperator, pluginCode string) (domain.PluginEnablePrecheckResponse, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" {
		return domain.PluginEnablePrecheckResponse{}, domain.NewPluginError("plugin_enable_precheck_invalid_request", "plugin_code 不能为空").
			WithStatus(400)
	}

	p, ok := s.repo.PluginByCode(pluginCode)
	if !ok || p.Code == "" {
		return domain.PluginEnablePrecheckResponse{}, pluginNotFound(pluginCode)
	}

	// Must be installed and not already enabled.
	if strings.TrimSpace(p.InstallStatus) == pluginregistry.StatusDiscovered || strings.TrimSpace(p.Status) == pluginregistry.StatusDiscovered {
		return domain.PluginEnablePrecheckResponse{}, pluginNotInstalled(pluginCode)
	}
	if strings.TrimSpace(p.Status) == pluginregistry.StatusEnabled {
		return domain.PluginEnablePrecheckResponse{}, domain.NewPluginError("plugin_enable_precheck_already_enabled", "插件已启用，无需启用前检查").
			WithStatus(400).
			WithDetail("plugin_code", pluginCode)
	}
	if strings.TrimSpace(p.Status) == pluginregistry.StatusArchived {
		return domain.PluginEnablePrecheckResponse{}, pluginArchived(pluginCode)
	}
	if strings.TrimSpace(p.Status) == pluginregistry.StatusMigrationFailed {
		return domain.PluginEnablePrecheckResponse{}, pluginMigrationFailed(pluginCode, "")
	}
	if strings.TrimSpace(p.Status) == pluginregistry.StatusDependencyMissing {
		return domain.PluginEnablePrecheckResponse{}, domain.NewPluginError(PluginErrDependencyMissing, "插件当前处于 dependency_missing 状态，不能执行启用前检查").
			WithStatus(400).
			WithDetail("plugin_code", pluginCode).
			WithSuggestion("请先修复依赖插件状态或版本后重试。")
	}
	// Failed/invalid states should not proceed.
	if strings.TrimSpace(p.Status) == pluginregistry.StatusConfigInvalid {
		return domain.PluginEnablePrecheckResponse{}, domain.NewPluginError("plugin_enable_precheck_invalid_status", "插件状态不允许执行启用前检查").
			WithStatus(400).
			WithDetail("plugin_code", pluginCode).
			WithDetail("status", p.Status).
			WithSuggestion("请先修复插件配置/状态异常后重试。")
	}

	// Must have a passed precheck + compat-check(can_install=true) chain for safety.
	precheck, compat, chainErr := s.latestEnablePrecheckChainForPlugin(p.Code, p.Version)
	if chainErr != nil {
		return domain.PluginEnablePrecheckResponse{}, chainErr
	}

	started := Now()
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.enable_precheck.requested",
		Target:    p.Code,
		Metadata: mustJSON(map[string]any{
			"plugin_code": p.Code,
			"version":     p.Version,
			"actor":       strings.TrimSpace(operator.Name),
		}),
		CreatedAt: Now(),
	})
	record := domain.PluginEnablePrecheckRecord{
		PluginCode: p.Code,
		Version:    p.Version,
		Status:     domain.PluginEnablePrecheckStatusChecking,
		CanEnable:  false,
		CoreVersion: currentCoreVersion(),
		CreatedBy:  operator.ID,
		StartedAt:  started,
		CreatedAt:  started,
	}
	record, _ = s.repo.AppendPluginEnablePrecheck(record)

	// Evaluate.
	ev := s.evaluateEnablePrecheck(p, precheck, compat)

	record.Status = ev.status
	record.CanEnable = ev.canEnable
	record.InstalledPath = asString(ev.fileIntegrityResult["installed_path"])
	record.ManifestSHA256 = asString(ev.manifestResult["manifest_sha256"])
	record.FileIntegrityResultJSON = mustJSON(scrubAnyForSnapshot(ev.fileIntegrityResult))
	record.ManifestResultJSON = mustJSON(scrubAnyForSnapshot(ev.manifestResult))
	record.DependencyResultJSON = mustJSON(scrubAnyForSnapshot(ev.dependencyResult))
	record.ConfigResultJSON = mustJSON(scrubAnyForSnapshot(ev.configResult))
	record.MigrationResultJSON = mustJSON(scrubAnyForSnapshot(ev.migrationResult))
	record.PermissionResultJSON = mustJSON(scrubAnyForSnapshot(ev.permissionResult))
	record.MenuResultJSON = mustJSON(scrubAnyForSnapshot(ev.menuResult))
	record.RouteResultJSON = mustJSON(scrubAnyForSnapshot(ev.routeResult))
	record.HookResultJSON = mustJSON(scrubAnyForSnapshot(ev.hookResult))
	record.ContentTypeResultJSON = mustJSON(scrubAnyForSnapshot(ev.contentTypeResult))
	record.RuntimeResultJSON = mustJSON(scrubAnyForSnapshot(ev.runtimeResult))
	record.WarningsJSON = mustJSON(enableWarningMessages(ev.warnings))
	record.ErrorsJSON = mustJSON(enableErrorMessages(ev.errors))
	record.SummaryJSON = mustJSON(scrubAnyForSnapshot(ev.summary))
	record.FinishedAt = Now()
	record, _ = s.repo.SavePluginEnablePrecheck(record)

	action := "plugin.enable_precheck.success"
	if record.Status == domain.PluginEnablePrecheckStatusWarning {
		action = "plugin.enable_precheck.success"
	}
	if record.Status == domain.PluginEnablePrecheckStatusConfigInvalid {
		action = "plugin.enable_precheck.config_invalid"
	} else if record.Status == domain.PluginEnablePrecheckStatusMigrationPending {
		action = "plugin.enable_precheck.migration_pending"
	} else if record.Status == domain.PluginEnablePrecheckStatusFileIntegrityFailed || record.Status == domain.PluginEnablePrecheckStatusManifestChanged {
		action = "plugin.enable_precheck.file_integrity_failed"
	} else if record.Status != domain.PluginEnablePrecheckStatusPassed && record.Status != domain.PluginEnablePrecheckStatusWarning {
		action = "plugin.enable_precheck.failed"
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    action,
		Target:    p.Code,
		Metadata: mustJSON(map[string]any{
			"plugin_code":  p.Code,
			"version":      p.Version,
			"status":       record.Status,
			"can_enable":   record.CanEnable,
			"precheck_id":  precheck.ID,
			"compat_check": compat.ID,
			"warnings":     enableWarningMessages(ev.warnings),
			"errors":       enableErrorMessages(ev.errors),
		}),
		CreatedAt: Now(),
	})

	return enablePrecheckResponse(record), nil
}

func enableErrorMessages(items []enableIssue) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Code+": "+it.Message)
	}
	sort.Strings(out)
	return out
}

func enableWarningMessages(items []enableIssue) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Code+": "+it.Message)
	}
	sort.Strings(out)
	return out
}

func (s *Service) evaluateEnablePrecheck(p domain.Plugin, precheck domain.PluginPackagePrecheckRecord, compat domain.PluginPackageCompatCheckRecord) enableEvaluation {
	coreVersion := currentCoreVersion()
	ev := enableEvaluation{
		status:    domain.PluginEnablePrecheckStatusPassed,
		canEnable: true,
		errors:    []enableIssue{},
		warnings:  []enableIssue{},
	}
	addErr := func(code, section, path, msg string, details map[string]any) {
		ev.errors = append(ev.errors, enableIssue{Code: code, Section: section, Path: path, Message: msg, Details: details})
	}
	addWarn := func(code, section, path, msg string, details map[string]any) {
		ev.warnings = append(ev.warnings, enableIssue{Code: code, Section: section, Path: path, Message: msg, Details: details})
	}

	// 1) File integrity: re-scan the package path captured by precheck (or fallback).
	fileRes, fileIssuesErr := s.enablePrecheckFileIntegrity(p, precheck)
	ev.fileIntegrityResult = fileRes
	if fileIssuesErr != nil {
		addErr(fileIssuesErr.Code, "file_integrity", "", fileIssuesErr.Message, fileIssuesErr.Details)
	}

	// 1.1) Record chain snapshot (precheck + compat-check). No mutation; this is for traceability.
	ev.runtimeResult = map[string]any{
		"precheck_id":     precheck.ID,
		"precheck_status": precheck.Status,
		"compat_check_id": compat.ID,
		"compat_status":   compat.Status,
		"can_install":     compat.CanInstall,
	}

	// 2) Manifest re-validate from stored snapshot.
	manifestRes, manifestIssues := s.enablePrecheckManifest(p, coreVersion)
	ev.manifestResult = manifestRes
	for _, it := range manifestIssues {
		if it.Code == "plugin_core_version_incompatible" {
			addErr("plugin_enable_precheck_core_incompatible", "core", "compatible_core_version", it.Message, it.Details)
		} else {
			addErr(it.Code, "manifest", it.Path, it.Message, it.Details)
		}
	}

	// 3) Dependency re-check (reuse existing readiness logic but with structured output).
	depRes, depIssues := s.enablePrecheckDependencies(p)
	ev.dependencyResult = depRes
	for _, it := range depIssues {
		if it.Details != nil && it.Details["required"] == false {
			addWarn(it.Code, "dependencies", it.Path, it.Message, it.Details)
		} else {
			addErr(it.Code, "dependencies", it.Path, it.Message, it.Details)
		}
	}

	// 4) Config validity.
	configRes, configIssues := s.enablePrecheckConfig(p)
	ev.configResult = configRes
	for _, it := range configIssues {
		addErr(it.Code, "config", it.Path, it.Message, it.Details)
	}

	// 5) Migration status.
	migRes, migIssues := s.enablePrecheckMigrations(p)
	ev.migrationResult = migRes
	for _, it := range migIssues {
		addErr(it.Code, "migrations", it.Path, it.Message, it.Details)
	}

	// 6) Conflict checks (permissions/menus/routes/hooks/content types) against current installed plugins.
	existing := s.repo.Plugins()
	permSet := map[string]bool{}
	for _, perm := range p.Permissions {
		permSet[strings.TrimSpace(perm.Code)] = true
	}
	permErrors, permResult := checkCompatPermissions(p.PluginManifest, permSet, existing)
	ev.permissionResult = permResult
	for _, it := range permErrors {
		addErr(it.Code, "permissions", it.Path, it.Message, it.Details)
	}
	menuErrors, menuWarnings, menuResult := checkCompatMenus(p.PluginManifest, permSet, existing)
	ev.menuResult = menuResult
	for _, it := range menuErrors {
		addErr(it.Code, "menus", it.Path, it.Message, it.Details)
	}
	for _, it := range menuWarnings {
		addWarn(it.Code, "menus", it.Path, it.Message, it.Details)
	}
	routeErrors, routeResult := checkCompatRoutes(p.PluginManifest, permSet, existing)
	ev.routeResult = routeResult
	for _, it := range routeErrors {
		addErr(it.Code, "routes", it.Path, it.Message, it.Details)
	}
	hookErrors, hookResult := checkCompatHooks(p.PluginManifest)
	ev.hookResult = hookResult
	for _, it := range hookErrors {
		addErr(it.Code, "hooks", it.Path, it.Message, it.Details)
	}

	// content_types owner/conflict check against enabled plugins only (stricter for enable).
	ctRes, ctIssues := s.enablePrecheckContentTypes(p, existing)
	ev.contentTypeResult = ctRes
	for _, it := range ctIssues {
		addErr(it.Code, "content_types", it.Path, it.Message, it.Details)
	}

	// 7) Runtime readiness summary (does not enable/register).
	for k, v := range map[string]any{
		"will_register":        false,
		"does_not_enable":      true,
		"does_not_execute":     true,
		"does_not_load_plugin": true,
	} {
		ev.runtimeResult[k] = v
	}

	ev.canEnable = len(ev.errors) == 0
	ev.summary = map[string]any{
		"file_integrity_ok": noEnableIssue(ev.errors, "file_integrity"),
		"manifest_ok":       noEnableIssue(ev.errors, "manifest") && noEnableIssue(ev.errors, "core"),
		"dependencies_ok":   noEnableIssue(ev.errors, "dependencies"),
		"config_ok":         noEnableIssue(ev.errors, "config"),
		"migrations_ok":     noEnableIssue(ev.errors, "migrations"),
		"permissions_ok":    noEnableIssue(ev.errors, "permissions"),
		"menus_ok":          noEnableIssue(ev.errors, "menus"),
		"routes_ok":         noEnableIssue(ev.errors, "routes"),
		"hooks_ok":          noEnableIssue(ev.errors, "hooks"),
		"content_types_ok":  noEnableIssue(ev.errors, "content_types"),
		"errors_count":      len(ev.errors),
		"warnings_count":    len(ev.warnings),
		"can_enable":        ev.canEnable,
		"enable_blocked":    !ev.canEnable,
		"core_version":      coreVersion,
		"plugin_status":     p.Status,
	}

	ev.status = enableStatus(ev)
	return ev
}

func noEnableIssue(items []enableIssue, section string) bool {
	for _, it := range items {
		if it.Section == section {
			return false
		}
	}
	return true
}

func enableStatus(ev enableEvaluation) string {
	if len(ev.errors) == 0 {
		if len(ev.warnings) > 0 {
			return domain.PluginEnablePrecheckStatusWarning
		}
		return domain.PluginEnablePrecheckStatusPassed
	}
	for _, it := range ev.errors {
		switch it.Section {
		case "dependencies":
			return domain.PluginEnablePrecheckStatusDependencyMissing
		case "config":
			return domain.PluginEnablePrecheckStatusConfigInvalid
		case "migrations":
			return domain.PluginEnablePrecheckStatusMigrationPending
		case "file_integrity":
			return domain.PluginEnablePrecheckStatusFileIntegrityFailed
		case "permissions", "routes", "menus", "content_types":
			return domain.PluginEnablePrecheckStatusConflictDetected
		}
	}
	return domain.PluginEnablePrecheckStatusFailed
}

type precheckIssue struct {
	Code    string
	Message string
	Path    string
	Details map[string]any
}

func (s *Service) enablePrecheckManifest(p domain.Plugin, coreVersion string) (map[string]any, []precheckIssue) {
	issues := []precheckIssue{}
	manifest := p.PluginManifest
	validation := pluginregistry.ValidatePluginManifest(manifest, s.repo.Plugins(), coreVersion)
	res := map[string]any{
		"valid":      validation.Valid,
		"errors":     validation.Errors,
		"warnings":   validation.Warnings,
		"checksum":   validation.Checksum,
		"core":       validation.Compatibility,
		"plugin_code": manifest.Code,
		"version":    manifest.Version,
		"manifest_sha256": validation.Checksum,
	}
	if !validation.Valid {
		for _, msg := range validation.Errors {
			issues = append(issues, precheckIssue{Code: "plugin_enable_precheck_manifest_invalid", Message: msg})
		}
	}
	// core incompat is already included in Errors; add structured issue for easier UI.
	if validation.Compatibility.Status == pluginregistry.CompatibilityIncompatible {
		issues = append(issues, precheckIssue{
			Code:    "plugin_core_version_incompatible",
			Path:    "compatible_core_version",
			Message: "当前 Core 版本不满足插件兼容范围",
			Details: map[string]any{
				"core_version":           coreVersion,
				"compatible_core_version": manifest.CompatibleCoreVersion,
				"messages":               validation.Compatibility.Messages,
			},
		})
	}
	return res, issues
}

func (s *Service) enablePrecheckDependencies(p domain.Plugin) (map[string]any, []precheckIssue) {
	checks, summary, errs, warns := pluginregistry.ValidatePluginDependencies(p.PluginManifest, s.repo.Plugins())
	issues := []precheckIssue{}
	for _, msg := range errs {
		issues = append(issues, precheckIssue{Code: "plugin_dependency_missing", Message: msg, Details: map[string]any{"required": true}})
	}
	for _, msg := range warns {
		issues = append(issues, precheckIssue{Code: "plugin_dependency_optional_missing", Message: msg, Details: map[string]any{"required": false}})
	}
	return map[string]any{"checks": checks, "summary": summary}, issues
}

func (s *Service) enablePrecheckConfig(p domain.Plugin) (map[string]any, []precheckIssue) {
	issues := []precheckIssue{}
	if err := pluginregistry.ValidateConfigJSON(p, p.ConfigJSON); err != nil {
		msg := strings.TrimSpace(err.Error())
		path := ""
		if strings.HasPrefix(msg, "$") {
			parts := strings.Fields(msg)
			if len(parts) > 0 {
				path = parts[0]
			}
		}
		issues = append(issues, precheckIssue{
			Code:    "plugin_config_schema_invalid",
			Path:    path,
			Message: msg,
		})
	}
	return map[string]any{"valid": len(issues) == 0}, issues
}

func (s *Service) enablePrecheckMigrations(p domain.Plugin) (map[string]any, []precheckIssue) {
	issues := []precheckIssue{}
	items, err := s.pluginMigrationsWithDefinitions(p.Code)
	if err != nil {
		issues = append(issues, precheckIssue{Code: "plugin_enable_precheck_migration_unavailable", Message: err.Error()})
		return map[string]any{"status": "unknown"}, issues
	}
	pending := 0
	failed := 0
	for _, it := range items {
		switch strings.TrimSpace(it.Status) {
		case "pending":
			pending++
		case "failed":
			failed++
		}
	}
	if failed > 0 {
		issues = append(issues, precheckIssue{Code: "plugin_migration_failed", Message: "存在失败迁移，禁止启用"})
	}
	if pending > 0 {
		issues = append(issues, precheckIssue{Code: "plugin_migration_pending", Message: "存在未完成迁移，禁止启用"})
	}
	return map[string]any{"pending": pending, "failed": failed, "total": len(items), "will_execute": false}, issues
}

type fileIntegrityErr struct {
	Code    string
	Message string
	Details map[string]any
}

func (s *Service) enablePrecheckFileIntegrity(p domain.Plugin, precheck domain.PluginPackagePrecheckRecord) (map[string]any, *fileIntegrityErr) {
	// This repo currently persists plugin manifest/config in DB but does not persist an installed filesystem snapshot.
	// For package-installed plugins, we re-check the package path captured by precheck (preferred),
	// falling back to storage/plugins/packages/{code}.
	out := map[string]any{
		"checked":        false,
		"installed_path": "",
		"note":           "当前未持久化 installed_path；启用前文件完整性检查以预检记录 package_path 复检为准（若存在）。",
	}

	if strings.TrimSpace(p.SourceType) != "local_package" {
		// Built-in/manifest plugins have no package directory; treat as checked=false + warning only.
		return out, nil
	}

	rel := strings.TrimSpace(precheck.PackagePath)
	if rel == "" {
		rel = filepath.ToSlash(filepath.Join("storage", "plugins", "packages", p.Code))
	}
	abs, _, err := pluginregistry.NormalizePluginPackagePath(rel)
	if err != nil {
		// If package path is missing, block enabling for safety.
		return out, &fileIntegrityErr{
			Code:    "plugin_enable_precheck_installed_path_missing",
			Message: "未找到插件源包目录，无法执行文件完整性复检",
			Details: map[string]any{"candidate_path": rel},
		}
	}
	st, statErr := os.Stat(abs)
	if statErr != nil || !st.IsDir() {
		return out, &fileIntegrityErr{
			Code:    "plugin_enable_precheck_installed_path_missing",
			Message: "插件源包目录不存在，无法执行文件完整性复检",
			Details: map[string]any{"candidate_path": rel},
		}
	}

	// Re-run scan + checksum (if checksums.json exists) + read manifest.json and compare checksum with stored manifest checksum.
	scan, scanErr := pluginregistry.ScanPluginPackage(abs)
	out["checked"] = true
	out["installed_path"] = rel
	out["file_scan"] = scan
	if scanErr != nil {
		return out, &fileIntegrityErr{Code: "plugin_enable_precheck_file_scan_failed", Message: scanErr.Error()}
	}
	if len(scan.DangerousFiles) > 0 {
		return out, &fileIntegrityErr{Code: "plugin_package_dangerous_file", Message: "插件包包含危险文件，禁止启用"}
	}

	checksum, cerr := pluginregistry.VerifyPluginPackageChecksums(abs, scan)
	out["checksum"] = checksum
	if cerr != nil {
		if api, ok := cerr.(*domain.APIError); ok && api != nil {
			return out, &fileIntegrityErr{Code: api.Code, Message: api.Message, Details: api.Details}
		}
		return out, &fileIntegrityErr{Code: "plugin_enable_precheck_checksum_failed", Message: cerr.Error()}
	}

	manifestPath := filepath.Join(abs, "manifest.json")
	raw, rerr := os.ReadFile(manifestPath)
	if rerr != nil {
		return out, &fileIntegrityErr{Code: "plugin_enable_precheck_manifest_missing", Message: "manifest.json 不存在或无法读取"}
	}
	var manifest domain.PluginManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return out, &fileIntegrityErr{Code: "plugin_enable_precheck_manifest_invalid", Message: "manifest.json 不是合法 JSON"}
	}
	manifest = pluginregistry.NormalizeManifest(manifest)
	got := pluginregistry.ManifestChecksum(manifest)
	out["manifest_sha256"] = got
	out["expected_manifest_sha256"] = strings.TrimSpace(p.ManifestChecksum)
	if strings.TrimSpace(p.ManifestChecksum) != "" && got != strings.TrimSpace(p.ManifestChecksum) {
		return out, &fileIntegrityErr{
			Code:    "plugin_enable_precheck_manifest_changed",
			Message: "manifest 与安装时快照不一致，禁止启用",
			Details: map[string]any{"expected": p.ManifestChecksum, "actual": got},
		}
	}
	return out, nil
}

func (s *Service) enablePrecheckContentTypes(p domain.Plugin, existing []domain.Plugin) (map[string]any, []precheckIssue) {
	issues := []precheckIssue{}
	conflicts := []string{}
	enabledTypes := map[string]string{}
	for _, it := range existing {
		if strings.TrimSpace(it.Status) != pluginregistry.StatusEnabled {
			continue
		}
		for _, t := range it.ContentTypes {
			enabledTypes[pluginregistry.NormalizeContentType(t)] = it.Code
		}
		for _, def := range it.ContentTypeDefs {
			enabledTypes[pluginregistry.NormalizeContentType(def.Type)] = it.Code
		}
	}
	for _, def := range p.ContentTypeDefs {
		if def.PluginCode != "" && def.PluginCode != p.Code {
			issues = append(issues, precheckIssue{Code: "plugin_enable_precheck_content_type_owner_invalid", Path: def.Type + ".plugin_code", Message: "content_type.plugin_code 必须等于当前 plugin_code"})
		}
		typ := pluginregistry.NormalizeContentType(def.Type)
		if owner := enabledTypes[typ]; owner != "" && owner != p.Code {
			conflicts = append(conflicts, typ)
		}
		if strings.TrimSpace(def.CreatePermission) != "" {
			found := false
			for _, perm := range p.Permissions {
				if strings.TrimSpace(perm.Code) == strings.TrimSpace(def.CreatePermission) {
					found = true
					break
				}
			}
			if !found {
				issues = append(issues, precheckIssue{Code: "plugin_enable_precheck_permission_missing", Path: def.Type + ".create_permission", Message: "create_permission 未在 permissions 中声明"})
			}
		}
	}
	if len(conflicts) > 0 {
		issues = append(issues, precheckIssue{Code: "plugin_enable_precheck_content_type_conflict", Message: "content_type 与已启用插件冲突", Details: map[string]any{"content_types": uniqueStrings(conflicts)}})
	}
	return map[string]any{"conflicts": uniqueStrings(conflicts), "checked_against_enabled": true}, issues
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return fmt.Sprintf("%v", v)
}

func enablePrecheckResponse(record domain.PluginEnablePrecheckRecord) domain.PluginEnablePrecheckResponse {
	parseAny := func(raw string) any {
		if strings.TrimSpace(raw) == "" {
			return nil
		}
		var out any
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return nil
		}
		return out
	}
	parseStrings := func(raw string) []string {
		out := []string{}
		if strings.TrimSpace(raw) != "" {
			_ = json.Unmarshal([]byte(raw), &out)
		}
		return out
	}
	return domain.PluginEnablePrecheckResponse{
		PluginEnablePrecheckRecord: record,
		FileIntegrityResult:        parseAny(record.FileIntegrityResultJSON),
		ManifestResult:             parseAny(record.ManifestResultJSON),
		DependencyResult:           parseAny(record.DependencyResultJSON),
		ConfigResult:               parseAny(record.ConfigResultJSON),
		MigrationResult:            parseAny(record.MigrationResultJSON),
		PermissionResult:           parseAny(record.PermissionResultJSON),
		MenuResult:                 parseAny(record.MenuResultJSON),
		RouteResult:                parseAny(record.RouteResultJSON),
		HookResult:                 parseAny(record.HookResultJSON),
		ContentTypeResult:          parseAny(record.ContentTypeResultJSON),
		RuntimeResult:              parseAny(record.RuntimeResultJSON),
		Warnings:                   parseStrings(record.WarningsJSON),
		Errors:                     parseStrings(record.ErrorsJSON),
		Summary:                    parseAny(record.SummaryJSON),
	}
}

func (s *Service) ListPluginEnablePrechecks(filter domain.PluginEnablePrecheckFilter) (domain.PluginEnablePrecheckListResponse, error) {
	items, total, err := s.repo.PluginEnablePrechecks(filter)
	if err != nil {
		return domain.PluginEnablePrecheckListResponse{}, err
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	out := make([]domain.PluginEnablePrecheckResponse, 0, len(items))
	for _, it := range items {
		out = append(out, enablePrecheckResponse(it))
	}
	return domain.PluginEnablePrecheckListResponse{
		Items: out,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) GetPluginEnablePrecheck(id int64) (domain.PluginEnablePrecheckResponse, error) {
	it, ok := s.repo.PluginEnablePrecheckByID(id)
	if !ok || it.ID <= 0 {
		return domain.PluginEnablePrecheckResponse{}, domain.NewPluginError("plugin_enable_precheck_not_found", "启用前检查记录不存在").
			WithStatus(404).
			WithSuggestion("请刷新启用前检查列表后重试。")
	}
	return enablePrecheckResponse(it), nil
}

func (s *Service) DeletePluginEnablePrecheckAs(operator PluginEnablePrecheckOperator, id int64) (domain.PluginEnablePrecheckResponse, error) {
	it, ok := s.repo.PluginEnablePrecheckByID(id)
	if !ok || it.ID <= 0 {
		return domain.PluginEnablePrecheckResponse{}, domain.NewPluginError("plugin_enable_precheck_not_found", "启用前检查记录不存在").
			WithStatus(404)
	}
	it.Status = domain.PluginEnablePrecheckStatusDeleted
	it.CanEnable = false
	it.UpdatedAt = Now()
	it, _ = s.repo.SavePluginEnablePrecheck(it)

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.enable_precheck.deleted",
		Target:    it.PluginCode,
		Metadata: mustJSON(map[string]any{
			"enable_precheck_id": it.ID,
			"plugin_code":        it.PluginCode,
			"version":            it.Version,
			"status":             it.Status,
		}),
		CreatedAt: Now(),
	})

	return enablePrecheckResponse(it), nil
}
