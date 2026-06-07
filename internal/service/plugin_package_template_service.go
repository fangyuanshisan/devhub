package service

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/plugins/scaffold"
)

const pluginPackageTemplateRoot = "storage/plugins/drafts"

var pluginTemplateReservedCodes = map[string]bool{
	"core":                    true,
	"qa":                      true,
	"docs":                    true,
	"wiki":                    true,
	"official_links":          true,
	"official_announcement":   true,
	"official_webhook_notify": true,
	"official-links":          true,
	"official-announcement":   true,
	"official-webhook-notify": true,
}

func (s *Service) PreviewPluginPackageTemplate(req domain.PluginPackageTemplateRequest) (domain.PluginPackageTemplatePreviewResponse, error) {
	preview, err := s.previewPluginPackageTemplate(req)
	if err != nil {
		return domain.PluginPackageTemplatePreviewResponse{}, err
	}
	return domain.PluginPackageTemplatePreviewResponse{
		Template: preview,
		Status:   "ok",
		Warnings: []string{
			"预览不会写入文件；正式初始化会固定写入 storage/plugins/drafts/{code}。",
			"后台初始化模板不会生成 registry.example.go，registry 接入说明改写入 docs/registry-example.md。",
			"初始化只生成声明型模板，不安装、不启用、不执行 SQL、不执行第三方代码。",
		},
	}, nil
}

func (s *Service) CreatePluginPackageTemplate(req domain.PluginPackageTemplateRequest) (domain.PluginPackageTemplateCreateResponse, error) {
	preview, err := s.previewPluginPackageTemplate(req)
	if err != nil {
		return domain.PluginPackageTemplateCreateResponse{}, err
	}
	if len(preview.Conflicts) > 0 {
		return domain.PluginPackageTemplateCreateResponse{}, pluginTemplateConflictError(preview.Conflicts, preview.PackagePath)
	}
	root, err := serviceProjectRoot()
	if err != nil {
		return domain.PluginPackageTemplateCreateResponse{}, err
	}
	outputRoot := filepath.Join(root, pluginPackageTemplateRoot)

	result, err := scaffold.Generate(scaffold.Options{
		Code:               preview.Code,
		Name:               preview.Name,
		PluginType:         preview.PluginType,
		ContentType:        preview.ContentType,
		ContentName:        preview.ContentName,
		Description:        preview.Description,
		Author:             preview.Author,
		MountPoint:         strings.TrimSpace(req.MountPoint),
		ComponentKey:       strings.TrimSpace(req.ComponentKey),
		HealthCheckPath:    strings.TrimSpace(req.HealthCheckPath),
		TimeoutMS:          req.TimeoutMS,
		FailurePolicy:      strings.TrimSpace(req.FailurePolicy),
		Output:             outputRoot,
		WithConfig:         req.WithConfig,
		WithHooks:          req.WithHooks,
		WithMigration:      req.WithMigration,
		IncludeRegistryDoc: true,
		AllowExistingEmpty: true,
		Force:              false,
	})
	if err != nil {
		return domain.PluginPackageTemplateCreateResponse{}, pluginTemplateError(err, preview.PackagePath)
	}

	relPath := filepath.ToSlash(filepath.Join(pluginPackageTemplateRoot, result.Manifest.Code))
	dry, err := s.DryRunPluginPackage(relPath)
	if err != nil {
		return domain.PluginPackageTemplateCreateResponse{}, err
	}

	warnings := append([]string{}, dry.Warnings...)
	if strings.ToLower(strings.TrimSpace(dry.Status)) == "blocked" {
		warnings = append(warnings, "初始化后 package dry-run 被阻断，请按风险报告修复后再提交安装审批。")
	}
	return domain.PluginPackageTemplateCreateResponse{
		Message:  "插件包模板已初始化，并已完成 package dry-run",
		Template: preview,
		DryRun:   dry,
		Status:   dry.Status,
		Warnings: uniqueStrings(warnings),
		Errors:   dry.Errors,
	}, nil
}

