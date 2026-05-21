package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const (
	externalServiceHookStatusPending        = "pending"
	externalServiceHookStatusRunning        = "running"
	externalServiceHookStatusSuccess        = "success"
	externalServiceHookStatusFailed         = "failed"
	externalServiceHookStatusTimeout        = "timeout"
	externalServiceHookStatusRetryScheduled = "retry_scheduled"
	externalServiceHookStatusRetryExhausted = "retry_exhausted"
	externalServiceHookStatusSkipped        = "skipped"

	externalServiceReasonPluginDisabled        = "PLUGIN_DISABLED"
	externalServiceReasonPluginArchived        = "PLUGIN_ARCHIVED"
	externalServiceReasonPluginSoftUninstalled = "PLUGIN_SOFT_UNINSTALLED"
	externalServiceReasonCommunityDisabled     = "COMMUNITY_PLUGIN_DISABLED"
	externalServiceReasonServiceDisabled       = "EXTERNAL_SERVICE_DISABLED"
	externalServiceReasonEndpointMissing       = "ENDPOINT_MISSING"
	externalServiceReasonEndpointInvalid       = "ENDPOINT_INVALID"
	externalServiceReasonTokenMissing          = "TOKEN_MISSING"
	externalServiceReasonTokenInvalid          = "TOKEN_INVALID"
	externalServiceReasonHookPaused            = "HOOK_DISABLED_BY_FAILURE_POLICY"
)

func (s *Service) enqueueExternalServiceHookDeliveries(event pluginregistry.HookEvent) {
	if s == nil || s.repo == nil {
		return
	}
	code := strings.TrimSpace(event.Ctx.PluginCode)
	if code == "" || code == pluginregistry.CoreCode {
		return
	}
	plugin, ok := s.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return
	}
	for _, hook := range plugin.Hooks {
		if !externalServiceHookMatches(event, hook) {
			continue
		}
		s.enqueueExternalServiceHookDelivery(event, plugin, hook)
	}
}

func externalServiceHookMatches(event pluginregistry.HookEvent, hook domain.HookDefinition) bool {
	if strings.TrimSpace(hook.Name) != strings.TrimSpace(event.Name) {
		return false
	}
	if strings.TrimSpace(hook.ServiceType) != externalServiceTypeExternal {
		return false
	}
	if hook.Enabled != nil && !*hook.Enabled {
		return false
	}
	mode := strings.TrimSpace(hook.Mode)
	return mode == "" || mode == string(pluginregistry.HookNonBlocking)
}

func (s *Service) enqueueExternalServiceHookDelivery(event pluginregistry.HookEvent, plugin domain.Plugin, hook domain.HookDefinition) {
	executionID := "extsvc_exec_" + randomHex(12)
	eventID := "evt_" + randomHex(12)
	idempotencyKey := strings.TrimSpace(event.Ctx.RequestID)
	if idempotencyKey == "" {
		idempotencyKey = executionID
	}
	cfg, ok := s.repo.PluginExternalServiceConfig(plugin.Code)
	now := time.Now()
	if reason := s.externalServiceDeliverySkipReason(event, plugin, hook, cfg, ok); reason != "" {
		s.appendExternalServiceHookExecution(event, plugin, hook, cfg, externalServiceHookExecutionRecord{
			ExecutionID:    executionID,
			EventID:        eventID,
			IdempotencyKey: idempotencyKey,
			Status:         externalServiceHookStatusSkipped,
			ErrorCode:      reason,
			ErrorMessage:   externalServiceReasonText(reason),
			StartedAt:      now,
			FinishedAt:     now,
			HealthBefore:   cfg.LastHealthStatus,
			HealthAfter:    cfg.LastHealthStatus,
		})
		return
	}
	pending := s.appendExternalServiceHookExecution(event, plugin, hook, cfg, externalServiceHookExecutionRecord{
		ExecutionID:    executionID,
		EventID:        eventID,
		IdempotencyKey: idempotencyKey,
		Status:         externalServiceHookStatusPending,
		Attempt:        0,
		MaxAttempts:    externalServiceHookMaxAttempts(hook),
		StartedAt:      now,
		FinishedAt:     now,
		HealthBefore:   cfg.LastHealthStatus,
		HealthAfter:    cfg.LastHealthStatus,
	})
	go s.deliverExternalServiceHook(event, plugin, hook, cfg, pending, executionID, eventID, idempotencyKey)
}

