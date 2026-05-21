package official_announcement

import "devhub-gin-backend/internal/domain"

const Code = "official_announcement"

// Definition returns the built-in official announcement plugin registration.
//
// v1.8.3 scope:
//   - This is a built-in official plugin shipped with DevHub.
//   - Frontend mounting uses official allowlisted mount points and an official
//     component key; the Host helper still serves the built-in iframe page.
//   - No third-party code execution; no remote iframe URL.
func Definition() domain.Plugin {
	return domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:           Code,
			PluginCode:     Code,
			Name:           "官方公告插件",
			Version:        "0.1.0",
			Description:    "用于验证官方插件前端挂载（Host + iframe + postMessage）的最小闭环；不执行第三方不可信代码。",
			IsSystem:       true,
			MinCoreVersion: "v1.8.1",
			Permissions: []domain.PermissionDefinition{
				{PluginCode: Code, Code: "official_announcement.view", Name: "查看公告插件入口", Description: "查看官方公告插件前后台挂载入口", Scope: "global"},
				{PluginCode: Code, Code: "official_announcement.manage", Name: "管理公告插件", Description: "管理官方公告插件配置与预览入口", Scope: "global"},
			},
			Menus: []domain.MenuDefinition{
				{PluginCode: Code, Code: "official_announcement_admin", Key: "official_announcement_admin", Title: "公告插件", Short: "公告", Path: "/plugins?code=official_announcement", Location: "admin", Area: "admin", Icon: "BellFilled", Permission: "plugin.read", SortOrder: 260},
			},
			Routes: []domain.RouteDefinition{
				{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/api/v1/plugins/official-announcement/context", Handler: "official_announcement.host_context"},
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/plugins/official-announcement/audit-events", Handler: "official_announcement.audit_events"},
				{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/plugins/official-announcement/iframe", Handler: "official_announcement.iframe"},
			},
			FrontendMounts: []domain.FrontendMountDefinition{
				{PluginCode: Code, MountPoint: "frontend.home.section", ComponentKey: "official.announcement.card", RenderMode: "official_component", ConfigRef: "resolved_config"},
				{PluginCode: Code, MountPoint: "frontend.community.section", ComponentKey: "official.announcement.card", RenderMode: "official_component", ConfigRef: "resolved_config"},
				{PluginCode: Code, MountPoint: "admin.plugin.detail.preview", ComponentKey: "official.announcement.card", RenderMode: "official_component", ConfigRef: "resolved_config"},
			},
			ConfigSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"enabled": map[string]any{
						"type":        "boolean",
						"default":     true,
						"description": "是否启用公告展示",
					},
					"message": map[string]any{
						"type":        "string",
						"default":     "",
						"maxLength":   200,
						"description": "公告文案（纯文本；最大 200 字符）",
					},
					"link_text": map[string]any{
						"type":        "string",
						"default":     "",
						"maxLength":   50,
						"description": "链接文案（纯文本；最大 50 字符）",
					},
					"link_url": map[string]any{
						"type":        "string",
						"default":     "/",
						"maxLength":   300,
						"pattern":     `^/([^/[:space:]][^[:space:]]*)?$`,
						"description": "链接地址，仅允许站内绝对路径（以 / 开头，禁止远程 URL、javascript:, data:, file: 等）。",
					},
					"dismissible": map[string]any{
						"type":        "boolean",
						"default":     false,
						"description": "是否允许关闭（前端仅做展示；持久化关闭不在 v1.8.3 范围）",
					},
				},
				"required": []any{"enabled", "message"},
			},
		},
		Status: "enabled",
	}
}
