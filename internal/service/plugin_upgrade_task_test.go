package service

import (
	"testing"

	"devhub-gin-backend/internal/domain"
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