type externalServiceHookExecutionRecord struct {
	ExecutionID         string
	EventID             string
	IdempotencyKey      string
	Status              string
	ResponseStatus      int
	ErrorCode           string
	ErrorMessage        string
	ResponseBodyExcerpt string
	RequestBodySHA256   string
	Attempt             int
	MaxAttempts         int
	NextRetryAt         string
	StartedAt           time.Time
	FinishedAt          time.Time
	DurationMS          int
	HealthBefore        string
	HealthAfter         string
	ExtraMetadata       map[string]any
}

func (s *Service) deliverExternalServiceHook(event pluginregistry.HookEvent, plugin domain.Plugin, hook domain.HookDefinition, cfg domain.PluginExternalServiceConfig, pending domain.HookExecution, executionID, eventID, idempotencyKey string) {
	maxAttempts := externalServiceHookMaxAttempts(hook)
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		started := time.Now()
		s.appendExternalServiceHookExecution(event, plugin, hook, cfg, externalServiceHookExecutionRecord{
			ExecutionID:    executionID,
			EventID:        eventID,
			IdempotencyKey: idempotencyKey,
			Status:         externalServiceHookStatusRunning,
			Attempt:        attempt,
			MaxAttempts:    maxAttempts,
			StartedAt:      started,
			FinishedAt:     started,
			HealthBefore:   cfg.LastHealthStatus,
			HealthAfter:    cfg.LastHealthStatus,
		})
		result := s.performExternalServiceHookAttempt(event, plugin, hook, cfg, executionID, eventID, idempotencyKey)
		result.Attempt = attempt
		result.MaxAttempts = maxAttempts
		result.HealthBefore = cfg.LastHealthStatus
		if result.Status == externalServiceHookStatusSuccess {
			saved := s.updateExternalServiceHealthFromDelivery(plugin.Code, cfg, true, result.ErrorCode, result.ErrorMessage)
			result.HealthAfter = saved.LastHealthStatus
			s.appendExternalServiceHookExecution(event, plugin, hook, saved, result)
			s.auditExternalServiceDelivery(event, saved, result)
			return
		}
		if externalServiceShouldRetry(result.Status, result.ResponseStatus, result.ErrorCode) && attempt < maxAttempts {
			result.Status = externalServiceHookStatusRetryScheduled
			result.NextRetryAt = formatHookTime(time.Now().Add(externalServiceRetryDelay(attempt)))
			saved := s.updateExternalServiceHealthFromDelivery(plugin.Code, cfg, false, result.ErrorCode, result.ErrorMessage)
			result.HealthAfter = saved.LastHealthStatus
			s.appendExternalServiceHookExecution(event, plugin, hook, saved, result)
			s.auditExternalServiceDelivery(event, saved, result)
			cfg = saved
			time.Sleep(externalServiceRetryDelay(attempt))
			continue
		}
		finalStatus := result.Status
		if externalServiceShouldRetry(result.Status, result.ResponseStatus, result.ErrorCode) && attempt >= maxAttempts {
			finalStatus = externalServiceHookStatusRetryExhausted
		}
		result.Status = finalStatus
		saved := s.updateExternalServiceHealthFromDelivery(plugin.Code, cfg, false, result.ErrorCode, result.ErrorMessage)
		result.HealthAfter = saved.LastHealthStatus
		s.appendExternalServiceHookExecution(event, plugin, hook, saved, result)
		s.auditExternalServiceDelivery(event, saved, result)
		return
	}
	_ = pending
}

