package domain

// SecretCenterStatusResponse is a read-only status view for Core SecretCenter.
//
// Security rules:
// - Never include plaintext secret values.
// - Never include encrypted_value.
// - Only expose ref/metadata counts and readiness hints.
type SecretCenterStatusResponse struct {
	Status          string         `json:"status"` // ok|warning|blocked
	SecretRefCount  int            `json:"secret_ref_count"`
	NamespaceCounts map[string]int `json:"namespace_counts,omitempty"`
	Notes           []string       `json:"notes,omitempty"`
}