func (s *Service) ExportPluginPackageTemplateZip(req domain.PluginPackageTemplateRequest) ([]byte, string, domain.PluginPackageTemplatePreview, error) {
	preview, err := s.previewPluginPackageTemplate(req)
	if err != nil {
		return nil, "", domain.PluginPackageTemplatePreview{}, err
	}
	if len(preview.Conflicts) > 0 {
		return nil, "", domain.PluginPackageTemplatePreview{}, pluginTemplateConflictError(preview.Conflicts, preview.PackagePath)
	}
	tmp, err := os.MkdirTemp("", "devhub-plugin-template-*")
	if err != nil {
		return nil, "", domain.PluginPackageTemplatePreview{}, fmt.Errorf("创建临时导出目录失败：%w", err)
	}
	defer os.RemoveAll(tmp)
	result, err := scaffold.Generate(scaffold.Options{
		Code:               preview.Code,
		Name:               preview.Name,
		PluginType:         preview.PluginType,
		ContentType:        preview.ContentType,
		ContentName:        preview.ContentName,
		Description:        preview.Description,
		Author:             preview.Author,
		MountPoint:         strings.TrimSpace(req.MountPoint),
		ComponentKey:       strings.TrimSpace(req.ComponentKey),
		HealthCheckPath:    strings.TrimSpace(req.HealthCheckPath),
		TimeoutMS:          req.TimeoutMS,
		FailurePolicy:      strings.TrimSpace(req.FailurePolicy),
		Output:             tmp,
		WithConfig:         req.WithConfig,
		WithHooks:          req.WithHooks,
		WithMigration:      req.WithMigration,
		IncludeRegistryDoc: true,
		AllowExistingEmpty: true,
	})
	if err != nil {
		return nil, "", domain.PluginPackageTemplatePreview{}, pluginTemplateError(err, preview.PackagePath)
	}
	data, err := zipDirectory(result.Dir, preview.Code)
	if err != nil {
		return nil, "", domain.PluginPackageTemplatePreview{}, err
	}
	return data, preview.Code + ".zip", preview, nil
}

func (s *Service) previewPluginPackageTemplate(req domain.PluginPackageTemplateRequest) (domain.PluginPackageTemplatePreview, error) {
	root, err := serviceProjectRoot()
	if err != nil {
		return domain.PluginPackageTemplatePreview{}, err
	}
	outputRoot := filepath.Join(root, pluginPackageTemplateRoot)
	req = normalizePluginTemplateRequest(req)
	res, err := scaffold.Preview(scaffold.Options{
		Code:               req.Code,
		Name:               req.Name,
		PluginType:         req.PluginType,
		ContentType:        req.ContentType,
		ContentName:        req.ContentName,
		Description:        req.Description,
		Author:             req.Author,
		MountPoint:         req.MountPoint,
		ComponentKey:       req.ComponentKey,
		HealthCheckPath:    req.HealthCheckPath,
		TimeoutMS:          req.TimeoutMS,
		FailurePolicy:      req.FailurePolicy,
		Output:             outputRoot,
		WithConfig:         req.WithConfig,
		WithHooks:          req.WithHooks,
		WithMigration:      req.WithMigration,
		IncludeRegistryDoc: true,
		AllowExistingEmpty: true,
		Force:              false,
	})
	if err != nil {
		return domain.PluginPackageTemplatePreview{}, pluginTemplateError(err, "")
	}
	code := strings.TrimSpace(res.Manifest.Code)
	packagePath := filepath.ToSlash(filepath.Join(pluginPackageTemplateRoot, code))
	if _, _, err := pluginregistry.NormalizePluginPackagePath(packagePath); err != nil {
		return domain.PluginPackageTemplatePreview{}, err
	}
	files := make([]string, 0, len(res.Files))
	for _, file := range res.Files {
		rel, err := filepath.Rel(res.Dir, file)
		if err != nil {
			rel = filepath.Base(file)
		}
		files = append(files, filepath.ToSlash(rel))
	}
	sort.Strings(files)
	conflicts := s.pluginTemplateConflicts(res.Manifest, packagePath)
	return domain.PluginPackageTemplatePreview{
		Code:            code,
		Name:            strings.TrimSpace(res.Manifest.Name),
		PluginType:      req.PluginType,
		ContentType:     firstManifestContentType(res.Manifest.ContentTypes),
		ContentName:     firstManifestContentName(res.Manifest.ContentTypeDefs),
		Description:     strings.TrimSpace(res.Manifest.Description),
		Author:          strings.TrimSpace(res.Manifest.Author),
		OutputDir:       packagePath,
		PackagePath:     packagePath,
		Files:           files,
		FileTree:        files,
		Permissions:     manifestPermissionCodes(res.Manifest),
		Menus:           manifestMenuCodes(res.Manifest),
		Hooks:           manifestHookNames(res.Manifest),
		Migrations:      manifestMigrationFiles(files),
		FrontendMounts:  res.Manifest.FrontendMounts,
		ExternalService: res.Manifest.ExternalService,
		Generated: map[string]string{
			"code":         code,
			"content_type": firstManifestContentType(res.Manifest.ContentTypes),
			"content_name": firstManifestContentName(res.Manifest.ContentTypeDefs),
			"author":       strings.TrimSpace(res.Manifest.Author),
		},
		Summary: map[string]any{
			"plugin_type":      req.PluginType,
			"permissions":      len(res.Manifest.Permissions),
			"menus":            len(res.Manifest.Menus),
			"hooks":            len(res.Manifest.Hooks),
			"migrations":       len(manifestMigrationFiles(files)),
			"frontend_mounts":  len(res.Manifest.FrontendMounts),
			"external_service": res.Manifest.ExternalService != nil,
			"file_count":       len(files),
		},
		Conflicts:     conflicts,
		WillOverwrite: false,
	}, nil
}

