package service

import (
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestPluginUpgradeImpactRejectsSystemPlugin(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, _ = repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                  "qa",
			Name:                  "QA",
			Version:               "1.0.0",
			CompatibleCoreVersion: ">=1.0.0",
			SourceType:            "builtin",
			IsSystem:              true,
		},
		Status:        "enabled",
		InstallStatus: "installed",
	})
	_, err := svc.PluginUpgradeImpact(PluginUpgradeOperator{ID: 1, Name: "tester"}, "qa", 1)
	apiErr, _ := err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.Code != "plugin_upgrade_system_forbidden" {
		t.Fatalf("expected system forbidden, got %v", err)
	}
}

func TestPluginUpgradeImpactRejectsCodeMismatch(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	// installed external plugin
	_, _ = repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                  "up_demo",
			Name:                  "Upgrade Demo",
			Version:               "1.0.0",
			CompatibleCoreVersion: ">=1.0.0",
			SourceType:            "local_package",
		},
		Status:        "disabled",
		InstallStatus: "installed",
	})

	pre, _ := repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PluginCode:     "other",
		Version:        "2.0.0",
		Status:         domain.PluginPackagePrecheckStatusPassed,
		ManifestJSON:   compatManifest("other", ">=1.0.0", ""),
		ChecksumStatus: "ok",
		PackagePath:    "storage/plugins/staging/test/other",
	})
	compat, _ := repo.AppendPluginPackageCompatCheck(domain.PluginPackageCompatCheckRecord{
		PackagePrecheckID: pre.ID,
		PluginCode:        "other",
		Version:           "2.0.0",
		Status:            domain.PluginPackageCompatCheckStatusPassed,
		CanInstall:        true,
	})

	_, err := svc.PluginUpgradeImpact(PluginUpgradeOperator{ID: 1, Name: "tester"}, "up_demo", compat.ID)
	apiErr, _ := err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.Code != "plugin_upgrade_target_code_mismatch" {
		t.Fatalf("expected code mismatch, got %v", err)
	}
}

func TestPluginUpgradeImpactRejectsLowerOrSameVersion(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	// installed external plugin
	_, _ = repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                  "up_demo",
			Name:                  "Upgrade Demo",
			Version:               "2.0.0",
			CompatibleCoreVersion: ">=1.0.0",
			SourceType:            "local_package",
		},
		Status:        "disabled",
		InstallStatus: "installed",
	})

	pre, _ := repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PluginCode:     "up_demo",
		Version:        "1.0.0",
		Status:         domain.PluginPackagePrecheckStatusPassed,
		ManifestJSON:   compatManifest("up_demo", ">=1.0.0", ""),
		ChecksumStatus: "ok",
		PackagePath:    "storage/plugins/staging/test/up_demo",
	})
	compat, _ := repo.AppendPluginPackageCompatCheck(domain.PluginPackageCompatCheckRecord{
		PackagePrecheckID: pre.ID,
		PluginCode:        "up_demo",
		Version:           "1.0.0",
		Status:            domain.PluginPackageCompatCheckStatusPassed,
		CanInstall:        true,
	})

	_, err := svc.PluginUpgradeImpact(PluginUpgradeOperator{ID: 1, Name: "tester"}, "up_demo", compat.ID)
	apiErr, _ := err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.Code != "plugin_version_same_version" {
		t.Fatalf("expected same/lower version rejected, got %v", err)
	}
}

func TestPluginUpgradeImpactRejectsPrecheckNotPassed(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, _ = repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                  "up_demo",
			Name:                  "Upgrade Demo",
			Version:               "1.0.0",
			CompatibleCoreVersion: ">=1.0.0",
			SourceType:            "local_package",
		},
		Status:        "disabled",
		InstallStatus: "installed",
	})

	pre, _ := repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PluginCode:     "up_demo",
		Version:        "2.0.0",
		Status:         "failed",
		ManifestJSON:   compatManifest("up_demo", ">=1.0.0", ""),
		ChecksumStatus: "ok",
		PackagePath:    "storage/plugins/staging/test/up_demo",
	})
	compat, _ := repo.AppendPluginPackageCompatCheck(domain.PluginPackageCompatCheckRecord{
		PackagePrecheckID: pre.ID,
		PluginCode:        "up_demo",
		Version:           "2.0.0",
		Status:            domain.PluginPackageCompatCheckStatusPassed,
		CanInstall:        true,
	})

	_, err := svc.PluginUpgradeImpact(PluginUpgradeOperator{ID: 1, Name: "tester"}, "up_demo", compat.ID)
	apiErr, _ := err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.Code != "plugin_upgrade_target_precheck_not_passed" {
		t.Fatalf("expected precheck not passed, got %v", err)
	}
}

func TestPluginUpgradeImpactRejectsCompatNotInstallable(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, _ = repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                  "up_demo",
			Name:                  "Upgrade Demo",
			Version:               "1.0.0",
			CompatibleCoreVersion: ">=1.0.0",
			SourceType:            "local_package",
		},
		Status:        "disabled",
		InstallStatus: "installed",
	})

	pre, _ := repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PluginCode:     "up_demo",
		Version:        "2.0.0",
		Status:         domain.PluginPackagePrecheckStatusPassed,
		ManifestJSON:   compatManifest("up_demo", ">=1.0.0", ""),
		ChecksumStatus: "ok",
		PackagePath:    "storage/plugins/staging/test/up_demo",
	})
	compat, _ := repo.AppendPluginPackageCompatCheck(domain.PluginPackageCompatCheckRecord{
		PackagePrecheckID: pre.ID,
		PluginCode:        "up_demo",
		Version:           "2.0.0",
		Status:            domain.PluginPackageCompatCheckStatusFailed,
		CanInstall:        false,
	})

	_, err := svc.PluginUpgradeImpact(PluginUpgradeOperator{ID: 1, Name: "tester"}, "up_demo", compat.ID)
	apiErr, _ := err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.Code != "plugin_upgrade_target_not_installable" {
		t.Fatalf("expected compat not installable, got %v", err)
	}
}

