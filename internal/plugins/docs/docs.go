package docs

import "devhub-gin-backend/internal/domain"

const Code = "docs"

// Definition returns the built-in documentation plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		Code:         Code,
		PluginCode:   Code,
		Name:         "文档插件",
		Version:      "1.0.0",
		Status:       "enabled",
		Description:  "提供 document 内容类型、文档空间、文档树和文档详情。",
		ContentTypes: []string{"document"},
		Menus: []domain.PluginMenu{
			{PluginCode: Code, Key: "docs", Title: "文档管理", Short: "文档", Path: "/docs", Area: "admin", Icon: "Notebook", Permission: "docs.document.create"},
			{PluginCode: Code, Key: "docs-moderator", Title: "文档治理", Short: "文档", Path: "/moderator/topics?content_type=document", Area: "moderator", Icon: "Notebook", Permission: "docs.document.create"},
		},
		Permissions: []domain.PluginPermission{
			{PluginCode: Code, Code: "docs.document.create", Name: "创建文档", Scope: "community,category"},
			{PluginCode: Code, Code: "docs.document.update", Name: "更新文档", Scope: "community,category"},
			{PluginCode: Code, Code: "docs.document.audit", Name: "审核文档", Scope: "community,category"},
			{PluginCode: Code, Code: "docs.space.manage", Name: "管理文档空间", Scope: "community"},
		},
		Routes: []domain.PluginRoute{
			{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "docs.document.create"},
			{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/topics/:id", Handler: "docs.document.detail"},
			{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=document", Handler: "docs.document.audit"},
		},
	}
}
