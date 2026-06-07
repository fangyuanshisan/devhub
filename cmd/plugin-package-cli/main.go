package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

type checkSummary struct {
	Kind            string                          `json:"kind"`
	Path            string                          `json:"path,omitempty"`
	PluginCode      string                          `json:"plugin_code,omitempty"`
	Name            string                          `json:"name,omitempty"`
	Version         string                          `json:"version,omitempty"`
	Status          string                          `json:"status"`
	RiskLevel       string                          `json:"risk_level,omitempty"`
	Blockers        []string                        `json:"blockers,omitempty"`
	Warnings        []string                        `json:"warnings,omitempty"`
	Suggestions     []string                        `json:"suggestions,omitempty"`
	Migrations      []migrationSummary              `json:"migrations,omitempty"`
	FrontendMounts  []frontendMountSummary          `json:"frontend_mounts,omitempty"`
	ExternalService *externalServiceSummary         `json:"external_service,omitempty"`
	ForbiddenFiles  []domain.PluginPackageFileEntry `json:"forbidden_files,omitempty"`
	ChecksumStatus  string                          `json:"checksum_status,omitempty"`
	ManifestValid   bool                            `json:"manifest_valid"`
	Raw             any                             `json:"raw,omitempty"`
}

type migrationSummary struct {
	Path        string `json:"path"`
	WillExecute bool   `json:"will_execute"`
}

type frontendMountSummary struct {
	MountPoint   string `json:"mount_point"`
	ComponentKey string `json:"component_key"`
	RenderMode   string `json:"render_mode,omitempty"`
	Status       string `json:"status"`
}

