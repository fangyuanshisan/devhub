package plugins

import (
	"encoding/json"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/plugins/aiworks"
	"devhub-gin-backend/internal/plugins/docs"
	"devhub-gin-backend/internal/plugins/jobs"
	"devhub-gin-backend/internal/plugins/projects"
	"devhub-gin-backend/internal/plugins/qa"
	"devhub-gin-backend/internal/plugins/wiki"
)

const (
	StatusDiscovered        = "discovered"
	StatusInstalled         = "installed"
	StatusMigrated          = "migrated"
	StatusConfigured        = "configured"
	StatusEnabled           = "enabled"
	StatusDisabled          = "disabled"
	StatusRunning           = "running"
	StatusArchived          = "archived"
	StatusConfigInvalid     = "config_invalid"
	StatusMigrationPending  = "migration_pending"
	StatusMigrationFailed   = "migration_failed"
	StatusDependencyMissing = "dependency_missing"
	CoreCode                = "core"
)

// ValidGlobalStatus reports whether a plugin status is accepted by the platform model.
func ValidGlobalStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case StatusDiscovered, StatusInstalled, StatusMigrated, StatusConfigured, StatusEnabled, StatusDisabled, StatusRunning, StatusArchived, StatusConfigInvalid, StatusMigrationPending, StatusMigrationFailed, StatusDependencyMissing:
		return true
	default:
		return false
	}
}

// ApplyLifecycle derives the current lifecycle fields from the persisted
// plugin status. These fields are response-level metadata so we can expose the
// lifecycle model without forcing a destructive schema rewrite.
func ApplyLifecycle(plugin domain.Plugin) domain.Plugin {
	status := strings.TrimSpace(plugin.Status)
	if status == "" {
		status = StatusDiscovered
	}
	plugin.Status = status
	if plugin.InstallStatus == "" {
		if status == StatusDiscovered {
			plugin.InstallStatus = StatusDiscovered
		} else {
			plugin.InstallStatus = StatusInstalled
		}
	}
	if plugin.InstalledAt == "" {
		plugin.InstalledAt = plugin.CreatedAt
	}
	if plugin.LastHealthCheckAt == "" {
		plugin.LastHealthCheckAt = plugin.UpdatedAt
	}
	if status == StatusArchived && plugin.ArchivedAt == "" {
		plugin.ArchivedAt = plugin.UpdatedAt
	}
	if plugin.LifecycleStatus == "" {
		plugin.LifecycleStatus = lifecycleStatusFor(status)
	}
	if plugin.StatusReason == "" {
		plugin.StatusReason = lifecycleReasonFor(status)
	}
	return plugin
}

func lifecycleStatusFor(status string) string {
	switch strings.TrimSpace(status) {
	case StatusEnabled, StatusRunning:
		return StatusRunning
	case StatusDisabled, StatusInstalled, StatusMigrated, StatusConfigured:
		return StatusInstalled
	case StatusArchived:
		return StatusArchived
	case StatusConfigInvalid:
		return StatusConfigInvalid
	case StatusMigrationPending:
		return StatusMigrationPending
	case StatusMigrationFailed:
		return StatusMigrationFailed
	case StatusDependencyMissing:
		return StatusDependencyMissing
	default:
		return StatusDiscovered
	}
}

func lifecycleReasonFor(status string) string {
	switch strings.TrimSpace(status) {
	case StatusEnabled, StatusRunning:
		return "插件已启用并可参与新发布、菜单和治理入口"
	case StatusDisabled:
		return "插件已安装但全局禁用，仅影响新发布和入口展示"
	case StatusInstalled, StatusMigrated, StatusConfigured:
		return "插件已安装，等待启用或进一步治理"
	case StatusArchived:
		return "插件已归档，禁止新建内容；历史内容、配置、迁移和审计记录保留"
	case StatusConfigInvalid:
		return "插件配置未通过校验"
	case StatusMigrationPending:
		return "插件存在待处理迁移"
	case StatusMigrationFailed:
		return "插件存在失败迁移"
	case StatusDependencyMissing:
		return "插件依赖缺失"
	default:
		return "系统已发现插件定义，尚未完成安装生命周期"
	}
}

// Definitions returns built-in system plugin definitions.
func Definitions() []domain.Plugin {
	return []domain.Plugin{
		qa.Definition(),
		docs.Definition(),
		wiki.Definition(),
		projects.Definition(),
		jobs.Definition(),
		aiworks.Definition(),
	}
}

// AllContentTypes returns all canonical content types from built-in definitions.
func AllContentTypes() []string {
	seen := map[string]bool{}
	out := []string{}
	for _, def := range Definitions() {
		for _, typ := range def.ContentTypes {
			typ = NormalizeContentType(typ)
			if typ == "" || seen[typ] {
				continue
			}
			seen[typ] = true
			out = append(out, typ)
		}
	}
	sort.Strings(out)
	return out
}

