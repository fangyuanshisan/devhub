package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const (
	remoteIndexMaxBodyBytes = 2 * 1024 * 1024
	remoteIndexTimeout      = 10 * time.Second
)

type RemoteIndexOperator struct {
	ID   int64
	Name string
}

func (s *Service) ListPluginRemoteIndexes(filter domain.PluginRemoteIndexFilter) (domain.PluginRemoteIndexListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	items, total, err := s.repo.PluginRemoteIndexes(filter)
	if err != nil {
		return domain.PluginRemoteIndexListResponse{}, err
	}
	summary := domain.PluginRemoteIndexSummary{Total: total}
	all, _, _ := s.repo.PluginRemoteIndexes(domain.PluginRemoteIndexFilter{Page: 1, PageSize: 1000})
	for _, it := range all {
		switch strings.TrimSpace(it.Status) {
		case domain.PluginRemoteIndexStatusEnabled:
			summary.Enabled++
		case domain.PluginRemoteIndexStatusDisabled:
			summary.Disabled++
		}
		if it.LastFetchStatus == "failed" {
			summary.Failed++
		}
	}
	return domain.PluginRemoteIndexListResponse{Items: items, Pagination: domain.Pagination{Page: filter.Page, PageSize: filter.PageSize, Total: total}, Summary: summary}, nil
}

func (s *Service) CreatePluginRemoteIndex(operator RemoteIndexOperator, req domain.PluginRemoteIndexSource) (domain.PluginRemoteIndexSource, error) {
	record, warnings, err := normalizeRemoteIndexSource(req)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	if _, ok := s.repo.PluginRemoteIndexBySourceID(record.SourceID); ok {
		return domain.PluginRemoteIndexSource{}, domain.NewPluginError("plugin_remote_index_schema_invalid", "远程索引 source_id 已存在").
			WithStatus(409).WithDetail("source_id", record.SourceID).WithSuggestion("请修改 source_id，或编辑已有远程索引源。")
	}
	record.CreatedBy = operator.ID
	record.UpdatedBy = operator.ID
	record.CreatedAt = Now()
	record.UpdatedAt = Now()
	if len(warnings) > 0 {
		record.LastErrorMessage = strings.Join(warnings, "；")
	}
	out, err := s.repo.AppendPluginRemoteIndex(record)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	s.logRemoteIndex(operator, "plugin.remote_index.created", out, "")
	return out, nil
}

func (s *Service) UpdatePluginRemoteIndex(operator RemoteIndexOperator, id int64, req domain.PluginRemoteIndexSource) (domain.PluginRemoteIndexSource, error) {
	existing, err := s.GetPluginRemoteIndex(id)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	record, warnings, err := normalizeRemoteIndexSource(req)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	if other, ok := s.repo.PluginRemoteIndexBySourceID(record.SourceID); ok && other.ID != id {
		return domain.PluginRemoteIndexSource{}, domain.NewPluginError("plugin_remote_index_schema_invalid", "远程索引 source_id 已存在").
			WithStatus(409).WithDetail("source_id", record.SourceID)
	}
	record.ID = id
	record.CreatedBy = existing.CreatedBy
	record.CreatedAt = existing.CreatedAt
	record.UpdatedBy = operator.ID
	record.UpdatedAt = Now()
	record.LastFetchStatus = existing.LastFetchStatus
	record.LastFetchAt = existing.LastFetchAt
	record.LastErrorCode = existing.LastErrorCode
	record.LastErrorMessage = existing.LastErrorMessage
	record.LastIndexHash = existing.LastIndexHash
	record.MetadataJSON = existing.MetadataJSON
	if len(warnings) > 0 && record.LastErrorMessage == "" {
		record.LastErrorMessage = strings.Join(warnings, "；")
	}
	out, err := s.repo.SavePluginRemoteIndex(record)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	s.logRemoteIndex(operator, "plugin.remote_index.updated", out, "")
	return out, nil
}

func (s *Service) GetPluginRemoteIndex(id int64) (domain.PluginRemoteIndexSource, error) {
	record, ok := s.repo.PluginRemoteIndexByID(id)
	if !ok {
		return domain.PluginRemoteIndexSource{}, domain.NewPluginError("plugin_remote_index_not_found", "远程索引源不存在").WithStatus(404).WithDetail("id", id)
	}
	return record, nil
}

