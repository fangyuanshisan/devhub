package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"devhub-gin-backend/internal/domain"
)

func TestNormalizePluginPackagePath_AllowsExamplesPlugins(t *testing.T) {
	abs, clean, err := NormalizePluginPackagePath("examples/plugins/demo_notice")
	if err != nil {
		t.Fatalf("NormalizePluginPackagePath failed: %v", err)
	}
	if clean != filepath.Clean("examples/plugins/demo_notice") {
		t.Fatalf("unexpected clean path: %q", clean)
	}
	if abs == "" {
		t.Fatalf("expected abs path")
	}
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("expected abs path exists: %v", err)
	}
}

func TestNormalizePluginPackagePath_RejectsTraversal(t *testing.T) {
	_, _, err := NormalizePluginPackagePath("../examples/plugins/demo_notice")
	if err == nil {
		t.Fatalf("expected error")
	}
	if api, ok := err.(*domain.APIError); ok {
		if api.Code != "plugin_package_path_invalid" {
			t.Fatalf("expected code plugin_package_path_invalid, got %q", api.Code)
		}
	}
}

func TestScanPluginPackage_SafeExample(t *testing.T) {
	abs, _, err := NormalizePluginPackagePath("examples/plugins/demo_notice")
	if err != nil {
		t.Fatalf("NormalizePluginPackagePath failed: %v", err)
	}
	scan, err := ScanPluginPackage(abs)
	if err != nil {
		t.Fatalf("ScanPluginPackage failed: %v", err)
	}
	if scan.TotalFiles == 0 {
		t.Fatalf("expected files > 0")
	}
	if len(scan.DangerousFiles) != 0 {
		t.Fatalf("expected no dangerous files, got %#v", scan.DangerousFiles)
	}
}
