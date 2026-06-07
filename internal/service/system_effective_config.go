package service

import (
	"encoding/json"
	"net/url"
	"os"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

func (s *Service) SystemEffectiveConfig() (domain.SystemEffectiveConfigResponse, error) {
	allowlist, err := s.ExternalServiceHTTPAllowlist()
	if err != nil {
		return domain.SystemEffectiveConfigResponse{}, err
	}
	secrets, _, err := s.repo.SecretRefs(domain.SecretRefFilter{Page: 1, PageSize: 100})
	if err != nil {
		return domain.SystemEffectiveConfigResponse{}, err
	}
	for i := range secrets {
		secrets[i] = s.enrichSecretRefForDisplay(secrets[i])
	}
	externalServices := s.externalServiceEffectiveConfigs()
	keyStatus, _ := s.PluginConfigKeyStatus()
	keyStatus.EnvExamples = nil
	secretStatus, _ := s.SecretCenterStatus()
	resp := domain.SystemEffectiveConfigResponse{
		DevHubVersion:                currentCoreVersion(),
		StoreMode:                    currentStoreMode(),
		GeneratedAt:                  Now(),
		RootKeyStatus:                keyStatus,
		SecretCenterStatus:           secretStatus,
		ExternalServiceHTTPAllowlist: allowlist,
		HTTPAllowlistSource:          effectiveHTTPAllowlistSource(allowlist),
		ExternalServices:             externalServices,
		WebhookCallbackSecurity:      s.webhookCallbackSecuritySummary(),
		Secrets:                      secrets,
		QuickLinks: map[string]string{
			"system_settings":      "/admin-next/system?tab=effective",
			"external_policy":      "/admin-next/system?tab=external",
			"secret_center":        "/admin-next/system?tab=secret-center",
			"root_key_status":      "/admin-next/system?tab=keys",
			"audit":                "/admin-next/system?tab=audit",
			"webhook_governance":   "/admin-next/plugins/webhooks",
			"external_executions":  "/admin-next/plugins/webhooks?tab=external_service",
			"hook_failures":        "/admin-next/plugins/webhooks?tab=external_service&status=failed",
			"retry_scheduled":      "/admin-next/plugins/webhooks?tab=external_service&status=retry_scheduled",
			"manual_retry_records": "/admin-next/plugins/webhooks?tab=external_service",
		},
		Notes: []string{
			"非敏感运行配置明文展示；token、secret、Authorization、root key 与密文字段不会回显。",
			"token 通过 token_ref/secret_ref、状态、key_id 与脱敏后缀展示，真实明文仅供服务端运行时内部使用。",
		},
	}
	resp.NextSteps = s.effectiveConfigNextSteps(resp)
	resp.DiagnosticText = buildEffectiveConfigDiagnosticText(resp)
	return resp, nil
}

func (s *Service) PluginExternalServiceEffectiveConfig(pluginCode string) (domain.PluginExternalServiceEffectiveConfig, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	plugin, ok := s.PluginByCode(pluginCode)
	if !ok || plugin.Code == "" {
		return domain.PluginExternalServiceEffectiveConfig{}, pluginNotFound(pluginCode)
	}
	cfg, ok := s.PluginExternalServiceConfig(pluginCode)
	if !ok {
		return domain.PluginExternalServiceEffectiveConfig{
			PluginCode:     plugin.Code,
			PluginName:     plugin.Name,
			TokenStatus:    "missing",
			TokenMessage:   "未配置 external_service 运行配置。",
			TokenAvailable: false,
			HTTPPolicy:     ptrExternalServiceHTTPPolicy(s.externalServiceHTTPPolicy()),
		}, nil
	}
	return s.externalServiceEffectiveConfig(plugin, cfg), nil
}

func (s *Service) externalServiceEffectiveConfigs() []domain.PluginExternalServiceEffectiveConfig {
	plugins := s.Plugins()
	out := make([]domain.PluginExternalServiceEffectiveConfig, 0, len(plugins))
	for _, plugin := range plugins {
		cfg, ok := s.PluginExternalServiceConfig(plugin.Code)
		if !ok {
			continue
		}
		out = append(out, s.externalServiceEffectiveConfig(plugin, cfg))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PluginCode < out[j].PluginCode })
	return out
}

