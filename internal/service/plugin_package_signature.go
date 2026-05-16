package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const (
	pluginPackageSignatureMaxBytesDefault = int64(64 * 1024)
	pluginPackageSignatureStagingDefault  = "storage/plugins/staging/signatures"
)

type PluginPackageSignatureOperator struct {
	ID   int64
	Name string
}

func pluginPackageSignatureMaxBytes() int64 {
	if v := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_SIGNATURE_MAX_BYTES")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return pluginPackageSignatureMaxBytesDefault
}

func pluginPackageSignatureStagingRoot() string {
	if v := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_SIGNATURE_STAGING_ROOT")); v != "" {
		return filepath.ToSlash(v)
	}
	return pluginPackageSignatureStagingDefault
}

func requireVerifiedSignatureForInstall() bool {
	// Default: require verified signature for staging/precheck driven install/upgrade.
	// Set to 0 to allow unsigned in dev/local environments (still audited and warned).
	v := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_REQUIRE_SIGNED_PACKAGES"))
	if v == "" {
		return true
	}
	return v != "0" && strings.ToLower(v) != "false"
}

func (s *Service) VerifyPluginPackageSignatureForPrecheckAs(operator PluginPackageSignatureOperator, precheckID int64) (domain.PluginPackageSignatureRecord, error) {
	precheck, ok := s.repo.PluginPackagePrecheckByID(precheckID)
	if !ok || precheck.ID <= 0 {
		return domain.PluginPackageSignatureRecord{}, domain.NewPluginError("plugin_package_precheck_not_found", "插件包预检记录不存在").
			WithStatus(404).
			WithSuggestion("请先完成插件包解压安全检查与 manifest 预校验。")
	}
	if precheck.Status != domain.PluginPackagePrecheckStatusPassed {
		return domain.PluginPackageSignatureRecord{}, domain.NewPluginError("plugin_package_signature_precheck_not_passed", "只有预检通过的插件包才能验签").
			WithStatus(400).
			WithDetail("precheck_id", precheck.ID).
			WithDetail("precheck_status", precheck.Status).
			WithSuggestion("请修复预检错误后重新执行预检。")
	}
	if precheck.PackageDownloadID <= 0 {
		return domain.PluginPackageSignatureRecord{}, domain.NewPluginError("plugin_package_signature_source_missing", "预检记录缺少 package_download_id，无法验签").
			WithStatus(400).
			WithDetail("precheck_id", precheck.ID).
			WithSuggestion("请从 staging 下载链路创建预检记录后重试。")
	}
	download, ok := s.repo.PluginPackageDownloadByID(precheck.PackageDownloadID)
	if !ok {
		return domain.PluginPackageSignatureRecord{}, domain.NewPluginError("plugin_package_download_not_found", "预检关联的 staging 下载记录不存在").
			WithStatus(404).
			WithDetail("package_download_id", precheck.PackageDownloadID)
	}
	if download.Status != domain.PluginPackageDownloadStatusDownloaded {
		return domain.PluginPackageSignatureRecord{}, domain.NewPluginError("plugin_package_signature_source_invalid", "只有 sha256 校验通过的 staging 包才能验签").
			WithStatus(400).
			WithDetail("download_id", download.ID).
			WithDetail("download_status", download.Status).
			WithSuggestion("请先提供 sha256 并确保下载校验通过。")
	}

	started := Now()
	record := domain.PluginPackageSignatureRecord{
		PackageDownloadID: download.ID,
		PackagePrecheckID: precheck.ID,
		PluginCode:        strings.TrimSpace(precheck.PluginCode),
		Version:           strings.TrimSpace(precheck.Version),
		Status:            domain.PluginPackageSignatureStatusPending,
		SignatureURL:      strings.TrimSpace(download.SignatureURL),
		PackageSHA256:     strings.TrimSpace(download.SHA256Actual),
		CreatedBy:         operator.ID,
		CreatedAt:         started,
		UpdatedAt:         started,
	}
	record, _ = s.repo.AppendPluginPackageSignature(record)

	saveFailed := func(status string, err error) (domain.PluginPackageSignatureRecord, error) {
		record.Status = status
		record.ErrorMessage = err.Error()
		record.UpdatedAt = Now()
		_, _ = s.repo.SavePluginPackageSignature(record)
		return record, err
	}

	root, err := serviceProjectRoot()
	if err != nil {
		return saveFailed(domain.PluginPackageSignatureStatusFailed, domain.NewPluginError("plugin_package_signature_failed", "读取项目根目录失败").
			WithStatus(500))
	}
	packageAbs := filepath.Join(root, filepath.FromSlash(precheck.PackagePath))
	if precheck.PackagePath == "" || !pathInside(root, packageAbs) {
		return saveFailed(domain.PluginPackageSignatureStatusFailed, domain.NewPluginError("plugin_package_signature_failed", "package_path 非法").
			WithStatus(400).
			WithDetail("package_path", precheck.PackagePath))
	}
	manifestRaw, rerr := os.ReadFile(filepath.Join(packageAbs, "manifest.json"))
	if rerr != nil {
		return saveFailed(domain.PluginPackageSignatureStatusFailed, domain.NewPluginError("plugin_package_manifest_missing", "未找到 manifest.json，无法验签").
			WithStatus(400).
			WithSuggestion("请先完成 manifest 预校验并确认解压目录存在。"))
	}
	manifestSum := sha256.Sum256(manifestRaw)
	manifestSHA := hex.EncodeToString(manifestSum[:])
	record.ManifestSHA256 = manifestSHA

	manifest, _, derr := pluginregistry.DecodePluginManifestJSON(manifestRaw)
	if derr != nil {
		return saveFailed(domain.PluginPackageSignatureStatusFailed, domain.NewPluginError("plugin_package_manifest_invalid", "manifest.json 不合法，无法验签").
			WithStatus(400).
			WithSuggestion("请修复 manifest.json 后重试。"))
	}
	if record.PluginCode == "" {
		record.PluginCode = strings.TrimSpace(manifest.Code)
	}
	if record.Version == "" {
		record.Version = strings.TrimSpace(manifest.Version)
	}

	sigURLBytes, sigURLPath, sigURLErr := s.maybeDownloadDetachedSignature(download, record.ID)
	if sigURLErr != nil {
		return saveFailed(domain.PluginPackageSignatureStatusFailed, sigURLErr)
	}
	var sigLocalBytes []byte
	localPath := filepath.Join(packageAbs, "devhub-signature.json")
	if b, err := os.ReadFile(localPath); err == nil && len(b) > 0 {
		sigLocalBytes = b
	}
	if len(sigURLBytes) == 0 && len(sigLocalBytes) == 0 {
		record.Status = domain.PluginPackageSignatureStatusUnsigned
		record.VerifiedAt = Now()
		record.UpdatedAt = Now()
		out, _ := s.repo.SavePluginPackageSignature(record)
		return out, nil
	}
	if len(sigURLBytes) > 0 && len(sigLocalBytes) > 0 && !jsonEqualBytes(sigURLBytes, sigLocalBytes) {
		return saveFailed(domain.PluginPackageSignatureStatusFailed, domain.NewPluginError("plugin_package_signature_payload_invalid", "signature_url 与包内 devhub-signature.json 不一致").
			WithStatus(400).
			WithSuggestion("请确保 detached signature 与包内签名文件一致，或仅保留一种来源。"))
	}

	sigBytes := sigURLBytes
	if len(sigBytes) == 0 {
		sigBytes = sigLocalBytes
		record.SignatureFilePath = filepath.ToSlash(filepath.Join(strings.TrimSpace(precheck.PackagePath), "devhub-signature.json"))
	} else {
		record.SignatureFilePath = sigURLPath
	}

	var sig domain.DevHubDetachedSignatureFile
	if err := json.Unmarshal(sigBytes, &sig); err != nil {
		return saveFailed(domain.PluginPackageSignatureStatusFailed, domain.NewPluginError("plugin_package_signature_invalid", "签名文件不是合法 JSON").
			WithStatus(400).
			WithSuggestion("请提供合法的 devhub-signature.json。"))
	}

	record.PublisherID = strings.TrimSpace(sig.SignaturePayload.PublisherID)
	if record.PublisherID == "" {
		record.PublisherID = strings.TrimSpace(sig.PublisherID)
	}
	record.KeyID = strings.TrimSpace(sig.SignaturePayload.KeyID)
	if record.KeyID == "" {
		record.KeyID = strings.TrimSpace(sig.KeyID)
	}
	record.Algorithm = strings.TrimSpace(sig.Algorithm)
	record.SignatureBase64 = strings.TrimSpace(sig.Signature)
	record.SignaturePayloadJSON = mustJSON(sig.SignaturePayload)

	trusted, found := s.repo.PluginTrustedPublisherByKey(record.PublisherID, record.KeyID)
	if !found {
		record.Status = domain.PluginPackageSignatureStatusUntrustedPublisher
		record.VerifiedAt = Now()
		record.UpdatedAt = Now()
		out, _ := s.repo.SavePluginPackageSignature(record)
		return out, nil
	}

	verifyRes, verr := pluginregistry.VerifyDetachedPluginPackageSignature(sig, manifest, record.PackageSHA256, record.ManifestSHA256, trusted)
	if verr != nil {
		status := verifyRes.Status
		if strings.TrimSpace(status) == "" {
			status = domain.PluginPackageSignatureStatusFailed
		}
		return saveFailed(status, verr)
	}
	record.Status = verifyRes.Status
	record.VerifiedAt = Now()
	record.WarningsJSON = mustJSON(verifyRes.Warnings)
	record.ErrorMessage = verifyRes.ErrorMessage
	record.UpdatedAt = Now()
	out, _ := s.repo.SavePluginPackageSignature(record)
	return out, nil
}

