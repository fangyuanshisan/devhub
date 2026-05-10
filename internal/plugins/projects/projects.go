package projects

import "devhub-gin-backend/internal/domain"

const Code = "projects"

// Definition returns the built-in projects plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:           Code,
			PluginCode:     Code,
			Name:           "开源项目插件",
			Version:        "1.0.0",
			Description:    "提供 project 内容类型，用于项目发布、组件推荐和开源协作。",
			IsSystem:       true,
			ContentTypes:   []string{"project"},
			MinCoreVersion: "v1.3.0",
			ContentTypeDefs: []domain.ContentTypeDefinition{
				{
					Type:             "project",
					Name:             "开源项目",
					PluginCode:       Code,
					CreatePermission: "projects.project.create",
					EditPermission:   "post.update",
					DeletePermission: "post.delete",
					AuditPermission:  "projects.project.audit",
					DefaultStatus:    "publish",
					AllowComment:     true,
					AllowLike:        true,
					AllowFavorite:    true,
					SEOType:          "TechArticle",
				},
			},
			Menus: []domain.MenuDefinition{
				{PluginCode: Code, Code: "projects", Key: "projects", Title: "项目管理", Short: "项目", Path: "/projects", Location: "admin", Area: "admin", Icon: "FolderOpened", Permission: "projects.project.audit", SortOrder: 500},
				{PluginCode: Code, Code: "projects-moderator", Key: "projects-moderator", Title: "项目治理", Short: "项目", Path: "/moderator/topics?content_type=project", Location: "moderator", Area: "moderator", Icon: "FolderOpened", Permission: "projects.project.audit", SortOrder: 500},
			},
			Permissions: []domain.PermissionDefinition{
				{PluginCode: Code, Code: "projects.project.create", Name: "发布项目", Description: "创建 project 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "projects.project.audit", Name: "审核项目", Description: "审核 project 内容", Scope: "community,category"},
			},
			Routes: []domain.RouteDefinition{
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "projects.project.create"},
				{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/topics/:id", Handler: "projects.project.detail"},
				{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=project", Handler: "projects.project.audit"},
			},
			ConfigSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			Hooks: []domain.HookDefinition{
				{PluginCode: Code, Name: "OnSEOBuild", Description: "构建项目详情 SEO 元信息", Critical: false, FailurePolicy: "log"},
				{PluginCode: Code, Name: "OnSearchIndex", Description: "构建项目搜索索引元信息", Critical: false, FailurePolicy: "log"},
			},
		},
		Status: "enabled",
	}
}
