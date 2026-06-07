package domain

import "strings"

const (
	WebhookEventStatusPending     = "pending"
	WebhookEventStatusDelivering  = "delivering"
	WebhookEventStatusDelivered   = "delivered"
	WebhookEventStatusFailed      = "failed"
	WebhookEventStatusSkipped     = "skipped"
	WebhookEventStatusCircuitOpen = "circuit_open"

	WebhookDeliveryStatusPending        = "pending"
	WebhookDeliveryStatusSending        = "sending"
	WebhookDeliveryStatusSuccess        = "success"
	WebhookDeliveryStatusFailed         = "failed"
	WebhookDeliveryStatusRetryScheduled = "retry_scheduled"
	WebhookDeliveryStatusRetryExhausted = "retry_exhausted"
	WebhookDeliveryStatusSkipped        = "skipped"
	WebhookDeliveryStatusCircuitOpen    = "circuit_open"

	WebhookCircuitBreakerStatusClosed   = "closed"
	WebhookCircuitBreakerStatusOpen     = "open"
	WebhookCircuitBreakerStatusHalfOpen = "half_open"
)

type WebhookEvent struct {
	ID           int64  `json:"id"`
	EventID      string `json:"event_id"`
	EventName    string `json:"event_name"`
	EventType    string `json:"event_type"`
	PluginCode   string `json:"plugin_code"`
	HookName     string `json:"hook_name"`
	Mode         string `json:"mode"`
	CommunityID  int64  `json:"community_id,omitempty"`
	ActorType    string `json:"actor_type,omitempty"`
	ActorID      int64  `json:"actor_id,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   int64  `json:"resource_id,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	PayloadJSON  string `json:"payload_json,omitempty"`
	MetadataJSON string `json:"metadata_json,omitempty"`
	Status       string `json:"status"`
	OccurredAt   string `json:"occurred_at,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type WebhookEventFilter struct {
	PluginCode  string
	HookName    string
	Mode        string
	Status      string
	CommunityID int64
	ActorType   string
	ActorID     int64
	RequestID   string
	Page        int
	PageSize    int
}

type WebhookEventListResponse struct {
	Items      []WebhookEvent `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type WebhookDelivery struct {
	ID                  int64  `json:"id"`
	DeliveryID          string `json:"delivery_id"`
	EventID             string `json:"event_id"`
	PluginCode          string `json:"plugin_code"`
	HookName            string `json:"hook_name,omitempty"`
	TargetURL           string `json:"target_url"`
	Status              string `json:"status"`
	Attempt             int    `json:"attempt"`
	MaxAttempts         int    `json:"max_attempts"`
	SignatureAlg        string `json:"signature_alg,omitempty"`
	SecretRef           string `json:"secret_ref,omitempty"`
	BodySHA256          string `json:"body_sha256,omitempty"`
	SignatureStatus     string `json:"signature_status,omitempty"`
	SignedAt            string `json:"signed_at,omitempty"`
	SignatureError      string `json:"signature_error,omitempty"`
	RequestHeadersJSON  string `json:"request_headers_json,omitempty"`
	RequestBodySHA256   string `json:"request_body_sha256,omitempty"`
	ResponseStatus      int    `json:"response_status,omitempty"`
	ResponseBodyExcerpt string `json:"response_body_excerpt,omitempty"`
	ErrorMessage        string `json:"error_message,omitempty"`
	RetryReason         string `json:"retry_reason,omitempty"`
	DurationMS          int64  `json:"duration_ms,omitempty"`
	NextRetryAt         string `json:"next_retry_at,omitempty"`
	StartedAt           string `json:"started_at,omitempty"`
	FinishedAt          string `json:"finished_at,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

type WebhookDeliveryFilter struct {
	PluginCode string
	HookName   string
	Status     string
	EventID    string
	DeliveryID string
	DueBefore  string
	Page       int
	PageSize   int
}

type WebhookDeliveryListResponse struct {
	Items      []WebhookDelivery `json:"items"`
	Pagination Pagination        `json:"pagination"`
}

type WebhookCircuitBreaker struct {
	ID               int64  `json:"id"`
	PluginCode       string `json:"plugin_code"`
	TargetURL        string `json:"target_url"`
	Status           string `json:"status"`
	FailureCount     int    `json:"failure_count"`
	SuccessCount     int    `json:"success_count"`
	OpenedAt         string `json:"opened_at,omitempty"`
	ClosedAt         string `json:"closed_at,omitempty"`
	NextProbeAt      string `json:"next_probe_at,omitempty"`
	LastErrorMessage string `json:"last_error_message,omitempty"`
	LastFailureAt    string `json:"last_failure_at,omitempty"`
	LastSuccessAt    string `json:"last_success_at,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

type WebhookCircuitBreakerFilter struct {
	PluginCode string
	Status     string
	Page       int
	PageSize   int
}

type WebhookCircuitBreakerListResponse struct {
	Items      []WebhookCircuitBreaker `json:"items"`
	Pagination Pagination              `json:"pagination"`
}

func (f WebhookEventFilter) Normalize() WebhookEventFilter {
	f.PluginCode = strings.TrimSpace(f.PluginCode)
	f.HookName = strings.TrimSpace(f.HookName)
	f.Mode = strings.TrimSpace(f.Mode)
	f.Status = strings.TrimSpace(f.Status)
	f.ActorType = strings.TrimSpace(f.ActorType)
	f.RequestID = strings.TrimSpace(f.RequestID)
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	return f
}

func (f WebhookDeliveryFilter) Normalize() WebhookDeliveryFilter {
	f.PluginCode = strings.TrimSpace(f.PluginCode)
	f.HookName = strings.TrimSpace(f.HookName)
	f.Status = strings.TrimSpace(f.Status)
	f.EventID = strings.TrimSpace(f.EventID)
	f.DeliveryID = strings.TrimSpace(f.DeliveryID)
	f.DueBefore = strings.TrimSpace(f.DueBefore)
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	return f
}

func (f WebhookCircuitBreakerFilter) Normalize() WebhookCircuitBreakerFilter {
	f.PluginCode = strings.TrimSpace(f.PluginCode)
	f.Status = strings.TrimSpace(f.Status)
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	return f
}
