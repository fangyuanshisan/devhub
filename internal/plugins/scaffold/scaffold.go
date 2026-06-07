package scaffold

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

var (
	codePattern        = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)
	contentTypePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
)

const (
	PluginTypeContent         = "content"
	PluginTypeExternalService = "external_service"
	PluginTypeAdminTool       = "admin_tool"
	PluginTypeFrontendMount   = "frontend_mount"
)

type PluginTemplateGenerator struct{}

type Options struct {
	Code               string
	Name               string
	PluginType         string
	ContentType        string
	ContentName        string
	Description        string
	Author             string
	MountPoint         string
	ComponentKey       string
	HealthCheckPath    string
	TimeoutMS          int
	FailurePolicy      string
	Output             string
	WithConfig         bool
	WithHooks          bool
	WithMigration      bool
	IncludeRegistryDoc bool
	AllowExistingEmpty bool
	Force              bool
}

type Result struct {
	Dir      string
	Files    []string
	Manifest domain.PluginManifest
}

type PreviewResult struct {
	Dir      string
	Files    []string
	Manifest domain.PluginManifest
}

func Generate(opts Options) (Result, error) {
	return PluginTemplateGenerator{}.Generate(opts)
}

func (PluginTemplateGenerator) Generate(opts Options) (Result, error) {
	preview, files, err := prepare(opts)
	if err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(preview.Dir, 0o755); err != nil {
		return Result{}, err
	}

	written := make([]string, 0, len(files))
	for name, data := range files {
		path := filepath.Join(preview.Dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Result{}, err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return Result{}, err
		}
		written = append(written, path)
	}
	sort.Strings(written)
	return Result{Dir: preview.Dir, Files: written, Manifest: preview.Manifest}, nil
}

func Preview(opts Options) (PreviewResult, error) {
	return PluginTemplateGenerator{}.Preview(opts)
}

func (PluginTemplateGenerator) Preview(opts Options) (PreviewResult, error) {
	preview, _, err := prepare(opts)
	if err != nil {
		return PreviewResult{}, err
	}
	return preview, nil
}

func marshalManifest(manifest domain.PluginManifest) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Manifest(opts Options) domain.PluginManifest {
	opts = normalizeOptions(opts)
	switch opts.PluginType {
	case PluginTypeExternalService:
		return externalServiceManifest(opts)
	case PluginTypeAdminTool:
		return adminToolManifest(opts)
	case PluginTypeFrontendMount:
		return frontendMountManifest(opts)
	default:
		return contentManifest(opts)
	}
}

