package service

import (
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestSoftUninstallBlocksWhenRequiredDependentEnabled(t *testing.T) {
	mem := store.NewMemoryStore()
	svc := New(mem)

	// Target plugin (remote/local package installed)
	_, _ = mem.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{Code: "demo", Name: "Demo", Version: "1.0.0"},
		Status:         pluginregistry.StatusEnabled,
		SourceType:     "local_package",
	})
	// Dependent plugin enabled and requires demo.
	_, _ = mem.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:    "dep",
			Name:    "Dep",
			Version: "1.0.0",
			Dependencies: []domain.PluginDependency{
				{Code: "demo", Required: true},
			},
		},
		Status:     pluginregistry.StatusEnabled,
		SourceType: "local_package",
	})

	_, err := svc.SoftUninstallPluginAs(PluginUninstallOperator{ID: 1, Name: "admin"}, "demo", map[string]any{"reason": "test"})
	if err == nil {
		t.Fatalf("expected dependency blocked error")
	}
	// Task should be created and failed.
	items, total, _ := mem.PluginUninstallTasks(domain.PluginUninstallTaskFilter{PluginCode: "demo", Page: 1, PageSize: 10})
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 uninstall task, got total=%d items=%d", total, len(items))
	}
	if items[0].Status != domain.PluginUninstallTaskStatusFailed {
		t.Fatalf("expected failed uninstall task, got %q", items[0].Status)
	}
	// Plugin should remain enabled.
	p, _ := mem.PluginByCode("demo")
	if p.Status != pluginregistry.StatusEnabled {
		t.Fatalf("expected plugin remain enabled, got %q", p.Status)
	}
}

func TestSoftUninstallArchivesPluginAndWritesTask(t *testing.T) {
	mem := store.NewMemoryStore()
	svc := New(mem)

	_, _ = mem.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{Code: "demo2", Name: "Demo2", Version: "1.0.0"},
		Status:         pluginregistry.StatusDisabled,
		SourceType:     "local_package",
	})

	resp, err := svc.SoftUninstallPluginAs(PluginUninstallOperator{ID: 2, Name: "admin"}, "demo2", map[string]any{"reason": "no longer needed"})
	if err != nil {
		t.Fatalf("soft uninstall failed: %v", err)
	}
	if resp.Status != domain.PluginUninstallTaskStatusSoftDone {
		t.Fatalf("expected soft_uninstalled, got %q", resp.Status)
	}
	p, _ := mem.PluginByCode("demo2")
	if p.Status != pluginregistry.StatusArchived {
		t.Fatalf("expected archived plugin, got %q", p.Status)
	}
	// Should write at least one admin log event.
	logs := mem.AdminLogs("admin")
	found := false
	for _, it := range logs {
		if it.Action == "plugin.soft_uninstall.success" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected plugin.soft_uninstall.success admin log")
	}
}
