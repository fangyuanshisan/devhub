package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestAdminPluginPackageUploadHTTP_AuthAndAudit(t *testing.T) {
	cleanupHTTPUploadStorage(t)
	router := NewRouter(service.New(store.NewMemoryStore()))

	user := userToken(t, router, "admin")
	body, contentType := multipartUploadBody(t, "demo.zip", makeHTTPUploadZip(t, "http_upload_auth"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/packages/upload", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+user)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Fatalf("expected user token rejected, got %d: %s", w.Code, w.Body.String())
	}

	admin := adminToken(t, router)
	body, contentType = multipartUploadBody(t, "demo.txt", []byte("not zip"))
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/packages/upload", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid type 400, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("plugin_package_upload_invalid_type")) {
		t.Fatalf("expected structured code, got %s", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs", nil)
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected audit logs, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("plugin.package.upload.failed")) {
		t.Fatalf("expected upload failed audit log, got %s", w.Body.String())
	}
}

func TestAdminPluginPackageUploadHTTP_OKDetailAndPromote(t *testing.T) {
	cleanupHTTPUploadStorage(t)
	root := httpTestProjectRoot(t)
	_ = os.RemoveAll(filepath.Join(root, "storage", "plugins", "packages", "http_upload_ok"))
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Join(root, "storage", "plugins", "packages", "http_upload_ok")) })

	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)
	body, contentType := multipartUploadBody(t, "http_upload_ok.zip", makeHTTPUploadZip(t, "http_upload_ok"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/packages/upload", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected upload ok, got %d: %s", w.Code, w.Body.String())
	}
	var upload struct {
		UploadID   string `json:"upload_id"`
		Status     string `json:"status"`
		CanPromote bool   `json:"can_promote"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &upload); err != nil {
		t.Fatalf("decode upload: %v", err)
	}
	if upload.UploadID == "" || !upload.CanPromote || (upload.Status != "ok" && upload.Status != "warning") {
		t.Fatalf("unexpected upload response: %s", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/packages/uploads/"+upload.UploadID, nil)
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || bytes.Contains(w.Body.Bytes(), []byte(root)) {
		t.Fatalf("expected relative-path detail, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/packages/uploads/"+upload.UploadID+"/promote", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected promote ok, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("storage/plugins/packages/http_upload_ok")) {
		t.Fatalf("expected promoted package path, got %s", w.Body.String())
	}
}

func TestAdminPluginPackageUploadHTTP_LifecycleAPIs(t *testing.T) {
	cleanupHTTPUploadStorage(t)
	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)
	body, contentType := multipartUploadBody(t, "http_upload_lifecycle.zip", makeHTTPUploadZip(t, "http_upload_lifecycle"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/packages/upload", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected upload ok, got %d: %s", w.Code, w.Body.String())
	}
	var upload struct {
		UploadID string `json:"upload_id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &upload)
	if upload.UploadID == "" {
		t.Fatalf("missing upload id: %s", w.Body.String())
	}

	for _, call := range []struct {
		method string
		path   string
		body   string
		want   string
	}{
		{http.MethodGet, "/api/v1/admin/plugins/packages/uploads", "", "http_upload_lifecycle"},
		{http.MethodPost, "/api/v1/admin/plugins/packages/uploads/" + upload.UploadID + "/rescan", `{}`, `"staged"`},
		{http.MethodPost, "/api/v1/admin/plugins/packages/uploads/" + upload.UploadID + "/approval", `{"reason":"import"}`, `"approval_pending"`},
		{http.MethodPost, "/api/v1/admin/plugins/packages/uploads/" + upload.UploadID + "/approve", `{"comment":"ok"}`, `"approved"`},
	} {
		req = httptest.NewRequest(call.method, call.path, strings.NewReader(call.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+admin)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s %s expected ok, got %d: %s", call.method, call.path, w.Code, w.Body.String())
		}
		if !bytes.Contains(w.Body.Bytes(), []byte(call.want)) {
			t.Fatalf("%s %s expected %q in %s", call.method, call.path, call.want, w.Body.String())
		}
	}
}

