package domain

const (
	PluginRemoteIndexStatusEnabled  = "enabled"
	PluginRemoteIndexStatusDisabled = "disabled"
)

type PluginRemoteIndexSource struct {
	ID               int64  `json:"id"`
	SourceID         string `json:"source_id"`
	Name             string `json:"name"`
	IndexURL         string `json:"index_url"`
	Homepage         string `json:"homepage,omitempty"`
	Description      string `json:"description,omitempty"`
	Status           string `json:"status"`
	TrustPolicy      string `json:"trust_policy,omitempty"`
	LastFetchStatus  string `json:"last_fetch_status,omitempty"`
	LastFetchAt      string `json:"last_fetch_at,omitempty"`
	LastErrorCode    string `json:"last_error_code,omitempty"`
	LastErrorMessage string `json:"last_error_message,omitempty"`
	LastIndexHash    string `json:"last_index_hash,omitempty"`
	MetadataJSON     string `json:"metadata_json,omitempty"`
	CreatedBy        int64  `json:"created_by,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	UpdatedBy        int64  `json:"updated_by,omitempty"`
	UpdatedAt        string `json:"updated_at,omitempty"`
}

type PluginRemoteIndexFilter struct {
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

type PluginRemoteIndexSummary struct {
	Total    int `json:"total"`
	Enabled  int `json:"enabled"`
	Disabled int `json:"disabled"`
	Failed   int `json:"failed"`
}

type PluginRemoteIndexListResponse struct {
	Items      []PluginRemoteIndexSource `json:"items"`
	Pagination Pagination                `json:"pagination"`
	Summary    PluginRemoteIndexSummary  `json:"summary"`
}

type PluginRemoteIndexSourceMeta struct {
	SourceID    string `json:"source_id"`
	Name        string `json:"name"`
	Homepage    string `json:"homepage,omitempty"`
	Description string `json:"description,omitempty"`
}

type PluginRemoteIndexDocument struct {
	SchemaVersion string                       `json:"schema_version"`
	GeneratedAt   string                       `json:"generated_at"`
	Source        PluginRemoteIndexSourceMeta  `json:"source"`
	Plugins       []PluginRemoteIndexPluginDoc `json:"plugins"`
}

type PluginRemoteIndexPluginDoc struct {
	Code          string                        `json:"code"`
	Name          string                        `json:"name"`
	Description   string                        `json:"description,omitempty"`
	LatestVersion string                        `json:"latest_version"`
	Versions      []PluginRemoteIndexVersionDoc `json:"versions"`
}

type PluginRemoteIndexVersionDoc struct {
	Version               string   `json:"version"`
	MinCoreVersion        string   `json:"min_core_version,omitempty"`
	CompatibleCoreVersion string   `json:"compatible_core_version,omitempty"`
	PackageURL            string   `json:"package_url"`
	PackageSHA256         string   `json:"package_sha256"`
	ManifestSHA256        string   `json:"manifest_sha256,omitempty"`
	SignatureURL          string   `json:"signature_url,omitempty"`
	PublisherID           string   `json:"publisher_id"`
	PublicKeyID           string   `json:"public_key_id"`
	License               string   `json:"license,omitempty"`
	Tags                  []string `json:"tags,omitempty"`
	CreatedAt             string   `json:"created_at,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
}

type PluginRemoteIndexValidation struct {
	Valid    bool                    `json:"valid"`
	Warnings []PluginPackageRiskItem `json:"warnings,omitempty"`
	Errors   []PluginPackageRiskItem `json:"errors,omitempty"`
}

type PluginRemoteIndexFetchResponse struct {
	Source     PluginRemoteIndexSource     `json:"source"`
	Document   PluginRemoteIndexDocument   `json:"document"`
	Validation PluginRemoteIndexValidation `json:"validation"`
	IndexHash  string                      `json:"index_hash"`
	Warnings   []PluginPackageRiskItem     `json:"warnings,omitempty"`
	Errors     []PluginPackageRiskItem     `json:"errors,omitempty"`
}

type PluginRemotePluginListItem struct {
	Code                  string                  `json:"code"`
	Name                  string                  `json:"name"`
	Description           string                  `json:"description,omitempty"`
	LatestVersion         string                  `json:"latest_version"`
	PublisherID           string                  `json:"publisher_id,omitempty"`
	PublicKeyID           string                  `json:"public_key_id,omitempty"`
	License               string                  `json:"license,omitempty"`
	MinCoreVersion        string                  `json:"min_core_version,omitempty"`
	CompatibleCoreVersion string                  `json:"compatible_core_version,omitempty"`
	PackageSHA256         string                  `json:"package_sha256,omitempty"`
	SignatureURL          string                  `json:"signature_url,omitempty"`
	Installed             bool                    `json:"installed"`
	LocalVersion          string                  `json:"local_version,omitempty"`
	InstalledStatus       string                  `json:"installed_status,omitempty"`
	VersionStatus         string                  `json:"version_status"`
	CoreCompatibility     PluginCoreCompatibility `json:"core_compatibility"`
	PublisherTrustStatus  string                  `json:"publisher_trust_status"`
	RiskLevel             string                  `json:"risk_level"`
	RiskSummary           string                  `json:"risk_summary"`
	RiskItems             []PluginPackageRiskItem `json:"risk_items,omitempty"`
}

type PluginRemotePluginListResponse struct {
	Items      []PluginRemotePluginListItem `json:"items"`
	Pagination Pagination                   `json:"pagination"`
	Summary    map[string]int               `json:"summary"`
}

type PluginRemoteVersionDetail struct {
	PluginRemoteIndexVersionDoc
	CoreCompatibility    PluginCoreCompatibility `json:"core_compatibility"`
	PublisherTrustStatus string                  `json:"publisher_trust_status"`
	RiskLevel            string                  `json:"risk_level"`
	RiskItems            []PluginPackageRiskItem `json:"risk_items,omitempty"`
}

type PluginRemotePluginDetailResponse struct {
	Source          PluginRemoteIndexSource     `json:"source"`
	Plugin          PluginRemoteIndexPluginDoc  `json:"plugin"`
	Versions        []PluginRemoteVersionDetail `json:"versions"`
	Installed       bool                        `json:"installed"`
	LocalVersion    string                      `json:"local_version,omitempty"`
	InstalledStatus string                      `json:"installed_status,omitempty"`
	Readonly        bool                        `json:"readonly"`
	ReadonlyMessage string                      `json:"readonly_message"`
}
