package service

import (
	"bytes"
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
	baseRel := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-repo")
	dir := filepath.Join(root, filepath.FromSlash(baseRel), "repo_empty")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(baseRel))) })

	resp, err := svc.ListPluginPackages(filepath.ToSlash(filepath.Join(baseRel, "repo_empty")), PluginPackageRepositoryFilter{})
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

func TestDeletePluginPackageRepositoryPackage_PromotedUninstalledOnly(t *testing.T) {
	_ = cleanupUploadStorage(t)
	root := mustProjectRoot(t)
	code := "demo_notice_repo_delete"
	target := filepath.Join(root, "storage", "plugins", "packages", code)
	_ = os.RemoveAll(target)
	t.Cleanup(func() { _ = os.RemoveAll(target) })

	repo := store.NewMemoryStore()
	svc := New(repo)
	raw := zipMinimalPackageForUpload(t, code, "")
	upload, err := svc.UploadPluginPackageZip(code+".zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip: %v", err)
	}
	promoted, err := svc.PromotePluginPackageUpload(upload.UploadID, false)
	if err != nil {
		t.Fatalf("PromotePluginPackageUpload: %v", err)
	}
	preview, err := svc.CleanupPluginPackageRepository(domain.PluginPackageCleanupRequest{DryRun: true, Statuses: []string{"ok", "warning"}})
	if err != nil {
		t.Fatalf("CleanupPluginPackageRepository dry-run: %v", err)
	}
	if preview.ConfirmToken == "" {
		t.Fatalf("expected confirm token: %#v", preview)
	}
	deleted, err := svc.DeletePluginPackageRepositoryPackage(promoted.PackagePath)
	if err != nil {
		t.Fatalf("DeletePluginPackageRepositoryPackage: %v", err)
	}
	if deleted.DeletedCount != 1 {
		t.Fatalf("expected deleted package: %#v", deleted)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected repository package removed, stat err=%v", err)
	}
}

func TestDeletePluginPackageRepositoryPackage_InstalledBlocked(t *testing.T) {
	_ = cleanupUploadStorage(t)
	root := mustProjectRoot(t)
	code := "demo_notice_repo_installed"
	target := filepath.Join(root, "storage", "plugins", "packages", code)
	_ = os.RemoveAll(target)
	t.Cleanup(func() { _ = os.RemoveAll(target) })

	repo := store.NewMemoryStore()
	svc := New(repo)
	raw := zipMinimalPackageForUpload(t, code, "")
	upload, err := svc.UploadPluginPackageZip(code+".zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip: %v", err)
	}
	promoted, err := svc.PromotePluginPackageUpload(upload.UploadID, false)
	if err != nil {
		t.Fatalf("PromotePluginPackageUpload: %v", err)
	}
	_, _ = repo.SavePlugin(domain.Plugin{PluginManifest: domain.PluginManifest{Code: code, Version: "1.0.0"}, Status: "disabled", SourceType: "local_package"})
	if _, err := svc.DeletePluginPackageRepositoryPackage(promoted.PackagePath); apiCodeForUpload(err) != "plugin_package_repository_installed" {
		t.Fatalf("expected installed guard, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "manifest.json")); err != nil {
		t.Fatalf("installed package must remain: %v", err)
	}
}

