package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const pluginExportRoot = "storage/plugins/exports"

var pluginExportSafePartPattern = regexp.MustCompile(`[^a-z0-9._-]+`)

func (s *Service) DryRunPluginPackageExport(code string, req domain.PluginPackageExportRequest) (domain.PluginPackageExportDryRunResponse, error) {
	plugin, manifest, err := s.exportablePluginManifest(code)
	if err != nil {
		return domain.PluginPackageExportDryRunResponse{}, err
	}
	outputDir, err := s.resolvePluginExportDir(manifest, req.OutputDir)
	if err != nil {
		return domain.PluginPackageExportDryRunResponse{}, err
	}
	files := pluginExportFileList(manifest, req)
	warnings := []string{}
	if plugin.IsSystem {
		warnings = append(warnings, "导出内置系统插件时会将 is_system 写为 false，导出包仅作为声明型插件包使用。")
	}
	if !req.IncludePublisher {
		warnings = append(warnings, "未导出 publisher.json；后续本地插件包 dry-run 会提示发布者信息缺失。")
	}
	if !req.IncludeSignatureStub {
		warnings = append(warnings, "未导出 signature.json；后续本地插件包 dry-run 会提示未签名。")
	}
	return domain.PluginPackageExportDryRunResponse{
		PluginCode: manifest.Code,
		Version:    manifest.Version,
		Status:     "ok",
		ExportPreview: domain.PluginPackageExportPreview{
			Files:                   files,
			OutputDir:               outputDir,
			ContainsSensitiveValues: false,
			ContainsUserData:        false,
			ContainsRuntimeCode:     false,
			ContainsExternalSQL:     false,
		},
		Warnings: uniqueStrings(warnings),
	}, nil
}

