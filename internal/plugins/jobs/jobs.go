package jobs

import "devhub-gin-backend/internal/domain"

const Code = "jobs"

// Definition returns the built-in jobs plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:           Code,
			PluginCode:     Code,
			Name:           "招聘插件",
			Version:        "1.0.0",
			Description:    "提供 job 内容类型，用于招聘、内推和团队介绍。",
			IsSystem:       true,
			ContentTypes:   []string{"job"},
			MinCoreVersion: "v1.3.0",
			ContentTypeDefs: []domain.ContentTypeDefinition{
				{
					Type:             "job",
					Name:             "招聘内推",
					PluginCode:       Code,
					CreatePermission: "jobs.job.create",
					EditPermission:   "post.update",
					DeletePermission: "post.delete",
					AuditPermission:  "jobs.job.audit",
					DefaultStatus:    "publish",
					AllowComment:     true,
					AllowLike:        true,
					AllowFavorite:    true,
					SEOType:          "Article",
				},
			},
			Menus: []domain.MenuDefinition{
				{PluginCode: Code, Code: "jobs", Key: "jobs", Title: "招聘管理", Short: "招聘", Path: "/jobs", Location: "admin", Area: "admin", Icon: "Briefcase", Permission: "jobs.job.audit", SortOrder: 600},
				{PluginCode: Code, Code: "jobs-moderator", Key: "jobs-moderator", Title: "招聘治理", Short: "招聘", Path: "/moderator/topics?content_type=job", Location: "moderator", Area: "moderator", Icon: "Briefcase", Permission: "jobs.job.audit", SortOrder: 600},
			},
			Permissions: []domain.PermissionDefinition{
				{PluginCode: Code, Code: "jobs.job.create", Name: "发布招聘", Description: "创建 job 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "jobs.job.audit", Name: "审核招聘", Description: "审核 job 内容", Scope: "community,category"},
			},
			Routes: []domain.RouteDefinition{
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "jobs.job.create"},
				{PluginCode: Code, Area: "frontend", Method: "GET", Path: "/topics/:id", Handler: "jobs.job.detail"},
				{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=job", Handler: "jobs.job.audit"},
			},
			ConfigSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			Hooks: []domain.HookDefinition{
				{PluginCode: Code, Name: "OnSEOBuild", Description: "构建招聘详情 SEO 元信息", Critical: false, FailurePolicy: "log"},
			},
		},
		Status: "enabled",
	}
}
