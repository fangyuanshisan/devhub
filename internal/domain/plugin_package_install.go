package domain

// PluginPackageInstallRequest is used by POST /api/v1/admin/plugins/packages/install.
type PluginPackageInstallRequest struct {
	Path              string `json:"path"`
	ConfirmRiskLevel  string `json:"confirm_risk_level,omitempty"`
	AllowWithWarnings bool   `json:"allow_with_warnings,omitempty"`
	// Optional context: used internally for approvals/operations linking.
	ApprovalID  int64  `json:"approval_id,omitempty"`
	OperationID string `json:"operation_id,omitempty"`
}

type PluginPackageInstallResult struct {
	Installed          bool `json:"installed"`
	CreatedConfig      bool `json:"created_config,omitempty"`
	CreatedMigrations  int  `json:"created_migrations,omitempty"`
	CreatedPermissions int  `json:"created_permissions,omitempty"`
	CreatedMenus       int  `json:"created_menus,omitempty"`
	CreatedRoutes      int  `json:"created_routes,omitempty"`
}

type PluginPackageInstallResponse struct {
	Message       string                      `json:"message,omitempty"`
	OperationID   string                      `json:"operation_id,omitempty"`
	Plugin        Plugin                      `json:"plugin"`
	Package       PluginPackageInfo           `json:"package"`
	Checksum      PluginPackageChecksumResult `json:"checksum,omitempty"`
	RiskLevel     string                      `json:"risk_level,omitempty"`
	InstallResult PluginPackageInstallResult  `json:"install_result"`
	Warnings      []string                    `json:"warnings,omitempty"`
}