func (s *Service) ListPluginPackageSignatures(filter domain.PluginPackageSignatureFilter) (domain.PluginPackageSignatureListResponse, error) {
	items, total, err := s.repo.PluginPackageSignatures(filter)
	if err != nil {
		return domain.PluginPackageSignatureListResponse{}, err
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	summary := map[string]int{"total": total}
	return domain.PluginPackageSignatureListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		Summary: summary,
	}, nil
}

func (s *Service) GetPluginPackageSignature(id int64) (domain.PluginPackageSignatureRecord, error) {
	it, ok := s.repo.PluginPackageSignatureByID(id)
	if !ok || it.ID <= 0 {
		return domain.PluginPackageSignatureRecord{}, domain.NewPluginError("plugin_package_signature_not_found", "验签记录不存在").
			WithStatus(404).
			WithSuggestion("请刷新验签列表后重试。")
	}
	return it, nil
}

func (s *Service) DeletePluginPackageSignature(id int64) (domain.PluginPackageSignatureRecord, error) {
	it, err := s.GetPluginPackageSignature(id)
	if err != nil {
		return domain.PluginPackageSignatureRecord{}, err
	}
	if it.Status == domain.PluginPackageSignatureStatusDeleted {
		return it, nil
	}
	it.Status = domain.PluginPackageSignatureStatusDeleted
	it.UpdatedAt = Now()
	return s.repo.SavePluginPackageSignature(it)
}

