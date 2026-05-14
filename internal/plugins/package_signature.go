package plugins

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

type publisherWire struct {
	PublisherID        string `json:"publisher_id"`
	Name               string `json:"name"`
	Homepage           string `json:"homepage"`
	Email              string `json:"email"`
	PublicKeyID        string `json:"public_key_id"`
	PublicKeyAlgorithm string `json:"public_key_algorithm"`
	PublicKey          string `json:"public_key"`
	TrustLevel         string `json:"trust_level"`
}

type signatureWire struct {
	Version     string   `json:"version"`
	Algorithm   string   `json:"algorithm"`
	SignedAt    string   `json:"signed_at"`
	PublisherID string   `json:"publisher_id"`
	PublicKeyID string   `json:"public_key_id"`
	SignedFiles []string `json:"signed_files"`
	Signature   string   `json:"signature"`
}

// VerifyPluginPackageSignature validates publisher.json + signature.json, matches local trusted publishers,
// and (when possible) verifies the signature.
//
// Signature payload (v1.5.0 P2-09):
// - message = sha256(raw_bytes_of_checksums.json)
// - signature = ed25519.Sign(public_key, message)
//
// Notes:
// - Missing signature.json does NOT block dry-run; it returns trust_status=unsigned + verification_status=missing.
// - Missing publisher.json does NOT block dry-run; trust can still be determined from trusted_publishers via signature publisher_id/public_key_id.
func VerifyPluginPackageSignature(packageDir string, scan domain.PluginPackageFileScan, checksum domain.PluginPackageChecksumResult) (domain.PluginPackageSignatureResult, error) {
	pubPath := filepath.Join(packageDir, "publisher.json")
	sigPath := filepath.Join(packageDir, "signature.json")

	res := domain.PluginPackageSignatureResult{
		SignatureFound:     false,
		PublisherFound:     false,
		TrustStatus:        "unsigned",
		VerificationStatus: "missing",
		SignedFiles:        []string{},
		UnsignedFiles:      []string{},
		Messages:           []string{},
	}

	publisherRaw, pubErr := os.ReadFile(pubPath)
	if pubErr == nil {
		res.PublisherFound = true
		var wire publisherWire
		if err := json.Unmarshal(publisherRaw, &wire); err != nil {
			return res, domain.NewPluginError("plugin_package_publisher_invalid", "publisher.json 不是合法 JSON").
				WithStatus(400).
				WithDetail("path", "publisher.json").
				WithDetail("reason", strings.TrimSpace(err.Error())).
				WithSuggestion("请修复 publisher.json 后重试。")
		}
	} else if pubErr != nil && !os.IsNotExist(pubErr) {
		return res, domain.NewPluginError("plugin_package_publisher_invalid", "读取 publisher.json 失败").
			WithStatus(400).
			WithDetail("path", "publisher.json").
			WithDetail("reason", strings.TrimSpace(pubErr.Error())).
			WithSuggestion("请检查插件包文件是否可读。")
	}

	sigRaw, sigErr := os.ReadFile(sigPath)
	if sigErr != nil {
		if os.IsNotExist(sigErr) {
			res.Messages = append(res.Messages, "未提供 signature.json（未签名）")
			return res, nil
		}
		return res, domain.NewPluginError("plugin_package_signature_invalid", "读取 signature.json 失败").
			WithStatus(400).
			WithDetail("path", "signature.json").
			WithDetail("reason", strings.TrimSpace(sigErr.Error())).
			WithSuggestion("请检查插件包文件是否可读。")
	}
	res.SignatureFound = true

	var wire signatureWire
	if err := json.Unmarshal(sigRaw, &wire); err != nil {
		return res, domain.NewPluginError("plugin_package_signature_invalid", "signature.json 不是合法 JSON").
			WithStatus(400).
			WithDetail("path", "signature.json").
			WithDetail("reason", strings.TrimSpace(err.Error())).
			WithSuggestion("请修复 signature.json 后重试。")
	}

	algo := strings.ToLower(strings.TrimSpace(wire.Algorithm))
	if algo == "" {
		algo = "ed25519"
	}
	res.Algorithm = algo
	res.PublisherID = strings.TrimSpace(wire.PublisherID)
	res.PublicKeyID = strings.TrimSpace(wire.PublicKeyID)

	if algo != "ed25519" {
		res.TrustStatus = "unknown"
		res.VerificationStatus = "unsupported"
		return res, domain.NewPluginError("plugin_package_signature_unsupported_algorithm", "不支持的签名算法").
			WithStatus(400).
			WithDetail("algorithm", wire.Algorithm).
			WithSuggestion("当前仅支持 ed25519。")
	}

	// Validate signed_files.
	signedFiles := normalizeSignedFiles(wire.SignedFiles)
	res.SignedFiles = signedFiles
	res.SignedFilesCount = len(signedFiles)
	if !containsString(signedFiles, "manifest.json") {
		res.VerificationStatus = "failed"
		res.TrustStatus = "unknown"
		return res, domain.NewPluginError("plugin_package_signature_manifest_unsigned", "signature.json 必须包含 manifest.json").
			WithStatus(400).
			WithDetail("path", "signature.json").
			WithDetail("missing", "manifest.json").
			WithSuggestion("请在 signed_files 中加入 manifest.json 后重试。")
	}
	if !containsString(signedFiles, "checksums.json") {
		res.VerificationStatus = "failed"
		res.TrustStatus = "unknown"
		return res, domain.NewPluginError("plugin_package_signature_signed_file_missing", "signature.json 必须包含 checksums.json").
			WithStatus(400).
			WithDetail("path", "signature.json").
			WithDetail("missing", "checksums.json").
			WithSuggestion("请在 signed_files 中加入 checksums.json 后重试。")
	}
	for _, p := range signedFiles {
		if strings.HasPrefix(p, "/") || filepath.IsAbs(p) {
			return res, domain.NewPluginError("plugin_package_signature_path_invalid", "signed_files 不允许为绝对路径").
				WithStatus(400).
				WithDetail("path", p).
				WithSuggestion("请使用插件包内相对路径。")
		}
		if strings.HasPrefix(p, "../") || p == ".." || strings.Contains(p, "/../") {
			return res, domain.NewPluginError("plugin_package_signature_path_invalid", "signed_files 不允许包含 ..").
				WithStatus(400).
				WithDetail("path", p).
				WithSuggestion("请移除路径穿越段。")
		}
		full := filepath.Join(packageDir, filepath.FromSlash(p))
		if _, err := os.Stat(full); err != nil {
			return res, domain.NewPluginError("plugin_package_signature_signed_file_missing", "signature.json 声明的 signed_files 文件不存在").
				WithStatus(400).
				WithDetail("path", p).
				WithSuggestion("请修复 signed_files 列表，或补齐缺失文件。")
		}
		// Ensure signed files are not dangerous by scan result.
		if isDangerousByScan(scan, p) {
			return res, domain.NewPluginError("plugin_package_dangerous_file", "签名包含危险文件，禁止").
				WithStatus(400).
				WithDetail("path", p).
				WithSuggestion("请移除危险文件后重新生成签名。")
		}
	}

	// Signature must not cover files outside checksums.json coverage (except manifest/checksums themselves).
	declared, derr := declaredChecksumPaths(packageDir)
	if derr == nil && len(declared) > 0 {
		for _, p := range signedFiles {
			if p == "manifest.json" || p == "checksums.json" {
				continue
			}
			if !declared[p] {
				return res, domain.NewPluginError("plugin_package_signature_invalid", "signature.json signed_files 不允许包含未被 checksums.json 覆盖的文件").
					WithStatus(400).
					WithDetail("path", p).
					WithSuggestion("请先将该文件加入 checksums.json，再重新生成签名。")
			}
		}
	}

	// Determine trust status by local trusted publishers config.
	trustedCfg, _, trustedErr := LoadTrustedPublishers()
	if trustedErr != nil {
		// Not blocking; keep trust unknown but surface message.
		res.Messages = append(res.Messages, "本地 trusted_publishers 不可用，无法判断发布者是否可信")
	}
	match := FindTrustedPublisher(trustedCfg, res.PublisherID, res.PublicKeyID)
	trustStatus := match.NormalizedStatus()
	if trustStatus == "blocked" {
		res.TrustStatus = "blocked"
		res.VerificationStatus = "failed"
		return res, domain.NewPluginError("plugin_package_publisher_blocked", "发布者已被本地策略 blocked，禁止导入预览").
			WithStatus(400).
			WithDetail("publisher_id", res.PublisherID).
			WithDetail("public_key_id", res.PublicKeyID).
			WithSuggestion("请移除该 publisher，或在 trusted_publishers 中调整为 trusted/unknown 后重试。")
	}
	if trustStatus == "revoked" {
		res.TrustStatus = "revoked"
		res.VerificationStatus = "failed"
		return res, domain.NewPluginError("plugin_package_publisher_revoked", "发布者已被本地策略 revoked，禁止导入预览").
			WithStatus(400).
			WithDetail("publisher_id", res.PublisherID).
			WithDetail("public_key_id", res.PublicKeyID).
			WithSuggestion("请更换可信发布者签名或移除该插件包。")
	}
	res.TrustStatus = trustStatus

	// Verify signature over sha256(checksums.json raw bytes).
	checksumsRaw, err := os.ReadFile(filepath.Join(packageDir, "checksums.json"))
	if err != nil {
		res.VerificationStatus = "failed"
		return res, domain.NewPluginError("plugin_package_signature_signed_file_missing", "未找到 checksums.json，无法验签").
			WithStatus(400).
			WithDetail("path", "checksums.json").
			WithSuggestion("请补齐 checksums.json 后重试。")
	}
	msg := sha256.Sum256(checksumsRaw)

	sigB64 := strings.TrimSpace(wire.Signature)
	sigBytes, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		res.VerificationStatus = "failed"
		return res, domain.NewPluginError("plugin_package_signature_invalid", "signature 字段不是合法的 base64 ed25519 签名").
			WithStatus(400).
			WithDetail("path", "signature.json").
			WithSuggestion("请重新生成 signature.json 后重试。")
	}

	var pubKeyBytes []byte
	if match.Found && strings.EqualFold(strings.TrimSpace(match.Publisher.PublicKeyAlgorithm), "ed25519") {
		if b, err := match.PublicKeyBytes(); err == nil && len(b) == ed25519.PublicKeySize {
			pubKeyBytes = b
		}
	}
	if len(pubKeyBytes) == 0 && res.PublisherFound {
		// Fallback to package publisher.json public key for technical verification (still unknown trust unless matched).
		var p publisherWire
		if err := json.Unmarshal(publisherRaw, &p); err == nil {
			if strings.EqualFold(strings.TrimSpace(p.PublicKeyAlgorithm), "ed25519") {
				b, derr := base64.StdEncoding.DecodeString(strings.TrimSpace(p.PublicKey))
				if derr == nil && len(b) == ed25519.PublicKeySize {
					pubKeyBytes = b
				}
			}
		}
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		// We cannot verify without a public key; treat as structural only (not blocked).
		res.VerificationStatus = "structural_only"
		if res.TrustStatus == "trusted" {
			res.TrustStatus = "unknown"
		}
		res.Messages = append(res.Messages, "缺少可用 public_key，未完成真实验签（仅结构校验）")
		return res, nil
	}

	if !ed25519.Verify(ed25519.PublicKey(pubKeyBytes), msg[:], sigBytes) {
		res.VerificationStatus = "failed"
		return res, domain.NewPluginError("plugin_package_signature_verification_failed", "签名验签失败，可能文件已被篡改").
			WithStatus(400).
			WithDetail("publisher_id", res.PublisherID).
			WithDetail("public_key_id", res.PublicKeyID).
			WithSuggestion("请检查 checksums.json 与签名是否匹配，或重新生成签名后重试。")
	}

	res.VerificationStatus = "verified"
	if res.TrustStatus == "" {
		res.TrustStatus = "unknown"
	}
	if res.TrustStatus == "unknown" {
		res.Messages = append(res.Messages, "验签通过，但 publisher 未在本地 trusted_publishers 中标记 trusted")
	}

	// Compute unsigned_files: declared checksum files + manifest/checksums not included in signed_files.
	unsigned := computeUnsignedFiles(signedFiles, checksum, scan)
	res.UnsignedFiles = unsigned
	if len(unsigned) > 0 {
		res.Messages = append(res.Messages, fmt.Sprintf("存在 %d 个未被签名覆盖的文件", len(unsigned)))
	}
	return res, nil
}

