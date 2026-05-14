package domain

// PluginPackageTemplateRequest is used by admin plugin package template preview/create APIs.
type PluginPackageTemplateRequest struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	ContentType   string `json:"content_type"`
	ContentName   string `json:"content_name"`
	Description   string `json:"description"`
	Author        string `json:"author"`
	WithConfig    bool   `json:"with_config"`
	WithHooks     bool   `json:"with_hooks"`
	WithMigration bool   `json:"with_migration"`
}

type PluginPackageTemplatePreview struct {
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	ContentType   string   `json:"content_type"`
	ContentName   string   `json:"content_name"`
	Description   string   `json:"description"`
	Author        string   `json:"author"`
	OutputDir     string   `json:"output_dir"`
	PackagePath   string   `json:"package_path"`
	Files         []string `json:"files"`
	WillOverwrite bool     `json:"will_overwrite"`
}

type PluginPackageTemplatePreviewResponse struct {
	Template PluginPackageTemplatePreview `json:"template"`
	Status   string                       `json:"status"`
	Warnings []string                     `json:"warnings,omitempty"`
	Errors   []string                     `json:"errors,omitempty"`
}

type PluginPackageTemplateCreateResponse struct {
	Message  string                       `json:"message,omitempty"`
	Template PluginPackageTemplatePreview `json:"template"`
	DryRun   PluginPackageDryRunResult    `json:"dry_run"`
	Status   string                       `json:"status"`
	Warnings []string                     `json:"warnings,omitempty"`
	Errors   []string                     `json:"errors,omitempty"`
}
