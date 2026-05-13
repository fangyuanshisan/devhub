package service

import (
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestInstallPluginPackage_Success_DisabledAndSource(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	resp, err := svc.InstallPluginPackage(domain.PluginPackageInstallRequest{
		Path:             "plugins-local/repository-fixtures/demo_notice_install",
		ConfirmRiskLevel: "low",
	})
	if err != nil {
		t.Fatalf("InstallPluginPackage: %v", err)
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
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, err := svc.InstallPluginPackage(domain.PluginPackageInstallRequest{Path: "plugins-local/repository-fixtures/demo_notice_install"})
	if err != nil {
		t.Fatalf("first install: %v", err)
	}
	_, err = svc.InstallPluginPackage(domain.PluginPackageInstallRequest{Path: "plugins-local/repository-fixtures/demo_notice_install"})
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
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, err := svc.InstallPluginPackage(domain.PluginPackageInstallRequest{Path: "plugins-local/repository-fixtures/dangerous_shell"})
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
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, err := svc.InstallPluginPackage(domain.PluginPackageInstallRequest{Path: "plugins-local/repository-fixtures/checksum_mismatch"})
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
