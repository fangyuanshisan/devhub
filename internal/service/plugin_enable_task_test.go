package service

import (
	"testing"
	"time"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestEnablePluginFromEnablePrecheck_Success(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	// installed plugin
	_, _ = repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                  "en_demo",
			Name:                  "Enable Demo",
			Version:               "1.0.0",
			CompatibleCoreVersion: ">=1.0.0",
			SourceType:            "local_package",
			ContentTypeDefs: []domain.ContentTypeDefinition{{
				Type:             "en_demo_item",
				Name:             "Item",
				PluginCode:       "en_demo",
				CreatePermission: "en_demo.item.create",
			}},
			Permissions: []domain.PermissionDefinition{{Code: "en_demo.item.create", Name: "create", Scope: "global"}},
		},
		Status:           "disabled",
		InstallStatus:    "installed",
		ManifestChecksum: "sha256_dummy",
		PackageChecksum:  "sha256_pkg_dummy",
	})

	// chain: package precheck + compat-check required by enable-precheck / enable.
	_, _ = repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PluginCode:     "en_demo",
		Version:        "1.0.0",
		Status:         domain.PluginPackagePrecheckStatusPassed,
		ManifestJSON:   compatManifest("en_demo", ">=1.0.0", ""),
		ChecksumStatus: "ok",
		PackagePath:    "plugins-local/repository-fixtures/demo_notice_install",
		CreatedBy:      1,
	})
	precheckItems, _, _ := repo.PluginPackagePrechecks(domain.PluginPackagePrecheckFilter{PluginCode: "en_demo", Page: 1, PageSize: 10})
	if len(precheckItems) == 0 {
		t.Fatalf("expected precheck record")
	}
	_, _ = repo.AppendPluginPackageCompatCheck(domain.PluginPackageCompatCheckRecord{
		PackagePrecheckID:     precheckItems[0].ID,
		PluginCode:            "en_demo",
		Version:               "1.0.0",
		Status:                domain.PluginPackageCompatCheckStatusPassed,
		CanInstall:            true,
		CoreVersion:           "v1.7.0",
		CreatedBy:             1,
		StartedAt:             time.Now().Format("2006-01-02 15:04:05"),
		FinishedAt:            time.Now().Format("2006-01-02 15:04:05"),
		CompatibleCoreVersion: ">=1.0.0",
	})

	// enable-precheck passed
	pre, _ := repo.AppendPluginEnablePrecheck(domain.PluginEnablePrecheckRecord{
		PluginCode: "en_demo",
		Version:    "1.0.0",
		Status:     domain.PluginEnablePrecheckStatusPassed,
		CanEnable:  true,
		FinishedAt: time.Now().Format("2006-01-02 15:04:05"),
	})

	resp, err := svc.EnablePluginFromEnablePrecheckAs(PluginEnableOperator{ID: 1, Name: "tester"}, pre.ID)
	if err != nil {
		t.Fatalf("EnablePluginFromEnablePrecheckAs: %v", err)
	}
	if resp.Status != domain.PluginEnableTaskStatusEnabled {
		t.Fatalf("expected enabled task status, got %q", resp.Status)
	}
	plugin, _ := repo.PluginByCode("en_demo")
	if plugin.Status != "enabled" {
		t.Fatalf("expected plugin enabled, got %q", plugin.Status)
	}
	if resp.Registered == nil || resp.Registered["permissions"] == 0 {
		t.Fatalf("expected registered summary")
	}
}

func TestEnablePluginFromEnablePrecheck_Expired(t *testing.T) {
	orig := getenv
	t.Cleanup(func() { getenv = orig })
	getenv = func(key string) string {
		if key == "DEVHUB_PLUGIN_ENABLE_PRECHECK_TTL_SECONDS" {
			return "1"
		}
		return ""
	}

	repo := store.NewMemoryStore()
	svc := New(repo)
	_, _ = repo.SavePlugin(domain.Plugin{PluginManifest: domain.PluginManifest{Code: "en_demo", Version: "1.0.0"}, Status: "disabled", InstallStatus: "installed"})
	pre, _ := repo.AppendPluginEnablePrecheck(domain.PluginEnablePrecheckRecord{
		PluginCode: "en_demo",
		Version:    "1.0.0",
		Status:     domain.PluginEnablePrecheckStatusPassed,
		CanEnable:  true,
		FinishedAt: time.Now().Add(-5 * time.Second).Format("2006-01-02 15:04:05"),
	})
	_, err := svc.EnablePluginFromEnablePrecheckAs(PluginEnableOperator{ID: 1, Name: "tester"}, pre.ID)
	api, _ := err.(*domain.APIError)
	if err == nil || api == nil || api.Code != "plugin_enable_precheck_expired" {
		t.Fatalf("expected expired error, got %v", err)
	}
}
