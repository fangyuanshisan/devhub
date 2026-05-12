package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"devhub-gin-backend/internal/domain"
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

	// qa 插件存在 config_schema 约束，这里使用合法配置值，测试重点是：public API 不应泄露 config_json/resolved_config。
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/plugins/qa/config", bytes.NewBufferString(`{"config_json":{"allow_anonymous_answer":true,"default_question_status":"publish"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+admin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected plugin config update success, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/communities/1/plugins/qa/config", bytes.NewBufferString(`{"config_json":{"allow_anonymous_answer":false,"default_question_status":"publish"}}`))
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

	// qa 插件存在 config_schema 约束，这里使用合法配置值，测试重点是：审计日志应记录结构化 old/new/metadata。
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/plugins/qa/config", bytes.NewBufferString(`{"config_json":{"allow_anonymous_answer":true,"default_question_status":"publish"}}`))
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

	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/communities/1/plugins/qa/config", bytes.NewBufferString(`{"config_json":{"allow_anonymous_answer":false,"default_question_status":"review"}}`))
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
	for _, want := range [][]byte{[]byte(`\"plugin_code\":\"qa\"`), []byte(`\"operation\":\"plugin_config\"`), []byte(`\"operation\":\"community_plugin_config\"`), []byte(`\"operation\":\"community_plugin_sort\"`), []byte(`\"changed_keys\"`), []byte(`default_question_status`)} {
		if !bytes.Contains(w.Body.Bytes(), want) {
			t.Fatalf("expected structured audit metadata %q in %s", want, w.Body.String())
		}
	}
}

func TestPluginMigrationFailureBlocksEnableAndRetryRestores(t *testing.T) {
	t.Setenv("CMS_STORE", "memory")
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := adminToken(t, router)

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

	if w := do(http.MethodPost, "/api/v1/admin/plugins/qa/disable", ""); w.Code != http.StatusOK {
		t.Fatalf("expected qa global disable success, got %d: %s", w.Code, w.Body.String())
	}
	if w := do(http.MethodPost, "/api/v1/admin/communities/1/plugins/qa/disable", ""); w.Code != http.StatusOK {
		t.Fatalf("expected qa community disable success, got %d: %s", w.Code, w.Body.String())
	}
	w := do(http.MethodPost, "/api/v1/admin/plugins/qa/migrations/qa_questions/e2e-fail", `{"error_message":"E2E forced migration failure"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected failed migration injection success, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"status":"failed"`)) || !bytes.Contains(w.Body.Bytes(), []byte("E2E forced migration failure")) {
		t.Fatalf("expected failed migration payload, got %s", w.Body.String())
	}

	w = do(http.MethodPost, "/api/v1/admin/plugins/qa/enable", "")
	if w.Code != http.StatusBadRequest || !bytes.Contains(w.Body.Bytes(), []byte("失败迁移")) {
		t.Fatalf("expected global enable blocked by failed migration, got %d: %s", w.Code, w.Body.String())
	}
	w = do(http.MethodPost, "/api/v1/admin/communities/1/plugins/qa/enable", "")
	if w.Code != http.StatusBadRequest || !bytes.Contains(w.Body.Bytes(), []byte("失败迁移")) {
		t.Fatalf("expected community enable blocked by failed migration, got %d: %s", w.Code, w.Body.String())
	}

	w = do(http.MethodGet, "/api/v1/admin/plugins/qa/migrations", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected migrations list, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"failed":1`)) || !bytes.Contains(w.Body.Bytes(), []byte("E2E forced migration failure")) {
		t.Fatalf("expected failed migration summary, got %s", w.Body.String())
	}

	w = do(http.MethodPost, "/api/v1/admin/plugins/qa/migrations/qa_questions/retry", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"success"`)) {
		t.Fatalf("expected retry success, got %d: %s", w.Code, w.Body.String())
	}
	w = do(http.MethodPost, "/api/v1/admin/plugins/qa/migrations/qa_questions/retry", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"success"`)) {
		t.Fatalf("expected repeated retry to remain success/no-op, got %d: %s", w.Code, w.Body.String())
	}
	w = do(http.MethodGet, "/api/v1/admin/plugins/qa/migrations", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected migrations list after retry, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"failed":0`)) || !bytes.Contains(w.Body.Bytes(), []byte(`"success"`)) {
		t.Fatalf("expected failed cleared after retry, got %s", w.Body.String())
	}

	if w = do(http.MethodPost, "/api/v1/admin/plugins/qa/enable", ""); w.Code != http.StatusOK {
		t.Fatalf("expected global enable after retry, got %d: %s", w.Code, w.Body.String())
	}
	if w = do(http.MethodPost, "/api/v1/admin/communities/1/plugins/qa/enable", ""); w.Code != http.StatusOK {
		t.Fatalf("expected community enable after retry, got %d: %s", w.Code, w.Body.String())
	}

	w = do(http.MethodGet, "/api/v1/admin/plugins/qa/audit-logs?action=plugin.migration&page_size=50", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected plugin audit logs, got %d: %s", w.Code, w.Body.String())
	}
	for _, want := range [][]byte{
		[]byte("plugin.migration.failed"),
		[]byte("plugin.migration.retry"),
		[]byte("plugin.migration.success"),
		[]byte("plugin_migration_test_injection"),
	} {
		if !bytes.Contains(w.Body.Bytes(), want) {
			t.Fatalf("expected audit marker %q in %s", want, w.Body.String())
		}
	}
}

