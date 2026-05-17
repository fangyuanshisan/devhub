package domain

import "strings"

const (
	PluginCallbackTokenStatusActive   = "active"
	PluginCallbackTokenStatusDisabled = "disabled"
	PluginCallbackTokenStatusRevoked  = "revoked"
	PluginCallbackTokenStatusExpired  = "expired"
)

// PluginCallbackToken is a bearer token for external plugin services to call DevHub Core callback APIs.
// NOTE: token plaintext is never stored; only token_hash is persisted.
type PluginCallbackToken struct {
	ID                   int64  `json:"id"`
	PluginCode           string `json:"plugin_code"`
	PluginInstallationID int64  `json:"plugin_installation_id,omitempty"`
	PublisherID          string `json:"publisher_id,omitempty"`
	TokenRef             string `json:"token_ref"`
	TokenHash            string `json:"token_hash,omitempty"`
	Name                 string `json:"name,omitempty"`
	Status               string `json:"status"`
	ScopesJSON           string `json:"scopes_json,omitempty"`
	CommunityScopeJSON   string `json:"community_scope_json,omitempty"`
	ExpiresAt            string `json:"expires_at,omitempty"`
	LastUsedAt           string `json:"last_used_at,omitempty"`
	LastUsedIP           string `json:"last_used_ip,omitempty"`
	CreatedBy            int64  `json:"created_by,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	RotatedAt            string `json:"rotated_at,omitempty"`
	RevokedAt            string `json:"revoked_at,omitempty"`
	RevokedReason        string `json:"revoked_reason,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
}

type PluginCallbackTokenFilter struct {
	PluginCode string
	Status     string
	Scope      string
	Page       int
	PageSize   int
}

type PluginCallbackTokenListResponse struct {
	Items      []PluginCallbackToken `json:"items"`
	Pagination Pagination            `json:"pagination"`
}

func (f PluginCallbackTokenFilter) Normalize() PluginCallbackTokenFilter {
	f.PluginCode = strings.TrimSpace(f.PluginCode)
	f.Status = strings.TrimSpace(f.Status)
	f.Scope = strings.TrimSpace(f.Scope)
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	return f
}