func declaredChecksumPaths(packageDir string) (map[string]bool, error) {
	raw, err := os.ReadFile(filepath.Join(packageDir, "checksums.json"))
	if err != nil {
		return nil, err
	}
	var wire checksumsWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, f := range wire.Files {
		p := filepath.ToSlash(filepath.Clean(strings.TrimSpace(f.Path)))
		p = strings.TrimPrefix(p, "./")
		if p == "" || p == "." {
			continue
		}
		out[p] = true
	}
	return out, nil
}

func normalizeSignedFiles(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, it := range items {
		p := filepath.ToSlash(filepath.Clean(strings.TrimSpace(it)))
		if p == "" || p == "." {
			continue
		}
		p = strings.TrimPrefix(p, "./")
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func containsString(items []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, it := range items {
		if strings.TrimSpace(it) == target {
			return true
		}
	}
	return false
}

func isDangerousByScan(scan domain.PluginPackageFileScan, path string) bool {
	p := filepath.ToSlash(strings.TrimSpace(path))
	for _, it := range scan.DangerousFiles {
		if filepath.ToSlash(strings.TrimSpace(it.Path)) == p {
			return true
		}
	}
	return false
}

func computeUnsignedFiles(signedFiles []string, checksum domain.PluginPackageChecksumResult, scan domain.PluginPackageFileScan) []string {
	want := map[string]bool{}
	// For a stable warning, consider all non-dangerous files (allowed + unknown) excluding checksums.json itself.
	for _, it := range append(scan.AllowedFiles, scan.UnknownFiles...) {
		p := filepath.ToSlash(strings.TrimSpace(it.Path))
		if p == "" || p == "checksums.json" || p == "signature.json" {
			continue
		}
		want[p] = true
	}
	// Also include manifest/checksums explicitly.
	want["manifest.json"] = true
	want["checksums.json"] = true

	for _, it := range signedFiles {
		delete(want, filepath.ToSlash(strings.TrimSpace(it)))
	}
	out := make([]string, 0, len(want))
	for p := range want {
		out = append(out, p)
	}
	sort.Strings(out)

	// If checksums status is missing/failed, unsigned_files signal is not meaningful; keep best-effort list.
	_ = checksum
	return out
}
