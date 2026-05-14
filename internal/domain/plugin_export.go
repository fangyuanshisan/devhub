package domain

// PluginPackageExportRequest is used by plugin package export dry-run/export APIs.
type PluginPackageExportRequest struct {
	IncludeDocs          bool   `json:"include_docs"`
	IncludeMigrations    bool   `json:"include_migrations"`
	IncludePublisher     bool   `json:"include_publisher"`
	IncludeSignatureStub bool   `json:"include_signature_stub"`
	Force                bool   `json:"force,omitempty"`
	OutputDir            string `json:"output_dir,omitempty"`
}

type PluginPackageExportPreview struct {
	Files                   []string `json:"files"`
	OutputDir               string   `json:"output_dir"`
	ContainsSensitiveValues bool     `json:"contains_sensitive_values"`
	ContainsUserData        bool     `json:"contains_user_data"`
	ContainsRuntimeCode     bool     `json:"contains_runtime_code"`
	ContainsExternalSQL     bool     `json:"contains_external_sql"`
}

type PluginPackageExportDryRunResponse struct {
	PluginCode    string                     `json:"plugin_code"`
	Version       string                     `json:"version"`
	Status        string                     `json:"status"`
	ExportPreview PluginPackageExportPreview `json:"export_preview"`
	Warnings      []string                   `json:"warnings,omitempty"`
	Errors        []string                   `json:"errors,omitempty"`
}

type PluginPackageExportResponse struct {
	Message             string   `json:"message,omitempty"`
	PluginCode          string   `json:"plugin_code"`
	Version             string   `json:"version"`
	OutputDir           string   `json:"output_dir"`
	Files               []string `json:"files"`
	ChecksumStatus      string   `json:"checksum_status"`
	PackageDryRunStatus string   `json:"package_dry_run_status"`
	Warnings            []string `json:"warnings,omitempty"`
}
