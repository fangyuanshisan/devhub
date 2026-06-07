package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"
	"devhub-gin-backend/internal/store"
)

func TestAdminWebhookEventsQueryRequiresAuthAndReturnsFilteredItems(t *testing.T) {
	repo := store.NewMemoryStore()
	_, _ = repo.AppendWebhookEvent(domain.WebhookEvent{
		EventID:     "evt_1",
		EventName:   "content.created",
		EventType:   "hook",
		PluginCode:  "qa",
		HookName:    "AfterCreateContent",
		Mode:        "non_blocking",
		CommunityID: 1,
		RequestID:   "req_1",
		Status:      domain.WebhookEventStatusDelivered,
	})
	_, _ = repo.AppendWebhookEvent(domain.WebhookEvent{
		EventID:     "evt_2",
		EventName:   "content.created",
		EventType:   "hook",
		PluginCode:  "docs",
		HookName:    "AfterCreateContent",
		Mode:        "non_blocking",
		CommunityID: 2,
		RequestID:   "req_2",
		Status:      domain.WebhookEventStatusFailed,
	})
	router := NewRouter(service.New(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/webhooks/events", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d: %s", w.Code, w.Body.String())
	}

	admin := adminToken(t, router)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/admin/plugins/webhooks/events?plugin_code=qa&status=delivered&page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+admin)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"event_id":"evt_1"`)) {
		t.Fatalf("expected qa event in response, got %s", w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte(`"event_id":"evt_2"`)) {
		t.Fatalf("did not expect docs event in filtered response, got %s", w.Body.String())
	}

	var body domain.WebhookEventListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Pagination.Page != 1 || body.Pagination.PageSize != 10 || body.Pagination.Total != 1 {
		t.Fatalf("unexpected pagination: %#v", body.Pagination)
	}
}
