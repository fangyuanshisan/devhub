package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestVerifyPluginPackageSignatureForPrecheckAs_Unsigned(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	download, _ := repo.AppendPluginPackageDownload(domain.PluginPackageDownloadRecord{
		PluginCode:     "sig_demo",
		Version:        "1.0.0",
		Status:         domain.PluginPackageDownloadStatusDownloaded,
		SHA256Actual:   "aa",
		SHA256Expected: "aa",
		CreatedAt:      Now(),
		UpdatedAt:      Now(),
	})
	pre, _ := repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PackageDownloadID: download.ID,
		PluginCode:        "sig_demo",
		Version:           "1.0.0",
		Status:            domain.PluginPackagePrecheckStatusPassed,
		PackagePath:       "storage/test-sig/unsigned",
		ChecksumStatus:    "ok",
		CreatedBy:         1,
		CreatedAt:         Now(),
		UpdatedAt:         Now(),
	})

	root, _ := serviceProjectRoot()
	pkgAbs := filepath.Join(root, filepath.FromSlash(pre.PackagePath))
	if err := os.MkdirAll(pkgAbs, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	manifestRaw := []byte(`{"code":"sig_demo","name":"Sig Demo","version":"1.0.0","compatible_core_version":">=1.7.0 <2.0.0"}`)
	if err := os.WriteFile(filepath.Join(pkgAbs, "manifest.json"), manifestRaw, 0o644); err != nil {
		t.Fatalf("WriteFile manifest: %v", err)
	}

	got, err := svc.VerifyPluginPackageSignatureForPrecheckAs(PluginPackageSignatureOperator{ID: 1, Name: "tester"}, pre.ID)
	if err != nil {
		t.Fatalf("VerifyPluginPackageSignatureForPrecheckAs: %v", err)
	}
	if got.Status != domain.PluginPackageSignatureStatusUnsigned {
		t.Fatalf("expected unsigned, got %s", got.Status)
	}
}