func (s *Service) performExternalServiceHookAttempt(event pluginregistry.HookEvent, plugin domain.Plugin, hook domain.HookDefinition, cfg domain.PluginExternalServiceConfig, executionID, eventID, idempotencyKey string) externalServiceHookExecutionRecord {
	started := time.Now()
	reqURL, _, err := externalServiceHookURL(cfg.EndpointURL, hook.Path)
	if err != nil {
		return externalServiceHookExecutionRecord{ExecutionID: executionID, EventID: eventID, IdempotencyKey: idempotencyKey, Status: externalServiceHookStatusFailed, ErrorCode: externalServiceReasonEndpointInvalid, ErrorMessage: "外部服务地址或 Hook 路径不合法", StartedAt: started, FinishedAt: time.Now()}
	}
	payload := externalServiceHookPayload(event, plugin.Code, hook.Name, executionID, eventID)
	body, _ := json.Marshal(payload)
	sum := sha256.Sum256(body)
	timeout := externalServiceHookTimeout(hook, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return externalServiceHookExecutionRecord{ExecutionID: executionID, EventID: eventID, IdempotencyKey: idempotencyKey, Status: externalServiceHookStatusFailed, ErrorCode: externalServiceReasonEndpointInvalid, ErrorMessage: "外部服务请求创建失败", StartedAt: started, FinishedAt: time.Now(), RequestBodySHA256: hex.EncodeToString(sum[:])}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DevHub-Plugin-Code", plugin.Code)
	req.Header.Set("X-DevHub-Hook-Name", hook.Name)
	req.Header.Set("X-DevHub-Execution-ID", executionID)
	req.Header.Set("X-DevHub-Event-ID", eventID)
	req.Header.Set("X-DevHub-Request-ID", event.Ctx.RequestID)
	req.Header.Set("X-DevHub-Idempotency-Key", idempotencyKey)
	req.Header.Set("X-DevHub-Timestamp", time.Now().UTC().Format(time.RFC3339))
	req.Header.Set("X-DevHub-Delivery-Mode", string(pluginregistry.HookNonBlocking))
	if normalizeExternalServiceAuthType(cfg.AuthType) == externalServiceAuthBearer {
		token, tokenErr := decryptExternalServiceToken(cfg.TokenCiphertext)
		if tokenErr != nil || strings.TrimSpace(token) == "" {
			code := externalServiceReasonTokenInvalid
			if strings.TrimSpace(cfg.TokenCiphertext) == "" {
				code = externalServiceReasonTokenMissing
			}
			finished := time.Now()
			return externalServiceHookExecutionRecord{ExecutionID: executionID, EventID: eventID, IdempotencyKey: idempotencyKey, Status: externalServiceHookStatusFailed, ErrorCode: code, ErrorMessage: externalServiceReasonText(code), StartedAt: started, FinishedAt: finished, DurationMS: int(finished.Sub(started).Milliseconds()), RequestBodySHA256: hex.EncodeToString(sum[:])}
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	finished := time.Now()
	durationMS := int(finished.Sub(started).Milliseconds())
	base := externalServiceHookExecutionRecord{ExecutionID: executionID, EventID: eventID, IdempotencyKey: idempotencyKey, StartedAt: started, FinishedAt: finished, DurationMS: durationMS, RequestBodySHA256: hex.EncodeToString(sum[:])}
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			base.Status = externalServiceHookStatusTimeout
			base.ErrorCode = "TIMEOUT"
			base.ErrorMessage = "外部服务请求超时"
			return base
		}
		base.Status = externalServiceHookStatusFailed
		base.ErrorCode = "REQUEST_FAILED"
		base.ErrorMessage = trimExternalServiceErr(err, "")
		return base
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	base.ResponseStatus = resp.StatusCode
	base.ResponseBodyExcerpt = trimExternalServiceExcerpt(string(raw))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		base.Status = externalServiceHookStatusSuccess
		return base
	}
	base.Status = externalServiceHookStatusFailed
	base.ErrorCode = "HTTP_" + strconv.Itoa(resp.StatusCode)
	base.ErrorMessage = externalServiceHTTPStatusText(resp.StatusCode)
	if base.ErrorMessage == "" {
		base.ErrorMessage = trimExternalServiceErr(nil, string(raw))
	}
	return base
}

func (s *Service) externalServiceDeliverySkipReason(event pluginregistry.HookEvent, plugin domain.Plugin, hook domain.HookDefinition, cfg domain.PluginExternalServiceConfig, hasConfig bool) string {
	switch strings.TrimSpace(plugin.Status) {
	case pluginregistry.StatusArchived:
		return externalServiceReasonPluginArchived
	case "soft_uninstalled":
		return externalServiceReasonPluginSoftUninstalled
	case pluginregistry.StatusDisabled:
		return externalServiceReasonPluginDisabled
	}
	if plugin.Status != pluginregistry.StatusEnabled {
		return externalServiceReasonPluginDisabled
	}
	if event.Ctx.CommunityID > 0 && !s.IsPluginEnabledForCommunity(event.Ctx.CommunityID, plugin.Code) {
		return externalServiceReasonCommunityDisabled
	}
	if !hasConfig || !cfg.Enabled {
		return externalServiceReasonServiceDisabled
	}
	if strings.TrimSpace(cfg.EndpointURL) == "" {
		return externalServiceReasonEndpointMissing
	}
	if normalizeExternalServiceFailurePolicy(firstNonEmpty(hook.FailurePolicy, cfg.FailurePolicy)) == externalServiceFailurePolicyDisableHook && (cfg.LastHealthStatus == externalServiceHealthError || cfg.Status == externalServiceHealthError) {
		return externalServiceReasonHookPaused
	}
	return ""
}

func (s *Service) appendExternalServiceHookExecution(event pluginregistry.HookEvent, plugin domain.Plugin, hook domain.HookDefinition, cfg domain.PluginExternalServiceConfig, rec externalServiceHookExecutionRecord) domain.HookExecution {
	now := time.Now()
	if rec.StartedAt.IsZero() {
		rec.StartedAt = now
	}
	if rec.FinishedAt.IsZero() {
		rec.FinishedAt = rec.StartedAt
	}
	if rec.MaxAttempts <= 0 {
		rec.MaxAttempts = externalServiceHookMaxAttempts(hook)
	}
	metadata := map[string]any{
		"execution_id":         rec.ExecutionID,
		"event_id":             rec.EventID,
		"idempotency_key":      rec.IdempotencyKey,
		"request_path":         normalizeExternalServiceHookPath(hook.Path),
		"attempt":              rec.Attempt,
		"max_attempts":         rec.MaxAttempts,
		"next_retry_at":        rec.NextRetryAt,
		"failure_policy":       externalServiceHookFailurePolicy(hook, cfg),
		"health_status_before": rec.HealthBefore,
		"health_status_after":  rec.HealthAfter,
	}
	for k, v := range rec.ExtraMetadata {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		metadata[key] = scrubAnyForSnapshot(v)
	}
	record := domain.HookExecution{
		HookName:            hook.Name,
		PluginCode:          plugin.Code,
		ServiceType:         externalServiceTypeExternal,
		EndpointURL:         cfg.EndpointURL,
		Mode:                string(pluginregistry.HookNonBlocking),
		ContentType:         event.Ctx.ContentType,
		ContentID:           event.Ctx.ContentID,
		CommunityID:         event.Ctx.CommunityID,
		CategoryID:          event.Ctx.CategoryID,
		ActorType:           string(event.Ctx.ActorType),
		ActorID:             event.Ctx.ActorID,
		UserID:              event.Ctx.UserID,
		AdminUserID:         event.Ctx.AdminUserID,
		RequestID:           event.Ctx.RequestID,
		StartedAt:           formatHookTime(rec.StartedAt),
		FinishedAt:          formatHookTime(rec.FinishedAt),
		DurationMS:          rec.DurationMS,
		Success:             rec.Status == externalServiceHookStatusSuccess,
		ErrorMessage:        rec.ErrorMessage,
		Blocking:            false,
		Status:              rec.Status,
		ResponseStatus:      rec.ResponseStatus,
		ResponseBodyExcerpt: rec.ResponseBodyExcerpt,
		RequestBodySHA256:   rec.RequestBodySHA256,
		ErrorCode:           rec.ErrorCode,
		Metadata:            mustJSON(metadata),
	}
	saved, err := s.repo.AppendHookExecution(record)
	if err != nil {
		return record
	}
	if rec.Status == externalServiceHookStatusSkipped {
		s.auditExternalServiceDelivery(event, cfg, rec)
	}
	return saved
}

func (s *Service) updateExternalServiceHealthFromDelivery(pluginCode string, base domain.PluginExternalServiceConfig, success bool, errorCode, errorMessage string) domain.PluginExternalServiceConfig {
	current, ok := s.repo.PluginExternalServiceConfig(pluginCode)
	if ok {
		base = current
	}
	before := base.LastHealthStatus
	now := Now()
	updated := base
	policy := normalizeExternalServiceFailurePolicy(updated.FailurePolicy)
	if success {
		updated.FailureCount = 0
		updated.LastHealthStatus = externalServiceHealthHealthy
		updated.Status = externalServiceHealthHealthy
		updated.LastSuccessAt = now
		updated.LastErrorMessage = ""
	} else {
		if policy != externalServiceFailurePolicyIgnore {
			updated.FailureCount++
			if updated.FailureCount >= normalizeExternalServiceThreshold(updated.ErrorThreshold, 5) && (policy == externalServiceFailurePolicyError || policy == externalServiceFailurePolicyDisableHook) {
				updated.LastHealthStatus = externalServiceHealthError
			} else if updated.FailureCount >= normalizeExternalServiceThreshold(updated.WarningThreshold, 3) {
				updated.LastHealthStatus = externalServiceHealthWarning
			}
			updated.Status = updated.LastHealthStatus
		}
		updated.LastFailureAt = now
		updated.LastErrorMessage = firstNonEmpty(errorMessage, errorCode)
	}
	updated.LastCheckedAt = now
	if !updated.Enabled {
		updated.Status = externalServiceHealthDisabled
		updated.LastHealthStatus = externalServiceHealthDisabled
	}
	saved, err := s.repo.SavePluginExternalServiceConfig(updated)
	if err != nil {
		return updated
	}
	if before != saved.LastHealthStatus {
		action := "plugin.external_service.health.recovered"
		switch saved.LastHealthStatus {
		case externalServiceHealthWarning:
			action = "plugin.external_service.health.warning"
		case externalServiceHealthError:
			action = "plugin.external_service.health.error"
		}
		s.repo.AppendAdminLog(domain.AdminLog{
			Site:      "admin",
			Type:      "system",
			Actor:     "system",
			ActorType: "system",
			Action:    action,
			Target:    "plugins#" + pluginCode,
			Metadata:  mustJSON(map[string]any{"plugin_code": pluginCode, "health_status_before": before, "health_status_after": saved.LastHealthStatus, "failure_count": saved.FailureCount}),
			CreatedAt: now,
		})
	}
	return saved
}

func (s *Service) auditExternalServiceDelivery(event pluginregistry.HookEvent, cfg domain.PluginExternalServiceConfig, rec externalServiceHookExecutionRecord) {
	action := "plugin.external_service.delivery." + rec.Status
	if rec.Status == externalServiceHookStatusSuccess {
		action = "plugin.external_service.delivery.success"
	}
	if rec.Status == externalServiceHookStatusFailed {
		action = "plugin.external_service.delivery.failed"
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Type:      "system",
		Actor:     firstNonEmpty(string(event.Ctx.ActorType), "system"),
		ActorType: firstNonEmpty(string(event.Ctx.ActorType), "system"),
		ActorID:   event.Ctx.ActorID,
		Action:    action,
		Target:    fmt.Sprintf("external_service#%s:%s", event.Ctx.PluginCode, event.Name),
		Metadata: mustJSON(map[string]any{
			"plugin_code":          event.Ctx.PluginCode,
			"hook_name":            event.Name,
			"execution_id":         rec.ExecutionID,
			"endpoint_url":         cfg.EndpointURL,
			"status":               rec.Status,
			"response_status":      rec.ResponseStatus,
			"attempt":              rec.Attempt,
			"max_attempts":         rec.MaxAttempts,
			"failure_policy":       cfg.FailurePolicy,
			"health_status_before": rec.HealthBefore,
			"health_status_after":  rec.HealthAfter,
			"error_code":           rec.ErrorCode,
			"request_id":           event.Ctx.RequestID,
		}),
		CreatedAt: Now(),
	})
}

