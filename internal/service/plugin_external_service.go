package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const (
	externalServiceTypeExternal             = "external_service"
	externalServiceHealthUnknown            = "unknown"
	externalServiceHealthHealthy            = "healthy"
	externalServiceHealthWarning            = "warning"
	externalServiceHealthError              = "error"
	externalServiceHealthDisabled           = "disabled"
	externalServiceHealthSkipped            = "skipped"
	externalServiceFailurePolicyIgnore      = "ignore"
	externalServiceFailurePolicyWarn        = "warn"
	externalServiceFailurePolicyError       = "error"
	externalServiceFailurePolicyDisableHook = "disable_hook"
	externalServiceAuthNone                 = "none"
	externalServiceAuthBearer               = "bearer"
)

type PluginExternalServiceOperator struct {
	ID   int64
	Name string
}

func (s *Service) PluginExternalServiceConfig(pluginCode string) (domain.PluginExternalServiceConfig, bool) {
	if s == nil || s.repo == nil {
		return domain.PluginExternalServiceConfig{}, false
	}
	cfg, ok := s.repo.PluginExternalServiceConfig(pluginCode)
	if !ok {
		return domain.PluginExternalServiceConfig{}, false
	}
	// Attach SecretCenter token metadata when token_ref is in secret://... form.
	if normalizeExternalServiceAuthType(cfg.AuthType) == externalServiceAuthBearer && domain.IsSecretRef(cfg.TokenRef) {
		if rec, ok := s.repo.SecretRefByRef(cfg.TokenRef); ok && rec.ID > 0 {
			rec.EncryptedValue = ""
			cfg.TokenSecret = &rec
		} else {
			// Keep response stable; surface missing ref via diagnostics.
			cfg.Diagnostics = append(cfg.Diagnostics, externalServiceEndpointDiagnostic(cfg.EndpointURL, "secret_ref_not_found", nil))
		}
	}
	return s.enrichExternalServiceConfigDiagnostics(cfg), true
}

func (s *Service) UpdatePluginExternalServiceConfig(operator PluginExternalServiceOperator, pluginCode string, req domain.PluginExternalServiceUpdateRequest) (domain.PluginExternalServiceConfig, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" {
		return domain.PluginExternalServiceConfig{}, domain.NewPluginError("plugin_external_service_invalid", "plugin_code 不能为空").WithStatus(http.StatusBadRequest)
	}
	plugin, ok := s.PluginByCode(pluginCode)
	if !ok || plugin.Code == "" {
		return domain.PluginExternalServiceConfig{}, pluginNotFound(pluginCode)
	}
	if plugin.Status == pluginregistry.StatusArchived {
		return domain.PluginExternalServiceConfig{}, pluginArchived(pluginCode)
	}
	current, _ := s.repo.PluginExternalServiceConfig(pluginCode)
	if err := s.validateExternalServiceUpdateRequest(req, current); err != nil {
		return domain.PluginExternalServiceConfig{}, err
	}
	tokenRef := current.TokenRef
	tokenCiphertext := current.TokenCiphertext
	tokenHash := current.TokenHash
	authType := normalizeExternalServiceAuthType(req.AuthType)
	if authType == externalServiceAuthNone {
		tokenRef = ""
		tokenCiphertext = ""
		tokenHash = ""
	}
	if authType == externalServiceAuthBearer && strings.TrimSpace(req.Token) != "" {
		// v1.8.4-S14: Store bearer token in Core SecretCenter and persist only token_ref.
		ref, err := s.ExternalServiceTokenRef(pluginCode)
		if err != nil {
			return domain.PluginExternalServiceConfig{}, err
		}
		secOp := SecretOperator{Type: "admin_user", ID: operator.ID, Name: operator.Name}
		if _, ok := s.repo.SecretRefByRef(ref); ok {
			if _, err := s.UpdateSecret(secOp, domain.SecretUpdateRequest{Ref: ref, Value: strings.TrimSpace(req.Token)}); err != nil {
				return domain.PluginExternalServiceConfig{}, err
			}
		} else {
			if _, err := s.CreateSecret(secOp, domain.SecretCreateRequest{
				Namespace:   secretCenterNamespaceExternalService,
				Name:        pluginCode + "/token",
				Value:       strings.TrimSpace(req.Token),
				Description: "external_service bearer token for plugin " + pluginCode,
			}); err != nil {
				return domain.PluginExternalServiceConfig{}, err
			}
		}
		tokenRef = ref
		tokenCiphertext = ""
		tokenHash = ""
	}
	enabled := current.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	status := current.Status
	if !enabled {
		status = externalServiceHealthDisabled
	} else if status == "" {
		status = externalServiceHealthUnknown
	}

	record := domain.PluginExternalServiceConfig{
		PluginCode:       pluginCode,
		ServiceType:      externalServiceTypeExternal,
		EndpointURL:      strings.TrimSpace(req.EndpointURL),
		HealthCheckPath:  normalizeHealthCheckPath(req.HealthCheckPath),
		TimeoutMS:        normalizeExternalServiceTimeout(req.TimeoutMS),
		FailurePolicy:    normalizeExternalServiceFailurePolicy(req.FailurePolicy),
		AuthType:         authType,
		TokenRef:         tokenRef,
		TokenCiphertext:  tokenCiphertext,
		TokenHash:        tokenHash,
		Enabled:          enabled,
		Status:           status,
		LastHealthStatus: firstNonEmpty(current.LastHealthStatus, externalServiceHealthUnknown),
		LastCheckedAt:    current.LastCheckedAt,
		LastSuccessAt:    current.LastSuccessAt,
		LastFailureAt:    current.LastFailureAt,
		FailureCount:     current.FailureCount,
		WarningThreshold: normalizeExternalServiceThreshold(req.WarningThreshold, 3),
		ErrorThreshold:   normalizeExternalServiceThreshold(req.ErrorThreshold, 5),
		LastErrorMessage: current.LastErrorMessage,
		CreatedAt:        current.CreatedAt,
	}
	saved, err := s.repo.SavePluginExternalServiceConfig(record)
	if err != nil {
		return domain.PluginExternalServiceConfig{}, err
	}
	if err := s.refreshPluginRegistry(pluginRegistryRefreshEvent{
		Trigger:    "after_external_service_change",
		PluginCode: pluginCode,
		ActorType:  "admin_user",
		ActorID:    operator.ID,
		ActorName:  firstNonEmpty(operator.Name, "system"),
		Status:     plugin.Status,
		NewVersion: plugin.Version,
	}); err != nil {
		return domain.PluginExternalServiceConfig{}, err
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.external_service.updated",
		Target:    "plugins#" + pluginCode,
		Metadata: mustJSON(map[string]any{
			"plugin_code":    pluginCode,
			"service_type":   externalServiceTypeExternal,
			"endpoint_url":   saved.EndpointURL,
			"enabled":        saved.Enabled,
			"auth_type":      saved.AuthType,
			"token_ref":      saved.TokenRef,
			"failure_policy": saved.FailurePolicy,
		}),
		CreatedAt: Now(),
	})
	if out, ok := s.PluginExternalServiceConfig(pluginCode); ok {
		return out, nil
	}
	return s.enrichExternalServiceConfigDiagnostics(saved), nil
}