func contentManifest(opts Options) domain.PluginManifest {
	resource := permissionResource(opts.ContentType)
	permissions := []domain.PermissionDefinition{
		{Code: fmt.Sprintf("%s.%s.create", opts.Code, resource), Name: fmt.Sprintf("创建%s", opts.ContentName), Description: fmt.Sprintf("允许创建%s内容。", opts.ContentName), Scope: "community", PluginCode: opts.Code},
		{Code: fmt.Sprintf("%s.%s.edit", opts.Code, resource), Name: fmt.Sprintf("编辑%s", opts.ContentName), Description: fmt.Sprintf("允许编辑%s内容。", opts.ContentName), Scope: "own", PluginCode: opts.Code},
		{Code: fmt.Sprintf("%s.%s.delete", opts.Code, resource), Name: fmt.Sprintf("删除%s", opts.ContentName), Description: fmt.Sprintf("允许删除%s内容。", opts.ContentName), Scope: "own", PluginCode: opts.Code},
		{Code: fmt.Sprintf("%s.%s.audit", opts.Code, resource), Name: fmt.Sprintf("审核%s", opts.ContentName), Description: fmt.Sprintf("允许审核%s内容。", opts.ContentName), Scope: "community", PluginCode: opts.Code},
		{Code: fmt.Sprintf("%s.manage", opts.Code), Name: fmt.Sprintf("管理%s", opts.Name), Description: fmt.Sprintf("允许管理%s插件内容和治理入口。", opts.Name), Scope: "global", PluginCode: opts.Code},
		{Code: fmt.Sprintf("%s.configure", opts.Code), Name: fmt.Sprintf("配置%s", opts.Name), Description: fmt.Sprintf("允许配置%s插件。", opts.Name), Scope: "global", PluginCode: opts.Code},
	}
	manifest := domain.PluginManifest{
		Code:                  opts.Code,
		Name:                  opts.Name,
		Version:               "1.0.0",
		Description:           opts.Description,
		Author:                opts.Author,
		CompatibleCoreVersion: ">=1.4.0",
		MinCoreVersion:        "1.4.0",
		IsSystem:              false,
		ContentTypes:          []string{opts.ContentType},
		ContentTypeDefs: []domain.ContentTypeDefinition{{
			Type:             opts.ContentType,
			Name:             opts.ContentName,
			PluginCode:       opts.Code,
			CreatePermission: fmt.Sprintf("%s.%s.create", opts.Code, resource),
			EditPermission:   fmt.Sprintf("%s.%s.edit", opts.Code, resource),
			DeletePermission: fmt.Sprintf("%s.%s.delete", opts.Code, resource),
			AuditPermission:  fmt.Sprintf("%s.%s.audit", opts.Code, resource),
			DefaultStatus:    "draft",
			AllowComment:     true,
			AllowLike:        true,
			AllowFavorite:    true,
			SEOType:          "Article",
		}},
		Permissions: permissions,
		Menus: []domain.MenuDefinition{
			{Code: opts.Code + ".admin", Title: opts.Name, Path: "/admin-next/" + opts.Code, Area: "admin", Location: "admin", Permission: opts.Code + ".manage", PluginCode: opts.Code, SortOrder: 300},
			{Code: opts.Code + ".moderator", Title: opts.Name, Path: "/moderator/" + opts.Code, Area: "moderator", Location: "moderator", Permission: fmt.Sprintf("%s.%s.audit", opts.Code, resource), PluginCode: opts.Code, SortOrder: 300},
			{Code: opts.Code + ".frontend", Title: opts.Name, Path: "/" + opts.Code, Area: "frontend", Location: "frontend", Permission: fmt.Sprintf("%s.%s.create", opts.Code, resource), PluginCode: opts.Code, SortOrder: 300},
		},
		Routes: []domain.RouteDefinition{
			{Area: "admin", Method: "GET", Path: "/api/v1/admin/" + opts.Code, Handler: "reserved:" + opts.Code + ".admin.list", Auth: "admin", Permission: opts.Code + ".manage", PluginCode: opts.Code},
			{Area: "frontend", Method: "GET", Path: "/" + opts.Code, Handler: "reserved:" + opts.Code + ".frontend.index", Auth: "optional", Permission: fmt.Sprintf("%s.%s.create", opts.Code, resource), PluginCode: opts.Code},
		},
		Dependencies: []domain.PluginDependency{},
		ConfigSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           map[string]any{},
		},
		Hooks:      []domain.HookDefinition{},
		Migrations: []domain.PluginMigrationDefinition{},
		Assets:     []string{},
	}
	if opts.WithConfig {
		manifest.ConfigSchema = map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []any{"enabled", "items_per_page"},
			"properties": map[string]any{
				"enabled": map[string]any{
					"type":        "boolean",
					"title":       "启用扩展功能",
					"description": "控制插件声明功能是否在配置层启用。",
					"default":     true,
				},
				"items_per_page": map[string]any{
					"type":        "integer",
					"title":       "每页数量",
					"description": "列表类页面每页展示数量。",
					"min":         float64(1),
					"max":         float64(100),
					"default":     float64(20),
				},
				"display_mode": map[string]any{
					"type":        "string",
					"title":       "展示模式",
					"description": "示例枚举配置。",
					"enum":        []any{"compact", "standard"},
					"default":     "standard",
				},
			},
		}
	}
	if opts.WithHooks {
		manifest.Hooks = []domain.HookDefinition{
			{Name: pluginregistry.HookAfterCreateContent, PluginCode: opts.Code, Description: "声明创建内容后的 non-blocking Hook；当前不会动态加载第三方处理器。", Mode: string(pluginregistry.HookNonBlocking), Blocking: false, Critical: false, FailurePolicy: "log", TimeoutMS: 3000, FailureThreshold: 3},
		}
	}
	if opts.WithMigration {
		manifest.Migrations = []domain.PluginMigrationDefinition{{
			PluginCode:        opts.Code,
			MigrationVersion:  "1.0.0",
			MigrationName:     opts.Code + "_init",
			Direction:         "up",
			Checksum:          "sha256:declaration-only",
			Tables:            []string{opts.Code + "_items"},
			RollbackSupported: false,
			Description:       "声明型迁移示例。当前不会执行外部 raw SQL。",
		}}
	}
	return manifest
}

