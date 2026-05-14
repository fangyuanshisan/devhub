package service

import (
	"path/filepath"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/plugins/scaffold"
)

const pluginPackageTemplateRoot = "storage/plugins/packages"

func (s *Service) PreviewPluginPackageTemplate(req domain.PluginPackageTemplateRequest) (domain.PluginPackageTemplatePreviewResponse, error) {
	preview, err := previewPluginPackageTemplate(req)
	if err != nil {
		return domain.PluginPackageTemplatePreviewResponse{}, err
	}
	return domain.PluginPackageTemplatePreviewResponse{
		Template: preview,
		Status:   "ok",
		Warnings: []string{
			"预览不会写入文件；正式初始化会固定写入 storage/plugins/packages/{code}。",
			"后台初始化模板不会生成 registry.example.go，registry 接入说明改写入 docs/registry-example.md。",
		},
	}, nil
}

func (s *Service) CreatePluginPackageTemplate(req domain.PluginPackageTemplateRequest) (domain.PluginPackageTemplateCreateResponse, error) {
	preview, err := previewPluginPackageTemplate(req)
	if err != nil {
		return domain.PluginPackageTemplateCreateResponse{}, err
	}
	root, err := serviceProjectRoot()
	if err != nil {
		return domain.PluginPackageTemplateCreateResponse{}, err
	}
	outputRoot := filepath.Join(root, pluginPackageTemplateRoot)

	result, err := scaffold.Generate(scaffold.Options{
		Code:               req.Code,
		Name:               req.Name,
		ContentType:        req.ContentType,
		ContentName:        req.ContentName,
		Description:        req.Description,
		Author:             req.Author,
		Output:             outputRoot,
		WithConfig:         req.WithConfig,
		WithHooks:          req.WithHooks,
		WithMigration:      req.WithMigration,
		IncludeRegistryDoc: true,
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

func previewPluginPackageTemplate(req domain.PluginPackageTemplateRequest) (domain.PluginPackageTemplatePreview, error) {
	root, err := serviceProjectRoot()
	if err != nil {
		return domain.PluginPackageTemplatePreview{}, err
	}
	outputRoot := filepath.Join(root, pluginPackageTemplateRoot)
	res, err := scaffold.Preview(scaffold.Options{
		Code:               req.Code,
		Name:               req.Name,
		ContentType:        req.ContentType,
		ContentName:        req.ContentName,
		Description:        req.Description,
		Author:             req.Author,
		Output:             outputRoot,
		WithConfig:         req.WithConfig,
		WithHooks:          req.WithHooks,
		WithMigration:      req.WithMigration,
		IncludeRegistryDoc: true,
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
	return domain.PluginPackageTemplatePreview{
		Code:          code,
		Name:          strings.TrimSpace(res.Manifest.Name),
		ContentType:   firstManifestContentType(res.Manifest.ContentTypes),
		ContentName:   firstManifestContentName(res.Manifest.ContentTypeDefs),
		Description:   strings.TrimSpace(res.Manifest.Description),
		Author:        strings.TrimSpace(res.Manifest.Author),
		OutputDir:     packagePath,
		PackagePath:   packagePath,
		Files:         files,
		WillOverwrite: false,
	}, nil
}

func pluginTemplateError(err error, path string) error {
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(err.Error())
	code := "plugin_package_template_invalid"
	if strings.Contains(msg, "输出目录已存在") || strings.Contains(msg, "exists") {
		code = "plugin_package_template_exists"
	}
	apiErr := domain.NewPluginError(code, "插件包模板初始化参数无效").
		WithStatus(400).
		WithDetail("reason", msg).
		WithSuggestion("请检查 code/content_type 是否符合小写字母、数字、下划线规则，并确认 storage/plugins/packages/{code} 尚不存在。")
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
