package aiworks

import "devhub-gin-backend/internal/domain"

const Code = "ai_works"

// Definition returns the built-in AI works plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:           Code,
			PluginCode:     Code,
			Name:           "AI 作品插件",
			Version:        "1.0.0",
			Description:    "提供 ai_work 内容类型，用于 AI 工具、Agent、工作流和实验作品发布。",
			IsSystem:       true,
			ContentTypes:   []string{"ai_work"},
			MinCoreVersion: "v1.3.0",
			ContentTypeDefs: []domain.ContentTypeDefinition{
				{
					Type:             "ai_work",
					Name:             "AI 作品",
					PluginCode:       Code,
					CreatePermission: "ai_works.work.create",
					EditPermission:   "post.update",
					DeletePermission: "post.delete",
					AuditPermission:  "ai_works.work.audit",
					DefaultStatus:    "publish",
					AllowComment:     true,
					AllowLike:        true,
					AllowFavorite:    true,
					SEOType:          "TechArticle",
				},
			},
			Menus: []domain.MenuDefinition{
				{PluginCode: Code, Code: "ai_works", Key: "ai_works", Title: "AI 作品管理", Short: "AI作品", Path: "/ai-works", Location: "admin", Area: "admin", Icon: "MagicStick", Permission: "ai_works.work.audit", SortOrder: 700},
				{PluginCode: Code, Code: "ai_works-moderator", Key: "ai_works-moderator", Title: "AI 作品治理", Short: "AI作品", Path: "/moderator/topics?content_type=ai_work", Location: "moderator", Area: "moderator", Icon: "MagicStick", Permission: "ai_works.work.audit", SortOrder: 700},
			},
			Permissions: []domain.PermissionDefinition{
				{PluginCode: Code, Code: "ai_works.work.create", Name: "发布 AI 作品", Description: "创建 ai_work 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "ai_works.work.audit", Name: "审核 AI 作品", Description: "审核 ai_work 内容", Scope: "community,category"},
			},
			Routes: []domain.RouteDefinition{
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "ai_works.work.create"},
				{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/topics/:id", Handler: "ai_works.work.detail"},
				{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=ai_work", Handler: "ai_works.work.audit"},
			},
			ConfigSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			Hooks: []domain.HookDefinition{
				{PluginCode: Code, Name: "OnSEOBuild", Description: "构建 AI 作品详情 SEO 元信息", Critical: false, FailurePolicy: "log"},
				{PluginCode: Code, Name: "OnSearchIndex", Description: "构建 AI 作品搜索索引元信息", Critical: false, FailurePolicy: "log"},
			},
		},
		Status: "enabled",
	}
}
