package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
)

func mustProjectRootForPlugins(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "VERSION")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("unable to locate project root")
	return ""
}

func TestVerifyPluginPackageChecksums_Missing(t *testing.T) {
	root := mustProjectRootForPlugins(t)
	dir := filepath.Join(root, "examples", "plugins", "security-fixtures", "no_checksums")
	scan, err := ScanPluginPackage(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	got, err := VerifyPluginPackageChecksums(dir, scan)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Status != "missing" {
		t.Fatalf("expected missing, got %#v", got)
	}
	if len(got.Warnings) == 0 {
		t.Fatalf("expected warnings when checksums.json is missing")
	}
}

func TestVerifyPluginPackageChecksums_OK(t *testing.T) {
	root := mustProjectRootForPlugins(t)
	dir := filepath.Join(root, "examples", "plugins", "demo_notice")
	scan, err := ScanPluginPackage(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	got, err := VerifyPluginPackageChecksums(dir, scan)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Status != "ok" {
		t.Fatalf("expected ok, got %#v", got)
	}
	if len(got.Matched) == 0 {
		t.Fatalf("expected matched files")
	}
	if len(got.Mismatched) != 0 {
		t.Fatalf("expected no mismatched")
	}
}

func TestVerifyPluginPackageChecksums_MismatchBlocked(t *testing.T) {
	root := mustProjectRootForPlugins(t)
	dir := filepath.Join(root, "examples", "plugins", "security-fixtures", "checksum_mismatch")
	scan, err := ScanPluginPackage(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	got, err := VerifyPluginPackageChecksums(dir, scan)
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_checksum_mismatch" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
	if got.Status != "failed" {
		t.Fatalf("expected failed status, got %#v", got)
	}
}

func TestVerifyPluginPackageChecksums_UnsupportedAlgorithmBlocked(t *testing.T) {
	root := mustProjectRootForPlugins(t)
	dir := filepath.Join(root, ".devhub", "plugins", "pkg_checksum_algo")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, ".devhub")) })

	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"code":"pkg_checksum_algo","name":"x","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	wire := map[string]any{
		"algorithm": "md5",
		"files":     []map[string]any{{"path": "manifest.json", "sha256": "deadbeef"}},
	}
	raw, _ := json.Marshal(wire)
	if err := os.WriteFile(filepath.Join(dir, "checksums.json"), raw, 0o644); err != nil {
		t.Fatalf("write checksums: %v", err)
	}
	scan, err := ScanPluginPackage(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	_, err = VerifyPluginPackageChecksums(dir, scan)
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_checksum_unsupported_algorithm" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}

func TestVerifyPluginPackageChecksums_InvalidJSONBlocked(t *testing.T) {
	root := mustProjectRootForPlugins(t)
	dir := filepath.Join(root, ".devhub", "plugins", "pkg_checksum_badjson")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, ".devhub")) })

	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"code":"pkg_checksum_badjson","name":"x","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "checksums.json"), []byte("{"), 0o644); err != nil {
		t.Fatalf("write checksums: %v", err)
	}
	scan, err := ScanPluginPackage(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	_, err = VerifyPluginPackageChecksums(dir, scan)
	if err == nil {
		t.Fatalf("expected error")
	}
	api, ok := err.(*domain.APIError)
	if !ok || api == nil {
		t.Fatalf("expected APIError, got %T %v", err, err)
	}
	if api.Code != "plugin_package_checksum_invalid" {
		t.Fatalf("unexpected code: %q", api.Code)
	}
}
