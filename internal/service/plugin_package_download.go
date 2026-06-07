package service

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
)

const (
	pluginPackageDownloadStagingRootDefault = "storage/plugins/staging/downloads"
	pluginPackageDownloadDefaultMaxBytes    = int64(20 * 1024 * 1024)
)

var pluginPackageDownloadCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

type PluginPackageDownloadOperator struct {
	ID   int64
	Name string
}

func pluginPackageDownloadStagingRoot() string {
	if v := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_STAGING_ROOT")); v != "" {
		return filepath.ToSlash(v)
	}
	return pluginPackageDownloadStagingRootDefault
}

func pluginPackageDownloadMaxBytes() int64 {
	if v := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_MAX_BYTES")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return pluginPackageDownloadDefaultMaxBytes
}

func (s *Service) DownloadPluginPackageToStagingAs(operator PluginPackageDownloadOperator, req domain.PluginPackageDownloadRequest) (domain.PluginPackageDownloadRecord, error) {
	req.PluginCode = strings.TrimSpace(req.PluginCode)
	req.Version = strings.TrimSpace(req.Version)
	req.PackageURL = strings.TrimSpace(req.PackageURL)
	req.SHA256 = strings.ToLower(strings.TrimSpace(req.SHA256))
	req.SignatureURL = strings.TrimSpace(req.SignatureURL)
	now := Now()
	record := domain.PluginPackageDownloadRecord{
		PluginCode:     req.PluginCode,
		Version:        req.Version,
		SourceURL:      req.PackageURL,
		SignatureURL:   req.SignatureURL,
		Status:         domain.PluginPackageDownloadStatusPending,
		SHA256Expected: req.SHA256,
		CreatedBy:      operator.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	record, _ = s.repo.AppendPluginPackageDownload(record)

	failRecord := func(status string, apiErr error) (domain.PluginPackageDownloadRecord, error) {
		record.Status = status
		record.UpdatedAt = Now()
		record.ErrorMessage = apiErr.Error()
		var pe *domain.APIError
		if errors.As(apiErr, &pe) {
			record.ErrorCode = pe.Code
		} else {
			record.ErrorCode = "plugin_package_download_failed"
		}
		record, _ = s.repo.SavePluginPackageDownload(record)
		return record, apiErr
	}

	if !pluginPackageDownloadCodePattern.MatchString(req.PluginCode) {
		return failRecord(domain.PluginPackageDownloadStatusRejected, domain.NewPluginError("plugin_package_download_invalid_request", "插件 code 不合法").
			WithStatus(400).
			WithSuggestion("plugin_code 必须使用小写字母、数字和下划线，并以字母开头。"))
	}
	if req.Version == "" {
		return failRecord(domain.PluginPackageDownloadStatusRejected, domain.NewPluginError("plugin_package_download_invalid_request", "插件版本不能为空").
			WithStatus(400).
			WithSuggestion("请提供远程插件包版本号。"))
	}
	u, ext, err := validatePluginPackageDownloadURL(req.PackageURL)
	if err != nil {
		return failRecord(domain.PluginPackageDownloadStatusRejected, err)
	}
	if req.SHA256 != "" && !isSHA256Hex(req.SHA256) {
		return failRecord(domain.PluginPackageDownloadStatusRejected, domain.NewPluginError("plugin_package_download_checksum_invalid", "sha256 格式不合法").
			WithStatus(400).
			WithSuggestion("sha256 必须是 64 位十六进制摘要。"))
	}

	root, err := serviceProjectRoot()
	if err != nil {
		return failRecord(domain.PluginPackageDownloadStatusFailed, domain.NewPluginError("plugin_package_download_failed", "读取项目根目录失败").
			WithStatus(500).
			WithSuggestion("请确认服务运行目录包含 DevHub 项目文件。"))
	}
	stagingRelRoot := pluginPackageDownloadStagingRoot()
	stagingAbsRoot := filepath.Join(root, filepath.FromSlash(stagingRelRoot))
	if err := os.MkdirAll(stagingAbsRoot, 0o755); err != nil {
		return failRecord(domain.PluginPackageDownloadStatusFailed, domain.NewPluginError("plugin_package_download_failed", "创建 staging 目录失败").
			WithStatus(500).
			WithSuggestion("请检查 storage/plugins/staging 目录权限。"))
	}

	record.Status = domain.PluginPackageDownloadStatusDownloading
	record, _ = s.repo.SavePluginPackageDownload(record)

	tmpFile, err := os.CreateTemp(stagingAbsRoot, "remote_download_*.part")
	if err != nil {
		return failRecord(domain.PluginPackageDownloadStatusFailed, domain.NewPluginError("plugin_package_download_failed", "创建临时下载文件失败").
			WithStatus(500).
			WithSuggestion("请检查 staging 目录写入权限。"))
	}
	tmpPath := tmpFile.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	finalURL, contentType, size, actualSHA, err := downloadPluginPackageURL(u.String(), tmpFile, pluginPackageDownloadMaxBytes())
	_ = tmpFile.Close()
	record.FinalURL = finalURL
	record.ContentType = contentType
	record.FileSize = size
	record.SHA256Actual = actualSHA
	if err != nil {
		return failRecord(domain.PluginPackageDownloadStatusFailed, err)
	}
	if req.SHA256 != "" && !strings.EqualFold(req.SHA256, actualSHA) {
		return failRecord(domain.PluginPackageDownloadStatusChecksumFailed, domain.NewPluginError("plugin_package_download_checksum_mismatch", "远程插件包 sha256 校验失败").
			WithStatus(400).
			WithDetail("sha256_expected", req.SHA256).
			WithDetail("sha256_actual", actualSHA).
			WithSuggestion("请确认远程索引中的 package_sha256 与实际下载文件一致后重试。"))
	}

	fileName := buildPluginPackageDownloadFileName(req.PluginCode, req.Version, actualSHA, ext)
	finalRel := filepath.ToSlash(filepath.Join(stagingRelRoot, fileName))
	finalAbs := filepath.Join(stagingAbsRoot, fileName)
	if !pathInside(stagingAbsRoot, finalAbs) {
		return failRecord(domain.PluginPackageDownloadStatusRejected, domain.NewPluginError("plugin_package_download_path_invalid", "staging 文件路径非法").
			WithStatus(400).
			WithSuggestion("请检查 plugin_code 和 version 是否包含非法字符。"))
	}
	if err := os.Rename(tmpPath, finalAbs); err != nil {
		return failRecord(domain.PluginPackageDownloadStatusFailed, domain.NewPluginError("plugin_package_download_failed", "移动 staging 文件失败").
			WithStatus(500).
			WithSuggestion("请检查 staging 目录权限和磁盘空间。"))
	}
	cleanupTmp = false

	record.FileName = fileName
	record.StagingPath = finalRel
	record.DownloadedAt = Now()
	if req.SHA256 == "" {
		record.Status = domain.PluginPackageDownloadStatusChecksumMissing
	} else {
		record.Status = domain.PluginPackageDownloadStatusDownloaded
	}
	record.ErrorCode = ""
	record.ErrorMessage = ""
	record.UpdatedAt = Now()
	return s.repo.SavePluginPackageDownload(record)
}

func (s *Service) ListPluginPackageDownloads(filter domain.PluginPackageDownloadFilter) (domain.PluginPackageDownloadListResponse, error) {
	items, total, err := s.repo.PluginPackageDownloads(filter)
	if err != nil {
		return domain.PluginPackageDownloadListResponse{}, err
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
	for _, status := range []string{
		domain.PluginPackageDownloadStatusPending,
		domain.PluginPackageDownloadStatusDownloading,
		domain.PluginPackageDownloadStatusDownloaded,
		domain.PluginPackageDownloadStatusChecksumFailed,
		domain.PluginPackageDownloadStatusRejected,
		domain.PluginPackageDownloadStatusFailed,
		domain.PluginPackageDownloadStatusChecksumMissing,
		domain.PluginPackageDownloadStatusDeleted,
	} {
		summary[status] = 0
	}
	all, _, _ := s.repo.PluginPackageDownloads(domain.PluginPackageDownloadFilter{Page: 1, PageSize: 100000})
	for _, it := range all {
		summary[it.Status]++
	}
	return domain.PluginPackageDownloadListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		Summary: summary,
	}, nil
}

func (s *Service) GetPluginPackageDownload(id int64) (domain.PluginPackageDownloadRecord, error) {
	record, ok := s.repo.PluginPackageDownloadByID(id)
	if !ok {
		return domain.PluginPackageDownloadRecord{}, domain.NewPluginError("plugin_package_staging_not_found", "staging 插件包不存在").
			WithStatus(404).
			WithSuggestion("请刷新 staging 列表后重试。")
	}
	return record, nil
}

func (s *Service) DeletePluginPackageDownload(id int64) (domain.PluginPackageDownloadRecord, error) {
	record, ok := s.repo.PluginPackageDownloadByID(id)
	if !ok {
		return domain.PluginPackageDownloadRecord{}, domain.NewPluginError("plugin_package_staging_not_found", "staging 插件包不存在").
			WithStatus(404).
			WithSuggestion("请刷新 staging 列表后重试。")
	}
	if record.Status == domain.PluginPackageDownloadStatusDeleted {
		return record, nil
	}
	if record.StagingPath != "" {
		root, err := serviceProjectRoot()
		if err != nil {
			return domain.PluginPackageDownloadRecord{}, domain.NewPluginError("plugin_package_staging_delete_failed", "读取项目根目录失败").WithStatus(500)
		}
		stagingAbsRoot := filepath.Join(root, filepath.FromSlash(pluginPackageDownloadStagingRoot()))
		targetAbs := filepath.Join(root, filepath.FromSlash(record.StagingPath))
		if pathInside(stagingAbsRoot, targetAbs) {
			_ = os.Remove(targetAbs)
		}
	}
	record.Status = domain.PluginPackageDownloadStatusDeleted
	record.DeletedAt = Now()
	record.UpdatedAt = Now()
	return s.repo.SavePluginPackageDownload(record)
}

func validatePluginPackageDownloadURL(raw string) (*url.URL, string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u == nil || u.Host == "" {
		return nil, "", domain.NewPluginError("plugin_package_download_url_invalid", "插件包 URL 不合法").
			WithStatus(400).
			WithSuggestion("请提供完整的 https 插件包 URL。")
	}
	if strings.ToLower(u.Scheme) != "https" {
		return nil, "", domain.NewPluginError("plugin_package_download_url_invalid", "仅允许 https 插件包 URL").
			WithStatus(400).
			WithDetail("scheme", u.Scheme).
			WithSuggestion("请使用 https；file/ftp/http/gopher 等协议均被拒绝。")
	}
	ext := packageDownloadExtension(u.Path)
	if ext == "" {
		return nil, "", domain.NewPluginError("plugin_package_download_type_unsupported", "插件包格式不受支持").
			WithStatus(400).
			WithSuggestion("当前仅允许 .zip、.tar.gz、.tgz 插件包。")
	}
	if err := validatePluginPackageDownloadHost(u.Hostname()); err != nil {
		return nil, "", err
	}
	return u, ext, nil
}

func downloadPluginPackageURL(raw string, out io.Writer, maxBytes int64) (string, string, int64, string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
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
				return domain.NewPluginError("plugin_package_download_redirect_blocked", "远程插件包重定向次数超过限制").
					WithStatus(400).
					WithSuggestion("请确认下载地址不会发生超过 3 次跳转。")
			}
			_, _, err := validatePluginPackageDownloadURL(req.URL.String())
			return err
		},
	}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return "", "", 0, "", domain.NewPluginError("plugin_package_download_url_invalid", "插件包 URL 不合法").WithStatus(400)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", 0, "", normalizePluginPackageDownloadError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.Request.URL.String(), resp.Header.Get("Content-Type"), 0, "", domain.NewPluginError("plugin_package_download_failed", "远程插件包下载失败").
			WithStatus(400).
			WithDetail("status_code", resp.StatusCode).
			WithSuggestion("请确认远程插件包地址可访问。")
	}
	if resp.ContentLength > maxBytes {
		return resp.Request.URL.String(), resp.Header.Get("Content-Type"), 0, "", pluginPackageDownloadTooLarge(maxBytes)
	}
	hasher := sha256.New()
	limited := &io.LimitedReader{R: resp.Body, N: maxBytes + 1}
	written, err := io.Copy(io.MultiWriter(out, hasher), limited)
	if err != nil {
		return resp.Request.URL.String(), resp.Header.Get("Content-Type"), written, "", domain.NewPluginError("plugin_package_download_failed", "读取远程插件包失败").
			WithStatus(400).
			WithSuggestion("请检查网络连接并重试。")
	}
	if written > maxBytes {
		return resp.Request.URL.String(), resp.Header.Get("Content-Type"), written, "", pluginPackageDownloadTooLarge(maxBytes)
	}
	return resp.Request.URL.String(), resp.Header.Get("Content-Type"), written, hex.EncodeToString(hasher.Sum(nil)), nil
}

