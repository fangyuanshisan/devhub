package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestOfficialAnnouncementContextUsesSafeRedactedConfig(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := service.New(repo)
	if _, err := svc.SetPluginConfig("official_announcement", `{"enabled":true,"message":"  系统公告  ","link_text":"详情","link_url":"/news","dismissible":true}`); err != nil {
		t.Fatalf("SetPluginConfig: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugins/official-announcement/context?mount_id=m1&area=frontend", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Visible bool           `json:"visible"`
		Config  map[string]any `json:"config"`
		Context map[string]any `json:"context"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Visible {
		t.Fatalf("expected visible context, got %s", w.Body.String())
	}
	if got := body.Config["message"]; got != "系统公告" {
		t.Fatalf("expected trimmed message, got %#v", got)
	}
	if got := body.Config["link_url"]; got != "/news" {
		t.Fatalf("expected internal link, got %#v", got)
	}
	if body.Context["plugin_code"] != "official_announcement" {
		t.Fatalf("expected plugin context, got %#v", body.Context)
	}
}

func TestOfficialAnnouncementContextRejectsUnsafeLinkAtConfigBoundary(t *testing.T) {
	svc := service.New(store.NewMemoryStore())
	if _, err := svc.SetPluginConfig("official_announcement", `{"enabled":true,"message":"公告","link_url":"https://evil.example/"}`); err == nil {
		t.Fatal("expected unsafe external link to fail schema validation")
	}
}

func TestOfficialAnnouncementContextGatingStates(t *testing.T) {
	svc := service.New(store.NewMemoryStore())
	if _, err := svc.SetPluginConfig("official_announcement", `{"enabled":true,"message":"公告","link_text":"详情","link_url":"/","dismissible":true}`); err != nil {
		t.Fatalf("SetPluginConfig: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugins/official-announcement/context?mount_id=m1&area=frontend&community_slug=php", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"visible":true`) {
		t.Fatalf("expected visible announcement, got %d: %s", w.Code, w.Body.String())
	}

	if _, err := svc.SetCommunityPluginStatus(1, "official_announcement", pluginregistry.StatusDisabled); err != nil {
		t.Fatalf("disable community plugin: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugins/official-announcement/context?mount_id=m1&area=frontend&community_slug=php", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"visible":false`) || !strings.Contains(w.Body.String(), "community_plugin_disabled") {
		t.Fatalf("expected community disabled context, got %d: %s", w.Code, w.Body.String())
	}

	if _, err := svc.SetPluginStatus("official_announcement", pluginregistry.StatusDisabled); err != nil {
		t.Fatalf("disable plugin: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugins/official-announcement/context?mount_id=m1&area=frontend", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"visible":false`) || !strings.Contains(w.Body.String(), "plugin_disabled") {
		t.Fatalf("expected plugin disabled context, got %d: %s", w.Code, w.Body.String())
	}

	if _, err := svc.SetPluginStatus("official_announcement", pluginregistry.StatusArchived); err != nil {
		t.Fatalf("archive plugin: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugins/official-announcement/context?mount_id=m1&area=frontend", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"visible":false`) || !strings.Contains(w.Body.String(), "plugin_soft_uninstalled") {
		t.Fatalf("expected archived context, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOfficialAnnouncementAdminContextRequiresAdminToken(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugins/official-announcement/context?mount_id=m1&area=admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without admin token, got %d: %s", w.Code, w.Body.String())
	}

	user := userToken(t, router, "admin")
	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugins/official-announcement/context?mount_id=m1&area=admin", nil)
	req.Header.Set("Authorization", "Bearer "+user)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for frontend user token, got %d: %s", w.Code, w.Body.String())
	}

	admin := adminToken(t, router)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/plugins/official-announcement/context?mount_id=m1&area=admin", nil)
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin token, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOfficialAnnouncementAuditEventsSanitizesMetadata(t *testing.T) {
	repo := store.NewMemoryStore()
	router := NewRouter(service.New(repo))
	body := `{"mount_id":"m1","area":"frontend","action":"official_announcement.clicked","request_id":"r1","metadata":{"token":"secret","link_url":"/news"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/plugins/official-announcement/audit-events", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	logs := repo.AdminLogs("admin")
	var got *domain.AdminLog
	for i := range logs {
		if logs[i].Action == "official_announcement.clicked" {
			got = &logs[i]
			break
		}
	}
	if got == nil {
		t.Fatalf("expected audit log, got %#v", logs)
	}
	if strings.Contains(got.Metadata, "secret") || strings.Contains(got.Metadata, "token") {
		t.Fatalf("expected credential metadata to be stripped, got %s", got.Metadata)
	}
	if !strings.Contains(got.Metadata, "/news") {
		t.Fatalf("expected safe metadata to remain, got %s", got.Metadata)
	}
}