// ContentTypesByPlugin returns canonical content types owned by a plugin.
func ContentTypesByPlugin(code string) []string {
	def, ok := DefinitionByCode(code)
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	out := []string{}
	for _, typ := range def.ContentTypes {
		typ = NormalizeContentType(typ)
		if typ == "" || seen[typ] {
			continue
		}
		seen[typ] = true
		out = append(out, typ)
	}
	sort.Strings(out)
	return out
}

// MenusByArea returns menus from built-in definitions for a given area.
func MenusByArea(area string) []domain.PluginMenu {
	area = strings.TrimSpace(area)
	menus := []domain.PluginMenu{}
	for _, def := range Definitions() {
		for _, menu := range def.Menus {
			menuArea := strings.TrimSpace(firstNonBlank(menu.Area, menu.Location))
			if area == "" || menuArea == area {
				if menu.Area == "" {
					menu.Area = menuArea
				}
				if menu.Location == "" {
					menu.Location = menuArea
				}
				menus = append(menus, menu)
			}
		}
	}
	return menus
}

// PermissionsByPlugin returns permissions for a built-in plugin.
func PermissionsByPlugin(code string) []domain.PluginPermission {
	def, ok := DefinitionByCode(code)
	if !ok {
		return nil
	}
	return append([]domain.PluginPermission(nil), def.Permissions...)
}

// RoutesByPlugin returns routes for a built-in plugin.
func RoutesByPlugin(code string) []domain.PluginRoute {
	def, ok := DefinitionByCode(code)
	if !ok {
		return nil
	}
	return append([]domain.PluginRoute(nil), def.Routes...)
}

// DefinitionByCode returns a plugin definition by code.
func DefinitionByCode(code string) (domain.Plugin, bool) {
	code = strings.TrimSpace(code)
	for _, def := range Definitions() {
		if def.Code == code || def.PluginCode == code {
			return def, true
		}
	}
	return domain.Plugin{}, false
}

// ContentTypeDefinitions returns normalized content type definitions for a plugin.
func ContentTypeDefinitions(code string) []domain.ContentTypeDefinition {
	def, ok := DefinitionByCode(code)
	if !ok {
		return nil
	}
	items := make([]domain.ContentTypeDefinition, 0, len(def.ContentTypeDefs))
	for _, item := range def.ContentTypeDefs {
		cp := item
		cp.Type = NormalizeContentType(cp.Type)
		aliases := make([]string, 0, len(cp.Aliases))
		for _, alias := range cp.Aliases {
			alias = strings.TrimSpace(alias)
			if alias != "" {
				aliases = append(aliases, alias)
			}
		}
		if len(aliases) == 0 {
			aliases = ContentTypeAliases(cp.Type)
		}
		cp.Aliases = aliases
		items = append(items, cp)
	}
	return items
}

// ContentTypeDefinitionByType resolves a canonical or alias type to a content type definition.
func ContentTypeDefinitionByType(contentType string) (domain.ContentTypeDefinition, bool) {
	want := NormalizeContentType(contentType)
	if want == "" {
		return domain.ContentTypeDefinition{}, false
	}
	for _, def := range Definitions() {
		for _, item := range ContentTypeDefinitions(def.Code) {
			if item.Type == want {
				return item, true
			}
			for _, alias := range item.Aliases {
				if NormalizeContentType(alias) == want || strings.TrimSpace(alias) == strings.TrimSpace(contentType) {
					return item, true
				}
			}
		}
	}
	if !ValidContentType(want) {
		return domain.ContentTypeDefinition{}, false
	}
	return domain.ContentTypeDefinition{
		Type:             want,
		Name:             want,
		PluginCode:       CoreCode,
		Aliases:          ContentTypeAliases(want),
		CreatePermission: "core.topic.create",
		EditPermission:   "post.update",
		DeletePermission: "post.delete",
		DefaultStatus:    "publish",
		AllowComment:     true,
		AllowLike:        true,
		AllowFavorite:    true,
		SEOType:          "Article",
	}, true
}

// HookDefinitions returns declared hooks for a plugin.
func HookDefinitions(code string) []domain.HookDefinition {
	def, ok := DefinitionByCode(code)
	if !ok {
		return nil
	}
	return append([]domain.HookDefinition(nil), def.Hooks...)
}

// MigrationDefinitions returns built-in migration declarations for a plugin.
func MigrationDefinitions(code string) []domain.PluginMigrationDefinition {
	def, ok := DefinitionByCode(code)
	if !ok {
		return nil
	}
	items := append([]domain.PluginMigrationDefinition(nil), def.Migrations...)
	for i := range items {
		if items[i].PluginCode == "" {
			items[i].PluginCode = def.Code
		}
		if items[i].MigrationVersion == "" {
			items[i].MigrationVersion = def.Version
		}
		if items[i].Direction == "" {
			items[i].Direction = "up"
		}
	}
	return items
}

