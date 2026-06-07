package domain

// PluginPackageTemplateRequest is used by admin plugin package template preview/create APIs.
type PluginPackageTemplateRequest struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	PluginType      string `json:"plugin_type"`
	ContentType     string `json:"content_type"`
	ContentName     string `json:"content_name"`
	Description     string `json:"description"`
	Author          string `json:"author"`
	PublisherMode   string `json:"publisher_mode,omitempty"`
	PublisherID     string `json:"publisher_id,omitempty"`
	MountPoint      string `json:"mount_point,omitempty"`
	ComponentKey    string `json:"component_key,omitempty"`
	HealthCheckPath string `json:"health_check_path,omitempty"`
	TimeoutMS       int    `json:"timeout_ms,omitempty"`
	FailurePolicy   string `json:"failure_policy,omitempty"`
	WithConfig      bool   `json:"with_config"`
	WithHooks       bool   `json:"with_hooks"`
	WithMigration   bool   `json:"with_migration"`
}

type PluginPackageTemplatePreview struct {
	Code            string                          `json:"code"`
	Name            string                          `json:"name"`
	PluginType      string                          `json:"plugin_type"`
	ContentType     string                          `json:"content_type,omitempty"`
	ContentName     string                          `json:"content_name,omitempty"`
	Description     string                          `json:"description"`
	Author          string                          `json:"author"`
	OutputDir       string                          `json:"output_dir"`
	PackagePath     string                          `json:"package_path"`
	Files           []string                        `json:"files"`
	FileTree        []string                        `json:"file_tree,omitempty"`
	Permissions     []string                        `json:"permissions,omitempty"`
	Menus           []string                        `json:"menus,omitempty"`
	Hooks           []string                        `json:"hooks,omitempty"`
	Migrations      []string                        `json:"migrations,omitempty"`
	FrontendMounts  []FrontendMountDefinition       `json:"frontend_mounts,omitempty"`
	ExternalService *PluginExternalService          `json:"external_service,omitempty"`
	Generated       map[string]string               `json:"generated,omitempty"`
	Summary         map[string]any                  `json:"summary,omitempty"`
	Conflicts       []PluginPackageTemplateConflict `json:"conflicts,omitempty"`
	WillOverwrite   bool                            `json:"will_overwrite"`
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

type PluginPackageTemplateConflict struct {
	Field      string `json:"field"`
	Value      string `json:"value"`
	Target     string `json:"target,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}