func (s *Service) externalServiceEffectiveConfig(plugin domain.Plugin, cfg domain.PluginExternalServiceConfig) domain.PluginExternalServiceEffectiveConfig {
	out := domain.PluginExternalServiceEffectiveConfig{
		PluginCode:        plugin.Code,
		PluginName:        plugin.Name,
		EndpointURL:       cfg.EndpointURL,
		HealthCheckPath:   cfg.HealthCheckPath,
		Enabled:           cfg.Enabled,
		AuthType:          normalizeExternalServiceAuthType(cfg.AuthType),
		TimeoutMS:         cfg.TimeoutMS,
		FailurePolicy:     cfg.FailurePolicy,
		CurrentHealth:     firstNonEmpty(cfg.LastHealthStatus, cfg.Status, "unknown"),
		LastHealthStatus:  firstNonEmpty(cfg.LastHealthStatus, cfg.Status, "unknown"),
		LastCheckedAt:     cfg.LastCheckedAt,
		LastHealthCheckAt: cfg.LastCheckedAt,
		LastSuccessAt:     cfg.LastSuccessAt,
		LastFailureAt:     cfg.LastFailureAt,
		LastErrorAt:       cfg.LastFailureAt,
		LastErrorMessage:  cfg.LastErrorMessage,
		LastErrorSummary:  summarizeExternalServiceError(cfg.LastErrorMessage),
		TokenRef:          cfg.TokenRef,
		TokenStatus:       "not_required",
		TokenAvailable:    true,
		TokenMessage:      "当前 auth_type 不需要 token。",
		ConfigSource:      "plugin runtime config",
		TokenSource:       "SecretCenter",
		Troubleshooting: map[string]string{
			"configure":    "/admin-next/plugins/overview?tab=list&plugin_code=" + plugin.Code + "&detail_tab=runtime",
			"health_check": "/api/v1/admin/plugins/" + plugin.Code + "/external-service/health-check",
			"executions":   "/admin-next/plugins/webhooks?tab=external_service&ext_plugin_code=" + plugin.Code,
			"audit":        "/admin-next/system?tab=audit&plugin_code=" + plugin.Code,
		},
		Diagnostics: cfg.Diagnostics,
		HTTPPolicy:  cfg.HTTPPolicy,
	}
	if out.HTTPPolicy == nil {
		out.HTTPPolicy = ptrExternalServiceHTTPPolicy(s.externalServiceHTTPPolicy())
	}
	out.EndpointScheme, out.EndpointOrigin, out.AllowlistMatched, out.AllowlistSource, out.AllowlistMessage = s.externalServiceEndpointAllowlistStatus(cfg.EndpointURL)
	if !out.Enabled {
		out.NextSteps = append(out.NextSteps, "external_service 当前未启用；如需投递或健康检查，请先到来源配置启用。")
	}
	if strings.TrimSpace(out.LastErrorMessage) != "" || strings.Contains(strings.ToLower(out.CurrentHealth), "error") {
		out.NextSteps = append(out.NextSteps, "最近健康检查或投递失败，请查看运行记录中的 failed / timeout / retry_scheduled 记录，必要时手动重试。")
	}
	if containsDockerLoopbackDiagnostic(cfg.Diagnostics) {
		out.NextSteps = append(out.NextSteps, "Docker 后端内的 127.0.0.1 指向容器自身；宿主机 receiver 请使用 host.docker.internal 或 host gateway，并按需配置 HTTP Allowlist。")
	}
	if out.EndpointScheme == "http" && !out.AllowlistMatched {
		out.NextSteps = append(out.NextSteps, "非 localhost HTTP endpoint 未命中 allowlist；生产建议改用 HTTPS，或到外部服务策略新增 exact origin。")
	}
	if normalizeExternalServiceAuthType(cfg.AuthType) != externalServiceAuthBearer {
		return out
	}
	out.TokenAvailable = false
	if strings.TrimSpace(cfg.TokenRef) == "" {
		out.TokenStatus = "missing"
		out.TokenSource = "未配置"
		out.TokenMessage = "未配置 token_ref。如果该服务需要鉴权，请在 external_service 配置中保存 token，系统会写入 SecretCenter 并生成 token_ref。"
		out.NextSteps = append(out.NextSteps, "在插件 external_service 配置中保存 token，生成 token_ref 后再执行健康检查。")
		out.NextSteps = append(out.NextSteps, "确认 root key 已通过启动环境变量配置；后台不会保存或生成 root key。")
		return out
	}
	out.Troubleshooting["secret"] = "/admin-next/system?tab=secret-center&ref=" + cfg.TokenRef
	rec, ok := s.repo.SecretRefByRef(cfg.TokenRef)
	if !ok || rec.ID == 0 {
		out.TokenStatus = "not_found"
		out.TokenMessage = "token_ref 指向的 Secret 不存在，请重新保存 token 或检查 SecretCenter。"
		out.NextSteps = append(out.NextSteps, "重新保存 external_service token，或在 SecretCenter 中查看该 token_ref 的来源与审计。")
		return out
	}
	rec = s.enrichSecretRefForDisplay(rec)
	out.TokenStatus = firstNonEmpty(rec.Status, domain.SecretRefStatusActive)
	out.TokenNamespace = rec.Namespace
	out.TokenName = rec.Name
	out.TokenMasked = rec.MaskedValue
	out.TokenKeyID = rec.KeyID
	out.TokenLastUsedAt = rec.LastUsedAt
	out.TokenRotatedAt = rec.RotatedAt
	out.TokenUsageType = rec.UsageType
	out.TokenSourceType = rec.SourceType
	out.TokenSourceID = rec.SourceID
	out.TokenSourceCode = rec.SourceCode
	out.TokenAvailable = rec.Available
	switch rec.Status {
	case domain.SecretRefStatusDisabled:
		out.TokenMessage = "该 token_ref 已禁用，投递将被阻断或失败。请轮换 token 或启用新的 Secret。"
		out.NextSteps = append(out.NextSteps, "打开 SecretCenter 查看禁用影响预览，并跳回来源配置轮换新 token。")
	case domain.SecretRefStatusRevoked:
		out.TokenMessage = "该 token_ref 已吊销，投递将被阻断或失败。请轮换 token 或启用新的 Secret。"
		out.NextSteps = append(out.NextSteps, "吊销不可直接恢复；请回到 external_service 来源配置重新写入 token。")
	default:
		out.TokenMessage = "token_ref 可用；明文不会回显。"
	}
	for _, diag := range cfg.Diagnostics {
		if diag.NeedsAllowlist || diag.SafetyRejected || diag.Code == "endpoint_safety_rejected" {
			out.NextSteps = append(out.NextSteps, "HTTP endpoint 被 allowlist 策略拒绝时，请到外部服务策略新增 exact origin，或改用 HTTPS。")
			break
		}
	}
	return out
}