func externalServiceManifest(opts Options) domain.PluginManifest {
	enabled := true
	timeout := opts.TimeoutMS
	if timeout <= 0 {
		timeout = 3000
	}
	failurePolicy := strings.TrimSpace(opts.FailurePolicy)
	if failurePolicy == "" {
		failurePolicy = "warn"
	}
	hookPath := strings.TrimSpace(opts.HealthCheckPath)
	if hookPath == "" {
		hookPath = "/hooks/content.after_create"
	}
	permissions := baseManagePermissions(opts)
	manifest := baseManifest(opts, permissions)
	manifest.Menus = []domain.MenuDefinition{
		{Code: opts.Code + ".external_service", Title: opts.Name, Path: "/admin-next/plugins/" + opts.Code + "/external-service", Area: "admin", Location: "admin", Permission: opts.Code + ".manage", PluginCode: opts.Code, SortOrder: 320},
	}
	manifest.ExternalService = &domain.PluginExternalService{
		Endpoint:           "https://example.com",
		SubscribedHooks:    []string{pluginregistry.HookAfterCreateContent},
		TimeoutMS:          timeout,
		FailurePolicy:      failurePolicy,
		RetryPolicy:        "linear",
		SignatureAlgorithm: "hmac-sha256",
	}
	if opts.WithHooks {
		manifest.Hooks = []domain.HookDefinition{{
			Name:          pluginregistry.HookAfterCreateContent,
			PluginCode:    opts.Code,
			Description:   "声明 external_service non-blocking Hook；当前不会执行本地第三方代码。",
			Mode:          string(pluginregistry.HookNonBlocking),
			ServiceType:   "external_service",
			Path:          hookPath,
			Method:        "POST",
			RetryEnabled:  true,
			MaxAttempts:   3,
			Enabled:       &enabled,
			Blocking:      false,
			Critical:      false,
			FailurePolicy: failurePolicy,
			TimeoutMS:     timeout,
		}}
	}
	if opts.WithConfig {
		manifest.ConfigSchema = externalServiceConfigSchema(timeout, failurePolicy)
	}
	return manifest
}

func adminToolManifest(opts Options) domain.PluginManifest {
	permissions := baseManagePermissions(opts)
	manifest := baseManifest(opts, permissions)
	manifest.Menus = []domain.MenuDefinition{
		{Code: opts.Code + ".admin", Title: opts.Name, Path: "/admin-next/tools/" + opts.Code, Area: "admin", Location: "admin", Permission: opts.Code + ".manage", PluginCode: opts.Code, SortOrder: 330},
	}
	if opts.WithMigration {
		manifest.Migrations = []domain.PluginMigrationDefinition{migrationDefinition(opts)}
	}
	return manifest
}

func frontendMountManifest(opts Options) domain.PluginManifest {
	permissions := baseManagePermissions(opts)
	manifest := baseManifest(opts, permissions)
	manifest.FrontendMounts = []domain.FrontendMountDefinition{{
		PluginCode:   opts.Code,
		MountPoint:   opts.MountPoint,
		ComponentKey: opts.ComponentKey,
		RenderMode:   pluginregistry.FrontendRenderModeOfficialComponent,
		ConfigRef:    "resolved_config",
		PropsSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"title": map[string]any{"type": "string", "default": opts.Name},
			},
		},
	}}
	manifest.Menus = []domain.MenuDefinition{
		{Code: opts.Code + ".mounts", Title: opts.Name, Path: "/admin-next/plugins/" + opts.Code + "/mounts", Area: "admin", Location: "admin", Permission: opts.Code + ".manage", PluginCode: opts.Code, SortOrder: 340},
	}
	return manifest
}