func (s *Service) SetPluginRemoteIndexStatus(operator RemoteIndexOperator, id int64, status string) (domain.PluginRemoteIndexSource, error) {
	record, err := s.GetPluginRemoteIndex(id)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != domain.PluginRemoteIndexStatusEnabled && status != domain.PluginRemoteIndexStatusDisabled {
		return domain.PluginRemoteIndexSource{}, domain.NewPluginError("plugin_remote_index_schema_invalid", "远程索引状态不合法").
			WithStatus(400).WithDetail("status", status).WithSuggestion("状态仅支持 enabled / disabled。")
	}
	record.Status = status
	record.UpdatedBy = operator.ID
	record.UpdatedAt = Now()
	out, err := s.repo.SavePluginRemoteIndex(record)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, err
	}
	s.logRemoteIndex(operator, "plugin.remote_index."+status, out, "")
	return out, nil
}

func (s *Service) DeletePluginRemoteIndex(operator RemoteIndexOperator, id int64) error {
	record, err := s.GetPluginRemoteIndex(id)
	if err != nil {
		return err
	}
	if err := s.repo.DeletePluginRemoteIndex(id); err != nil {
		return err
	}
	s.logRemoteIndex(operator, "plugin.remote_index.deleted", record, "")
	return nil
}

func (s *Service) FetchPluginRemoteIndex(operator RemoteIndexOperator, id int64) (domain.PluginRemoteIndexFetchResponse, error) {
	record, err := s.GetPluginRemoteIndex(id)
	if err != nil {
		return domain.PluginRemoteIndexFetchResponse{}, err
	}
	if record.Status == domain.PluginRemoteIndexStatusDisabled {
		return domain.PluginRemoteIndexFetchResponse{}, domain.NewPluginError("plugin_remote_index_disabled", "远程索引源已禁用").
			WithStatus(409).WithDetail("source_id", record.SourceID).WithSuggestion("请先启用索引源，再执行拉取。")
	}
	raw, contentType, err := fetchRemoteIndexJSON(record.IndexURL)
	if err != nil {
		record.LastFetchStatus = "failed"
		record.LastFetchAt = Now()
		record.LastErrorCode = pluginErrorCode(err, "plugin_remote_index_fetch_failed")
		record.LastErrorMessage = err.Error()
		record.UpdatedBy = operator.ID
		_, _ = s.repo.SavePluginRemoteIndex(record)
		s.logRemoteIndex(operator, "plugin.remote_index.fetch_failed", record, record.LastErrorCode)
		return domain.PluginRemoteIndexFetchResponse{}, err
	}
	var doc domain.PluginRemoteIndexDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		apiErr := domain.NewPluginError("plugin_remote_index_invalid_json", "远程索引 JSON 无法解析").
			WithStatus(400).WithSuggestion("请检查 index.json 是否是合法 JSON。")
		record.LastFetchStatus = "failed"
		record.LastFetchAt = Now()
		record.LastErrorCode = apiErr.Code
		record.LastErrorMessage = apiErr.Message
		record.UpdatedBy = operator.ID
		_, _ = s.repo.SavePluginRemoteIndex(record)
		s.logRemoteIndex(operator, "plugin.remote_index.fetch_failed", record, apiErr.Code)
		return domain.PluginRemoteIndexFetchResponse{}, apiErr
	}
	validation := s.validateRemoteIndexDocument(doc)
	if contentType != "" && !strings.Contains(strings.ToLower(contentType), "json") {
		validation.Warnings = append(validation.Warnings, remoteIndexRisk("plugin_remote_index_schema_invalid", "warning", "远程索引 Content-Type 不是 JSON", "建议远程源返回 application/json。"))
	}
	indexHash := "sha256:" + sha256Hex(raw)
	metadata, _ := json.Marshal(map[string]any{"document": doc})
	record.LastFetchStatus = "ok"
	if !validation.Valid {
		record.LastFetchStatus = "invalid"
	}
	record.LastFetchAt = Now()
	record.LastErrorCode = ""
	record.LastErrorMessage = ""
	if !validation.Valid && len(validation.Errors) > 0 {
		record.LastErrorCode = validation.Errors[0].Code
		record.LastErrorMessage = validation.Errors[0].Message
	}
	record.LastIndexHash = indexHash
	record.MetadataJSON = string(metadata)
	record.UpdatedBy = operator.ID
	record, _ = s.repo.SavePluginRemoteIndex(record)
	s.logRemoteIndex(operator, "plugin.remote_index.fetched", record, "")
	return domain.PluginRemoteIndexFetchResponse{
		Source:     record,
		Document:   doc,
		Validation: validation,
		IndexHash:  indexHash,
		Warnings:   validation.Warnings,
		Errors:     validation.Errors,
	}, nil
}

