package domain

const (
	PluginPackageSignatureStatusPending              = "pending"
	PluginPackageSignatureStatusVerified             = "verified"
	PluginPackageSignatureStatusFailed               = "failed"
	PluginPackageSignatureStatusUnsigned             = "unsigned"
	PluginPackageSignatureStatusUntrustedPublisher   = "untrusted_publisher"
	PluginPackageSignatureStatusKeyRevokedOrDisabled = "key_revoked"
	PluginPackageSignatureStatusKeyExpired           = "key_expired"
	PluginPackageSignatureStatusHashMismatch         = "hash_mismatch"
	PluginPackageSignatureStatusPayloadMismatch      = "payload_mismatch"
	PluginPackageSignatureStatusAlgorithmUnsupported = "algorithm_unsupported"
	PluginPackageSignatureStatusDeleted              = "deleted"
)

// DevHubSignaturePayload is the canonical payload used for detached signature verification.
//
// Notes:
// - This payload is signed (canonical JSON bytes) with Ed25519.
// - Do not add map fields: keep a stable struct layout for canonical bytes.
type DevHubSignaturePayload struct {
	PluginCode            string `json:"plugin_code"`
	Version               string `json:"version"`
	PackageSHA256         string `json:"package_sha256"`
	ManifestSHA256        string `json:"manifest_sha256"`
	PublisherID           string `json:"publisher_id"`
	KeyID                 string `json:"key_id"`
	CompatibleCoreVersion string `json:"compatible_core_version,omitempty"`
}

// DevHubDetachedSignatureFile describes devhub-signature.json (detached signature metadata).
type DevHubDetachedSignatureFile struct {
	SchemaVersion    string                 `json:"schema_version"`
	Algorithm        string                 `json:"algorithm"`
	PublisherID      string                 `json:"publisher_id"`
	KeyID            string                 `json:"key_id"`
	PluginCode       string                 `json:"plugin_code"`
	Version          string                 `json:"version"`
	PackageSHA256    string                 `json:"package_sha256"`
	ManifestSHA256   string                 `json:"manifest_sha256"`
	Signature        string                 `json:"signature"`
	SignedAt         string                 `json:"signed_at,omitempty"`
	SignaturePayload DevHubSignaturePayload `json:"signature_payload"`
}

type PluginPackageSignatureRecord struct {
	ID                   int64  `json:"id"`
	PackageDownloadID    int64  `json:"package_download_id,omitempty"`
	PackagePrecheckID    int64  `json:"package_precheck_id,omitempty"`
	PackageCompatID      int64  `json:"package_compat_check_id,omitempty"`
	PluginCode           string `json:"plugin_code"`
	Version              string `json:"version"`
	PublisherID          string `json:"publisher_id,omitempty"`
	KeyID                string `json:"key_id,omitempty"`
	Algorithm            string `json:"algorithm,omitempty"`
	Status               string `json:"status"`
	SignatureURL         string `json:"signature_url,omitempty"`
	SignatureFilePath    string `json:"signature_file_path,omitempty"`
	PackageSHA256        string `json:"package_sha256,omitempty"`
	ManifestSHA256       string `json:"manifest_sha256,omitempty"`
	SignaturePayloadJSON string `json:"signature_payload_json,omitempty"`
	SignatureBase64      string `json:"signature_base64,omitempty"`
	VerifiedAt           string `json:"verified_at,omitempty"`
	ErrorMessage         string `json:"error_message,omitempty"`
	WarningsJSON         string `json:"warnings_json,omitempty"`
	CreatedBy            int64  `json:"created_by,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
}

type PluginPackageSignatureFilter struct {
	Status            string
	PluginCode        string
	PackageDownloadID int64
	PackagePrecheckID int64
	Keyword           string
	Page              int
	PageSize          int
}

type PluginPackageSignatureListResponse struct {
	Items      []PluginPackageSignatureRecord `json:"items"`
	Pagination Pagination                     `json:"pagination"`
	Summary    map[string]int                 `json:"summary"`
}