func TestAdminPluginPackageRepositoryCleanupRoutesExist(t *testing.T) {
	cleanupHTTPUploadStorage(t)
	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)
	for _, path := range []string{
		"/api/v1/admin/plugins/packages/cleanup/preview",
		"/api/v1/admin/plugins/packages/cleanup",
		"/api/v1/admin/plugins/packages/repository/cleanup",
	} {
		body := `{"dry_run":true,"statuses":["blocked","invalid"]}`
		if path != "/api/v1/admin/plugins/packages/cleanup/preview" {
			previewReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/packages/cleanup/preview", strings.NewReader(body))
			previewReq.Header.Set("Content-Type", "application/json")
			previewReq.Header.Set("Authorization", "Bearer "+admin)
			previewW := httptest.NewRecorder()
			router.ServeHTTP(previewW, previewReq)
			if previewW.Code != http.StatusOK {
				t.Fatalf("preview before %s expected ok, got %d: %s", path, previewW.Code, previewW.Body.String())
			}
			body = `{"dry_run":false,"statuses":["blocked","invalid"],"confirm_token":"` + jsonField(t, previewW.Body.Bytes(), "confirm_token") + `"}`
		}
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+admin)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound {
			t.Fatalf("%s route must exist, got 404", path)
		}
		if w.Code != http.StatusOK {
			t.Fatalf("%s expected ok, got %d: %s", path, w.Code, w.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/plugins/packages/repository", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code == http.StatusNotFound {
		t.Fatalf("repository delete route must exist, got 404")
	}
}

func jsonField(t *testing.T, raw []byte, field string) string {
	t.Helper()
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("decode json: %v body=%s", err, string(raw))
	}
	value, _ := obj[field].(string)
	return strings.TrimSpace(value)
}

func multipartUploadBody(t *testing.T, filename string, raw []byte) ([]byte, string) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := fw.Write(raw); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("multipart close: %v", err)
	}
	return body.Bytes(), mw.FormDataContentType()
}

func makeHTTPUploadZip(t *testing.T, code string) []byte {
	t.Helper()
	manifest := strings.ReplaceAll(`{"code":"__CODE__","name":"HTTP Upload","version":"1.0.0","compatible_core_version":">=1.4.0","is_system":false,"content_types":["http_upload_item"],"content_type_definitions":[{"type":"http_upload_item","name":"HTTP Upload Item","plugin_code":"__CODE__","create_permission":"__CODE__.item.create","edit_permission":"__CODE__.item.edit","delete_permission":"__CODE__.item.delete","audit_permission":"__CODE__.item.audit","default_status":"draft","allow_comment":true,"allow_like":true,"allow_favorite":true,"seo_type":"Article"}],"permissions":[{"code":"__CODE__.item.create","name":"create","scope":"community"},{"code":"__CODE__.item.edit","name":"edit","scope":"own"},{"code":"__CODE__.item.delete","name":"delete","scope":"own"},{"code":"__CODE__.item.audit","name":"audit","scope":"community"}],"menus":[{"code":"__CODE__.admin","title":"HTTP","path":"/admin-next/__CODE__","location":"admin","area":"admin","permission":"__CODE__.item.audit"}],"routes":[{"area":"admin","method":"GET","path":"/api/v1/admin/__CODE__","handler":"reserved","auth":"admin","permission":"__CODE__.item.audit"}]}`, "__CODE__", code)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, raw := range map[string][]byte{
		"manifest.json":            []byte(manifest),
		"README.md":                []byte("# HTTP Upload\n"),
		"config.example.json":      []byte("{}"),
		"docs/registry-example.md": []byte("reserved\n"),
	} {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create: %v", err)
		}
		if _, err := w.Write(raw); err != nil {
			t.Fatalf("zip write: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func cleanupHTTPUploadStorage(t *testing.T) {
	t.Helper()
	root := httpTestProjectRoot(t)

	// Use allowlisted "storage/plugins/packages" and ensure it's writable in dev env.
	uploads := "storage/plugins/packages/http-test-uploads"
	staging := "storage/plugins/packages/http-test-staging"
	quarantine := "storage/plugins/packages/http-test-quarantine"
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_UPLOADS_ROOT", uploads)
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_STAGING_ROOT", staging)
	t.Setenv("DEVHUB_PLUGIN_PACKAGE_QUARANTINE_ROOT", quarantine)

	for _, rel := range []string{uploads, staging, quarantine} {
		_ = os.MkdirAll(filepath.Join(root, filepath.FromSlash(rel)), 0o755)
		_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(rel)))
	}
	t.Cleanup(func() {
		for _, rel := range []string{uploads, staging, quarantine} {
			_ = os.RemoveAll(filepath.Join(root, filepath.FromSlash(rel)))
		}
	})
}

func httpTestProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	t.Fatalf("project root not found")
	return ""
}