func externalServiceHookPayload(event pluginregistry.HookEvent, pluginCode, hookName, executionID, eventID string) map[string]any {
	resourceType := firstNonEmpty(event.Ctx.ContentType, "unknown")
	resourceID := event.Ctx.ContentID
	data := map[string]any{}
	if event.Topic != nil {
		data["topic_id"] = event.Topic.ID
		data["title"] = event.Topic.Title
		data["content_type"] = event.Topic.ContentType
		data["status"] = event.Topic.Status
	}
	if event.Comment != nil {
		data["comment_id"] = event.Comment.ID
		data["topic_id"] = event.Comment.TopicID
	}
	if event.Ctx.Metadata != nil {
		data["metadata"] = scrubAnyForSnapshot(event.Ctx.Metadata)
	}
	return map[string]any{
		"schema_version": "1",
		"plugin_code":    pluginCode,
		"hook_name":      hookName,
		"event_id":       eventID,
		"execution_id":   executionID,
		"resource_type":  resourceType,
		"resource_id":    resourceID,
		"community_id":   event.Ctx.CommunityID,
		"actor": map[string]any{
			"type": string(event.Ctx.ActorType),
			"id":   event.Ctx.ActorID,
		},
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"data":        data,
	}
}

func externalServiceHookURL(endpoint, hookPath string) (string, string, error) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", "", errors.New("invalid endpoint")
	}
	p := normalizeExternalServiceHookPath(hookPath)
	if p == "" {
		return "", "", errors.New("invalid hook path")
	}
	u.Path = path.Join(strings.TrimRight(u.Path, "/"), p)
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), p, nil
}