func TestCleanupPluginPackageRepository_TestPackagesPreviewExecuteAndSkipsInstalled(t *testing.T) {
	_ = cleanupUploadStorage(t)
	root := mustProjectRoot(t)
	codes := []string{"s8cleanup_e2e_upload_promote_cleanup", "s8cleanup_fixture_cleanup_installed"}
	for _, code := range codes {
		target := filepath.Join(root, "storage", "plugins", "packages", code)
		_ = os.RemoveAll(target)
		t.Cleanup(func() { _ = os.RemoveAll(target) })
	}

	repo := store.NewMemoryStore()
	svc := New(repo)
	raw := zipMinimalPackageForUpload(t, codes[0], "")
	upload, err := svc.UploadPluginPackageZip(codes[0]+".zip", int64(len(raw)), bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip cleanup target: %v", err)
	}
	promoted, err := svc.PromotePluginPackageUpload(upload.UploadID, false)
	if err != nil {
		t.Fatalf("PromotePluginPackageUpload cleanup target: %v", err)
	}
	rawInstalled := zipMinimalPackageForUpload(t, codes[1], "")
	uploadInstalled, err := svc.UploadPluginPackageZip(codes[1]+".zip", int64(len(rawInstalled)), bytes.NewReader(rawInstalled))
	if err != nil {
		t.Fatalf("UploadPluginPackageZip installed: %v", err)
	}
	promotedInstalled, err := svc.PromotePluginPackageUpload(uploadInstalled.UploadID, false)
	if err != nil {
		t.Fatalf("PromotePluginPackageUpload installed: %v", err)
	}
	_, _ = repo.SavePlugin(domain.Plugin{PluginManifest: domain.PluginManifest{Code: codes[1], Version: "1.0.0"}, Status: "enabled", SourceType: "local_package"})

	req := domain.PluginPackageCleanupRequest{Scope: "test_packages", Prefixes: []string{"s8cleanup_"}}
	preview, err := svc.PreviewPluginPackageRepositoryCleanup(req)
	if err != nil {
		t.Fatalf("PreviewPluginPackageRepositoryCleanup: %v", err)
	}
	if preview.WillDeleteCount == 0 || preview.ConfirmToken == "" {
		t.Fatalf("expected cleanup preview candidate: %#v", preview)
	}
	if preview.SkippedCount == 0 {
		t.Fatalf("expected installed package skipped: %#v", preview)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(promoted.PackagePath), "manifest.json")); err != nil {
		t.Fatalf("preview must not delete package: %v", err)
	}

	req.ConfirmToken = preview.ConfirmToken
	cleaned, err := svc.CleanupPluginPackageRepository(req)
	if err != nil {
		t.Fatalf("CleanupPluginPackageRepository: %v", err)
	}
	if cleaned.DeletedCount != 1 || cleaned.SkippedCount == 0 {
		t.Fatalf("expected one deleted and installed skipped: %#v", cleaned)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(promoted.PackagePath))); !os.IsNotExist(err) {
		t.Fatalf("expected test package storage deleted, stat err=%v", err)
	}
	if _, ok := repo.PluginPackageUploadByUploadID(upload.UploadID); ok {
		t.Fatalf("expected promoted upload record removed with local repository package")
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(promotedInstalled.PackagePath), "manifest.json")); err != nil {
		t.Fatalf("installed package must remain: %v", err)
	}
	if _, ok := repo.PluginPackageUploadByUploadID(uploadInstalled.UploadID); !ok {
		t.Fatalf("installed package upload record must remain")
	}
}

func TestDeletePluginPackageRepositoryPackage_InvalidUnpromotedAllowed(t *testing.T) {
	root := mustProjectRoot(t)
	baseRel := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-repo-delete-invalid")
	pathRel := filepath.ToSlash(filepath.Join(baseRel, "invalid_unpromoted"))
	pathAbs := filepath.Join(root, filepath.FromSlash(pathRel))
	_ = os.RemoveAll(pathAbs)
	if err := os.MkdirAll(pathAbs, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pathAbs, "README.md"), []byte("invalid test package"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(baseRel))) })

	svc := New(store.NewMemoryStore())
	deleted, err := svc.DeletePluginPackageRepositoryPackage(pathRel)
	if err != nil {
		t.Fatalf("DeletePluginPackageRepositoryPackage invalid unpromoted: %v", err)
	}
	if deleted.DeletedCount != 1 {
		t.Fatalf("expected deleted invalid package: %#v", deleted)
	}
	if _, err := os.Stat(pathAbs); !os.IsNotExist(err) {
		t.Fatalf("expected invalid package removed, stat err=%v", err)
	}
}

func TestDeletePluginPackageRepositoryPackage_ValidUnpromotedUninstalledAllowed(t *testing.T) {
	root := mustProjectRoot(t)
	baseRel := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-repo-delete-valid")
	pathRel := filepath.ToSlash(filepath.Join(baseRel, "valid_unpromoted"))
	pathAbs := filepath.Join(root, filepath.FromSlash(pathRel))
	_ = os.RemoveAll(pathAbs)
	if err := os.MkdirAll(pathAbs, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	manifest := `{"code":"valid_unpromoted","name":"Valid Unpromoted","version":"1.0.0","description":"test","author":"DevHub","content_types":[],"permissions":[],"menus":[],"hooks":[]}`
	if err := os.WriteFile(filepath.Join(pathAbs, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(baseRel))) })

	svc := New(store.NewMemoryStore())
	deleted, err := svc.DeletePluginPackageRepositoryPackage(pathRel)
	if err != nil {
		t.Fatalf("DeletePluginPackageRepositoryPackage valid unpromoted: %v", err)
	}
	if deleted.DeletedCount != 1 {
		t.Fatalf("expected deleted valid package: %#v", deleted)
	}
	if _, err := os.Stat(pathAbs); !os.IsNotExist(err) {
		t.Fatalf("expected valid package removed, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(baseRel))); !os.IsNotExist(err) {
		t.Fatalf("expected empty parent removed, stat err=%v", err)
	}
}

func TestListPluginPackages_RepoNotFound(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, err := svc.ListPluginPackages("storage/plugins/packages/repo_not_exist", PluginPackageRepositoryFilter{})
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
