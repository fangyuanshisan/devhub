package service

import (
	"encoding/json"
	"strings"
)

const (
	approvalRedactedValue  = "[REDACTED]"
	approvalEncryptedValue = "[ENCRYPTED]"
)

var approvalSensitiveKeywords = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"access_token",
	"refresh_token",
	"api_key",
	"key",
	"credential",
	"app_secret",
	"aes_key",
	"encoding_aes_key",
	"private_key",
	"client_secret",
}

func scrubJSONStringForSnapshot(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	// Best-effort JSON scrub: keep it JSON if possible.
	if json.Valid([]byte(raw)) {
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err == nil {
			v = scrubAnyForSnapshot(v)
			out, _ := json.Marshal(v)
			return string(out)
		}
	}
	// Non-JSON: just prevent leaking ciphertext markers.
	if strings.Contains(raw, "enc:v1:") {
		return strings.ReplaceAll(raw, "enc:v1:", approvalEncryptedValue+":")
	}
	if strings.Contains(raw, "enc:v2:") {
		return strings.ReplaceAll(raw, "enc:v2:", approvalEncryptedValue+":")
	}
	return raw
}

// scrubManifestJSONForSnapshot keeps manifest structure intact (required for execute),
// but it ensures ciphertext markers are never persisted.
func scrubManifestJSONForSnapshot(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if json.Valid([]byte(raw)) {
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err == nil {
			v = scrubCiphertextOnly(v)
			out, _ := json.Marshal(v)
			return string(out)
		}
	}
	if strings.Contains(raw, "enc:v1:") {
		return strings.ReplaceAll(raw, "enc:v1:", approvalEncryptedValue+":")
	}
	if strings.Contains(raw, "enc:v2:") {
		return strings.ReplaceAll(raw, "enc:v2:", approvalEncryptedValue+":")
	}
	return raw
}

func scrubAnyForSnapshot(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			key := strings.TrimSpace(k)
			if isSensitiveKey(key) {
				out[key] = scrubSensitiveValue(vv)
				continue
			}
			out[key] = scrubAnyForSnapshot(vv)
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, it := range t {
			out = append(out, scrubAnyForSnapshot(it))
		}
		return out
	case string:
		s := t
		if strings.Contains(s, "enc:v1:") {
			return approvalEncryptedValue
		}
		if strings.Contains(s, "enc:v2:") {
			return approvalEncryptedValue
		}
		return s
	default:
		return v
	}
}

func scrubCiphertextOnly(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			out[k] = scrubCiphertextOnly(vv)
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, it := range t {
			out = append(out, scrubCiphertextOnly(it))
		}
		return out
	case string:
		if strings.Contains(t, "enc:v1:") {
			return approvalEncryptedValue
		}
		if strings.Contains(t, "enc:v2:") {
			return approvalEncryptedValue
		}
		return t
	default:
		return v
	}
}

func scrubSensitiveValue(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case string:
		if strings.TrimSpace(t) == "" {
			return ""
		}
		return approvalRedactedValue
	default:
		// For nested objects/arrays under sensitive keys, do a deep scrub anyway.
		return scrubAnyForSnapshot(v)
	}
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return false
	}
	for _, kw := range approvalSensitiveKeywords {
		if strings.Contains(key, kw) {
			return true
		}
	}
	return false
}
