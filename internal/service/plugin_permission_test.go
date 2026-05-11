package service

import (
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestCreatePermissionForContentTypeUsesPluginDefinitions(t *testing.T) {
	cases := []struct {
		contentType string
		pluginCode  string
		want        string
	}{
		{contentType: "question", pluginCode: "qa", want: "qa.question.create"},
		{contentType: "document", pluginCode: "docs", want: "docs.document.create"},
		{contentType: "wiki_page", pluginCode: "wiki", want: "wiki.page.create"},
		{contentType: "project", pluginCode: "projects", want: "projects.project.create"},
		{contentType: "job", pluginCode: "jobs", want: "jobs.job.create"},
		{contentType: "ai_work", pluginCode: "ai_works", want: "ai_works.work.create"},
		{contentType: "article", pluginCode: "core", want: "core.topic.create"},
		{contentType: "news", pluginCode: "core", want: "core.topic.create"},
	}

	for _, tc := range cases {
		t.Run(tc.contentType, func(t *testing.T) {
			if got := CreatePermissionForContentType(tc.contentType, tc.pluginCode); got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestPostCreateOnlyBridgesCoreTopicCreate(t *testing.T) {
	svc := New(store.NewMemoryStore())

	_, err := svc.CreateTopic(domain.CreateTopicRequest{
		UserID:      1,
		CommunityID: 1,
		CategoryID:  102,
		ContentType: "question",
		Title:       "E2E permission question",
		Summary:     "permission matrix check",
		Content:     "this question should be blocked without qa question create permission",
		ActorContext: domain.ActorContext{
			UserID:      1,
			Permissions: []string{"post.create", "core.topic.create"},
		},
	})
	if err == nil {
		t.Fatal("expected post.create/core.topic.create to be insufficient for question")
	}
	if !strings.Contains(err.Error(), "qa.question.create") {
		t.Fatalf("expected missing qa.question.create error, got %v", err)
	}

	topic, err := svc.CreateTopic(domain.CreateTopicRequest{
		UserID:      1,
		CommunityID: 1,
		CategoryID:  101,
		ContentType: "article",
		Title:       "E2E permission article",
		Summary:     "permission matrix check",
		Content:     "this article should keep working with the legacy post create bridge",
		ActorContext: domain.ActorContext{
			UserID:      1,
			Permissions: []string{"post.create"},
		},
	})
	if err != nil {
		t.Fatalf("expected post.create to remain compatible for core article, got %v", err)
	}
	if topic.ContentType != "article" || topic.PluginCode != "core" {
		t.Fatalf("expected normalized core article, got content_type=%s plugin_code=%s", topic.ContentType, topic.PluginCode)
	}
}

func TestPluginCreatePermissionAllowsPluginTopic(t *testing.T) {
	svc := New(store.NewMemoryStore())

	topic, err := svc.CreateTopic(domain.CreateTopicRequest{
		UserID:      1,
		CommunityID: 1,
		CategoryID:  102,
		ContentType: "question",
		Title:       "E2E permission allowed question",
		Summary:     "permission matrix check",
		Content:     "this question should be allowed with qa question create permission",
		ActorContext: domain.ActorContext{
			UserID:      1,
			Permissions: []string{"qa.question.create"},
		},
	})
	if err != nil {
		t.Fatalf("expected qa.question.create to allow question, got %v", err)
	}
	if topic.ContentType != "question" || topic.PluginCode != "qa" {
		t.Fatalf("expected normalized qa question, got content_type=%s plugin_code=%s", topic.ContentType, topic.PluginCode)
	}
}