func baseManifest(opts Options, permissions []domain.PermissionDefinition) domain.PluginManifest {
	manifest := domain.PluginManifest{
		Code:                  opts.Code,
		Name:                  opts.Name,
		Version:               "1.0.0",
		Description:           opts.Description,
		Author:                opts.Author,
		CompatibleCoreVersion: ">=1.4.0",
		MinCoreVersion:        "1.4.0",
		IsSystem:              false,
		ContentTypes:          []string{},
		ContentTypeDefs:       []domain.ContentTypeDefinition{},
		Permissions:           permissions,
		Menus:                 []domain.MenuDefinition{},
		Routes:                []domain.RouteDefinition{},
		Dependencies:          []domain.PluginDependency{},
		ConfigSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           map[string]any{},
		},
		Hooks:      []domain.HookDefinition{},
		Migrations: []domain.PluginMigrationDefinition{},
		Assets:     []string{},
	}
	if opts.WithConfig {
		manifest.ConfigSchema = map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []any{"enabled"},
			"properties": map[string]any{
				"enabled": map[string]any{
					"type":        "boolean",
					"title":       "启用插件声明",
					"description": "控制插件声明能力是否在配置层启用。",
					"default":     true,
				},
			},
		}
	}
	return manifest
}

func baseManagePermissions(opts Options) []domain.PermissionDefinition {
	return []domain.PermissionDefinition{
		{Code: fmt.Sprintf("%s.manage", opts.Code), Name: fmt.Sprintf("管理%s", opts.Name), Description: fmt.Sprintf("允许管理%s插件。", opts.Name), Scope: "global", PluginCode: opts.Code},
		{Code: fmt.Sprintf("%s.configure", opts.Code), Name: fmt.Sprintf("配置%s", opts.Name), Description: fmt.Sprintf("允许配置%s插件。", opts.Name), Scope: "global", PluginCode: opts.Code},
		{Code: fmt.Sprintf("%s.menu.view", opts.Code), Name: fmt.Sprintf("访问%s菜单", opts.Name), Description: fmt.Sprintf("允许访问%s后台入口。", opts.Name), Scope: "global", PluginCode: opts.Code},
	}
}

func migrationDefinition(opts Options) domain.PluginMigrationDefinition {
	return domain.PluginMigrationDefinition{
		PluginCode:        opts.Code,
		MigrationVersion:  "1.0.0",
		MigrationName:     opts.Code + "_init",
		Direction:         "up",
		Checksum:          "sha256:declaration-only",
		Tables:            []string{strings.ReplaceAll(opts.Code, "-", "_") + "_items"},
		RollbackSupported: false,
		Description:       "声明型迁移示例。当前不会执行外部 raw SQL。",
	}
}

func externalServiceConfigSchema(timeout int, failurePolicy string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []any{"enabled", "endpoint_url"},
		"properties": map[string]any{
			"enabled":           map[string]any{"type": "boolean", "title": "启用外部服务", "default": false},
			"endpoint_url":      map[string]any{"type": "string", "title": "服务地址", "format": "uri", "default": "https://example.com"},
			"health_check_path": map[string]any{"type": "string", "title": "健康检查路径", "pattern": "^/(?:[^/\\s][^\\s]*)?$", "default": "/health"},
			"timeout_ms":        map[string]any{"type": "integer", "title": "超时时间", "min": float64(100), "max": float64(10000), "default": float64(timeout)},
			"failure_policy":    map[string]any{"type": "string", "title": "失败策略", "enum": []any{"warn", "log", "ignore"}, "default": failurePolicy},
		},
	}
}

func normalizeOptions(opts Options) Options {
	opts.Code = strings.TrimSpace(opts.Code)
	opts.Name = strings.TrimSpace(opts.Name)
	opts.PluginType = normalizePluginType(opts.PluginType)
	opts.ContentType = strings.TrimSpace(opts.ContentType)
	opts.ContentName = strings.TrimSpace(opts.ContentName)
	opts.Description = strings.TrimSpace(opts.Description)
	opts.Author = strings.TrimSpace(opts.Author)
	opts.Output = strings.TrimSpace(opts.Output)
	if opts.Output == "" {
		opts.Output = "examples/plugins"
	}
	if opts.Code == "" && opts.Name != "" {
		opts.Code = SlugifyCode(opts.Name)
	}
	if opts.PluginType == PluginTypeContent && opts.ContentType == "" && opts.Code != "" {
		contentCode := strings.ReplaceAll(opts.Code, "-", "_")
		opts.ContentType = strings.TrimSuffix(contentCode, "s") + "_item"
	}
	if opts.PluginType == PluginTypeContent && opts.ContentName == "" {
		opts.ContentName = opts.Name + "内容"
	}
	if opts.Description == "" {
		opts.Description = opts.Name + "声明型插件模板。"
	}
	if opts.Author == "" {
		opts.Author = "DevHub Team"
	}
	if opts.MountPoint == "" {
		opts.MountPoint = pluginregistry.FrontendMountAdminPluginPreview
	}
	if opts.ComponentKey == "" {
		opts.ComponentKey = pluginregistry.FrontendComponentAnnouncementCard
	}
	return opts
}