func normalizeExternalServiceHookPath(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || strings.Contains(v, "://") || strings.Contains(v, "\\") || strings.Contains(v, "..") {
		return ""
	}
	if !strings.HasPrefix(v, "/") {
		v = "/" + v
	}
	cleaned := path.Clean(v)
	if cleaned == "." || cleaned == "/" || strings.Contains(cleaned, "..") {
		return ""
	}
	return cleaned
}

func externalServiceHookMaxAttempts(hook domain.HookDefinition) int {
	if !hook.RetryEnabled {
		return 1
	}
	if hook.MaxAttempts <= 0 {
		return 3
	}
	if hook.MaxAttempts > 5 {
		return 5
	}
	return hook.MaxAttempts
}

func externalServiceHookTimeout(hook domain.HookDefinition, cfg domain.PluginExternalServiceConfig) int {
	if hook.TimeoutMS > 0 {
		return normalizeExternalServiceTimeout(hook.TimeoutMS)
	}
	return normalizeExternalServiceTimeout(cfg.TimeoutMS)
}

func externalServiceHookFailurePolicy(hook domain.HookDefinition, cfg domain.PluginExternalServiceConfig) string {
	policy := normalizeExternalServiceFailurePolicy(hook.FailurePolicy)
	if policy == "" {
		policy = normalizeExternalServiceFailurePolicy(cfg.FailurePolicy)
	}
	return policy
}

