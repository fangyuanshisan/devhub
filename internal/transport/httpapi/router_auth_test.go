package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestMain(m *testing.M) {
	_ = os.Chdir("../../..")
	os.Exit(m.Run())
}

func TestAuthRequiredForWriteAPI(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewBufferString(`{"site":"go","board":"qa","title":"t","content":"c"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminLoginReturnsTokenPair(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBufferString(`{"account":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.AccessToken == "" || body.RefreshToken == "" {
		t.Fatalf("expected token pair, got %#v", body)
	}
}

func TestAdminEndpointRequiresToken(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestScopedAdminEntry(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	req := httptest.NewRequest(http.MethodGet, "/admin/php", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301, got %d: %s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Location"); got != "/admin-next?site=php" {
		t.Fatalf("expected redirect to /admin-next?site=php, got %q", got)
	}
}

func TestAdminOverviewUsesScopedSiteQuery(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := adminToken(t, router)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview?site=php", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		SiteStats []struct {
			Site string `json:"site"`
		} `json:"site_stats"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.SiteStats) != 1 || body.SiteStats[0].Site != "php" {
		t.Fatalf("expected php-only stats, got %#v", body.SiteStats)
	}
}

func TestGenericCommunityAPIsReturnSeedData(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	cases := []struct {
		path     string
		minTotal int
	}{
		{"/api/v1/communities", 5},
		{"/api/v1/topics", 20},
		{"/api/v1/topics?community_slug=php", 4},
		{"/api/v1/topics?community_slug=ai", 6},
		{"/api/v1/search/topics?keyword=go", 4},
		{"/api/v1/communities/php/tags", 3},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d: %s", tc.path, w.Code, w.Body.String())
		}
		if totalItems(w.Body.Bytes()) < tc.minTotal {
			t.Fatalf("%s expected at least %d items, got body: %s", tc.path, tc.minTotal, w.Body.String())
		}
	}
}

func TestCreateTopicFlowInMemoryMode(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	payload := `{
		"community_slug":"php",
		"category_id":101,
		"content_type":"article",
		"title":"刚发布的 PHP Topic",
		"summary":"发布流程测试",
		"content":"这是一条用于验证 DevHub 发布闭环的正文内容。",
		"tags":["Laravel"]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/topics", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var created struct {
		ID        int64  `json:"id"`
		DetailURL string `json:"detail_url"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.DetailURL == "" {
		t.Fatalf("expected created topic id and detail url, got %#v", created)
	}

	for _, path := range []string{
		"/api/v1/topics?community_slug=php",
		"/api/v1/search/topics?keyword=%E5%88%9A%E5%8F%91%E5%B8%83",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d: %s", path, w.Code, w.Body.String())
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("刚发布的 PHP Topic")) {
			t.Fatalf("%s expected created topic in response, got %s", path, w.Body.String())
		}
	}
}

func totalItems(body []byte) int {
	var paged struct {
		Items []json.RawMessage `json:"items"`
		Total int               `json:"total"`
	}
	if json.Unmarshal(body, &paged) == nil && paged.Items != nil {
		if paged.Total > 0 {
			return paged.Total
		}
		return len(paged.Items)
	}
	var direct []json.RawMessage
	if json.Unmarshal(body, &direct) == nil {
		return len(direct)
	}
	return 0
}

func adminToken(t *testing.T, router http.Handler) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBufferString(`{"account":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	var body struct {
		AccessToken string `json:"access_token"`
		Token       string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.AccessToken != "" {
		return body.AccessToken
	}
	return body.Token
}