func validateOptions(opts Options) error {
	if !codePattern.MatchString(opts.Code) {
		return errors.New("code 必填，且只能使用小写字母、数字、下划线、中划线，并以小写字母开头，长度 2-64")
	}
	if opts.Name == "" {
		return errors.New("name 不能为空")
	}
	if opts.PluginType == PluginTypeContent && opts.ContentType == "" {
		return errors.New("content_type 不能为空")
	}
	if opts.PluginType == PluginTypeContent && !contentTypePattern.MatchString(opts.ContentType) {
		return errors.New("content_type 只能使用小写字母、数字、下划线，并以小写字母开头，长度 2-64")
	}
	if opts.PluginType == PluginTypeFrontendMount {
		if !pluginregistry.IsFrontendMountPointAllowed(opts.MountPoint) {
			return fmt.Errorf("mount_point 不在官方允许列表中：%s", opts.MountPoint)
		}
		if !pluginregistry.IsFrontendComponentAllowed(opts.ComponentKey) {
			return fmt.Errorf("component_key 不在官方允许列表中：%s", opts.ComponentKey)
		}
	}
	return nil
}

func normalizePluginType(value string) string {
	return NormalizePluginType(value)
}

func NormalizePluginType(value string) string {
	switch strings.TrimSpace(value) {
	case "", "content", "content_plugin":
		return PluginTypeContent
	case "external_service", "service":
		return PluginTypeExternalService
	case "admin_tool", "tool":
		return PluginTypeAdminTool
	case "frontend_mount", "frontend":
		return PluginTypeFrontendMount
	default:
		return PluginTypeContent
	}
}

func SlugifyCode(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	var b strings.Builder
	lastSep := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastSep = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastSep = false
		case r == '_' || r == '-':
			if !lastSep && b.Len() > 0 {
				b.WriteRune('_')
				lastSep = true
			}
		case unicode.IsSpace(r):
			if !lastSep && b.Len() > 0 {
				b.WriteRune('_')
				lastSep = true
			}
		}
	}
	slug := strings.Trim(b.String(), "_-")
	for strings.Contains(slug, "__") {
		slug = strings.ReplaceAll(slug, "__", "_")
	}
	if slug == "" || slug[0] < 'a' || slug[0] > 'z' {
		sum := sha1.Sum([]byte(name))
		slug = "plugin_" + hex.EncodeToString(sum[:4])
	}
	if len(slug) > 64 {
		slug = strings.Trim(slug[:64], "_-")
	}
	return slug
}

func permissionResource(contentType string) string {
	resource := strings.Trim(contentType, "_")
	if resource == "" {
		return "item"
	}
	if strings.HasSuffix(resource, "_item") {
		resource = strings.TrimSuffix(resource, "_item")
	}
	resource = strings.ReplaceAll(resource, "_", ".")
	return strings.Trim(resource, ".")
}

func configJSON(opts Options) string {
	if !opts.WithConfig {
		return "{}"
	}
	if normalizePluginType(opts.PluginType) == PluginTypeExternalService {
		return "{\n  \"enabled\": false,\n  \"endpoint_url\": \"https://example.com\",\n  \"failure_policy\": \"warn\",\n  \"health_check_path\": \"/health\",\n  \"timeout_ms\": 3000\n}"
	}
	if normalizePluginType(opts.PluginType) != PluginTypeContent {
		return "{\n  \"enabled\": true\n}"
	}
	return "{\n  \"display_mode\": \"standard\",\n  \"enabled\": true,\n  \"items_per_page\": 20\n}"
}

