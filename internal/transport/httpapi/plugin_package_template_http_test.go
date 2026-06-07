package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestAdminPluginPackageTemplateHTTPPreviewGenerateExport(t *testing.T) {
	root := httpTestProjectRoot(t)
	code := "http_tpl_s9"
	draftDir := filepath.Join(root, "storage", "plugins", "drafts", code)
	_ = os.RemoveAll(draftDir)
	t.Cleanup(func() { _ = os.RemoveAll(draftDir) })

	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)
	payload := `{"code":"http_tpl_s9","name":"HTTP Template S9","plugin_type":"content","content_type":"http_tpl_s9_item","content_name":"模板内容","author":"DevHub Team","with_config":true,"with_hooks":true,"with_migration":true}`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/templates/preview", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("preview failed: %d %s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(draftDir); !os.IsNotExist(err) {
		t.Fatalf("preview should not write draft dir, stat err=%v", err)
	}
	var preview struct {
		Template struct {
			Code       string   `json:"code"`
			PluginType string   `json:"plugin_type"`
			Migrations []string `json:"migrations"`
			FileTree   []string `json:"file_tree"`
		} `json:"template"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if preview.Template.Code != code || preview.Template.PluginType != "content" || len(preview.Template.Migrations) != 1 {
		t.Fatalf("unexpected preview body: %s", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/templates", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("generate failed: %d %s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(filepath.Join(draftDir, "migrations", "001_init.sql")); err != nil {
		t.Fatalf("expected migrations/001_init.sql: %v", err)
	}
	if _, err := os.Stat(filepath.Join(draftDir, "001_schema.sql")); !os.IsNotExist(err) {
		t.Fatalf("must not generate root 001_schema.sql, stat err=%v", err)
	}

	_ = os.RemoveAll(draftDir)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/templates/export", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("export failed: %d %s %s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte("001_schema.sql")) || bytes.Contains(w.Body.Bytes(), []byte(".go")) || bytes.Contains(w.Body.Bytes(), []byte(".wasm")) {
		t.Fatalf("export contains forbidden content marker")
	}
}
