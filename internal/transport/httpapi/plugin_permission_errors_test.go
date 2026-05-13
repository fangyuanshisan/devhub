package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
	"github.com/gin-gonic/gin"
)

func TestRequirePermissionReturnsStructuredPluginError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := &Server{svc: service.New(store.NewMemoryStore())}
	r := gin.New()
	r.GET("/t",
		func(c *gin.Context) {
			c.Set("auth_user", domain.AuthUser{ID: 1, Username: "operator", TokenType: "admin", Permissions: []string{}})
			c.Next()
		},
		srv.requirePermission("plugin.read"),
		func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) },
	)
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body == "" || !containsAll(body, []string{`"code":"plugin_permission_denied"`, `"permission_code":"plugin.read"`}) {
		t.Fatalf("expected structured code and permission_code, got: %s", body)
	}
}

func TestRequirePermissionKeepsLegacyForNonPluginEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv := &Server{svc: service.New(store.NewMemoryStore())}
	r := gin.New()
	r.GET("/t",
		func(c *gin.Context) {
			c.Set("auth_user", domain.AuthUser{ID: 1, Username: "operator", TokenType: "admin", Permissions: []string{}})
			c.Next()
		},
		srv.requirePermission("site.read"),
		func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) },
	)
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
	if body := strings.TrimSpace(w.Body.String()); body == "" || body != `{"error":"无权限"}` {
		t.Fatalf("expected legacy error only, got: %s", body)
	}
}

func containsAll(body string, fragments []string) bool {
	for _, f := range fragments {
		if !strings.Contains(body, f) {
			return false
		}
	}
	return true
}
