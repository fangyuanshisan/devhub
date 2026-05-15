package domain

const (
	PluginPackageUploadStatusUploaded               = "uploaded"
	PluginPackageUploadStatusScanned                = "scanned"
	PluginPackageUploadStatusBlocked                = "blocked"
	PluginPackageUploadStatusStaged                 = "staged"
	PluginPackageUploadStatusApprovalPending        = "approval_pending"
	PluginPackageUploadStatusApprovalRejected       = "approval_rejected"
	PluginPackageUploadStatusApproved               = "approved"
	PluginPackageUploadStatusPromoted               = "promoted"
	PluginPackageUploadStatusInstallApprovalPending = "install_approval_pending"
	PluginPackageUploadStatusInstalled              = "installed"
	PluginPackageUploadStatusCanceled               = "canceled"
	PluginPackageUploadStatusExpired                = "expired"
	PluginPackageUploadStatusDeleted                = "deleted"
	PluginPackageUploadStatusFailed                 = "failed"
)

type PluginPackageZipScan struct {
	TotalEntries     int      `json:"total_entries"`
	CompressedSize   int64    `json:"compressed_size"`
	UncompressedSize int64    `json:"uncompressed_size"`
	Warnings         []string `json:"warnings,omitempty"`
	Errors           []string `json:"errors,omitempty"`
}

type PluginPackageUploadResult struct {
	UploadID           string                          `json:"upload_id"`
	Filename           string                          `json:"filename"`
	Status             string                          `json:"status"`
	StagingPath        string                          `json:"staging_path"`
	PackagePath        string                          `json:"package_path,omitempty"`
	ZipScan            PluginPackageZipScan            `json:"zip_scan"`
	FileScan           PluginPackageFileScan           `json:"file_scan"`
	Checksum           PluginPackageChecksumResult     `json:"checksum"`
	Signature          PluginPackageSignatureResult    `json:"signature,omitempty"`
	ManifestValidation PluginPackageManifestValidation `json:"manifest_validation"`
	InstallDryRun      PluginManifestValidationResult  `json:"install_dry_run"`
	RiskReport         PluginPackageRiskReport         `json:"risk_report"`
	CanPromote         bool                            `json:"can_promote"`
	CanSubmitApproval  bool                            `json:"can_submit_approval"`
	Warnings           []string                        `json:"warnings,omitempty"`
	Errors             []string                        `json:"errors,omitempty"`
	Record             *PluginPackageUploadRecord      `json:"record,omitempty"`
	Actions            []PluginPackageUploadAction     `json:"actions,omitempty"`
}

type PluginPackagePromoteResponse struct {
	Message     string                    `json:"message,omitempty"`
	UploadID    string                    `json:"upload_id"`
	PackagePath string                    `json:"package_path"`
	Status      string                    `json:"status"`
	DryRun      PluginPackageDryRunResult `json:"dry_run"`
	Warnings    []string                  `json:"warnings,omitempty"`
}

type PluginPackageUploadRecord struct {
	ID                     int64  `json:"id,omitempty"`
	UploadID               string `json:"upload_id"`
	OriginalFilename       string `json:"original_filename"`
	UploadedBy             int64  `json:"uploaded_by,omitempty"`
	UploadedByName         string `json:"uploaded_by_name,omitempty"`
	UploadedAt             string `json:"uploaded_at,omitempty"`
	Status                 string `json:"status"`
	PackageCode            string `json:"package_code,omitempty"`
	PackageName            string `json:"package_name,omitempty"`
	PackageVersion         string `json:"package_version,omitempty"`
	UploadPath             string `json:"upload_path,omitempty"`
	StagingPath            string `json:"staging_path,omitempty"`
	PackagePath            string `json:"package_path,omitempty"`
	PromotedPath           string `json:"promoted_path,omitempty"`
	CompressedSize         int64  `json:"compressed_size,omitempty"`
	UncompressedSize       int64  `json:"uncompressed_size,omitempty"`
	FileCount              int    `json:"file_count,omitempty"`
	ChecksumStatus         string `json:"checksum_status,omitempty"`
	SignatureStatus        string `json:"signature_status,omitempty"`
	PublisherID            string `json:"publisher_id,omitempty"`
	TrustStatus            string `json:"trust_status,omitempty"`
	RiskLevel              string `json:"risk_level,omitempty"`
	RiskReportJSON         string `json:"risk_report_json,omitempty"`
	ZipScanJSON            string `json:"zip_scan_json,omitempty"`
	FileScanJSON           string `json:"file_scan_json,omitempty"`
	ManifestValidationJSON string `json:"manifest_validation_json,omitempty"`
	InstallDryRunJSON      string `json:"install_dry_run_json,omitempty"`
	ApprovalID             int64  `json:"approval_id,omitempty"`
	InstallApprovalID      int64  `json:"install_approval_id,omitempty"`
	ExpiresAt              string `json:"expires_at,omitempty"`
	DeletedAt              string `json:"deleted_at,omitempty"`
	ErrorCode              string `json:"error_code,omitempty"`
	ErrorMessage           string `json:"error_message,omitempty"`
	MetadataJSON           string `json:"metadata_json,omitempty"`
	CreatedAt              string `json:"created_at,omitempty"`
	UpdatedAt              string `json:"updated_at,omitempty"`
}