func (s *Service) ListRemoteIndexPlugins(id int64, keyword string, page, pageSize int) (domain.PluginRemotePluginListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	record, doc, err := s.remoteIndexDocument(id)
	if err != nil {
		return domain.PluginRemotePluginListResponse{}, err
	}
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	all := make([]domain.PluginRemotePluginListItem, 0, len(doc.Plugins))
	for _, plugin := range doc.Plugins {
		if keyword != "" {
			hay := strings.ToLower(strings.Join([]string{plugin.Code, plugin.Name, plugin.Description, plugin.LatestVersion}, " "))
			if !strings.Contains(hay, keyword) {
				continue
			}
		}
		all = append(all, s.enrichRemotePlugin(plugin))
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Code < all[j].Code })
	total := len(all)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	summary := map[string]int{"total": total, "trusted": 0, "unknown": 0, "blocked": 0, "installed": 0, "update_available": 0, "incompatible": 0}
	for _, it := range all {
		if it.Installed {
			summary["installed"]++
		}
		if it.VersionStatus == "update_available" {
			summary["update_available"]++
		}
		switch it.PublisherTrustStatus {
		case "trusted":
			summary["trusted"]++
		case "blocked", "revoked":
			summary["blocked"]++
		default:
			summary["unknown"]++
		}
		if it.CoreCompatibility.Status == pluginregistry.CompatibilityIncompatible {
			summary["incompatible"]++
		}
	}
	_ = record
	return domain.PluginRemotePluginListResponse{Items: append([]domain.PluginRemotePluginListItem(nil), all[start:end]...), Pagination: domain.Pagination{Page: page, PageSize: pageSize, Total: total}, Summary: summary}, nil
}

func (s *Service) GetRemoteIndexPlugin(id int64, code string) (domain.PluginRemotePluginDetailResponse, error) {
	record, doc, err := s.remoteIndexDocument(id)
	if err != nil {
		return domain.PluginRemotePluginDetailResponse{}, err
	}
	for _, plugin := range doc.Plugins {
		if plugin.Code != code {
			continue
		}
		details := make([]domain.PluginRemoteVersionDetail, 0, len(plugin.Versions))
		for _, ver := range plugin.Versions {
			compat := pluginregistry.CheckPluginVersionCompatibility(domain.PluginManifest{MinCoreVersion: ver.MinCoreVersion, CompatibleCoreVersion: ver.CompatibleCoreVersion}, currentCoreVersion())
			riskLevel, risks := s.remoteVersionRisks(ver, compat)
			details = append(details, domain.PluginRemoteVersionDetail{
				PluginRemoteIndexVersionDoc: ver,
				CoreCompatibility:           compat,
				PublisherTrustStatus:        s.remotePublisherTrust(ver.PublisherID, ver.PublicKeyID),
				RiskLevel:                   riskLevel,
				RiskItems:                   risks,
			})
		}
		installed, installedOK := s.repo.PluginByCode(plugin.Code)
		return domain.PluginRemotePluginDetailResponse{
			Source:          record,
			Plugin:          plugin,
			Versions:        details,
			Installed:       installedOK,
			LocalVersion:    installed.Version,
			InstalledStatus: installed.Status,
			Readonly:        true,
			ReadonlyMessage: "当前仅展示远程索引元数据；不会下载、安装、执行代码或动态加载前端资产。",
		}, nil
	}
	return domain.PluginRemotePluginDetailResponse{}, domain.NewPluginError("plugin_remote_index_plugin_not_found", "远程索引中未找到插件").WithStatus(404).WithDetail("code", code)
}

