package service

import (
	"strings"
	"testing"

	"devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestValidateAndInstallPluginManifestLifecycle(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	raw := []byte(`{
		"code": "notes",
		"name": "Notes Plugin",
		"version": "1.0.0",
		"description": "manifest install test",
		"compatible_core_version": ">=1.3.0",
		"content_types": [
			{
				"type": "note",
				"name": "Note",
				"create_permission": "notes.note.create",
				"edit_permission": "notes.note.edit",
				"delete_permission": "notes.note.delete"
			}
		],
		"permissions": [
			{"code": "notes.note.create", "name": "Create Note", "scope": "community"},
			{"code": "notes.note.edit", "name": "Edit Note", "scope": "own"},
			{"code": "notes.note.delete", "name": "Delete Note", "scope": "own"}
		],
		"menus": [
			{"code": "notes.frontend", "title": "Notes", "path": "/notes", "location": "frontend", "permission": "notes.note.create"}
		],
		"routes": [
			{"area": "frontend", "method": "GET", "path": "/notes", "permission": "notes.note.create"}
		],
		"hooks": [
			{"name": "BeforeCreateContent", "mode": "blocking", "failure_policy": "block", "timeout_ms": 1000}
		],
		"config_schema": {
			"type": "object",
			"properties": {
				"enabled": {"type": "boolean", "default": true}
			}
		},
		"migrations": [
			{"migration_version": "1.0.0", "migration_name": "notes_init", "direction": "up", "checksum": "sha256:test"}
		]
	}`)

	validation, err := svc.ValidatePluginManifestJSON(raw)
	if err != nil {
		t.Fatalf("ValidatePluginManifestJSON failed: %v", err)
	}
	if !validation.Valid {
		t.Fatalf("expected valid manifest, got %#v", validation)
	}

	plugin, validation, err := svc.InstallPluginManifest(raw)
	if err != nil {
		t.Fatalf("InstallPluginManifest failed: %v", err)
	}
	if !validation.Valid {
		t.Fatalf("expected install validation valid, got %#v", validation)
	}
	if plugin.Code != "notes" || plugin.Status != plugins.StatusDisabled || plugin.SourceType != "manifest" {
		t.Fatalf("unexpected installed plugin: %#v", plugin)
	}
	if _, ok := svc.PluginByCode("notes"); !ok {
		t.Fatal("installed plugin should be queryable")
	}
	if got := svc.ContentTypeCreatePermission("note", "notes"); got != "notes.note.create" {
		t.Fatalf("unexpected create permission: %q", got)
	}
	if _, err := svc.SetPluginStatus("notes", plugins.StatusEnabled); err != nil {
		t.Fatalf("enable installed manifest plugin: %v", err)
	}
	if _, err := svc.ArchivePlugin("notes"); err != nil {
		t.Fatalf("archive installed manifest plugin: %v", err)
	}
	if _, err := svc.SetPluginStatus("notes", plugins.StatusEnabled); err == nil || !strings.Contains(err.Error(), "归档") {
		t.Fatalf("archived manifest plugin should not enable, got %v", err)
	}
	if _, err := svc.RestorePlugin("notes"); err != nil {
		t.Fatalf("restore installed manifest plugin: %v", err)
	}
	if _, err := svc.SetPluginStatus("notes", plugins.StatusEnabled); err != nil {
		t.Fatalf("enable after restore: %v", err)
	}
}

func TestBulkArchiveAndRestorePlugins(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	result := svc.BulkArchivePlugins([]string{"qa", "docs"})
	if len(result.Succeeded) == 0 {
		t.Fatalf("expected bulk archive to succeed for some plugins, got %#v", result)
	}
	if len(result.Failed) != 0 {
		t.Fatalf("expected bulk archive to succeed without failures, got %#v", result)
	}
	restore := svc.BulkRestorePlugins([]string{"qa", "docs"})
	if len(restore.Succeeded) == 0 {
		t.Fatalf("expected bulk restore to succeed for some plugins, got %#v", restore)
	}
}

func TestPluginUpgradeDryRun(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	raw := []byte(`{
		"code": "qa",
		"name": "问答插件",
		"version": "9.9.9",
		"description": "upgrade preview test",
		"compatible_core_version": ">=1.3.4",
		"content_types": [
			{
				"type": "question",
				"name": "Question",
				"create_permission": "qa.question.create"
			}
		],
		"permissions": [
			{"code": "qa.question.create", "name": "Create Question", "scope": "community"}
		],
		"menus": [],
		"routes": [],
		"hooks": [],
		"config_schema": {
			"type": "object",
			"properties": {}
		},
		"migrations": []
	}`)

	result, err := svc.PluginUpgradeDryRun("qa", raw)
	if err != nil {
		t.Fatalf("PluginUpgradeDryRun failed: %v", err)
	}
	if result.PluginCode != "qa" {
		t.Fatalf("unexpected plugin code: %#v", result)
	}
	if result.CurrentVersion == "" || result.NewVersion != "9.9.9" {
		t.Fatalf("unexpected versions: %#v", result)
	}
	if result.CompatibilityStatus != "compatible" {
		t.Fatalf("expected compatible upgrade preview, got %#v", result)
	}
	if len(result.ChangedKeys) == 0 {
		t.Fatalf("expected changed keys, got %#v", result)
	}
	if result.Diff["current"] == nil || result.Diff["new"] == nil {
		t.Fatalf("expected diff snapshots, got %#v", result)
	}
}

func TestUpgradePluginManifest(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	base := []byte(`{
		"code": "notes",
		"name": "Notes Plugin",
		"version": "1.0.0",
		"description": "upgrade target",
		"compatible_core_version": ">=1.3.0",
		"content_types": [{"type":"note","name":"Note","plugin_code":"notes","create_permission":"notes.note.create"}],
		"permissions": [{"code":"notes.note.create","name":"Create","scope":"community"}],
		"menus": [],
		"routes": [],
		"hooks": [],
		"config_schema": {"type":"object","properties":{}},
		"migrations": []
	}`)
	if _, _, err := svc.InstallPluginManifest(base); err != nil {
		t.Fatalf("install manifest failed: %v", err)
	}

	upgrade := []byte(`{
		"code": "notes",
		"name": "Notes Plugin",
		"version": "2.0.0",
		"description": "upgrade target v2",
		"compatible_core_version": ">=1.3.0",
		"content_types": [{"type":"note","name":"Note","plugin_code":"notes","create_permission":"notes.note.create"}],
		"permissions": [{"code":"notes.note.create","name":"Create","scope":"community"}],
		"menus": [],
		"routes": [],
		"hooks": [],
		"config_schema": {"type":"object","properties":{}},
		"migrations": []
	}`)

	result, err := svc.UpgradePluginManifest("notes", upgrade)
	if err != nil {
		t.Fatalf("UpgradePluginManifest failed: %v", err)
	}
	if result.Plugin.Code != "notes" || result.Plugin.Version != "2.0.0" {
		t.Fatalf("unexpected upgraded plugin: %#v", result.Plugin)
	}
	if result.Plugin.Status != plugins.StatusDisabled {
		t.Fatalf("expected upgraded plugin to remain disabled, got %s", result.Plugin.Status)
	}
	if result.CurrentVersion != "1.0.0" || result.NewVersion != "2.0.0" {
		t.Fatalf("unexpected upgrade preview: %#v", result.PluginUpgradeDryRunResult)
	}
}
