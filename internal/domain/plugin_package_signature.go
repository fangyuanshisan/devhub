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
	Version          string   `json:"version,omitempty"`
	Algorithm        string   `json:"algorithm,omitempty"`
	SignedAt         string   `json:"signed_at,omitempty"`
	PublisherID      string   `json:"publisher_id,omitempty"`
	PublicKeyID      string   `json:"public_key_id,omitempty"`
	PayloadAlgorithm string   `json:"payload_algorithm,omitempty"`
	Payload          string   `json:"payload,omitempty"`
	SignedFiles      []string `json:"signed_files,omitempty"`
	Signature        string   `json:"signature,omitempty"`
}

// PluginPackageSignatureResult is included in package dry-run/detail responses.
//
// trust_status: trusted|unknown|blocked|revoked|unsigned
// verification_status: verified|failed|missing|unsupported|publisher_unknown
type PluginPackageSignatureResult struct {
	SignatureFound     bool     `json:"signature_found"`
	PublisherFound     bool     `json:"publisher_found"`
	Algorithm          string   `json:"algorithm,omitempty"`
	Payload            string   `json:"payload,omitempty"`
	PayloadAlgorithm   string   `json:"payload_algorithm,omitempty"`
	PublisherID        string   `json:"publisher_id,omitempty"`
	PublicKeyID        string   `json:"public_key_id,omitempty"`
	Fingerprint        string   `json:"fingerprint,omitempty"`
	TrustStatus        string   `json:"trust_status,omitempty"`
	VerificationStatus string   `json:"verification_status,omitempty"`
	Verified           bool     `json:"verified"`
	SignedFiles        []string `json:"signed_files,omitempty"`
	SignedFilesCount   int      `json:"signed_files_count,omitempty"`
	UnsignedFiles      []string `json:"unsigned_files,omitempty"`
	Messages           []string `json:"messages,omitempty"`
}

type PluginTrustedPublisher struct {
	ID                 int64  `json:"id"`
	PublisherID        string `json:"publisher_id"`
	Name               string `json:"name"`
	Homepage           string `json:"homepage,omitempty"`
	Email              string `json:"email,omitempty"`
	PublicKeyID        string `json:"public_key_id"`
	PublicKeyAlgorithm string `json:"public_key_algorithm"`
	PublicKey          string `json:"public_key"`
	Fingerprint        string `json:"fingerprint"`
	Status             string `json:"status"`
	Notes              string `json:"notes,omitempty"`
	CreatedBy          int64  `json:"created_by,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedBy          int64  `json:"updated_by,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
	RevokedAt          string `json:"revoked_at,omitempty"`
	BlockedAt          string `json:"blocked_at,omitempty"`
	ExpiresAt          string `json:"expires_at,omitempty"`
	MetadataJSON       string `json:"metadata_json,omitempty"`
}

type PluginTrustedPublisherFilter struct {
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

type PluginTrustedPublisherListResponse struct {
	Items      []PluginTrustedPublisher `json:"items"`
	Pagination Pagination               `json:"pagination"`
	Summary    map[string]int           `json:"summary"`
}
