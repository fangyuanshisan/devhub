package domain

import "strings"

const (
	PluginCallbackRequestStatusAccepted = "accepted"
	PluginCallbackRequestStatusRejected = "rejected"
	PluginCallbackRequestStatusFailed   = "failed"
)

type PluginCallbackRequest struct {
	ID             int64  `json:"id"`
	RequestID      string `json:"request_id"`
	PluginCode     string `json:"plugin_code"`
	TokenRef       string `json:"token_ref,omitempty"`
	APIPath        string `json:"api_path"`
	Method         string `json:"method"`
	ScopeRequired  string `json:"scope_required,omitempty"`
	Status         string `json:"status"`
	ResponseStatus int    `json:"response_status"`
	ErrorCode      string `json:"error_code,omitempty"`
	ErrorMessage   string `json:"error_message,omitempty"`
	CommunityID    int64  `json:"community_id,omitempty"`
	ActorType      string `json:"actor_type,omitempty"`
	ActorID        int64  `json:"actor_id,omitempty"`
	IPAddress      string `json:"ip_address,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
	DurationMS     int64  `json:"duration_ms,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
}

type PluginCallbackRequestFilter struct {
	PluginCode string
	TokenRef   string
	Status     string
	RequestID  string
	Page       int
	PageSize   int
}

type PluginCallbackRequestListResponse struct {
	Items      []PluginCallbackRequest `json:"items"`
	Pagination Pagination              `json:"pagination"`
}

func (f PluginCallbackRequestFilter) Normalize() PluginCallbackRequestFilter {
	f.PluginCode = strings.TrimSpace(f.PluginCode)
	f.TokenRef = strings.TrimSpace(f.TokenRef)
	f.Status = strings.TrimSpace(f.Status)
	f.RequestID = strings.TrimSpace(f.RequestID)
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 || f.PageSize > 100 {
		f.PageSize = 20
	}
	return f
}
