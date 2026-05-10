package qa

import "devhub-gin-backend/internal/domain"

const Code = "qa"

// Definition returns the built-in question-answer plugin registration.
func Definition() domain.Plugin {
	return domain.Plugin{
		Code:         Code,
		PluginCode:   Code,
		Name:         "问答插件",
		Version:      "1.0.0",
		Status:       "enabled",
		Description:  "提供 question 内容类型、问题发布、回答、采纳和已解决状态。",
		ContentTypes: []string{"question"},
		Menus: []domain.PluginMenu{
			{PluginCode: Code, Key: "qa", Title: "问答管理", Short: "问答", Path: "/qa", Area: "admin", Icon: "QuestionFilled", Permission: "qa.question.audit"},
			{PluginCode: Code, Key: "qa-moderator", Title: "问答治理", Short: "问答", Path: "/moderator/topics?content_type=question", Area: "moderator", Icon: "QuestionFilled", Permission: "qa.question.audit"},
		},
		Permissions: []domain.PluginPermission{
			{PluginCode: Code, Code: "qa.question.create", Name: "发布问题", Scope: "community,category"},
			{PluginCode: Code, Code: "qa.question.audit", Name: "审核问题", Scope: "community,category"},
			{PluginCode: Code, Code: "qa.answer.create", Name: "提交回答", Scope: "community,category"},
			{PluginCode: Code, Code: "qa.answer.accept", Name: "采纳回答", Scope: "community,category"},
		},
		Routes: []domain.PluginRoute{
			{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics", Handler: "qa.question.create"},
			{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics/:id/comments", Handler: "qa.answer.create"},
			{PluginCode: Code, Area: "frontend", Method: "POST", Path: "/api/v1/topics/:id/comments/:commentId/accept", Handler: "qa.answer.accept"},
			{PluginCode: Code, Area: "admin", Method: "GET", Path: "/api/v1/admin/posts?content_type=question", Handler: "qa.question.audit"},
		},
	}
}
