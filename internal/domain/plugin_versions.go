package domain

type PluginVersionSource string

const (
	PluginVersionSourceInstalled       PluginVersionSource = "installed"
	PluginVersionSourceLocalPackage    PluginVersionSource = "local_package"
	PluginVersionSourceUploadedPackage PluginVersionSource = "uploaded_package"
	PluginVersionSourceRemoteIndex     PluginVersionSource = "remote_index"
	PluginVersionSourceExportedPackage PluginVersionSource = "exported_package"
)

type PluginVersionFilter struct {
	PluginCode string
	Source     string
	Status     string
	Keyword    string
	Page       int
	PageSize   int
}

type PluginVersionRecord struct {
	PluginCode            string                  `json:"plugin_code"`
	PluginName            string                  `json:"plugin_name,omitempty"`
	Version               string                  `json:"version"`
	Source                string                  `json:"source"`
	Status                string                  `json:"status"`
	PackagePath           string                  `json:"package_path,omitempty"`
	RemoteIndexID         int64                   `json:"remote_index_id,omitempty"`
	RemoteSourceID        string                  `json:"remote_source_id,omitempty"`
	RiskLevel             string                  `json:"risk_level,omitempty"`
	RiskSummary           string                  `json:"risk_summary,omitempty"`
	ChecksumStatus        string                  `json:"checksum_status,omitempty"`
	SignatureStatus       string                  `json:"signature_status,omitempty"`
	PublisherID           string                  `json:"publisher_id,omitempty"`
	PublicKeyID           string                  `json:"public_key_id,omitempty"`
	TrustStatus           string                  `json:"trust_status,omitempty"`
	CoreCompatibility     PluginCoreCompatibility `json:"core_compatibility"`
	InstalledVersion      string                  `json:"installed_version,omitempty"`
	IsInstalled           bool                    `json:"is_installed"`
	IsUpgradeCandidate    bool                    `json:"is_upgrade_candidate"`
	Readonly              bool                    `json:"readonly"`
	ReadonlyMessage       string                  `json:"readonly_message,omitempty"`
	CreatedAt             string                  `json:"created_at,omitempty"`
	UpdatedAt             string                  `json:"updated_at,omitempty"`
	PackageSHA256         string                  `json:"package_sha256,omitempty"`
	SignatureURL          string                  `json:"signature_url,omitempty"`
	CompatibleCoreVersion string                  `json:"compatible_core_version,omitempty"`
	MinCoreVersion        string                  `json:"min_core_version,omitempty"`
}

type PluginVersionRepositoryItem struct {
	PluginCode          string                `json:"plugin_code"`
	PluginName          string                `json:"plugin_name,omitempty"`
	InstalledVersion    string                `json:"installed_version,omitempty"`
	LatestLocalVersion  string                `json:"latest_local_version,omitempty"`
	LatestRemoteVersion string                `json:"latest_remote_version,omitempty"`
	UpdateAvailable     bool                  `json:"update_available"`
	Sources             []string              `json:"sources"`
	RiskSummary         string                `json:"risk_summary,omitempty"`
	Versions            []PluginVersionRecord `json:"versions,omitempty"`
}

type PluginVersionRepositorySummary struct {
	Total           int `json:"total"`
	Installed       int `json:"installed"`
	LocalPackages   int `json:"local_packages"`
	UploadedPackage int `json:"uploaded_packages"`
	RemoteIndex     int `json:"remote_index"`
	UpdateAvailable int `json:"update_available"`
	Readonly        int `json:"readonly"`
}

type PluginVersionRepositoryListResponse struct {
	Items      []PluginVersionRepositoryItem  `json:"items"`
	Pagination Pagination                     `json:"pagination"`
	Summary    PluginVersionRepositorySummary `json:"summary"`
}

type PluginVersionDetailResponse struct {
	PluginCode       string                `json:"plugin_code"`
	PluginName       string                `json:"plugin_name,omitempty"`
	InstalledVersion string                `json:"installed_version,omitempty"`
	Versions         []PluginVersionRecord `json:"versions"`
}

type PluginUpgradeDiffRequest struct {
	Source        string `json:"source"`
	PackagePath   string `json:"package_path,omitempty"`
	RemoteIndexID int64  `json:"remote_index_id,omitempty"`
}

type PluginUpgradeDiffSummary struct {
	Added    int `json:"added"`
	Removed  int `json:"removed"`
	Changed  int `json:"changed"`
	HighRisk int `json:"high_risk"`
	Blocked  int `json:"blocked"`
}

type PluginManifestDiffItem struct {
	Section   string `json:"section"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	RiskLevel string `json:"risk_level"`
	Before    any    `json:"before,omitempty"`
	After     any    `json:"after,omitempty"`
	Message   string `json:"message,omitempty"`
}

type PluginManifestDiffSection struct {
	Section   string                   `json:"section"`
	Title     string                   `json:"title"`
	RiskLevel string                   `json:"risk_level"`
	Items     []PluginManifestDiffItem `json:"items"`
}

type PluginUpgradeDiffResult struct {
	PluginCode        string                      `json:"plugin_code"`
	CurrentVersion    string                      `json:"current_version"`
	TargetVersion     string                      `json:"target_version"`
	Source            string                      `json:"source"`
	Status            string                      `json:"status"`
	Summary           PluginUpgradeDiffSummary    `json:"summary"`
	DiffSections      []PluginManifestDiffSection `json:"diff_sections"`
	RiskReport        PluginPackageRiskReport     `json:"risk_report"`
	Compatibility     PluginCoreCompatibility     `json:"compatibility"`
	Dependencies      PluginDependencySummary     `json:"dependencies"`
	PackageRiskReport PluginPackageRiskReport     `json:"package_risk_report,omitempty"`
	Warnings          []string                    `json:"warnings,omitempty"`
	Errors            []string                    `json:"errors,omitempty"`
	Readonly          bool                        `json:"readonly,omitempty"`
	ReadonlyMessage   string                      `json:"readonly_message,omitempty"`
}
