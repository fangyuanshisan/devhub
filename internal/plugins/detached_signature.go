package plugins

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
)

const (
	DetachedSignatureAlgorithmEd25519 = "ed25519"
)

// CanonicalSignaturePayloadBytes returns stable canonical JSON bytes for signing/verifying.
//
// Important:
// - payload must be a struct with stable field order; do not pass a map.
func CanonicalSignaturePayloadBytes(payload domain.DevHubSignaturePayload) ([]byte, error) {
	return json.Marshal(payload)
}

type DetachedSignatureVerifyResult struct {
	Status       string   `json:"status"`
	Warnings     []string `json:"warnings,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
}

func VerifyDetachedPluginPackageSignature(
	sig domain.DevHubDetachedSignatureFile,
	manifest domain.PluginManifest,
	packageSHA256 string,
	manifestSHA256 string,
	trusted domain.PluginTrustedPublisher,
) (DetachedSignatureVerifyResult, error) {
	res := DetachedSignatureVerifyResult{Status: domain.PluginPackageSignatureStatusFailed}

	algo := strings.ToLower(strings.TrimSpace(sig.Algorithm))
	if algo == "" {
		algo = strings.ToLower(strings.TrimSpace(sig.SignaturePayload.KeyID)) // no-op fallback
	}
	if !strings.EqualFold(strings.TrimSpace(sig.SchemaVersion), "1") {
		res.Status = domain.PluginPackageSignatureStatusFailed
		return res, domain.NewPluginError("plugin_package_signature_invalid", "签名文件 schema_version 不受支持").
			WithStatus(400).
			WithDetail("schema_version", sig.SchemaVersion).
			WithSuggestion("请使用 schema_version=1 的 devhub-signature.json。")
	}
	if algo != DetachedSignatureAlgorithmEd25519 {
		res.Status = domain.PluginPackageSignatureStatusAlgorithmUnsupported
		return res, domain.NewPluginError("plugin_package_signature_unsupported_algorithm", "签名算法不受支持").
			WithStatus(400).
			WithDetail("algorithm", sig.Algorithm).
			WithSuggestion("请使用 Ed25519 签名。")
	}

	payload := sig.SignaturePayload
	// Require payload consistency with top-level fields.
	if strings.TrimSpace(sig.PublisherID) != "" && strings.TrimSpace(payload.PublisherID) != "" && strings.TrimSpace(sig.PublisherID) != strings.TrimSpace(payload.PublisherID) {
		res.Status = domain.PluginPackageSignatureStatusPayloadMismatch
		return res, domain.NewPluginError("plugin_package_signature_payload_invalid", "签名 payload 与顶层 publisher_id 不一致").
			WithStatus(400).
			WithSuggestion("请重新生成 devhub-signature.json。")
	}
	if strings.TrimSpace(sig.KeyID) != "" && strings.TrimSpace(payload.KeyID) != "" && strings.TrimSpace(sig.KeyID) != strings.TrimSpace(payload.KeyID) {
		res.Status = domain.PluginPackageSignatureStatusPayloadMismatch
		return res, domain.NewPluginError("plugin_package_signature_payload_invalid", "签名 payload 与顶层 key_id 不一致").
			WithStatus(400).
			WithSuggestion("请重新生成 devhub-signature.json。")
	}

	// Bind hashes and manifest identity.
	if strings.TrimSpace(payload.PackageSHA256) == "" || !strings.EqualFold(strings.TrimSpace(payload.PackageSHA256), strings.TrimSpace(packageSHA256)) {
		res.Status = domain.PluginPackageSignatureStatusHashMismatch
		return res, domain.NewPluginError("plugin_package_signature_payload_invalid", "签名 payload 的 package_sha256 与实际包摘要不一致").
			WithStatus(400).
			WithDetail("package_sha256_payload", payload.PackageSHA256).
			WithDetail("package_sha256_actual", packageSHA256).
			WithSuggestion("请确认下载链路未被替换，或重新生成签名后重试。")
	}
	if strings.TrimSpace(payload.ManifestSHA256) == "" || !strings.EqualFold(strings.TrimSpace(payload.ManifestSHA256), strings.TrimSpace(manifestSHA256)) {
		res.Status = domain.PluginPackageSignatureStatusHashMismatch
		return res, domain.NewPluginError("plugin_package_signature_payload_invalid", "签名 payload 的 manifest_sha256 与实际 manifest 摘要不一致").
			WithStatus(400).
			WithDetail("manifest_sha256_payload", payload.ManifestSHA256).
			WithDetail("manifest_sha256_actual", manifestSHA256).
			WithSuggestion("请确认 manifest.json 未被篡改，或重新生成签名后重试。")
	}
	if strings.TrimSpace(payload.PluginCode) == "" || strings.TrimSpace(payload.PluginCode) != strings.TrimSpace(manifest.Code) {
		res.Status = domain.PluginPackageSignatureStatusPayloadMismatch
		return res, domain.NewPluginError("plugin_package_signature_payload_invalid", "签名 payload 的 plugin_code 与 manifest 不一致").
			WithStatus(400).
			WithDetail("plugin_code_payload", payload.PluginCode).
			WithDetail("plugin_code_manifest", manifest.Code)
	}
	if strings.TrimSpace(payload.Version) == "" || strings.TrimSpace(payload.Version) != strings.TrimSpace(manifest.Version) {
		res.Status = domain.PluginPackageSignatureStatusPayloadMismatch
		return res, domain.NewPluginError("plugin_package_signature_payload_invalid", "签名 payload 的 version 与 manifest 不一致").
			WithStatus(400).
			WithDetail("version_payload", payload.Version).
			WithDetail("version_manifest", manifest.Version)
	}
	if strings.TrimSpace(manifest.CompatibleCoreVersion) != "" && strings.TrimSpace(payload.CompatibleCoreVersion) != "" && strings.TrimSpace(payload.CompatibleCoreVersion) != strings.TrimSpace(manifest.CompatibleCoreVersion) {
		res.Status = domain.PluginPackageSignatureStatusPayloadMismatch
		return res, domain.NewPluginError("plugin_package_signature_payload_invalid", "签名 payload 的 compatible_core_version 与 manifest 不一致").
			WithStatus(400).
			WithDetail("compatible_core_version_payload", payload.CompatibleCoreVersion).
			WithDetail("compatible_core_version_manifest", manifest.CompatibleCoreVersion)
	}

	// Trusted publisher must match.
	if strings.TrimSpace(trusted.PublisherID) == "" || strings.TrimSpace(trusted.PublicKeyID) == "" {
		res.Status = domain.PluginPackageSignatureStatusUntrustedPublisher
		return res, domain.NewPluginError("plugin_package_signature_publisher_unknown", "缺少可信发布者公钥，无法验签").
			WithStatus(400).
			WithSuggestion("请在后台添加 trusted publisher 后重试。")
	}
	if strings.TrimSpace(payload.PublisherID) != strings.TrimSpace(trusted.PublisherID) || strings.TrimSpace(payload.KeyID) != strings.TrimSpace(trusted.PublicKeyID) {
		res.Status = domain.PluginPackageSignatureStatusUntrustedPublisher
		return res, domain.NewPluginError("plugin_package_signature_publisher_unknown", "签名声明的 publisher/key 不在本地可信发布者中").
			WithStatus(400).
			WithDetail("publisher_id", payload.PublisherID).
			WithDetail("key_id", payload.KeyID).
			WithSuggestion("请确认 publisher_id/key_id 并在后台配置对应可信公钥。")
	}
	switch strings.ToLower(strings.TrimSpace(trusted.Status)) {
	case "blocked":
		res.Status = domain.PluginPackageSignatureStatusKeyRevokedOrDisabled
		return res, domain.NewPluginError("plugin_package_signature_publisher_blocked", "可信发布者已被 block，禁止使用其签名").
			WithStatus(400).
			WithDetail("publisher_id", trusted.PublisherID).
			WithDetail("key_id", trusted.PublicKeyID)
	case "revoked":
		res.Status = domain.PluginPackageSignatureStatusKeyRevokedOrDisabled
		return res, domain.NewPluginError("plugin_package_signature_publisher_revoked", "可信发布者已被 revoke，禁止使用其签名").
			WithStatus(400).
			WithDetail("publisher_id", trusted.PublisherID).
			WithDetail("key_id", trusted.PublicKeyID)
	}
	if strings.TrimSpace(trusted.ExpiresAt) != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(trusted.ExpiresAt), time.Local); err == nil {
			if !t.IsZero() && time.Now().After(t) {
				res.Status = domain.PluginPackageSignatureStatusKeyExpired
				return res, domain.NewPluginError("plugin_package_signature_key_expired", "可信发布者公钥已过期，禁止使用其签名").
					WithStatus(400).
					WithDetail("publisher_id", trusted.PublisherID).
					WithDetail("key_id", trusted.PublicKeyID).
					WithDetail("expires_at", trusted.ExpiresAt)
			}
		}
	}

	pubBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(trusted.PublicKey))
	if err != nil || len(pubBytes) != ed25519.PublicKeySize {
		res.Status = domain.PluginPackageSignatureStatusFailed
		return res, domain.NewPluginError("plugin_trusted_publisher_invalid_key", "可信发布者公钥不是合法的 Ed25519 公钥").
			WithStatus(400).
			WithDetail("publisher_id", trusted.PublisherID).
			WithDetail("key_id", trusted.PublicKeyID)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(sig.Signature))
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		res.Status = domain.PluginPackageSignatureStatusFailed
		return res, domain.NewPluginError("plugin_package_signature_invalid", "signature 字段不是合法的 base64 Ed25519 签名").
			WithStatus(400).
			WithSuggestion("请重新生成 devhub-signature.json 后重试。")
	}

	canonical, err := CanonicalSignaturePayloadBytes(payload)
	if err != nil {
		res.Status = domain.PluginPackageSignatureStatusFailed
		return res, fmt.Errorf("canonical payload encode: %w", err)
	}
	// Extra binding: also sign sha256(canonical) to be stable for large payloads.
	// But for v1.7.1 we verify over canonical bytes directly; sha256 is computed only for debug.
	_ = sha256.Sum256(canonical)

	if !ed25519.Verify(ed25519.PublicKey(pubBytes), canonical, sigBytes) {
		res.Status = domain.PluginPackageSignatureStatusFailed
		return res, domain.NewPluginError("plugin_package_signature_verification_failed", "插件包签名验签失败").
			WithStatus(400).
			WithDetail("publisher_id", payload.PublisherID).
			WithDetail("key_id", payload.KeyID).
			WithSuggestion("请确认签名与 payload 匹配，或重新从可信来源获取插件包。")
	}

	res.Status = domain.PluginPackageSignatureStatusVerified
	return res, nil
}
