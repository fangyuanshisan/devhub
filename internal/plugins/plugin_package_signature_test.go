package plugins_test

import (
	"testing"

	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestPluginPackageSignature_TrustedVerified(t *testing.T) {
	svc := service.New(store.NewMemoryStore())

	res, err := svc.DryRunPluginPackage("examples/plugins/demo_signed_notice")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.Status == "blocked" {
		t.Fatalf("unexpected blocked: %#v", res)
	}
	if !res.Signature.SignatureFound {
		t.Fatalf("expected signature_found=true, got %#v", res.Signature)
	}
	if !res.Signature.PublisherFound {
		t.Fatalf("expected publisher_found=true, got %#v", res.Signature)
	}
	if res.Signature.VerificationStatus != "verified" {
		t.Fatalf("expected verified, got %#v", res.Signature)
	}
	if res.Signature.TrustStatus != "trusted" {
		t.Fatalf("expected trusted, got %#v", res.Signature)
	}
	if res.Signature.PublisherID != "devhub-official" {
		t.Fatalf("unexpected publisher_id: %#v", res.Signature)
	}
}

func TestPluginPackageSignature_UnknownPublisherVerified(t *testing.T) {
	svc := service.New(store.NewMemoryStore())

	res, err := svc.DryRunPluginPackage("examples/plugins/security-fixtures/signature_unknown_publisher")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.Status == "blocked" {
		t.Fatalf("unexpected blocked: %#v", res)
	}
	if res.Signature.VerificationStatus != "verified" {
		t.Fatalf("expected verified, got %#v", res.Signature)
	}
	if res.Signature.TrustStatus != "unknown" {
		t.Fatalf("expected unknown trust, got %#v", res.Signature)
	}
	if res.RiskReport.Level != "high" && res.RiskReport.Level != "medium" {
		// keep tolerant: risk rules may be tuned
		t.Fatalf("unexpected risk level: %#v", res.RiskReport)
	}
}

func TestPluginPackageSignature_UnsupportedAlgorithmBlocked(t *testing.T) {
	svc := service.New(store.NewMemoryStore())

	res, err := svc.DryRunPluginPackage("examples/plugins/security-fixtures/signature_unsupported_algorithm")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.Status != "blocked" {
		t.Fatalf("expected blocked, got %#v", res)
	}
	if res.BlockedCode != "plugin_package_signature_unsupported_algorithm" {
		t.Fatalf("unexpected blocked_code: %#v", res.BlockedCode)
	}
}

func TestPluginPackageSignature_PublisherRevokedBlocked(t *testing.T) {
	svc := service.New(store.NewMemoryStore())

	res, err := svc.DryRunPluginPackage("examples/plugins/security-fixtures/publisher_revoked")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.Status != "blocked" {
		t.Fatalf("expected blocked, got %#v", res)
	}
	if res.BlockedCode != "plugin_package_publisher_revoked" {
		t.Fatalf("unexpected blocked_code: %#v", res.BlockedCode)
	}
}
