package service

import (
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func mustProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "VERSION")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("unable to locate project root")
	return ""
}

func TestDryRunPluginPackage_SafeExampleOK(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	res, err := svc.DryRunPluginPackage("examples/plugins/demo_notice")
	if err != nil {
		t.Fatalf("DryRunPluginPackage failed: %v", err)
	}
	if res.Package.Code != "demo_notice" {
		t.Fatalf("unexpected package code: %#v", res.Package)
	}
	if res.Status != "ok" && res.Status != "warning" {
		t.Fatalf("unexpected status: %q", res.Status)
	}
	if res.Checksum.Status != "ok" && res.Checksum.Status != "warning" {
		t.Fatalf("unexpected checksum status: %#v", res.Checksum)
	}
	if res.RiskReport.Level == "" {
		t.Fatalf("expected risk report")
	}
	if len(res.FileScan.DangerousFiles) != 0 {
		t.Fatalf("expected no dangerous files, got %#v", res.FileScan.DangerousFiles)
	}
	if !res.ManifestValidation.Valid {
		t.Fatalf("expected manifest validation ok, got %#v", res.ManifestValidation)
	}
}

func TestDryRunPluginPackage_RejectsTraversal(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, err := svc.DryRunPluginPackage("../examples/plugins/demo_notice")
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_path_invalid" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}

func TestDryRunPluginPackage_ManifestMissing(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	root := mustProjectRoot(t)
	// Ensure temp dir is under allowlisted .devhub/plugins
	dir := filepath.Join(root, ".devhub", "plugins", "pkg_no_manifest")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, ".devhub")) })

	_, err := svc.DryRunPluginPackage(".devhub/plugins/pkg_no_manifest")
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_manifest_missing" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}

func TestDryRunPluginPackage_DangerousFileBlocked(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	root := mustProjectRoot(t)
	pkgDir := filepath.Join(root, ".devhub", "plugins", "pkg_danger")
	_ = os.RemoveAll(pkgDir)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, ".devhub")) })

	manifest := []byte(`{"code":"pkg_danger","name":"Danger","version":"1.0.0","compatible_core_version":">=1.4.0","is_system":false,"content_types":["danger_item"],"content_type_definitions":[{"type":"danger_item","name":"Danger Item","plugin_code":"pkg_danger","create_permission":"pkg_danger.item.create","edit_permission":"pkg_danger.item.edit","delete_permission":"pkg_danger.item.delete","audit_permission":"pkg_danger.item.audit","default_status":"draft","allow_comment":true,"allow_like":true,"allow_favorite":true,"seo_type":"Article"}],"permissions":[{"code":"pkg_danger.item.create","name":"create","scope":"community"},{"code":"pkg_danger.item.edit","name":"edit","scope":"own"},{"code":"pkg_danger.item.delete","name":"delete","scope":"own"},{"code":"pkg_danger.item.audit","name":"audit","scope":"community"}],"menus":[{"code":"pkg_danger.admin","title":"danger","path":"/admin-next/pkg_danger","location":"admin","area":"admin","permission":"pkg_danger.item.audit"}],"routes":[{"area":"admin","method":"GET","path":"/api/v1/admin/pkg_danger","handler":"reserved","auth":"admin","permission":"pkg_danger.item.audit"}]}`)
	if err := os.WriteFile(filepath.Join(pkgDir, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "hack.sh"), []byte("#!/bin/sh\necho hacked\n"), 0o644); err != nil {
		t.Fatalf("write sh: %v", err)
	}

	res, err := svc.DryRunPluginPackage(".devhub/plugins/pkg_danger")
	if err != nil {
		t.Fatalf("DryRunPluginPackage failed: %v", err)
	}
	if res.Status != "blocked" {
		t.Fatalf("expected blocked, got %q", res.Status)
	}
	if res.RiskReport.Level != "blocked" {
		t.Fatalf("expected risk blocked, got %#v", res.RiskReport)
	}
	if len(res.FileScan.DangerousFiles) == 0 {
		t.Fatalf("expected dangerous files")
	}
}