func (s *Service) ExportPluginPackage(code string, req domain.PluginPackageExportRequest) (domain.PluginPackageExportResponse, error) {
	preview, err := s.DryRunPluginPackageExport(code, req)
	if err != nil {
		return domain.PluginPackageExportResponse{}, err
	}
	plugin, manifest, err := s.exportablePluginManifest(code)
	if err != nil {
		return domain.PluginPackageExportResponse{}, err
	}
	root, absOut, err := normalizePluginExportOutputDir(preview.ExportPreview.OutputDir)
	if err != nil {
		return domain.PluginPackageExportResponse{}, err
	}
	if _, statErr := os.Stat(absOut); statErr == nil {
		if !req.Force {
			return domain.PluginPackageExportResponse{}, domain.NewPluginError("plugin_export_output_exists", "导出目录已存在").
				WithStatus(400).
				WithDetail("output_dir", preview.ExportPreview.OutputDir).
				WithSuggestion("请更换 output_dir，或设置 force=true 后重试。")
		}
		if remErr := os.RemoveAll(absOut); remErr != nil {
			return domain.PluginPackageExportResponse{}, domain.NewPluginError("plugin_export_failed", "清理已有导出目录失败").
				WithStatus(500).
				WithDetail("output_dir", preview.ExportPreview.OutputDir).
				WithSuggestion("请检查 storage/plugins/exports 目录权限后重试。")
		}
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return domain.PluginPackageExportResponse{}, statErr
	}

	if err := os.MkdirAll(absOut, 0755); err != nil {
		return domain.PluginPackageExportResponse{}, domain.NewPluginError("plugin_export_failed", "创建导出目录失败").
			WithStatus(500).
			WithDetail("output_dir", preview.ExportPreview.OutputDir).
			WithSuggestion("请检查 storage/plugins/exports 目录权限后重试。")
	}
	cleanupOnError := true
	defer func() {
		if cleanupOnError {
			_ = os.RemoveAll(absOut)
		}
	}()

	written := []string{}
	writeJSON := func(rel string, value any) error {
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		raw = append(raw, '\n')
		return writeExportFile(absOut, rel, raw, &written)
	}
	if err := writeJSON("manifest.json", manifest); err != nil {
		return domain.PluginPackageExportResponse{}, err
	}
	if err := writeExportFile(absOut, "README.md", []byte(buildPluginExportReadme(plugin, manifest)), &written); err != nil {
		return domain.PluginPackageExportResponse{}, err
	}
	configExample := buildConfigExample(manifest.ConfigSchema)
	if err := writeJSON("config.example.json", configExample); err != nil {
		return domain.PluginPackageExportResponse{}, err
	}
	if req.IncludeDocs {
		if err := writeExportFile(absOut, "docs/usage.md", []byte(buildPluginExportUsage(manifest)), &written); err != nil {
			return domain.PluginPackageExportResponse{}, err
		}
	}
	if req.IncludeMigrations && len(manifest.Migrations) > 0 {
		if err := writeJSON("migrations/exported_migrations.json", map[string]any{"migrations": manifest.Migrations}); err != nil {
			return domain.PluginPackageExportResponse{}, err
		}
	}
	if req.IncludePublisher {
		if err := writeJSON("publisher.json", buildPublisherStub(manifest)); err != nil {
			return domain.PluginPackageExportResponse{}, err
		}
	}
	if req.IncludeSignatureStub {
		if err := writeJSON("signature.json", buildSignatureStub(manifest, written)); err != nil {
			return domain.PluginPackageExportResponse{}, err
		}
	}

	checksumFiles, err := buildExportChecksums(absOut, written)
	if err != nil {
		return domain.PluginPackageExportResponse{}, domain.NewPluginError("plugin_export_checksum_failed", "生成 checksums.json 失败").
			WithStatus(500).
			WithDetail("output_dir", preview.ExportPreview.OutputDir).
			WithSuggestion("请检查导出目录文件权限后重试。")
	}
	if err := writeJSON("checksums.json", map[string]any{"algorithm": "sha256", "files": checksumFiles}); err != nil {
		return domain.PluginPackageExportResponse{}, err
	}
	written = append(written, "checksums.json")
	sort.Strings(written)

	dry, dryErr := s.DryRunPluginPackage(relativeToRoot(root, absOut))
	packageStatus := ""
	warnings := append([]string{}, preview.Warnings...)
	if dryErr != nil {
		if apiErr, ok := dryErr.(*domain.APIError); ok && apiErr != nil {
			warnings = append(warnings, fmt.Sprintf("[%s] %s", apiErr.Code, apiErr.Message))
			packageStatus = "failed"
		} else {
			warnings = append(warnings, dryErr.Error())
			packageStatus = "failed"
		}
	} else {
		packageStatus = dry.Status
		if dry.Status == "blocked" {
			warnings = append(warnings, "导出后 package dry-run 被阻断，请检查导出包。")
		}
		warnings = append(warnings, dry.Warnings...)
	}
	cleanupOnError = false

	return domain.PluginPackageExportResponse{
		Message:             "插件已导出为本地声明型插件包",
		PluginCode:          manifest.Code,
		Version:             manifest.Version,
		OutputDir:           preview.ExportPreview.OutputDir,
		Files:               written,
		ChecksumStatus:      "generated",
		PackageDryRunStatus: packageStatus,
		Warnings:            uniqueStrings(warnings),
	}, nil
}

