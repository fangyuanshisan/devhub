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

func TestInstalledManifestPluginMigrationsCanRun(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	raw := []byte(`{
		"code": "manifest_migrate",
		"name": "Manifest Migrate",
		"version": "1.0.0",
		"compatible_core_version": ">=1.3.0",
		"content_types": [
			{"type": "manifest_migrate_item", "name": "Manifest Item", "create_permission": "manifest_migrate.item.create"}
		],
		"permissions": [
			{"code": "manifest_migrate.item.create", "name": "Create Manifest Item", "scope": "community"}
		],
		"migrations": [
			{
				"plugin_code": "manifest_migrate",
				"migration_version": "1.0.0",
				"migration_name": "manifest_migrate_init",
				"direction": "up",
				"checksum": "sha256:no-op"
			}
		]
	}`)

	if _, _, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("InstallPluginManifest failed: %v", err)
	}
	before, err := svc.PluginMigrations("manifest_migrate")
	if err != nil {
		t.Fatalf("PluginMigrations before run failed: %v", err)
	}
	if len(before) != 1 || before[0].Status != "pending" {
		t.Fatalf("expected pending manifest migration, got %#v", before)
	}

	ran, err := svc.RunAllPluginMigrations("manifest_migrate", "test")
	if err != nil {
		t.Fatalf("RunAllPluginMigrations failed: %v", err)
	}
	if len(ran) != 1 || ran[0].Status != "success" {
		t.Fatalf("expected successful manifest migration run, got %#v", ran)
	}
	after, err := svc.PluginMigrations("manifest_migrate")
	if err != nil {
		t.Fatalf("PluginMigrations after run failed: %v", err)
	}
	if len(after) != 1 || after[0].Status != "success" || after[0].FinishedAt == "" {
		t.Fatalf("expected manifest migration to remain successful, got %#v", after)
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

func TestFrontendMountsForRuntimeAllowlistAndStatus(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	raw := []byte(`{
		"code": "frontsafe",
		"name": "Frontend Safe",
		"version": "1.0.0",
		"frontend_mounts": [
			{
				"mount_point": "frontend.home.section",
				"component_key": "official.announcement.card",
				"render_mode": "official_component",
				"config_ref": "resolved_config",
				"props": {"title":"hello","token":"do-not-leak","webhook_secret":"hide"}
			}
		]
	}`)
	if _, _, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("install frontend manifest: %v", err)
	}

	disabled := svc.FrontendMountsForRuntime("frontend.home.section", 0)
	if len(disabled.Items) == 0 {
		t.Fatalf("expected at least one official mount at runtime: %#v", disabled)
	}
	for _, item := range disabled.Items {
		if item.PluginCode == "frontsafe" {
			t.Fatalf("disabled plugin should not render frontend mounts: %#v", disabled)
		}
	}

	if _, err := svc.SetPluginStatus("frontsafe", plugins.StatusEnabled); err != nil {
		t.Fatalf("enable frontend plugin: %v", err)
	}
	enabled := svc.FrontendMountsForRuntime("frontend.home.section", 0)
	foundFrontsafe := false
	for _, item := range enabled.Items {
		if item.PluginCode != "frontsafe" {
			continue
		}
		foundFrontsafe = true
		if item.Props["token"] != nil || item.Props["webhook_secret"] != nil {
			t.Fatalf("sensitive props leaked to frontend runtime: %#v", item.Props)
		}
		if item.Props["title"] != "hello" {
			t.Fatalf("safe prop missing from frontend runtime: %#v", item.Props)
		}
	}
	if !foundFrontsafe {
		t.Fatalf("expected frontsafe mount after enable, got %#v", enabled.Items)
	}

	communityMissing := svc.FrontendMountsForRuntime("frontend.home.section", 1)
	for _, item := range communityMissing.Items {
		if item.PluginCode == "frontsafe" {
			t.Fatalf("community without plugin enablement should not render frontsafe mounts: %#v", communityMissing)
		}
	}
	if _, err := svc.SetCommunityPluginStatus(1, "frontsafe", plugins.StatusEnabled); err != nil {
		t.Fatalf("enable frontend plugin for community: %v", err)
	}
	communityEnabled := svc.FrontendMountsForRuntime("frontend.home.section", 1)
	foundCommunityFrontsafe := false
	for _, item := range communityEnabled.Items {
		if item.PluginCode == "frontsafe" {
			foundCommunityFrontsafe = true
		}
	}
	if !foundCommunityFrontsafe {
		t.Fatalf("expected community enabled frontend mount, got %#v", communityEnabled.Items)
	}

	historical, _ := repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:    "legacy_frontend",
			Name:    "Legacy Frontend",
			Version: "1.0.0",
			FrontendMounts: []domain.FrontendMountDefinition{
				{PluginCode: "legacy_frontend", MountPoint: "frontend.home.section", ComponentKey: "legacy.remote.widget"},
			},
		},
		Status: plugins.StatusEnabled,
	})
	if historical.Code == "" {
		t.Fatal("expected historical plugin to save")
	}
	withLegacy := svc.FrontendMountsForRuntime("frontend.home.section", 0)
	for _, item := range withLegacy.Items {
		if item.PluginCode == "legacy_frontend" {
			t.Fatalf("unknown historical component should be skipped, got %#v", withLegacy)
		}
	}
	if len(withLegacy.Warnings) == 0 {
		t.Fatalf("expected warning for skipped historical mount")
	}

	if _, err := svc.ArchivePlugin("frontsafe"); err != nil {
		t.Fatalf("archive frontend plugin: %v", err)
	}
	archived := svc.FrontendMountsForRuntime("frontend.home.section", 0)
	for _, item := range archived.Items {
		if item.PluginCode == "frontsafe" {
			t.Fatalf("archived plugin should not render frontend mounts: %#v", archived)
		}
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

func TestPluginUpgradeDryRunStructuredPlansAndConfirm(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	base := []byte(`{
		"code": "articles",
		"name": "Articles",
		"version": "1.0.0",
		"compatible_core_version": ">=1.3.0",
		"content_types": [{"type":"article","name":"Article","plugin_code":"articles","create_permission":"articles.create"}],
		"permissions": [{"code":"articles.create","name":"Create","scope":"community"},{"code":"articles.delete","name":"Delete","scope":"community"}],
		"menus": [{"code":"articles.frontend","title":"Articles","path":"/articles","location":"frontend","permission":"articles.create"}],
		"routes": [{"area":"frontend","method":"GET","path":"/articles","permission":"articles.create"}],
		"hooks": [{"name":"BeforeCreateContent","mode":"blocking","failure_policy":"block","timeout_ms":1000}],
		"config_schema": {"type":"object","properties":{"token":{"type":"string"}}},
		"migrations": []
	}`)
	if _, _, err := svc.InstallPluginManifest(base); err != nil {
		t.Fatalf("install base failed: %v", err)
	}
	upgrade := []byte(`{
		"code": "articles",
		"name": "Articles v2",
		"version": "2.0.0",
		"compatible_core_version": ">=1.3.0",
		"content_types": [
			{"type":"article","name":"Article","plugin_code":"articles","create_permission":"articles.create"},
			{"type":"featured_article","name":"Featured Article","plugin_code":"articles","create_permission":"articles.create"}
		],
		"permissions": [{"code":"articles.create","name":"Create","scope":"community"}],
		"menus": [{"code":"articles.frontend","title":"Articles","path":"/articles","location":"frontend","permission":"articles.create"}],
		"routes": [{"area":"frontend","method":"GET","path":"/articles","permission":"articles.create"}],
		"hooks": [{"name":"AfterCreateContent","mode":"non_blocking","service_type":"external_service","path":"/hooks/articles.after_create","method":"POST","timeout_ms":3000,"retry_enabled":true,"max_attempts":3,"failure_policy":"warn","enabled":true}],
		"config_schema": {"type":"object","properties":{"token":{"type":"number"},"endpoint_url":{"type":"string"}}},
		"migrations": [{"plugin_code":"articles","migration_version":"2.0.0","migration_name":"articles_2","direction":"up","checksum":"sha256:test","rollback_supported":false}],
		"assets": ["assets/articles.css"],
		"external_service": {"endpoint":"https://example.com/hooks/articles","timeout_ms":3000,"failure_policy":"warn","retry_policy":"linear"}
	}`)

	preview, err := svc.PluginUpgradeDryRun("articles", upgrade)
	if err != nil {
		t.Fatalf("PluginUpgradeDryRun failed: %v", err)
	}
	if preview.RiskLevel != "warning" || !preview.ConfirmRequired {
		t.Fatalf("expected warning confirm required preview, got %#v", preview)
	}
	if preview.VersionPlan.Compare != "upgrade" || !preview.VersionPlan.CodeMatched {
		t.Fatalf("unexpected version plan: %#v", preview.VersionPlan)
	}
	if preview.ChangeSummary.HighRisk == 0 || preview.ImpactSummary.AffectedPermissionsCount == 0 {
		t.Fatalf("expected structured change and impact summary, got %#v", preview)
	}
	if len(preview.MigrationPlan.Items) == 0 || len(preview.HookPlan.Items) == 0 {
		t.Fatalf("expected migration and hook plans, got %#v", preview)
	}
	foundSecret := false
	for _, section := range preview.DiffSections {
		for _, item := range section.Items {
			if item.Path == "config_schema.token" && item.After != "[REDACTED]" {
				t.Fatalf("expected secret redaction, got %#v", item)
			}
			if item.Path == "config_schema.token" {
				foundSecret = true
			}
		}
	}
	if !foundSecret {
		t.Fatalf("expected secret field diff, got %#v", preview.DiffSections)
	}
	if _, err := svc.UpgradePluginManifestConfirmed("articles", upgrade, false); err == nil {
		t.Fatal("expected confirm-required upgrade to fail without confirmation")
	} else if apiErr, ok := err.(*domain.APIError); !ok || apiErr.Code != "plugin_upgrade_confirm_required" {
		t.Fatalf("expected confirm-required error, got %v", err)
	}
	result, err := svc.UpgradePluginManifestConfirmed("articles", upgrade, true)
	if err != nil {
		t.Fatalf("confirmed upgrade failed: %v", err)
	}
	if result.Plugin.Version != "2.0.0" {
		t.Fatalf("unexpected upgraded plugin: %#v", result.Plugin)
	}
}

func TestPluginUpgradeDryRunBlocksSameVersionEvenWithConfirm(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	base := []byte(`{"code":"notes","name":"Notes","version":"1.0.0","compatible_core_version":">=1.3.0","content_types":[],"permissions":[],"menus":[],"routes":[],"hooks":[],"config_schema":{"type":"object","properties":{}},"migrations":[]}`)
	if _, _, err := svc.InstallPluginManifest(base); err != nil {
		t.Fatalf("install base failed: %v", err)
	}
	same := []byte(`{"code":"notes","name":"Notes","version":"1.0.0","compatible_core_version":">=1.3.0","content_types":[],"permissions":[],"menus":[],"routes":[],"hooks":[],"config_schema":{"type":"object","properties":{}},"migrations":[]}`)
	if _, err := svc.UpgradePluginManifestConfirmed("notes", same, true); err == nil {
		t.Fatal("expected same-version upgrade to be blocked")
	} else if apiErr, ok := err.(*domain.APIError); !ok || apiErr.Code != "plugin_upgrade_blocked" {
		t.Fatalf("expected blocked error, got %v", err)
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

	result, err := svc.UpgradePluginManifestConfirmed("notes", upgrade, true)
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