func (s *Service) RunPluginExternalServiceHealthCheck(operator PluginExternalServiceOperator, pluginCode string) (domain.PluginExternalServiceHealthCheckResponse, error) {
	cfg, ok := s.PluginExternalServiceConfig(pluginCode)
	if !ok {
		return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_not_found", "external_service 配置不存在").WithStatus(http.StatusNotFound)
	}
	if !cfg.Enabled {
		return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_disabled", "external_service 已停用").
			WithStatus(http.StatusBadRequest).
			WithSuggestion(externalServiceDiagnosticSuggestion(cfg.EndpointURL, externalServiceReasonServiceDisabled, nil)).
			WithDiagnostic(externalServiceEndpointDiagnostic(cfg.EndpointURL, externalServiceReasonServiceDisabled, nil))
	}
	plugin, ok := s.PluginByCode(pluginCode)
	if !ok || plugin.Code == "" {
		return domain.PluginExternalServiceHealthCheckResponse{}, pluginNotFound(pluginCode)
	}
	if plugin.Status == pluginregistry.StatusDisabled || plugin.Status == pluginregistry.StatusArchived {
		return s.recordExternalServiceHealth(operator, pluginCode, cfg, externalServiceHealthSkipped, 0, "", "skipped", 0, nil)
	}
	if strings.TrimSpace(cfg.EndpointURL) == "" {
		return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_endpoint_missing", "external_service.endpoint_url 不能为空").WithStatus(http.StatusBadRequest)
	}
	if s.externalServiceEndpointSafetyRejected(cfg.EndpointURL) {
		return s.recordExternalServiceHealth(operator, pluginCode, cfg, externalServiceHealthError, 0, "", "endpoint_safety_rejected", 0, nil)
	}

	started := time.Now()
	reqURL, err := url.Parse(cfg.EndpointURL)
	if err != nil {
		return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_invalid_endpoint", "服务地址不合法").WithStatus(http.StatusBadRequest)
	}
	reqURL.Path = path.Join(strings.TrimRight(reqURL.Path, "/"), normalizeHealthCheckPath(cfg.HealthCheckPath))
	client := &http.Client{Timeout: time.Duration(normalizeExternalServiceTimeout(cfg.TimeoutMS)) * time.Millisecond}
	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return domain.PluginExternalServiceHealthCheckResponse{}, err
	}
	req.Header.Set("User-Agent", "DevHub-ExternalService-HealthCheck/1.0")
	if cfg.AuthType == externalServiceAuthBearer {
		token := ""
		if domain.IsSecretRef(cfg.TokenRef) {
			plain, _, rerr := s.ResolveSecretInternal(cfg.TokenRef)
			if rerr != nil {
				// Keep external_service-facing error codes stable.
				code := "plugin_external_service_token_missing"
				msg := "external_service bearer token 缺失"
				if apiErr, ok := rerr.(*domain.APIError); ok {
					switch strings.TrimSpace(apiErr.Code) {
					case "secret_ref_disabled":
						code = "plugin_external_service_token_disabled"
						msg = "external_service token_ref 已停用"
					case "secret_ref_revoked":
						code = "plugin_external_service_token_revoked"
						msg = "external_service token_ref 已吊销"
					case "secret_ref_decrypt_failed":
						code = "plugin_external_service_token_invalid"
						msg = "external_service token 无法解密"
					}
				}
				return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError(code, msg).
					WithStatus(http.StatusBadRequest).
					WithDetail("plugin_code", pluginCode).
					WithDetail("token_ref", cfg.TokenRef).
					WithSuggestion(externalServiceDiagnosticSuggestion(cfg.EndpointURL, externalServiceReasonTokenMissing, nil)).
					WithDiagnostic(externalServiceEndpointDiagnostic(cfg.EndpointURL, externalServiceReasonTokenMissing, nil))
			}
			token = plain
		} else {
			// Backward compat: legacy token_ciphertext storage.
			if strings.TrimSpace(cfg.TokenCiphertext) == "" {
				return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_token_missing", "external_service bearer token 缺失").
					WithStatus(http.StatusBadRequest).
					WithDetail("plugin_code", pluginCode).
					WithSuggestion(externalServiceDiagnosticSuggestion(cfg.EndpointURL, externalServiceReasonTokenMissing, nil)).
					WithDiagnostic(externalServiceEndpointDiagnostic(cfg.EndpointURL, externalServiceReasonTokenMissing, nil))
			}
			plain, derr := decryptExternalServiceToken(cfg.TokenCiphertext)
			if derr != nil {
				return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_token_invalid", "external_service token 无法解密").
					WithStatus(http.StatusBadRequest).
					WithDetail("plugin_code", pluginCode).
					WithSuggestion("请检查插件配置加密 key 是否可用，必要时重新写入 external_service Bearer Token；不会回显 token 明文。").
					WithDiagnostic(externalServiceEndpointDiagnostic(cfg.EndpointURL, externalServiceReasonTokenInvalid, derr))
			}
			token = plain
		}
		if strings.TrimSpace(token) == "" {
			return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_token_missing", "external_service bearer token 缺失").
				WithStatus(http.StatusBadRequest).
				WithDetail("plugin_code", pluginCode).
				WithSuggestion(externalServiceDiagnosticSuggestion(cfg.EndpointURL, externalServiceReasonTokenMissing, nil)).
				WithDiagnostic(externalServiceEndpointDiagnostic(cfg.EndpointURL, externalServiceReasonTokenMissing, nil))
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	duration := time.Since(started).Milliseconds()
	if err != nil {
		diag := externalServiceNetworkDiagnostic(cfg.EndpointURL, err, err)
		return s.recordExternalServiceHealth(operator, pluginCode, cfg, diag.Status, 0, "", diag.Code, duration, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	status := externalServiceHealthHealthy
	errorCode := ""
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		status = externalServiceHealthHealthy
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		status = externalServiceHealthWarning
		errorCode = "http_status_failed"
	default:
		status = externalServiceHealthError
		errorCode = "http_status_failed"
	}
	return s.recordExternalServiceHealth(operator, pluginCode, cfg, status, resp.StatusCode, string(body), errorCode, duration, nil)
}

func (s *Service) recordExternalServiceHealth(operator PluginExternalServiceOperator, pluginCode string, cfg domain.PluginExternalServiceConfig, status string, responseStatus int, responseBody string, errorCode string, durationMS int64, cause error) (domain.PluginExternalServiceHealthCheckResponse, error) {
	now := Now()
	nextFailureCount := cfg.FailureCount
	lastHealth := externalServiceHealthHealthy
	switch status {
	case externalServiceHealthHealthy:
		nextFailureCount = 0
		lastHealth = externalServiceHealthHealthy
	case externalServiceHealthSkipped:
		lastHealth = externalServiceHealthSkipped
	case externalServiceHealthWarning:
		nextFailureCount++
		if nextFailureCount >= normalizeExternalServiceThreshold(cfg.WarningThreshold, 3) {
			lastHealth = externalServiceHealthWarning
		} else {
			lastHealth = externalServiceHealthHealthy
		}
	default:
		nextFailureCount++
		switch normalizeExternalServiceFailurePolicy(cfg.FailurePolicy) {
		case externalServiceFailurePolicyIgnore:
			lastHealth = cfg.LastHealthStatus
		case externalServiceFailurePolicyWarn:
			if nextFailureCount >= normalizeExternalServiceThreshold(cfg.WarningThreshold, 3) {
				lastHealth = externalServiceHealthWarning
			} else {
				lastHealth = externalServiceHealthHealthy
			}
		default:
			if nextFailureCount >= normalizeExternalServiceThreshold(cfg.ErrorThreshold, 5) {
				lastHealth = externalServiceHealthError
			} else if nextFailureCount >= normalizeExternalServiceThreshold(cfg.WarningThreshold, 3) {
				lastHealth = externalServiceHealthWarning
			} else {
				lastHealth = externalServiceHealthHealthy
			}
		}
	}
	updated := cfg
	updated.LastHealthStatus = lastHealth
	updated.LastCheckedAt = now
	updated.LastErrorMessage = externalServiceDiagnosticMessage(cfg.EndpointURL, errorCode, cause, responseStatus, responseBody)
	updated.FailureCount = nextFailureCount
	if status == externalServiceHealthHealthy {
		updated.LastSuccessAt = now
		updated.LastErrorMessage = ""
	} else {
		updated.LastFailureAt = now
	}
	updated.Status = lastHealth
	if !updated.Enabled {
		updated.Status = externalServiceHealthDisabled
	}
	saved, saveErr := s.repo.SavePluginExternalServiceConfig(updated)
	if saveErr != nil {
		return domain.PluginExternalServiceHealthCheckResponse{}, saveErr
	}
	execStatus := status
	if execStatus == externalServiceHealthHealthy {
		execStatus = "success"
	}
	success := status == externalServiceHealthHealthy
	execMetadata := map[string]any{
		"plugin_code":          pluginCode,
		"service_type":         externalServiceTypeExternal,
		"failure_policy":       saved.FailurePolicy,
		"auth_type":            saved.AuthType,
		"health_status_before": cfg.LastHealthStatus,
		"health_status_after":  saved.LastHealthStatus,
	}
	if strings.TrimSpace(errorCode) != "" {
		execMetadata["diagnostic_failure_type"] = externalServiceFailureType(errorCode)
		execMetadata["diagnostic_suggestion"] = externalServiceDiagnosticSuggestion(saved.EndpointURL, errorCode, cause)
		execMetadata["diagnostics"] = []domain.PluginExternalServiceDiagnostic{externalServiceEndpointDiagnostic(saved.EndpointURL, errorCode, cause)}
	}
	exec := domain.HookExecution{
		PluginCode:          pluginCode,
		ServiceType:         externalServiceTypeExternal,
		EndpointURL:         saved.EndpointURL,
		HookName:            "external_service.health_check",
		Mode:                string(pluginregistry.HookNonBlocking),
		RequestID:           hookRequestID(""),
		StartedAt:           now,
		FinishedAt:          now,
		DurationMS:          int(durationMS),
		Success:             success,
		ErrorMessage:        saved.LastErrorMessage,
		Status:              execStatus,
		ResponseStatus:      responseStatus,
		ResponseBodyExcerpt: trimExternalServiceExcerpt(responseBody),
		ErrorCode:           errorCode,
		Metadata:            mustJSON(execMetadata),
	}
	if _, err := s.repo.AppendHookExecution(exec); err == nil {
		// best-effort record only.
	}
	if cause != nil {
		s.repo.AppendAdminLog(domain.AdminLog{
			Site:      "admin",
			Actor:     firstNonEmpty(operator.Name, "system"),
			ActorType: "admin_user",
			ActorID:   operator.ID,
			Action:    "plugin.external_service.health.failed",
			Target:    "plugins#" + pluginCode,
			Metadata:  mustJSON(map[string]any{"plugin_code": pluginCode, "status": saved.LastHealthStatus, "error": trimExternalServiceErr(cause, responseBody), "duration_ms": durationMS}),
			CreatedAt: now,
		})
	}
	diagnostics := []domain.PluginExternalServiceDiagnostic{}
	if strings.TrimSpace(errorCode) != "" {
		diagnostics = append(diagnostics, externalServiceEndpointDiagnostic(saved.EndpointURL, errorCode, cause))
	}
	return domain.PluginExternalServiceHealthCheckResponse{
		PluginCode:   pluginCode,
		ServiceType:  saved.ServiceType,
		EndpointURL:  saved.EndpointURL,
		HealthStatus: saved.LastHealthStatus,
		Status:       saved.Status,
		CheckedAt:    saved.LastCheckedAt,
		DurationMS:   durationMS,
		Message:      saved.LastErrorMessage,
		ErrorCode:    errorCode,
		Suggestion:   externalServiceDiagnosticSuggestion(saved.EndpointURL, errorCode, cause),
		Diagnostics:  diagnostics,
		HTTPPolicy:   ptrExternalServiceHTTPPolicy(s.externalServiceHTTPPolicy()),
	}, nil
}

func (s *Service) validateExternalServiceUpdateRequest(req domain.PluginExternalServiceUpdateRequest, current domain.PluginExternalServiceConfig) error {
	endpoint := strings.TrimSpace(req.EndpointURL)
	if endpoint == "" {
		return domain.NewPluginError("plugin_external_service_invalid_endpoint", "external_service.endpoint_url 不能为空").WithStatus(http.StatusBadRequest)
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return domain.NewPluginError("plugin_external_service_invalid_endpoint", "external_service.endpoint_url 必须是合法 URL").WithStatus(http.StatusBadRequest)
	}
	if s.externalServiceEndpointSafetyRejected(endpoint) {
		diag := externalServiceEndpointDiagnostic(endpoint, "endpoint_safety_rejected", nil)
		return domain.NewPluginError("endpoint_safety_rejected", "安全策略拒绝该 external_service HTTP endpoint").
			WithStatus(http.StatusBadRequest).
			WithDetail("endpoint_url", sanitizeExternalServiceEndpointForDiagnostics(endpoint)).
			WithDetail("safety_rejected", true).
			WithDetail("needs_allowlist", true).
			WithDetail("allowlist_env", "DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST").
			WithDetail("allowlist_example", externalServiceAllowlistExample()).
			WithDetail("policy", s.externalServiceHTTPPolicy()).
			WithSuggestion("非 localhost 的 HTTP endpoint 默认不允许。生产建议使用 HTTPS；本地开发可通过 DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST 显式放行该 origin，例如 DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build。").
			WithDiagnostic(diag)
	}
	if strings.HasPrefix(strings.ToLower(endpoint), "javascript:") || strings.HasPrefix(strings.ToLower(endpoint), "data:") || strings.HasPrefix(strings.ToLower(endpoint), "file:") || strings.HasPrefix(strings.ToLower(endpoint), "ftp:") {
		return domain.NewPluginError("plugin_external_service_invalid_endpoint", "external_service.endpoint_url 协议不被允许").WithStatus(http.StatusBadRequest)
	}
	timeout := normalizeExternalServiceTimeout(req.TimeoutMS)
	if timeout <= 0 {
		return domain.NewPluginError("plugin_external_service_invalid_timeout", "timeout_ms 不合法").WithStatus(http.StatusBadRequest)
	}
	policy := normalizeExternalServiceFailurePolicy(req.FailurePolicy)
	if policy == "" {
		return domain.NewPluginError("plugin_external_service_invalid_policy", "failure_policy 不合法").WithStatus(http.StatusBadRequest)
	}
	authType := normalizeExternalServiceAuthType(req.AuthType)
	if authType == "" {
		return domain.NewPluginError("plugin_external_service_invalid_auth_type", "auth_type 不合法").WithStatus(http.StatusBadRequest)
	}
	if authType == externalServiceAuthBearer && strings.TrimSpace(req.Token) == "" && strings.TrimSpace(current.TokenCiphertext) == "" && !domain.IsSecretRef(current.TokenRef) {
		return domain.NewPluginError("plugin_external_service_invalid_token", "bearer token 不能为空").WithStatus(http.StatusBadRequest)
	}
	return nil
}

func normalizeHealthCheckPath(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "/health"
	}
	if strings.Contains(v, "://") {
		return "/health"
	}
	v = strings.ReplaceAll(v, "\\", "/")
	if !strings.HasPrefix(v, "/") {
		v = "/" + v
	}
	cleaned := path.Clean(v)
	if cleaned == "." || cleaned == "/" {
		return "/health"
	}
	if strings.Contains(cleaned, "..") {
		return "/health"
	}
	return cleaned
}

func normalizeExternalServiceTimeout(timeoutMS int) int {
	if timeoutMS <= 0 {
		return 3000
	}
	if timeoutMS < 500 {
		return 500
	}
	if timeoutMS > 10000 {
		return 10000
	}
	return timeoutMS
}

func normalizeExternalServiceFailurePolicy(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", externalServiceFailurePolicyWarn:
		return externalServiceFailurePolicyWarn
	case externalServiceFailurePolicyIgnore, externalServiceFailurePolicyError, externalServiceFailurePolicyDisableHook:
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

func normalizeExternalServiceAuthType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", externalServiceAuthNone:
		return externalServiceAuthNone
	case externalServiceAuthBearer:
		return externalServiceAuthBearer
	default:
		return ""
	}
}

