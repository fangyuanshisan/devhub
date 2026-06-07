package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestAdminPluginHookExecutionsQueryRequiresAuthAndPermission(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/qa/hooks/executions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}

	// user token should be forbidden.
	user := userToken(t, router, "admin")
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/qa/hooks/executions", nil)
	req.Header.Set("Authorization", "Bearer "+user)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Backend admin auth middleware rejects non-admin tokens as 401 (invalid admin token),
	// so we assert "not ok" without faking a 403 expectation.
	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminPluginHookExecutionsQueryReturnsItems(t *testing.T) {
	repo := store.NewMemoryStore()
	_, _ = repo.AppendHookExecution(domain.HookExecution{
		HookName:     "BeforeCreateContent",
		PluginCode:   "qa",
		Mode:         "blocking",
		Blocking:     true,
		Success:      false,
		ErrorMessage: "e2e hook fail",
		StartedAt:    "2026-05-12 10:00:00",
		FinishedAt:   "2026-05-12 10:00:00",
	})
	router := NewRouter(service.New(repo))
	token := adminToken(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/qa/hooks/executions?success=false&page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"items"`)) || !bytes.Contains(w.Body.Bytes(), []byte(`"total"`)) {
		t.Fatalf("expected list response, got %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"hook_name":"BeforeCreateContent"`)) {
		t.Fatalf("expected hook item, got %s", w.Body.String())
	}
}

func TestAdminPluginHookExecutionsGlobalQueryReturnsExternalServiceFailures(t *testing.T) {
	repo := store.NewMemoryStore()
	_, _ = repo.AppendHookExecution(domain.HookExecution{
		HookName:     "AfterCreateContent",
		PluginCode:   "qa",
		ServiceType:  "external_service",
		Mode:         "non_blocking",
		Success:      false,
		Status:       "failed",
		ErrorMessage: "external service failed",
		StartedAt:    "2026-05-12 10:00:00",
		FinishedAt:   "2026-05-12 10:00:01",
	})
	router := NewRouter(service.New(repo))
	token := adminToken(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/hooks/executions?service_type=external_service&success=false&page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"plugin_code":"qa"`)) || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"failed"`)) {
		t.Fatalf("expected failed external_service hook item, got %s", w.Body.String())
	}
}
