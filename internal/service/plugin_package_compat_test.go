package service

import (
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestPluginPackageCompatCheckPassed(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	precheck := appendCompatPrecheck(t, repo, "compat_demo", compatManifest("compat_demo", ">=1.0.0", ""))

	res, err := svc.RunPluginPackageCompatCheckAs(PluginPackageCompatOperator{ID: 7}, precheck.ID)
	if err != nil {
		t.Fatalf("RunPluginPackageCompatCheckAs: %v", err)
	}
	if !res.CanInstall || res.Status != domain.PluginPackageCompatCheckStatusPassed {
		t.Fatalf("expected passed can_install, got status=%s can_install=%v errors=%v warnings=%v", res.Status, res.CanInstall, res.Errors, res.Warnings)
	}
	if _, ok := repo.PluginByCode("compat_demo"); ok {
		t.Fatalf("compat-check must not install plugin")
	}
}

func TestPluginPackageCompatCheckRejectsNonPassedPrecheck(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	precheck := appendCompatPrecheck(t, repo, "compat_failed", compatManifest("compat_failed", ">=1.7.0", ""))
	precheck.Status = domain.PluginPackagePrecheckStatusManifestInvalid
	precheck, _ = repo.SavePluginPackagePrecheck(precheck)

	_, err := svc.RunPluginPackageCompatCheckAs(PluginPackageCompatOperator{}, precheck.ID)
	apiErr, _ := err.(*domain.APIError)
	if err == nil || apiErr == nil || apiErr.Code != "plugin_package_compat_precheck_not_passed" {
		t.Fatalf("expected precheck status error, got %v", err)
	}
}

func TestPluginPackageCompatCheckCoreIncompatible(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	precheck := appendCompatPrecheck(t, repo, "compat_core", compatManifest("compat_core", ">=9.0.0", ""))

	res, err := svc.RunPluginPackageCompatCheckAs(PluginPackageCompatOperator{}, precheck.ID)
	if err != nil {
		t.Fatalf("RunPluginPackageCompatCheckAs: %v", err)
	}
	if res.Status != domain.PluginPackageCompatCheckStatusIncompatible || res.CanInstall {
		t.Fatalf("expected incompatible blocked, got %#v", res)
	}
}

func TestPluginPackageCompatCheckRequiredAndOptionalDependencies(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	required := appendCompatPrecheck(t, repo, "compat_dep_required", compatManifest("compat_dep_required", ">=1.0.0", `"dependencies":[{"code":"missing_required","required":true}],`))

	requiredRes, err := svc.RunPluginPackageCompatCheckAs(PluginPackageCompatOperator{}, required.ID)
	if err != nil {
		t.Fatalf("required dependency check: %v", err)
	}
	if requiredRes.Status != domain.PluginPackageCompatCheckStatusDependencyMissing || requiredRes.CanInstall {
		t.Fatalf("expected required dependency missing, got %#v", requiredRes)
	}

	optional := appendCompatPrecheck(t, repo, "compat_dep_optional", compatManifest("compat_dep_optional", ">=1.0.0", `"dependencies":[{"code":"missing_optional","required":false}],`))
	optionalRes, err := svc.RunPluginPackageCompatCheckAs(PluginPackageCompatOperator{}, optional.ID)
	if err != nil {
		t.Fatalf("optional dependency check: %v", err)
	}
	if optionalRes.Status != domain.PluginPackageCompatCheckStatusWarning || !optionalRes.CanInstall {
		t.Fatalf("expected optional dependency warning, got %#v", optionalRes)
	}
}

func TestPluginPackageCompatCheckPermissionRouteHookConfigMigration(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	raw := `{
		"code":"compat_bad",
		"name":"Compat Bad",
		"version":"1.0.0",
		"min_core_version":"1.0.0",
		"compatible_core_version":">=1.7.0",
		"content_types":[{"type":"compat_bad_item","name":"Item","create_permission":"compat_bad.item.create"}],
		"permissions":[{"code":"core.admin","name":"bad","scope":"global"},{"code":"compat_bad.item.create","name":"create","scope":"global"}],
		"menus":[{"code":"compat_bad.menu","title":"Bad","path":"https://example.com","area":"frontend","permission":"compat_bad.item.create"}],
		"routes":[{"area":"admin","method":"GET","path":"/api/v1/admin/plugins/evil","handler":"noop","permission":"compat_bad.item.create"}],
		"hooks":[{"name":"UnknownHook","mode":"blocking","failure_policy":"ignore"}],
		"config_schema":{"type":"object","properties":{"enabled":{"type":"boolean"}},"additionalProperties":false},
		"default_config":{"enabled":"yes","extra":true},
		"migrations":[{"migration_version":"1.0.0","migration_name":"../bad","direction":"down"}]
	}`
	precheck := appendCompatPrecheck(t, repo, "compat_bad", raw)

	res, err := svc.RunPluginPackageCompatCheckAs(PluginPackageCompatOperator{}, precheck.ID)
	if err != nil {
		t.Fatalf("RunPluginPackageCompatCheckAs: %v", err)
	}
	joined := strings.Join(res.Errors, "\n")
	for _, want := range []string{
		"permission_forbidden",
		"menu_path_invalid",
		"route_sensitive_path",
		"hook_unknown",
		"config_default_invalid",
		"migration_direction_unsupported",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected %s in errors:\n%s", want, joined)
		}
	}
	if res.CanInstall {
		t.Fatalf("expected can_install=false")
	}
}

