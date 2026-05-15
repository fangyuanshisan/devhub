package service

import (
	"encoding/json"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestPluginVersionRepositoryAggregatesInstalledUploadedAndRemote(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	if _, err := repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{Code: "version_demo", Name: "Version Demo", Version: "1.0.0"},
		Status:         pluginregistry.StatusDisabled,
	}); err != nil {
		t.Fatalf("save plugin: %v", err)
	}
	if _, err := repo.AppendPluginPackageUpload(domain.PluginPackageUploadRecord{
		UploadID:       "upload-version-demo",
		Status:         domain.PluginPackageUploadStatusStaged,
		PackageCode:    "version_demo",
		PackageName:    "Version Demo",
		PackageVersion: "1.1.0",
		PackagePath:    "storage/plugins/staging/upload-version-demo/version_demo",
		RiskLevel:      "low",
	}); err != nil {
		t.Fatalf("save upload: %v", err)
	}
	metadata, _ := json.Marshal(map[string]any{"document": domain.PluginRemoteIndexDocument{
		SchemaVersion: "1",
		Source:        domain.PluginRemoteIndexSourceMeta{SourceID: "idx", Name: "Index"},
		Plugins: []domain.PluginRemoteIndexPluginDoc{{
			Code:          "version_demo",
			Name:          "Version Demo",
			LatestVersion: "1.2.0",
			Versions: []domain.PluginRemoteIndexVersionDoc{{
				Version:       "1.2.0",
				PackageURL:    "https://example.com/version_demo.zip",
				PackageSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				PublisherID:   "unknown",
				PublicKeyID:   "unknown-key",
			}},
		}},
	}})
	if _, err := repo.AppendPluginRemoteIndex(domain.PluginRemoteIndexSource{SourceID: "idx", Name: "Index", IndexURL: "https://example.com/index.json", Status: domain.PluginRemoteIndexStatusEnabled, MetadataJSON: string(metadata)}); err != nil {
		t.Fatalf("save remote index: %v", err)
	}

	got, err := svc.ListPluginVersionRepository(domain.PluginVersionFilter{PluginCode: "version_demo", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListPluginVersionRepository: %v", err)
	}
	if got.Pagination.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("unexpected repository result: %#v", got)
	}
	item := got.Items[0]
	if item.InstalledVersion != "1.0.0" || item.LatestRemoteVersion != "1.2.0" || !item.UpdateAvailable {
		t.Fatalf("unexpected version aggregation: %#v", item)
	}
	if len(item.Versions) < 3 {
		t.Fatalf("expected installed/uploaded/remote versions, got %#v", item.Versions)
	}
}

func TestPluginVersionUpgradeDiffBlocksSameDowngradeAndRemoteReadonly(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	if _, err := repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{Code: "version_guard", Name: "Version Guard", Version: "1.2.0"},
		Status:         pluginregistry.StatusDisabled,
	}); err != nil {
		t.Fatalf("save plugin: %v", err)
	}
	if _, err := svc.PluginVersionUpgradeDiff("version_guard", "1.2.0", domain.PluginUpgradeDiffRequest{Source: string(domain.PluginVersionSourceLocalPackage)}); err == nil {
		t.Fatal("expected same version to be rejected")
	} else if apiErr, ok := err.(*domain.APIError); !ok || apiErr.Code != "plugin_version_same_version" {
		t.Fatalf("unexpected same-version error: %T %v", err, err)
	}
	if _, err := svc.PluginVersionUpgradeDiff("version_guard", "1.1.0", domain.PluginUpgradeDiffRequest{Source: string(domain.PluginVersionSourceLocalPackage)}); err == nil {
		t.Fatal("expected downgrade to be rejected")
	} else if apiErr, ok := err.(*domain.APIError); !ok || apiErr.Code != "plugin_version_downgrade_forbidden" {
		t.Fatalf("unexpected downgrade error: %T %v", err, err)
	}
	if _, err := svc.PluginVersionUpgradeDiff("version_guard", "1.3.0", domain.PluginUpgradeDiffRequest{Source: string(domain.PluginVersionSourceRemoteIndex)}); err == nil {
		t.Fatal("expected remote index upgrade diff to be readonly")
	} else if apiErr, ok := err.(*domain.APIError); !ok || apiErr.Code != "plugin_version_remote_readonly" {
		t.Fatalf("unexpected remote readonly error: %T %v", err, err)
	}
}

func TestBuildPluginManifestDiffHighRiskAndSensitiveRedaction(t *testing.T) {
	current := domain.PluginManifest{
		Code:         "diff_demo",
		Name:         "Diff Demo",
		Version:      "1.0.0",
		ContentTypes: []string{"diff_item"},
		Permissions:  []domain.PermissionDefinition{{Code: "diff.item.create", Name: "Create"}},
		ConfigSchema: map[string]any{"type": "object", "properties": map[string]any{"oauth": map[string]any{"type": "object"}, "app_secret": map[string]any{"type": "string", "format": "password"}}},
		Hooks:        []domain.HookDefinition{{Name: "AfterCreateContent", Blocking: false}},
	}
	target := domain.PluginManifest{
		Code:         "diff_demo",
		Name:         "Diff Demo",
		Version:      "1.1.0",
		ContentTypes: []string{},
		Permissions:  []domain.PermissionDefinition{{Code: "diff.manage", Name: "Manage", Scope: "admin"}},
		ConfigSchema: map[string]any{"type": "object", "required": []any{"app_secret"}, "properties": map[string]any{"app_secret": map[string]any{"type": "integer", "format": "password"}}},
		Dependencies: []domain.PluginDependency{{Code: "search", Required: true, Version: ">=1.0.0"}},
		Hooks:        []domain.HookDefinition{{Name: "AfterCreateContent", Blocking: true}},
		Migrations:   []domain.PluginMigrationDefinition{{MigrationName: "diff_110", MigrationVersion: "1.1.0", Direction: "up"}},
	}
	sections, summary := buildPluginManifestDiff(current, target)
	if summary.HighRisk < 5 {
		t.Fatalf("expected multiple high risk changes, got summary=%#v sections=%#v", summary, sections)
	}
	foundRedaction := false
	for _, section := range sections {
		for _, item := range section.Items {
			if item.Path == "config_schema.app_secret" && (item.Before == "[REDACTED]" || item.After == "[REDACTED]") {
				foundRedaction = true
			}
		}
	}
	if !foundRedaction {
		t.Fatalf("expected sensitive config diff redaction, got %#v", sections)
	}
}