func normalizeExternalServiceThreshold(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	if v > 100 {
		return 100
	}
	return v
}

func externalServiceEndpointSafetyRejected(endpoint string) bool {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || u == nil {
		return false
	}
	return strings.EqualFold(u.Scheme, "http") && !allowedLocalEndpoint(u) && !allowedExternalServiceHTTPAllowlist(u)
}

func (s *Service) externalServiceEndpointSafetyRejected(endpoint string) bool {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || u == nil {
		return false
	}
	return strings.EqualFold(u.Scheme, "http") && !allowedLocalEndpoint(u) && !s.allowedExternalServiceHTTPAllowlist(u)
}

func allowedLocalEndpoint(u *url.URL) bool {
	if u == nil {
		return false
	}
	if strings.ToLower(u.Scheme) != "http" {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func allowedExternalServiceHTTPAllowlist(u *url.URL) bool {
	if u == nil || strings.ToLower(u.Scheme) != "http" {
		return false
	}
	target := normalizeExternalServiceEndpointOrigin(u)
	if target == "" {
		return false
	}
	for _, item := range envExternalServiceHTTPAllowlistEntries() {
		if item.Origin == target {
			return true
		}
	}
	return false
}

func (s *Service) allowedExternalServiceHTTPAllowlist(u *url.URL) bool {
	if u == nil || strings.ToLower(u.Scheme) != "http" {
		return false
	}
	target := normalizeExternalServiceEndpointOrigin(u)
	if target == "" {
		return false
	}
	for _, item := range envExternalServiceHTTPAllowlistEntries() {
		if item.Origin == target {
			return true
		}
	}
	if s != nil {
		admin, _ := s.externalServiceAdminHTTPAllowlist()
		for _, item := range admin {
			allowed := strings.TrimSpace(item.Origin)
			if allowed != "" && allowed == target {
				return true
			}
		}
	}
	return false
}

func allowedExternalServiceHTTPAllowlistEnvOnly(u *url.URL) bool {
	if u == nil || strings.ToLower(u.Scheme) != "http" {
		return false
	}
	target := normalizeExternalServiceEndpointOrigin(u)
	if target == "" {
		return false
	}
	for _, item := range splitExternalServiceAllowlist(os.Getenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST")) {
		allowed, err := validateExternalServiceHTTPAllowlistOrigin(item, true)
		if err != nil {
			allowed = normalizeExternalServiceAllowlistItem(item)
		}
		if allowed != "" && allowed == target {
			return true
		}
	}
	return false
}

type externalServiceNetworkFailure struct {
	Code   string
	Status string
}

func externalServiceNetworkDiagnostic(_ string, err error, cause error) externalServiceNetworkFailure {
	msg := strings.ToLower(strings.TrimSpace(firstNonEmpty(errorString(err), errorString(cause))))
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "connect: connection refused") {
		return externalServiceNetworkFailure{Code: "network_connection_refused", Status: externalServiceHealthError}
	}
	if strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "i/o timeout") {
		return externalServiceNetworkFailure{Code: "network_timeout", Status: "timeout"}
	}
	return externalServiceNetworkFailure{Code: "network_request_failed", Status: externalServiceHealthError}
}

