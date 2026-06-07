package domain

const (
	PluginPackagePrecheckStatusPassed              = "passed"
	PluginPackagePrecheckStatusFailed              = "failed"
	PluginPackagePrecheckStatusRejected            = "rejected"
	PluginPackagePrecheckStatusUnsafeFilesDetected = "unsafe_files_detected"
	PluginPackagePrecheckStatusManifestInvalid     = "manifest_invalid"
	PluginPackagePrecheckStatusDeleted             = "deleted"

	PluginPackageCompatCheckStatusPending                   = "pending"
	PluginPackageCompatCheckStatusChecking                  = "checking"
	PluginPackageCompatCheckStatusPassed                    = "passed"
	PluginPackageCompatCheckStatusWarning                   = "warning"
	PluginPackageCompatCheckStatusFailed                    = "failed"
	PluginPackageCompatCheckStatusIncompatible              = "incompatible"
	PluginPackageCompatCheckStatusDependencyMissing         = "dependency_missing"
	PluginPackageCompatCheckStatusDependencyVersionMismatch = "dependency_version_mismatch"
	PluginPackageCompatCheckStatusConflictDetected          = "conflict_detected"
	PluginPackageCompatCheckStatusDeleted                   = "deleted"
)

type PluginPackagePrecheckRecord struct {
	ID                int64  `json:"id"`
	PackageDownloadID int64  `json:"package_download_id"`
	PluginCode        string `json:"plugin_code"`
	Version           string `json:"version"`
	Status            string `json:"status"`
	ManifestJSON      string `json:"manifest_json,omitempty"`
	PackagePath       string `json:"package_path,omitempty"`
	StagingPath       string `json:"staging_path,omitempty"`
	ChecksumStatus    string `json:"checksum_status,omitempty"`
	ErrorCode         string `json:"error_code,omitempty"`
	ErrorMessage      string `json:"error_message,omitempty"`
	CreatedBy         int64  `json:"created_by"`
	StartedAt         string `json:"started_at,omitempty"`
	FinishedAt        string `json:"finished_at,omitempty"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type PluginPackagePrecheckFilter struct {
	Status     string
	PluginCode string
	Keyword    string
	Page       int
	PageSize   int
}

type PluginPackageCompatCheckRecord struct {
	ID                     int64  `json:"id"`
	PackageDownloadID      int64  `json:"package_download_id"`
	PackagePrecheckID      int64  `json:"package_precheck_id"`
	PluginCode             string `json:"plugin_code"`
	Version                string `json:"version"`
	Status                 string `json:"status"`
	CanInstall             bool   `json:"can_install"`
	CoreVersion            string `json:"core_version"`
	CompatibleCoreVersion  string `json:"compatible_core_version"`
	DependencyResultJSON   string `json:"dependency_result_json,omitempty"`
	ConflictResultJSON     string `json:"conflict_result_json,omitempty"`
	PermissionResultJSON   string `json:"permission_result_json,omitempty"`
	RouteResultJSON        string `json:"route_result_json,omitempty"`
	MenuResultJSON         string `json:"menu_result_json,omitempty"`
	HookResultJSON         string `json:"hook_result_json,omitempty"`
	ConfigSchemaResultJSON string `json:"config_schema_result_json,omitempty"`
	MigrationResultJSON    string `json:"migration_result_json,omitempty"`
	WarningsJSON           string `json:"warnings_json,omitempty"`
	ErrorsJSON             string `json:"errors_json,omitempty"`
	SummaryJSON            string `json:"summary_json,omitempty"`
	CreatedBy              int64  `json:"created_by"`
	StartedAt              string `json:"started_at,omitempty"`
	FinishedAt             string `json:"finished_at,omitempty"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

type PluginPackageCompatCheckFilter struct {
	Status            string
	PluginCode        string
	PackagePrecheckID int64
	Keyword           string
	Page              int
	PageSize          int
}

type PluginPackageCompatCheckResponse struct {
	PluginPackageCompatCheckRecord
	DependencyResult   any      `json:"dependency_result,omitempty"`
	ConflictResult     any      `json:"conflict_result,omitempty"`
	PermissionResult   any      `json:"permission_result,omitempty"`
	RouteResult        any      `json:"route_result,omitempty"`
	MenuResult         any      `json:"menu_result,omitempty"`
	HookResult         any      `json:"hook_result,omitempty"`
	ConfigSchemaResult any      `json:"config_schema_result,omitempty"`
	MigrationResult    any      `json:"migration_result,omitempty"`
	Warnings           []string `json:"warnings"`
	Errors             []string `json:"errors"`
	Summary            any      `json:"summary,omitempty"`
}

type PluginPackageCompatCheckListResponse struct {
	Items      []PluginPackageCompatCheckResponse `json:"items"`
	Pagination Pagination                         `json:"pagination"`
	Summary    map[string]int                     `json:"summary"`
}
