package domain

// PluginConfigVersionScope indicates the config scope.
// scope: global | community
type PluginConfigVersionScope = string

const (
	PluginConfigScopeGlobal    PluginConfigVersionScope = "global"
	PluginConfigScopeCommunity PluginConfigVersionScope = "community"
)

// PluginConfigDiffItem is a redacted diff item for display and audit metadata.
// type: added | removed | changed | unchanged
type PluginConfigDiffItem struct {
	Path   string `json:"path"`
	Type   string `json:"type"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

// PluginConfigVersion stores one snapshot of plugin config changes.
//
// NOTE: config_json may be stored as-is (same as current config storage policy),
// but APIs must return redacted values for sensitive fields.
type PluginConfigVersion struct {
	ID                  int64  `json:"id"`
	PluginCode          string `json:"plugin_code"`
	Scope               string `json:"scope"`
	CommunityID         int64  `json:"community_id,omitempty"`
	VersionNo           int    `json:"version_no"`
	ConfigJSON          string `json:"config_json,omitempty"`
	ConfigHash          string `json:"config_hash,omitempty"`
	ChangedKeysJSON     string `json:"changed_keys_json,omitempty"`
	DiffJSON            string `json:"diff_json,omitempty"`
	Source              string `json:"source,omitempty"`
	OperatorType        string `json:"operator_type,omitempty"`
	OperatorID          int64  `json:"operator_id,omitempty"`
	OperatorName        string `json:"operator_name,omitempty"`
	Reason              string `json:"reason,omitempty"`
	PreviousVersionID   int64  `json:"previous_version_id,omitempty"`
	RollbackFromVersion int64  `json:"rollback_from_version_id,omitempty"`
	MetadataJSON        string `json:"metadata_json,omitempty"`
	CreatedAt           string `json:"created_at,omitempty"`
}

type PluginConfigVersionListItem struct {
	ID           int64    `json:"id"`
	PluginCode   string   `json:"plugin_code"`
	Scope        string   `json:"scope"`
	CommunityID  int64    `json:"community_id,omitempty"`
	VersionNo    int      `json:"version_no"`
	ChangedKeys  []string `json:"changed_keys,omitempty"`
	Source       string   `json:"source,omitempty"`
	OperatorName string   `json:"operator_name,omitempty"`
	ConfigHash   string   `json:"config_hash,omitempty"`
	CreatedAt    string   `json:"created_at,omitempty"`
}

type PluginConfigVersionListResponse struct {
	Items      []PluginConfigVersionListItem `json:"items"`
	Pagination Pagination                    `json:"pagination"`
}

type PluginConfigVersionDetailResponse struct {
	Version      PluginConfigVersionListItem `json:"version"`
	ConfigJSON   any                         `json:"config_json,omitempty"`
	Diff         []PluginConfigDiffItem      `json:"diff,omitempty"`
	ChangedKeys  []string                    `json:"changed_keys,omitempty"`
	Warnings     []string                    `json:"warnings,omitempty"`
	Errors       []string                    `json:"errors,omitempty"`
	RawScopeInfo map[string]any              `json:"scope_info,omitempty"`
}

type PluginConfigRollbackDryRunResponse struct {
	PluginCode  string `json:"plugin_code"`
	Scope       string `json:"scope"`
	Status      string `json:"status"`
	BlockedCode string `json:"blocked_code,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`

	TargetVersion  PluginConfigVersionListItem `json:"target_version"`
	CurrentVersion PluginConfigVersionListItem `json:"current_version"`

	SchemaValidation struct {
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors,omitempty"`
	} `json:"schema_validation"`

	ChangedKeys []string               `json:"changed_keys,omitempty"`
	Diff        []PluginConfigDiffItem `json:"diff,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	Errors      []string               `json:"errors,omitempty"`
}
