package service

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestPluginPackageTemplatePreviewDoesNotWriteFiles(t *testing.T) {
	root := mustProjectRoot(t)
	code := "tpl_preview_test"
	pkgDir := filepath.Join(root, "storage", "plugins", "drafts", code)
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
	if res.Template.PackagePath != "storage/plugins/drafts/"+code {
		t.Fatalf("unexpected package path: %q", res.Template.PackagePath)
	}
	if _, err := os.Stat(pkgDir); !os.IsNotExist(err) {
		t.Fatalf("preview should not create directory, stat err=%v", err)
	}
	if len(res.Template.Files) == 0 {
		t.Fatalf("expected preview files")
	}
	if len(res.Template.Permissions) == 0 || len(res.Template.Menus) == 0 || len(res.Template.Migrations) != 1 {
		t.Fatalf("expected preview summary lists: %#v", res.Template)
	}
}

func TestPluginPackageTemplateCreateDryRunAndOmitsGoFile(t *testing.T) {
	root := mustProjectRoot(t)
	code := "tpl_create_test"
	pkgDir := filepath.Join(root, "storage", "plugins", "drafts", code)
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
	if res.Template.PackagePath != "storage/plugins/drafts/"+code {
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
	if _, err := os.Stat(filepath.Join(pkgDir, "migrations", "001_init.sql")); err != nil {
		t.Fatalf("migrations/001_init.sql should exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pkgDir, "001_schema.sql")); !os.IsNotExist(err) {
		t.Fatalf("root 001_schema.sql must not be generated, stat err=%v", err)
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

func TestPluginPackageTemplateCreateReusesEmptyDirectory(t *testing.T) {
	root := mustProjectRoot(t)
	code := "tpl_create_empty_dir_test"
	pkgDir := filepath.Join(root, "storage", "plugins", "drafts", code)
	_ = os.RemoveAll(pkgDir)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir empty package dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(pkgDir) })

	svc := New(store.NewMemoryStore())
	res, err := svc.CreatePluginPackageTemplate(domain.PluginPackageTemplateRequest{
		Code:        code,
		Name:        "Template Empty Dir Test",
		ContentType: "tpl_empty_item",
		ContentName: "空目录内容",
		Description: "template empty dir",
		Author:      "DevHub",
	})
	if err != nil {
		t.Fatalf("CreatePluginPackageTemplate with empty dir: %v", err)
	}
	if res.Template.PackagePath != "storage/plugins/drafts/"+code {
		t.Fatalf("unexpected package path: %q", res.Template.PackagePath)
	}
	if _, err := os.Stat(filepath.Join(pkgDir, "manifest.json")); err != nil {
		t.Fatalf("manifest should exist: %v", err)
	}
}

func TestPluginPackageTemplatePreviewTypesAndConflicts(t *testing.T) {
	root := mustProjectRoot(t)
	code := "tpl_front_mount_test"
	pkgDir := filepath.Join(root, "storage", "plugins", "drafts", code)
	_ = os.RemoveAll(pkgDir)
	t.Cleanup(func() { _ = os.RemoveAll(pkgDir) })

	svc := New(store.NewMemoryStore())
	res, err := svc.PreviewPluginPackageTemplate(domain.PluginPackageTemplateRequest{
		Code:         code,
		Name:         "Front Mount Test",
		PluginType:   "frontend_mount",
		Author:       "DevHub",
		WithConfig:   true,
		WithHooks:    true,
		MountPoint:   "admin.plugin.detail.preview",
		ComponentKey: "official.announcement.card",
	})
	if err != nil {
		t.Fatalf("PreviewPluginPackageTemplate frontend_mount: %v", err)
	}
	if res.Template.ContentType != "" || len(res.Template.FrontendMounts) != 1 {
		t.Fatalf("frontend_mount should hide content type and include allowlisted mount: %#v", res.Template)
	}

	conflict, err := svc.PreviewPluginPackageTemplate(domain.PluginPackageTemplateRequest{
		Code: "core",
		Name: "Reserved Core",
	})
	if err != nil {
		t.Fatalf("reserved preview should return conflicts, not fail: %v", err)
	}
	if len(conflict.Template.Conflicts) == 0 {
		t.Fatalf("expected reserved code conflict")
	}
}

func TestPluginPackageTemplateRejectsPathTraversalCode(t *testing.T) {
	svc := New(store.NewMemoryStore())
	_, err := svc.CreatePluginPackageTemplate(domain.PluginPackageTemplateRequest{
		Code:        "../evil",
		Name:        "Traversal",
		ContentType: "evil_item",
		ContentName: "恶意内容",
	})
	if err == nil {
		t.Fatal("expected path traversal code to be rejected")
	}
}

func TestPluginPackageTemplateExportZipOmitsExecutableCode(t *testing.T) {
	svc := New(store.NewMemoryStore())
	data, name, preview, err := svc.ExportPluginPackageTemplateZip(domain.PluginPackageTemplateRequest{
		Code:          "tpl_export_test",
		Name:          "Template Export Test",
		ContentType:   "tpl_export_item",
		ContentName:   "导出内容",
		Author:        "DevHub",
		WithConfig:    true,
		WithHooks:     true,
		WithMigration: true,
	})
	if err != nil {
		t.Fatalf("ExportPluginPackageTemplateZip: %v", err)
	}
	if name != "tpl_export_test.zip" || len(data) == 0 || preview.Code != "tpl_export_test" {
		t.Fatalf("unexpected export response: name=%q size=%d preview=%#v", name, len(data), preview)
	}
	if bytes.Contains(data, []byte("001_schema.sql")) || bytes.Contains(data, []byte(".go")) || bytes.Contains(data, []byte(".wasm")) {
		t.Fatalf("export zip contains forbidden file marker")
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open export zip: %v", err)
	}
	entries := map[string]bool{}
	for _, item := range zr.File {
		if strings.Contains(item.Name, "../") || strings.HasPrefix(item.Name, "/") {
			t.Fatalf("export zip contains unsafe entry: %s", item.Name)
		}
		entries[item.Name] = true
	}
	for _, name := range []string{
		"tpl_export_test/manifest.json",
		"tpl_export_test/docs/usage.md",
		"tpl_export_test/docs/content-types.md",
		"tpl_export_test/docs/permissions.md",
		"tpl_export_test/docs/hooks.md",
		"tpl_export_test/docs/migrations.md",
		"tpl_export_test/docs/registry-example.md",
		"tpl_export_test/migrations/001_init.sql",
	} {
		if !entries[name] {
			t.Fatalf("export zip missing %s, entries=%v", name, entries)
		}
	}
}
