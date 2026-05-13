package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestCommunityNavigationReturnsFrontendMenusAndReasons(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/communities/php/navigation", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected navigation ok, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Items []struct {
			PluginCode string `json:"plugin_code"`
			Location   string `json:"location"`
			Visible    bool   `json:"visible"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	foundQA := false
	for _, item := range body.Items {
		if item.PluginCode == "qa" && item.Location == "community_nav" {
			foundQA = true
			if !item.Visible {
				t.Fatalf("expected qa nav visible when enabled, got %s", w.Body.String())
			}
		}
	}
	if !foundQA {
		t.Fatalf("expected qa frontend menu in community navigation, got %s", w.Body.String())
	}

	disable := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/qa/disable", nil)
	disable.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, disable)
	if w.Code != http.StatusOK {
		t.Fatalf("expected disable ok, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/communities/php/navigation", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected navigation ok, got %d: %s", w.Code, w.Body.String())
	}
	foundQA = false
	var body2 struct {
		Items []struct {
			PluginCode  string `json:"plugin_code"`
			Location    string `json:"location"`
			Visible     bool   `json:"visible"`
			ReasonCode  string `json:"reason_code"`
			Reason      string `json:"reason"`
			ContentType string `json:"content_type"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body2); err != nil {
		t.Fatal(err)
	}
	for _, item := range body2.Items {
		if item.PluginCode == "qa" && item.Location == "community_nav" {
			foundQA = true
			if item.Visible {
				t.Fatalf("expected qa nav hidden when disabled, got %s", w.Body.String())
			}
			if item.ReasonCode != "plugin_disabled" || !bytes.Contains([]byte(item.Reason), []byte("未启用")) {
				t.Fatalf("expected plugin_disabled reason, got %s", w.Body.String())
			}
		}
	}
	if !foundQA {
		t.Fatalf("expected qa frontend menu entry to remain in list with visible=false, got %s", w.Body.String())
	}
}

func TestCommunityCreateOptionsRequireLogin(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/communities/php/create-options", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected create options ok, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Items []struct {
			ContentType string `json:"content_type"`
			Visible     bool   `json:"visible"`
			ReasonCode  string `json:"reason_code"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	foundArticle := false
	for _, item := range body.Items {
		if item.ContentType == "article" {
			foundArticle = true
			if item.Visible {
				t.Fatalf("expected article create option hidden for guest, got %s", w.Body.String())
			}
			if item.ReasonCode != "plugin_login_required" {
				t.Fatalf("expected plugin_login_required, got %s", w.Body.String())
			}
		}
	}
	if !foundArticle {
		t.Fatalf("expected article option, got %s", w.Body.String())
	}

	token := userToken(t, router, "admin")
	req = httptest.NewRequest(http.MethodGet, "/api/v1/communities/php/create-options", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected create options ok, got %d: %s", w.Code, w.Body.String())
	}
	body = struct {
		Items []struct {
			ContentType string `json:"content_type"`
			Visible     bool   `json:"visible"`
			ReasonCode  string `json:"reason_code"`
		} `json:"items"`
	}{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	foundArticle = false
	for _, item := range body.Items {
		if item.ContentType == "article" {
			foundArticle = true
			if !item.Visible {
				t.Fatalf("expected article create option visible for logged-in user, got %s", w.Body.String())
			}
		}
	}
	if !foundArticle {
		t.Fatalf("expected article option, got %s", w.Body.String())
	}
}

func TestAdminPluginMenusPreviewRejectsUserToken(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := userToken(t, router, "admin")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/qa/menus/preview?community_slug=php", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Fatalf("expected 401/403 for non-admin token, got %d: %s", w.Code, w.Body.String())
	}
}