func (s *Service) exportablePluginManifest(code string) (domain.Plugin, domain.PluginManifest, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.Plugin{}, domain.PluginManifest{}, domain.NewPluginError("plugin_not_found", "插件不存在").
			WithStatus(404).
			WithSuggestion("请提供有效插件 code。")
	}
	plugin, ok := s.repo.PluginByCode(code)
	if !ok {
		return domain.Plugin{}, domain.PluginManifest{}, domain.NewPluginError("plugin_not_found", "插件不存在").
			WithStatus(404).
			WithDetail("plugin_code", code).
			WithSuggestion("请确认插件已安装后再导出。")
	}
	manifest := plugin.PluginManifest
	if strings.TrimSpace(plugin.ManifestJSON) != "" {
		var decoded domain.PluginManifest
		if err := json.Unmarshal([]byte(plugin.ManifestJSON), &decoded); err == nil && strings.TrimSpace(decoded.Code) != "" {
			manifest = decoded
		}
	}
	manifest.Code = firstNonBlank(manifest.Code, plugin.Code)
	manifest.Name = firstNonBlank(manifest.Name, plugin.Name)
	manifest.Version = firstNonBlank(manifest.Version, plugin.Version)
	manifest.Status = ""
	manifest.SourceType = ""
	manifest.PluginCode = ""
	manifest.IsSystem = false
	if strings.TrimSpace(manifest.Code) == "" || strings.TrimSpace(manifest.Version) == "" {
		return domain.Plugin{}, domain.PluginManifest{}, domain.NewPluginError("plugin_export_manifest_invalid", "插件 manifest 缺少 code 或 version，无法导出").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithSuggestion("请修复插件 manifest 后重试。")
	}
	raw, _ := json.Marshal(manifest)
	existing := make([]domain.Plugin, 0, len(s.repo.Plugins()))
	for _, item := range s.repo.Plugins() {
		if item.Code == plugin.Code {
			continue
		}
		existing = append(existing, item)
	}
	validation := pluginregistry.ValidatePluginManifestJSON(raw, existing, currentCoreVersion())
	if !validation.Valid {
		return domain.Plugin{}, domain.PluginManifest{}, domain.NewPluginError("plugin_export_manifest_invalid", "导出的 manifest 未通过校验").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithDetail("errors", validation.Errors).
			WithSuggestion("请修复插件声明后再导出。")
	}
	return plugin, validation.NormalizedManifest, nil
}

func (s *Service) resolvePluginExportDir(manifest domain.PluginManifest, outputDir string) (string, error) {
	if strings.TrimSpace(outputDir) != "" {
		clean, _, err := normalizePluginExportOutputDir(outputDir)
		return clean, err
	}
	name := fmt.Sprintf("%s-%s-%s", safeExportPart(manifest.Code), safeExportPart(manifest.Version), time.Now().UTC().Format("20060102T150405Z"))
	return filepath.ToSlash(filepath.Join(pluginExportRoot, name)), nil
}

func normalizePluginExportOutputDir(input string) (clean string, abs string, err error) {
	clean = filepath.Clean(strings.TrimSpace(input))
	if clean == "" || clean == "." {
		return "", "", domain.NewPluginError("plugin_export_path_invalid", "导出路径不能为空").
			WithStatus(400).
			WithSuggestion("请使用 storage/plugins/exports 下的相对路径。")
	}
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." || strings.Contains(clean, string(filepath.Separator)+".."+string(filepath.Separator)) {
		return "", "", domain.NewPluginError("plugin_export_path_invalid", "导出路径不合法").
			WithStatus(400).
			WithDetail("output_dir", input).
			WithSuggestion("导出目录必须是 storage/plugins/exports 下的相对路径，不能包含路径穿越。")
	}
	clean = filepath.ToSlash(clean)
	if !strings.HasPrefix(clean, pluginExportRoot+"/") {
		return "", "", domain.NewPluginError("plugin_export_path_invalid", "导出目录不在允许范围内").
			WithStatus(400).
			WithDetail("output_dir", input).
			WithSuggestion("请导出到 storage/plugins/exports/ 下。")
	}
	root, err := serviceProjectRoot()
	if err != nil {
		return "", "", err
	}
	abs = filepath.Clean(filepath.Join(root, filepath.FromSlash(clean)))
	exportRootAbs := filepath.Clean(filepath.Join(root, pluginExportRoot))
	rel, relErr := filepath.Rel(exportRootAbs, abs)
	if relErr != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return "", "", domain.NewPluginError("plugin_export_path_invalid", "导出目录不在允许范围内").
			WithStatus(400).
			WithDetail("output_dir", input).
			WithSuggestion("请导出到 storage/plugins/exports/ 下的子目录。")
	}
	return clean, abs, nil
}

func serviceProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir = filepath.Clean(dir)
	for i := 0; i < 12; i++ {
		if fileExists(filepath.Join(dir, "go.mod")) && fileExists(filepath.Join(dir, "VERSION")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("读取项目根目录失败")
}

func writeExportFile(root string, rel string, raw []byte, written *[]string) error {
	rel = filepath.ToSlash(filepath.Clean(strings.TrimSpace(rel)))
	if rel == "" || rel == "." || strings.HasPrefix(rel, "../") || strings.Contains(rel, "/../") || filepath.IsAbs(rel) {
		return domain.NewPluginError("plugin_export_path_invalid", "导出文件路径不合法").
			WithStatus(400).
			WithDetail("path", rel).
			WithSuggestion("导出文件必须使用包内相对路径。")
	}
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0644); err != nil {
		return err
	}
	*written = append(*written, rel)
	return nil
}

func buildExportChecksums(root string, files []string) ([]map[string]string, error) {
	unique := uniqueStrings(files)
	sort.Strings(unique)
	out := make([]map[string]string, 0, len(unique))
	for _, rel := range unique {
		if rel == "checksums.json" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(raw)
		out = append(out, map[string]string{"path": rel, "sha256": hex.EncodeToString(sum[:])})
	}
	return out, nil
}

func pluginExportFileList(manifest domain.PluginManifest, req domain.PluginPackageExportRequest) []string {
	files := []string{"manifest.json", "README.md", "config.example.json"}
	if req.IncludeDocs {
		files = append(files, "docs/usage.md")
	}
	if req.IncludeMigrations && len(manifest.Migrations) > 0 {
		files = append(files, "migrations/exported_migrations.json")
	}
	if req.IncludePublisher {
		files = append(files, "publisher.json")
	}
	if req.IncludeSignatureStub {
		files = append(files, "signature.json")
	}
	files = append(files, "checksums.json")
	sort.Strings(files)
	return files
}

func buildPluginExportReadme(plugin domain.Plugin, manifest domain.PluginManifest) string {
	now := time.Now().UTC().Format(time.RFC3339)
	lines := []string{
		fmt.Sprintf("# %s", firstNonBlank(manifest.Name, manifest.Code)),
		"",
		fmt.Sprintf("- code: `%s`", manifest.Code),
		fmt.Sprintf("- version: `%s`", manifest.Version),
		fmt.Sprintf("- description: %s", firstNonBlank(manifest.Description, "暂无说明")),
		fmt.Sprintf("- exported_at: `%s`", now),
		fmt.Sprintf("- devhub_core_version: `%s`", currentCoreVersion()),
		fmt.Sprintf("- source_status: `%s`", plugin.Status),
		"",
		"## 内容类型",
		"",
		joinMarkdownList(manifest.ContentTypes),
		"",
		"## 权限摘要",
		"",
		joinPermissionList(manifest.Permissions),
		"",
		"## 配置说明",
		"",
		"`config.example.json` 仅为示例配置，不是当前环境配置备份；敏感字段使用 `REPLACE_ME` 占位。",
		"",
		"## 依赖说明",
		"",
		joinDependencyList(manifest.Dependencies),
		"",
		"## 安装方式",
		"",
		"1. 将目录复制到 DevHub 本地插件仓库，例如 `storage/plugins/packages/`。",
		"2. 在后台执行本地插件包 dry-run，检查 checksum、risk_report、依赖和 Core 兼容状态。",
		"3. 通过校验后再按审批/安装流程安装，安装后默认 disabled，需要手动配置并启用。",
		"",
		"## 安全边界",
		"",
		"- 本导出包不包含第三方运行时代码。",
		"- 本导出包不包含外部 SQL。",
		"- 本导出包不包含敏感配置明文或密文。",
		"- 本导出包不包含用户数据、通知数据、审计日志或 Hook 执行历史。",
		"",
	}
	return strings.Join(lines, "\n")
}

func buildPluginExportUsage(manifest domain.PluginManifest) string {
	return strings.Join([]string{
		fmt.Sprintf("# %s 使用说明", firstNonBlank(manifest.Name, manifest.Code)),
		"",
		"这是由 DevHub 后台导出的声明型插件包说明文件。",
		"",
		"导入前请先执行本地插件包 dry-run，确认 manifest、checksums、risk_report、依赖和 Core 兼容状态均符合预期。",
		"",
	}, "\n")
}