func externalServiceShouldRetry(status string, responseStatus int, errorCode string) bool {
	if errorCode == externalServiceReasonTokenMissing || errorCode == externalServiceReasonTokenInvalid || errorCode == externalServiceReasonEndpointInvalid {
		return false
	}
	if status == externalServiceHookStatusTimeout {
		return true
	}
	if responseStatus == http.StatusTooManyRequests {
		return true
	}
	return responseStatus >= 500
}

func externalServiceRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 20 * time.Millisecond
	}
	return time.Duration(attempt*20) * time.Millisecond
}

func externalServiceHTTPStatusText(status int) string {
	switch {
	case status == http.StatusTooManyRequests:
		return "外部服务请求过多，等待重试"
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return "外部服务认证失败"
	case status >= 400 && status < 500:
		return "外部服务返回客户端错误"
	case status >= 500:
		return "外部服务异常"
	default:
		return ""
	}
}

func externalServiceReasonText(reason string) string {
	switch reason {
	case externalServiceReasonPluginDisabled:
		return "当前插件未启用，已跳过外部服务投递"
	case externalServiceReasonPluginArchived:
		return "当前插件已归档，已跳过外部服务投递"
	case externalServiceReasonPluginSoftUninstalled:
		return "当前插件已软卸载，已跳过外部服务投递"
	case externalServiceReasonCommunityDisabled:
		return "当前子站未启用该插件，已跳过外部服务投递"
	case externalServiceReasonServiceDisabled:
		return "external_service 已停用，已跳过投递"
	case externalServiceReasonEndpointMissing:
		return "外部服务地址缺失"
	case externalServiceReasonEndpointInvalid:
		return "外部服务地址不合法"
	case externalServiceReasonTokenMissing:
		return "external_service bearer token 缺失"
	case externalServiceReasonTokenInvalid:
		return "external_service bearer token 无法解密"
	case externalServiceReasonHookPaused:
		return "该 Hook 已因连续失败被暂停"
	default:
		return "外部服务投递已跳过"
	}
}