func externalServiceDiagnosticMessage(endpoint, code string, err error, responseStatus int, responseBody string) string {
	switch strings.TrimSpace(code) {
	case "network_connection_refused":
		return "外部服务连接被拒绝。" + externalServiceDiagnosticSuggestion(endpoint, code, err)
	case "network_timeout":
		return "外部服务请求超时。请检查 receiver 是否启动、端口是否可达、Docker 网络是否连通，以及 timeout_ms 是否过短。"
	case "http_status_failed":
		if responseStatus > 0 {
			return "外部服务返回非 2xx/3xx 状态：" + http.StatusText(responseStatus)
		}
		return "外部服务返回非 2xx/3xx 状态。"
	case "endpoint_safety_rejected":
		return "安全策略拒绝该 HTTP endpoint。" + externalServiceHTTPPolicyText()
	case externalServiceReasonTokenMissing, "token_missing":
		return "external_service bearer token 缺失。"
	case externalServiceReasonTokenInvalid, "token_invalid":
		return "external_service bearer token 无法解密或不可用。"
	case externalServiceReasonServiceDisabled, "service_disabled":
		return "external_service 已停用，未执行投递。"
	case externalServiceReasonPluginDisabled, externalServiceReasonPluginArchived, externalServiceReasonPluginSoftUninstalled, "plugin_disabled":
		return "插件当前不可用，未执行 external_service 投递。"
	case externalServiceReasonCommunityDisabled, "community_plugin_disabled":
		return "当前子站未启用该插件，未执行 external_service 投递。"
	}
	msg := trimExternalServiceErr(err, responseBody)
	if msg != "" {
		return msg
	}
	if responseStatus > 0 {
		return http.StatusText(responseStatus)
	}
	return ""
}

