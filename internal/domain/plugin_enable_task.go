package domain

const (
	PluginEnableTaskStatusPending      = "pending"
	PluginEnableTaskStatusEnabling     = "enabling"
	PluginEnableTaskStatusEnabled      = "enabled"
	PluginEnableTaskStatusFailed       = "failed"
	PluginEnableTaskStatusRolledBack   = "rolled_back"
	PluginEnableTaskStatusRollbackFail = "rollback_failed"
	PluginEnableTaskStatusDeleted      = "deleted"
)

type PluginEnableTask struct {
	ID int64 `json:"id"`

	PluginCode string `json:"plugin_code"`
	Version    string `json:"version"`

	PluginInstallTaskID    int64 `json:"plugin_install_task_id,omitempty"`
	PluginEnablePrecheckID int64 `json:"plugin_enable_precheck_id,omitempty"`

	Status         string `json:"status"`
	PreviousStatus string `json:"previous_status,omitempty"`
	NewStatus      string `json:"new_status,omitempty"`

	RegisteredContentTypesJSON string `json:"registered_content_types_json,omitempty"`
	RegisteredPermissionsJSON  string `json:"registered_permissions_json,omitempty"`
	RegisteredMenusJSON        string `json:"registered_menus_json,omitempty"`
	RegisteredRoutesJSON       string `json:"registered_routes_json,omitempty"`
	RegisteredHooksJSON        string `json:"registered_hooks_json,omitempty"`
	EffectiveConfigJSON        string `json:"effective_config_json,omitempty"`

	ErrorsJSON      string `json:"errors_json,omitempty"`
	WarningsJSON    string `json:"warnings_json,omitempty"`
	RollbackLogJSON string `json:"rollback_log_json,omitempty"`

	StartedAt   string `json:"started_at,omitempty"`
	FinishedAt  string `json:"finished_at,omitempty"`
	DurationMS  int64  `json:"duration_ms,omitempty"`
	EnabledBy   int64  `json:"enabled_by,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type PluginEnableTaskFilter struct {
	Status     string
	PluginCode string
	Keyword    string
	Page       int
	PageSize   int
}

type PluginEnableTaskResponse struct {
	PluginEnableTask

	Registered map[string]int `json:"registered,omitempty"`
	Errors     []string       `json:"errors,omitempty"`
	Warnings   []string       `json:"warnings,omitempty"`
	Rollback   []string       `json:"rollback,omitempty"`
}

type PluginEnableTaskListResponse struct {
	Items      []PluginEnableTaskResponse `json:"items"`
	Pagination Pagination                `json:"pagination"`
}

