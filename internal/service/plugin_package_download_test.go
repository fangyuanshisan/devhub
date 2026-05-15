package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestDownloadPluginPackageToStaging_SHA256OK(t *testing.T) {
	stagingRoot := cleanupDownloadStorage(t)
	raw := []byte("devhub plugin zip fixture")
	sum := sha256Hex(raw)
	server := testTLSServer(t, raw)

	repo := store.NewMemoryStore()
	svc := New(repo)
	record, err := svc.DownloadPluginPackageToStagingAs(PluginPackageDownloadOperator{ID: 7}, domain.PluginPackageDownloadRequest{
		PluginCode: "demo_notice",
		Version:    "1.0.0",
		PackageURL: server.URL + "/demo_notice-1.0.0.zip",
		SHA256:     sum,
	})
	if err != nil {
		t.Fatalf("DownloadPluginPackageToStagingAs: %v", err)
	}
	if record.Status != domain.PluginPackageDownloadStatusDownloaded {
		t.Fatalf("expected downloaded, got %q", record.Status)
	}
	if record.SHA256Actual != sum || record.FileSize != int64(len(raw)) {
		t.Fatalf("unexpected hash/size: %#v", record)
	}
	if !strings.HasPrefix(record.StagingPath, stagingRoot+"/") {
		t.Fatalf("expected staging path under %s, got %q", stagingRoot, record.StagingPath)
	}
	if _, err := os.Stat(filepath.Join(mustProjectRoot(t), filepath.FromSlash(record.StagingPath))); err != nil {
		t.Fatalf("expected staging file: %v", err)
	}
	if _, ok := repo.PluginByCode("demo_notice"); ok {
		t.Fatalf("download must not install plugin")
	}
}

func TestDownloadPluginPackageToStaging_RejectsUnsafeURL(t *testing.T) {
	cleanupDownloadStorageNoLocal(t)
	svc := New(store.NewMemoryStore())
	cases := []string{
		"http://example.com/demo.zip",
		"file:///tmp/demo.zip",
		"https://localhost/demo.zip",
		"https://127.0.0.1/demo.zip",
		"https://10.0.0.2/demo.zip",
	}
	for _, rawURL := range cases {
		_, err := svc.DownloadPluginPackageToStagingAs(PluginPackageDownloadOperator{}, domain.PluginPackageDownloadRequest{
			PluginCode: "demo_notice",
			Version:    "1.0.0",
			PackageURL: rawURL,
			SHA256:     strings.Repeat("a", 64),
		})
		if err == nil {
			t.Fatalf("expected rejection for %s", rawURL)
		}
		if code := apiCodeForUpload(err); code != "plugin_package_download_url_invalid" && code != "plugin_package_download_url_forbidden" {
			t.Fatalf("unexpected code for %s: %s", rawURL, code)
		}
	}
}

func TestDownloadPluginPackageToStaging_ChecksumMismatchCleansFile(t *testing.T) {
	stagingRoot := cleanupDownloadStorage(t)
	raw := []byte("devhub plugin zip fixture")
	server := testTLSServer(t, raw)

	svc := New(store.NewMemoryStore())
	_, err := svc.DownloadPluginPackageToStagingAs(PluginPackageDownloadOperator{}, domain.PluginPackageDownloadRequest{
		PluginCode: "demo_notice",
		Version:    "1.0.0",
		PackageURL: server.URL + "/demo_notice-1.0.0.zip",
		SHA256:     strings.Repeat("b", 64),
	})
	if err == nil || apiCodeForUpload(err) != "plugin_package_download_checksum_mismatch" {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
	entries, _ := os.ReadDir(filepath.Join(mustProjectRoot(t), filepath.FromSlash(stagingRoot)))
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".part") {
			t.Fatalf("expected no finalized staging file after checksum mismatch, found %s", entry.Name())
		}
	}
}

func TestDownloadPluginPackageToStaging_NoSHA256IsChecksumMissing(t *testing.T) {
	cleanupDownloadStorage(t)
	raw := []byte("devhub plugin zip fixture")
	server := testTLSServer(t, raw)
	svc := New(store.NewMemoryStore())
	record, err := svc.DownloadPluginPackageToStagingAs(PluginPackageDownloadOperator{}, domain.PluginPackageDownloadRequest{
		PluginCode: "demo_notice",
		Version:    "1.0.0",
		PackageURL: server.URL + "/demo_notice-1.0.0.zip",
	})
	if err != nil {
		t.Fatalf("DownloadPluginPackageToStagingAs: %v", err)
	}
	if record.Status != domain.PluginPackageDownloadStatusChecksumMissing {
		t.Fatalf("expected checksum_missing, got %q", record.Status)
	}
}

func cleanupDownloadStorage(t *testing.T) string {
	t.Helper()
	root := mustProjectRoot(t)
	staging := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-download-staging")
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_STAGING_ROOT", staging)
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_ALLOW_LOCAL", "1")
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_INSECURE_TLS", "1")
	_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(staging)))
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(staging)))
	})
	return staging
}

func cleanupDownloadStorageNoLocal(t *testing.T) string {
	t.Helper()
	root := mustProjectRoot(t)
	staging := ensureWritableTestStorageDir(t, "storage/plugins/packages/test-download-staging")
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_DOWNLOAD_STAGING_ROOT", staging)
	_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(staging)))
	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(staging)))
	})
	return staging
}

func testTLSServer(t *testing.T, raw []byte) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		_, _ = w.Write(raw)
	}))
	t.Cleanup(server.Close)
	return server
}