func TestPluginArchiveRestoreAPIBlocksCreationAndAudits(t *testing.T) {
	t.Setenv("CMS_STORE", "memory")
	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)
	user := userToken(t, router, "admin")

	adminDo := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+admin)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	userDo := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+user)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	w := adminDo(http.MethodPost, "/api/v1/admin/plugins/qa/archive", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"archived"`)) {
		t.Fatalf("expected archive success, got %d: %s", w.Code, w.Body.String())
	}
	w = userDo(http.MethodPost, "/api/v1/topics", `{"community_id":1,"category_id":101,"content_type":"question","title":"Archived QA should fail","content":"body should be long enough for validation"}`)
	if w.Code != http.StatusBadRequest || !bytes.Contains(w.Body.Bytes(), []byte("archived")) {
		t.Fatalf("expected archived plugin to block question creation, got %d: %s", w.Code, w.Body.String())
	}
	w = adminDo(http.MethodPost, "/api/v1/admin/communities/1/plugins/qa/enable", "")
	if w.Code != http.StatusBadRequest || !bytes.Contains(w.Body.Bytes(), []byte("归档")) {
		t.Fatalf("expected archived plugin to block community enable, got %d: %s", w.Code, w.Body.String())
	}
	w = adminDo(http.MethodPost, "/api/v1/admin/plugins/qa/restore", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"disabled"`)) {
		t.Fatalf("expected restore to disabled, got %d: %s", w.Code, w.Body.String())
	}
	w = adminDo(http.MethodPost, "/api/v1/admin/plugins/qa/enable", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"enabled"`)) {
		t.Fatalf("expected enable after restore, got %d: %s", w.Code, w.Body.String())
	}
	w = adminDo(http.MethodGet, "/api/v1/admin/plugins/qa/audit-logs?action=plugin.&page_size=80", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected plugin audit logs, got %d: %s", w.Code, w.Body.String())
	}
	for _, want := range [][]byte{[]byte("plugin.archived"), []byte("plugin.restored")} {
		if !bytes.Contains(w.Body.Bytes(), want) {
			t.Fatalf("expected audit marker %q in %s", want, w.Body.String())
		}
	}
}

func TestHookFailureInjectionBlocksAndRecordsNonBlockingFailures(t *testing.T) {
	t.Setenv("CMS_STORE", "memory")
	router := NewRouter(service.New(store.NewMemoryStore()))
	admin := adminToken(t, router)
	user := userToken(t, router, "admin")

	adminDo := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+admin)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	userDo := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+user)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}
	createQuestion := func(title string) *httptest.ResponseRecorder {
		t.Helper()
		payload := map[string]any{
			"community_id":   1,
			"community_slug": "php",
			"category_id":    102,
			"content_type":   "question",
			"title":          title,
			"summary":        "E2E Hook API test summary",
			"content":        "这是一段用于 HookBus API 测试的正文内容，长度满足发布校验。",
			"tags":           []string{},
		}
		raw, _ := json.Marshal(payload)
		return userDo(http.MethodPost, "/api/v1/topics", string(raw))
	}

	if w := adminDo(http.MethodPost, "/api/v1/admin/plugins/qa/enable", ""); w.Code != http.StatusOK {
		t.Fatalf("expected qa global enable success, got %d: %s", w.Code, w.Body.String())
	}
	if w := adminDo(http.MethodPost, "/api/v1/admin/communities/1/plugins/qa/enable", ""); w.Code != http.StatusOK {
		t.Fatalf("expected qa community enable success, got %d: %s", w.Code, w.Body.String())
	}
	if w := adminDo(http.MethodPost, "/api/v1/admin/categories/102/enable", ""); w.Code != http.StatusOK {
		t.Fatalf("expected qa category enable success, got %d: %s", w.Code, w.Body.String())
	}

	blockingErr := "E2E blocking hook failure"
	w := adminDo(http.MethodPost, "/api/v1/admin/plugins/qa/hooks/BeforeCreateContent/e2e-fail", `{"mode":"blocking","error_message":"`+blockingErr+`"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected blocking hook injection success, got %d: %s", w.Code, w.Body.String())
	}
	blockedTitle := "E2E Hook Blocked Topic"
	w = createQuestion(blockedTitle)
	if w.Code != http.StatusBadRequest || !bytes.Contains(w.Body.Bytes(), []byte(blockingErr)) {
		t.Fatalf("expected blocking hook to reject create, got %d: %s", w.Code, w.Body.String())
	}
	w = adminDo(http.MethodGet, "/api/v1/admin/posts?q=E2E%20Hook%20Blocked%20Topic", "")
	if w.Code == http.StatusOK && bytes.Contains(w.Body.Bytes(), []byte(blockedTitle)) {
		t.Fatalf("blocking hook should not create dirty topic, got %s", w.Body.String())
	}
	w = adminDo(http.MethodGet, "/api/v1/admin/plugins/qa/hooks", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte("BeforeCreateContent")) || !bytes.Contains(w.Body.Bytes(), []byte(blockingErr)) || !bytes.Contains(w.Body.Bytes(), []byte(`"success":false`)) {
		t.Fatalf("expected blocked hook execution in stats, got %d: %s", w.Code, w.Body.String())
	}
	w = adminDo(http.MethodGet, "/api/v1/admin/plugins/qa/audit-logs?action=plugin.hook.blocked&page_size=50", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte("plugin.hook.blocked")) || !bytes.Contains(w.Body.Bytes(), []byte(blockingErr)) {
		t.Fatalf("expected plugin.hook.blocked audit, got %d: %s", w.Code, w.Body.String())
	}
	if w := adminDo(http.MethodPost, "/api/v1/admin/plugins/qa/hooks/BeforeCreateContent/e2e-fail", `{"clear":true}`); w.Code != http.StatusOK {
		t.Fatalf("expected blocking hook injection clear success, got %d: %s", w.Code, w.Body.String())
	}

	nonBlockingErr := "E2E non-blocking hook failure"
	w = adminDo(http.MethodPost, "/api/v1/admin/plugins/qa/hooks/AfterCreateContent/e2e-fail", `{"mode":"non_blocking","error_message":"`+nonBlockingErr+`"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected non-blocking hook injection success, got %d: %s", w.Code, w.Body.String())
	}
	w = createQuestion("E2E Hook NonBlocking Topic")
	if w.Code != http.StatusCreated {
		t.Fatalf("expected non-blocking hook not to block create, got %d: %s", w.Code, w.Body.String())
	}
	w = adminDo(http.MethodGet, "/api/v1/admin/plugins/qa/hooks", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte("AfterCreateContent")) || !bytes.Contains(w.Body.Bytes(), []byte(nonBlockingErr)) || !bytes.Contains(w.Body.Bytes(), []byte(`"success":false`)) {
		t.Fatalf("expected failed non-blocking hook execution in stats, got %d: %s", w.Code, w.Body.String())
	}
	w = adminDo(http.MethodGet, "/api/v1/admin/plugins/qa/audit-logs?action=plugin.hook.failed&page_size=50", "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte("plugin.hook.failed")) || !bytes.Contains(w.Body.Bytes(), []byte(nonBlockingErr)) {
		t.Fatalf("expected plugin.hook.failed audit, got %d: %s", w.Code, w.Body.String())
	}
	if w := adminDo(http.MethodPost, "/api/v1/admin/plugins/qa/hooks/AfterCreateContent/e2e-fail", `{"clear":true}`); w.Code != http.StatusOK {
		t.Fatalf("expected non-blocking hook injection clear success, got %d: %s", w.Code, w.Body.String())
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

func TestFrontendUserTokenCannotCallPluginGovernanceAPI(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := userToken(t, router, "admin")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/plugins/qa/disable", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Fatalf("expected frontend user token to be rejected by plugin governance API, got %d: %s", w.Code, w.Body.String())
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

func TestAdminPostsFiltersByPluginCodeAndContentType(t *testing.T) {
	router := NewRouter(service.New(store.NewMemoryStore()))
	token := adminToken(t, router)

	getPosts := func(path string) struct {
		Items []domain.Post `json:"items"`
		Total int           `json:"total"`
	} {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d: %s", path, w.Code, w.Body.String())
		}
		var resp struct {
			Items []domain.Post `json:"items"`
			Total int           `json:"total"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode %s: %v body=%s", path, err, w.Body.String())
		}
		return resp
	}

	qa := getPosts("/api/v1/admin/posts?site=portal&board=all&plugin_code=qa&content_type=question&page_size=1000")
	if qa.Total == 0 || len(qa.Items) == 0 {
		t.Fatalf("expected qa question rows, got total=%d items=%d", qa.Total, len(qa.Items))
	}
	for _, post := range qa.Items {
		if post.PluginCode != "qa" || post.ContentType != "question" {
			t.Fatalf("expected only qa/question rows, got id=%d plugin_code=%q content_type=%q title=%q", post.ID, post.PluginCode, post.ContentType, post.Title)
		}
	}

	docs := getPosts("/api/v1/admin/posts?site=portal&board=all&plugin_code=docs&content_type=document&page_size=1000")
	if docs.Total == 0 || len(docs.Items) == 0 {
		t.Fatalf("expected docs document rows, got total=%d items=%d", docs.Total, len(docs.Items))
	}
	for _, post := range docs.Items {
		if post.PluginCode != "docs" || post.ContentType != "document" {
			t.Fatalf("expected only docs/document rows, got id=%d plugin_code=%q content_type=%q title=%q", post.ID, post.PluginCode, post.ContentType, post.Title)
		}
	}

	mismatch := getPosts("/api/v1/admin/posts?site=portal&board=all&plugin_code=qa&content_type=document&page_size=1000")
	if mismatch.Total != 0 || len(mismatch.Items) != 0 {
		t.Fatalf("expected no rows for mismatched qa/document filter, got total=%d items=%d first=%#v", mismatch.Total, len(mismatch.Items), mismatch.Items)
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
