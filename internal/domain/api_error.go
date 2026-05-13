package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIError is a structured error payload for HTTP APIs.
//
// Compatibility note:
// - The HTTP layer should still populate legacy `error` for old clients.
// - `Message` is user-facing and safe to render directly.
// - `Details` must not contain secrets; sensitive fields should be masked upstream.
type APIError struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	Suggestion string         `json:"suggestion,omitempty"`
	HTTPStatus int            `json:"-"`
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	if strings.TrimSpace(e.Code) != "" {
		return e.Code
	}
	return "请求失败"
}

func (e *APIError) WithStatus(status int) *APIError {
	if e == nil {
		return nil
	}
	e.HTTPStatus = status
	return e
}

func (e *APIError) WithDetail(key string, value any) *APIError {
	if e == nil {
		return nil
	}
	if e.Details == nil {
		e.Details = map[string]any{}
	}
	e.Details[key] = value
	return e
}

func (e *APIError) WithSuggestion(s string) *APIError {
	if e == nil {
		return nil
	}
	e.Suggestion = strings.TrimSpace(s)
	return e
}

// CloneDetails returns a shallow copy of Details for safe mutation.
func (e *APIError) CloneDetails() map[string]any {
	if e == nil || e.Details == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(e.Details))
	for k, v := range e.Details {
		out[k] = v
	}
	return out
}

func NewAPIError(code, message string) *APIError {
	return &APIError{Code: strings.TrimSpace(code), Message: strings.TrimSpace(message)}
}

func NewPluginError(code, message string) *APIError {
	return NewAPIError(code, message)
}

// MaskPotentialSecret masks value for audit/details logging.
func MaskPotentialSecret(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return ""
		}
		if len(s) <= 4 {
			return "***"
		}
		return s[:2] + "***" + s[len(s)-2:]
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return "***"
		}
		if len(raw) > 128 {
			return fmt.Sprintf("[masked %d bytes]", len(raw))
		}
		return "***"
	}
}
