package domain

const (
	PluginEnablePrecheckStatusPending            = "pending"
	PluginEnablePrecheckStatusChecking           = "checking"
	PluginEnablePrecheckStatusPassed             = "passed"
	PluginEnablePrecheckStatusWarning            = "warning"
	PluginEnablePrecheckStatusFailed             = "failed"
	PluginEnablePrecheckStatusBlocked            = "blocked"
	PluginEnablePrecheckStatusDependencyMissing  = "dependency_missing"
	PluginEnablePrecheckStatusConfigInvalid      = "config_invalid"
	PluginEnablePrecheckStatusMigrationPending   = "migration_pending"
	PluginEnablePrecheckStatusFileIntegrityFailed = "file_integrity_failed"
	PluginEnablePrecheckStatusManifestChanged    = "manifest_changed"
	PluginEnablePrecheckStatusConflictDetected   = "conflict_detected"
	PluginEnablePrecheckStatusDeleted            = "deleted"
)

type PluginEnablePrecheckRecord struct {
	ID int64 `json:"id"`

	PluginCode string `json:"plugin_code"`
	Version    string `json:"version"`

	// Optional linking fields (reserved for future install-task based enable flow).
	PluginInstallTaskID  int64 `json:"plugin_install_task_id,omitempty"`
	PluginInstallationID int64 `json:"plugin_installation_id,omitempty"`

	Status    string `json:"status"`
	CanEnable bool   `json:"can_enable"`

	CoreVersion   string `json:"core_version,omitempty"`
	InstalledPath string `json:"installed_path,omitempty"`
	ManifestSHA256 string `json:"manifest_sha256,omitempty"`

	FileIntegrityResultJSON string `json:"file_integrity_result_json,omitempty"`
	ManifestResultJSON       string `json:"manifest_result_json,omitempty"`
	DependencyResultJSON     string `json:"dependency_result_json,omitempty"`
	ConfigResultJSON         string `json:"config_result_json,omitempty"`
	MigrationResultJSON      string `json:"migration_result_json,omitempty"`
	PermissionResultJSON     string `json:"permission_result_json,omitempty"`
	MenuResultJSON           string `json:"menu_result_json,omitempty"`
	RouteResultJSON          string `json:"route_result_json,omitempty"`
	HookResultJSON           string `json:"hook_result_json,omitempty"`
	ContentTypeResultJSON    string `json:"content_type_result_json,omitempty"`
	RuntimeResultJSON        string `json:"runtime_result_json,omitempty"`

	WarningsJSON string `json:"warnings_json,omitempty"`
	ErrorsJSON   string `json:"errors_json,omitempty"`
	SummaryJSON  string `json:"summary_json,omitempty"`

	CreatedBy  int64  `json:"created_by,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type PluginEnablePrecheckFilter struct {
	Status     string
	PluginCode string
	Keyword    string
	Page       int
	PageSize   int
}

type PluginEnablePrecheckResponse struct {
	PluginEnablePrecheckRecord

	FileIntegrityResult any `json:"file_integrity_result,omitempty"`
	ManifestResult      any `json:"manifest_result,omitempty"`
	DependencyResult    any `json:"dependency_result,omitempty"`
	ConfigResult        any `json:"config_result,omitempty"`
	MigrationResult     any `json:"migration_result,omitempty"`
	PermissionResult    any `json:"permission_result,omitempty"`
	MenuResult          any `json:"menu_result,omitempty"`
	RouteResult         any `json:"route_result,omitempty"`
	HookResult          any `json:"hook_result,omitempty"`
	ContentTypeResult   any `json:"content_type_result,omitempty"`
	RuntimeResult       any `json:"runtime_result,omitempty"`

	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
	Summary  any      `json:"summary,omitempty"`
}

type PluginEnablePrecheckListResponse struct {
	Items      []PluginEnablePrecheckResponse `json:"items"`
	Pagination Pagination                    `json:"pagination"`
}

