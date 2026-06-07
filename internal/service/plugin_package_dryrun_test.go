package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
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

func TestDryRunPluginPackage_OfficialWebhookNotifyExampleOK(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	res, err := svc.DryRunPluginPackage("examples/plugins/official_webhook_notify")
	if err != nil {
		t.Fatalf("DryRunPluginPackage official_webhook_notify failed: %v", err)
	}
	if strings.EqualFold(res.Status, "blocked") {
		t.Fatalf("official_webhook_notify should not be blocked: %#v", res)
	}
	if res.Package.Code != "official_webhook_notify" {
		t.Fatalf("unexpected package code: %#v", res.Package)
	}
	found := false
	for _, hook := range res.InstallDryRun.NormalizedManifest.Hooks {
		if hook.Name == "AfterCreateContent" && hook.ServiceType == "external_service" && hook.Mode == string(pluginregistry.HookNonBlocking) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected external_service non-blocking AfterCreateContent hook: %#v", res.InstallDryRun.NormalizedManifest.Hooks)
	}
	if res.InstallDryRun.NormalizedManifest.ExternalService == nil {
		t.Fatalf("expected official_webhook_notify external_service metadata")
	}
	if !stringSliceContains(res.InstallDryRun.NormalizedManifest.ExternalService.SubscribedHooks, pluginregistry.HookAfterCreateContent) {
		t.Fatalf("expected official_webhook_notify to subscribe AfterCreateContent, got %#v", res.InstallDryRun.NormalizedManifest.ExternalService)
	}
}

func TestDryRunPluginPackage_OfficialLinksExampleOK(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	res, err := svc.DryRunPluginPackage("examples/plugins/official_links")
	if err != nil {
		t.Fatalf("DryRunPluginPackage official_links failed: %v", err)
	}
	if strings.EqualFold(res.Status, "blocked") {
		t.Fatalf("official_links should not be blocked: %#v", res)
	}
	if res.Package.Code != "official_links" {
		t.Fatalf("unexpected package code: %#v", res.Package)
	}
	if !res.ManifestValidation.Valid {
		t.Fatalf("expected manifest validation ok, got %#v", res.ManifestValidation)
	}
	if len(res.FileScan.DangerousFiles) != 0 {
		t.Fatalf("expected no dangerous files, got %#v", res.FileScan.DangerousFiles)
	}
	if len(res.MigrationPlan) != 1 || res.MigrationPlan[0].Path != "migrations/001_init.sql" {
		t.Fatalf("expected official_links no-op migration plan, got %#v", res.MigrationPlan)
	}
	if res.MigrationPlan[0].WillExecute {
		t.Fatalf("dry-run migration plan must not execute SQL: %#v", res.MigrationPlan[0])
	}
	foundContentType := false
	for _, def := range res.InstallDryRun.NormalizedManifest.ContentTypeDefs {
		if def.Type == "friend_link" && def.PluginCode == "official_links" {
			foundContentType = true
		}
	}
	if !foundContentType {
		t.Fatalf("expected friend_link content type definition: %#v", res.InstallDryRun.NormalizedManifest.ContentTypeDefs)
	}
	wantPerms := map[string]bool{
		"official_links.menu.view":     false,
		"official_links.link.create":   false,
		"official_links.link.manage":   false,
		"official_links.config.manage": false,
	}
	for _, perm := range res.InstallDryRun.NormalizedManifest.Permissions {
		if _, ok := wantPerms[perm.Code]; ok {
			wantPerms[perm.Code] = true
		}
	}
	for code, ok := range wantPerms {
		if !ok {
			t.Fatalf("expected permission %s in manifest permissions: %#v", code, res.InstallDryRun.NormalizedManifest.Permissions)
		}
	}
}

