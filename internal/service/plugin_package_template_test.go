package service

import (
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestPluginPackageTemplatePreviewDoesNotWriteFiles(t *testing.T) {
	root := mustProjectRoot(t)
	code := "tpl_preview_test"
	pkgDir := filepath.Join(root, "storage", "plugins", "packages", code)
	_ = os.RemoveAll(pkgDir)
	t.Cleanup(func() { _ = os.RemoveAll(pkgDir) })

	svc := New(store.NewMemoryStore())
	res, err := svc.PreviewPluginPackageTemplate(domain.PluginPackageTemplateRequest{
		Code:          code,
		Name:          "Template Preview Test",
		ContentType:   "tpl_preview_item",
		ContentName:   "预览内容",
		Description:   "template preview",
		Author:        "DevHub",
		WithConfig:    true,
		WithHooks:     true,
		WithMigration: true,
	})
	if err != nil {
		t.Fatalf("PreviewPluginPackageTemplate: %v", err)
	}
	if res.Template.PackagePath != "storage/plugins/packages/"+code {
		t.Fatalf("unexpected package path: %q", res.Template.PackagePath)
	}
	if _, err := os.Stat(pkgDir); !os.IsNotExist(err) {
		t.Fatalf("preview should not create directory, stat err=%v", err)
	}
	if len(res.Template.Files) == 0 {
		t.Fatalf("expected preview files")
	}
}

func TestPluginPackageTemplateCreateDryRunAndOmitsGoFile(t *testing.T) {
	root := mustProjectRoot(t)
	code := "tpl_create_test"
	pkgDir := filepath.Join(root, "storage", "plugins", "packages", code)
	_ = os.RemoveAll(pkgDir)
	t.Cleanup(func() { _ = os.RemoveAll(pkgDir) })

	svc := New(store.NewMemoryStore())
	res, err := svc.CreatePluginPackageTemplate(domain.PluginPackageTemplateRequest{
		Code:          code,
		Name:          "Template Create Test",
		ContentType:   "tpl_create_item",
		ContentName:   "创建内容",
		Description:   "template create",
		Author:        "DevHub",
		WithConfig:    true,
		WithHooks:     true,
		WithMigration: true,
	})
	if err != nil {
		t.Fatalf("CreatePluginPackageTemplate: %v", err)
	}
	if res.Template.PackagePath != "storage/plugins/packages/"+code {
		t.Fatalf("unexpected package path: %q", res.Template.PackagePath)
	}
	if _, err := os.Stat(filepath.Join(pkgDir, "manifest.json")); err != nil {
		t.Fatalf("manifest should exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pkgDir, "registry.example.go")); !os.IsNotExist(err) {
		t.Fatalf("admin template must not generate registry.example.go, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(pkgDir, "docs", "registry-example.md")); err != nil {
		t.Fatalf("registry docs should exist: %v", err)
	}
	if res.DryRun.Status == "blocked" {
		t.Fatalf("template dry-run should not be blocked: %#v", res.DryRun)
	}
	if len(res.DryRun.FileScan.DangerousFiles) != 0 {
		t.Fatalf("expected no dangerous files: %#v", res.DryRun.FileScan.DangerousFiles)
	}
	if !res.DryRun.ManifestValidation.Valid {
		t.Fatalf("manifest should validate: %#v", res.DryRun.ManifestValidation)
	}
}
