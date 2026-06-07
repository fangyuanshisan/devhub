package qa

import "devhub-gin-backend/internal/domain"

const Code = "qa"

// Definition returns the built-in question-answer plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:           Code,
			PluginCode:     Code,
			Name:           "问答插件",
			Version:        "1.0.0",
			Description:    "提供 question 内容类型、问题发布、回答、采纳和已解决状态。",
			IsSystem:       true,
			ContentTypes:   []string{"question"},
			Dependencies:   nil,
			MinCoreVersion: "v1.3.0",
			ContentTypeDefs: []domain.ContentTypeDefinition{
				{
					Type:             "question",
					Name:             "问答",
					PluginCode:       Code,
					CreatePermission: "qa.question.create",
					EditPermission:   "post.update",
					DeletePermission: "post.delete",
					AuditPermission:  "qa.question.audit",
					DefaultStatus:    "publish",
					AllowComment:     true,
					AllowLike:        true,
					AllowFavorite:    true,
					SEOType:          "QAPage",
				},
			},
			Menus: []domain.MenuDefinition{
				{PluginCode: Code, Code: "qa", Key: "qa", Title: "问答管理", Short: "问答", Path: "/qa", Location: "admin", Area: "admin", Icon: "QuestionFilled", Permission: "qa.question.audit", SortOrder: 200},
				{PluginCode: Code, Code: "qa-moderator", Key: "qa-moderator", Title: "问答治理", Short: "问答", Path: "/moderator/topics?content_type=question", Location: "moderator", Area: "moderator", Icon: "QuestionFilled", Permission: "qa.question.audit", SortOrder: 200},
				{
					PluginCode:              Code,
					Code:                    "qa-community-nav",
					Key:                     "qa-community-nav",
					Title:                   "问答",
					Description:             "浏览问答内容",
					Path:                    "/search/",
					Route:                   "/search/",
					Location:                "community_nav",
					Area:                    "frontend",
					Icon:                    "QuestionFilled",
					ContentType:             "question",
					RequireLogin:            false,
					RequireCommunityEnabled: true,
					RequireCategoryBinding:  false,
					VisibleWhen:             []string{"plugin_enabled", "community_enabled", "dependency_satisfied", "config_valid"},
					SortOrder:               120,
					Order:                   120,
					Permission:              "",
					Badge:                   "",
				},
			},
			Permissions: []domain.PermissionDefinition{
				{PluginCode: Code, Code: "qa.question.create", Name: "发布问题", Description: "创建 question 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "qa.question.audit", Name: "审核问题", Description: "审核 question 内容", Scope: "community,category"},
				{PluginCode: Code, Code: "qa.answer.create", Name: "提交回答", Description: "在 question 下创建回答", Scope: "community,category"},
				{PluginCode: Code, Code: "qa.answer.accept", Name: "采纳回答", Description: "采纳最佳回答", Scope: "own,community"},
			},
			Routes: []domain.RouteDefinition{
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "qa.question.create"},
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics/:id/comments", Handler: "qa.answer.create"},
				{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics/:id/comments/:commentId/accept", Handler: "qa.answer.accept"},
				{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=question", Handler: "qa.question.audit"},
			},
			ConfigSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"allow_anonymous_answer":    map[string]any{"type": "boolean", "description": "是否允许匿名回答"},
					"require_accept_permission": map[string]any{"type": "boolean", "default": true, "description": "采纳答案是否要求 qa.answer.accept 权限"},
					"default_question_status":   map[string]any{"type": "string", "enum": []any{"publish", "review"}, "default": "publish", "description": "问题默认状态"},
				},
				"required": []any{"allow_anonymous_answer", "default_question_status"},
			},
			Hooks: []domain.HookDefinition{
				{PluginCode: Code, Name: "AfterCreateComment", Description: "回答创建后同步问答状态", Critical: false, FailurePolicy: "log"},
				{PluginCode: Code, Name: "OnSEOBuild", Description: "构建问答详情页 SEO 元信息", Critical: false, FailurePolicy: "log"},
			},
			Migrations: []domain.PluginMigrationDefinition{
				{PluginCode: Code, MigrationVersion: "1.0.0", MigrationName: "qa_questions", Direction: "up", Checksum: "builtin:qa:qa_questions:v1", Tables: []string{"qa_questions"}, RollbackSupported: false, Description: "确认 qa_questions 问题扩展表"},
				{PluginCode: Code, MigrationVersion: "1.0.0", MigrationName: "qa_answers", Direction: "up", Checksum: "builtin:qa:qa_answers:v1", Tables: []string{"qa_answers"}, RollbackSupported: false, Description: "确认 qa_answers 回答扩展表"},
			},
		},
		Status: "enabled",
	}
}