func TestDryRunPluginPackage_OfficialTemplatesOK(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		code            string
		wantWebhookHook bool
	}{
		{
			name: "declarative content template",
			path: "examples/plugins/templates/declarative-content",
			code: "official_links_template",
		},
		{
			name:            "external service webhook template",
			path:            "examples/plugins/templates/external-service-webhook",
			code:            "official_webhook_notify_template",
			wantWebhookHook: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := store.NewMemoryStore()
			svc := New(repo)

			res, err := svc.DryRunPluginPackage(tt.path)
			if err != nil {
				t.Fatalf("DryRunPluginPackage(%s) failed: %v", tt.path, err)
			}
			if strings.EqualFold(res.Status, "blocked") {
				t.Fatalf("template should not be blocked: %#v", res)
			}
			if res.Package.Code != tt.code {
				t.Fatalf("unexpected package code: %#v", res.Package)
			}
			if len(res.FileScan.DangerousFiles) != 0 {
				t.Fatalf("expected no dangerous files, got %#v", res.FileScan.DangerousFiles)
			}
			if !res.ManifestValidation.Valid {
				t.Fatalf("expected manifest validation ok, got %#v", res.ManifestValidation)
			}
			for _, hook := range res.InstallDryRun.NormalizedManifest.Hooks {
				if hook.Mode == string(pluginregistry.HookBlocking) || hook.Blocking {
					t.Fatalf("official templates must not declare blocking hooks: %#v", hook)
				}
				if hook.ServiceType == "external_service" && hook.Mode != string(pluginregistry.HookNonBlocking) {
					t.Fatalf("external_service hooks must be non-blocking: %#v", hook)
				}
			}
			hasWebhookHook := false
			for _, hook := range res.InstallDryRun.NormalizedManifest.Hooks {
				if hook.Name == "AfterCreateContent" && hook.ServiceType == "external_service" && hook.Mode == string(pluginregistry.HookNonBlocking) {
					hasWebhookHook = true
				}
			}
			if tt.wantWebhookHook && !hasWebhookHook {
				t.Fatalf("expected external_service non-blocking AfterCreateContent hook: %#v", res.InstallDryRun.NormalizedManifest.Hooks)
			}
			if tt.wantWebhookHook {
				if res.InstallDryRun.NormalizedManifest.ExternalService == nil {
					t.Fatalf("expected template external_service metadata")
				}
				if !stringSliceContains(res.InstallDryRun.NormalizedManifest.ExternalService.SubscribedHooks, pluginregistry.HookAfterCreateContent) {
					t.Fatalf("expected template to subscribe AfterCreateContent, got %#v", res.InstallDryRun.NormalizedManifest.ExternalService)
				}
			}
		})
	}
}

func stringSliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func TestConfigSchemaPatternsAreJSCompatible(t *testing.T) {
	wantPathPattern := `^/(?:[^/\s][^\s]*)?$`

	announcement, ok := pluginregistry.DefinitionByCode("official_announcement")
	if !ok {
		t.Fatal("official_announcement definition missing")
	}
	annSchema, ok := announcement.ConfigSchema.(map[string]any)
	if !ok {
		t.Fatalf("official_announcement config_schema should be an object: %#v", announcement.ConfigSchema)
	}
	annProps, _ := annSchema["properties"].(map[string]any)
	linkURL, _ := annProps["link_url"].(map[string]any)
	if got := linkURL["pattern"]; got != wantPathPattern {
		t.Fatalf("official_announcement link_url pattern = %v, want %q", got, wantPathPattern)
	}
	if err := pluginregistry.ValidateConfigJSON(announcement, `{"enabled":true,"message":"公告","link_text":"详情","link_url":"/news","dismissible":false}`); err != nil {
		t.Fatalf("official_announcement config should validate: %v", err)
	}
	if err := pluginregistry.ValidateConfigJSON(announcement, `{"enabled":true,"message":"公告","link_text":"详情","link_url":"https://evil.example/","dismissible":false}`); err == nil {
		t.Fatal("official_announcement should reject remote link_url")
	}

	root := mustProjectRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "examples/plugins/templates/external-service-webhook/manifest.json"))
	if err != nil {
		t.Fatalf("read official webhook template manifest: %v", err)
	}
	var manifest domain.PluginManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("unmarshal official webhook template manifest: %v", err)
	}
	schema, ok := manifest.ConfigSchema.(map[string]any)
	if !ok {
		t.Fatalf("official webhook template config_schema should be an object: %#v", manifest.ConfigSchema)
	}
	props, _ := schema["properties"].(map[string]any)
	healthCheck, _ := props["health_check_path"].(map[string]any)
	if got := healthCheck["pattern"]; got != wantPathPattern {
		t.Fatalf("health_check_path pattern = %v, want %q", got, wantPathPattern)
	}
	if err := pluginregistry.ValidateConfigJSON(domain.Plugin{PluginManifest: manifest}, `{"enabled":true,"endpoint_url":"http://127.0.0.1:18090","health_check_path":"/health","timeout_ms":3000,"failure_policy":"warn","auth_type":"none","warning_threshold":3,"error_threshold":5}`); err != nil {
		t.Fatalf("official webhook template config should validate: %v", err)
	}
	if err := pluginregistry.ValidateConfigJSON(domain.Plugin{PluginManifest: manifest}, `{"enabled":true,"endpoint_url":"http://127.0.0.1:18090","health_check_path":" /health","timeout_ms":3000,"failure_policy":"warn","auth_type":"none","warning_threshold":3,"error_threshold":5}`); err == nil {
		t.Fatal("health_check_path with leading space should be rejected")
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
	baseRel := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-pkg-dryrun")
	dir := filepath.Join(root, filepath.FromSlash(baseRel), "pkg_no_manifest")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(baseRel))) })

	_, err := svc.DryRunPluginPackage(filepath.ToSlash(filepath.Join(baseRel, "pkg_no_manifest")))
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
	baseRel := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-pkg-dryrun")
	pkgDir := filepath.Join(root, filepath.FromSlash(baseRel), "pkg_danger")
	_ = os.RemoveAll(pkgDir)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(baseRel))) })

	manifest := []byte(`{"code":"pkg_danger","name":"Danger","version":"1.0.0","compatible_core_version":">=1.4.0","is_system":false,"content_types":["danger_item"],"content_type_definitions":[{"type":"danger_item","name":"Danger Item","plugin_code":"pkg_danger","create_permission":"pkg_danger.item.create","edit_permission":"pkg_danger.item.edit","delete_permission":"pkg_danger.item.delete","audit_permission":"pkg_danger.item.audit","default_status":"draft","allow_comment":true,"allow_like":true,"allow_favorite":true,"seo_type":"Article"}],"permissions":[{"code":"pkg_danger.item.create","name":"create","scope":"community"},{"code":"pkg_danger.item.edit","name":"edit","scope":"own"},{"code":"pkg_danger.item.delete","name":"delete","scope":"own"},{"code":"pkg_danger.item.audit","name":"audit","scope":"community"}],"menus":[{"code":"pkg_danger.admin","title":"danger","path":"/admin-next/pkg_danger","location":"admin","area":"admin","permission":"pkg_danger.item.audit"}],"routes":[{"area":"admin","method":"GET","path":"/api/v1/admin/pkg_danger","handler":"reserved","auth":"admin","permission":"pkg_danger.item.audit"}]}`)
	if err := os.WriteFile(filepath.Join(pkgDir, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "hack.sh"), []byte("#!/bin/sh\necho hacked\n"), 0o644); err != nil {
		t.Fatalf("write sh: %v", err)
	}

	res, err := svc.DryRunPluginPackage(filepath.ToSlash(filepath.Join(baseRel, "pkg_danger")))
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

