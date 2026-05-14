package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestPluginPackageExport_DryRunDoesNotWrite(t *testing.T) {
	svc := New(store.NewMemoryStore())
	installExportTestPlugin(t, svc)
	root, _ := serviceProjectRoot()
	output := "storage/plugins/exports/export_test_plugin-dryrun-" + time.Now().UTC().Format("20060102150405")
	outputAbs := filepath.Join(root, filepath.FromSlash(output))
	t.Cleanup(func() { _ = os.RemoveAll(outputAbs) })

	res, err := svc.DryRunPluginPackageExport("export_test_plugin", domain.PluginPackageExportRequest{
		IncludeDocs:       true,
		IncludeMigrations: true,
		OutputDir:         output,
	})
	if err != nil {
		t.Fatalf("DryRunPluginPackageExport: %v", err)
	}
	if res.Status != "ok" {
		t.Fatalf("unexpected status: %s", res.Status)
	}
	if _, err := os.Stat(outputAbs); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create output dir, stat err=%v", err)
	}
	if !containsStringForTest(res.ExportPreview.Files, "manifest.json") || !containsStringForTest(res.ExportPreview.Files, "checksums.json") {
		t.Fatalf("missing core export files: %#v", res.ExportPreview.Files)
	}
}

func TestPluginPackageExport_WritesPackageAndSanitizesConfig(t *testing.T) {
	svc := New(store.NewMemoryStore())
	installExportTestPlugin(t, svc)
	root, _ := serviceProjectRoot()
	output := "storage/plugins/exports/export_test_plugin-full-" + time.Now().UTC().Format("20060102150405")
	outputAbs := filepath.Join(root, filepath.FromSlash(output))
	t.Cleanup(func() { _ = os.RemoveAll(outputAbs) })

	res, err := svc.ExportPluginPackage("export_test_plugin", domain.PluginPackageExportRequest{
		IncludeDocs:       true,
		IncludeMigrations: true,
		OutputDir:         output,
	})
	if err != nil {
		t.Fatalf("ExportPluginPackage: %v", err)
	}
	if res.ChecksumStatus != "generated" {
		t.Fatalf("unexpected checksum status: %s", res.ChecksumStatus)
	}
	if res.PackageDryRunStatus == "blocked" || res.PackageDryRunStatus == "failed" {
		t.Fatalf("exported package dry-run should not be blocked, got %s warnings=%v", res.PackageDryRunStatus, res.Warnings)
	}

	for _, rel := range []string{"manifest.json", "README.md", "config.example.json", "checksums.json", "docs/usage.md", "migrations/exported_migrations.json"} {
		if _, err := os.Stat(filepath.Join(outputAbs, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}

	rawConfig, err := os.ReadFile(filepath.Join(outputAbs, "config.example.json"))
	if err != nil {
		t.Fatalf("read config.example.json: %v", err)
	}
	configText := string(rawConfig)
	if strings.Contains(configText, "enc:v1:") || strings.Contains(configText, "real-secret") {
		t.Fatalf("config.example.json leaked sensitive value: %s", configText)
	}
	if !strings.Contains(configText, "REPLACE_ME") {
		t.Fatalf("config.example.json should use REPLACE_ME for sensitive fields: %s", configText)
	}

	verifyExportChecksum(t, outputAbs, "manifest.json")
	verifyExportChecksum(t, outputAbs, "README.md")
	verifyExportChecksum(t, outputAbs, "config.example.json")

	dry, err := svc.DryRunPluginPackage(output)
	if err != nil {
		t.Fatalf("DryRunPluginPackage(export): %v", err)
	}
	if dry.Status == "blocked" {
		t.Fatalf("exported package should not be blocked: %#v", dry)
	}
}

func TestPluginPackageExport_OutputExistsRequiresForce(t *testing.T) {
	svc := New(store.NewMemoryStore())
	installExportTestPlugin(t, svc)
	root, _ := serviceProjectRoot()
	output := "storage/plugins/exports/export_test_plugin-exists-" + time.Now().UTC().Format("20060102150405")
	outputAbs := filepath.Join(root, filepath.FromSlash(output))
	t.Cleanup(func() { _ = os.RemoveAll(outputAbs) })

	if _, err := svc.ExportPluginPackage("export_test_plugin", domain.PluginPackageExportRequest{OutputDir: output}); err != nil {
		t.Fatalf("first export: %v", err)
	}
	_, err := svc.ExportPluginPackage("export_test_plugin", domain.PluginPackageExportRequest{OutputDir: output})
	if err == nil {
		t.Fatalf("expected output exists error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api.Code != "plugin_export_output_exists" {
		t.Fatalf("unexpected error: %T %v", err, err)
	}
	if _, err := svc.ExportPluginPackage("export_test_plugin", domain.PluginPackageExportRequest{OutputDir: output, Force: true}); err != nil {
		t.Fatalf("force export: %v", err)
	}
}

func installExportTestPlugin(t *testing.T, svc *Service) {
	t.Helper()
	raw := []byte(`{
  "code": "export_test_plugin",
  "name": "Export Test Plugin",
  "version": "1.0.0",
  "description": "Export test plugin.",
  "author": "DevHub Test",
  "compatible_core_version": ">=1.4.0",
  "content_types": ["export_test_item"],
  "content_type_definitions": [
    {
      "type": "export_test_item",
      "name": "Export Test Item",
      "plugin_code": "export_test_plugin",
      "create_permission": "export_test_plugin.item.create",
      "edit_permission": "export_test_plugin.item.edit",
      "delete_permission": "export_test_plugin.item.delete",
      "audit_permission": "export_test_plugin.item.audit",
      "default_status": "draft",
      "allow_comment": true,
      "allow_like": true,
      "allow_favorite": true,
      "seo_type": "Article"
    }
  ],
  "permissions": [
    { "plugin_code": "export_test_plugin", "code": "export_test_plugin.item.create", "name": "Create", "scope": "community" },
    { "plugin_code": "export_test_plugin", "code": "export_test_plugin.item.edit", "name": "Edit", "scope": "own" },
    { "plugin_code": "export_test_plugin", "code": "export_test_plugin.item.delete", "name": "Delete", "scope": "own" },
    { "plugin_code": "export_test_plugin", "code": "export_test_plugin.item.audit", "name": "Audit", "scope": "community" },
    { "plugin_code": "export_test_plugin", "code": "export_test_plugin.manage", "name": "Manage", "scope": "global" },
    { "plugin_code": "export_test_plugin", "code": "export_test_plugin.configure", "name": "Configure", "scope": "global" }
  ],
  "menus": [
    { "plugin_code": "export_test_plugin", "code": "export_test_plugin.admin", "title": "Export Test", "path": "/admin-next/export-test", "location": "admin", "area": "admin", "permission": "export_test_plugin.manage", "sort_order": 300 }
  ],
  "routes": [
    { "plugin_code": "export_test_plugin", "area": "admin", "method": "GET", "path": "/api/v1/admin/export-test", "handler": "reserved:export_test_plugin.list", "auth": "admin", "permission": "export_test_plugin.manage" }
  ],
  "config_schema": {
    "type": "object",
    "required": ["enabled", "app_secret"],
    "properties": {
      "enabled": { "type": "boolean", "default": true },
      "title_prefix": { "type": "string", "default": "[Export]" },
      "app_secret": { "type": "string", "format": "password", "default": "real-secret" },
      "oauth": {
        "type": "object",
        "properties": {
          "client_secret": { "type": "string", "x-sensitive": true, "default": "nested-secret" }
        }
      }
    }
  },
  "migrations": [
    { "plugin_code": "export_test_plugin", "migration_version": "1.0.0", "migration_name": "export_test_init", "direction": "up", "checksum": "sha256:declaration-only", "tables": ["export_test_items"], "rollback_supported": false, "description": "declaration only" }
  ]
}`)
	if _, _, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("InstallPluginManifest: %v", err)
	}
}

func verifyExportChecksum(t *testing.T, root string, rel string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "checksums.json"))
	if err != nil {
		t.Fatalf("read checksums: %v", err)
	}
	var parsed struct {
		Files []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"files"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse checksums: %v", err)
	}
	want := ""
	for _, item := range parsed.Files {
		if item.Path == rel {
			want = item.SHA256
			break
		}
	}
	if want == "" {
		t.Fatalf("checksum missing for %s", rel)
	}
	fileRaw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	sum := sha256.Sum256(fileRaw)
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("checksum mismatch for %s: got %s want %s", rel, got, want)
	}
}

func containsStringForTest(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