type pluginPackageDownloadDialer struct {
	Timeout   time.Duration
	KeepAlive time.Duration
}

func (d *pluginPackageDownloadDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	if err := validatePluginPackageDownloadHost(host); err != nil {
		return nil, err
	}
	resolver := net.DefaultResolver
	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{Timeout: d.Timeout, KeepAlive: d.KeepAlive}
	var lastErr error
	for _, ip := range ips {
		if !allowLocalPluginPackageDownload() && blockedPluginPackageDownloadIP(ip.IP) {
			lastErr = errors.New("blocked private ip")
			continue
		}
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no public ip resolved")
	}
	return nil, domain.NewPluginError("plugin_package_download_url_forbidden", "插件包下载地址解析到禁止访问的地址").
		WithStatus(400).
		WithSuggestion("请使用公网 https 下载地址，禁止 localhost、内网和链路本地地址。")
}

func validatePluginPackageDownloadHost(host string) error {
	host = strings.Trim(strings.ToLower(strings.TrimSpace(host)), "[]")
	if host == "" {
		return domain.NewPluginError("plugin_package_download_url_forbidden", "插件包下载地址缺少主机名").
			WithStatus(400).
			WithSuggestion("请使用公网 https 下载地址。")
	}
	if allowLocalPluginPackageDownload() {
		return nil
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return domain.NewPluginError("plugin_package_download_url_forbidden", "插件包下载地址禁止使用 localhost").
			WithStatus(400).
			WithSuggestion("请使用公网 https 下载地址。")
	}
	if ip := net.ParseIP(host); ip != nil {
		if blockedPluginPackageDownloadIP(ip) {
			return domain.NewPluginError("plugin_package_download_url_forbidden", "插件包下载地址指向内网或本机地址").
				WithStatus(400).
				WithSuggestion("请使用公网 https 下载地址。")
		}
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return domain.NewPluginError("plugin_package_download_url_invalid", "插件包下载地址域名无法解析").
			WithStatus(400).
			WithSuggestion("请确认域名可以解析到公网地址。")
	}
	for _, ip := range ips {
		if blockedPluginPackageDownloadIP(ip.IP) {
			return domain.NewPluginError("plugin_package_download_url_forbidden", "插件包下载地址解析到内网或本机地址").
				WithStatus(400).
				WithSuggestion("请使用公网 https 下载地址；重定向和 DNS 解析结果都会被校验。")
		}
	}
	return nil
}

func blockedPluginPackageDownloadIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return true
	}
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsUnspecified() || addr.IsMulticast() {
		return true
	}
	if addr.Is6() && (strings.HasPrefix(addr.String(), "fc") || strings.HasPrefix(addr.String(), "fd") || strings.HasPrefix(addr.String(), "fe80:")) {
		return true
	}
	return false
}

func packageDownloadExtension(path string) string {
	lower := strings.ToLower(strings.TrimSpace(path))
	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return ".tar.gz"
	case strings.HasSuffix(lower, ".tgz"):
		return ".tgz"
	case strings.HasSuffix(lower, ".zip"):
		return ".zip"
	default:
		return ""
	}
}

func buildPluginPackageDownloadFileName(code, version, sha, ext string) string {
	code = safeDownloadName(code)
	version = safeDownloadName(version)
	prefix := sha
	if len(prefix) > 12 {
		prefix = prefix[:12]
	}
	return fmt.Sprintf("%s-%s-%s-%d%s", code, version, prefix, timeNow().Unix(), ext)
}

func safeDownloadName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "._-")
	if out == "" {
		return "plugin"
	}
	return out
}

func pathInside(root, target string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	return err == nil && rel != "." && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel)
}

func isSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func pluginPackageDownloadTooLarge(maxBytes int64) error {
	return domain.NewPluginError("plugin_package_download_too_large", "远程插件包超过大小限制").
		WithStatus(400).
		WithDetail("max_bytes", maxBytes).
		WithSuggestion("请将远程插件包控制在 20MB 以内，或调整受控配置后重试。")
}

func normalizePluginPackageDownloadError(err error) error {
	var pluginErr *domain.APIError
	if errors.As(err, &pluginErr) {
		return err
	}
	return domain.NewPluginError("plugin_package_download_failed", "远程插件包下载失败").
		WithStatus(400).
		WithSuggestion("请确认下载 URL 可访问、证书有效且未跳转到被禁止地址。")
}

func pluginPackageDownloadTLSConfig() *tls.Config {
	if strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_INSECURE_TLS")) == "1" {
		return &tls.Config{InsecureSkipVerify: true} //nolint:gosec // 测试环境自签 HTTPS fixture 使用，默认关闭。
	}
	return nil
}

func allowLocalPluginPackageDownload() bool {
	return strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_ALLOW_LOCAL")) == "1"
}