func joinMarkdownList(items []string) string {
	if len(items) == 0 {
		return "- 无"
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		out = append(out, fmt.Sprintf("- `%s`", strings.TrimSpace(item)))
	}
	if len(out) == 0 {
		return "- 无"
	}
	return strings.Join(out, "\n")
}

func joinPermissionList(items []domain.PermissionDefinition) string {
	if len(items) == 0 {
		return "- 无"
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("- `%s` %s", item.Code, strings.TrimSpace(item.Name)))
	}
	return strings.Join(out, "\n")
}

func joinDependencyList(items []domain.PluginDependency) string {
	if len(items) == 0 {
		return "- 无 required 依赖"
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		required := "required"
		if !item.Required {
			required = "optional"
		}
		out = append(out, fmt.Sprintf("- `%s` `%s` %s", item.Code, firstNonBlank(item.Version, "*"), required))
	}
	return strings.Join(out, "\n")
}

func buildConfigExample(schema any) any {
	m, ok := schema.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	props, _ := m["properties"].(map[string]any)
	out := map[string]any{}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		out[key] = exampleValueForSchema(props[key], key, key)
	}
	return out
}

func exampleValueForSchema(schema any, key string, path string) any {
	if pluginregistry.IsSensitiveField(schema, key, path) {
		return "REPLACE_ME"
	}
	m, _ := schema.(map[string]any)
	if v, ok := m["default"]; ok {
		return v
	}
	typ, _ := m["type"].(string)
	switch strings.TrimSpace(typ) {
	case "boolean":
		return false
	case "number", "integer":
		return 0
	case "array":
		return []any{}
	case "object":
		props, _ := m["properties"].(map[string]any)
		out := map[string]any{}
		keys := make([]string, 0, len(props))
		for k := range props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, child := range keys {
			out[child] = exampleValueForSchema(props[child], child, joinDot(path, child))
		}
		return out
	case "string":
		if enum, ok := m["enum"].([]any); ok && len(enum) > 0 {
			if s, ok := enum[0].(string); ok {
				return s
			}
		}
		return ""
	default:
		if enum, ok := m["enum"].([]any); ok && len(enum) > 0 {
			return enum[0]
		}
		return ""
	}
}

func buildPublisherStub(manifest domain.PluginManifest) map[string]any {
	return map[string]any{
		"publisher_id":         "unknown",
		"name":                 firstNonBlank(manifest.Author, "Unknown Publisher"),
		"homepage":             firstNonBlank(manifest.Homepage, ""),
		"email":                "",
		"public_key_id":        "",
		"public_key_algorithm": "ed25519",
		"public_key":           "",
		"trust_level":          "unknown",
	}
}

func buildSignatureStub(manifest domain.PluginManifest, files []string) map[string]any {
	signed := []string{"manifest.json", "checksums.json"}
	for _, rel := range files {
		if rel == "signature.json" || rel == "checksums.json" {
			continue
		}
		if rel == "manifest.json" || rel == "README.md" || rel == "config.example.json" {
			signed = append(signed, rel)
		}
	}
	signed = uniqueStrings(signed)
	sort.Strings(signed)
	return map[string]any{
		"version":       "1",
		"algorithm":     "ed25519",
		"signed_at":     time.Now().UTC().Format(time.RFC3339),
		"publisher_id":  "unknown",
		"public_key_id": "",
		"signed_files":  signed,
		"signature":     "",
		"note":          fmt.Sprintf("signature stub for %s; not a verified signature", manifest.Code),
	}
}

func safeExportPart(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = pluginExportSafePartPattern.ReplaceAllString(s, "_")
	s = strings.Trim(s, "._-")
	if s == "" {
		return "plugin"
	}
	return s
}

func joinDot(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func relativeToRoot(root string, abs string) string {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return filepath.ToSlash(abs)
	}
	return filepath.ToSlash(rel)
}
