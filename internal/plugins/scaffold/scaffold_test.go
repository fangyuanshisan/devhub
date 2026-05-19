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
		Code:                   "demo_links",
		Name:                   "Demo Links",
		ContentType:            "demo_link",
		ContentName:            "演示链接",
		Description:            "Demo links plugin",
		Author:                 "DevHub",
		Output:                 dir,
		WithConfig:             true,
		WithHooks:              true,
		WithMigration:          true,
		IncludeRegistryExample: true,
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
	for _, name := range []string{"README.md", "config.example.json", "content-type.md", "permissions.md", "hooks.md", "migrations.md", "registry.example.go"} {
		if _, err := os.Stat(filepath.Join(result.Dir, name)); err != nil {
			t.Fatalf("expected generated file %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(result.Dir, "001_schema.sql")); !os.IsNotExist(err) {
		t.Fatalf("scaffold must not generate deprecated root 001_schema.sql, stat err=%v", err)
	}
}

func TestGenerateRejectsInvalidCode(t *testing.T) {
	_, err := Generate(Options{Code: "Demo-Links", Name: "Demo Links", ContentType: "demo_link", Output: t.TempDir()})
	if err == nil {
		t.Fatal("expected invalid code to fail")
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