func readme(opts Options) string {
	entries := []string{
		"- manifest.json：插件声明。",
		"- config.example.json：可用于全局或子站配置的示例配置。",
		"- docs/usage.md：初始化、预检和安装前检查说明。",
		"- docs/content-types.md：内容数据类型声明说明。",
		"- docs/permissions.md：权限声明说明。",
	}
	if opts.WithHooks {
		entries = append(entries, "- docs/hooks.md：Hook 声明边界。")
	}
	if opts.WithMigration {
		entries = append(entries, "- docs/migrations.md：迁移声明边界。", "- migrations/001_init.sql：唯一模板迁移计划文件。")
	}
	if opts.IncludeRegistryDoc {
		entries = append(entries, "- docs/registry-example.md：内置系统插件接入示例说明，不会被动态加载。")
	}
	return fmt.Sprintf(`# %s

这是 DevHub 声明型插件模板。它只包含 manifest、配置、文档和 migrations 计划文件，不会被系统自动扫描或动态加载。

## 使用方式

1. 打开 manifest.json。
2. 将 JSON 复制到后台 /admin-next/plugins 的 Manifest 校验 / dry-run / install 流程。
3. 先执行 validate，再执行 dry-run；确认 impact、权限、菜单、Hook 和迁移声明后再 install。

## 当前目录内容

%s

## 如何定义插件能力

- 内容类型：编辑 manifest.json 的 content_types 与 content_type_definitions。
- 权限：编辑 permissions，权限码需要和内容类型中的 create/edit/delete/audit permission 对应。
- 菜单：编辑 menus；菜单展示仍会受插件状态、子站状态、scope 和 permission 共同过滤。
- 配置：编辑 config_schema，并用 config.example.json 验证示例配置。
- Hook：编辑 hooks；当前只是声明和治理元信息，或 external_service non-blocking 投递声明，不会加载第三方处理器。
- 迁移：编辑 migrations；当前只生成 migrations/001_init.sql 作为 dry-run 计划输入，不执行外部 SQL。

## 安全边界

- 不执行动态代码。
- 不生成 Go/JS/WASM/binary/package scripts 等可执行模板。
- 不上传插件包。
- 不接入插件市场。
- 不执行外部 SQL。
- 不支持 migration down。
- 不生成 blocking Hook。
- 不执行远程 Webhook。
- 不动态加载前端资产。
- 不破坏 Core 表。
- 不绕过权限校验。
- 不影响历史内容访问。
- 不破坏 /topics/:id SEO。
- 不在示例配置中保存明文敏感字段。
`, opts.Name, strings.Join(entries, "\n"))
}

func usageDoc(opts Options) string {
	return fmt.Sprintf(`# 使用说明

- plugin_code: %s
- plugin_type: %s

生成后的模板仍需经过后台 Manifest 校验、package check、precheck 和 install dry-run。模板生成本身不会创建插件记录、不会安装、不会启用、不会执行 SQL，也不会执行第三方代码。
`, opts.Code, normalizePluginType(opts.PluginType))
}

func contentTypeDoc(opts Options) string {
	if normalizePluginType(opts.PluginType) != PluginTypeContent {
		return fmt.Sprintf(`# 内容数据类型

plugin_type=%s 默认不声明 content_type。只有内容型插件才需要内容数据类型。
`, normalizePluginType(opts.PluginType))
	}
	return fmt.Sprintf(`# 内容类型

- plugin_code: %s
- content_type: %s
- content_name: %s

内容类型由 manifest.json 的 content_type_definitions 声明。发布链路仍由后端校验插件全局状态、子站插件状态、板块绑定、allowed_content_types 和 create_permission。
`, opts.Code, opts.ContentType, opts.ContentName)
}

func permissionsDoc(opts Options) string {
	if normalizePluginType(opts.PluginType) != PluginTypeContent {
		manifest := Manifest(opts)
		lines := make([]string, 0, len(manifest.Permissions))
		for _, permission := range manifest.Permissions {
			lines = append(lines, "- "+permission.Code)
		}
		return fmt.Sprintf(`# 权限

推荐权限码：

%s

前端隐藏按钮不是权限控制。所有管理和配置操作必须由后端按权限码强校验。
`, strings.Join(lines, "\n"))
	}
	resource := permissionResource(opts.ContentType)
	return fmt.Sprintf(`# 权限

推荐权限码：

- %s.%s.create
- %s.%s.edit
- %s.%s.delete
- %s.%s.audit
- %s.manage
- %s.configure

前端隐藏按钮不是权限控制。所有创建、编辑、删除、审核和配置操作必须由后端按权限码强校验。
`, opts.Code, resource, opts.Code, resource, opts.Code, resource, opts.Code, resource, opts.Code, opts.Code)
}

