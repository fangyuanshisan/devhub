package service

import (
	"errors"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestDispatchHookBlockingRecordsAndAudits(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	err := svc.DispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookBeforeCreateContent,
		Mode: pluginregistry.HookBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  "qa",
			ContentType: "document",
			CommunityID: 1,
			CategoryID:  101,
			ActorType:   pluginregistry.HookActorUser,
			ActorID:     1,
			Actor:       domain.ActorContext{UserID: 1},
		},
	})
	if err == nil {
		t.Fatal("expected blocking hook to reject mismatched content type")
	}
	records, err := svc.HookExecutions("qa", 10)
	if err != nil {
		t.Fatalf("HookExecutions failed: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected hook execution record")
	}
	if records[0].Success {
		t.Fatalf("expected failed hook record, got %#v", records[0])
	}
	if !records[0].Blocking {
		t.Fatalf("expected blocking hook record, got %#v", records[0])
	}
	logs, _ := repo.AdminLogsByFilter(domain.AdminLogFilter{Action: "plugin.hook.blocked", Target: "hooks#qa"})
	if len(logs) == 0 {
		t.Fatal("expected plugin.hook.blocked audit log")
	}
}

func TestDispatchHookNonBlockingFailureRecordsWithoutBlocking(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	svc.RegisterHookHandler(pluginregistry.HookAfterCreateContent, "qa", func(pluginregistry.HookEvent) error {
		return errors.New("e2e non-blocking hook failure")
	})

	err := svc.DispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterCreateContent,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  "qa",
			ContentType: "question",
			CommunityID: 1,
			CategoryID:  101,
			ContentID:   999,
			ActorType:   pluginregistry.HookActorSystem,
		},
	})
	if err != nil {
		t.Fatalf("non-blocking hook should not block caller, got %v", err)
	}
	records, err := svc.HookExecutions("qa", 10)
	if err != nil {
		t.Fatalf("HookExecutions failed: %v", err)
	}
	foundFailure := false
	for _, record := range records {
		if record.HookName == pluginregistry.HookAfterCreateContent && !record.Success && strings.Contains(record.ErrorMessage, "non-blocking") {
			foundFailure = true
			break
		}
	}
	if !foundFailure {
		t.Fatalf("expected failed non-blocking execution record, got %#v", records)
	}
	logs, _ := repo.AdminLogsByFilter(domain.AdminLogFilter{Action: "plugin.hook.failed", Target: "hooks#qa"})
	if len(logs) == 0 {
		t.Fatal("expected plugin.hook.failed audit log")
	}
}

func TestSetPluginStatusEnabledChecksMigrationReadiness(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	if _, err := svc.SetPluginStatus("docs", pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("pending built-in no-op migrations should not block enable: %v", err)
	}

	_, err := repo.SavePluginMigration(domain.PluginMigration{
		PluginCode:       "qa",
		MigrationVersion: "1.0.0",
		Version:          "1.0.0",
		MigrationName:    "qa_questions",
		Direction:        "up",
		Status:           "failed",
		ErrorMessage:     "e2e failed migration",
	})
	if err != nil {
		t.Fatalf("seed failed migration: %v", err)
	}
	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err == nil || !strings.Contains(err.Error(), "失败迁移") {
		t.Fatalf("expected failed migration to block enable, got %v", err)
	}
	if _, err := svc.SetCommunityPluginStatus(1, "qa", pluginregistry.StatusEnabled); err == nil || !strings.Contains(err.Error(), "失败迁移") {
		t.Fatalf("expected failed migration to block community enable, got %v", err)
	}
}

func TestPluginArchiveRestoreLifecycle(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	archived, err := svc.ArchivePlugin("qa")
	if err != nil {
		t.Fatalf("ArchivePlugin qa: %v", err)
	}
	if archived.Status != pluginregistry.StatusArchived || archived.LifecycleStatus != pluginregistry.StatusArchived {
		t.Fatalf("expected archived lifecycle, got %#v", archived)
	}
	if svc.IsPluginEnabled("qa") {
		t.Fatal("archived plugin should not be globally enabled")
	}
	if _, _, err := svc.ValidateTopicPluginAccess(1, 101, "question"); err == nil || !strings.Contains(err.Error(), "当前状态为 archived") {
		t.Fatalf("expected archived plugin to block topic creation, got %v", err)
	}
	if _, err := svc.SetCommunityPluginStatus(1, "qa", pluginregistry.StatusEnabled); err == nil || !strings.Contains(err.Error(), "归档") {
		t.Fatalf("expected archived plugin to block community enable, got %v", err)
	}

	restored, err := svc.RestorePlugin("qa")
	if err != nil {
		t.Fatalf("RestorePlugin qa: %v", err)
	}
	if restored.Status != pluginregistry.StatusDisabled {
		t.Fatalf("restore should not auto-enable plugin, got %#v", restored)
	}
	if svc.IsPluginEnabled("qa") {
		t.Fatal("restored plugin should remain disabled until admin enables it")
	}
	if _, err := svc.SetPluginStatus("qa", pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable after restore: %v", err)
	}

	if _, err := svc.ArchivePlugin("qa"); err != nil {
		t.Fatalf("archive qa again: %v", err)
	}
	if _, err := repo.SavePluginMigration(domain.PluginMigration{
		PluginCode:       "qa",
		MigrationVersion: "1.0.0",
		Version:          "1.0.0",
		MigrationName:    "qa_questions",
		Direction:        "up",
		Status:           "failed",
		ErrorMessage:     "restore blocked by migration",
	}); err != nil {
		t.Fatalf("seed failed migration: %v", err)
	}
	if _, err := svc.RestorePlugin("qa"); err == nil || !strings.Contains(err.Error(), "失败迁移") {
		t.Fatalf("expected failed migration to block restore, got %v", err)
	}
}