func (s *Service) ManualRetryExternalServiceHookExecution(operator PluginExternalServiceOperator, pluginCode string, executionRecordID int64) (domain.PluginExternalServiceManualRetryResponse, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" || executionRecordID <= 0 {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_webhook_retry_invalid", "重试参数不合法").WithStatus(http.StatusBadRequest)
	}
	source, ok := s.repo.HookExecutionByID(executionRecordID)
	if !ok || source.ID == 0 {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_hook_execution_not_found", "Hook 执行记录不存在").WithStatus(http.StatusNotFound)
	}
	if strings.TrimSpace(source.PluginCode) != pluginCode {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_webhook_retry_cross_plugin", "不能跨插件重试 Hook 执行记录").WithStatus(http.StatusBadRequest).
			WithDetail("plugin_code", pluginCode).
			WithDetail("source_plugin_code", source.PluginCode)
	}
	if strings.TrimSpace(source.ServiceType) != externalServiceTypeExternal {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_webhook_retry_not_external_service", "只允许重试 external_service Hook 执行记录").WithStatus(http.StatusBadRequest)
	}
	if source.Blocking || strings.TrimSpace(source.Mode) == string(pluginregistry.HookBlocking) {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_webhook_retry_not_allowed", "internal/builtin Hook 不能手动重试").WithStatus(http.StatusBadRequest)
	}
	if !externalServiceManualRetryAllowed(source.Status, source.Success) {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_webhook_retry_status_not_allowed", "当前 Hook 执行状态不允许手动重试").WithStatus(http.StatusBadRequest).
			WithDetail("status", source.Status)
	}
	if source.HookName == "external_service.health_check" {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_webhook_retry_health_check_not_allowed", "健康检查记录不能作为 Webhook 投递重试").WithStatus(http.StatusBadRequest)
	}
	plugin, ok := s.PluginByCode(pluginCode)
	if !ok || plugin.Code == "" {
		return domain.PluginExternalServiceManualRetryResponse{}, pluginNotFound(pluginCode)
	}
	switch strings.TrimSpace(plugin.Status) {
	case pluginregistry.StatusArchived:
		return domain.PluginExternalServiceManualRetryResponse{}, pluginArchived(pluginCode)
	case pluginregistry.StatusDisabled, "soft_uninstalled":
		return domain.PluginExternalServiceManualRetryResponse{}, pluginDisabled(pluginCode)
	}
	if plugin.Status != pluginregistry.StatusEnabled {
		return domain.PluginExternalServiceManualRetryResponse{}, pluginDisabled(pluginCode)
	}
	cfg, hasConfig := s.repo.PluginExternalServiceConfig(pluginCode)
	if !hasConfig || !cfg.Enabled {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_external_service_disabled", "external_service 已停用，不能手动重试").WithStatus(http.StatusBadRequest)
	}
	hook, foundHook := findExternalServiceHook(plugin, source.HookName)
	if !foundHook {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_webhook_retry_hook_not_found", "未找到可重试的 external_service Hook 声明").WithStatus(http.StatusBadRequest).
			WithDetail("hook_name", source.HookName)
	}
	event := hookEventFromExecution(source, pluginCode)
	if reason := s.externalServiceDeliverySkipReason(event, plugin, hook, cfg, hasConfig); reason != "" {
		return domain.PluginExternalServiceManualRetryResponse{}, domain.NewPluginError("plugin_webhook_retry_skipped", externalServiceReasonText(reason)).WithStatus(http.StatusBadRequest).
			WithDetail("reason", reason)
	}

	executionID := "extsvc_retry_" + randomHex(12)
	eventID := "evt_retry_" + randomHex(12)
	idempotencyKey := executionID
	if source.RequestID != "" {
		idempotencyKey = source.RequestID + ":manual_retry:" + executionID
	}
	started := time.Now()
	s.appendExternalServiceHookExecution(event, plugin, hook, cfg, externalServiceHookExecutionRecord{
		ExecutionID:    executionID,
		EventID:        eventID,
		IdempotencyKey: idempotencyKey,
		Status:         externalServiceHookStatusRunning,
		Attempt:        1,
		MaxAttempts:    1,
		StartedAt:      started,
		FinishedAt:     started,
		HealthBefore:   cfg.LastHealthStatus,
		HealthAfter:    cfg.LastHealthStatus,
		ExtraMetadata:  manualRetryMetadata(operator, source),
	})
	result := s.performExternalServiceHookAttempt(event, plugin, hook, cfg, executionID, eventID, idempotencyKey)
	result.Attempt = 1
	result.MaxAttempts = 1
	result.HealthBefore = cfg.LastHealthStatus
	result.ExtraMetadata = manualRetryMetadata(operator, source)
	success := result.Status == externalServiceHookStatusSuccess
	if !success && externalServiceShouldRetry(result.Status, result.ResponseStatus, result.ErrorCode) {
		result.Status = externalServiceHookStatusRetryExhausted
	}
	savedCfg := s.updateExternalServiceHealthFromDelivery(plugin.Code, cfg, success, result.ErrorCode, result.ErrorMessage)
	result.HealthAfter = savedCfg.LastHealthStatus
	savedExecution := s.appendExternalServiceHookExecution(event, plugin, hook, savedCfg, result)
	s.auditExternalServiceDelivery(event, savedCfg, result)
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Type:      "system",
		Actor:     firstNonEmpty(operator.Name, "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.webhook.manual_retry",
		Target:    fmt.Sprintf("external_service#%s:%s", pluginCode, source.HookName),
		Metadata: mustJSON(map[string]any{
			"plugin_code":         pluginCode,
			"hook_name":           source.HookName,
			"source_record_id":    source.ID,
			"source_execution_id": externalServiceExecutionIDFromRecord(source),
			"retry_record_id":     savedExecution.ID,
			"retry_execution_id":  executionID,
			"status":              result.Status,
			"request_id":          source.RequestID,
			"manual_retry":        true,
		}),
		CreatedAt: Now(),
	})
	message := "手动重试已完成"
	if result.Status != externalServiceHookStatusSuccess {
		message = "手动重试已记录，请在执行记录中查看失败原因"
	}
	return domain.PluginExternalServiceManualRetryResponse{
		PluginCode:        pluginCode,
		SourceExecutionID: externalServiceExecutionIDFromRecord(source),
		RetryExecutionID:  executionID,
		RetryRecordID:     savedExecution.ID,
		Status:            result.Status,
		Message:           message,
	}, nil
}

