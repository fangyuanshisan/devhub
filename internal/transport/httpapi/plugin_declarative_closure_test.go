package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestAdminPostCreateSupportsDeclarativePluginContentType(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := adminToken(t, router)
	code := "official_links_http"
	contentType := "friend_link_http"

	do := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	manifest := `{
		"code":"` + code + `",
		"name":"声明型友情链接插件",
		"version":"1.0.0",
		"compatible_core_version":">=1.4.0",
		"is_system":false,
		"content_types":[{"type":"` + contentType + `","name":"友情链接","plugin_code":"` + code + `","create_permission":"` + code + `.link.create"}],
		"permissions":[
			{"plugin_code":"` + code + `","code":"` + code + `.link.create","name":"创建友情链接","scope":"community"},
			{"plugin_code":"` + code + `","code":"` + code + `.menu.view","name":"查看友情链接菜单","scope":"community"}
		],
		"menus":[{"plugin_code":"` + code + `","code":"` + code + `.admin.links","title":"友情链接管理","path":"/admin-next/plugins/overview?tab=list","area":"admin","permission":"` + code + `.menu.view"}],
		"routes":[],
		"hooks":[],
		"config_schema":{"type":"object","properties":{"enabled":{"type":"boolean","default":true}}},
		"migrations":[]
	}`
	if w := do(http.MethodPost, "/api/v1/admin/plugins/install", manifest); w.Code != http.StatusCreated {
		t.Fatalf("install declarative plugin failed: %d %s", w.Code, w.Body.String())
	}
	if w := do(http.MethodPost, "/api/v1/admin/plugins/"+code+"/enable", ""); w.Code != http.StatusOK {
		t.Fatalf("enable declarative plugin failed: %d %s", w.Code, w.Body.String())
	}
	if w := do(http.MethodPost, "/api/v1/admin/communities/1/plugins/"+code+"/enable", ""); w.Code != http.StatusOK {
		t.Fatalf("enable community plugin failed: %d %s", w.Code, w.Body.String())
	}

	categoryBody := `{"name":"友情链接","slug":"links-http","type":"` + contentType + `","content_type":"` + contentType + `","plugin_code":"` + code + `","allowed_content_types":["` + contentType + `"],"visible":true,"nav_visible":true}`
	categoryResp := do(http.MethodPost, "/api/v1/admin/communities/1/categories", categoryBody)
	if categoryResp.Code != http.StatusCreated {
		t.Fatalf("create category failed: %d %s", categoryResp.Code, categoryResp.Body.String())
	}
	var category struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(categoryResp.Body.Bytes(), &category); err != nil || category.ID == 0 {
		t.Fatalf("invalid category response: %v %s", err, categoryResp.Body.String())
	}

	createBody := `{"site":"php","board":"community","category_id":` + intString(category.ID) + `,"content_type":"` + contentType + `","plugin_code":"` + code + `","title":"声明型友情链接","summary":"fixture summary","content":"https://example.com declarative link body"}`
	created := do(http.MethodPost, "/api/v1/admin/posts", createBody)
	if created.Code != http.StatusCreated {
		t.Fatalf("expected declarative content create success, got %d: %s", created.Code, created.Body.String())
	}
	if !bytes.Contains(created.Body.Bytes(), []byte(`"plugin_code":"`+code+`"`)) || !bytes.Contains(created.Body.Bytes(), []byte(`"content_type":"`+contentType+`"`)) {
		t.Fatalf("expected plugin ownership in created post, got %s", created.Body.String())
	}

	search := do(http.MethodGet, "/api/v1/search/topics?content_type="+contentType, "")
	if search.Code != http.StatusOK {
		t.Fatalf("search declarative content failed: %d %s", search.Code, search.Body.String())
	}
	if !bytes.Contains(search.Body.Bytes(), []byte(`"content_type":"`+contentType+`"`)) {
		t.Fatalf("expected declarative content type in search result, got %s", search.Body.String())
	}
	if bytes.Contains(search.Body.Bytes(), []byte(`"content_type":"article"`)) {
		t.Fatalf("search content_type filter was ignored, got %s", search.Body.String())
	}

	blockedBody := strings.Replace(createBody, `"category_id":`+intString(category.ID), `"category_id":101`, 1)
	blocked := do(http.MethodPost, "/api/v1/admin/posts", blockedBody)
	if blocked.Code != http.StatusBadRequest || !strings.Contains(blocked.Body.String(), "当前板块") {
		t.Fatalf("expected disallowed category failure, got %d: %s", blocked.Code, blocked.Body.String())
	}

	if w := do(http.MethodPost, "/api/v1/admin/communities/1/plugins/"+code+"/disable", ""); w.Code != http.StatusOK {
		t.Fatalf("disable community plugin failed: %d %s", w.Code, w.Body.String())
	}
	communityBlocked := do(http.MethodPost, "/api/v1/admin/posts", createBody)
	if communityBlocked.Code != http.StatusBadRequest || !strings.Contains(communityBlocked.Body.String(), "当前子站未启用") {
		t.Fatalf("expected community disabled failure, got %d: %s", communityBlocked.Code, communityBlocked.Body.String())
	}

	if w := do(http.MethodPost, "/api/v1/admin/communities/1/plugins/"+code+"/enable", ""); w.Code != http.StatusOK {
		t.Fatalf("re-enable community plugin failed: %d %s", w.Code, w.Body.String())
	}
	if w := do(http.MethodPost, "/api/v1/admin/plugins/"+code+"/disable", ""); w.Code != http.StatusOK {
		t.Fatalf("disable global plugin failed: %d %s", w.Code, w.Body.String())
	}
	globalBlocked := do(http.MethodPost, "/api/v1/admin/posts", createBody)
	if globalBlocked.Code != http.StatusBadRequest || !strings.Contains(globalBlocked.Body.String(), "插件未启用") {
		t.Fatalf("expected global disabled failure, got %d: %s", globalBlocked.Code, globalBlocked.Body.String())
	}
}

func intString(v int64) string {
	return strconv.FormatInt(v, 10)
}