func TestDryRunPluginPackage_MigrationsDirectoryOnlyAndDeprecatedRootSchema(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	root := mustProjectRoot(t)
	baseRel := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-pkg-dryrun-migrations")
	pkgDir := filepath.Join(root, filepath.FromSlash(baseRel), "pkg_migrations")
	_ = os.RemoveAll(pkgDir)
	if err := os.MkdirAll(filepath.Join(pkgDir, "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(baseRel))) })

	manifest := minimalPackageManifest(t, "pkg_migrations")
	if err := os.WriteFile(filepath.Join(pkgDir, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "001_schema.sql"), []byte("create table deprecated_root(id bigint);\n"), 0o644); err != nil {
		t.Fatalf("write deprecated schema: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "migrations", "002_add_name.sql"), []byte("alter table demo add column name varchar(64);\n"), 0o644); err != nil {
		t.Fatalf("write migration 002: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "migrations", "001_init.sql"), []byte("create table demo(id bigint);\n"), 0o644); err != nil {
		t.Fatalf("write migration 001: %v", err)
	}

	res, err := svc.DryRunPluginPackage(filepath.ToSlash(filepath.Join(baseRel, "pkg_migrations")))
	if err != nil {
		t.Fatalf("DryRunPluginPackage failed: %v", err)
	}
	if res.Status == "blocked" {
		t.Fatalf("deprecated root schema should warn, not block: %#v", res)
	}
	if len(res.MigrationPlan) != 2 {
		t.Fatalf("expected only migrations/*.sql in migration plan, got %#v", res.MigrationPlan)
	}
	if res.MigrationPlan[0].Path != "migrations/001_init.sql" || res.MigrationPlan[1].Path != "migrations/002_add_name.sql" {
		t.Fatalf("expected sorted migrations plan, got %#v", res.MigrationPlan)
	}
	for _, item := range res.MigrationPlan {
		if item.WillExecute {
			t.Fatalf("dry-run migration plan must not execute SQL: %#v", item)
		}
		if item.Path == "001_schema.sql" {
			t.Fatalf("root 001_schema.sql must not be migration plan: %#v", res.MigrationPlan)
		}
	}
	if !containsString(res.Warnings, "001_schema.sql 已废弃") {
		t.Fatalf("expected deprecated warning, got %#v", res.Warnings)
	}
	if len(res.FileScan.DangerousFiles) != 0 {
		t.Fatalf("root 001_schema.sql should not be dangerous, got %#v", res.FileScan.DangerousFiles)
	}
}

func TestDryRunPluginPackage_RootOtherSQLStillBlocked(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	root := mustProjectRoot(t)
	baseRel := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-pkg-dryrun-root-sql")
	pkgDir := filepath.Join(root, filepath.FromSlash(baseRel), "pkg_root_sql")
	_ = os.RemoveAll(pkgDir)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(baseRel))) })

	if err := os.WriteFile(filepath.Join(pkgDir, "manifest.json"), minimalPackageManifest(t, "pkg_root_sql"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "002_schema.sql"), []byte("drop table demo;\n"), 0o644); err != nil {
		t.Fatalf("write root sql: %v", err)
	}

	res, err := svc.DryRunPluginPackage(filepath.ToSlash(filepath.Join(baseRel, "pkg_root_sql")))
	if err != nil {
		t.Fatalf("DryRunPluginPackage failed: %v", err)
	}
	if res.Status != "blocked" {
		t.Fatalf("expected root SQL blocked, got %q warnings=%v", res.Status, res.Warnings)
	}
	if res.BlockedCode != "plugin_package_dangerous_file" {
		t.Fatalf("expected dangerous file blocked code, got %q", res.BlockedCode)
	}
}

func minimalPackageManifest(t *testing.T, code string) []byte {
	t.Helper()
	manifest := map[string]any{
		"code":                    code,
		"name":                    code,
		"version":                 "1.0.0",
		"compatible_core_version": ">=1.4.0",
		"is_system":               false,
		"content_types":           []string{code + "_item"},
		"content_type_definitions": []map[string]any{{
			"type":              code + "_item",
			"name":              "Item",
			"plugin_code":       code,
			"create_permission": code + ".item.create",
			"edit_permission":   code + ".item.edit",
			"delete_permission": code + ".item.delete",
			"audit_permission":  code + ".item.audit",
			"default_status":    "draft",
			"allow_comment":     true,
			"allow_like":        true,
			"allow_favorite":    true,
			"seo_type":          "Article",
		}},
		"permissions": []map[string]any{{
			"code":  code + ".item.create",
			"name":  "create",
			"scope": "community",
		}, {
			"code":  code + ".item.edit",
			"name":  "edit",
			"scope": "own",
		}, {
			"code":  code + ".item.delete",
			"name":  "delete",
			"scope": "own",
		}, {
			"code":  code + ".item.audit",
			"name":  "audit",
			"scope": "community",
		}},
		"menus": []map[string]any{{
			"code":       code + ".admin",
			"title":      code,
			"path":       "/admin-next/" + code,
			"location":   "admin",
			"area":       "admin",
			"permission": code + ".item.audit",
		}},
		"routes": []map[string]any{{
			"area":       "admin",
			"method":     "GET",
			"path":       "/api/v1/admin/" + code,
			"handler":    "reserved",
			"auth":       "admin",
			"permission": code + ".item.audit",
		}},
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	return raw
}

func containsString(items []string, sub string) bool {
	for _, item := range items {
		if strings.Contains(item, sub) {
			return true
		}
	}
	return false
}