func normalizeRemoteIndexSource(req domain.PluginRemoteIndexSource) (domain.PluginRemoteIndexSource, []string, error) {
	req.SourceID = strings.TrimSpace(req.SourceID)
	req.Name = strings.TrimSpace(req.Name)
	req.IndexURL = strings.TrimSpace(req.IndexURL)
	req.Homepage = strings.TrimSpace(req.Homepage)
	req.Description = strings.TrimSpace(req.Description)
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	req.TrustPolicy = strings.ToLower(strings.TrimSpace(req.TrustPolicy))
	if req.Status == "" {
		req.Status = domain.PluginRemoteIndexStatusEnabled
	}
	if req.TrustPolicy == "" {
		req.TrustPolicy = "readonly"
	}
	if req.SourceID == "" || req.Name == "" || req.IndexURL == "" {
		return domain.PluginRemoteIndexSource{}, nil, domain.NewPluginError("plugin_remote_index_schema_invalid", "source_id/name/index_url 必填").WithStatus(400)
	}
	if req.Status != domain.PluginRemoteIndexStatusEnabled && req.Status != domain.PluginRemoteIndexStatusDisabled {
		return domain.PluginRemoteIndexSource{}, nil, domain.NewPluginError("plugin_remote_index_schema_invalid", "远程索引状态不合法").WithStatus(400).WithDetail("status", req.Status)
	}
	if req.TrustPolicy != "readonly" && req.TrustPolicy != "trusted_metadata" && req.TrustPolicy != "blocked" {
		return domain.PluginRemoteIndexSource{}, nil, domain.NewPluginError("plugin_remote_index_schema_invalid", "远程索引 trust_policy 不合法").WithStatus(400).WithDetail("trust_policy", req.TrustPolicy)
	}
	normalized, warnings, err := validateRemoteIndexURL(req.IndexURL)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, nil, err
	}
	req.IndexURL = normalized
	return req, warnings, nil
}

func validateRemoteIndexURL(raw string) (string, []string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", nil, domain.NewPluginError("plugin_remote_index_invalid_url", "远程索引 URL 不合法").WithStatus(400).WithDetail("index_url", raw).WithSuggestion("请填写 http 或 https 的 index.json 地址。")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", nil, domain.NewPluginError("plugin_remote_index_invalid_url", "远程索引 URL 协议不支持").WithStatus(400).WithDetail("scheme", parsed.Scheme).WithSuggestion("仅支持 http / https；禁止 file:// 等本地协议。")
	}
	host := parsed.Hostname()
	if isForbiddenRemoteHost(host) {
		return "", nil, domain.NewPluginError("plugin_remote_index_forbidden_url", "远程索引 URL 指向本机或内网地址，已拦截").
			WithStatus(400).WithDetail("host", host).WithSuggestion("请使用公网可访问的只读 HTTPS 索引地址；测试环境可设置 DEVHUB_ALLOW_LOCAL_REMOTE_INDEX=1。")
	}
	warnings := []string{}
	if scheme == "http" {
		warnings = append(warnings, "index_url 使用 http，生产环境建议使用 https")
	}
	return parsed.String(), warnings, nil
}

func isForbiddenRemoteHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("DEVHUB_ALLOW_LOCAL_REMOTE_INDEX")), "1") {
		return false
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

func fetchRemoteIndexJSON(indexURL string) ([]byte, string, error) {
	normalized, _, err := validateRemoteIndexURL(indexURL)
	if err != nil {
		return nil, "", err
	}
	client := &http.Client{
		Timeout: remoteIndexTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return domain.NewPluginError("plugin_remote_index_fetch_failed", "远程索引重定向次数过多").WithStatus(400)
			}
			_, _, err := validateRemoteIndexURL(req.URL.String())
			return err
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), remoteIndexTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalized, nil)
	if err != nil {
		return nil, "", domain.NewPluginError("plugin_remote_index_invalid_url", "远程索引 URL 不合法").WithStatus(400)
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", domain.NewPluginError("plugin_remote_index_fetch_timeout", "远程索引拉取超时").WithStatus(504).WithSuggestion("请检查远程索引服务可用性或稍后重试。")
		}
		return nil, "", domain.NewPluginError("plugin_remote_index_fetch_failed", "远程索引拉取失败").WithStatus(502).WithDetail("error", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", domain.NewPluginError("plugin_remote_index_fetch_failed", "远程索引返回非 2xx 状态").WithStatus(502).WithDetail("status", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, remoteIndexMaxBodyBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", domain.NewPluginError("plugin_remote_index_fetch_failed", "读取远程索引失败").WithStatus(502)
	}
	if int64(len(raw)) > remoteIndexMaxBodyBytes {
		return nil, "", domain.NewPluginError("plugin_remote_index_response_too_large", "远程索引响应超过大小限制").WithStatus(413).WithDetail("limit_bytes", remoteIndexMaxBodyBytes)
	}
	return raw, resp.Header.Get("Content-Type"), nil
}

func (s *Service) validateRemoteIndexDocument(doc domain.PluginRemoteIndexDocument) domain.PluginRemoteIndexValidation {
	out := domain.PluginRemoteIndexValidation{Valid: true}
	addErr := func(code, msg, suggestion string) {
		out.Valid = false
		out.Errors = append(out.Errors, remoteIndexRisk(code, "blocked", msg, suggestion))
	}
	addWarn := func(code, msg, suggestion string) {
		out.Warnings = append(out.Warnings, remoteIndexRisk(code, "warning", msg, suggestion))
	}
	if strings.TrimSpace(doc.SchemaVersion) == "" {
		addErr("plugin_remote_index_schema_invalid", "schema_version 缺失", "请按 DevHub 远程索引 JSON 规范补齐 schema_version。")
	}
	if strings.TrimSpace(doc.GeneratedAt) == "" {
		addWarn("plugin_remote_index_schema_invalid", "generated_at 缺失", "建议记录索引生成时间。")
	}
	if strings.TrimSpace(doc.Source.SourceID) == "" || strings.TrimSpace(doc.Source.Name) == "" {
		addErr("plugin_remote_index_schema_invalid", "source.source_id/source.name 缺失", "请补齐远程索引来源信息。")
	}
	if len(doc.Plugins) == 0 {
		addWarn("plugin_remote_index_schema_invalid", "plugins 为空", "远程索引当前没有可展示插件。")
	}
	for _, plugin := range doc.Plugins {
		if strings.TrimSpace(plugin.Code) == "" || strings.TrimSpace(plugin.Name) == "" || strings.TrimSpace(plugin.LatestVersion) == "" {
			addErr("plugin_remote_index_schema_invalid", "插件 code/name/latest_version 缺失", "请补齐 plugins[].code/name/latest_version。")
		}
		if len(plugin.Versions) == 0 {
			addErr("plugin_remote_index_schema_invalid", "插件 versions 为空", "每个插件至少需要一个版本元数据。")
		}
		for _, ver := range plugin.Versions {
			if strings.TrimSpace(ver.Version) == "" || strings.TrimSpace(ver.PackageURL) == "" {
				addErr("plugin_remote_index_schema_invalid", "版本 version/package_url 缺失", "请补齐 versions[].version/package_url。")
			}
			if strings.TrimSpace(ver.PackageSHA256) == "" {
				addWarn("plugin_remote_index_package_url_invalid", "package_sha256 缺失", "远程包 checksum 是元数据完整性提示，缺失会标记高风险。")
			}
			if err := validateRemotePackageURL(ver.PackageURL); err != nil {
				addErr(pluginErrorCode(err, "plugin_remote_index_package_url_invalid"), "package_url 不合法："+err.Error(), "请使用 http/https 且避免 file/javascript 等危险协议。")
			}
			if strings.TrimSpace(ver.SignatureURL) == "" {
				addWarn("plugin_package_signature_missing", "signature_url 缺失", "未签名或未提供签名地址的远程包只能作为风险元数据展示。")
			} else if err := validateRemotePackageURL(ver.SignatureURL); err != nil {
				addWarn("plugin_remote_index_package_url_invalid", "signature_url 不推荐："+err.Error(), "建议使用 HTTPS signature_url。")
			}
		}
	}
	return out
}

func validateRemotePackageURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return domain.NewPluginError("plugin_remote_index_package_url_invalid", "远程包 URL 不合法").WithStatus(400)
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return domain.NewPluginError("plugin_remote_index_package_url_invalid", "远程包 URL 协议不支持").WithStatus(400).WithDetail("scheme", parsed.Scheme)
	}
	return nil
}

func (s *Service) remoteIndexDocument(id int64) (domain.PluginRemoteIndexSource, domain.PluginRemoteIndexDocument, error) {
	record, err := s.GetPluginRemoteIndex(id)
	if err != nil {
		return domain.PluginRemoteIndexSource{}, domain.PluginRemoteIndexDocument{}, err
	}
	if strings.TrimSpace(record.MetadataJSON) == "" {
		return record, domain.PluginRemoteIndexDocument{}, domain.NewPluginError("plugin_remote_index_fetch_failed", "远程索引尚未拉取").WithStatus(409).WithSuggestion("请先点击拉取索引。")
	}
	var payload struct {
		Document domain.PluginRemoteIndexDocument `json:"document"`
	}
	if err := json.Unmarshal([]byte(record.MetadataJSON), &payload); err != nil || payload.Document.SchemaVersion == "" {
		return record, domain.PluginRemoteIndexDocument{}, domain.NewPluginError("plugin_remote_index_invalid_json", "远程索引缓存无法解析").WithStatus(500)
	}
	return record, payload.Document, nil
}

func (s *Service) enrichRemotePlugin(plugin domain.PluginRemoteIndexPluginDoc) domain.PluginRemotePluginListItem {
	ver := latestRemoteVersion(plugin)
	compat := pluginregistry.CheckPluginVersionCompatibility(domain.PluginManifest{MinCoreVersion: ver.MinCoreVersion, CompatibleCoreVersion: ver.CompatibleCoreVersion}, currentCoreVersion())
	riskLevel, risks := s.remoteVersionRisks(ver, compat)
	installed, ok := s.repo.PluginByCode(plugin.Code)
	versionStatus := "not_installed"
	if ok {
		versionStatus = "installed"
		if cmp := pluginregistry.CompareVersionStrings(plugin.LatestVersion, installed.Version); cmp > 0 {
			versionStatus = "update_available"
		} else if cmp < 0 {
			versionStatus = "local_newer"
		}
	}
	return domain.PluginRemotePluginListItem{
		Code:                  plugin.Code,
		Name:                  plugin.Name,
		Description:           plugin.Description,
		LatestVersion:         plugin.LatestVersion,
		PublisherID:           ver.PublisherID,
		PublicKeyID:           ver.PublicKeyID,
		License:               ver.License,
		MinCoreVersion:        ver.MinCoreVersion,
		CompatibleCoreVersion: ver.CompatibleCoreVersion,
		PackageSHA256:         ver.PackageSHA256,
		SignatureURL:          ver.SignatureURL,
		Installed:             ok,
		LocalVersion:          installed.Version,
		InstalledStatus:       installed.Status,
		VersionStatus:         versionStatus,
		CoreCompatibility:     compat,
		PublisherTrustStatus:  s.remotePublisherTrust(ver.PublisherID, ver.PublicKeyID),
		RiskLevel:             riskLevel,
		RiskSummary:           summarizeRemoteRisks(riskLevel, risks),
		RiskItems:             risks,
	}
}

func latestRemoteVersion(plugin domain.PluginRemoteIndexPluginDoc) domain.PluginRemoteIndexVersionDoc {
	for _, ver := range plugin.Versions {
		if ver.Version == plugin.LatestVersion {
			return ver
		}
	}
	if len(plugin.Versions) > 0 {
		return plugin.Versions[0]
	}
	return domain.PluginRemoteIndexVersionDoc{}
}

func (s *Service) remotePublisherTrust(publisherID, publicKeyID string) string {
	if strings.TrimSpace(publisherID) == "" || strings.TrimSpace(publicKeyID) == "" {
		return "unknown"
	}
	if pub, ok := s.repo.PluginTrustedPublisherByKey(publisherID, publicKeyID); ok {
		switch pub.Status {
		case "trusted", "blocked", "revoked":
			return pub.Status
		}
	}
	return "unknown"
}

