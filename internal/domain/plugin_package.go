package domain

// PluginPackageInfo describes basic metadata derived from scanning a local plugin package directory.
// It is used by the admin "local plugin package dry-run" preview API.
type PluginPackageInfo struct {
	Path               string `json:"path"`
	DirName            string `json:"dir_name,omitempty"`
	Name               string `json:"name,omitempty"`
	Code               string `json:"code,omitempty"`
	Version            string `json:"version,omitempty"`
	ManifestFound      bool   `json:"manifest_found"`
	ReadmeFound        bool   `json:"readme_found"`
	ConfigExampleFound bool   `json:"config_example_found"`
	ChecksumFound      bool   `json:"checksum_found,omitempty"`
}

type PluginPackageFileEntry struct {
	Path string `json:"path"`
	Size int64  `json:"size,omitempty"`
	Rule string `json:"rule,omitempty"`
}

type PluginPackageFileScan struct {
	TotalFiles     int                      `json:"total_files"`
	TotalSize      int64                    `json:"total_size"`
	AllowedFiles   []PluginPackageFileEntry `json:"allowed_files,omitempty"`
	UnknownFiles   []PluginPackageFileEntry `json:"unknown_files,omitempty"`
	DangerousFiles []PluginPackageFileEntry `json:"dangerous_files,omitempty"`
	Warnings       []string                 `json:"warnings,omitempty"`
	Errors         []string                 `json:"errors,omitempty"`
}

type PluginPackageManifestValidation struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type PluginPackageChecksumFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
}

type PluginPackageChecksumMismatch struct {
	Path     string `json:"path"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
}

// PluginPackageChecksumResult summarizes checksums.json verification.
// status: ok|warning|failed|missing
type PluginPackageChecksumResult struct {
	Algorithm  string                          `json:"algorithm"`
	Status     string                          `json:"status"`
	Matched    []PluginPackageChecksumFile     `json:"matched,omitempty"`
	Mismatched []PluginPackageChecksumMismatch `json:"mismatched,omitempty"`
	Missing    []string                        `json:"missing,omitempty"`
	Extra      []string                        `json:"extra,omitempty"`
	Warnings   []string                        `json:"warnings,omitempty"`
	Errors     []string                        `json:"errors,omitempty"`
}

type PluginPackageRiskItem struct {
	Code       string `json:"code"`
	Level      string `json:"level"`
	Path       string `json:"path,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

type PluginPackageRiskReport struct {
	Level   string                  `json:"level"`
	Score   int                     `json:"score"`
	Summary string                  `json:"summary"`
	Items   []PluginPackageRiskItem `json:"items,omitempty"`
}

// PluginPackageDryRunResult is returned by POST /api/v1/admin/plugins/packages/dry-run.
// status: ok|warning|blocked
type PluginPackageDryRunResult struct {
	Package            PluginPackageInfo               `json:"package"`
	FileScan           PluginPackageFileScan           `json:"file_scan"`
	Checksum           PluginPackageChecksumResult     `json:"checksum"`
	ManifestValidation PluginPackageManifestValidation `json:"manifest_validation"`
	InstallDryRun      PluginManifestValidationResult  `json:"install_dry_run"`
	RiskReport         PluginPackageRiskReport         `json:"risk_report"`
	Status             string                          `json:"status"`
	BlockedCode        string                          `json:"blocked_code,omitempty"`
	BlockedReasons     []string                        `json:"blocked_reasons,omitempty"`
	Warnings           []string                        `json:"warnings,omitempty"`
	Errors             []string                        `json:"errors,omitempty"`
}
