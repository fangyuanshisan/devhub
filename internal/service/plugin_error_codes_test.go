package service

import (
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestPluginNotFoundReturnsStructuredCode(t *testing.T) {
	svc := New(store.NewMemoryStore())
	if _, err := svc.SetPluginStatus("not_exist", pluginregistry.StatusEnabled); err == nil {
		t.Fatal("expected not found error")
	} else if apiErr, ok := err.(*domain.APIError); ok {
		if apiErr.Code != PluginErrNotFound {
			t.Fatalf("unexpected code: %s", apiErr.Code)
		}
		if apiErr.HTTPStatus != 404 {
			t.Fatalf("unexpected http status: %d", apiErr.HTTPStatus)
		}
	} else {
		t.Fatalf("expected *domain.APIError, got %T", err)
	}
}

func TestPluginConfigInvalidReturnsStructuredCode(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	// Write invalid config directly into runtime store (bypassing SetPluginConfig validation),
	// then verify enable readiness blocks with structured error.
	qa, _ := repo.PluginByCode("qa")
	qa.ConfigJSON = `{"default_question_status":123}`
	if _, err := repo.SavePlugin(qa); err != nil {
		t.Fatalf("save qa: %v", err)
	}
	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err == nil {
		t.Fatal("expected enable to fail due to invalid config")
	} else if apiErr, ok := err.(*domain.APIError); ok {
		if apiErr.Code != PluginErrConfigSchemaInvalid {
			t.Fatalf("unexpected code: %s", apiErr.Code)
		}
	} else {
		t.Fatalf("expected *domain.APIError, got %T", err)
	}
}