func externalServiceDiagnosticSuggestion(endpoint, code string, err error) string {
	u, _ := url.Parse(strings.TrimSpace(endpoint))
	host := ""
	if u != nil {
		host = strings.ToLower(strings.TrimSpace(u.Hostname()))
	}
	switch strings.TrimSpace(code) {
	case "network_connection_refused":
		if host == "127.0.0.1" || host == "localhost" {
			return "DevHub 后端如果运行在 Docker 容器中，127.0.0.1 指向容器自身。若 Webhook 接收端运行在宿主机，请使用 host.docker.internal、Docker host gateway 地址，或宿主机局域网 IP。非 localhost HTTP 还需要 DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST 放行。"
		}
		return "请确认接收服务已启动、端口可达、防火墙未拦截，并确认该 HTTP origin 已在 DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST 中显式放行。"
	case "endpoint_safety_rejected":
		return externalServiceHTTPPolicyText()
	case "network_timeout":
		return "请确认 receiver 已启动、端口可达、Docker 网络连通，并能在 timeout_ms 内响应；本地联调可先访问 /health，再查看 receiver 日志。"
	case externalServiceReasonTokenMissing, "token_missing":
		return "请在 external_service 运行配置中写入新的 Bearer Token；接口和日志不会回显 token 明文。"
	case externalServiceReasonTokenInvalid, "token_invalid":
		return "请检查插件配置加密 key 是否可用，必要时重新写入 external_service Bearer Token；不会回显 token 明文。"
	case externalServiceReasonServiceDisabled, "service_disabled":
		return "请在 external_service 运行配置中启用投递后再执行健康检查或重试。"
	case externalServiceReasonPluginDisabled, externalServiceReasonPluginArchived, externalServiceReasonPluginSoftUninstalled, "plugin_disabled":
		return "请先启用或恢复插件，再执行 external_service 健康检查或投递。"
	case externalServiceReasonCommunityDisabled, "community_plugin_disabled":
		return "请在当前子站启用该插件，或切换到已启用插件的子站。"
	}
	return ""
}

