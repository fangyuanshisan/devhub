package service

import (
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestListPluginPackages_EmptyRepo(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	root := mustProjectRoot(t)
	dir := filepath.Join(root, ".devhub", "plugins", "repo_empty")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, ".devhub")) })

	resp, err := svc.ListPluginPackages(".devhub/plugins/repo_empty", PluginPackageRepositoryFilter{})
	if err != nil {
		t.Fatalf("ListPluginPackages: %v", err)
	}
	if len(resp.Items) != 0 {
		t.Fatalf("expected empty items, got %#v", resp.Items)
	}
	if resp.Summary.Total != 0 {
		t.Fatalf("expected summary total 0, got %#v", resp.Summary)
	}
}

func TestListPluginPackages_RepoNotFound(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, err := svc.ListPluginPackages(".devhub/plugins/repo_not_exist", PluginPackageRepositoryFilter{})
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_repository_not_found" && api.Code != "plugin_package_path_invalid" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}

func TestListPluginPackages_FiltersAndSummary(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	resp, err := svc.ListPluginPackages("plugins-local/repository-fixtures", PluginPackageRepositoryFilter{Status: "all", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListPluginPackages: %v", err)
	}
	if resp.Summary.Total == 0 {
		t.Fatalf("expected non-empty summary")
	}
	found := false
	for _, it := range resp.Items {
		if it.Code == "demo_notice" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected demo_notice in items, got %#v", resp.Items)
	}

	onlyBlocked, err := svc.ListPluginPackages("plugins-local/repository-fixtures", PluginPackageRepositoryFilter{Status: "blocked", Page: 1, PageSize: 50})
	if err != nil {
		t.Fatalf("ListPluginPackages blocked: %v", err)
	}
	for _, it := range onlyBlocked.Items {
		if it.Status != "blocked" {
			t.Fatalf("expected blocked only, got %#v", it)
		}
	}
}
