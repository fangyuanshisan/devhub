package plugins

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
)

func TestNormalizeZipEntryNameRejectsZipSlip(t *testing.T) {
	cases := []string{
		"../manifest.json",
		"/etc/passwd",
		`C:\Windows\system32\evil`,
		"dir/../../evil",
		"%2e%2e/manifest.json",
	}
	for _, tc := range cases {
		if _, err := normalizeZipEntryName(tc); err == nil {
			t.Fatalf("expected %q to be rejected", tc)
		}
	}
}

func TestExtractPluginPackageZip_BlocksNestedArchiveAndSymlink(t *testing.T) {
	tests := []struct {
		name string
		make func(t *testing.T, zw *zip.Writer)
		code string
	}{
		{
			name: "nested_zip",
			make: func(t *testing.T, zw *zip.Writer) {
				addZipFile(t, zw, "manifest.json", []byte(`{}`))
				addZipFile(t, zw, "assets/inner.zip", []byte("nested"))
			},
			code: "plugin_package_zip_nested_archive_forbidden",
		},
		{
			name: "symlink",
			make: func(t *testing.T, zw *zip.Writer) {
				addZipFile(t, zw, "manifest.json", []byte(`{}`))
				h := &zip.FileHeader{Name: "link"}
				h.SetMode(os.ModeSymlink | 0o777)
				w, err := zw.CreateHeader(h)
				if err != nil {
					t.Fatalf("CreateHeader: %v", err)
				}
				_, _ = w.Write([]byte("manifest.json"))
			},
			code: "plugin_package_zip_symlink_forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zipPath := writeTestZip(t, tt.make)
			staging := t.TempDir()
			_, err := ExtractPluginPackageZip(zipPath, staging, "storage/plugins/staging/test")
			if err == nil {
				t.Fatalf("expected blocked")
			}
			if apiCode(err) != tt.code {
				t.Fatalf("unexpected code: %s (%v)", apiCode(err), err)
			}
		})
	}
}

func TestExtractPluginPackageZip_Limits(t *testing.T) {
	t.Run("too_many_files", func(t *testing.T) {
		zipPath := writeTestZip(t, func(t *testing.T, zw *zip.Writer) {
			addZipFile(t, zw, "manifest.json", []byte(`{}`))
			for i := 0; i < PluginPackageUploadMaxFiles+1; i++ {
				addZipFile(t, zw, filepath.ToSlash(filepath.Join("docs", "f"+string(rune('a'+i%26))+"_"+string(rune('a'+(i/26)%26))+".md")), []byte("x"))
			}
		})
		_, err := ExtractPluginPackageZip(zipPath, t.TempDir(), "storage/plugins/staging/test")
		if apiCode(err) != "plugin_package_zip_too_many_files" {
			t.Fatalf("unexpected code: %s (%v)", apiCode(err), err)
		}
	})

	t.Run("single_file_too_large", func(t *testing.T) {
		zipPath := writeTestZip(t, func(t *testing.T, zw *zip.Writer) {
			addZipFile(t, zw, "manifest.json", []byte(`{}`))
			addZipFile(t, zw, "assets/big.bin", bytes.Repeat([]byte("a"), int(PluginPackageUploadMaxFileSize+1)))
		})
		_, err := ExtractPluginPackageZip(zipPath, t.TempDir(), "storage/plugins/staging/test")
		if apiCode(err) != "plugin_package_zip_file_too_large" {
			t.Fatalf("unexpected code: %s (%v)", apiCode(err), err)
		}
	})

	t.Run("total_size_exceeded", func(t *testing.T) {
		zipPath := writeTestZip(t, func(t *testing.T, zw *zip.Writer) {
			addZipFile(t, zw, "manifest.json", []byte(`{}`))
			payload := bytes.Repeat([]byte("a"), int(PluginPackageUploadMaxFileSize))
			for i := 0; i < int(PluginPackageUploadMaxTotalSize/PluginPackageUploadMaxFileSize)+1; i++ {
				addZipFile(t, zw, filepath.ToSlash(filepath.Join("assets", "part_"+string(rune('a'+i))+".bin")), payload)
			}
		})
		_, err := ExtractPluginPackageZip(zipPath, t.TempDir(), "storage/plugins/staging/test")
		if apiCode(err) != "plugin_package_zip_total_size_exceeded" {
			t.Fatalf("unexpected code: %s (%v)", apiCode(err), err)
		}
	})
}

func writeTestZip(t *testing.T, fill func(t *testing.T, zw *zip.Writer)) string {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fill(t, zw)
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	path := filepath.Join(t.TempDir(), "pkg.zip")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}
	return path
}

func addZipFile(t *testing.T, zw *zip.Writer, name string, raw []byte) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("zip create %s: %v", name, err)
	}
	if _, err := w.Write(raw); err != nil {
		t.Fatalf("zip write %s: %v", name, err)
	}
}

func apiCode(err error) string {
	if api, ok := err.(*domain.APIError); ok && api != nil {
		return api.Code
	}
	return ""
}