type PluginPackageUploadFilter struct {
	Status      string
	RiskLevel   string
	Keyword     string
	UploadedBy  int64
	PackageCode string
	PublisherID string
	TrustStatus string
	Page        int
	PageSize    int
}

type PluginPackageUploadListItem struct {
	UploadID          string `json:"upload_id"`
	OriginalFilename  string `json:"original_filename"`
	PackageCode       string `json:"package_code,omitempty"`
	PackageName       string `json:"package_name,omitempty"`
	PackageVersion    string `json:"package_version,omitempty"`
	Status            string `json:"status"`
	RiskLevel         string `json:"risk_level,omitempty"`
	ChecksumStatus    string `json:"checksum_status,omitempty"`
	SignatureStatus   string `json:"signature_status,omitempty"`
	TrustStatus       string `json:"trust_status,omitempty"`
	UploadedBy        int64  `json:"uploaded_by,omitempty"`
	UploadedByName    string `json:"uploaded_by_name,omitempty"`
	UploadedAt        string `json:"uploaded_at,omitempty"`
	ExpiresAt         string `json:"expires_at,omitempty"`
	PromotedPath      string `json:"promoted_path,omitempty"`
	ApprovalID        int64  `json:"approval_id,omitempty"`
	InstallApprovalID int64  `json:"install_approval_id,omitempty"`
	ErrorCode         string `json:"error_code,omitempty"`
	RiskSummary       string `json:"risk_summary,omitempty"`
}

type PluginPackageUploadSummary struct {
	Total                  int `json:"total"`
	Uploaded               int `json:"uploaded"`
	Scanned                int `json:"scanned"`
	Staged                 int `json:"staged"`
	Blocked                int `json:"blocked"`
	ApprovalPending        int `json:"approval_pending"`
	ApprovalRejected       int `json:"approval_rejected"`
	Approved               int `json:"approved"`
	Promoted               int `json:"promoted"`
	InstallApprovalPending int `json:"install_approval_pending"`
	Installed              int `json:"installed"`
	Canceled               int `json:"canceled"`
	Expired                int `json:"expired"`
	Deleted                int `json:"deleted"`
	Failed                 int `json:"failed"`
}

type PluginPackageUploadListResponse struct {
	Items      []PluginPackageUploadListItem `json:"items"`
	Pagination Pagination                    `json:"pagination"`
	Summary    PluginPackageUploadSummary    `json:"summary"`
}

type PluginPackageUploadAction struct {
	Action     string `json:"action"`
	Enabled    bool   `json:"enabled"`
	ReasonCode string `json:"reason_code,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type PluginPackageUploadDetailResponse struct {
	Record             PluginPackageUploadRecord       `json:"record"`
	ZipScan            PluginPackageZipScan            `json:"zip_scan"`
	FileScan           PluginPackageFileScan           `json:"file_scan"`
	Checksum           PluginPackageChecksumResult     `json:"checksum"`
	Signature          PluginPackageSignatureResult    `json:"signature,omitempty"`
	RiskReport         PluginPackageRiskReport         `json:"risk_report"`
	ManifestValidation PluginPackageManifestValidation `json:"manifest_validation"`
	InstallDryRun      PluginManifestValidationResult  `json:"install_dry_run"`
	Approval           *PluginApprovalRequest          `json:"approval,omitempty"`
	InstallApproval    *PluginApprovalRequest          `json:"install_approval,omitempty"`
	Actions            []PluginPackageUploadAction     `json:"actions"`
}

type PluginPackageUploadCleanupResponse struct {
	Scanned int      `json:"scanned"`
	Expired int      `json:"expired"`
	Cleaned int      `json:"cleaned"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}
