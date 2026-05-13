package docs

import "devhub-gin-backend/internal/domain"

const Code = "docs"

// Definition returns the built-in documentation plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:           Code,
			PluginCode:     Code,
			Name:           "文档插件",
			Version:        "1.0.0",
			Description:    "提供 document 内容类型、文档空间、文档树和文档详情。",
			IsSystem:       true,
			ContentTypes:   []string{"document"},
			Dependencies:   nil,
			MinCoreVersion: "v1.3.0",
			ContentTypeDefs: []domain.ContentTypeDefinition{
				{
					Type:             "document",
					Name:             "文档",
					PluginCode:       Code,
					Aliases:          []string{"doc"},
					CreatePermission: "docs.document.create",
					EditPermission:   "docs.document.update",
					DeletePermission: "post.delete",
					AuditPermission:  "docs.document.audit",
					DefaultStatus:    "publish",
					AllowComment:     true,
					AllowLike:        true,
					AllowFavorite:    true,
					SEOType:          "TechArticle",
				},
			},
			Menus: []domain.MenuDefinition{
				{PluginCode: Code, Code: "docs", Key: "docs", Title: "文档管理", Short: "文档", Path: "/docs", Location: "admin", Area: "admin", Icon: "Notebook", Permission: "docs.document.create", SortOrder: 300},
				{PluginCode: Code, Code: "docs-moderator", Key: "docs-moderator", Title: "文档治理", Short: "文档", Path: "/moderator/topics?content_type=document", Location: "moderator", Area: "moderator", Icon: "Notebook", Permission: "docs.document.create", SortOrder: 300},
				{
					PluginCode:              Code,
					Code:                    "docs-community-nav",
					Key:                     "docs-community-nav",
					Title:                   "文档",
					Description:             "浏览文档内容",
					Path:                    "/search/",
					Route:                   "/search/",
					Location:                "community_nav",
					Area:                    "frontend",
					Icon:                    "Notebook",
					ContentType:             "document",
					RequireLogin:            false,
					RequireCommunityEnabled: true,
					RequireCategoryBinding:  false,
					VisibleWhen:             []string{"plugin_enabled", "community_enabled", "dependency_satisfied", "config_valid"},
					SortOrder:               130,
					Order:                   130,
				},
			},
			Permissions: []domain.PermissionDefinition{
				{PluginCode: Code, Code: "docs.document.create", Name: "创建文档", Description: "创建 document 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "docs.document.update", Name: "更新文档", Description: "更新 document 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "docs.document.audit", Name: "审核文档", Description: "审核 document 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "docs.space.manage", Name: "管理文档空间", Description: "维护文档空间与结构", Scope: "community"},
			},
			Routes: []domain.RouteDefinition{
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "docs.document.create"},
				{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/topics/:id", Handler: "docs.document.detail"},
				{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=document", Handler: "docs.document.audit"},
			},
			ConfigSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"allow_public_spaces": map[string]any{"type": "boolean", "default": true, "description": "是否允许公开文档空间"},
					"max_tree_depth":      map[string]any{"type": "integer", "min": float64(1), "max": float64(20), "default": float64(5), "description": "文档树最大深度"},
				},
				"required": []any{"allow_public_spaces"},
			},
			Hooks: []domain.HookDefinition{
				{PluginCode: Code, Name: "BeforeCreateContent", Description: "创建文档前预留文档树约束校验", Critical: true, FailurePolicy: "rollback"},
				{PluginCode: Code, Name: "OnSEOBuild", Description: "构建文档详情 SEO 元信息", Critical: false, FailurePolicy: "log"},
				{PluginCode: Code, Name: "OnSearchIndex", Description: "构建文档搜索索引元信息", Critical: false, FailurePolicy: "log"},
			},
			Migrations: []domain.PluginMigrationDefinition{
				{PluginCode: Code, MigrationVersion: "1.0.0", MigrationName: "docs_spaces", Direction: "up", Checksum: "builtin:docs:docs_spaces:v1", Tables: []string{"docs_spaces"}, RollbackSupported: false, Description: "确认 docs_spaces 文档空间表"},
				{PluginCode: Code, MigrationVersion: "1.0.0", MigrationName: "docs_documents", Direction: "up", Checksum: "builtin:docs:docs_documents:v1", Tables: []string{"docs_documents"}, RollbackSupported: false, Description: "确认 docs_documents 文档扩展表"},
			},
		},
		Status: "enabled",
	}
}
