package plugins

import (
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
)

func TestCheckVersionConstraint(t *testing.T) {
	cases := []struct {
		version    string
		constraint string
		want       bool
	}{
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "1.2.4", false},
		{"1.2.3", ">=1.2.0", true},
		{"1.2.3", "<2.0.0", true},
		{"1.2.3", ">=1.2.0 <2.0.0", true},
		{"2.0.0", ">=1.2.0 <2.0.0", false},
		{"1.8.0", "^1.7.0", true},
		{"2.0.0", "^1.7.0", false},
		{"1.7.5", "~1.7.0", true},
		{"1.8.0", "~1.7.0", false},
	}
	for _, tc := range cases {
		got, err := CheckVersionConstraint(tc.version, tc.constraint)
		if err != nil {
			t.Fatalf("CheckVersionConstraint(%q,%q) error: %v", tc.version, tc.constraint, err)
		}
		if got != tc.want {
			t.Fatalf("CheckVersionConstraint(%q,%q)=%v want %v", tc.version, tc.constraint, got, tc.want)
		}
	}
	if _, err := CheckVersionConstraint("1.2.3", "^bad"); err == nil {
		t.Fatal("unsupported constraint should error")
	}
}

func TestCheckPluginVersionCompatibility(t *testing.T) {
	ok := CheckPluginVersionCompatibility(domain.PluginManifest{MinCoreVersion: "1.4.0", CompatibleCoreVersion: ">=1.4.0 <2.0.0"}, "v1.4.0")
	if ok.Status != CompatibilityCompatible {
		t.Fatalf("expected compatible, got %#v", ok)
	}
	high := CheckPluginVersionCompatibility(domain.PluginManifest{MinCoreVersion: "9.0.0"}, "v1.4.0")
	if high.Status != CompatibilityIncompatible {
		t.Fatalf("expected incompatible min core, got %#v", high)
	}
	missing := CheckPluginVersionCompatibility(domain.PluginManifest{}, "v1.4.0")
	if missing.Status != CompatibilityWarning {
		t.Fatalf("expected warning for missing min_core_version, got %#v", missing)
	}
}

func TestValidatePluginDependencies(t *testing.T) {
	existing := []domain.Plugin{
		{PluginManifest: domain.PluginManifest{Code: "qa", Name: "QA", Version: "1.2.0"}, Status: StatusEnabled},
		{PluginManifest: domain.PluginManifest{Code: "docs", Name: "Docs", Version: "1.0.0"}, Status: StatusDisabled},
	}
	manifest := domain.PluginManifest{
		Code: "consumer",
		Dependencies: []domain.PluginDependency{
			{Code: "qa", Version: ">=1.0.0", Required: true},
			{Code: "docs", Required: true},
			{Code: "missing", Required: false},
		},
	}
	checks, summary, errors, warnings := ValidatePluginDependencies(manifest, existing)
	if len(checks) != 3 || summary.Satisfied != 1 || summary.Blocking != 1 || summary.Warnings != 1 {
		t.Fatalf("unexpected dependency summary checks=%#v summary=%#v", checks, summary)
	}
	if !strings.Contains(strings.Join(errors, "\n"), "docs") {
		t.Fatalf("expected disabled docs error, got %v", errors)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "optional") {
		t.Fatalf("expected optional warning, got %v", warnings)
	}
}

func TestValidatePluginDependenciesVersionMismatchAndCycles(t *testing.T) {
	existing := []domain.Plugin{
		{
			PluginManifest: domain.PluginManifest{
				Code:         "plugin_b",
				Name:         "B",
				Version:      "1.0.0",
				Dependencies: []domain.PluginDependency{{Code: "plugin_c", Required: true}},
			},
			Status: StatusEnabled,
		},
		{
			PluginManifest: domain.PluginManifest{
				Code:         "plugin_c",
				Name:         "C",
				Version:      "1.0.0",
				Dependencies: []domain.PluginDependency{{Code: "plugin_a", Required: true}},
			},
			Status: StatusEnabled,
		},
	}
	self := domain.PluginManifest{Code: "plugin_a", Dependencies: []domain.PluginDependency{{Code: "plugin_a", Required: true}}}
	_, _, selfErrors, _ := ValidatePluginDependencies(self, existing)
	if !strings.Contains(strings.Join(selfErrors, "\n"), "不能依赖自身") {
		t.Fatalf("expected self dependency error, got %v", selfErrors)
	}

	cycle := domain.PluginManifest{Code: "plugin_a", Dependencies: []domain.PluginDependency{{Code: "plugin_b", Required: true}}}
	checks, _, cycleErrors, _ := ValidatePluginDependencies(cycle, existing)
	if !strings.Contains(strings.Join(cycleErrors, "\n"), "plugin_a -> plugin_b -> plugin_c -> plugin_a") {
		t.Fatalf("expected cycle chain, got checks=%#v errors=%v", checks, cycleErrors)
	}

	versionExisting := []domain.Plugin{{PluginManifest: domain.PluginManifest{Code: "plugin_b", Name: "B", Version: "1.0.0"}, Status: StatusEnabled}}
	version := domain.PluginManifest{Code: "plugin_a", Dependencies: []domain.PluginDependency{{Code: "plugin_b", Version: ">=2.0.0", Required: true}}}
	_, _, versionErrors, _ := ValidatePluginDependencies(version, versionExisting)
	if !strings.Contains(strings.Join(versionErrors, "\n"), "不满足 >=2.0.0") {
		t.Fatalf("expected version mismatch error, got %v", versionErrors)
	}
}

func TestDependencyDiff(t *testing.T) {
	current := []domain.PluginDependency{{Code: "qa", Version: ">=1.0.0", Required: true}, {Code: "docs", Required: false}}
	next := []domain.PluginDependency{{Code: "qa", Version: ">=2.0.0", Required: true}, {Code: "wiki", Required: true}}
	diff := DependencyDiff(current, next)
	if len(diff.Added) != 1 || diff.Added[0].Code != "wiki" {
		t.Fatalf("unexpected added diff: %#v", diff)
	}
	if len(diff.Removed) != 1 || diff.Removed[0].Code != "docs" {
		t.Fatalf("unexpected removed diff: %#v", diff)
	}
	if len(diff.VersionChanged) != 1 || diff.VersionChanged[0].Code != "qa" {
		t.Fatalf("unexpected version diff: %#v", diff)
	}
}