type externalServiceSummary struct {
	Declared      bool   `json:"declared"`
	Endpoint      string `json:"endpoint,omitempty"`
	HealthPath    string `json:"health_check_path,omitempty"`
	TimeoutMS     int    `json:"timeout_ms,omitempty"`
	FailurePolicy string `json:"failure_policy,omitempty"`
	Status        string `json:"status"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var summary checkSummary
	var err error
	switch os.Args[1] {
	case "check":
		summary, err = runCheck(os.Args[2:])
	case "check-builtin":
		summary, err = runCheckBuiltin(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		var exitErr exitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	if summary.Status == "blocked" {
		os.Exit(1)
	}
}

type exitError struct {
	Code int
}

func (e exitError) Error() string { return fmt.Sprintf("exit %d", e.Code) }

func runCheck(args []string) (checkSummary, error) {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	path := fs.String("path", "", "plugin package directory or zip path")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		usage()
		return checkSummary{}, exitError{Code: 2}
	}
	if strings.TrimSpace(*path) == "" && fs.NArg() > 0 {
		*path = fs.Arg(0)
	}
	if strings.TrimSpace(*path) == "" {
		fmt.Fprintln(os.Stderr, "缺少 --path")
		return checkSummary{}, exitError{Code: 2}
	}
	checkPath := strings.TrimSpace(*path)
	cleanup := func() {}
	if strings.EqualFold(filepath.Ext(checkPath), ".zip") {
		dir, clean, err := extractZipPackage(checkPath)
		if err != nil {
			summary := checkSummary{
				Kind:          "package",
				Path:          checkPath,
				Status:        "blocked",
				RiskLevel:     "blocked",
				ManifestValid: false,
				Blockers:      []string{err.Error()},
				Suggestions:   []string{"请确认 zip 包结构合法，且 manifest.json 位于包根目录或唯一顶层目录内。"},
			}
			printSummary(summary, *jsonOut)
			return summary, nil
		}
		checkPath = dir
		cleanup = clean
	}
	defer cleanup()

	svc := service.New(store.NewMemoryStore())
	res, err := svc.DryRunPluginPackage(checkPath)
	if err != nil {
		summary := errorSummary(checkPath, err)
		printSummary(summary, *jsonOut)
		return summary, nil
	}
	summary := summarizeDryRun(res)
	printSummary(summary, *jsonOut)
	return summary, nil
}

func runCheckBuiltin(args []string) (checkSummary, error) {
	fs := flag.NewFlagSet("check-builtin", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	code := fs.String("code", "", "builtin plugin code")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		usage()
		return checkSummary{}, exitError{Code: 2}
	}
	if strings.TrimSpace(*code) == "" && fs.NArg() > 0 {
		*code = fs.Arg(0)
	}
	plugin, ok := pluginregistry.DefinitionByCode(strings.TrimSpace(*code))
	if !ok {
		summary := checkSummary{
			Kind:        "builtin",
			PluginCode:  strings.TrimSpace(*code),
			Status:      "blocked",
			RiskLevel:   "blocked",
			Blockers:    []string{"未找到内置插件定义"},
			Suggestions: []string{"请确认内置插件 code 是否正确。"},
		}
		printSummary(summary, *jsonOut)
		return summary, nil
	}
	blockers := []string{}
	warnings := []string{}
	for _, mount := range plugin.FrontendMounts {
		for _, msg := range pluginregistry.ValidateFrontendMount(mount) {
			blockers = append(blockers, msg)
		}
	}
	if err := pluginregistry.ValidateConfigJSON(plugin, sampleConfigForBuiltin(plugin.Code)); err != nil {
		blockers = append(blockers, "config_schema 编译或样例配置校验失败："+err.Error())
	}
	if strings.TrimSpace(plugin.Code) == "" || strings.TrimSpace(plugin.Name) == "" || strings.TrimSpace(plugin.Version) == "" {
		blockers = append(blockers, "内置插件 code/name/version 不能为空")
	}
	status := "passed"
	risk := "low"
	if len(blockers) > 0 {
		status = "blocked"
		risk = "blocked"
	} else if len(warnings) > 0 {
		status = "warning"
		risk = "medium"
	}
	summary := checkSummary{
		Kind:           "builtin",
		PluginCode:     plugin.Code,
		Name:           plugin.Name,
		Version:        plugin.Version,
		Status:         status,
		RiskLevel:      risk,
		Blockers:       uniqueStrings(blockers),
		Warnings:       uniqueStrings(warnings),
		Suggestions:    suggestionsFor(blockers, warnings),
		ManifestValid:  len(blockers) == 0,
		FrontendMounts: summarizeFrontendMounts(plugin.FrontendMounts),
	}
	printSummary(summary, *jsonOut)
	return summary, nil
}

func summarizeDryRun(res domain.PluginPackageDryRunResult) checkSummary {
	status := strings.TrimSpace(res.Status)
	if status == "ok" || status == "" {
		status = "passed"
	}
	blockers := append([]string{}, res.Errors...)
	for _, item := range res.RiskReport.Items {
		if strings.EqualFold(item.Level, "blocked") {
			blockers = append(blockers, formatRiskItem(item))
		}
	}
	warnings := append([]string{}, res.Warnings...)
	for _, item := range res.RiskReport.Items {
		if !strings.EqualFold(item.Level, "blocked") {
			warnings = append(warnings, formatRiskItem(item))
		}
	}
	manifest := res.InstallDryRun.NormalizedManifest
	return checkSummary{
		Kind:            "package",
		Path:            res.Package.Path,
		PluginCode:      res.Package.Code,
		Name:            res.Package.Name,
		Version:         res.Package.Version,
		Status:          status,
		RiskLevel:       firstNonBlank(res.RiskReport.Level, status),
		Blockers:        uniqueStrings(blockers),
		Warnings:        uniqueStrings(warnings),
		Suggestions:     suggestionsFor(blockers, warnings),
		Migrations:      summarizeMigrations(res.MigrationPlan),
		FrontendMounts:  summarizeFrontendMounts(manifest.FrontendMounts),
		ExternalService: summarizeExternalService(manifest.ExternalService, manifest),
		ForbiddenFiles:  res.FileScan.DangerousFiles,
		ChecksumStatus:  res.Checksum.Status,
		ManifestValid:   res.ManifestValidation.Valid,
	}
}

func errorSummary(path string, err error) checkSummary {
	msg := strings.TrimSpace(err.Error())
	code := ""
	suggestion := "请修复阻断项后重新执行插件包校验。"
	var apiErr *domain.APIError
	if errors.As(err, &apiErr) {
		code = strings.TrimSpace(apiErr.Code)
		msg = firstNonBlank(apiErr.Message, msg)
		suggestion = firstNonBlank(apiErr.Suggestion, suggestion)
	}
	if code != "" {
		msg = fmt.Sprintf("[%s] %s", code, msg)
	}
	return checkSummary{
		Kind:          "package",
		Path:          path,
		Status:        "blocked",
		RiskLevel:     "blocked",
		Blockers:      []string{msg},
		Suggestions:   []string{suggestion},
		ManifestValid: false,
	}
}

func printSummary(summary checkSummary, jsonOut bool) {
	if jsonOut {
		raw, _ := json.MarshalIndent(summary, "", "  ")
		fmt.Println(string(raw))
		return
	}
	fmt.Printf("插件包校验结果: %s\n", summary.Status)
	if summary.Kind != "" {
		fmt.Printf("类型: %s\n", summary.Kind)
	}
	if summary.Path != "" {
		fmt.Printf("路径: %s\n", summary.Path)
	}
	if summary.PluginCode != "" {
		fmt.Printf("plugin_code: %s\n", summary.PluginCode)
	}
	if summary.Name != "" {
		fmt.Printf("name: %s\n", summary.Name)
	}
	if summary.Version != "" {
		fmt.Printf("version: %s\n", summary.Version)
	}
	if summary.RiskLevel != "" {
		fmt.Printf("risk_level: %s\n", summary.RiskLevel)
	}
	fmt.Printf("manifest_valid: %v\n", summary.ManifestValid)
	if summary.ChecksumStatus != "" {
		fmt.Printf("checksum_status: %s\n", summary.ChecksumStatus)
	}
	printList("blockers", summary.Blockers)
	printList("warnings", summary.Warnings)
	printList("建议修复动作", summary.Suggestions)
	if len(summary.Migrations) > 0 {
		fmt.Println("migrations:")
		for _, item := range summary.Migrations {
			fmt.Printf("  - %s will_execute=%v\n", item.Path, item.WillExecute)
		}
	}
	if len(summary.FrontendMounts) > 0 {
		fmt.Println("frontend_mount allowlist:")
		for _, item := range summary.FrontendMounts {
			fmt.Printf("  - %s / %s / %s => %s\n", item.MountPoint, item.ComponentKey, item.RenderMode, item.Status)
		}
	}
	if summary.ExternalService != nil && summary.ExternalService.Declared {
		fmt.Printf("external_service: endpoint=%s health_check_path=%s timeout_ms=%d failure_policy=%s status=%s\n",
			summary.ExternalService.Endpoint,
			summary.ExternalService.HealthPath,
			summary.ExternalService.TimeoutMS,
			summary.ExternalService.FailurePolicy,
			summary.ExternalService.Status,
		)
	}
	if len(summary.ForbiddenFiles) > 0 {
		fmt.Println("forbidden_files:")
		for _, item := range summary.ForbiddenFiles {
			fmt.Printf("  - %s (%s)\n", item.Path, item.Rule)
		}
	}
}

func printList(title string, items []string) {
	fmt.Printf("%s:\n", title)
	if len(items) == 0 {
		fmt.Println("  - 无")
		return
	}
	for _, item := range items {
		fmt.Printf("  - %s\n", item)
	}
}

func extractZipPackage(path string) (string, func(), error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", func() {}, err
	}
	reader, err := zip.OpenReader(abs)
	if err != nil {
		return "", func() {}, fmt.Errorf("读取 zip 失败：%w", err)
	}
	defer reader.Close()
	var manifestRaw []byte
	for _, f := range reader.File {
		name := filepath.Clean(f.Name)
		if strings.HasSuffix(name, "manifest.json") && !f.FileInfo().IsDir() {
			in, err := f.Open()
			if err != nil {
				return "", func() {}, err
			}
			manifestRaw, err = io.ReadAll(in)
			_ = in.Close()
			if err != nil {
				return "", func() {}, err
			}
			break
		}
	}
	manifestCode := "package"
	if len(manifestRaw) > 0 {
		var manifest map[string]any
		if err := json.Unmarshal(manifestRaw, &manifest); err == nil {
			if code := firstNonBlank(asString(manifest["code"]), asString(manifest["plugin_code"])); code != "" {
				manifestCode = code
			}
		}
	}
	root, err := os.Getwd()
	if err != nil {
		return "", func() {}, err
	}
	tempRoot := filepath.Join(root, ".devhub", "plugins")
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		return "", func() {}, err
	}
	tmp, err := os.MkdirTemp(tempRoot, manifestCode+"-check-*")
	if err != nil {
		return "", func() {}, err
	}
	workRoot := filepath.Join(tmp, manifestCode)
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		_ = os.RemoveAll(tmp)
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }
	for _, f := range reader.File {
		name := filepath.Clean(f.Name)
		if name == "." || strings.HasPrefix(name, "..") || filepath.IsAbs(name) {
			cleanup()
			return "", func() {}, fmt.Errorf("zip 包包含不安全路径：%s", f.Name)
		}
		dst := filepath.Join(workRoot, name)
		if !strings.HasPrefix(dst, workRoot+string(os.PathSeparator)) && dst != workRoot {
			cleanup()
			return "", func() {}, fmt.Errorf("zip 包包含路径穿越：%s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dst, 0o755); err != nil {
				cleanup()
				return "", func() {}, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			cleanup()
			return "", func() {}, err
		}
		in, err := f.Open()
		if err != nil {
			cleanup()
			return "", func() {}, err
		}
		out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			_ = in.Close()
			cleanup()
			return "", func() {}, err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		_ = in.Close()
		if copyErr != nil {
			cleanup()
			return "", func() {}, copyErr
		}
		if closeErr != nil {
			cleanup()
			return "", func() {}, closeErr
		}
	}
	if fileExists(filepath.Join(workRoot, "manifest.json")) {
		return workRoot, cleanup, nil
	}
	entries, err := os.ReadDir(workRoot)
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	candidates := []string{}
	for _, ent := range entries {
		if ent.IsDir() && fileExists(filepath.Join(workRoot, ent.Name(), "manifest.json")) {
			candidates = append(candidates, filepath.Join(workRoot, ent.Name()))
		}
	}
	if len(candidates) == 1 {
		return candidates[0], cleanup, nil
	}
	cleanup()
	return "", func() {}, fmt.Errorf("zip 包根目录缺少 manifest.json")
}

func summarizeMigrations(items []domain.PluginPackageMigrationPlanItem) []migrationSummary {
	out := make([]migrationSummary, 0, len(items))
	for _, item := range items {
		out = append(out, migrationSummary{Path: item.Path, WillExecute: item.WillExecute})
	}
	return out
}

func summarizeFrontendMounts(items []domain.FrontendMountDefinition) []frontendMountSummary {
	out := make([]frontendMountSummary, 0, len(items))
	for _, mount := range items {
		status := "passed"
		if len(pluginregistry.ValidateFrontendMount(mount)) > 0 {
			status = "blocked"
		}
		out = append(out, frontendMountSummary{
			MountPoint:   mount.MountPoint,
			ComponentKey: mount.ComponentKey,
			RenderMode:   mount.RenderMode,
			Status:       status,
		})
	}
	return out
}

func summarizeExternalService(svc *domain.PluginExternalService, manifest domain.PluginManifest) *externalServiceSummary {
	if svc == nil {
		return nil
	}
	return &externalServiceSummary{
		Declared:      true,
		Endpoint:      svc.Endpoint,
		HealthPath:    configDefaultString(manifest.ConfigSchema, "health_check_path"),
		TimeoutMS:     svc.TimeoutMS,
		FailurePolicy: svc.FailurePolicy,
		Status:        "passed",
	}
}

func configDefaultString(schema any, key string) string {
	root, ok := schema.(map[string]any)
	if !ok {
		return ""
	}
	props, ok := root["properties"].(map[string]any)
	if !ok {
		return ""
	}
	item, ok := props[key].(map[string]any)
	if !ok {
		return ""
	}
	if value, ok := item["default"].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func suggestionsFor(blockers, warnings []string) []string {
	out := []string{}
	if len(blockers) > 0 {
		out = append(out, "先修复 blockers；blocked 会返回非零退出码，禁止进入 upload / promote / install dry-run。")
	}
	if len(warnings) > 0 {
		out = append(out, "warnings 可继续本地校验，但建议在上传前补齐 checksums、README、签名或迁移说明。")
	}
	if len(out) == 0 {
		out = append(out, "可继续执行 upload -> precheck -> promote -> install dry-run。")
	}
	return uniqueStrings(out)
}

func sampleConfigForBuiltin(code string) string {
	switch strings.TrimSpace(code) {
	case "official_announcement":
		return `{"enabled":true,"message":"CLI 校验公告","link_text":"查看详情","link_url":"/","dismissible":false}`
	default:
		return `{}`
	}
}

func formatRiskItem(item domain.PluginPackageRiskItem) string {
	msg := strings.TrimSpace(item.Message)
	if item.Code != "" {
		msg = "[" + item.Code + "] " + msg
	}
	if item.Path != "" {
		msg += " (" + item.Path + ")"
	}
	if item.Suggestion != "" {
		msg += "；建议：" + item.Suggestion
	}
	return msg
}

func uniqueStrings(items []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func firstNonBlank(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `用法:
  plugin-package-cli check --path <plugin-dir-or-zip> [--json]
  plugin-package-cli check-builtin --code official_announcement [--json]

说明:
  check 复用 DevHub package dry-run，仅解析本地文件，不执行 SQL、不执行插件代码、不执行 package scripts。
  status=blocked 返回退出码 1；status=warning/passed 返回退出码 0。`)
}
