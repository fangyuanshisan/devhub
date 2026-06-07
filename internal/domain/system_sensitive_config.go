package domain

// SystemSensitiveConfigStatusResponse is a read-only view for startup/runtime sensitive config boundaries.
//
// Security rules:
// - Never include plaintext secrets/tokens/keys.
// - Only show readiness/status, key_id, and env examples.
// - All fields are sourced from startup env/runtime state (not user-editable via UI).
type SystemSensitiveConfigStatusResponse struct {
	PluginConfigKeyring PluginConfigKeyStatusResponse   `json:"plugin_config_keyring"`
	ExternalServiceHTTP PluginExternalServiceHTTPPolicy `json:"external_service_http_policy"`
	SecretCenter        SecretCenterStatusResponse      `json:"secret_center"`
	Notes               []string                        `json:"notes,omitempty"`
}
