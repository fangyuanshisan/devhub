package wiki

import "devhub-gin-backend/internal/domain"

const Code = "wiki"

// Definition returns the built-in wiki plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		Code:         Code,
		PluginCode:   Code,
		Name:         "Wiki 插件",
		Version:      "1.0.0",
		Status:       "enabled",
		Description:  "提供 wiki_page 内容类型、协作编辑、版本历史和回滚能力。",
		ContentTypes: []string{"wiki_page"},
		Menus: []domain.PluginMenu{
			{PluginCode: Code, Key: "wiki", Title: "Wiki 管理", Short: "Wiki", Path: "/wiki", Area: "admin", Icon: "Document", Permission: "wiki.page.edit"},
			{PluginCode: Code, Key: "wiki-moderator", Title: "Wiki 治理", Short: "Wiki", Path: "/moderator/topics?content_type=wiki_page", Area: "moderator", Icon: "Document", Permission: "wiki.page.edit"},
		},
		Permissions: []domain.PluginPermission{
			{PluginCode: Code, Code: "wiki.page.create", Name: "创建 Wiki 页面", Scope: "community,category"},
			{PluginCode: Code, Code: "wiki.page.edit", Name: "编辑 Wiki 页面", Scope: "community,category"},
			{PluginCode: Code, Code: "wiki.page.audit", Name: "审核 Wiki 页面", Scope: "community,category"},
			{PluginCode: Code, Code: "wiki.page.version.rollback", Name: "回滚 Wiki 版本", Scope: "community,category"},
		},
		Routes: []domain.PluginRoute{
			{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "wiki.page.create"},
			{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/topics/:id", Handler: "wiki.page.detail"},
			{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=wiki_page", Handler: "wiki.page.audit"},
		},
	}
}
