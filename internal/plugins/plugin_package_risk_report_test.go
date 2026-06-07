package plugins_test

import (
	"testing"

	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestPluginPackageRiskReport_Levels(t *testing.T) {
	svc := service.New(store.NewMemoryStore())

	safe, err := svc.DryRunPluginPackage("examples/plugins/demo_notice")
	if err != nil {
		t.Fatalf("safe dry-run: %v", err)
	}
	if safe.RiskReport.Level != "low" && safe.RiskReport.Level != "medium" {
		t.Fatalf("expected safe low/medium, got %#v", safe.RiskReport)
	}

	noSum, err := svc.DryRunPluginPackage("examples/plugins/security-fixtures/no_checksums")
	if err != nil {
		t.Fatalf("no_checksums dry-run: %v", err)
	}
	if noSum.RiskReport.Level != "medium" && noSum.RiskReport.Level != "high" {
		t.Fatalf("expected no_checksums medium/high, got %#v", noSum.RiskReport)
	}

	mismatch, err := svc.DryRunPluginPackage("examples/plugins/security-fixtures/checksum_mismatch")
	if err != nil {
		t.Fatalf("checksum_mismatch dry-run: %v", err)
	}
	if mismatch.RiskReport.Level != "blocked" {
		t.Fatalf("expected mismatch blocked, got %#v", mismatch.RiskReport)
	}

	danger, err := svc.DryRunPluginPackage("examples/plugins/security-fixtures/dangerous_shell")
	if err != nil {
		t.Fatalf("dangerous_shell dry-run: %v", err)
	}
	if danger.RiskReport.Level != "blocked" {
		t.Fatalf("expected danger blocked, got %#v", danger.RiskReport)
	}
}