func (s *Service) maybeDownloadDetachedSignature(download domain.PluginPackageDownloadRecord, signatureRecordID int64) ([]byte, string, error) {
	rawURL := strings.TrimSpace(download.SignatureURL)
	if rawURL == "" {
		return nil, "", nil
	}
	u, err := validatePluginPackageSignatureURL(rawURL)
	if err != nil {
		return nil, "", err
	}

	root, err := serviceProjectRoot()
	if err != nil {
		return nil, "", domain.NewPluginError("plugin_package_signature_failed", "读取项目根目录失败").WithStatus(500)
	}
	relRoot := pluginPackageSignatureStagingRoot()
	absRoot := filepath.Join(root, filepath.FromSlash(relRoot))
	if err := os.MkdirAll(absRoot, 0o755); err != nil {
		return nil, "", domain.NewPluginError("plugin_package_signature_failed", "创建签名 staging 目录失败").
			WithStatus(500).
			WithSuggestion("请检查 storage/plugins/staging 目录权限。")
	}

	tmpFile, err := os.CreateTemp(absRoot, "sig_*.part")
	if err != nil {
		return nil, "", domain.NewPluginError("plugin_package_signature_failed", "创建签名临时文件失败").
			WithStatus(500)
	}
	tmpPath := tmpFile.Name()
	cleanupTmp := true
	defer func() {
		_ = tmpFile.Close()
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	finalURL, contentType, size, rawBytes, err := downloadPluginSignatureURL(u.String(), tmpFile, pluginPackageSignatureMaxBytes())
	_ = finalURL
	_ = contentType
	if err != nil {
		return nil, "", err
	}
	if size <= 0 || len(rawBytes) == 0 {
		return nil, "", domain.NewPluginError("plugin_package_signature_failed", "下载签名文件为空").WithStatus(400)
	}

	// Move to stable path.
	fileName := "signature-" + strconv.FormatInt(download.ID, 10) + "-" + strconv.FormatInt(signatureRecordID, 10) + ".json"
	finalRel := filepath.ToSlash(filepath.Join(relRoot, fileName))
	finalAbs := filepath.Join(absRoot, fileName)
	if !pathInside(absRoot, finalAbs) {
		return nil, "", domain.NewPluginError("plugin_package_signature_failed", "签名 staging 路径非法").WithStatus(400)
	}
	if err := os.Rename(tmpPath, finalAbs); err != nil {
		return nil, "", domain.NewPluginError("plugin_package_signature_failed", "移动签名 staging 文件失败").WithStatus(500)
	}
	cleanupTmp = false
	return rawBytes, finalRel, nil
}

func validatePluginPackageSignatureURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil || u.Host == "" {
		return nil, domain.NewPluginError("plugin_package_signature_url_invalid", "签名文件 URL 不合法").
			WithStatus(400).
			WithSuggestion("请提供完整的 https signature_url。")
	}
	if strings.ToLower(u.Scheme) != "https" {
		return nil, domain.NewPluginError("plugin_package_signature_url_invalid", "仅允许 https signature_url").
			WithStatus(400).
			WithDetail("scheme", u.Scheme).
			WithSuggestion("请使用 https；file/ftp/http/gopher 等协议均被拒绝。")
	}
	if !strings.HasSuffix(strings.ToLower(u.Path), ".json") {
		return nil, domain.NewPluginError("plugin_package_signature_url_invalid", "签名文件必须为 .json").
			WithStatus(400).
			WithSuggestion("请提供 devhub-signature.json 的 URL。")
	}
	// Reuse the same host/IP restrictions as package download.
	if err := validatePluginPackageDownloadHost(u.Hostname()); err != nil {
		return nil, err
	}
	return u, nil
}

