package domain

// PluginPackageRepositoryListItem is returned by GET /api/v1/admin/plugins/packages.
type PluginPackageRepositoryListItem struct {
	Path           string `json:"path"`
	Code           string `json:"code,omitempty"`
	Name           string `json:"name,omitempty"`
	Version        string `json:"version,omitempty"`
	ManifestFound  bool   `json:"manifest_found"`
	ChecksumFound  bool   `json:"checksum_found"`
	SignatureFound bool   `json:"signature_found,omitempty"`
	PublisherFound bool   `json:"publisher_found,omitempty"`
	ReadmeFound    bool   `json:"readme_found"`

	// status: ok|warning|blocked|invalid
	Status string `json:"status"`

	RiskLevel   string `json:"risk_level,omitempty"`
	RiskSummary string `json:"risk_summary,omitempty"`

	ChecksumStatus string `json:"checksum_status,omitempty"`
	ManifestValid  *bool  `json:"manifest_valid,omitempty"`

	Signature *PluginPackageSignatureResult `json:"signature,omitempty"`

	TotalFiles int   `json:"total_files,omitempty"`
	TotalSize  int64 `json:"total_size,omitempty"`

	UpdatedAt int64 `json:"updated_at,omitempty"`

	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

type PluginPackageRepositorySummary struct {
	Total   int `json:"total"`
	OK      int `json:"ok"`
	Warning int `json:"warning"`
	Blocked int `json:"blocked"`
	Invalid int `json:"invalid"`
}

type PluginPackageRepositoryListResponse struct {
	Items      []PluginPackageRepositoryListItem `json:"items"`
	Pagination Pagination                        `json:"pagination"`
	Summary    PluginPackageRepositorySummary    `json:"summary"`
}
