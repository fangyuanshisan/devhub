package service

import (
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestEnablePrecheckRejectsNotInstalled(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, err := svc.RunPluginEnablePrecheckAs(PluginEnablePrecheckOperator{ID: 1, Name: "tester"}, "not_exist")
	apiErr, _ := err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.HTTPStatus != 404 {
		t.Fatalf("expected 404 plugin not found, got %v", err)
	}
}

func TestEnablePrecheckRequiresPassedPrecheckAndCompatCheck(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	// seed an installed plugin record (minimal fields for enable-precheck).
	_, _ = repo.SavePlugin(domain.Plugin{
		PluginManifest: domain.PluginManifest{
			Code:                 "ep_demo",
			Name:                 "Enable Precheck Demo",
			Version:              "1.0.0",
			CompatibleCoreVersion: ">=1.0.0",
			SourceType:            "local_package",
		},
		Status:        "disabled",
		InstallStatus: "installed",
		ManifestChecksum: "sha256_dummy",
	})

	_, err := svc.RunPluginEnablePrecheckAs(PluginEnablePrecheckOperator{ID: 1, Name: "tester"}, "ep_demo")
	apiErr, _ := err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.Code != "plugin_enable_precheck_precheck_missing" {
		t.Fatalf("expected precheck missing error, got %v", err)
	}

	// add a passed precheck, but missing compat-check.
	_, _ = repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PluginCode:     "ep_demo",
		Version:        "1.0.0",
		Status:         domain.PluginPackagePrecheckStatusPassed,
		ManifestJSON:   compatManifest("ep_demo", ">=1.0.0", ""),
		ChecksumStatus: "ok",
		PackagePath:    "storage/plugins/staging/test/ep_demo",
		CreatedBy:      1,
	})

	_, err = svc.RunPluginEnablePrecheckAs(PluginEnablePrecheckOperator{ID: 1, Name: "tester"}, "ep_demo")
	apiErr, _ = err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.Code != "plugin_enable_precheck_compat_missing" {
		t.Fatalf("expected compat missing error, got %v", err)
	}
}

func TestEnablePrecheckDeleteAuditUsesOperator(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	rec, _ := repo.AppendPluginEnablePrecheck(domain.PluginEnablePrecheckRecord{PluginCode: "ep_demo", Version: "1.0.0"})
	_, err := svc.DeletePluginEnablePrecheckAs(PluginEnablePrecheckOperator{ID: 9, Name: "op"}, rec.ID)
	if err != nil {
		t.Fatalf("DeletePluginEnablePrecheckAs: %v", err)
	}
	logs, _ := repo.AdminLogsByFilter(domain.AdminLogFilter{Action: "plugin.enable_precheck.deleted", Page: 1, PageSize: 50})
	found := false
	for _, it := range logs {
		if it.Action == "plugin.enable_precheck.deleted" && it.ActorID == 9 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected audit log with actor_id=9")
	}
}
