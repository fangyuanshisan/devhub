package service

import (
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestPluginApprovals_Install_CreateApproveExecute(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	created, err := svc.CreatePluginApproval(PluginApprovalOperator{ID: 1, Name: "admin"}, domain.PluginApprovalCreateRequest{
		Action:      "install",
		PackagePath: "plugins-local/repository-fixtures/demo_notice_install",
		Reason:      "安装示例插件用于验收审批流",
	})
	if err != nil {
		t.Fatalf("CreatePluginApproval: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected id")
	}
	if created.Status != domain.PluginApprovalStatusPending {
		t.Fatalf("expected pending, got %q", created.Status)
	}
	if created.PluginCode != "demo_notice_install" {
		t.Fatalf("unexpected plugin_code: %q", created.PluginCode)
	}
	if strings.Contains(created.DryRunJSON, "enc:v1:") || strings.Contains(created.RiskReportJSON, "enc:v1:") {
		t.Fatalf("should not persist ciphertext markers")
	}

	approved, err := svc.ApprovePluginApproval(PluginApprovalOperator{ID: 1, Name: "admin"}, created.ID, "通过")
	if err != nil {
		t.Fatalf("ApprovePluginApproval: %v", err)
	}
	if approved.Status != domain.PluginApprovalStatusApproved {
		t.Fatalf("expected approved, got %q", approved.Status)
	}

	executed, err := svc.ExecutePluginApproval(PluginApprovalOperator{ID: 1, Name: "admin"}, created.ID)
	if err != nil {
		t.Fatalf("ExecutePluginApproval: %v", err)
	}
	if executed.Status != domain.PluginApprovalStatusExecuted {
		t.Fatalf("expected executed, got %q", executed.Status)
	}
	if _, ok := repo.PluginByCode("demo_notice_install"); !ok {
		t.Fatalf("expected plugin installed")
	}
	plugin, _ := repo.PluginByCode("demo_notice_install")
	if plugin.Status != "disabled" {
		t.Fatalf("expected installed plugin disabled, got %q", plugin.Status)
	}
}

func TestPluginApprovals_RejectRequiresComment(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	created, err := svc.CreatePluginApproval(PluginApprovalOperator{ID: 1, Name: "admin"}, domain.PluginApprovalCreateRequest{
		Action:      "install",
		PackagePath: "plugins-local/repository-fixtures/demo_notice_install",
	})
	if err != nil {
		t.Fatalf("CreatePluginApproval: %v", err)
	}
	_, err = svc.RejectPluginApproval(PluginApprovalOperator{ID: 1, Name: "admin"}, created.ID, "")
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_approval_reject_reason_required" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}