func (s *Service) externalServiceEndpointAllowlistStatus(endpoint string) (scheme, origin string, matched bool, source, message string) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || u == nil || u.Scheme == "" {
		return "", "", false, "unknown", "endpoint_url 缺失或格式异常。"
	}
	scheme = strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" {
		return scheme, "", true, "default", "HTTPS endpoint 默认允许。"
	}
	origin = normalizeExternalServiceEndpointOrigin(u)
	if allowedLocalEndpoint(u) {
		return scheme, origin, true, "default", "localhost / 127.0.0.1 / ::1 HTTP 默认允许；Docker 中请注意 127.0.0.1 指向容器自身。"
	}
	if origin == "" {
		return scheme, origin, false, "unknown", "无法解析 HTTP origin，请检查 endpoint_url。"
	}
	envMatched := false
	for _, item := range envExternalServiceHTTPAllowlistEntries() {
		if item.Origin == origin {
			envMatched = true
			break
		}
	}
	adminMatched := false
	if admin, err := s.externalServiceAdminHTTPAllowlist(); err == nil {
		for _, item := range admin {
			if item.Origin == origin {
				adminMatched = true
				break
			}
		}
	}
	switch {
	case envMatched && adminMatched:
		return scheme, origin, true, "merged", "该 HTTP origin 同时命中环境变量与后台配置来源。"
	case envMatched:
		return scheme, origin, true, "environment", "该 HTTP origin 命中环境变量来源。"
	case adminMatched:
		return scheme, origin, true, "admin_setting", "该 HTTP origin 命中后台配置来源。"
	default:
		return scheme, origin, false, "empty", "非 localhost HTTP endpoint 未命中 allowlist；建议改用 HTTPS 或新增 exact origin。"
	}
}