func normalizePluginTemplateRequest(req domain.PluginPackageTemplateRequest) domain.PluginPackageTemplateRequest {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" && req.Name != "" {
		req.Code = scaffold.SlugifyCode(req.Name)
	}
	req.PluginType = normalizeServicePluginType(req.PluginType)
	req.ContentType = strings.TrimSpace(req.ContentType)
	req.ContentName = strings.TrimSpace(req.ContentName)
	if req.PluginType != scaffold.PluginTypeContent {
		req.ContentType = ""
		req.ContentName = ""
	}
	req.Description = strings.TrimSpace(req.Description)
	req.Author = strings.TrimSpace(req.Author)
	if req.Author == "" {
		req.Author = "DevHub Team"
	}
	if req.WithConfig == false && req.WithHooks == false && req.WithMigration == false {
		req.WithConfig = true
		req.WithHooks = req.PluginType == scaffold.PluginTypeContent || req.PluginType == scaffold.PluginTypeExternalService
		req.WithMigration = req.PluginType == scaffold.PluginTypeContent
	}
	return req
}

func normalizeServicePluginType(value string) string {
	return scaffold.NormalizePluginType(value)
}

func (s *Service) pluginTemplateConflicts(manifest domain.PluginManifest, packagePath string) []domain.PluginPackageTemplateConflict {
	conflicts := []domain.PluginPackageTemplateConflict{}
	code := strings.TrimSpace(manifest.Code)
	if pluginTemplateReservedCodes[code] {
		conflicts = append(conflicts, domain.PluginPackageTemplateConflict{Field: "code", Value: code, Target: "reserved", Message: "该插件编码为系统保留值，不能使用。", Suggestion: "请换一个业务插件编码。"})
	}
	if _, ok := s.repo.PluginByCode(code); ok {
		conflicts = append(conflicts, domain.PluginPackageTemplateConflict{Field: "code", Value: code, Target: "installed_plugin", Message: "当前插件编码已存在，请更换。", Suggestion: "请使用未安装过的插件编码。"})
	}
	for _, rel := range []string{packagePath, filepath.ToSlash(filepath.Join("storage/plugins/packages", code))} {
		if abs, _, err := pluginregistry.NormalizePluginPackagePath(rel); err == nil {
			if info, statErr := os.Stat(abs); statErr == nil && info.IsDir() && !dirIsEmptyForTemplate(abs) {
				conflicts = append(conflicts, domain.PluginPackageTemplateConflict{Field: "draft_directory", Value: rel, Target: "filesystem", Message: "草稿或本地仓库目录已存在。", Suggestion: "请清理已有目录，或更换插件编码。"})
			}
		}
	}
	knownTypes := map[string]string{}
	for _, plugin := range s.repo.Plugins() {
		for _, typ := range plugin.ContentTypes {
			knownTypes[strings.TrimSpace(typ)] = plugin.Code
		}
		for _, def := range plugin.ContentTypeDefs {
			knownTypes[strings.TrimSpace(def.Type)] = plugin.Code
		}
	}
	for _, typ := range manifest.ContentTypes {
		if owner := strings.TrimSpace(knownTypes[strings.TrimSpace(typ)]); owner != "" && owner != code {
			conflicts = append(conflicts, domain.PluginPackageTemplateConflict{Field: "content_type", Value: typ, Target: owner, Message: "内容数据类型已被其他插件使用。", Suggestion: "请重新生成或手动修改内容数据类型。"})
		}
	}
	seen := map[string]string{}
	for _, permission := range manifest.Permissions {
		if prev := seen[permission.Code]; prev != "" {
			conflicts = append(conflicts, domain.PluginPackageTemplateConflict{Field: "permission", Value: permission.Code, Target: prev, Message: "模板内权限码重复。", Suggestion: "请修改插件编码或权限声明。"})
		}
		seen[permission.Code] = permission.PluginCode
	}
	menuSeen := map[string]bool{}
	for _, menu := range manifest.Menus {
		if menuSeen[menu.Code] {
			conflicts = append(conflicts, domain.PluginPackageTemplateConflict{Field: "menu", Value: menu.Code, Message: "模板内菜单 key 重复。", Suggestion: "请修改插件编码或菜单声明。"})
		}
		menuSeen[menu.Code] = true
	}
	routeSeen := map[string]bool{}
	for _, route := range manifest.Routes {
		key := strings.Join([]string{route.Area, route.Method, route.Path}, " ")
		if routeSeen[key] {
			conflicts = append(conflicts, domain.PluginPackageTemplateConflict{Field: "route", Value: key, Message: "模板内路由 key 重复。", Suggestion: "请修改路由声明。"})
		}
		routeSeen[key] = true
	}
	mountSeen := map[string]bool{}
	for _, mount := range manifest.FrontendMounts {
		key := strings.Join([]string{mount.MountPoint, mount.ComponentKey}, " ")
		if mountSeen[key] {
			conflicts = append(conflicts, domain.PluginPackageTemplateConflict{Field: "frontend_mount", Value: key, Message: "模板内前端挂载重复。", Suggestion: "请修改挂载点或组件 key。"})
		}
		mountSeen[key] = true
	}
	return conflicts
}