func externalServiceManualRetryAllowed(status string, success bool) bool {
	if success || strings.TrimSpace(status) == externalServiceHookStatusSuccess {
		return false
	}
	switch strings.TrimSpace(status) {
	case externalServiceHookStatusFailed, externalServiceHookStatusTimeout, externalServiceHookStatusRetryScheduled, externalServiceHookStatusRetryExhausted:
		return true
	default:
		return false
	}
}

func findExternalServiceHook(plugin domain.Plugin, hookName string) (domain.HookDefinition, bool) {
	for _, hook := range plugin.Hooks {
		if strings.TrimSpace(hook.Name) != strings.TrimSpace(hookName) {
			continue
		}
		if strings.TrimSpace(hook.ServiceType) != externalServiceTypeExternal {
			continue
		}
		if hook.Enabled != nil && !*hook.Enabled {
			continue
		}
		mode := strings.TrimSpace(hook.Mode)
		if mode != "" && mode != string(pluginregistry.HookNonBlocking) {
			continue
		}
		return hook, true
	}
	return domain.HookDefinition{}, false
}

func hookEventFromExecution(source domain.HookExecution, pluginCode string) pluginregistry.HookEvent {
	event := pluginregistry.HookEvent{
		Name: source.HookName,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			HookName:    source.HookName,
			RequestID:   source.RequestID,
			PluginCode:  pluginCode,
			ContentType: source.ContentType,
			CommunityID: source.CommunityID,
			CategoryID:  source.CategoryID,
			ContentID:   source.ContentID,
			ActorType:   pluginregistry.HookActorType(source.ActorType),
			ActorID:     source.ActorID,
			UserID:      source.UserID,
			AdminUserID: source.AdminUserID,
			Metadata: map[string]any{
				"manual_retry":        true,
				"source_record_id":    source.ID,
				"source_execution_id": externalServiceExecutionIDFromRecord(source),
			},
		},
	}
	if event.Ctx.ActorType == "" {
		event.Ctx.ActorType = pluginregistry.HookActorSystem
	}
	return event
}

func manualRetryMetadata(operator PluginExternalServiceOperator, source domain.HookExecution) map[string]any {
	return map[string]any{
		"manual_retry":        true,
		"source_record_id":    source.ID,
		"source_execution_id": externalServiceExecutionIDFromRecord(source),
		"operator_actor":      firstNonEmpty(operator.Name, "system"),
		"operator_actor_id":   operator.ID,
		"request_id":          source.RequestID,
	}
}

func externalServiceExecutionIDFromRecord(record domain.HookExecution) string {
	meta := map[string]any{}
	_ = json.Unmarshal([]byte(record.Metadata), &meta)
	if v, ok := meta["execution_id"].(string); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	if record.ID > 0 {
		return strconv.FormatInt(record.ID, 10)
	}
	return ""
}
