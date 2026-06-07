package domain

import "strings"

const (
	PluginWebhookSecretStatusActive   = "active"
	PluginWebhookSecretStatusPrevious = "previous"
	PluginWebhookSecretStatusDisabled = "disabled"
	PluginWebhookSecretStatusRevoked  = "revoked"
	PluginWebhookSecretStatusExpired  = "expired"
)

type PluginWebhookSecret struct {
	ID                int64  `json:"id"`
	PluginCode        string `json:"plugin_code"`
	TargetURL         string `json:"target_url"`
	SecretRef         string `json:"secret_ref"`
	SecretCiphertext  string `json:"secret_ciphertext,omitempty"`
	SecretHash        string `json:"secret_hash,omitempty"`
	Version           int    `json:"version"`
	Status            string `json:"status"`
	RotationGroup     string `json:"rotation_group,omitempty"`
	PreviousSecretRef string `json:"previous_secret_ref,omitempty"`
	ActiveFrom        string `json:"active_from,omitempty"`
	ActiveUntil       string `json:"active_until,omitempty"`
	GraceUntil        string `json:"grace_until,omitempty"`
	CreatedBy         int64  `json:"created_by,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
	RotatedAt         string `json:"rotated_at,omitempty"`
	RevokedAt         string `json:"revoked_at,omitempty"`
	LastUsedAt        string `json:"last_used_at,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

type PluginWebhookSecretFilter struct {
	PluginCode string
	Status     string
	SecretRef  string
	Page       int
	PageSize   int
}

type PluginWebhookSecretListResponse struct {
	Items      []PluginWebhookSecret `json:"items"`
	Pagination Pagination            `json:"pagination"`
}

func (f PluginWebhookSecretFilter) Normalize() PluginWebhookSecretFilter {
	f.PluginCode = strings.TrimSpace(f.PluginCode)
	f.Status = strings.TrimSpace(f.Status)
	f.SecretRef = strings.TrimSpace(f.SecretRef)
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	return f
}
