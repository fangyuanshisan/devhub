package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
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
	return s.repo.PluginExternalServiceConfig(pluginCode)
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
	if err := validateExternalServiceUpdateRequest(req, current); err != nil {
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
		if tokenRef == "" {
			tokenRef = "extsvc_" + randomHex(12)
		}
		enc, hash, err := encryptExternalServiceToken(strings.TrimSpace(req.Token))
		if err != nil {
			return domain.PluginExternalServiceConfig{}, err
		}
		tokenCiphertext = enc
		tokenHash = hash
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
			"failure_policy": saved.FailurePolicy,
		}),
		CreatedAt: Now(),
	})
	return saved, nil
}

func (s *Service) RunPluginExternalServiceHealthCheck(operator PluginExternalServiceOperator, pluginCode string) (domain.PluginExternalServiceHealthCheckResponse, error) {
	cfg, ok := s.PluginExternalServiceConfig(pluginCode)
	if !ok {
		return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_not_found", "external_service 配置不存在").WithStatus(http.StatusNotFound)
	}
	if !cfg.Enabled {
		return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_disabled", "external_service 已停用").WithStatus(http.StatusBadRequest)
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
	if cfg.AuthType == externalServiceAuthBearer && cfg.TokenCiphertext != "" {
		token, derr := decryptExternalServiceToken(cfg.TokenCiphertext)
		if derr != nil {
			return domain.PluginExternalServiceHealthCheckResponse{}, domain.NewPluginError("plugin_external_service_token_invalid", "external_service token 无法解密").WithStatus(http.StatusBadRequest)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	duration := time.Since(started).Milliseconds()
	if err != nil {
		return s.recordExternalServiceHealth(operator, pluginCode, cfg, "timeout", 0, "", "timeout", duration, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	status := externalServiceHealthHealthy
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		status = externalServiceHealthHealthy
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		status = externalServiceHealthWarning
	default:
		status = externalServiceHealthError
	}
	return s.recordExternalServiceHealth(operator, pluginCode, cfg, status, resp.StatusCode, string(body), "", duration, nil)
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
	updated.LastErrorMessage = trimExternalServiceErr(cause, responseBody)
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
	exec := domain.HookExecution{
		PluginCode:          pluginCode,
		ServiceType:         externalServiceTypeExternal,
		EndpointURL:         saved.EndpointURL,
		HookName:            "external_service.health_check",
		Mode:                string(pluginregistry.HookNonBlocking),
		StartedAt:           now,
		FinishedAt:          now,
		DurationMS:          int(durationMS),
		Success:             success,
		ErrorMessage:        saved.LastErrorMessage,
		Status:              execStatus,
		ResponseStatus:      responseStatus,
		ResponseBodyExcerpt: trimExternalServiceExcerpt(responseBody),
		ErrorCode:           errorCode,
		Metadata:            mustJSON(map[string]any{"plugin_code": pluginCode, "service_type": externalServiceTypeExternal, "failure_policy": saved.FailurePolicy, "auth_type": saved.AuthType, "health_status_before": cfg.LastHealthStatus, "health_status_after": saved.LastHealthStatus}),
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
	return domain.PluginExternalServiceHealthCheckResponse{
		PluginCode:   pluginCode,
		ServiceType:  saved.ServiceType,
		EndpointURL:  saved.EndpointURL,
		HealthStatus: saved.LastHealthStatus,
		Status:       saved.Status,
		CheckedAt:    saved.LastCheckedAt,
		DurationMS:   durationMS,
		Message:      saved.LastErrorMessage,
	}, nil
}

func validateExternalServiceUpdateRequest(req domain.PluginExternalServiceUpdateRequest, current domain.PluginExternalServiceConfig) error {
	endpoint := strings.TrimSpace(req.EndpointURL)
	if endpoint == "" {
		return domain.NewPluginError("plugin_external_service_invalid_endpoint", "external_service.endpoint_url 不能为空").WithStatus(http.StatusBadRequest)
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return domain.NewPluginError("plugin_external_service_invalid_endpoint", "external_service.endpoint_url 必须是合法 URL").WithStatus(http.StatusBadRequest)
	}
	if u.Scheme != "https" && !allowedLocalEndpoint(u) {
		return domain.NewPluginError("plugin_external_service_invalid_endpoint", "生产环境建议使用 https；仅本地开发允许 http://localhost 或 127.0.0.1").WithStatus(http.StatusBadRequest)
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
	if authType == externalServiceAuthBearer && strings.TrimSpace(req.Token) == "" && strings.TrimSpace(current.TokenCiphertext) == "" {
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