// ResolvePluginConfig merges schema defaults, global config and community config.
func ResolvePluginConfig(def domain.Plugin, globalConfigJSON, communityConfigJSON string) map[string]any {
	out := map[string]any{}
	defaults := map[string]any{}
	if schema, ok := def.ConfigSchema.(map[string]any); ok {
		out["default"] = schema
		defaults = defaultsFromSchema(schema)
	}
	global, hasGlobal := parseConfigJSON(globalConfigJSON)
	community, hasCommunity := parseConfigJSON(communityConfigJSON)
	if hasGlobal {
		out["global"] = global
	}
	if hasCommunity {
		out["community"] = community
	}
	out["effective"] = mergeConfigValues(defaults, len(defaults) > 0, global, hasGlobal, community, hasCommunity)
	return out
}

func defaultsFromSchema(schema map[string]any) map[string]any {
	out := map[string]any{}
	props, _ := schema["properties"].(map[string]any)
	for key, rawSchema := range props {
		subSchema, ok := rawSchema.(map[string]any)
		if !ok {
			continue
		}
		if value, ok := subSchema["default"]; ok {
			out[key] = value
		}
	}
	return out
}

func parseConfigJSON(raw string) (any, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return raw, true
	}
	return out, true
}

func mergeConfigValues(defaults any, hasDefaults bool, global any, hasGlobal bool, community any, hasCommunity bool) any {
	merged := map[string]any{}
	if d, ok := defaults.(map[string]any); ok {
		for key, value := range d {
			merged[key] = value
		}
	}
	if g, ok := global.(map[string]any); ok {
		for key, value := range g {
			merged[key] = value
		}
	}
	if c, ok := community.(map[string]any); ok {
		for key, value := range c {
			merged[key] = value
		}
	}
	if len(merged) > 0 {
		return merged
	}
	if hasCommunity {
		return community
	}
	if hasGlobal {
		return global
	}
	if hasDefaults {
		return defaults
	}
	return map[string]any{}
}

func firstNonBlank(items ...string) string {
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			return item
		}
	}
	return ""
}

// NormalizeContentType maps legacy core content types to plugin-owned content types.
func NormalizeContentType(contentType string) string {
	switch strings.TrimSpace(contentType) {
	case "doc":
		return "document"
	case "wiki":
		return "wiki_page"
	default:
		return strings.TrimSpace(contentType)
	}
}

// ContentTypeAliases returns accepted aliases for a canonical content type.
func ContentTypeAliases(contentType string) []string {
	switch NormalizeContentType(contentType) {
	case "document":
		return []string{"document", "doc"}
	case "wiki_page":
		return []string{"wiki_page", "wiki"}
	default:
		if strings.TrimSpace(contentType) == "" {
			return nil
		}
		return []string{NormalizeContentType(contentType)}
	}
}

// PluginCodeForContentType maps a content type to the owning plugin.
func PluginCodeForContentType(contentType string) string {
	switch NormalizeContentType(contentType) {
	case "question":
		return qa.Code
	case "document":
		return docs.Code
	case "wiki_page":
		return wiki.Code
	case "project":
		return projects.Code
	case "job":
		return jobs.Code
	case "ai_work":
		return aiworks.Code
	default:
		return CoreCode
	}
}

// ValidContentType checks core and plugin-owned content type names.
func ValidContentType(contentType string) bool {
	switch NormalizeContentType(contentType) {
	case "article", "question", "project", "ai_work", "job", "wiki_page", "document", "news":
		return true
	default:
		return false
	}
}

// DefaultAllowedContentTypes returns the category allow-list for a primary content type.
func DefaultAllowedContentTypes(contentType string) []string {
	return ContentTypeAliases(contentType)
}

// ContentTypeAllowed checks a category allow-list, accepting legacy aliases.
func ContentTypeAllowed(allowed []string, contentType string) bool {
	if len(allowed) == 0 {
		return true
	}
	want := NormalizeContentType(contentType)
	for _, item := range allowed {
		if NormalizeContentType(item) == want {
			return true
		}
	}
	return false
}

// MergeRuntimeState overlays persisted status/timestamps onto a static definition.
func MergeRuntimeState(def domain.Plugin, runtime domain.Plugin) domain.Plugin {
	if runtime.Code != "" {
		def.Code = runtime.Code
	}
	if runtime.PluginCode != "" {
		def.PluginCode = runtime.PluginCode
	}
	if runtime.Name != "" {
		def.Name = runtime.Name
	}
	if runtime.Version != "" {
		def.Version = runtime.Version
	}
	if runtime.Status != "" {
		def.Status = runtime.Status
	}
	if runtime.Description != "" {
		def.Description = runtime.Description
	}
	def.CreatedAt = runtime.CreatedAt
	def.UpdatedAt = runtime.UpdatedAt
	return ApplyLifecycle(def)
}
