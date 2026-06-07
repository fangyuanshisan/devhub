package domain

const (
	PluginUninstallTypeSoft = "soft"

	PluginUninstallTaskStatusPending      = "pending"
	PluginUninstallTaskStatusUninstalling = "uninstalling"
	PluginUninstallTaskStatusSoftDone     = "soft_uninstalled"
	PluginUninstallTaskStatusFailed       = "failed"
	PluginUninstallTaskStatusRolledBack   = "rolled_back"
	PluginUninstallTaskStatusRollbackFail = "rollback_failed"
	PluginUninstallTaskStatusDeleted      = "deleted"
)

type PluginUninstallTask struct {
	ID                   int64  `json:"id"`
	PluginCode           string `json:"plugin_code"`
	Version              string `json:"version"`
	PluginInstallationID int64  `json:"plugin_installation_id,omitempty"`
	PluginEnableTaskID   int64  `json:"plugin_enable_task_id,omitempty"`
	Status               string `json:"status"`
	UninstallType        string `json:"uninstall_type"`
	PreviousStatus       string `json:"previous_status,omitempty"`
	NewStatus            string `json:"new_status,omitempty"`

	AffectedContentsCount    int    `json:"affected_contents_count,omitempty"`
	AffectedCommunitiesCount int    `json:"affected_communities_count,omitempty"`
	DependentPluginsJSON     string `json:"dependent_plugins_json,omitempty"`

	UnregisteredContentTypesJSON string `json:"unregistered_content_types_json,omitempty"`
	UnregisteredPermissionsJSON  string `json:"unregistered_permissions_json,omitempty"`
	UnregisteredMenusJSON        string `json:"unregistered_menus_json,omitempty"`
	UnregisteredRoutesJSON       string `json:"unregistered_routes_json,omitempty"`
	UnregisteredHooksJSON        string `json:"unregistered_hooks_json,omitempty"`

	PreservedFilesJSON      string `json:"preserved_files_json,omitempty"`
	PreservedConfigsJSON    string `json:"preserved_configs_json,omitempty"`
	PreservedMigrationsJSON string `json:"preserved_migrations_json,omitempty"`

	ErrorsJSON      string `json:"errors_json,omitempty"`
	WarningsJSON    string `json:"warnings_json,omitempty"`
	RollbackLogJSON string `json:"rollback_log_json,omitempty"`

	StartedAt   string `json:"started_at,omitempty"`
	FinishedAt  string `json:"finished_at,omitempty"`
	DurationMS  int64  `json:"duration_ms,omitempty"`
	RequestedBy int64  `json:"requested_by,omitempty"`
	Reason      string `json:"reason,omitempty"`

	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type PluginUninstallTaskFilter struct {
	Status     string
	PluginCode string
	Keyword    string
	Page       int
	PageSize   int
}

type PluginUninstallTaskResponse struct {
	PluginUninstallTask
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Rollback []string `json:"rollback"`
}

type PluginUninstallTaskListResponse struct {
	Items      []PluginUninstallTaskResponse `json:"items"`
	Pagination Pagination                    `json:"pagination"`
}
