package domain

const (
	PluginPackageDownloadStatusPending         = "pending"
	PluginPackageDownloadStatusDownloading     = "downloading"
	PluginPackageDownloadStatusDownloaded      = "downloaded"
	PluginPackageDownloadStatusChecksumFailed  = "checksum_failed"
	PluginPackageDownloadStatusRejected        = "rejected"
	PluginPackageDownloadStatusFailed          = "failed"
	PluginPackageDownloadStatusChecksumMissing = "checksum_missing"
	PluginPackageDownloadStatusDeleted         = "deleted"
)

type PluginPackageDownloadRequest struct {
	PluginCode string `json:"plugin_code"`
	Version    string `json:"version"`
	PackageURL string `json:"package_url"`
	SHA256     string `json:"sha256"`
}

type PluginPackageDownloadRecord struct {
	ID             int64  `json:"id"`
	PluginCode     string `json:"plugin_code"`
	Version        string `json:"version"`
	SourceURL      string `json:"source_url"`
	FinalURL       string `json:"final_url"`
	Status         string `json:"status"`
	FileName       string `json:"file_name"`
	StagingPath    string `json:"staging_path"`
	FileSize       int64  `json:"file_size"`
	SHA256Expected string `json:"sha256_expected"`
	SHA256Actual   string `json:"sha256_actual"`
	ContentType    string `json:"content_type"`
	ErrorCode      string `json:"error_code"`
	ErrorMessage   string `json:"error_message"`
	CreatedBy      int64  `json:"created_by"`
	CreatedAt      string `json:"created_at"`
	DownloadedAt   string `json:"downloaded_at"`
	DeletedAt      string `json:"deleted_at"`
	UpdatedAt      string `json:"updated_at"`
}

type PluginPackageDownloadFilter struct {
	Status     string
	PluginCode string
	Keyword    string
	Page       int
	PageSize   int
}

type PluginPackageDownloadListResponse struct {
	Items      []PluginPackageDownloadRecord `json:"items"`
	Pagination Pagination                    `json:"pagination"`
	Summary    map[string]int                `json:"summary"`
}