func externalServiceHTTPPolicyText() string {
	return "HTTPS 默认允许；localhost / 127.0.0.1 / ::1 的 HTTP 默认允许；非 localhost 的 HTTP endpoint 默认不允许。生产建议使用 HTTPS；本地开发可通过 DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST 显式放行该 origin，例如 DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build。"
}

func externalServiceAllowlistExample() string {
	return "DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build"
}

func externalServiceHTTPPolicy() domain.PluginExternalServiceHTTPPolicy {
	return domain.PluginExternalServiceHTTPPolicy{
		HTTPSAllowed:               true,
		LocalhostHTTPAllowed:       true,
		NonLocalHTTPNeedsAllowlist: true,
		AllowlistEnv:               "DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST",
		AllowlistOrigins:           externalServiceAllowlistOrigins(),
	}
}

func (s *Service) externalServiceHTTPPolicy() domain.PluginExternalServiceHTTPPolicy {
	admin, _ := s.externalServiceAdminHTTPAllowlist()
	return buildExternalServiceHTTPAllowlistResponse(admin).Policy
}

func externalServiceAllowlistOrigins() []string {
	rawItems := splitExternalServiceAllowlist(os.Getenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST"))
	out := make([]string, 0, len(rawItems))
	seen := map[string]bool{}
	for _, item := range rawItems {
		origin := normalizeExternalServiceAllowlistItem(item)
		if origin == "" || seen[origin] {
			continue
		}
		seen[origin] = true
		out = append(out, origin)
	}
	return out
}

func (s *Service) enrichExternalServiceConfigDiagnostics(cfg domain.PluginExternalServiceConfig) domain.PluginExternalServiceConfig {
	cfg.HTTPPolicy = ptrExternalServiceHTTPPolicy(s.externalServiceHTTPPolicy())
	endpoint := strings.TrimSpace(cfg.EndpointURL)
	if endpoint == "" {
		return cfg
	}
	u, err := url.Parse(endpoint)
	if err == nil && u != nil {
		host := strings.ToLower(strings.TrimSpace(u.Hostname()))
		if strings.EqualFold(u.Scheme, "http") && (host == "127.0.0.1" || host == "localhost") && externalServiceRuntimeMayBeDocker() {
			cfg.Diagnostics = append(cfg.Diagnostics, domain.PluginExternalServiceDiagnostic{
				Type:        "docker_loopback_hint",
				Code:        "network_connection_refused",
				FailureType: "network_connection_refused",
				EndpointURL: endpoint,
				Message:     "DevHub 后端在容器中运行时，127.0.0.1 指向容器自身。",
				Suggestion:  "若 Webhook 接收端运行在宿主机，请使用 host.docker.internal、Docker host gateway 地址，或宿主机局域网 IP；host gateway HTTP 需要 DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST 放行。",
			})
		}
		if strings.EqualFold(u.Scheme, "http") && !allowedLocalEndpoint(u) && !s.allowedExternalServiceHTTPAllowlist(u) {
			cfg.Diagnostics = append(cfg.Diagnostics, externalServiceEndpointDiagnostic(endpoint, "endpoint_safety_rejected", nil))
		}
	}
	return cfg
}

func externalServiceEndpointDiagnostic(endpoint, code string, err error) domain.PluginExternalServiceDiagnostic {
	u, _ := url.Parse(strings.TrimSpace(endpoint))
	needsAllowlist := false
	safetyRejected := false
	if u != nil && strings.EqualFold(u.Scheme, "http") && !allowedLocalEndpoint(u) && !allowedExternalServiceHTTPAllowlist(u) {
		needsAllowlist = true
		safetyRejected = true
	}
	return domain.PluginExternalServiceDiagnostic{
		Type:             externalServiceFailureType(code),
		Code:             strings.TrimSpace(code),
		FailureType:      externalServiceFailureType(code),
		EndpointURL:      sanitizeExternalServiceEndpointForDiagnostics(endpoint),
		Message:          externalServiceDiagnosticMessage(endpoint, code, err, 0, ""),
		Suggestion:       externalServiceDiagnosticSuggestion(endpoint, code, err),
		SafetyRejected:   safetyRejected,
		NeedsAllowlist:   needsAllowlist,
		AllowlistEnv:     "DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST",
		AllowlistExample: externalServiceAllowlistExample(),
		AllowlistOrigins: externalServiceAllowlistOrigins(),
	}
}

func sanitizeExternalServiceEndpointForDiagnostics(endpoint string) string {
	cleaned := strings.TrimSpace(endpoint)
	u, err := url.Parse(cleaned)
	if err != nil || u == nil {
		return cleaned
	}
	if u.User != nil {
		u.User = url.User("[REDACTED]")
	}
	q := u.Query()
	for key := range q {
		if isSensitiveDiagnosticKey(key) {
			q.Set(key, "[REDACTED]")
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func isSensitiveDiagnosticKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return false
	}
	for _, part := range []string{"token", "secret", "authorization", "password", "credential"} {
		if strings.Contains(normalized, part) {
			return true
		}
	}
	return false
}

func externalServiceFailureType(code string) string {
	switch strings.TrimSpace(code) {
	case "endpoint_safety_rejected":
		return "endpoint_safety_rejected"
	case "network_connection_refused":
		return "network_connection_refused"
	case "network_timeout":
		return "network_timeout"
	case "http_status_failed":
		return "http_status_failed"
	case externalServiceReasonTokenMissing, "token_missing":
		return "token_missing"
	case externalServiceReasonTokenInvalid, "token_invalid":
		return "token_invalid"
	case externalServiceReasonServiceDisabled, "service_disabled":
		return "service_disabled"
	case externalServiceReasonPluginDisabled, externalServiceReasonPluginArchived, externalServiceReasonPluginSoftUninstalled, "plugin_disabled":
		return "plugin_disabled"
	case externalServiceReasonCommunityDisabled, "community_plugin_disabled":
		return "community_plugin_disabled"
	default:
		return strings.TrimSpace(code)
	}
}

func ptrExternalServiceHTTPPolicy(v domain.PluginExternalServiceHTTPPolicy) *domain.PluginExternalServiceHTTPPolicy {
	return &v
}

func externalServiceRuntimeMayBeDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	raw, err := os.ReadFile("/proc/1/cgroup")
	if err != nil {
		return false
	}
	text := strings.ToLower(string(raw))
	return strings.Contains(text, "docker") || strings.Contains(text, "containerd") || strings.Contains(text, "kubepods")
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func splitExternalServiceAllowlist(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
}

func normalizeExternalServiceAllowlistItem(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || strings.ToLower(u.Scheme) != "http" {
		return ""
	}
	return normalizeExternalServiceEndpointOrigin(u)
}

func normalizeExternalServiceEndpointOrigin(u *url.URL) string {
	if u == nil {
		return ""
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" {
		return ""
	}
	port := strings.TrimSpace(u.Port())
	if port == "" {
		port = "80"
	}
	return "http://" + net.JoinHostPort(host, port)
}

func encryptExternalServiceToken(token string) (string, string, error) {
	kr, ok, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil {
		return "", "", err
	}
	if !ok || kr == nil {
		return "", "", errors.New("缺少外部服务 token 加密密钥")
	}
	keyID, key := kr.CurrentKey()
	sum := sha256.Sum256([]byte(token))
	enc, err := pluginregistry.EncryptStringV2(keyID, key, token)
	if err != nil {
		return "", "", err
	}
	return enc, hex.EncodeToString(sum[:]), nil
}

func decryptExternalServiceToken(ciphertext string) (string, error) {
	kr, ok, err := pluginregistry.LoadPluginConfigKeyring()
	if err != nil {
		return "", err
	}
	if !ok || kr == nil {
		return "", errors.New("缺少外部服务 token 解密密钥")
	}
	plain, _, err := pluginregistry.DecryptStringWithKeyring(kr, ciphertext)
	return plain, err
}

func trimExternalServiceErr(err error, body string) string {
	if err != nil {
		msg := strings.TrimSpace(err.Error())
		if len(msg) > 500 {
			msg = msg[:500]
		}
		return msg
	}
	msg := strings.TrimSpace(body)
	if len(msg) > 500 {
		msg = msg[:500]
	}
	return msg
}

func trimExternalServiceExcerpt(body string) string {
	body = strings.TrimSpace(body)
	if len(body) > 500 {
		body = body[:500]
	}
	return body
}
