package domain

// PluginPackagePublisher describes publisher.json metadata inside a plugin package.
// Notes:
// - publisher.json is never a trust source by itself; trust is determined by local trusted_publishers config.
// - Never store or return any private key.
type PluginPackagePublisher struct {
	PublisherID        string `json:"publisher_id"`
	Name               string `json:"name,omitempty"`
	Homepage           string `json:"homepage,omitempty"`
	Email              string `json:"email,omitempty"`
	PublicKeyID        string `json:"public_key_id,omitempty"`
	PublicKeyAlgorithm string `json:"public_key_algorithm,omitempty"`
	PublicKey          string `json:"public_key,omitempty"`
	TrustLevel         string `json:"trust_level,omitempty"`
}

// PluginPackageSignature describes signature.json metadata inside a plugin package.
type PluginPackageSignature struct {
	Version     string   `json:"version,omitempty"`
	Algorithm   string   `json:"algorithm,omitempty"`
	SignedAt    string   `json:"signed_at,omitempty"`
	PublisherID string   `json:"publisher_id,omitempty"`
	PublicKeyID string   `json:"public_key_id,omitempty"`
	SignedFiles []string `json:"signed_files,omitempty"`
	Signature   string   `json:"signature,omitempty"`
}

// PluginPackageSignatureResult is included in package dry-run/detail responses.
//
// trust_status: trusted|unknown|blocked|revoked|unsigned
// verification_status: verified|failed|missing|unsupported|structural_only
type PluginPackageSignatureResult struct {
	SignatureFound     bool     `json:"signature_found"`
	PublisherFound     bool     `json:"publisher_found"`
	Algorithm          string   `json:"algorithm,omitempty"`
	PublisherID        string   `json:"publisher_id,omitempty"`
	PublicKeyID        string   `json:"public_key_id,omitempty"`
	TrustStatus        string   `json:"trust_status,omitempty"`
	VerificationStatus string   `json:"verification_status,omitempty"`
	SignedFiles        []string `json:"signed_files,omitempty"`
	SignedFilesCount   int      `json:"signed_files_count,omitempty"`
	UnsignedFiles      []string `json:"unsigned_files,omitempty"`
	Messages           []string `json:"messages,omitempty"`
}
