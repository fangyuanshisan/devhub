package domain

const (
	PluginOperationTypeInstall = "install"
	PluginOperationTypeUpgrade = "upgrade"

	PluginOperationStatusCreated   = "created"
	PluginOperationStatusApplied   = "applied"
	PluginOperationStatusFailed    = "failed"
	PluginOperationStatusRecovered = "recovered"
	PluginOperationStatusAbandoned = "abandoned"
)

type PluginOperationSnapshot struct {
	ID int64 `json:"id,omitempty"`

	OperationID   string `json:"operation_id"`
	OperationType string `json:"operation_type"`
	PluginCode    string `json:"plugin_code"`

	FromVersion string `json:"from_version,omitempty"`
	ToVersion   string `json:"to_version,omitempty"`

	PackagePath   string `json:"package_path,omitempty"`
	PackageSource string `json:"package_source,omitempty"`
	ApprovalID    int64  `json:"approval_id,omitempty"`

	BeforePluginJSON       string `json:"before_plugin_json,omitempty"`
	BeforeManifestJSON     string `json:"before_manifest_json,omitempty"`
	BeforeConfigJSON       string `json:"before_config_json,omitempty"`
	BeforeConfigVersionID  int64  `json:"before_config_version_id,omitempty"`
	BeforeMigrationsJSON   string `json:"before_migrations_json,omitempty"`
	BeforePermissionsJSON  string `json:"before_permissions_json,omitempty"`
	BeforeMenusJSON        string `json:"before_menus_json,omitempty"`
	BeforeRoutesJSON       string `json:"before_routes_json,omitempty"`
	BeforeDependenciesJSON string `json:"before_dependencies_json,omitempty"`
	BeforeStatus           string `json:"before_status,omitempty"`

	AfterManifestJSON string `json:"after_manifest_json,omitempty"`
	DryRunJSON        string `json:"dry_run_json,omitempty"`
	RiskReportJSON    string `json:"risk_report_json,omitempty"`
	DiffJSON          string `json:"diff_json,omitempty"`

	ChecksumSummaryJSON  string `json:"checksum_summary_json,omitempty"`
	SignatureSummaryJSON string `json:"signature_summary_json,omitempty"`

	Status       string `json:"status"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`

	CreatedBy    int64  `json:"created_by,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	MetadataJSON string `json:"metadata_json,omitempty"`
}

type PluginOperationFilter struct {
	PluginCode    string
	OperationType string
	Status        string
	Page          int
	PageSize      int
}

type PluginOperationListItem struct {
	OperationID   string `json:"operation_id"`
	OperationType string `json:"operation_type"`
	PluginCode    string `json:"plugin_code"`
	FromVersion   string `json:"from_version,omitempty"`
	ToVersion     string `json:"to_version,omitempty"`
	PackageSource string `json:"package_source,omitempty"`
	Status        string `json:"status"`
	CreatedBy     int64  `json:"created_by,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
}

type PluginOperationListResponse struct {
	Items      []PluginOperationListItem `json:"items"`
	Pagination Pagination                `json:"pagination"`
}

type PluginOperationRecoverDryRunResponse struct {
	Operation      PluginOperationSnapshot `json:"operation"`
	Status         string                  `json:"status"`
	Summary        string                  `json:"summary,omitempty"`
	Detected       []string                `json:"detected_residues,omitempty"`
	CleanupPlan    []string                `json:"cleanup_plan,omitempty"`
	Warnings       []string                `json:"warnings,omitempty"`
	Errors         []string                `json:"errors,omitempty"`
	AllowedActions []string                `json:"allowed_actions,omitempty"`
}

type PluginOperationCleanupResponse struct {
	OperationID string   `json:"operation_id"`
	Status      string   `json:"status"`
	Cleaned     []string `json:"cleaned,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Errors      []string `json:"errors,omitempty"`
}

type PluginUpgradeRollbackDryRunRequest struct {
	OperationID string `json:"operation_id"`
}

type PluginUpgradeRollbackDryRunResponse struct {
	PluginCode     string `json:"plugin_code"`
	OperationID    string `json:"operation_id"`
	FromVersion    string `json:"from_version,omitempty"`
	ToVersion      string `json:"to_version,omitempty"`
	CurrentVersion string `json:"current_version,omitempty"`
	Status         string `json:"status"`

	DiffSections []PluginManifestDiffSection `json:"diff_sections,omitempty"`
	DiffSummary  PluginUpgradeDiffSummary    `json:"diff_summary,omitempty"`

	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}