func hooksDoc(opts Options) string {
	return `# Hook 声明

HookDefinition 当前是声明和治理边界，不是第三方动态 Hook 运行时。模板只生成 non-blocking 声明，不生成 blocking Hook。

- non-blocking Hook 失败不阻断主流程，但会写入 hook_executions 与 plugin.hook.failed 审计。
- 当前不支持第三方动态 Hook 处理器。
- 当前不支持远程 Webhook 执行。
`
}

func migrationsDoc(opts Options) string {
	return `# Migration 声明

当前 migration 是内置 up/no-op 与记录型迁移边界。模板中的 migrations 仅为声明示例，不会执行外部 SQL。

当前只允许 migrations/001_init.sql 作为模板迁移计划文件，不生成根目录 001_schema.sql。

当前不支持：

- 外部 raw SQL 执行。
- migration down。
- hard rollback。
- 自动备份。
- 外部插件迁移包。
`
}

func registryExampleDoc(opts Options) string {
	return fmt.Sprintf(`# Registry 示例说明

本文件仅用于解释如何把声明型插件接入 DevHub 源码 registry；它不是动态加载入口，也不会被本地插件包扫描器当作运行时代码执行。

## 说明

- plugin_code: %s
- plugin_type: %s
- content_type: %s
- content_name: %s

如需成为内置系统插件，必须在 DevHub 源码中显式注册并随系统一起编译发布。
`, opts.Code, normalizePluginType(opts.PluginType), opts.ContentType, opts.ContentName)
}

func prepare(opts Options) (PreviewResult, map[string][]byte, error) {
	opts = normalizeOptions(opts)
	if err := validateOptions(opts); err != nil {
		return PreviewResult{}, nil, err
	}

	dir := filepath.Join(opts.Output, opts.Code)
	if info, err := os.Stat(dir); err == nil && !opts.Force {
		if !opts.AllowExistingEmpty || !info.IsDir() || !dirIsEmpty(dir) {
			return PreviewResult{}, nil, fmt.Errorf("输出目录已存在：%s（如需覆盖请加 --force）", dir)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return PreviewResult{}, nil, err
	}
	if opts.Force {
		if err := os.RemoveAll(dir); err != nil {
			return PreviewResult{}, nil, err
		}
	}

	manifest := Manifest(opts)
	manifestRaw, err := marshalManifest(manifest)
	if err != nil {
		return PreviewResult{}, nil, err
	}
	validation := pluginregistry.ValidatePluginManifestJSON(manifestRaw, pluginregistry.Definitions(), "v1.4.0")
	if !validation.Valid {
		return PreviewResult{}, nil, fmt.Errorf("生成 manifest 未通过校验：%s", strings.Join(validation.Errors, "; "))
	}
	if err := pluginregistry.ValidateConfigJSON(domain.Plugin{PluginManifest: manifest}, configJSON(opts)); err != nil {
		return PreviewResult{}, nil, fmt.Errorf("生成 config.example.json 未通过 config_schema 校验：%w", err)
	}

	files := map[string][]byte{
		"manifest.json":         manifestRaw,
		"README.md":             []byte(readme(opts)),
		"config.example.json":   []byte(configJSON(opts) + "\n"),
		"docs/usage.md":         []byte(usageDoc(opts)),
		"docs/permissions.md":   []byte(permissionsDoc(opts)),
		"docs/content-types.md": []byte(contentTypeDoc(opts)),
	}
	if opts.WithHooks {
		files["docs/hooks.md"] = []byte(hooksDoc(opts))
	}
	if opts.WithMigration {
		files["docs/migrations.md"] = []byte(migrationsDoc(opts))
		files["migrations/001_init.sql"] = []byte("-- DevHub 声明型插件迁移占位文件；dry-run 只生成计划，不执行 SQL。\n")
	}
	if opts.IncludeRegistryDoc {
		files["docs/registry-example.md"] = []byte(registryExampleDoc(opts))
	}
	written := make([]string, 0, len(files))
	for name := range files {
		written = append(written, filepath.Join(dir, name))
	}
	sort.Strings(written)
	return PreviewResult{Dir: dir, Files: written, Manifest: manifest}, files, nil
}

func dirIsEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) == 0
}
