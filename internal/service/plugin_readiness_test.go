package service

import (
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestPluginReadinessBlockedByMissingDependency(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	consumer := domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:         "consumer",
			Name:         "Consumer",
			Version:      "1.0.0",
			Dependencies: []domain.PluginDependency{{Code: "qa", Version: ">=1.0.0", Required: true}},
		},
		Status: pluginregistry.StatusDisabled,
	}
	if _, err := repo.SavePlugin(consumer); err != nil {
		t.Fatalf("save consumer: %v", err)
	}
	// Force dependency to be disabled so readiness blocks.
	qa, _ := repo.PluginByCode("qa")
	qa.Status = pluginregistry.StatusDisabled
	if _, err := repo.SavePlugin(qa); err != nil {
		t.Fatalf("save qa: %v", err)
	}
	result, err := svc.PluginReadiness("consumer", "enable")
	if err != nil {
		t.Fatalf("readiness should not error: %v", err)
	}
	if result.Status != "blocked" {
		t.Fatalf("expected blocked, got %s", result.Status)
	}
	found := false
	for _, check := range result.Checks {
		if check.DependencyCode == "qa" && check.Status == "blocked" {
			found = true
			if check.Code == "" {
				t.Fatalf("expected check to include code")
			}
		}
	}
	if !found {
		t.Fatalf("expected dependency check for qa")
	}
}
