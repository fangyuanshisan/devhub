package plugins

import (
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/plugins/docs"
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
			if area == "" || menu.Area == area {
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
