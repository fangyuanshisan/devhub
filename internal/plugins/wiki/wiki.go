package wiki

import "devhub-gin-backend/internal/domain"

const Code = "wiki"

// Definition returns the built-in wiki plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:           Code,
			PluginCode:     Code,
			Name:           "Wiki 插件",
			Version:        "1.0.0",
			Description:    "提供 wiki_page 内容类型、协作编辑、版本历史和回滚能力。",
			IsSystem:       true,
			ContentTypes:   []string{"wiki_page"},
			Dependencies:   nil,
			MinCoreVersion: "v1.3.0",
			ContentTypeDefs: []domain.ContentTypeDefinition{
				{
					Type:             "wiki_page",
					Name:             "Wiki 页面",
					PluginCode:       Code,
					Aliases:          []string{"wiki"},
					CreatePermission: "wiki.page.create",
					EditPermission:   "wiki.page.edit",
					DeletePermission: "post.delete",
					AuditPermission:  "wiki.page.audit",
					DefaultStatus:    "publish",
					AllowComment:     true,
					AllowLike:        true,
					AllowFavorite:    true,
					SEOType:          "TechArticle",
				},
			},
			Menus: []domain.MenuDefinition{
				{PluginCode: Code, Code: "wiki", Key: "wiki", Title: "Wiki 管理", Short: "Wiki", Path: "/wiki", Location: "admin", Area: "admin", Icon: "Document", Permission: "wiki.page.edit", SortOrder: 400},
				{PluginCode: Code, Code: "wiki-moderator", Key: "wiki-moderator", Title: "Wiki 治理", Short: "Wiki", Path: "/moderator/topics?content_type=wiki_page", Location: "moderator", Area: "moderator", Icon: "Document", Permission: "wiki.page.edit", SortOrder: 400},
			},
			Permissions: []domain.PermissionDefinition{
				{PluginCode: Code, Code: "wiki.page.create", Name: "创建 Wiki 页面", Description: "创建 wiki_page 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "wiki.page.edit", Name: "编辑 Wiki 页面", Description: "编辑 wiki_page 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "wiki.page.audit", Name: "审核 Wiki 页面", Description: "审核 wiki_page 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "wiki.page.version.rollback", Name: "回滚 Wiki 版本", Description: "回滚 wiki_page 历史版本", Scope: "community,category"},
			},
			Routes: []domain.RouteDefinition{
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "wiki.page.create"},
				{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/topics/:id", Handler: "wiki.page.detail"},
				{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=wiki_page", Handler: "wiki.page.audit"},
			},
			ConfigSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			Hooks: []domain.HookDefinition{
				{PluginCode: Code, Name: "BeforeUpdateContent", Description: "更新 Wiki 页面前预留版本快照校验", Critical: true, FailurePolicy: "rollback"},
				{PluginCode: Code, Name: "AfterUpdateContent", Description: "更新 Wiki 页面后写入版本记录", Critical: false, FailurePolicy: "log"},
				{PluginCode: Code, Name: "OnSEOBuild", Description: "构建 Wiki 详情 SEO 元信息", Critical: false, FailurePolicy: "log"},
			},
		},
		Status: "enabled",
	}
}
