package service

import (
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
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

func TestDeclarativePluginManifestCapabilitiesClosedLoop(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	raw := []byte(`{
		"code": "memo",
		"name": "Memo Plugin",
		"version": "1.0.0",
		"description": "declarative closure test",
		"compatible_core_version": ">=1.3.0",
		"content_types": [
			{
				"type": "memo_note",
				"name": "Memo Note",
				"create_permission": "memo.note.create",
				"edit_permission": "memo.note.edit"
			}
		],
		"permissions": [
			{"code": "memo.note.create", "name": "创建备忘", "scope": "community"},
			{"code": "memo.note.edit", "name": "编辑备忘", "scope": "community"}
		],
		"menus": [
			{"code": "memo.admin", "title": "备忘治理", "path": "/plugins/overview?tab=content", "area": "admin", "permission": "memo.note.create"},
			{"code": "memo.frontend", "title": "备忘", "path": "/memo", "location": "frontend", "permission": "memo.note.create", "content_type": "memo_note", "require_community_enabled": true}
		],
		"routes": [],
		"hooks": [],
		"config_schema": {
			"type": "object",
			"properties": {
				"enabled": {"type": "boolean", "default": true}
			}
		},
		"migrations": []
	}`)

	if _, _, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("install manifest plugin: %v", err)
	}
	if _, err := svc.SetPluginStatus("memo", plugins.StatusEnabled); err != nil {
		t.Fatalf("enable manifest plugin: %v", err)
	}
	if _, err := svc.SetCommunityPluginStatus(1, "memo", plugins.StatusEnabled); err != nil {
		t.Fatalf("enable manifest plugin for community: %v", err)
	}
	if !svc.IsPluginEnabledForCommunity(1, "memo") {
		t.Fatal("manifest plugin should be enabled for community after runtime reload")
	}

	category, err := svc.CreateCategory(1, domain.CategoryRequest{
		Name:                "备忘",
		Slug:                "memo",
		ContentType:         "memo_note",
		PluginCode:          "memo",
		AllowedContentTypes: []string{"memo_note"},
	})
	if err != nil {
		t.Fatalf("create manifest content type category: %v", err)
	}
	normalized, pluginCode, err := svc.ValidateTopicPluginAccess(1, category.ID, "memo_note")
	if err != nil {
		t.Fatalf("validate manifest content type access: %v", err)
	}
	if normalized != "memo_note" || pluginCode != "memo" {
		t.Fatalf("unexpected manifest content type resolution: content_type=%s plugin_code=%s", normalized, pluginCode)
	}
	if got := svc.ContentTypeCreatePermission("memo_note", "memo"); got != "memo.note.create" {
		t.Fatalf("unexpected manifest create permission: %s", got)
	}
	topic, err := svc.CreateTopic(domain.CreateTopicRequest{
		UserID:      1,
		CommunityID: 1,
		CategoryID:  category.ID,
		ContentType: "memo_note",
		Title:       "声明型备忘",
		Summary:     "declarative memo summary",
		Content:     "declarative memo content",
		ActorContext: domain.ActorContext{
			UserID:      1,
			Permissions: []string{"memo.note.create"},
		},
	})
	if err != nil {
		t.Fatalf("create manifest plugin topic: %v", err)
	}
	if topic.PluginCode != "memo" || topic.ContentType != "memo_note" {
		t.Fatalf("topic should preserve manifest plugin ownership, got %#v", topic)
	}

	foundPermissionGroup := false
	for _, item := range svc.AdminPermissions() {
		if item.Code == "plugin.memo" && containsString(item.Ops, "memo.note.create") {
			foundPermissionGroup = true
			break
		}
	}
	if !foundPermissionGroup {
		t.Fatalf("admin permission matrix should include manifest plugin permissions: %#v", svc.AdminPermissions())
	}

	communityPlugins, err := svc.CommunityPlugins(1)
	if err != nil {
		t.Fatalf("community plugins: %v", err)
	}
	foundMenu := false
	for _, plugin := range communityPlugins {
		if plugin.Code != "memo" {
			continue
		}
		for _, menu := range plugin.Menus {
			if menu.Code == "memo.frontend" {
				foundMenu = true
			}
		}
	}
	if !foundMenu {
		t.Fatalf("community runtime should expose manifest plugin menus: %#v", communityPlugins)
	}

	if _, err := svc.SetCommunityPluginStatus(1, "memo", plugins.StatusDisabled); err != nil {
		t.Fatalf("disable manifest plugin for community: %v", err)
	}
	if _, _, err := svc.ValidateTopicPluginAccess(1, category.ID, "memo_note"); err == nil || !strings.Contains(err.Error(), "当前子站未启用") {
		t.Fatalf("community disabled should block manifest content type creation, got %v", err)
	}
	if _, err := svc.SetCommunityPluginStatus(1, "memo", plugins.StatusEnabled); err != nil {
		t.Fatalf("re-enable manifest plugin for community: %v", err)
	}
	if _, err := svc.SetPluginStatus("memo", plugins.StatusDisabled); err != nil {
		t.Fatalf("disable manifest plugin globally: %v", err)
	}
	if _, _, err := svc.ValidateTopicPluginAccess(1, category.ID, "memo_note"); err == nil || !strings.Contains(err.Error(), "插件未启用") {
		t.Fatalf("global disabled should block manifest content type creation, got %v", err)
	}
	if _, err := svc.ArchivePlugin("memo"); err != nil {
		t.Fatalf("archive manifest plugin: %v", err)
	}
	if _, _, err := svc.ValidateTopicPluginAccess(1, category.ID, "memo_note"); err == nil || !strings.Contains(err.Error(), "archived") {
		t.Fatalf("archived manifest plugin should block new content, got %v", err)
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
