package service

import (
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestInstallPluginPackage_Success_DisabledAndSource(t *testing.T) {
	path := prepareInstallRepositoryFixture(t, "demo_notice_install_success")
	repo := store.NewMemoryStore()
	svc := New(repo)
	dry, err := svc.DryRunPluginPackage(path)
	if err != nil {
		t.Fatalf("DryRunPluginPackage: %v", err)
	}

	resp, err := svc.InstallPluginPackage(PluginOperationOperator{ID: 1, Name: "tester"}, domain.PluginPackageInstallRequest{
		Path:             path,
		DryRunID:         dry.DryRunID,
		ConfirmRiskLevel: "medium",
	})
	if err != nil {
		t.Fatalf("InstallPluginPackage: %v", err)
	}
	if resp.OperationID == "" {
		t.Fatalf("expected operation_id")
	}
	if resp.Plugin.Code != "demo_notice_install" {
		t.Fatalf("unexpected plugin code: %q", resp.Plugin.Code)
	}
	if resp.Plugin.Status != "disabled" {
		t.Fatalf("expected disabled, got %q", resp.Plugin.Status)
	}
	if resp.Plugin.SourceType != "local_package" {
		t.Fatalf("expected source_type local_package, got %q", resp.Plugin.SourceType)
	}
	if resp.Plugin.ManifestChecksum == "" {
		t.Fatalf("expected manifest checksum")
	}
	if resp.Plugin.PackageChecksum == "" {
		t.Fatalf("expected package checksum (manifest.json sha256)")
	}
	if _, ok := repo.PluginByCode("demo_notice_install"); !ok {
		t.Fatalf("expected plugin saved")
	}
}

func TestInstallPluginPackage_AlreadyInstalled(t *testing.T) {
	path := prepareInstallRepositoryFixture(t, "demo_notice_install_duplicate")
	repo := store.NewMemoryStore()
	svc := New(repo)
	dry, err := svc.DryRunPluginPackage(path)
	if err != nil {
		t.Fatalf("DryRunPluginPackage: %v", err)
	}

	_, err = svc.InstallPluginPackage(PluginOperationOperator{ID: 1, Name: "tester"}, domain.PluginPackageInstallRequest{
		Path:             path,
		DryRunID:         dry.DryRunID,
		ConfirmRiskLevel: "medium",
	})
	if err != nil {
		t.Fatalf("first install: %v", err)
	}
	dry, err = svc.DryRunPluginPackage(path)
	if err != nil {
		t.Fatalf("second dry-run: %v", err)
	}
	_, err = svc.InstallPluginPackage(PluginOperationOperator{ID: 1, Name: "tester"}, domain.PluginPackageInstallRequest{
		Path:             path,
		DryRunID:         dry.DryRunID,
		ConfirmRiskLevel: "medium",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_already_installed" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}

func TestInstallPluginPackage_Blocked_DangerousFile(t *testing.T) {
	path := prepareInstallRepositoryFixtureFrom(t, "plugins-local/repository-fixtures/dangerous_shell", "dangerous_shell_blocked")
	repo := store.NewMemoryStore()
	svc := New(repo)

	dry, dryErr := svc.DryRunPluginPackage(path)
	dryRunID := ""
	if dryErr == nil {
		dryRunID = dry.DryRunID
	}
	_, err := svc.InstallPluginPackage(PluginOperationOperator{ID: 1, Name: "tester"}, domain.PluginPackageInstallRequest{Path: path, DryRunID: dryRunID})
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_dangerous_file" && api.Code != "plugin_package_install_blocked" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
	if _, ok := repo.PluginByCode("demo_notice_install"); ok {
		t.Fatalf("should not install any plugin")
	}
}

func TestInstallPluginPackage_Blocked_ChecksumMismatch(t *testing.T) {
	path := prepareInstallRepositoryFixtureFrom(t, "plugins-local/repository-fixtures/checksum_mismatch", "checksum_mismatch_blocked")
	repo := store.NewMemoryStore()
	svc := New(repo)

	dry, dryErr := svc.DryRunPluginPackage(path)
	dryRunID := ""
	if dryErr == nil {
		dryRunID = dry.DryRunID
	}
	_, err := svc.InstallPluginPackage(PluginOperationOperator{ID: 1, Name: "tester"}, domain.PluginPackageInstallRequest{Path: path, DryRunID: dryRunID})
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_checksum_mismatch" && api.Code != "plugin_package_install_blocked" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}

func TestInstallPluginPackage_RequiresRepositoryAndDryRunID(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	if _, err := svc.InstallPluginPackage(PluginOperationOperator{ID: 1, Name: "tester"}, domain.PluginPackageInstallRequest{
		Path:             "plugins-local/repository-fixtures/demo_notice_install",
		ConfirmRiskLevel: "medium",
	}); err == nil {
		t.Fatalf("expected non-repository path to be rejected")
	} else if api, ok := err.(*domain.APIError); !ok || api.Code != "plugin_package_install_source_invalid" {
		t.Fatalf("unexpected source error: %T %v", err, err)
	}

	path := prepareInstallRepositoryFixture(t, "demo_notice_install_require_dryrun")
	if _, err := svc.InstallPluginPackage(PluginOperationOperator{ID: 1, Name: "tester"}, domain.PluginPackageInstallRequest{
		Path:             path,
		ConfirmRiskLevel: "medium",
	}); err == nil {
		t.Fatalf("expected missing dry-run id to be rejected")
	} else if api, ok := err.(*domain.APIError); !ok || api.Code != "plugin_package_install_dry_run_required" {
		t.Fatalf("unexpected dry-run error: %T %v", err, err)
	}
}

func TestInstallPluginPackage_DryRunIDMismatch(t *testing.T) {
	path := prepareInstallRepositoryFixture(t, "demo_notice_install_mismatch")
	repo := store.NewMemoryStore()
	svc := New(repo)
	dry, err := svc.DryRunPluginPackage(path)
	if err != nil {
		t.Fatalf("DryRunPluginPackage: %v", err)
	}
	root := mustProjectRoot(t)
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(path), "README.md"), []byte("changed after dry-run\n"), 0o644); err != nil {
		t.Fatalf("mutate package: %v", err)
	}
	_, err = svc.InstallPluginPackage(PluginOperationOperator{ID: 1, Name: "tester"}, domain.PluginPackageInstallRequest{
		Path:             path,
		DryRunID:         dry.DryRunID,
		ConfirmRiskLevel: "medium",
	})
	if err == nil {
		t.Fatalf("expected dry-run mismatch")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_install_dry_run_mismatch" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}

func prepareInstallRepositoryFixture(t *testing.T, name string) string {
	t.Helper()
	return prepareInstallRepositoryFixtureFrom(t, "plugins-local/repository-fixtures/demo_notice_install", name)
}

func prepareInstallRepositoryFixtureFrom(t *testing.T, src, name string) string {
	t.Helper()
	root := mustProjectRoot(t)
	dst := filepath.Join("storage", "plugins", "packages", name)
	dstAbs := filepath.Join(root, filepath.FromSlash(dst))
	if err := os.RemoveAll(dstAbs); err != nil {
		t.Fatalf("cleanup fixture: %v", err)
	}
	if err := copyPackageTree(filepath.Join(root, filepath.FromSlash(src)), dstAbs); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dstAbs) })
	return filepath.ToSlash(dst)
}