func appendCompatPrecheck(t *testing.T, repo serviceTestCompatRepo, code string, manifest string) domain.PluginPackagePrecheckRecord {
	t.Helper()
	rec, err := repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PluginCode:        code,
		Version:           "1.0.0",
		Status:            domain.PluginPackagePrecheckStatusPassed,
		ManifestJSON:      manifest,
		ChecksumStatus:    "ok",
		PackagePath:       "storage/plugins/staging/test/" + code,
		StagingPath:       "storage/plugins/staging/test",
		PackageDownloadID: 0,
		CreatedBy:         1,
	})
	if err != nil {
		t.Fatalf("AppendPluginPackagePrecheck: %v", err)
	}
	return rec
}

type serviceTestCompatRepo interface {
	AppendPluginPackagePrecheck(record domain.PluginPackagePrecheckRecord) (domain.PluginPackagePrecheckRecord, error)
	SavePluginPackagePrecheck(record domain.PluginPackagePrecheckRecord) (domain.PluginPackagePrecheckRecord, error)
	PluginByCode(code string) (domain.Plugin, bool)
}

func compatManifest(code, coreConstraint, extra string) string {
	raw := strings.ReplaceAll(`{
		"code":"__CODE__",
		"name":"Compat Demo",
		"version":"1.0.0",
		"min_core_version":"1.0.0",
		"compatible_core_version":"__CORE__",
		__EXTRA__
		"content_types":[{"type":"__CODE___item","name":"Item","create_permission":"__CODE__.item.create"}],
		"permissions":[{"code":"__CODE__.item.create","name":"create","scope":"global"}],
		"menus":[{"code":"__CODE__.menu","title":"Demo","path":"/__CODE__","area":"frontend","permission":"__CODE__.item.create"}],
		"routes":[{"area":"admin","method":"GET","path":"/api/v1/admin/__CODE__/health","handler":"noop","permission":"__CODE__.item.create"}],
		"hooks":[{"name":"BeforeCreateContent","mode":"blocking","failure_policy":"block","timeout_ms":1000}],
		"config_schema":{"type":"object","properties":{"enabled":{"type":"boolean","default":true}}},
		"default_config":{"enabled":true},
		"migrations":[{"migration_version":"1.0.0","migration_name":"__CODE___init","direction":"up","checksum":"sha256:test"}]
	}`, "__CODE__", code)
	raw = strings.ReplaceAll(raw, "__CORE__", coreConstraint)
	raw = strings.ReplaceAll(raw, "__EXTRA__", extra)
	return raw
}