func TestPluginHealthStatusSourcesAndAuditFilters(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	if _, err := svc.RunPluginMigration("docs", "docs_spaces", "test"); err != nil {
		t.Fatalf("run docs_spaces migration: %v", err)
	}
	if _, err := svc.RunPluginMigration("docs", "docs_documents", "test"); err != nil {
		t.Fatalf("run docs_documents migration: %v", err)
	}
	health, err := svc.PluginHealth("docs")
	if err != nil {
		t.Fatalf("PluginHealth docs: %v", err)
	}
	if health.Status != "healthy" {
		t.Fatalf("expected healthy docs, got %#v", health)
	}

	if _, err := svc.SetPluginConfig("qa", `{"default_question_status":123}`); err == nil {
		t.Fatal("expected invalid config to be rejected")
	}
	_, _ = repo.SetPluginStatus("qa", pluginregistry.StatusConfigInvalid)
	health, err = svc.PluginHealth("qa")
	if err != nil {
		t.Fatalf("PluginHealth qa config invalid: %v", err)
	}
	if health.Status != "config_invalid" || health.ConfigStatus != "invalid" {
		t.Fatalf("expected config_invalid health, got %#v", health)
	}

	_, _ = repo.SetPluginStatus("qa", pluginregistry.StatusEnabled)
	_, err = repo.SavePluginMigration(domain.PluginMigration{
		PluginCode:       "qa",
		MigrationVersion: "1.0.0",
		Version:          "1.0.0",
		MigrationName:    "qa_questions",
		Direction:        "up",
		Status:           "failed",
		ErrorMessage:     "health failed migration",
	})
	if err != nil {
		t.Fatalf("seed failed migration: %v", err)
	}
	health, err = svc.PluginHealth("qa")
	if err != nil {
		t.Fatalf("PluginHealth qa migration failed: %v", err)
	}
	if health.Status != "error" || health.MigrationStatus != "failed" {
		t.Fatalf("expected migration error health, got %#v", health)
	}

	if _, err := svc.RunPluginMigration("qa", "qa_questions", "test"); err != nil {
		t.Fatalf("retry migration: %v", err)
	}
	if _, err := svc.RunPluginMigration("qa", "qa_answers", "test"); err != nil {
		t.Fatalf("run qa_answers migration: %v", err)
	}
	for i := 0; i < 3; i++ {
		err = svc.DispatchHook(pluginregistry.HookEvent{
			Name: pluginregistry.HookBeforeCreateContent,
			Mode: pluginregistry.HookBlocking,
			Ctx: pluginregistry.HookContext{
				PluginCode:  "qa",
				ContentType: "document",
				CommunityID: 1,
				CategoryID:  101,
				ActorType:   pluginregistry.HookActorUser,
				ActorID:     1,
				RequestID:   "req-health-filter",
				Metadata:    map[string]any{"marker": "health-filter-meta"},
				Actor:       domain.ActorContext{UserID: 1},
			},
		})
		if err == nil {
			t.Fatal("expected blocking hook failure")
		}
	}
	health, err = svc.PluginHealth("qa")
	if err != nil {
		t.Fatalf("PluginHealth qa hook error: %v", err)
	}
	if health.Status != "hook_error" || health.HookStatus != "hook_error" {
		t.Fatalf("expected hook_error health, got %#v", health)
	}

	logs, total := repo.AdminLogsByFilter(domain.AdminLogFilter{
		PluginCode:  "qa",
		Action:      "plugin.hook.blocked",
		CommunityID: 1,
		Metadata:    "health-filter-meta",
		RequestID:   "req-health-filter",
		Page:        1,
		PageSize:    10,
	})
	if total == 0 || len(logs) == 0 {
		t.Fatalf("expected filtered audit logs, total=%d logs=%#v", total, logs)
	}
}
