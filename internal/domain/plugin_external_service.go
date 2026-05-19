package domain

// PluginExternalServiceConfig describes a persisted external_service runtime config.
// It is separate from manifest-time PluginExternalService declaration data.
type PluginExternalServiceConfig struct {
	PluginCode       string `json:"plugin_code"`
	ServiceType      string `json:"service_type"`
	EndpointURL      string `json:"endpoint_url,omitempty"`
	HealthCheckPath  string `json:"health_check_path,omitempty"`
	TimeoutMS        int    `json:"timeout_ms,omitempty"`
	FailurePolicy    string `json:"failure_policy,omitempty"`
	AuthType         string `json:"auth_type,omitempty"`
	TokenRef         string `json:"token_ref,omitempty"`
	TokenCiphertext  string `json:"-"`
	TokenHash        string `json:"-"`
	Enabled          bool   `json:"enabled"`
	Status           string `json:"status,omitempty"`
	LastHealthStatus string `json:"last_health_status,omitempty"`
	LastCheckedAt    string `json:"last_checked_at,omitempty"`
	LastSuccessAt    string `json:"last_success_at,omitempty"`
	LastFailureAt    string `json:"last_failure_at,omitempty"`
	FailureCount     int    `json:"failure_count,omitempty"`
	WarningThreshold int    `json:"warning_threshold,omitempty"`
	ErrorThreshold   int    `json:"error_threshold,omitempty"`
	LastErrorMessage string `json:"last_error_message,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
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
	PluginCode   string `json:"plugin_code"`
	ServiceType  string `json:"service_type"`
	EndpointURL  string `json:"endpoint_url,omitempty"`
	HealthStatus string `json:"health_status,omitempty"`
	Status       string `json:"status,omitempty"`
	CheckedAt    string `json:"checked_at,omitempty"`
	DurationMS   int64  `json:"duration_ms,omitempty"`
	Message      string `json:"message,omitempty"`
}