func effectiveHTTPAllowlistSource(resp domain.PluginExternalServiceHTTPAllowlistResponse) string {
	hasEnv := len(resp.EnvAllowlist) > 0
	hasAdmin := len(resp.AdminAllowlist) > 0
	switch {
	case hasEnv && hasAdmin:
		return "merged"
	case hasEnv:
		return "environment"
	case hasAdmin:
		return "admin_setting"
	case len(resp.EffectiveAllowlist) > 0:
		return "default"
	default:
		return "empty"
	}
}

func summarizeExternalServiceError(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return ""
	}
	message = strings.ReplaceAll(message, "\n", " ")
	if len([]rune(message)) <= 180 {
		return message
	}
	runes := []rune(message)
	return string(runes[:180]) + "..."
}

func containsDockerLoopbackDiagnostic(items []domain.PluginExternalServiceDiagnostic) bool {
	for _, item := range items {
		if item.Type == "docker_loopback_hint" || item.Code == "network_connection_refused" && strings.Contains(item.Suggestion, "Docker") {
			return true
		}
	}
	return false
}

func (s *Service) webhookCallbackSecuritySummary() domain.WebhookCallbackSecuritySummary {
	webhookItems, _, _ := s.repo.PluginWebhookSecrets(domain.PluginWebhookSecretFilter{Page: 1, PageSize: 100})
	callbackItems, _, _ := s.repo.PluginCallbackTokens(domain.PluginCallbackTokenFilter{Page: 1, PageSize: 100})
	out := domain.WebhookCallbackSecuritySummary{
		WebhookSecretByStatus: map[string]int{},
		CallbackTokenByStatus: map[string]int{},
		QuickLinks: map[string]string{
			"webhook_secrets":       "/admin-next/plugins/webhooks?tab=secrets",
			"callback_tokens":       "/admin-next/plugins/webhooks?tab=callback_tokens",
			"callback_requests":     "/admin-next/plugins/webhooks?tab=callback_requests",
			"external_executions":   "/admin-next/plugins/webhooks?tab=external_service",
			"audit":                 "/admin-next/system?tab=audit&metadata=webhook callback token_ref secret_ref",
			"manual_retry_failures": "/admin-next/plugins/webhooks?tab=external_service&status=failed",
		},
		Notes: []string{
			"Webhook Secret 明文只在创建或轮换响应中出现一次；当前视图只展示状态计数与 secret_ref 元数据。",
			"Callback Token 明文只在创建或轮换响应中出现一次；存储侧只保留 hash 与 token_ref 元数据。",
		},
	}
	for _, item := range webhookItems {
		status := firstNonEmpty(item.Status, domain.PluginWebhookSecretStatusActive)
		out.WebhookSecretByStatus[status]++
		out.WebhookSecretTotal++
		if status == domain.PluginWebhookSecretStatusActive || status == domain.PluginWebhookSecretStatusPrevious {
			out.ActiveWebhookSecrets++
		}
		if status == domain.PluginWebhookSecretStatusDisabled || status == domain.PluginWebhookSecretStatusRevoked {
			out.DisabledOrRevokedCount++
		}
		out.LastWebhookSecretUsedAt = maxTimeText(out.LastWebhookSecretUsedAt, item.LastUsedAt)
	}
	for _, item := range callbackItems {
		status := firstNonEmpty(item.Status, domain.PluginCallbackTokenStatusActive)
		out.CallbackTokenByStatus[status]++
		out.CallbackTokenTotal++
		if status == domain.PluginCallbackTokenStatusActive {
			out.ActiveCallbackTokens++
		}
		if status == domain.PluginCallbackTokenStatusDisabled || status == domain.PluginCallbackTokenStatusRevoked {
			out.DisabledOrRevokedCount++
		}
		out.LastCallbackTokenUsedAt = maxTimeText(out.LastCallbackTokenUsedAt, item.LastUsedAt)
	}
	return out
}