func dirIsEmptyForTemplate(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) == 0
}

func manifestPermissionCodes(manifest domain.PluginManifest) []string {
	out := make([]string, 0, len(manifest.Permissions))
	for _, item := range manifest.Permissions {
		out = append(out, strings.TrimSpace(item.Code))
	}
	sort.Strings(out)
	return out
}

func manifestMenuCodes(manifest domain.PluginManifest) []string {
	out := make([]string, 0, len(manifest.Menus))
	for _, item := range manifest.Menus {
		out = append(out, strings.TrimSpace(item.Code))
	}
	sort.Strings(out)
	return out
}

func manifestHookNames(manifest domain.PluginManifest) []string {
	out := make([]string, 0, len(manifest.Hooks))
	for _, item := range manifest.Hooks {
		name := strings.TrimSpace(item.Name)
		if item.ServiceType != "" {
			name += " / " + strings.TrimSpace(item.ServiceType)
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func manifestMigrationFiles(files []string) []string {
	out := []string{}
	for _, file := range files {
		if strings.HasPrefix(file, "migrations/") && strings.HasSuffix(strings.ToLower(file), ".sql") {
			out = append(out, file)
		}
	}
	sort.Strings(out)
	return out
}

func zipDirectory(dir, rootName string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(filepath.Join(rootName, rel))
		if strings.Contains(name, "../") || strings.HasPrefix(name, "/") {
			return fmt.Errorf("导出路径不安全：%s", name)
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = name
		header.Method = zip.Deflate
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		return err
	})
	if err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func pluginTemplateConflictError(conflicts []domain.PluginPackageTemplateConflict, path string) error {
	hash := sha1.Sum([]byte(fmt.Sprintf("%v", conflicts)))
	return domain.NewPluginError("plugin_package_template_conflict", "模板命名存在冲突").
		WithStatus(409).
		WithDetail("conflicts", conflicts).
		WithDetail("conflict_hash", hex.EncodeToString(hash[:6])).
		WithDetail("path", path).
		WithSuggestion("请根据 conflicts 提示修改插件编码、内容数据类型或模板字段后重试。")
}

func pluginTemplateError(err error, path string) error {
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(err.Error())
	code := "plugin_package_template_invalid"
	message := "插件包模板初始化参数无效"
	suggestion := "请检查 code/content_type 是否符合小写字母、数字、下划线规则，并确认 storage/plugins/drafts/{code} 尚不存在。"
	if strings.Contains(msg, "输出目录已存在") || strings.Contains(msg, "exists") {
		code = "plugin_package_template_exists"
		message = "插件包模板目标目录已存在"
		suggestion = "请先删除或清理 storage/plugins/drafts/{code} 下的已有文件后重试；空目录会自动复用。"
	}
	apiErr := domain.NewPluginError(code, message).
		WithStatus(400).
		WithDetail("reason", msg).
		WithSuggestion(suggestion)
	if strings.TrimSpace(path) != "" {
		apiErr.WithDetail("path", path)
	}
	return apiErr
}

func firstManifestContentType(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return strings.TrimSpace(items[0])
}

func firstManifestContentName(items []domain.ContentTypeDefinition) string {
	if len(items) == 0 {
		return ""
	}
	return strings.TrimSpace(items[0].Name)
}
