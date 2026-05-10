package plugins

import (
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
	StatusInstalled = "installed"
	StatusEnabled   = "enabled"
	StatusDisabled  = "disabled"
	CoreCode        = "core"
)

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

// ResolvePluginConfig merges default config schema, global config placeholder and community config placeholder.
// v1.3.x only persists community config_json; plugins.config_json remains a planned extension point.
func ResolvePluginConfig(def domain.Plugin, globalConfigJSON, communityConfigJSON string) map[string]any {
	out := map[string]any{}
	if schema, ok := def.ConfigSchema.(map[string]any); ok {
		out["default"] = schema
	}
	if strings.TrimSpace(globalConfigJSON) != "" {
		out["global_config_json"] = strings.TrimSpace(globalConfigJSON)
	}
	if strings.TrimSpace(communityConfigJSON) != "" {
		out["community_config_json"] = strings.TrimSpace(communityConfigJSON)
	}
	return out
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
	return def
}
