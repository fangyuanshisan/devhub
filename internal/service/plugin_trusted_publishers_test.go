package service

import (
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

const testEd25519PublicKey = "Jf8LxG7EtiK11AHz1Ce3zxPdzIoIk0cUWmPFsNyJAi0="

func TestTrustedPublishers_CreateDuplicateAndStatus(t *testing.T) {
	svc := New(store.NewMemoryStore())
	operator := TrustedPublisherOperator{ID: 1, Name: "admin"}

	req := domain.PluginTrustedPublisher{
		PublisherID:        "fixture-publisher",
		Name:               "Fixture Publisher",
		PublicKeyID:        "fixture-key-2026",
		PublicKeyAlgorithm: "ed25519",
		PublicKey:          testEd25519PublicKey,
		Homepage:           "https://example.com",
	}
	created, err := svc.CreatePluginTrustedPublisher(operator, req)
	if err != nil {
		t.Fatalf("create publisher: %v", err)
	}
	if created.ID == 0 || created.Status != "trusted" || created.Fingerprint == "" {
		t.Fatalf("unexpected created publisher: %#v", created)
	}

	if _, err := svc.CreatePluginTrustedPublisher(operator, req); err == nil {
		t.Fatalf("expected duplicate error")
	} else if api, ok := err.(*domain.APIError); !ok || api.Code != "plugin_trusted_publisher_duplicate" {
		t.Fatalf("unexpected duplicate error: %T %v", err, err)
	}

	blocked, err := svc.SetPluginTrustedPublisherStatus(operator, created.ID, "blocked", "test block")
	if err != nil {
		t.Fatalf("block publisher: %v", err)
	}
	if blocked.Status != "blocked" || blocked.BlockedAt == "" {
		t.Fatalf("expected blocked publisher, got %#v", blocked)
	}

	restored, err := svc.SetPluginTrustedPublisherStatus(operator, created.ID, "trusted", "test restore")
	if err != nil {
		t.Fatalf("restore publisher: %v", err)
	}
	if restored.Status != "trusted" || restored.BlockedAt != "" || restored.RevokedAt != "" {
		t.Fatalf("expected restored publisher, got %#v", restored)
	}
}

func TestTrustedPublishers_InvalidKeyRejected(t *testing.T) {
	svc := New(store.NewMemoryStore())
	operator := TrustedPublisherOperator{ID: 1, Name: "admin"}

	_, err := svc.CreatePluginTrustedPublisher(operator, domain.PluginTrustedPublisher{
		PublisherID:        "bad",
		Name:               "Bad",
		PublicKeyID:        "bad-key",
		PublicKeyAlgorithm: "rsa",
		PublicKey:          testEd25519PublicKey,
	})
	if err == nil {
		t.Fatalf("expected invalid algorithm error")
	}
	if api, ok := err.(*domain.APIError); !ok || api.Code != "plugin_trusted_publisher_invalid_key" {
		t.Fatalf("unexpected error: %T %v", err, err)
	}

	_, err = svc.CreatePluginTrustedPublisher(operator, domain.PluginTrustedPublisher{
		PublisherID:        "bad",
		Name:               "Bad",
		PublicKeyID:        "bad-key",
		PublicKeyAlgorithm: "ed25519",
		PublicKey:          "not-base64",
	})
	if err == nil {
		t.Fatalf("expected invalid public key error")
	}
	if api, ok := err.(*domain.APIError); !ok || api.Code != "plugin_trusted_publisher_invalid_key" {
		t.Fatalf("unexpected error: %T %v", err, err)
	}
}

func TestTrustedPublisher_BlockChangesPackageSignatureRisk(t *testing.T) {
	svc := New(store.NewMemoryStore())
	list, err := svc.ListPluginTrustedPublishers(domain.PluginTrustedPublisherFilter{Keyword: "devhub-official", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list publisher: %v", err)
	}
	if len(list.Items) == 0 {
		t.Fatalf("expected seeded devhub-official publisher")
	}
	if _, err := svc.SetPluginTrustedPublisherStatus(TrustedPublisherOperator{ID: 1, Name: "admin"}, list.Items[0].ID, "blocked", "test block"); err != nil {
		t.Fatalf("block publisher: %v", err)
	}

	res, err := svc.DryRunPluginPackage("examples/plugins/demo_signed_notice")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.Status != "blocked" {
		t.Fatalf("expected blocked package after publisher block, got %#v", res.Signature)
	}
	if res.Signature.TrustStatus != "blocked" {
		t.Fatalf("expected blocked trust status, got %#v", res.Signature)
	}
}
