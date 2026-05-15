package service

import (
	"encoding/json"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestPluginOperations_CleanupFailedInstall_RemovesResidues(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	// Simulate residue: a plugin record exists but install operation failed.
	_, _ = repo.SavePlugin(domain.Plugin{PluginManifest: domain.PluginManifest{Code: "demo_notice_install", Name: "Demo", Version: "1.0.0"}, Status: "disabled"})
	op, err := repo.AppendPluginOperationSnapshot(domain.PluginOperationSnapshot{
		OperationID:   "op_test_install_failed",
		OperationType: domain.PluginOperationTypeInstall,
		PluginCode:    "demo_notice_install",
		Status:        domain.PluginOperationStatusFailed,
	})
	if err != nil {
		t.Fatalf("AppendPluginOperationSnapshot: %v", err)
	}
	if op.OperationID == "" {
		t.Fatalf("expected operation id")
	}

	dry, err := svc.RecoverPluginOperationDryRun(op.OperationID)
	if err != nil {
		t.Fatalf("RecoverPluginOperationDryRun: %v", err)
	}
	if dry.Status == "blocked" {
		t.Fatalf("expected non-blocked dry-run")
	}

	res, err := svc.CleanupPluginOperation(PluginOperationOperator{ID: 1, Name: "tester"}, op.OperationID)
	if err != nil {
		t.Fatalf("CleanupPluginOperation: %v", err)
	}
	if res.Status != "ok" {
		t.Fatalf("unexpected status: %q", res.Status)
	}
	if _, ok := repo.PluginByCode("demo_notice_install"); ok {
		t.Fatalf("expected plugin removed after cleanup")
	}
}

func TestPluginOperations_RollbackDryRun_RequiresBeforeManifest(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	_, _ = repo.SavePlugin(domain.Plugin{PluginManifest: domain.PluginManifest{Code: "demo_notice_install", Name: "Demo", Version: "1.0.0"}, Status: "disabled"})
	_, _ = repo.AppendPluginOperationSnapshot(domain.PluginOperationSnapshot{
		OperationID:   "op_test_upgrade_missing_before",
		OperationType: domain.PluginOperationTypeUpgrade,
		PluginCode:    "demo_notice_install",
		FromVersion:   "1.0.0",
		ToVersion:     "1.1.0",
		Status:        domain.PluginOperationStatusApplied,
	})

	resp, err := svc.PluginUpgradeRollbackDryRun("demo_notice_install", domain.PluginUpgradeRollbackDryRunRequest{OperationID: "op_test_upgrade_missing_before"})
	if err != nil {
		t.Fatalf("PluginUpgradeRollbackDryRun: %v", err)
	}
	if resp.Status != "blocked" {
		t.Fatalf("expected blocked when before_manifest_json missing, got %q", resp.Status)
	}
}

func TestPluginOperations_RollbackDryRun_DiffWorks(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	// Current plugin.
	p, _ := repo.SavePlugin(domain.Plugin{PluginManifest: domain.PluginManifest{Code: "demo_notice_install", Name: "Demo", Version: "1.1.0", Description: "new"}, Status: "disabled"})

	before := domain.PluginManifest{Code: "demo_notice_install", Name: "Demo", Version: "1.0.0", Description: "old"}
	beforeRaw, _ := json.Marshal(before)
	_, _ = repo.AppendPluginOperationSnapshot(domain.PluginOperationSnapshot{
		OperationID:        "op_test_upgrade_before",
		OperationType:      domain.PluginOperationTypeUpgrade,
		PluginCode:         "demo_notice_install",
		FromVersion:        "1.0.0",
		ToVersion:          "1.1.0",
		BeforeManifestJSON: string(beforeRaw),
		Status:             domain.PluginOperationStatusApplied,
	})

	resp, err := svc.PluginUpgradeRollbackDryRun(p.Code, domain.PluginUpgradeRollbackDryRunRequest{OperationID: "op_test_upgrade_before"})
	if err != nil {
		t.Fatalf("PluginUpgradeRollbackDryRun: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("expected ok, got %q", resp.Status)
	}
	if len(resp.DiffSections) == 0 {
		t.Fatalf("expected diff sections")
	}
	if resp.DiffSummary.Changed == 0 && resp.DiffSummary.Added == 0 && resp.DiffSummary.Removed == 0 {
		t.Fatalf("expected non-empty diff summary")
	}
}
