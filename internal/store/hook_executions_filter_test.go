package store

import (
	"testing"

	"devhub-gin-backend/internal/domain"
)

func TestHookExecutionsByFilter_MemoryStoreFiltersAndPaging(t *testing.T) {
	repo := NewMemoryStore()
	_, _ = repo.AppendHookExecution(domain.HookExecution{HookName: "BeforeCreateContent", PluginCode: "qa", Mode: "blocking", Blocking: true, Success: false, StartedAt: "2026-05-12 10:00:00"})
	_, _ = repo.AppendHookExecution(domain.HookExecution{HookName: "AfterCreateContent", PluginCode: "qa", Mode: "non_blocking", Blocking: false, Success: true, StartedAt: "2026-05-12 10:01:00"})

	ok := true
	items, total, err := repo.HookExecutionsByFilter(domain.HookExecutionFilter{PluginCode: "qa", Success: &ok, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("HookExecutionsByFilter: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item, got total=%d len=%d", total, len(items))
	}
	if items[0].HookName != "AfterCreateContent" {
		t.Fatalf("unexpected item %#v", items[0])
	}
}
