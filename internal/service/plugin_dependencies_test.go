package service

import (
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestInstallPluginManifestBlocksRequiredDependency(t *testing.T) {
	svc := New(store.NewMemoryStore())
	_, validation, err := svc.InstallPluginManifest([]byte(dependencyManifest("consumer", "1.0.0", "missing_plugin", ">=1.0.0", true)))
	if err == nil {
		t.Fatal("expected install to fail for missing required dependency")
	}
	if validation.Valid || validation.DependencySummary.Blocking == 0 {
		t.Fatalf("expected blocking dependency validation, got %#v", validation)
	}
}

func TestInstallPluginManifestAllowsOptionalDependency(t *testing.T) {
	svc := New(store.NewMemoryStore())
	plugin, validation, err := svc.InstallPluginManifest([]byte(dependencyManifest("consumer", "1.0.0", "missing_plugin", ">=1.0.0", false)))
	if err != nil {
		t.Fatalf("optional dependency should not block install: %v", err)
	}
	if !validation.Valid || validation.DependencySummary.Warnings == 0 {
		t.Fatalf("expected optional dependency warning, got %#v", validation)
	}
	if plugin.Status != pluginregistry.StatusDisabled {
		t.Fatalf("installed manifest plugin should remain disabled, got %s", plugin.Status)
	}
}

func TestEnablePluginBlocksDependencyStatuses(t *testing.T) {
	cases := []struct {
		name   string
		status string
	}{
		{"disabled", pluginregistry.StatusDisabled},
		{"archived", pluginregistry.StatusArchived},
		{"migration_failed", pluginregistry.StatusMigrationFailed},
		{"config_invalid", pluginregistry.StatusConfigInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := store.NewMemoryStore()
			svc := New(repo)
			qa, _ := repo.PluginByCode("qa")
			qa.Status = tc.status
			if _, err := repo.SavePlugin(qa); err != nil {
				t.Fatalf("save qa: %v", err)
			}
			consumer := domain.Plugin{
				PluginManifest: domain.PluginManifest{
					Code:         "consumer",
					Name:         "Consumer",
					Version:      "1.0.0",
					Dependencies: []domain.PluginDependency{{Code: "qa", Version: ">=1.0.0", Required: true}},
				},
				Status: pluginregistry.StatusDisabled,
			}
			if _, err := repo.SavePlugin(consumer); err != nil {
				t.Fatalf("save consumer: %v", err)
			}
			if _, err := svc.SetPluginStatus("consumer", pluginregistry.StatusEnabled); err == nil {
				t.Fatal("expected dependency readiness error, got nil")
			} else if apiErr, ok := err.(*domain.APIError); ok {
				if apiErr.Code != PluginErrDependencyMissing && apiErr.Code != PluginErrDependencyDisabled && apiErr.Code != PluginErrDependencyArchived && apiErr.Code != PluginErrDependencyVersion && apiErr.Code != PluginErrDependencyCycle {
					t.Fatalf("unexpected plugin error code: %s", apiErr.Code)
				}
			} else {
				t.Fatalf("expected structured API error, got %T: %v", err, err)
			}
		})
	}
}

func TestEnablePluginAllowsSatisfiedDependency(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	consumer := domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:         "consumer",
			Name:         "Consumer",
			Version:      "1.0.0",
			Dependencies: []domain.PluginDependency{{Code: "qa", Version: ">=1.0.0", Required: true}},
		},
		Status: pluginregistry.StatusDisabled,
	}
	if _, err := repo.SavePlugin(consumer); err != nil {
		t.Fatalf("save consumer: %v", err)
	}
	if _, err := svc.SetPluginStatus("consumer", pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("satisfied dependency should allow enable: %v", err)
	}
}

func TestPluginUpgradeDryRunDependencyDiffAndUpgradeBlocksNewRequiredDependency(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	base := []byte(`{
		"code":"consumer",
		"name":"Consumer",
		"version":"1.0.0",
		"min_core_version":"1.4.0",
		"content_types":[{"type":"consumer_item","name":"Consumer Item","plugin_code":"consumer","create_permission":"consumer.item.create"}],
		"permissions":[{"code":"consumer.item.create","name":"Create","scope":"community"}],
		"menus":[],
		"routes":[],
		"config_schema":{"type":"object","properties":{}},
		"hooks":[],
		"migrations":[]
	}`)
	if _, _, err := svc.InstallPluginManifest(base); err != nil {
		t.Fatalf("install base: %v", err)
	}
	upgrade := []byte(dependencyManifest("consumer", "2.0.0", "missing_plugin", ">=1.0.0", true))
	preview, err := svc.PluginUpgradeDryRun("consumer", upgrade)
	if err != nil {
		t.Fatalf("upgrade dry-run should return preview: %v", err)
	}
	if preview.Validation.Valid || len(preview.DependencyDiff.Added) != 1 {
		t.Fatalf("expected invalid preview with added dependency, got %#v", preview)
	}
	if _, err := svc.UpgradePluginManifest("consumer", upgrade); err == nil {
		t.Fatal("upgrade should fail for missing required dependency")
	}
}

func dependencyManifest(code, version, depCode, depVersion string, required bool) string {
	requiredText := "false"
	if required {
		requiredText = "true"
	}
	return `{
		"code":"` + code + `",
		"name":"Consumer",
		"version":"` + version + `",
		"min_core_version":"1.4.0",
		"compatible_core_version":">=1.4.0 <2.0.0",
		"content_types":[{"type":"` + code + `_item","name":"Consumer Item","plugin_code":"` + code + `","create_permission":"` + code + `.item.create"}],
		"permissions":[{"code":"` + code + `.item.create","name":"Create","scope":"community"}],
		"menus":[],
		"routes":[],
		"config_schema":{"type":"object","properties":{}},
		"dependencies":[{"code":"` + depCode + `","version":"` + depVersion + `","required":` + requiredText + `,"reason":"test dependency"}],
		"hooks":[],
		"migrations":[]
	}`
}
