package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
)

const (
	DefaultWebhookMaxAttempts       = 5
	defaultWebhookCircuitThreshold  = 5
	defaultWebhookCircuitOpenFor    = 10 * time.Minute
	defaultWebhookManualRetryReason = "manual_retry"
)

type WebhookRetryResult struct {
	Processed int `json:"processed"`
	Success   int `json:"success"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
}

func (s *Service) WebhookDeliveriesAdmin(filter domain.WebhookDeliveryFilter) (domain.WebhookDeliveryListResponse, error) {
	items, total, err := s.repo.WebhookDeliveries(filter)
	if err != nil {
		return domain.WebhookDeliveryListResponse{}, err
	}
	f := filter.Normalize()
	return domain.WebhookDeliveryListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     f.Page,
			PageSize: f.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) WebhookCircuitBreakersAdmin(filter domain.WebhookCircuitBreakerFilter) (domain.WebhookCircuitBreakerListResponse, error) {
	items, total, err := s.repo.WebhookCircuitBreakers(filter)
	if err != nil {
		return domain.WebhookCircuitBreakerListResponse{}, err
	}
	f := filter.Normalize()
	return domain.WebhookCircuitBreakerListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     f.Page,
			PageSize: f.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) WebhookDeliveryByID(id int64) (domain.WebhookDelivery, bool) {
	return s.repo.WebhookDeliveryByID(id)
}

func (s *Service) WebhookCircuitBreakerByID(id int64) (domain.WebhookCircuitBreaker, bool) {
	return s.repo.WebhookCircuitBreakerByID(id)
}

func (s *Service) RetryDueWebhookDeliveries(ctx context.Context, limit int) (WebhookRetryResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	now := time.Now()
	due, err := s.repo.DueWebhookDeliveries(now.Format("2006-01-02 15:04:05"), limit)
	if err != nil {
		return WebhookRetryResult{}, err
	}
	result := WebhookRetryResult{}
	for _, d := range due {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		result.Processed++
		status, err := s.retryOneWebhookDelivery(ctx, d.ID, false)
		if err != nil {
			result.Failed++
			continue
		}
		switch status {
		case domain.WebhookDeliveryStatusSuccess:
			result.Success++
		case domain.WebhookDeliveryStatusCircuitOpen, domain.WebhookDeliveryStatusSkipped:
			result.Skipped++
		default:
			// failed / retry_scheduled / retry_exhausted
			result.Failed++
		}
	}
	return result, nil
}

func (s *Service) ManualRetryWebhookDelivery(ctx context.Context, deliveryID int64) (domain.WebhookDelivery, error) {
	status, err := s.retryOneWebhookDelivery(ctx, deliveryID, true)
	if err != nil {
		return domain.WebhookDelivery{}, err
	}
	out, ok := s.repo.WebhookDeliveryByID(deliveryID)
	if !ok {
		return domain.WebhookDelivery{}, errors.New("delivery not found")
	}
	out.Status = status
	return out, nil
}

func (s *Service) CloseWebhookCircuitBreaker(ctx context.Context, breakerID int64) (domain.WebhookCircuitBreaker, error) {
	b, ok := s.repo.WebhookCircuitBreakerByID(breakerID)
	if !ok {
		return domain.WebhookCircuitBreaker{}, errors.New("circuit breaker not found")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	b.Status = domain.WebhookCircuitBreakerStatusClosed
	b.FailureCount = 0
	b.SuccessCount = 0
	b.ClosedAt = now
	b.NextProbeAt = ""
	b.LastErrorMessage = ""
	b.UpdatedAt = now
	out, err := s.repo.UpsertWebhookCircuitBreaker(b)
	if err != nil {
		return domain.WebhookCircuitBreaker{}, err
	}
	return out, nil
}

func (s *Service) OpenWebhookCircuitBreaker(ctx context.Context, breakerID int64) (domain.WebhookCircuitBreaker, error) {
	b, ok := s.repo.WebhookCircuitBreakerByID(breakerID)
	if !ok {
		return domain.WebhookCircuitBreaker{}, errors.New("circuit breaker not found")
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	b.Status = domain.WebhookCircuitBreakerStatusOpen
	if b.FailureCount <= 0 {
		b.FailureCount = defaultWebhookCircuitThreshold
	}
	b.OpenedAt = now
	b.NextProbeAt = time.Now().Add(defaultWebhookCircuitOpenFor).Format("2006-01-02 15:04:05")
	b.UpdatedAt = now
	out, err := s.repo.UpsertWebhookCircuitBreaker(b)
	if err != nil {
		return domain.WebhookCircuitBreaker{}, err
	}
	return out, nil
}

func (s *Service) retryOneWebhookDelivery(ctx context.Context, id int64, manual bool) (string, error) {
	d, ok := s.repo.WebhookDeliveryByID(id)
	if !ok {
		return "", errors.New("delivery not found")
	}
	// Default: do not allow re-run successful deliveries (unless a future "replay" feature exists).
	if d.Status == domain.WebhookDeliveryStatusSuccess {
		return d.Status, errors.New("success delivery cannot be retried")
	}

	// Lock by state transition.
	from := domain.WebhookDeliveryStatusRetryScheduled
	if manual {
		// allow retry_exhausted & failed by manual, but still need lock
		if d.Status == domain.WebhookDeliveryStatusRetryExhausted {
			from = domain.WebhookDeliveryStatusRetryExhausted
		} else if d.Status == domain.WebhookDeliveryStatusFailed {
			from = domain.WebhookDeliveryStatusFailed
		}
	}
	ok2, err := s.repo.TryMarkWebhookDeliveryStatus(d.ID, from, domain.WebhookDeliveryStatusSending)
	if err != nil {
		return "", err
	}
	if !ok2 {
		// someone else is processing or status changed
		return d.Status, nil
	}
	// reload
	d, _ = s.repo.WebhookDeliveryByID(id)

	now := time.Now()
	d.StartedAt = now.Format("2006-01-02 15:04:05")
	d.UpdatedAt = d.StartedAt
	if manual {
		d.RetryReason = defaultWebhookManualRetryReason
	}
	_, _ = s.repo.SaveWebhookDelivery(d)

	// Circuit breaker gating.
	cb, _ := s.repo.WebhookCircuitBreakerByKey(d.PluginCode, d.TargetURL)
	cbStatus, allowProbe := shouldAllowDeliveryWithBreaker(cb, now)
	if cbStatus == domain.WebhookCircuitBreakerStatusOpen && !allowProbe {
		d.Status = domain.WebhookDeliveryStatusCircuitOpen
		d.FinishedAt = now.Format("2006-01-02 15:04:05")
		d.ErrorMessage = "circuit breaker is open"
		_, _ = s.repo.SaveWebhookDelivery(d)
		return d.Status, nil
	}
	if cbStatus == domain.WebhookCircuitBreakerStatusOpen && allowProbe {
		// move to half_open for probe
		cb.Status = domain.WebhookCircuitBreakerStatusHalfOpen
		cb.NextProbeAt = ""
		cb.UpdatedAt = now.Format("2006-01-02 15:04:05")
		_, _ = s.repo.UpsertWebhookCircuitBreaker(cb)
	}

	// Perform a minimal network attempt. This is a governance path; do not run any plugin code.
	// v1.7.6: sign + auth headers with HMAC-SHA256 based on active secret_ref.
	respStatus, retryAfterSeconds, errMsg := s.doSignedWebhookAttempt(ctx, &d)
	d.ResponseStatus = respStatus
	d.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
	d.DurationMS = time.Since(now).Milliseconds()

	// Local signing/auth failure: do not retry and do not count towards circuit breaker.
	if strings.TrimSpace(d.SignatureStatus) != "" && d.SignatureStatus != "signed" {
		d.Status = domain.WebhookDeliveryStatusFailed
		d.NextRetryAt = ""
		if strings.TrimSpace(d.ErrorMessage) == "" {
			d.ErrorMessage = truncateString(firstNonEmpty(errMsg, d.SignatureError, "signature failed"), 1000)
		}
		_, _ = s.repo.SaveWebhookDelivery(d)
		return d.Status, nil
	}

	if errMsg == "" && respStatus >= 200 && respStatus < 300 {
		d.Status = domain.WebhookDeliveryStatusSuccess
		d.ErrorMessage = ""
		d.NextRetryAt = ""
		_, _ = s.repo.SaveWebhookDelivery(d)

		// success closes breaker if half_open/open
		cb = onWebhookSuccess(cb, d.PluginCode, d.TargetURL, now)
		_, _ = s.repo.UpsertWebhookCircuitBreaker(cb)
		return d.Status, nil
	}

	// failure
	if errMsg == "" {
		errMsg = fmt.Sprintf("http_status=%d", respStatus)
	}
	d.ErrorMessage = truncateString(errMsg, 1000)
	// classify retryable
	retryable, retryReason := classifyRetryable(respStatus, errMsg)
	d.RetryReason = truncateString(firstNonEmpty(retryReason, d.RetryReason), 64)

	// update breaker
	cb = onWebhookFailure(cb, d.PluginCode, d.TargetURL, now, errMsg)
	_, _ = s.repo.UpsertWebhookCircuitBreaker(cb)

	if cb.Status == domain.WebhookCircuitBreakerStatusOpen {
		d.Status = domain.WebhookDeliveryStatusCircuitOpen
		d.NextRetryAt = ""
		_, _ = s.repo.SaveWebhookDelivery(d)
		return d.Status, nil
	}

	if !retryable {
		d.Status = domain.WebhookDeliveryStatusFailed
		d.NextRetryAt = ""
		_, _ = s.repo.SaveWebhookDelivery(d)
		return d.Status, nil
	}

	// schedule retry
	if d.MaxAttempts <= 0 {
		d.MaxAttempts = DefaultWebhookMaxAttempts
	}
	nextAttempt := d.Attempt + 1
	if nextAttempt > d.MaxAttempts {
		d.Status = domain.WebhookDeliveryStatusRetryExhausted
		d.NextRetryAt = ""
		_, _ = s.repo.SaveWebhookDelivery(d)
		return d.Status, nil
	}
	d.Attempt = nextAttempt
	next := computeNextRetryAt(now, nextAttempt, respStatus, retryAfterSeconds)
	d.NextRetryAt = next.Format("2006-01-02 15:04:05")
	d.Status = domain.WebhookDeliveryStatusRetryScheduled
	_, _ = s.repo.SaveWebhookDelivery(d)
	return d.Status, nil
}

func computeNextRetryAt(now time.Time, attempt int, respStatus int, retryAfterSeconds int) time.Time {
	// attempt here is the "next attempt number" (2..)
	// Retry-After support for 429.
	if respStatus == 429 {
		if retryAfterSeconds > 0 {
			return now.Add(time.Duration(retryAfterSeconds) * time.Second)
		}
	}
	switch attempt {
	case 2:
		return now.Add(1 * time.Minute)
	case 3:
		return now.Add(5 * time.Minute)
	case 4:
		return now.Add(15 * time.Minute)
	case 5:
		return now.Add(1 * time.Hour)
	default:
		return now.Add(1 * time.Hour)
	}
}

func classifyRetryable(respStatus int, errMsg string) (bool, string) {
	if respStatus == 0 {
		// network errors
		return true, "network_error"
	}
	if respStatus >= 200 && respStatus < 300 {
		return false, ""
	}
	if respStatus == 401 {
		return false, "unauthorized"
	}
	if respStatus == 403 {
		return false, "forbidden"
	}
	if respStatus == 408 {
		return true, "timeout"
	}
	if respStatus == 429 {
		return true, "rate_limited"
	}
	if respStatus >= 500 && respStatus <= 599 {
		return true, "server_error"
	}
	if respStatus >= 400 && respStatus <= 499 {
		return false, "client_error"
	}
	_ = errMsg
	return true, "unknown"
}

func shouldAllowDeliveryWithBreaker(cb domain.WebhookCircuitBreaker, now time.Time) (string, bool) {
	if cb.PluginCode == "" || cb.TargetURL == "" {
		return domain.WebhookCircuitBreakerStatusClosed, false
	}
	switch cb.Status {
	case domain.WebhookCircuitBreakerStatusClosed:
		return cb.Status, false
	case domain.WebhookCircuitBreakerStatusHalfOpen:
		// allow probe delivery.
		return cb.Status, true
	case domain.WebhookCircuitBreakerStatusOpen:
		if strings.TrimSpace(cb.NextProbeAt) == "" {
			return cb.Status, false
		}
		t, err := time.ParseInLocation("2006-01-02 15:04:05", cb.NextProbeAt, time.Local)
		if err != nil {
			return cb.Status, false
		}
		return cb.Status, !t.After(now)
	default:
		return cb.Status, false
	}
}

func onWebhookFailure(cb domain.WebhookCircuitBreaker, pluginCode, targetURL string, now time.Time, errMsg string) domain.WebhookCircuitBreaker {
	if cb.PluginCode == "" {
		cb.PluginCode = pluginCode
		cb.TargetURL = targetURL
		cb.Status = domain.WebhookCircuitBreakerStatusClosed
		cb.CreatedAt = now.Format("2006-01-02 15:04:05")
	}
	cb.LastErrorMessage = truncateString(errMsg, 1000)
	cb.LastFailureAt = now.Format("2006-01-02 15:04:05")
	cb.FailureCount++
	cb.SuccessCount = 0
	if cb.FailureCount >= defaultWebhookCircuitThreshold {
		cb.Status = domain.WebhookCircuitBreakerStatusOpen
		cb.OpenedAt = now.Format("2006-01-02 15:04:05")
		cb.NextProbeAt = now.Add(defaultWebhookCircuitOpenFor).Format("2006-01-02 15:04:05")
	}
	cb.UpdatedAt = now.Format("2006-01-02 15:04:05")
	return cb
}

func onWebhookSuccess(cb domain.WebhookCircuitBreaker, pluginCode, targetURL string, now time.Time) domain.WebhookCircuitBreaker {
	if cb.PluginCode == "" {
		cb.PluginCode = pluginCode
		cb.TargetURL = targetURL
		cb.CreatedAt = now.Format("2006-01-02 15:04:05")
	}
	cb.LastSuccessAt = now.Format("2006-01-02 15:04:05")
	cb.SuccessCount++
	cb.FailureCount = 0
	if cb.Status == domain.WebhookCircuitBreakerStatusHalfOpen || cb.Status == domain.WebhookCircuitBreakerStatusOpen {
		cb.Status = domain.WebhookCircuitBreakerStatusClosed
		cb.ClosedAt = now.Format("2006-01-02 15:04:05")
		cb.NextProbeAt = ""
	}
	cb.UpdatedAt = now.Format("2006-01-02 15:04:05")
	return cb
}

func doWebhookAttempt(ctx context.Context, target string) (int, int, string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return 0, 0, "empty target_url"
	}
	u, err := url.Parse(target)
	if err != nil {
		return 0, 0, "invalid target_url"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return 0, 0, "unsupported scheme"
	}
	// NOTE: we intentionally do not attempt SSRF protection here; governance is admin-only
	// and this is a minimal implementation. SSRF hardening can be added later.
	req, err := http.NewRequestWithContext(ctx, "POST", target, strings.NewReader(`{}`))
	if err != nil {
		return 0, 0, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err.Error()
	}
	defer resp.Body.Close()
	// consume small excerpt to keep connections reusable
	_, _ = io.ReadAll(io.LimitReader(resp.Body, 2048))
	retryAfterSeconds := parseRetryAfter(resp.Header.Get("Retry-After"))
	return resp.StatusCode, retryAfterSeconds, ""
}

func (s *Service) doSignedWebhookAttempt(ctx context.Context, d *domain.WebhookDelivery) (int, int, string) {
	if d == nil {
		return 0, 0, "nil delivery"
	}
	target := strings.TrimSpace(d.TargetURL)
	if target == "" {
		d.SignatureAlg = webhookSignatureAlgHMACSHA256
		d.SignatureStatus = "sign_failed"
		d.SignatureError = "empty target_url"
		_, _ = s.repo.SaveWebhookDelivery(*d)
		return 0, 0, "empty target_url"
	}
	u, err := url.Parse(target)
	if err != nil || u.Scheme == "" || u.Host == "" {
		d.SignatureAlg = webhookSignatureAlgHMACSHA256
		d.SignatureStatus = "sign_failed"
		d.SignatureError = "invalid target_url"
		_, _ = s.repo.SaveWebhookDelivery(*d)
		return 0, 0, "invalid target_url"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		d.SignatureAlg = webhookSignatureAlgHMACSHA256
		d.SignatureStatus = "sign_failed"
		d.SignatureError = "unsupported scheme"
		_, _ = s.repo.SaveWebhookDelivery(*d)
		return 0, 0, "unsupported scheme"
	}

	// Resolve secret: local failures should not be retried.
	resolved, sigStatus, sigErr := s.resolveActiveWebhookSecret(d.PluginCode, d.TargetURL)
	d.SignatureAlg = webhookSignatureAlgHMACSHA256
	d.SecretRef = strings.TrimSpace(resolved.Record.SecretRef)
	d.SignatureStatus = sigStatus
	if sigErr != nil {
		d.SignatureError = truncateString(sigErr.Error(), 500)
		// Audit local signature failures without leaking secret/ciphertext/signature.
		action := "plugin.webhook.signature.sign_failed"
		switch sigStatus {
		case "secret_missing":
			action = "plugin.webhook.signature.secret_missing"
		case "secret_disabled":
			action = "plugin.webhook.signature.secret_disabled"
		case "secret_revoked":
			action = "plugin.webhook.signature.secret_revoked"
		case "secret_expired":
			action = "plugin.webhook.signature.secret_expired"
		}
		s.repo.AppendAdminLog(domain.AdminLog{
			Site:      "admin",
			Actor:     "system",
			ActorType: "system",
			ActorID:   0,
			Action:    action,
			Target:    "webhook-deliveries#" + strings.TrimSpace(d.DeliveryID),
			Metadata: mustJSON(map[string]any{
				"plugin_code":      d.PluginCode,
				"event_id":         d.EventID,
				"delivery_id":      d.DeliveryID,
				"target_url":       d.TargetURL,
				"signature_alg":    d.SignatureAlg,
				"signature_status": d.SignatureStatus,
				"secret_ref":       d.SecretRef,
				"error_message":    d.SignatureError,
			}),
			CreatedAt: Now(),
		})
		_, _ = s.repo.SaveWebhookDelivery(*d)
		return 0, 0, "signing secret unavailable"
	}

	method := "POST"
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}

	body := []byte(`{}`)
	bodySum := sha256.Sum256(body)
	bodySHA := hex.EncodeToString(bodySum[:])
	d.BodySHA256 = bodySHA

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	signingString := ts + "." + method + "." + path + "." + bodySHA
	hexSig := signWebhookHMACSHA256([]byte(resolved.Plaintext), signingString)

	// Persist signature metadata; do not store signature itself.
	d.SignedAt = time.Now().Format("2006-01-02 15:04:05")
	d.SignatureError = ""
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     "system",
		ActorType: "system",
		ActorID:   0,
		Action:    "plugin.webhook.signature.signed",
		Target:    "webhook-deliveries#" + strings.TrimSpace(d.DeliveryID),
		Metadata: mustJSON(map[string]any{
			"plugin_code":      d.PluginCode,
			"event_id":         d.EventID,
			"delivery_id":      d.DeliveryID,
			"target_url":       d.TargetURL,
			"signature_alg":    d.SignatureAlg,
			"signature_status": d.SignatureStatus,
			"secret_ref":       d.SecretRef,
			"body_sha256":      d.BodySHA256,
		}),
		CreatedAt: Now(),
	})

	headers := map[string]string{
		"Content-Type":             "application/json",
		"X-DevHub-Event-ID":        strings.TrimSpace(d.EventID),
		"X-DevHub-Delivery-ID":     strings.TrimSpace(d.DeliveryID),
		"X-DevHub-Plugin-Code":     strings.TrimSpace(d.PluginCode),
		"X-DevHub-Timestamp":       ts,
		"X-DevHub-Signature-Alg":   webhookSignatureAlgHMACSHA256,
		"X-DevHub-Idempotency-Key": strings.TrimSpace(d.EventID),
		"X-DevHub-Request-ID":      strings.TrimSpace(d.DeliveryID),
		"X-DevHub-Body-SHA256":     bodySHA,
		"X-DevHub-Secret-Ref":      strings.TrimSpace(d.SecretRef),
	}

	// Store request headers JSON with signature redacted.
	headersForStore := map[string]string{}
	for k, v := range headers {
		headersForStore[k] = v
	}
	headersForStore["X-DevHub-Signature"] = "v1=[REDACTED]"
	if raw, err := json.Marshal(headersForStore); err == nil {
		d.RequestHeadersJSON = string(raw)
	}
	d.RequestBodySHA256 = bodySHA
	_, _ = s.repo.SaveWebhookDelivery(*d)

	req, err := http.NewRequestWithContext(ctx, method, target, strings.NewReader(string(body)))
	if err != nil {
		d.SignatureStatus = "sign_failed"
		d.SignatureError = truncateString(err.Error(), 500)
		_, _ = s.repo.SaveWebhookDelivery(*d)
		return 0, 0, err.Error()
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("X-DevHub-Signature", "v1="+hexSig)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err.Error()
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(io.LimitReader(resp.Body, 2048))
	retryAfterSeconds := parseRetryAfter(resp.Header.Get("Retry-After"))
	return resp.StatusCode, retryAfterSeconds, ""
}

func truncateString(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func parseRetryAfter(v string) int {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	// Retry-After can be delta-seconds or HTTP-date. First handle delta-seconds.
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return n
	}
	// fallback: HTTP date format
	if t, err := http.ParseTime(v); err == nil {
		secs := int(time.Until(t).Seconds())
		if secs > 0 {
			return secs
		}
	}
	return 0
}