func (s *Service) remoteVersionRisks(ver domain.PluginRemoteIndexVersionDoc, compat domain.PluginCoreCompatibility) (string, []domain.PluginPackageRiskItem) {
	items := []domain.PluginPackageRiskItem{}
	add := func(code, level, msg, suggestion string) {
		items = append(items, remoteIndexRisk(code, level, msg, suggestion))
	}
	if strings.TrimSpace(ver.PackageSHA256) == "" {
		add("plugin_remote_index_package_url_invalid", "blocked", "远程版本缺少 package_sha256", "请要求索引源补充 package_sha256；本轮不会下载该包。")
	}
	if err := validateRemotePackageURL(ver.PackageURL); err != nil {
		add("plugin_remote_index_package_url_invalid", "blocked", "package_url 协议不合法", "仅展示 http/https 元数据；禁止 file/javascript 等协议。")
	} else if strings.HasPrefix(strings.ToLower(ver.PackageURL), "http://") {
		add("plugin_remote_index_package_url_invalid", "warning", "package_url 使用 http", "生产环境建议远程包地址使用 HTTPS。")
	}
	if strings.TrimSpace(ver.SignatureURL) == "" {
		add("plugin_package_signature_missing", "warning", "远程版本未提供 signature_url", "未签名包不一定危险，但风险更高。")
	} else if strings.HasPrefix(strings.ToLower(ver.SignatureURL), "http://") {
		add("plugin_package_signature_missing", "warning", "signature_url 使用 http", "建议使用 HTTPS 签名地址。")
	}
	switch s.remotePublisherTrust(ver.PublisherID, ver.PublicKeyID) {
	case "blocked":
		add("plugin_remote_index_publisher_blocked", "blocked", "发布者在本地可信发布者中被阻断", "请不要使用该远程包；如需恢复，请由管理员审查发布者。")
	case "revoked":
		add("plugin_remote_index_publisher_blocked", "blocked", "发布者公钥已撤销", "请使用新的可信公钥重新发布插件包。")
	case "unknown":
		add("plugin_remote_index_publisher_unknown", "warning", "发布者未在本地可信发布者中", "远程索引不能自动建立信任；请管理员核验后再决定是否信任。")
	}
	if compat.Status == pluginregistry.CompatibilityIncompatible {
		add("plugin_remote_index_core_incompatible", "blocked", "远程版本与当前 Core 不兼容", strings.Join(compat.Messages, "；"))
	}
	return maxRemoteRisk(items), items
}

func remoteIndexRisk(code, level, msg, suggestion string) domain.PluginPackageRiskItem {
	return domain.PluginPackageRiskItem{Code: code, Level: level, Message: msg, Suggestion: suggestion}
}

func maxRemoteRisk(items []domain.PluginPackageRiskItem) string {
	level := "low"
	for _, it := range items {
		switch it.Level {
		case "blocked":
			return "blocked"
		case "high":
			if level != "blocked" {
				level = "high"
			}
		case "warning":
			if level == "low" {
				level = "warning"
			}
		}
	}
	return level
}

func summarizeRemoteRisks(level string, items []domain.PluginPackageRiskItem) string {
	if len(items) == 0 {
		return "远程索引元数据未发现明显风险；仍不会下载或安装插件包。"
	}
	return fmt.Sprintf("%s：%s", level, items[0].Message)
}

func sha256Hex(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func pluginErrorCode(err error, fallback string) string {
	if apiErr, ok := err.(*domain.APIError); ok && apiErr.Code != "" {
		return apiErr.Code
	}
	return fallback
}

func (s *Service) logRemoteIndex(operator RemoteIndexOperator, action string, record domain.PluginRemoteIndexSource, code string) {
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    action,
		Target:    record.SourceID,
		Metadata:  mustJSON(map[string]any{"source_id": record.SourceID, "index_url": record.IndexURL, "status": record.Status, "last_fetch_status": record.LastFetchStatus, "error_code": code}),
		CreatedAt: Now(),
	})
}
