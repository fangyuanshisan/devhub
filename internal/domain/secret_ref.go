package domain

// SecretRefStatus is the lifecycle status of a secret reference.
//
// Values are intentionally simple: active/disabled/revoked.
// Secrets are never deleted automatically because refs should remain stable for audit/debug.
const (
	SecretRefStatusActive   = "active"
	SecretRefStatusDisabled = "disabled"
	SecretRefStatusRevoked  = "revoked"
)

// SecretRefRecord is the persisted SecretCenter record.
//
// Security:
// - EncryptedValue must never be exposed via JSON/API.
type SecretRefRecord struct {
	ID             int64  `json:"id,omitempty"`
	Ref            string `json:"ref"`
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name,omitempty"`
	Type           string `json:"type,omitempty"`
	Usage          string `json:"usage,omitempty"`
	UsageType      string `json:"usage_type,omitempty"`
	SourceType     string `json:"source_type,omitempty"`
	SourceID       string `json:"source_id,omitempty"`
	SourceCode     string `json:"source_code,omitempty"`
	AssociatedWith string `json:"associated_with,omitempty"`
	MaskedValue    string `json:"masked_value,omitempty"`
	Available      bool   `json:"available"`
	KeyID          string `json:"key_id,omitempty"`
	EncryptedValue string `json:"-"`
	Status         string `json:"status,omitempty"`
	Description    string `json:"description,omitempty"`
	LastUsedAt     string `json:"last_used_at,omitempty"`
	UsageCount     int    `json:"usage_count,omitempty"`
	RotatedAt      string `json:"rotated_at,omitempty"`
	CreatedBy      int64  `json:"created_by,omitempty"`
	UpdatedBy      int64  `json:"updated_by,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

type SecretRefFilter struct {
	Namespace string `json:"namespace"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

func (f SecretRefFilter) Normalize() SecretRefFilter {
	out := f
	if out.Page <= 0 {
		out.Page = 1
	}
	if out.PageSize <= 0 || out.PageSize > 100 {
		out.PageSize = 20
	}
	return out
}

type SecretRefListResponse struct {
	Items      []SecretRefRecord `json:"items"`
	Pagination Pagination        `json:"pagination"`
}

type SecretCreateRequest struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

type SecretUpdateRequest struct {
	Ref   string `json:"ref"`
	Value string `json:"value"`
}

type SecretStatusRequest struct {
	Ref           string `json:"ref"`
	ConfirmRef    string `json:"confirm_ref,omitempty"`
	StrongConfirm bool   `json:"strong_confirm,omitempty"`
}

type SecretSourceInfo struct {
	Type                 string `json:"type"`
	Label                string `json:"label"`
	SourceID             string `json:"source_id,omitempty"`
	SourceCode           string `json:"source_code,omitempty"`
	PluginCode           string `json:"plugin_code,omitempty"`
	PluginName           string `json:"plugin_name,omitempty"`
	ConfigEntry          string `json:"config_entry,omitempty"`
	ManagementPage       string `json:"management_page,omitempty"`
	ManagementQueryHint  string `json:"management_query_hint,omitempty"`
	TestData             bool   `json:"test_data"`
	CanJump              bool   `json:"can_jump"`
	JumpDisabledReason   string `json:"jump_disabled_reason,omitempty"`
	RotationTarget       string `json:"rotation_target,omitempty"`
	RotationDisabledNote string `json:"rotation_disabled_note,omitempty"`
}

type SecretUsageRelationship struct {
	Type              string `json:"type"`
	Label             string `json:"label"`
	UsageType         string `json:"usage_type,omitempty"`
	SourceType        string `json:"source_type,omitempty"`
	SourceID          string `json:"source_id,omitempty"`
	SourceCode        string `json:"source_code,omitempty"`
	PluginCode        string `json:"plugin_code,omitempty"`
	PluginName        string `json:"plugin_name,omitempty"`
	ServiceName       string `json:"service_name,omitempty"`
	EndpointURL       string `json:"endpoint_url,omitempty"`
	Enabled           bool   `json:"enabled,omitempty"`
	CurrentHealth     string `json:"current_health,omitempty"`
	LastSuccessAt     string `json:"last_success_at,omitempty"`
	LastFailureAt     string `json:"last_failure_at,omitempty"`
	ConfigEntry       string `json:"config_entry,omitempty"`
	ManagementPage    string `json:"management_page,omitempty"`
	Status            string `json:"status,omitempty"`
	SecretRef         string `json:"secret_ref,omitempty"`
	TargetURL         string `json:"target_url,omitempty"`
	TokenRef          string `json:"token_ref,omitempty"`
	Unresolved        bool   `json:"unresolved,omitempty"`
	UnresolvedMessage string `json:"unresolved_message,omitempty"`
}

type SecretImpactPreview struct {
	Ref                        string                    `json:"ref"`
	Action                     string                    `json:"action"`
	Allowed                    bool                      `json:"allowed"`
	Status                     string                    `json:"status,omitempty"`
	Message                    string                    `json:"message,omitempty"`
	Warning                    string                    `json:"warning,omitempty"`
	RequiresStrongConfirmation bool                      `json:"requires_strong_confirmation,omitempty"`
	ConfirmationText           string                    `json:"confirmation_text,omitempty"`
	AffectedPlugins            int                       `json:"affected_plugins"`
	AffectedExternalServices   int                       `json:"affected_external_services"`
	AffectedWebhooks           int                       `json:"affected_webhooks"`
	AffectedCallbacks          int                       `json:"affected_callbacks"`
	PossibleFailedHealthChecks int                       `json:"possible_failed_health_checks"`
	PossibleFailedDeliveries   int                       `json:"possible_failed_deliveries"`
	UsageCountTotal            int                       `json:"usage_count_total,omitempty"`
	UsageCountLast24h          int                       `json:"usage_count_last_24h,omitempty"`
	UsageCountLast7d           int                       `json:"usage_count_last_7d,omitempty"`
	AffectedBusiness           []SecretUsageRelationship `json:"affected_business,omitempty"`
	Notes                      []string                  `json:"notes,omitempty"`
}

type SecretDetailResponse struct {
	Record      SecretRefRecord           `json:"record"`
	Source      SecretSourceInfo          `json:"source"`
	Usages      []SecretUsageRelationship `json:"usages"`
	Preview     *SecretImpactPreview      `json:"preview,omitempty"`
	SafetyNotes []string                  `json:"safety_notes,omitempty"`
}