func TestPluginUpgradeFromCompatRequiresConfirmForWarning(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	root, err := serviceProjectRoot()
	if err != nil {
		t.Fatalf("project root: %v", err)
	}
	packagePath := filepath.Join("storage", "plugins", "staging", "test", "up_demo_confirm")
	absPackagePath := filepath.Join(root, packagePath)
	_ = os.RemoveAll(absPackagePath)
	if err := os.MkdirAll(filepath.Join(absPackagePath, "migrations"), 0o755); err != nil {
		t.Fatalf("mkdir package path: %v", err)
	}
	defer os.RemoveAll(absPackagePath)
	if err := os.WriteFile(filepath.Join(absPackagePath, "manifest.json"), []byte(`{"code":"up_demo","name":"Upgrade Demo v2","version":"2.0.0","compatible_core_version":">=1.0.0","permissions":[{"code":"up_demo.read","name":"Read","scope":"community"}],"menus":[{"code":"up_demo.frontend","title":"Demo","path":"/demo","location":"frontend"}],"routes":[],"content_types":[],"hooks":[{"name":"AfterCreateContent","mode":"non_blocking","service_type":"external_service","path":"/hooks/up_demo","method":"POST","timeout_ms":3000,"retry_enabled":true,"max_attempts":3,"failure_policy":"warn","enabled":true}],"config_schema":{"type":"object","properties":{}},"migrations":[{"plugin_code":"up_demo","migration_version":"2.0.0","migration_name":"up_demo_2","direction":"up","checksum":"sha256:test","rollback_supported":false}],"assets":["assets/up_demo.css"],"external_service":{"endpoint":"https://example.com/hooks/up_demo","timeout_ms":3000,"failure_policy":"warn","retry_policy":"linear"}}`), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(absPackagePath, "README.md"), []byte("# demo"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	if err := os.WriteFile(filepath.Join(absPackagePath, "config.example.json"), []byte(`{"endpoint_url":"https://example.com/hooks/up_demo"}`), 0o644); err != nil {
		t.Fatalf("write config example: %v", err)
	}

	_, _ = repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                  "up_demo",
			Name:                  "Upgrade Demo",
			Version:               "1.0.0",
			CompatibleCoreVersion: ">=1.0.0",
			Permissions: []domain.PermissionDefinition{
				{Code: "up_demo.read", Name: "Read", Scope: "community"},
				{Code: "up_demo.write", Name: "Write", Scope: "community"},
			},
			Menus:      []domain.MenuDefinition{{Code: "up_demo.frontend", Title: "Demo", Path: "/demo", Location: "frontend"}},
			SourceType: "local_package",
		},
		Status:        pluginregistry.StatusDisabled,
		InstallStatus: "installed",
	})

	pre, _ := repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PluginCode:     "up_demo",
		Version:        "2.0.0",
		Status:         domain.PluginPackagePrecheckStatusPassed,
		ManifestJSON:   `{"code":"up_demo","name":"Upgrade Demo v2","version":"2.0.0","compatible_core_version":">=1.0.0","permissions":[{"code":"up_demo.read","name":"Read","scope":"community"}],"menus":[{"code":"up_demo.frontend","title":"Demo","path":"/demo","location":"frontend"}],"routes":[],"content_types":[],"hooks":[{"name":"AfterCreateContent","mode":"non_blocking","service_type":"external_service","path":"/hooks/up_demo","method":"POST","timeout_ms":3000,"retry_enabled":true,"max_attempts":3,"failure_policy":"warn","enabled":true}],"config_schema":{"type":"object","properties":{}},"migrations":[{"plugin_code":"up_demo","migration_version":"2.0.0","migration_name":"up_demo_2","direction":"up","checksum":"sha256:test","rollback_supported":false}],"assets":["assets/up_demo.css"],"external_service":{"endpoint":"https://example.com/hooks/up_demo","timeout_ms":3000,"failure_policy":"warn","retry_policy":"linear"}}`,
		ChecksumStatus: "ok",
		PackagePath:    packagePath,
	})
	compat, _ := repo.AppendPluginPackageCompatCheck(domain.PluginPackageCompatCheckRecord{
		PackagePrecheckID: pre.ID,
		PluginCode:        "up_demo",
		Version:           "2.0.0",
		Status:            domain.PluginPackageCompatCheckStatusPassed,
		CanInstall:        true,
	})

	_, err = svc.UpgradePluginFromCompatCheckAs(PluginUpgradeOperator{ID: 1, Name: "tester"}, "up_demo", PluginUpgradeRequest{TargetCompatCheckID: compat.ID})
	if err == nil {
		t.Fatal("expected warning upgrade to require confirm")
	}
	if apiErr, ok := err.(*domain.APIError); !ok || apiErr.Code != "plugin_upgrade_confirm_required" {
		t.Fatalf("expected confirm-required error, got %v", err)
	}
	res, err := svc.UpgradePluginFromCompatCheckAs(PluginUpgradeOperator{ID: 1, Name: "tester"}, "up_demo", PluginUpgradeRequest{TargetCompatCheckID: compat.ID, Confirm: true})
	if err != nil {
		t.Fatalf("confirmed package upgrade failed: %v", err)
	}
	if res.Status != domain.PluginUpgradeTaskStatusUpgraded {
		t.Fatalf("expected upgraded task, got %#v", res)
	}
}
