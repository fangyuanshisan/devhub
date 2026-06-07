package service

import (
	"errors"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestPluginRegistryRefreshAfterLifecycleAndConfigChanges(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusDisabled); err != nil {
		t.Fatalf("disable qa: %v", err)
	}
	if svc.IsPluginEnabled("qa") {
		t.Fatal("qa should be disabled after registry refresh")
	}
	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable qa: %v", err)
	}
	if !svc.IsPluginEnabled("qa") {
		t.Fatal("qa should be enabled after registry refresh")
	}

	if _, err := svc.SetCommunityPluginStatus(1, "qa", pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable community qa: %v", err)
	}
	if !svc.IsPluginEnabledForCommunity(1, "qa") {
		t.Fatal("qa should be enabled for community after registry refresh")
	}
	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusDisabled); err != nil {
		t.Fatalf("disable global qa: %v", err)
	}
	if svc.IsPluginEnabledForCommunity(1, "qa") {
		t.Fatal("global disabled plugin should invalidate community runtime snapshot")
	}
	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("re-enable global qa: %v", err)
	}
	if _, err := svc.SetCommunityPluginStatus(1, "qa", pluginregistry.StatusDisabled); err != nil {
		t.Fatalf("disable community qa: %v", err)
	}
	if svc.IsPluginEnabledForCommunity(1, "qa") {
		t.Fatal("qa community status should be disabled after registry refresh")
	}

	if _, err := svc.SetPluginConfig("official_announcement", `{"enabled":true,"message":"新的公告","link_text":"详情","link_url":"/","dismissible":false}`); err != nil {
		t.Fatalf("set official announcement config: %v", err)
	}
	plugin, ok := svc.PluginByCode("official_announcement")
	if !ok {
		t.Fatal("official_announcement should exist")
	}
	cfg, ok := plugin.ResolvedConfig.(map[string]any)
	effective, _ := cfg["effective"].(map[string]any)
	if !ok || effective["message"] != "新的公告" {
		t.Fatalf("expected refreshed resolved config, got %#v", plugin.ResolvedConfig)
	}

	logs, _ := repo.AdminLogsByFilter(domain.AdminLogFilter{Action: "plugin.registry.reload.after_config_change", Target: "plugins#official_announcement"})
	if len(logs) == 0 {
		t.Fatal("expected registry reload audit for config change")
	}
}

func TestPluginRegistryRefreshFailureKeepsOldRuntimeSnapshot(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable qa: %v", err)
	}
	if !svc.IsPluginEnabled("qa") {
		t.Fatal("qa should be enabled before injected reload failure")
	}

	svc.setPluginRegistryRefreshFailureForTest(func(event pluginRegistryRefreshEvent) error {
		if event.Trigger == "after_disable" {
			return errors.New("injected reload failure")
		}
		return nil
	})
	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusDisabled); err == nil {
		t.Fatal("expected reload failure to be returned")
	}
	if !svc.IsPluginEnabled("qa") {
		t.Fatal("old runtime snapshot should remain enabled after reload failure")
	}
	logs, _ := repo.AdminLogsByFilter(domain.AdminLogFilter{Action: "plugin.registry.reload.failed", Target: "plugins#qa"})
	if len(logs) == 0 {
		t.Fatal("expected failed registry reload audit log")
	}
}