func downloadPluginSignatureURL(raw string, out io.Writer, maxBytes int64) (string, string, int64, []byte, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&pluginPackageDownloadDialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     pluginPackageDownloadTLSConfig(),
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return domain.NewPluginError("plugin_package_signature_redirect_blocked", "签名文件重定向次数超过限制").
					WithStatus(400)
			}
			_, err := validatePluginPackageSignatureURL(req.URL.String())
			return err
		},
	}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return "", "", 0, nil, domain.NewPluginError("plugin_package_signature_url_invalid", "签名文件 URL 不合法").WithStatus(400)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", 0, nil, normalizePluginPackageDownloadError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.Request.URL.String(), resp.Header.Get("Content-Type"), 0, nil, domain.NewPluginError("plugin_package_signature_failed", "签名文件下载失败").
			WithStatus(400).
			WithDetail("status_code", resp.StatusCode)
	}
	if resp.ContentLength > maxBytes {
		return resp.Request.URL.String(), resp.Header.Get("Content-Type"), 0, nil, domain.NewPluginError("plugin_package_signature_response_too_large", "签名文件过大").
			WithStatus(400).
			WithDetail("max_bytes", maxBytes).
			WithSuggestion("签名文件默认限制 64KB。")
	}

	buf := &strings.Builder{}
	limited := &io.LimitedReader{R: resp.Body, N: maxBytes + 1}
	written, err := io.Copy(io.MultiWriter(out, buf), limited)
	if err != nil {
		return resp.Request.URL.String(), resp.Header.Get("Content-Type"), written, nil, domain.NewPluginError("plugin_package_signature_failed", "读取签名文件失败").
			WithStatus(400)
	}
	if written > maxBytes {
		return resp.Request.URL.String(), resp.Header.Get("Content-Type"), written, nil, domain.NewPluginError("plugin_package_signature_response_too_large", "签名文件过大").
			WithStatus(400).
			WithDetail("max_bytes", maxBytes)
	}
	rawBytes := []byte(buf.String())
	return resp.Request.URL.String(), resp.Header.Get("Content-Type"), written, rawBytes, nil
}

func jsonEqualBytes(a, b []byte) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	var ja any
	var jb any
	if err := json.Unmarshal(a, &ja); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &jb); err != nil {
		return false
	}
	return mustJSON(ja) == mustJSON(jb)
}
