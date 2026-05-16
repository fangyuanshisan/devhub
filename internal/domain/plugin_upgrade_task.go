package domain

const (
	PluginUpgradeTaskStatusPending      = "pending"
	PluginUpgradeTaskStatusAnalyzing    = "analyzing"
	PluginUpgradeTaskStatusUpgrading    = "upgrading"
	PluginUpgradeTaskStatusUpgraded     = "upgraded"
	PluginUpgradeTaskStatusFailed       = "failed"
	PluginUpgradeTaskStatusRolledBack   = "rolled_back"
	PluginUpgradeTaskStatusRollbackFail = "rollback_failed"
	PluginUpgradeTaskStatusDeleted      = "deleted"
)

// PluginUpgradeTask records a package-based upgrade (compat-check driven) for an installed plugin.
//
// It does NOT execute third-party code, does NOT run migrations, and does NOT auto-enable the upgraded version.
type PluginUpgradeTask struct {
	ID int64 `json:"id"`

	PluginCode string `json:"plugin_code"`
	OldVersion string `json:"old_version"`
	NewVersion string `json:"new_version"`

	OldPluginInstallationID int64 `json:"old_plugin_installation_id,omitempty"`
	NewPackageDownloadID    int64 `json:"new_package_download_id,omitempty"`
	NewPackagePrecheckID    int64 `json:"new_package_precheck_id,omitempty"`
	NewPackageCompatCheckID int64 `json:"new_package_compat_check_id,omitempty"`

	Status               string `json:"status"`
	PreviousPluginStatus string `json:"previous_plugin_status,omitempty"`
	NewPluginStatus      string `json:"new_plugin_status,omitempty"`

	BackupPath     string `json:"backup_path,omitempty"`
	OldInstallPath string `json:"old_install_path,omitempty"`
	NewInstallPath string `json:"new_install_path,omitempty"`

	ManifestDiffJSON    string `json:"manifest_diff_json,omitempty"`
	ConfigDiffJSON      string `json:"config_diff_json,omitempty"`
	PermissionDiffJSON  string `json:"permission_diff_json,omitempty"`
	MenuDiffJSON        string `json:"menu_diff_json,omitempty"`
	RouteDiffJSON       string `json:"route_diff_json,omitempty"`
	HookDiffJSON        string `json:"hook_diff_json,omitempty"`
	ContentTypeDiffJSON string `json:"content_type_diff_json,omitempty"`
	MigrationDiffJSON   string `json:"migration_diff_json,omitempty"`
	ImpactJSON          string `json:"impact_json,omitempty"`

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

type PluginUpgradeTaskFilter struct {
	Status     string
	PluginCode string
	Keyword    string
	Page       int
	PageSize   int
}

type PluginUpgradeTaskResponse struct {
	PluginUpgradeTask
	Errors       []string `json:"errors"`
	Warnings     []string `json:"warnings"`
	Rollback     []string `json:"rollback"`
	ManifestDiff any      `json:"manifest_diff,omitempty"`
	ConfigDiff   any      `json:"config_diff,omitempty"`
	Impact       any      `json:"impact,omitempty"`
}

type PluginUpgradeTaskListResponse struct {
	Items      []PluginUpgradeTaskResponse `json:"items"`
	Pagination Pagination                  `json:"pagination"`
}

type PluginUpgradeImpactResponse struct {
	PluginCode string `json:"plugin_code"`
	OldVersion string `json:"old_version"`
	NewVersion string `json:"new_version"`

	TargetCompatCheckID int64 `json:"target_compat_check_id"`
	PackageDownloadID   int64 `json:"package_download_id,omitempty"`
	PackagePrecheckID   int64 `json:"package_precheck_id,omitempty"`

	Status     string   `json:"status"`
	CanUpgrade bool     `json:"can_upgrade"`
	Errors     []string `json:"errors,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`

	Impact any `json:"impact,omitempty"`

	ManifestDiffSections []PluginManifestDiffSection `json:"manifest_diff_sections,omitempty"`
	ManifestDiffSummary  PluginUpgradeDiffSummary    `json:"manifest_diff_summary,omitempty"`
	DependencyDiff       PluginDependencyDiff        `json:"dependency_diff,omitempty"`
	ConfigDiff           map[string]any              `json:"config_diff,omitempty"`
	MigrationDiff        map[string]any              `json:"migration_diff,omitempty"`
}
