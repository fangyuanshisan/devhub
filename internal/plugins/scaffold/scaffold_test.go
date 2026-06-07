package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

func TestGenerateDemoPlugin(t *testing.T) {
	dir := t.TempDir()
	result, err := Generate(Options{
		Code:               "demo_links",
		Name:               "Demo Links",
		ContentType:        "demo_link",
		ContentName:        "演示链接",
		Description:        "Demo links plugin",
		Author:             "DevHub",
		Output:             dir,
		WithConfig:         true,
		WithHooks:          true,
		WithMigration:      true,
		IncludeRegistryDoc: true,
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if result.Dir != filepath.Join(dir, "demo_links") {
		t.Fatalf("unexpected dir: %s", result.Dir)
	}
	manifestPath := filepath.Join(result.Dir, "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !json.Valid(raw) {
		t.Fatalf("manifest is not valid JSON: %s", string(raw))
	}
	validation := pluginregistry.ValidatePluginManifestJSON(raw, pluginregistry.Definitions(), "v1.4.0")
	if !validation.Valid {
		t.Fatalf("manifest should validate, errors=%v warnings=%v", validation.Errors, validation.Warnings)
	}
	var manifest domain.PluginManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if err := pluginregistry.ValidateConfigJSON(domain.Plugin{PluginManifest: manifest}, "{\n  \"display_mode\": \"standard\",\n  \"enabled\": true,\n  \"items_per_page\": 20\n}"); err != nil {
		t.Fatalf("config example should validate: %v", err)
	}
	for _, name := range []string{"README.md", "config.example.json", "docs/usage.md", "docs/content-types.md", "docs/permissions.md", "docs/hooks.md", "docs/migrations.md", "docs/registry-example.md", "migrations/001_init.sql"} {
		if _, err := os.Stat(filepath.Join(result.Dir, name)); err != nil {
			t.Fatalf("expected generated file %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(result.Dir, "registry.example.go")); !os.IsNotExist(err) {
		t.Fatalf("scaffold must not generate registry.example.go, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(result.Dir, "001_schema.sql")); !os.IsNotExist(err) {
		t.Fatalf("scaffold must not generate deprecated root 001_schema.sql, stat err=%v", err)
	}
	for _, hook := range result.Manifest.Hooks {
		if hook.Blocking || hook.Mode == string(pluginregistry.HookBlocking) {
			t.Fatalf("template must not generate blocking hook: %#v", hook)
		}
	}
}

func TestGenerateRejectsInvalidCode(t *testing.T) {
	_, err := Generate(Options{Code: "Demo-Links", Name: "Demo Links", ContentType: "demo_link", Output: t.TempDir()})
	if err == nil {
		t.Fatal("expected invalid code to fail")
	}
}

func TestGenerateAllowsHyphenatedCode(t *testing.T) {
	dir := t.TempDir()
	result, err := Generate(Options{
		Code:          "demo-links",
		Name:          "Demo Links",
		ContentType:   "demo_links_item",
		ContentName:   "演示链接",
		Output:        dir,
		WithConfig:    true,
		WithMigration: true,
	})
	if err != nil {
		t.Fatalf("Generate with hyphenated code failed: %v", err)
	}
	if result.Manifest.Code != "demo-links" {
		t.Fatalf("unexpected code: %s", result.Manifest.Code)
	}
	if _, err := os.Stat(filepath.Join(result.Dir, "migrations", "001_init.sql")); err != nil {
		t.Fatalf("expected migrations/001_init.sql: %v", err)
	}
	if _, err := os.Stat(filepath.Join(result.Dir, "001_schema.sql")); !os.IsNotExist(err) {
		t.Fatalf("root 001_schema.sql must not be generated, stat err=%v", err)
	}
}

func TestGeneratePluginTypesUseDeclarativeDocsOnly(t *testing.T) {
	cases := []Options{
		{
			Code:          "tpl_external_service",
			Name:          "Template External Service",
			PluginType:    PluginTypeExternalService,
			Output:        t.TempDir(),
			WithConfig:    true,
			WithHooks:     true,
			WithMigration: false,
		},
		{
			Code:          "tpl_admin_tool",
			Name:          "Template Admin Tool",
			PluginType:    PluginTypeAdminTool,
			Output:        t.TempDir(),
			WithConfig:    true,
			WithMigration: false,
		},
		{
			Code:         "tpl_frontend_mount",
			Name:         "Template Frontend Mount",
			PluginType:   PluginTypeFrontendMount,
			Output:       t.TempDir(),
			WithConfig:   true,
			MountPoint:   "admin.plugin.detail.preview",
			ComponentKey: "official.announcement.card",
		},
	}
	for _, tc := range cases {
		result, err := Generate(tc)
		if err != nil {
			t.Fatalf("Generate %s failed: %v", tc.PluginType, err)
		}
		for _, forbidden := range []string{"registry.example.go", "001_schema.sql", "package.json", "index.js", "main.go"} {
			if _, err := os.Stat(filepath.Join(result.Dir, forbidden)); !os.IsNotExist(err) {
				t.Fatalf("%s template must not generate %s, stat err=%v", tc.PluginType, forbidden, err)
			}
		}
		if _, err := os.Stat(filepath.Join(result.Dir, "docs", "usage.md")); err != nil {
			t.Fatalf("%s template should generate docs/usage.md: %v", tc.PluginType, err)
		}
		for _, hook := range result.Manifest.Hooks {
			if hook.Blocking || hook.Mode == string(pluginregistry.HookBlocking) {
				t.Fatalf("%s template must not generate blocking hook: %#v", tc.PluginType, hook)
			}
		}
	}
}

func TestGenerateRefusesExistingDirectoryWithoutForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "demo_links"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_, err := Generate(Options{Code: "demo_links", Name: "Demo Links", ContentType: "demo_link", Output: dir})
	if err == nil {
		t.Fatal("expected existing directory to fail without force")
	}
}

func TestGenerateForceOverwritesDirectory(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "demo_links")
	if err := os.Mkdir(pluginDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write old file: %v", err)
	}
	if _, err := Generate(Options{Code: "demo_links", Name: "Demo Links", ContentType: "demo_link", Output: dir, Force: true}); err != nil {
		t.Fatalf("Generate with force failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pluginDir, "old.txt")); !os.IsNotExist(err) {
		t.Fatalf("old file should be removed by force, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(pluginDir, "manifest.json")); err != nil {
		t.Fatalf("manifest should exist after force: %v", err)
	}
}