func maxTimeText(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" {
		return b
	}
	if b == "" || a >= b {
		return a
	}
	return b
}

func (s *Service) effectiveConfigNextSteps(resp domain.SystemEffectiveConfigResponse) []string {
	steps := []string{}
	keyStatus, _ := s.PluginConfigKeyStatus()
	if strings.TrimSpace(keyStatus.Status) != "ok" {
		steps = append(steps, "root key 未就绪时只能通过启动环境变量或外部 Secret 系统注入，后台不会保存、生成或修改。")
	}
	for _, svc := range resp.ExternalServices {
		steps = append(steps, svc.NextSteps...)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(steps))
	for _, step := range steps {
		step = strings.TrimSpace(step)
		if step == "" || seen[step] {
			continue
		}
		seen[step] = true
		out = append(out, step)
	}
	return out
}

func buildEffectiveConfigDiagnosticText(resp domain.SystemEffectiveConfigResponse) string {
	type diagnosticService struct {
		PluginCode        string   `json:"plugin_code"`
		PluginName        string   `json:"plugin_name,omitempty"`
		EndpointURL       string   `json:"endpoint_url,omitempty"`
		EndpointOrigin    string   `json:"endpoint_origin,omitempty"`
		HealthCheckPath   string   `json:"health_check_path,omitempty"`
		Enabled           bool     `json:"enabled"`
		AuthType          string   `json:"auth_type,omitempty"`
		TimeoutMS         int      `json:"timeout_ms,omitempty"`
		FailurePolicy     string   `json:"failure_policy,omitempty"`
		CurrentHealth     string   `json:"current_health,omitempty"`
		LastHealthStatus  string   `json:"last_health_status,omitempty"`
		LastSuccessAt     string   `json:"last_success_at,omitempty"`
		LastErrorAt       string   `json:"last_error_at,omitempty"`
		TokenRef          string   `json:"token_ref,omitempty"`
		TokenStatus       string   `json:"token_status,omitempty"`
		TokenKeyID        string   `json:"token_key_id,omitempty"`
		TokenMasked       string   `json:"token_masked,omitempty"`
		TokenNamespace    string   `json:"token_namespace,omitempty"`
		TokenName         string   `json:"token_name,omitempty"`
		TokenUsageType    string   `json:"token_usage_type,omitempty"`
		TokenSourceType   string   `json:"token_source_type,omitempty"`
		TokenSourceID     string   `json:"token_source_id,omitempty"`
		TokenSourceCode   string   `json:"token_source_code,omitempty"`
		ConfigSource      string   `json:"config_source,omitempty"`
		TokenSource       string   `json:"token_source,omitempty"`
		AllowlistSource   string   `json:"http_allowlist_source,omitempty"`
		AllowlistMatched  bool     `json:"http_allowlist_matched"`
		AllowlistMessage  string   `json:"http_allowlist_message,omitempty"`
		NextSteps         []string `json:"next_steps,omitempty"`
		LastErrorSummary  string   `json:"last_error_summary,omitempty"`
		LastHealthCheckAt string   `json:"last_health_check_at,omitempty"`
	}
	payload := struct {
		DevHubVersion              string                                `json:"devhub_version"`
		StoreMode                  string                                `json:"store_mode"`
		GeneratedAt                string                                `json:"generated_at"`
		RootKeyStatus              string                                `json:"root_key_status"`
		RootKeyConfigured          bool                                  `json:"root_key_configured"`
		SecretCenterStatus         string                                `json:"secret_center_status"`
		SecretRefCount             int                                   `json:"secret_ref_count"`
		ExternalServices           []diagnosticService                   `json:"external_services"`
		WebhookCallbackSecurity    domain.WebhookCallbackSecuritySummary `json:"webhook_callback_security"`
		HTTPAllowlistSource        string                                `json:"http_allowlist_source"`
		HTTPAllowlistEffectiveList []string                              `json:"http_allowlist_effective_list"`
		HTTPAllowlistDefaults      []string                              `json:"http_allowlist_defaults"`
		HTTPAllowlistEnv           []string                              `json:"http_allowlist_env"`
		HTTPAllowlistAdmin         []string                              `json:"http_allowlist_admin"`
	}{
		DevHubVersion:              resp.DevHubVersion,
		StoreMode:                  resp.StoreMode,
		GeneratedAt:                resp.GeneratedAt,
		RootKeyStatus:              resp.RootKeyStatus.Status,
		RootKeyConfigured:          resp.RootKeyStatus.Status == "ok",
		SecretCenterStatus:         resp.SecretCenterStatus.Status,
		SecretRefCount:             resp.SecretCenterStatus.SecretRefCount,
		WebhookCallbackSecurity:    resp.WebhookCallbackSecurity,
		HTTPAllowlistSource:        resp.HTTPAllowlistSource,
		HTTPAllowlistEffectiveList: resp.ExternalServiceHTTPAllowlist.Policy.EffectiveAllowlist,
		HTTPAllowlistDefaults:      resp.ExternalServiceHTTPAllowlist.Policy.Defaults,
		HTTPAllowlistEnv:           resp.ExternalServiceHTTPAllowlist.Policy.EnvAllowlist,
		HTTPAllowlistAdmin:         resp.ExternalServiceHTTPAllowlist.Policy.AdminAllowlist,
	}
	for _, svc := range resp.ExternalServices {
		payload.ExternalServices = append(payload.ExternalServices, diagnosticService{
			PluginCode:        svc.PluginCode,
			PluginName:        svc.PluginName,
			EndpointURL:       sanitizeExternalServiceEndpointForDiagnostics(svc.EndpointURL),
			EndpointOrigin:    svc.EndpointOrigin,
			HealthCheckPath:   svc.HealthCheckPath,
			Enabled:           svc.Enabled,
			AuthType:          svc.AuthType,
			TimeoutMS:         svc.TimeoutMS,
			FailurePolicy:     svc.FailurePolicy,
			CurrentHealth:     svc.CurrentHealth,
			LastHealthStatus:  svc.LastHealthStatus,
			LastSuccessAt:     svc.LastSuccessAt,
			LastErrorAt:       svc.LastErrorAt,
			TokenRef:          svc.TokenRef,
			TokenStatus:       svc.TokenStatus,
			TokenKeyID:        svc.TokenKeyID,
			TokenMasked:       svc.TokenMasked,
			TokenNamespace:    svc.TokenNamespace,
			TokenName:         svc.TokenName,
			TokenUsageType:    svc.TokenUsageType,
			TokenSourceType:   svc.TokenSourceType,
			TokenSourceID:     svc.TokenSourceID,
			TokenSourceCode:   svc.TokenSourceCode,
			ConfigSource:      svc.ConfigSource,
			TokenSource:       svc.TokenSource,
			AllowlistSource:   svc.AllowlistSource,
			AllowlistMatched:  svc.AllowlistMatched,
			AllowlistMessage:  svc.AllowlistMessage,
			NextSteps:         svc.NextSteps,
			LastErrorSummary:  sanitizeExternalServiceEndpointForDiagnostics(svc.LastErrorSummary),
			LastHealthCheckAt: svc.LastHealthCheckAt,
		})
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return ""
	}
	return string(raw)
}

func currentStoreMode() string {
	if strings.TrimSpace(os.Getenv("CMS_STORE")) == "memory" {
		return "memory"
	}
	return "mysql"
}
