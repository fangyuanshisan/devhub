package plugins

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"devhub-gin-backend/internal/domain"
)

func TestVerifyDetachedPluginPackageSignature_Verified(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	payload := domain.DevHubSignaturePayload{
		PluginCode:            "demo",
		Version:               "1.0.0",
		PackageSHA256:         "aa",
		ManifestSHA256:        "bb",
		PublisherID:           "devhub-official",
		KeyID:                 "devhub-official-2026-01",
		CompatibleCoreVersion: ">=1.7.0 <2.0.0",
	}
	canonical, err := CanonicalSignaturePayloadBytes(payload)
	if err != nil {
		t.Fatalf("CanonicalSignaturePayloadBytes: %v", err)
	}
	sigBytes := ed25519.Sign(priv, canonical)
	sigFile := domain.DevHubDetachedSignatureFile{
		SchemaVersion:  "1",
		Algorithm:      "Ed25519",
		PublisherID:    payload.PublisherID,
		KeyID:          payload.KeyID,
		PluginCode:     payload.PluginCode,
		Version:        payload.Version,
		PackageSHA256:  payload.PackageSHA256,
		ManifestSHA256: payload.ManifestSHA256,
		Signature:      base64.StdEncoding.EncodeToString(sigBytes),
		SignaturePayload: domain.DevHubSignaturePayload{
			PluginCode:            payload.PluginCode,
			Version:               payload.Version,
			PackageSHA256:         payload.PackageSHA256,
			ManifestSHA256:        payload.ManifestSHA256,
			PublisherID:           payload.PublisherID,
			KeyID:                 payload.KeyID,
			CompatibleCoreVersion: payload.CompatibleCoreVersion,
		},
	}
	manifest := domain.PluginManifest{
		Code:                  payload.PluginCode,
		Name:                  "Demo",
		Version:               payload.Version,
		CompatibleCoreVersion: payload.CompatibleCoreVersion,
	}
	trusted := domain.PluginTrustedPublisher{
		PublisherID:        payload.PublisherID,
		PublicKeyID:        payload.KeyID,
		PublicKeyAlgorithm: "ed25519",
		PublicKey:          base64.StdEncoding.EncodeToString(pub),
		Status:             "trusted",
	}
	res, err := VerifyDetachedPluginPackageSignature(sigFile, manifest, payload.PackageSHA256, payload.ManifestSHA256, trusted)
	if err != nil {
		t.Fatalf("VerifyDetachedPluginPackageSignature: %v", err)
	}
	if res.Status != domain.PluginPackageSignatureStatusVerified {
		t.Fatalf("expected verified, got %s (%s)", res.Status, res.ErrorMessage)
	}
}

func TestVerifyDetachedPluginPackageSignature_KeyExpired(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	payload := domain.DevHubSignaturePayload{
		PluginCode:     "demo",
		Version:        "1.0.0",
		PackageSHA256:  "aa",
		ManifestSHA256: "bb",
		PublisherID:    "devhub-official",
		KeyID:          "devhub-official-2026-01",
	}
	canonical, _ := CanonicalSignaturePayloadBytes(payload)
	sigBytes := ed25519.Sign(priv, canonical)
	sigFile := domain.DevHubDetachedSignatureFile{
		SchemaVersion:    "1",
		Algorithm:        "ed25519",
		PublisherID:      payload.PublisherID,
		KeyID:            payload.KeyID,
		PluginCode:       payload.PluginCode,
		Version:          payload.Version,
		PackageSHA256:    payload.PackageSHA256,
		ManifestSHA256:   payload.ManifestSHA256,
		Signature:        base64.StdEncoding.EncodeToString(sigBytes),
		SignaturePayload: payload,
	}
	manifest := domain.PluginManifest{Code: payload.PluginCode, Name: "Demo", Version: payload.Version}
	trusted := domain.PluginTrustedPublisher{
		PublisherID:        payload.PublisherID,
		PublicKeyID:        payload.KeyID,
		PublicKeyAlgorithm: "ed25519",
		PublicKey:          base64.StdEncoding.EncodeToString(pub),
		Status:             "trusted",
		ExpiresAt:          "2000-01-01 00:00:00",
	}
	res, err := VerifyDetachedPluginPackageSignature(sigFile, manifest, payload.PackageSHA256, payload.ManifestSHA256, trusted)
	if err == nil {
		t.Fatalf("expected error")
	}
	if res.Status != domain.PluginPackageSignatureStatusKeyExpired {
		t.Fatalf("expected key_expired, got %s", res.Status)
	}
}
