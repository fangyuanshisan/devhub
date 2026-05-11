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
