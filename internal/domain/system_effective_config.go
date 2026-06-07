package domain

// SystemEffectiveConfigResponse is a redacted, human-readable view of the
// runtime configuration that is currently effective.
//
// Security rules:
// - Never include plaintext token/secret/Authorization values.
// - Never include encrypted_value or startup root keys.
// - Non-sensitive runtime fields are intentionally shown in clear text.
type SystemEffectiveConfigResponse struct {
	DevHubVersion                string                                     `json:"devhub_version"`
	StoreMode                    string                                     `json:"store_mode"`
	GeneratedAt                  string                                     `json:"generated_at"`
	RootKeyStatus                PluginConfigKeyStatusResponse              `json:"root_key_status"`
	SecretCenterStatus           SecretCenterStatusResponse                 `json:"secret_center_status"`
	ExternalServiceHTTPAllowlist PluginExternalServiceHTTPAllowlistResponse `json:"external_service_http_allowlist"`
	HTTPAllowlistSource          string                                     `json:"http_allowlist_source,omitempty"`
	ExternalServices             []PluginExternalServiceEffectiveConfig     `json:"external_services"`
	WebhookCallbackSecurity      WebhookCallbackSecuritySummary             `json:"webhook_callback_security"`
	Secrets                      []SecretRefRecord                          `json:"secrets,omitempty"`
	DiagnosticText               string                                     `json:"diagnostic_text,omitempty"`
	QuickLinks                   map[string]string                          `json:"quick_links,omitempty"`
	NextSteps                    []string                                   `json:"next_steps,omitempty"`
	Notes                        []string                                   `json:"notes,omitempty"`
}

type PluginExternalServiceEffectiveConfig struct {
	PluginCode        string                            `json:"plugin_code"`
	PluginName        string                            `json:"plugin_name,omitempty"`
	EndpointURL       string                            `json:"endpoint_url,omitempty"`
	HealthCheckPath   string                            `json:"health_check_path,omitempty"`
	Enabled           bool                              `json:"enabled"`
	AuthType          string                            `json:"auth_type,omitempty"`
	TimeoutMS         int                               `json:"timeout_ms,omitempty"`
	FailurePolicy     string                            `json:"failure_policy,omitempty"`
	CurrentHealth     string                            `json:"current_health,omitempty"`
	LastHealthStatus  string                            `json:"last_health_status,omitempty"`
	LastCheckedAt     string                            `json:"last_checked_at,omitempty"`
	LastHealthCheckAt string                            `json:"last_health_check_at,omitempty"`
	LastSuccessAt     string                            `json:"last_success_at,omitempty"`
	LastFailureAt     string                            `json:"last_failure_at,omitempty"`
	LastErrorAt       string                            `json:"last_error_at,omitempty"`
	LastErrorMessage  string                            `json:"last_error_message,omitempty"`
	LastErrorSummary  string                            `json:"last_error_summary,omitempty"`
	EndpointOrigin    string                            `json:"endpoint_origin,omitempty"`
	EndpointScheme    string                            `json:"endpoint_scheme,omitempty"`
	AllowlistSource   string                            `json:"http_allowlist_source,omitempty"`
	AllowlistMatched  bool                              `json:"http_allowlist_matched"`
	AllowlistMessage  string                            `json:"http_allowlist_message,omitempty"`
	TokenRef          string                            `json:"token_ref,omitempty"`
	TokenNamespace    string                            `json:"token_namespace,omitempty"`
	TokenName         string                            `json:"token_name,omitempty"`
	TokenStatus       string                            `json:"token_status,omitempty"`
	TokenMasked       string                            `json:"token_masked,omitempty"`
	TokenKeyID        string                            `json:"token_key_id,omitempty"`
	TokenLastUsedAt   string                            `json:"token_last_used_at,omitempty"`
	TokenRotatedAt    string                            `json:"token_rotated_at,omitempty"`
	TokenUsageType    string                            `json:"token_usage_type,omitempty"`
	TokenSourceType   string                            `json:"token_source_type,omitempty"`
	TokenSourceID     string                            `json:"token_source_id,omitempty"`
	TokenSourceCode   string                            `json:"token_source_code,omitempty"`
	TokenAvailable    bool                              `json:"token_available"`
	TokenMessage      string                            `json:"token_message,omitempty"`
	ConfigSource      string                            `json:"config_source,omitempty"`
	TokenSource       string                            `json:"token_source,omitempty"`
	NextSteps         []string                          `json:"next_steps,omitempty"`
	Troubleshooting   map[string]string                 `json:"troubleshooting,omitempty"`
	Diagnostics       []PluginExternalServiceDiagnostic `json:"diagnostics,omitempty"`
	HTTPPolicy        *PluginExternalServiceHTTPPolicy  `json:"http_policy,omitempty"`
}

type WebhookCallbackSecuritySummary struct {
	WebhookSecretTotal      int               `json:"webhook_secret_total"`
	WebhookSecretByStatus   map[string]int    `json:"webhook_secret_by_status,omitempty"`
	CallbackTokenTotal      int               `json:"callback_token_total"`
	CallbackTokenByStatus   map[string]int    `json:"callback_token_by_status,omitempty"`
	ActiveWebhookSecrets    int               `json:"active_webhook_secrets"`
	ActiveCallbackTokens    int               `json:"active_callback_tokens"`
	DisabledOrRevokedCount  int               `json:"disabled_or_revoked_count"`
	LastWebhookSecretUsedAt string            `json:"last_webhook_secret_used_at,omitempty"`
	LastCallbackTokenUsedAt string            `json:"last_callback_token_used_at,omitempty"`
	Notes                   []string          `json:"notes,omitempty"`
	QuickLinks              map[string]string `json:"quick_links,omitempty"`
}
