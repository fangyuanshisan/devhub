package domain

// PluginExternalServiceConfig describes a persisted external_service runtime config.
// It is separate from manifest-time PluginExternalService declaration data.
type PluginExternalServiceConfig struct {
	PluginCode       string                            `json:"plugin_code"`
	ServiceType      string                            `json:"service_type"`
	EndpointURL      string                            `json:"endpoint_url,omitempty"`
	HealthCheckPath  string                            `json:"health_check_path,omitempty"`
	TimeoutMS        int                               `json:"timeout_ms,omitempty"`
	FailurePolicy    string                            `json:"failure_policy,omitempty"`
	AuthType         string                            `json:"auth_type,omitempty"`
	TokenRef         string                            `json:"token_ref,omitempty"`
	TokenSecret      *SecretRefRecord                  `json:"token_secret,omitempty"`
	TokenCiphertext  string                            `json:"-"`
	TokenHash        string                            `json:"-"`
	Enabled          bool                              `json:"enabled"`
	Status           string                            `json:"status,omitempty"`
	LastHealthStatus string                            `json:"last_health_status,omitempty"`
	LastCheckedAt    string                            `json:"last_checked_at,omitempty"`
	LastSuccessAt    string                            `json:"last_success_at,omitempty"`
	LastFailureAt    string                            `json:"last_failure_at,omitempty"`
	FailureCount     int                               `json:"failure_count,omitempty"`
	WarningThreshold int                               `json:"warning_threshold,omitempty"`
	ErrorThreshold   int                               `json:"error_threshold,omitempty"`
	LastErrorMessage string                            `json:"last_error_message,omitempty"`
	CreatedAt        string                            `json:"created_at,omitempty"`
	UpdatedAt        string                            `json:"updated_at,omitempty"`
	Diagnostics      []PluginExternalServiceDiagnostic `json:"diagnostics,omitempty"`
	HTTPPolicy       *PluginExternalServiceHTTPPolicy  `json:"http_policy,omitempty"`
}

type PluginExternalServiceDiagnostic struct {
	Type             string   `json:"type"`
	Code             string   `json:"code,omitempty"`
	FailureType      string   `json:"failure_type,omitempty"`
	EndpointURL      string   `json:"endpoint_url,omitempty"`
	Message          string   `json:"message"`
	Suggestion       string   `json:"suggestion,omitempty"`
	SafetyRejected   bool     `json:"safety_rejected,omitempty"`
	NeedsAllowlist   bool     `json:"needs_allowlist,omitempty"`
	AllowlistEnv     string   `json:"allowlist_env,omitempty"`
	AllowlistExample string   `json:"allowlist_example,omitempty"`
	AllowlistOrigins []string `json:"allowlist_origins,omitempty"`
}

type PluginExternalServiceHTTPPolicy struct {
	HTTPSAllowed               bool     `json:"https_allowed"`
	LocalhostHTTPAllowed       bool     `json:"localhost_http_allowed"`
	NonLocalHTTPNeedsAllowlist bool     `json:"non_local_http_needs_allowlist"`
	AllowlistEnv               string   `json:"allowlist_env"`
	AllowlistOrigins           []string `json:"allowlist_origins,omitempty"`
	Defaults                   []string `json:"defaults,omitempty"`
	EnvAllowlist               []string `json:"env_allowlist,omitempty"`
	AdminAllowlist             []string `json:"admin_allowlist,omitempty"`
	EffectiveAllowlist         []string `json:"effective_allowlist,omitempty"`
}

type PluginExternalServiceHTTPAllowlistEntry struct {
	ID        string `json:"id,omitempty"`
	Origin    string `json:"origin"`
	Source    string `json:"source"`
	Usage     string `json:"usage,omitempty"`
	Status    string `json:"status,omitempty"`
	Deletable bool   `json:"deletable"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type PluginExternalServiceHTTPAllowlistResponse struct {
	Defaults           []PluginExternalServiceHTTPAllowlistEntry `json:"defaults"`
	EnvAllowlist       []PluginExternalServiceHTTPAllowlistEntry `json:"env_allowlist"`
	AdminAllowlist     []PluginExternalServiceHTTPAllowlistEntry `json:"admin_allowlist"`
	EffectiveAllowlist []PluginExternalServiceHTTPAllowlistEntry `json:"effective_allowlist"`
	Policy             PluginExternalServiceHTTPPolicy           `json:"policy"`
}

type PluginExternalServiceHTTPAllowlistUpdateRequest struct {
	Origin        string `json:"origin"`
	Usage         string `json:"usage,omitempty"`
	RiskConfirmed bool   `json:"risk_confirmed"`
}

// PluginExternalServiceUpdateRequest is the admin writable form for external_service config.
type PluginExternalServiceUpdateRequest struct {
	EndpointURL      string `json:"endpoint_url,omitempty"`
	HealthCheckPath  string `json:"health_check_path,omitempty"`
	TimeoutMS        int    `json:"timeout_ms,omitempty"`
	FailurePolicy    string `json:"failure_policy,omitempty"`
	AuthType         string `json:"auth_type,omitempty"`
	Token            string `json:"token,omitempty"`
	Enabled          *bool  `json:"enabled,omitempty"`
	WarningThreshold int    `json:"warning_threshold,omitempty"`
	ErrorThreshold   int    `json:"error_threshold,omitempty"`
}

// PluginExternalServiceHealthCheckResponse is returned by manual health checks.
type PluginExternalServiceHealthCheckResponse struct {
	PluginCode   string                            `json:"plugin_code"`
	ServiceType  string                            `json:"service_type"`
	EndpointURL  string                            `json:"endpoint_url,omitempty"`
	HealthStatus string                            `json:"health_status,omitempty"`
	Status       string                            `json:"status,omitempty"`
	CheckedAt    string                            `json:"checked_at,omitempty"`
	DurationMS   int64                             `json:"duration_ms,omitempty"`
	Message      string                            `json:"message,omitempty"`
	ErrorCode    string                            `json:"error_code,omitempty"`
	Suggestion   string                            `json:"suggestion,omitempty"`
	Diagnostics  []PluginExternalServiceDiagnostic `json:"diagnostics,omitempty"`
	HTTPPolicy   *PluginExternalServiceHTTPPolicy  `json:"http_policy,omitempty"`
}

// PluginExternalServiceManualRetryResponse is returned after a manual external_service hook retry.
type PluginExternalServiceManualRetryResponse struct {
	PluginCode        string `json:"plugin_code"`
	SourceExecutionID string `json:"source_execution_id"`
	RetryExecutionID  string `json:"retry_execution_id"`
	RetryRecordID     int64  `json:"retry_record_id,omitempty"`
	Status            string `json:"status"`
	Message           string `json:"message"`
}