func TestVerifyPluginPackageSignatureForPrecheckAs_Verified(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	publisherID := "devhub-official"
	keyID := "devhub-official-2026-01"
	_, err = repo.AppendPluginTrustedPublisher(domain.PluginTrustedPublisher{
		PublisherID:        publisherID,
		Name:               "DevHub Official",
		PublicKeyID:        keyID,
		PublicKeyAlgorithm: "ed25519",
		PublicKey:          base64.StdEncoding.EncodeToString(pub),
		Status:             "trusted",
		CreatedAt:          Now(),
		UpdatedAt:          Now(),
	})
	if err != nil {
		t.Fatalf("AppendPluginTrustedPublisher: %v", err)
	}

	packageBytes := []byte("fake package bytes")
	sum := sha256.Sum256(packageBytes)
	packageSHA := hex.EncodeToString(sum[:])
	download, _ := repo.AppendPluginPackageDownload(domain.PluginPackageDownloadRecord{
		PluginCode:     "sig_demo",
		Version:        "1.0.0",
		Status:         domain.PluginPackageDownloadStatusDownloaded,
		SHA256Actual:   packageSHA,
		SHA256Expected: packageSHA,
		CreatedAt:      Now(),
		UpdatedAt:      Now(),
	})

	pre := domain.PluginPackagePrecheckRecord{
		PackageDownloadID: download.ID,
		PluginCode:        "sig_demo",
		Version:           "1.0.0",
		Status:            domain.PluginPackagePrecheckStatusPassed,
		ChecksumStatus:    "ok",
		CreatedBy:         1,
		CreatedAt:         Now(),
		UpdatedAt:         Now(),
	}
	pre, _ = repo.AppendPluginPackagePrecheck(pre)

	root, _ := serviceProjectRoot()
	pkgRel := filepath.ToSlash(filepath.Join("storage/test-sig", "verified-"+strconv.FormatInt(pre.ID, 10)))
	pre.PackagePath = pkgRel
	pre, _ = repo.SavePluginPackagePrecheck(pre)

	pkgAbs := filepath.Join(root, filepath.FromSlash(pkgRel))
	if err := os.MkdirAll(pkgAbs, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	manifestRaw := []byte(`{"code":"sig_demo","name":"Sig Demo","version":"1.0.0","compatible_core_version":">=1.7.0 <2.0.0"}`)
	if err := os.WriteFile(filepath.Join(pkgAbs, "manifest.json"), manifestRaw, 0o644); err != nil {
		t.Fatalf("WriteFile manifest: %v", err)
	}
	manifestSum := sha256.Sum256(manifestRaw)
	manifestSHA := hex.EncodeToString(manifestSum[:])

	manifest, _, err := pluginregistry.DecodePluginManifestJSON(manifestRaw)
	if err != nil {
		t.Fatalf("DecodePluginManifestJSON: %v", err)
	}
	payload := domain.DevHubSignaturePayload{
		PluginCode:            manifest.Code,
		Version:               manifest.Version,
		PackageSHA256:         packageSHA,
		ManifestSHA256:        manifestSHA,
		PublisherID:           publisherID,
		KeyID:                 keyID,
		CompatibleCoreVersion: manifest.CompatibleCoreVersion,
	}
	canonical, _ := pluginregistry.CanonicalSignaturePayloadBytes(payload)
	sigBytes := ed25519.Sign(priv, canonical)
	sigFile := domain.DevHubDetachedSignatureFile{
		SchemaVersion:    "1",
		Algorithm:        "Ed25519",
		PublisherID:      publisherID,
		KeyID:            keyID,
		PluginCode:       manifest.Code,
		Version:          manifest.Version,
		PackageSHA256:    packageSHA,
		ManifestSHA256:   manifestSHA,
		Signature:        base64.StdEncoding.EncodeToString(sigBytes),
		SignaturePayload: payload,
	}
	raw, _ := json.Marshal(sigFile)
	if err := os.WriteFile(filepath.Join(pkgAbs, "devhub-signature.json"), raw, 0o644); err != nil {
		t.Fatalf("WriteFile signature: %v", err)
	}

	got, err := svc.VerifyPluginPackageSignatureForPrecheckAs(PluginPackageSignatureOperator{ID: 1, Name: "tester"}, pre.ID)
	if err != nil {
		t.Fatalf("VerifyPluginPackageSignatureForPrecheckAs: %v", err)
	}
	if got.Status != domain.PluginPackageSignatureStatusVerified {
		t.Fatalf("expected verified, got %s err=%s", got.Status, got.ErrorMessage)
	}
}

func TestVerifyPluginPackageSignatureForPrecheckAs_KeyExpired(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	publisherID := "devhub-official"
	keyID := "devhub-official-2026-01"
	_, err = repo.AppendPluginTrustedPublisher(domain.PluginTrustedPublisher{
		PublisherID:        publisherID,
		Name:               "DevHub Official",
		PublicKeyID:        keyID,
		PublicKeyAlgorithm: "ed25519",
		PublicKey:          base64.StdEncoding.EncodeToString(pub),
		Status:             "trusted",
		ExpiresAt:          "2000-01-01 00:00:00",
		CreatedAt:          Now(),
		UpdatedAt:          Now(),
	})
	if err != nil {
		t.Fatalf("AppendPluginTrustedPublisher: %v", err)
	}

	packageBytes := []byte("fake package bytes")
	sum := sha256.Sum256(packageBytes)
	packageSHA := hex.EncodeToString(sum[:])
	download, _ := repo.AppendPluginPackageDownload(domain.PluginPackageDownloadRecord{
		PluginCode:     "sig_demo",
		Version:        "1.0.0",
		Status:         domain.PluginPackageDownloadStatusDownloaded,
		SHA256Actual:   packageSHA,
		SHA256Expected: packageSHA,
		CreatedAt:      Now(),
		UpdatedAt:      Now(),
	})
	pre, _ := repo.AppendPluginPackagePrecheck(domain.PluginPackagePrecheckRecord{
		PackageDownloadID: download.ID,
		PluginCode:        "sig_demo",
		Version:           "1.0.0",
		Status:            domain.PluginPackagePrecheckStatusPassed,
		ChecksumStatus:    "ok",
		CreatedBy:         1,
		CreatedAt:         Now(),
		UpdatedAt:         Now(),
	})

	root, _ := serviceProjectRoot()
	pkgRel := filepath.ToSlash(filepath.Join("storage/test-sig", "expired-"+strconv.FormatInt(pre.ID, 10)))
	pre.PackagePath = pkgRel
	pre, _ = repo.SavePluginPackagePrecheck(pre)
	pkgAbs := filepath.Join(root, filepath.FromSlash(pkgRel))
	if err := os.MkdirAll(pkgAbs, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	manifestRaw := []byte(`{"code":"sig_demo","name":"Sig Demo","version":"1.0.0","compatible_core_version":">=1.7.0 <2.0.0"}`)
	_ = os.WriteFile(filepath.Join(pkgAbs, "manifest.json"), manifestRaw, 0o644)
	manifestSum := sha256.Sum256(manifestRaw)
	manifestSHA := hex.EncodeToString(manifestSum[:])
	manifest, _, _ := pluginregistry.DecodePluginManifestJSON(manifestRaw)

	payload := domain.DevHubSignaturePayload{
		PluginCode:            manifest.Code,
		Version:               manifest.Version,
		PackageSHA256:         packageSHA,
		ManifestSHA256:        manifestSHA,
		PublisherID:           publisherID,
		KeyID:                 keyID,
		CompatibleCoreVersion: manifest.CompatibleCoreVersion,
	}
	canonical, _ := pluginregistry.CanonicalSignaturePayloadBytes(payload)
	sigBytes := ed25519.Sign(priv, canonical)
	sigFile := domain.DevHubDetachedSignatureFile{
		SchemaVersion:    "1",
		Algorithm:        "Ed25519",
		PublisherID:      publisherID,
		KeyID:            keyID,
		PluginCode:       manifest.Code,
		Version:          manifest.Version,
		PackageSHA256:    packageSHA,
		ManifestSHA256:   manifestSHA,
		Signature:        base64.StdEncoding.EncodeToString(sigBytes),
		SignaturePayload: payload,
	}
	raw, _ := json.Marshal(sigFile)
	_ = os.WriteFile(filepath.Join(pkgAbs, "devhub-signature.json"), raw, 0o644)

	got, err := svc.VerifyPluginPackageSignatureForPrecheckAs(PluginPackageSignatureOperator{ID: 1, Name: "tester"}, pre.ID)
	if err == nil {
		t.Fatalf("expected error")
	}
	apiErr, _ := err.(*domain.APIError)
	if apiErr == nil || apiErr.Code != "plugin_package_signature_key_expired" {
		t.Fatalf("expected key_expired error, got %v", err)
	}
	if got.Status != domain.PluginPackageSignatureStatusKeyExpired {
		t.Fatalf("expected key_expired status, got %s", got.Status)
	}
}
