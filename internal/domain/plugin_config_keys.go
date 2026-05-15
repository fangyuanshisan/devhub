package domain

type PluginConfigKeyStatusResponse struct {
	CurrentKeyID      string   `json:"current_key_id"`
	LoadedKeyIDs      []string `json:"loaded_key_ids"`
	LegacyV1Supported bool     `json:"legacy_v1_supported"`
	KeyCount          int      `json:"key_count"`
	Status            string   `json:"status"` // ok|warning|blocked
	Warnings          []string `json:"warnings"`
}

type PluginConfigKeyRotationDryRunRequest struct {
	Scope                 string `json:"scope"` // all|plugin|community
	PluginCode            string `json:"plugin_code"`
	CommunityID           int64  `json:"community_id"`
	IncludeConfigVersions bool   `json:"include_config_versions"`
}

type PluginConfigKeyRotationDryRunSummary struct {
	TotalSensitiveValues int `json:"total_sensitive_values"`
	AlreadyCurrent       int `json:"already_current"`
	NeedsReencrypt       int `json:"needs_reencrypt"`
	LegacyV1             int `json:"legacy_v1"`
	DecryptFailed        int `json:"decrypt_failed"`
	MissingKey           int `json:"missing_key"`
}

type PluginConfigKeyRotationDryRunItem struct {
	PluginCode    string `json:"plugin_code"`
	Scope         string `json:"scope"` // global|community
	CommunityID   int64  `json:"community_id"`
	FieldPath     string `json:"field_path"`
	CipherVersion string `json:"cipher_version"` // v1|v2|plain|invalid
	KeyID         string `json:"key_id"`
	Status        string `json:"status"` // already_current|needs_reencrypt|decrypt_failed|missing_key|cipher_invalid
	Message       string `json:"message"`
}

type PluginConfigKeyRotationDryRunResponse struct {
	Status       string                               `json:"status"` // ok|warning|blocked
	CurrentKeyID string                               `json:"current_key_id"`
	Summary      PluginConfigKeyRotationDryRunSummary `json:"summary"`
	Items        []PluginConfigKeyRotationDryRunItem  `json:"items"`
	Warnings     []string                             `json:"warnings"`
	Errors       []APIError                           `json:"errors"`
}

type PluginConfigKeyRotationReencryptRequest struct {
	Scope                 string `json:"scope"` // all|plugin|community
	PluginCode            string `json:"plugin_code"`
	CommunityID           int64  `json:"community_id"`
	IncludeConfigVersions bool   `json:"include_config_versions"`
	ConfirmCurrentKeyID   string `json:"confirm_current_key_id"`
}

type PluginConfigKeyRotationReencryptResponse struct {
	Status       string                               `json:"status"` // ok|warning|blocked
	CurrentKeyID string                               `json:"current_key_id"`
	Summary      PluginConfigKeyRotationDryRunSummary `json:"summary"`
	UpdatedCount int                                  `json:"updated_count"`
	Warnings     []string                             `json:"warnings"`
	Errors       []APIError                           `json:"errors"`
}
