package domain

// PluginApprovalAction indicates the governance action that needs review.
// action: install | upgrade | enable | rollback
type PluginApprovalAction = string

const (
	PluginApprovalActionInstall  PluginApprovalAction = "install"
	PluginApprovalActionUpgrade  PluginApprovalAction = "upgrade"
	PluginApprovalActionEnable   PluginApprovalAction = "enable"
	PluginApprovalActionRollback PluginApprovalAction = "rollback"
)

// PluginApprovalStatus indicates the request status.
// status: pending | approved | rejected | canceled | executed | failed
type PluginApprovalStatus = string

const (
	PluginApprovalStatusPending  PluginApprovalStatus = "pending"
	PluginApprovalStatusApproved PluginApprovalStatus = "approved"
	PluginApprovalStatusRejected PluginApprovalStatus = "rejected"
	PluginApprovalStatusCanceled PluginApprovalStatus = "canceled"
	PluginApprovalStatusExecuted PluginApprovalStatus = "executed"
	PluginApprovalStatusFailed   PluginApprovalStatus = "failed"
)

type PluginApprovalRequest struct {
	ID        int64  `json:"id"`
	RequestNo string `json:"request_no,omitempty"`

	Action         string `json:"action"`
	PluginCode     string `json:"plugin_code"`
	PluginName     string `json:"plugin_name,omitempty"`
	CurrentVersion string `json:"current_version,omitempty"`
	TargetVersion  string `json:"target_version,omitempty"`

	PackagePath           string `json:"package_path,omitempty"`
	PackageChecksumStatus string `json:"package_checksum_status,omitempty"`
	PackageRiskLevel      string `json:"package_risk_level,omitempty"`

	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`

	RequestedBy     int64  `json:"requested_by,omitempty"`
	RequestedByName string `json:"requested_by_name,omitempty"`
	RequestedAt     string `json:"requested_at,omitempty"`

	ReviewedBy     int64  `json:"reviewed_by,omitempty"`
	ReviewedByName string `json:"reviewed_by_name,omitempty"`
	ReviewedAt     string `json:"reviewed_at,omitempty"`
	ReviewComment  string `json:"review_comment,omitempty"`

	ExecutedBy int64  `json:"executed_by,omitempty"`
	ExecutedAt string `json:"executed_at,omitempty"`

	ExecuteResultJSON string `json:"execute_result_json,omitempty"`

	ManifestJSON          string `json:"manifest_json,omitempty"`
	DryRunJSON            string `json:"dry_run_json,omitempty"`
	RiskReportJSON        string `json:"risk_report_json,omitempty"`
	DependencySummaryJSON string `json:"dependency_summary_json,omitempty"`
	CompatibilityJSON     string `json:"compatibility_json,omitempty"`
	ChangedKeysJSON       string `json:"changed_keys_json,omitempty"`
	DiffJSON              string `json:"diff_json,omitempty"`

	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	MetadataJSON string `json:"metadata_json,omitempty"`

	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type PluginApprovalCreateRequest struct {
	Action      string `json:"action"`
	PluginCode  string `json:"plugin_code,omitempty"`
	PackagePath string `json:"package_path,omitempty"`
	// ManifestJSON is a raw manifest payload (object) or a JSON string.
	ManifestJSON any    `json:"manifest_json,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type PluginApprovalReviewRequest struct {
	Comment string `json:"comment,omitempty"`
}

type PluginApprovalCancelRequest struct {
	Comment string `json:"comment,omitempty"`
}

type PluginApprovalFilter struct {
	Status      string
	Action      string
	PluginCode  string
	RequestedBy int64
	ReviewedBy  int64
	Page        int
	PageSize    int
}

type PluginApprovalListItem struct {
	ID              int64  `json:"id"`
	RequestNo       string `json:"request_no,omitempty"`
	Action          string `json:"action"`
	PluginCode      string `json:"plugin_code"`
	PluginName      string `json:"plugin_name,omitempty"`
	CurrentVersion  string `json:"current_version,omitempty"`
	TargetVersion   string `json:"target_version,omitempty"`
	Status          string `json:"status"`
	PackagePath     string `json:"package_path,omitempty"`
	RiskLevel       string `json:"risk_level,omitempty"`
	ChecksumStatus  string `json:"checksum_status,omitempty"`
	RequestedByName string `json:"requested_by_name,omitempty"`
	RequestedAt     string `json:"requested_at,omitempty"`
	ReviewedByName  string `json:"reviewed_by_name,omitempty"`
	ReviewedAt      string `json:"reviewed_at,omitempty"`
}

type PluginApprovalListResponse struct {
	Items      []PluginApprovalListItem `json:"items"`
	Pagination Pagination               `json:"pagination"`
}

type PluginApprovalDetailResponse struct {
	Request PluginApprovalRequest `json:"request"`
}
