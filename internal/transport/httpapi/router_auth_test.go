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

func TestLegacyPostsWriteReturnsGoneWhenAuthenticated(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := userToken(t, router, "admin")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewBufferString(`{"site":"php","board":"community","title":"legacy","content":"legacy content body"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusGone {
		t.Fatalf("expected 410 for deprecated posts write, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminPostCreateRespectsGlobalPluginStatus(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := adminToken(t, router)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/qa/disable", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected qa disable success, got %d: %s", w.Code, w.Body.String())
	}

	payload := `{"site":"php","board":"qa","title":"disabled qa check","summary":"check","content":"this content is long enough for validation"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/posts", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when qa is disabled, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("插件未启用")) {
		t.Fatalf("expected plugin disabled error, got %s", w.Body.String())
	}
}

func TestAdminPostUpdateRejectsOwnershipChange(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := adminToken(t, router)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/posts/1", bytes.NewBufferString(`{"board":"qa","title":"try move qa"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for admin ownership change, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("不允许修改内容板块或内容类型")) {
		t.Fatalf("expected ownership change error, got %s", w.Body.String())
	}
}

func TestPublicPluginAPIsHideConfig(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/plugins/qa/config", bytes.NewBufferString(`{"config_json":{"secret":"hidden"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected plugin config update success, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/communities/1/plugins/qa/config", bytes.NewBufferString(`{"config_json":{"community_secret":"hidden"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected community plugin config update success, got %d: %s", w.Code, w.Body.String())
	}

	for _, path := range []string{"/api/v1/plugins", "/api/v1/communities/php/plugins"} {
		req = httptest.NewRequest(http.MethodGet, path, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s expected 200, got %d: %s", path, w.Code, w.Body.String())
		}
		if bytes.Contains(w.Body.Bytes(), []byte("hidden")) || bytes.Contains(w.Body.Bytes(), []byte("config_json")) || bytes.Contains(w.Body.Bytes(), []byte("resolved_config")) {
			t.Fatalf("%s should not expose runtime config, got %s", path, w.Body.String())
		}
	}
}

func TestPluginConfigAuditAndInvalidJSON(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := adminToken(t, router)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/plugins/qa/config", bytes.NewBufferString(`{"config_json":{"limit":5}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected global plugin config success, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/plugins/qa/config", bytes.NewBufferString(`{"config_json":`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid json to fail, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/communities/1/plugins/qa/config", bytes.NewBufferString(`{"config_json":{"enabled":true}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected community plugin config success, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/communities/1/plugins/sort", bytes.NewBufferString(`{"codes":["docs","qa","wiki"]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected community plugin sort success, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected audit logs, got %d: %s", w.Code, w.Body.String())
	}
	for _, want := range [][]byte{[]byte("更新插件全局配置"), []byte("更新子站插件配置"), []byte("子站插件排序")} {
		if !bytes.Contains(w.Body.Bytes(), want) {
			t.Fatalf("expected audit log %q in %s", want, w.Body.String())
		}
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("metadata_json")) || !bytes.Contains(w.Body.Bytes(), []byte("old_value")) || !bytes.Contains(w.Body.Bytes(), []byte("new_value")) {
		t.Fatalf("expected structured audit fields in %s", w.Body.String())
	}
	for _, want := range [][]byte{[]byte(`\"plugin_code\":\"qa\"`), []byte(`\"operation\":\"plugin_config\"`), []byte(`\"operation\":\"community_plugin_config\"`), []byte(`\"operation\":\"community_plugin_sort\"`)} {
		if !bytes.Contains(w.Body.Bytes(), want) {
			t.Fatalf("expected structured audit metadata %q in %s", want, w.Body.String())
		}
	}
}

func TestModeratorPluginMenusRespectCommunityScopeAndPluginStatus(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)
	moderator := userToken(t, router, "operator")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/communities/1/plugins/qa/disable", nil)
	req.Header.Set("Authorization", "Bearer "+admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected qa community disable success, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/moderator/plugin-menus?community_slug=php", nil)
	req.Header.Set("Authorization", "Bearer "+moderator)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected moderator plugin menus, got %d: %s", w.Code, w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte("qa-moderator")) {
		t.Fatalf("qa moderator menu should be hidden after community disable: %s", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/moderator/plugin-menus?community_slug=go", nil)
	req.Header.Set("Authorization", "Bearer "+moderator)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected cross-community moderator menu request to fail, got %d: %s", w.Code, w.Body.String())
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

func TestUserAndAdminTokensAreSeparated(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	userToken := userToken(t, router, "admin")
	adminToken := adminToken(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden && w.Code != http.StatusUnauthorized {
		t.Fatalf("expected frontend token to be rejected by privileged admin API, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected admin token to be rejected by frontend auth API, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRegisteredMemoryUserCanLoginWithOwnPassword(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"username":"newuser","nickname":"新用户","email":"newuser@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected register success, got %d: %s", w.Code, w.Body.String())
	}
	var registered struct {
		AccessToken string `json:"access_token"`
		User        struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &registered); err != nil {
		t.Fatal(err)
	}
	if registered.AccessToken == "" || registered.User.Username != "newuser" {
		t.Fatalf("unexpected register response: %s", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"account":"newuser","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected registered user to login with own password, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"account":"newuser","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected wrong password to fail, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCommunityModeratorScopeUsesFrontendUserToken(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	moderatorToken := userToken(t, router, "operator")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports?site=php", nil)
	req.Header.Set("Authorization", "Bearer "+moderatorToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected php moderator to read php reports, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports?site=go", nil)
	req.Header.Set("Authorization", "Bearer "+moderatorToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected scoped fallback instead of cross-community access failure, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Items []struct {
			CommunityID int64 `json:"community_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, item := range body.Items {
		if item.CommunityID != 0 && item.CommunityID != 1 {
			t.Fatalf("php moderator should not see non-php report, got community_id=%d in %s", item.CommunityID, w.Body.String())
		}
	}
}

func TestModeratorWorkbenchAPIScope(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	normalUserToken := userToken(t, router, "admin")
	phpModeratorToken := userToken(t, router, "operator")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/moderator/communities", nil)
	req.Header.Set("Authorization", "Bearer "+normalUserToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected normal user to be rejected by moderator API, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/moderator/communities", nil)
	req.Header.Set("Authorization", "Bearer "+phpModeratorToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected php moderator communities, got %d: %s", w.Code, w.Body.String())
	}
	var communities struct {
		Items []struct {
			ID   int64  `json:"id"`
			Slug string `json:"slug"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &communities); err != nil {
		t.Fatal(err)
	}
	if len(communities.Items) != 1 || communities.Items[0].Slug != "php" {
		t.Fatalf("expected only php community, got %s", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/moderator/topics?community_id=2", nil)
	req.Header.Set("Authorization", "Bearer "+phpModeratorToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected cross-community moderator read to fail, got %d: %s", w.Code, w.Body.String())
	}
}

func TestModeratorActionsWriteAuditLog(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	phpModeratorToken := userToken(t, router, "operator")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/moderator/topics/1/hide", nil)
	req.Header.Set("Authorization", "Bearer "+phpModeratorToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected php moderator to hide php topic, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/v1/moderator/topics/7/hide", nil)
	req.Header.Set("Authorization", "Bearer "+phpModeratorToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected php moderator to be denied on go topic, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/moderator/audit-logs?community_id=1&actor_type=moderator", nil)
	req.Header.Set("Authorization", "Bearer "+phpModeratorToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected moderator audit logs, got %d: %s", w.Code, w.Body.String())
	}
	var logs struct {
		Items []struct {
			ActorType   string `json:"actor_type"`
			CommunityID int64  `json:"community_id"`
			Action      string `json:"action"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &logs); err != nil {
		t.Fatal(err)
	}
	if len(logs.Items) == 0 {
		t.Fatalf("expected moderator audit log, got %s", w.Body.String())
	}
	if logs.Items[0].ActorType != "moderator" || logs.Items[0].CommunityID != 1 || logs.Items[0].Action != "hide_topic" {
		t.Fatalf("unexpected audit log: %#v body=%s", logs.Items[0], w.Body.String())
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
	token := userToken(t, router, "admin")
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
	req.Header.Set("Authorization", "Bearer "+token)
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

func userToken(t *testing.T, router http.Handler, account string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"account":"`+account+`","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("frontend login failed: %d %s", w.Code, w.Body.String())
	}
	var body struct {
		AccessToken string `json:"access_token"`
		Token       string `json:"token"`
		TokenType   string `json:"token_type"`
		Audience    string `json:"aud"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.TokenType != "user" || body.Audience != "devhub_frontend" {
		t.Fatalf("expected frontend user token, got type=%q aud=%q body=%s", body.TokenType, body.Audience, w.Body.String())
	}
	if body.AccessToken != "" {
		return body.AccessToken
	}
	return body.Token
}
