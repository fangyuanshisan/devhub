package service

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestExternalServiceHealthCheckWarningAndRecovery(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	_, _ = svc.SetPluginStatus("qa", pluginregistry.StatusEnabled)

	fail := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("auth_type=none should not send Authorization header")
		}
		if fail {
			http.Error(w, "downstream unavailable", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	enabled := true
	cfg, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", domain.PluginExternalServiceUpdateRequest{
		EndpointURL:      server.URL,
		HealthCheckPath:  "/health",
		TimeoutMS:        800,
		FailurePolicy:    "warn",
		AuthType:         "none",
		Enabled:          &enabled,
		WarningThreshold: 1,
		ErrorThreshold:   2,
	})
	if err != nil {
		t.Fatalf("save external_service config: %v", err)
	}
	if cfg.TokenCiphertext != "" || cfg.TokenHash != "" {
		t.Fatal("auth_type=none should not persist token material")
	}

	res, err := svc.RunPluginExternalServiceHealthCheck(PluginExternalServiceOperator{Name: "tester"}, "qa")
	if err != nil {
		t.Fatalf("health check failure should be recorded, not returned as API error: %v", err)
	}
	if res.HealthStatus != "warning" {
		t.Fatalf("expected warning after threshold, got %#v", res)
	}
	saved, _ := repo.PluginExternalServiceConfig("qa")
	if saved.FailureCount != 1 {
		t.Fatalf("expected failure count 1, got %d", saved.FailureCount)
	}
	rows, total, err := repo.HookExecutionsByFilter(domain.HookExecutionFilter{PluginCode: "qa", ServiceType: "external_service", Page: 1, PageSize: 20})
	if err != nil || total != 1 || rows[0].Status == "" {
		t.Fatalf("expected external_service hook execution, total=%d rows=%#v err=%v", total, rows, err)
	}

	fail = false
	res, err = svc.RunPluginExternalServiceHealthCheck(PluginExternalServiceOperator{Name: "tester"}, "qa")
	if err != nil {
		t.Fatalf("health check recovery: %v", err)
	}
	if res.HealthStatus != "healthy" {
		t.Fatalf("expected healthy after recovery, got %#v", res)
	}
	saved, _ = repo.PluginExternalServiceConfig("qa")
	if saved.FailureCount != 0 {
		t.Fatalf("expected failure count reset, got %d", saved.FailureCount)
	}
}

func TestExternalServiceValidationAndDisabledPluginSkipped(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	enabled := true

	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", domain.PluginExternalServiceUpdateRequest{
		EndpointURL: "javascript:alert(1)",
		AuthType:    "none",
		Enabled:     &enabled,
	}); err == nil {
		t.Fatal("expected javascript endpoint to be rejected")
	}

	cfg, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", domain.PluginExternalServiceUpdateRequest{
		EndpointURL:      "http://127.0.0.1:65535",
		HealthCheckPath:  "/health",
		FailurePolicy:    "error",
		AuthType:         "none",
		Enabled:          &enabled,
		WarningThreshold: 1,
		ErrorThreshold:   2,
	})
	if err != nil {
		t.Fatalf("save local endpoint config: %v", err)
	}
	if cfg.HealthCheckPath != "/health" {
		t.Fatalf("unexpected health path %q", cfg.HealthCheckPath)
	}

	_, _ = svc.SetPluginStatus("qa", pluginregistry.StatusDisabled)
	res, err := svc.RunPluginExternalServiceHealthCheck(PluginExternalServiceOperator{Name: "tester"}, "qa")
	if err != nil {
		t.Fatalf("disabled plugin should record skipped without API error: %v", err)
	}
	if res.HealthStatus != "skipped" || res.Status != "skipped" {
		t.Fatalf("expected skipped health result, got %#v", res)
	}
	rows, total, err := repo.HookExecutionsByFilter(domain.HookExecutionFilter{PluginCode: "qa", ServiceType: "external_service", Page: 1, PageSize: 20})
	if err != nil || total == 0 {
		t.Fatalf("expected skipped execution record, total=%d err=%v", total, err)
	}
	if rows[0].Status != "skipped" {
		t.Fatalf("expected skipped execution status, got %#v", rows[0])
	}
}

func TestExternalServiceHTTPAllowlist(t *testing.T) {
	t.Setenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST", "")

	repo := store.NewMemoryStore()
	svc := New(repo)
	enabled := true

	req := domain.PluginExternalServiceUpdateRequest{
		EndpointURL:     "http://172.17.0.1:18081",
		HealthCheckPath: "/health",
		FailurePolicy:   "warn",
		AuthType:        "none",
		Enabled:         &enabled,
	}
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", req); err == nil {
		t.Fatal("expected non-local http endpoint to be rejected without allowlist")
	} else {
		var apiErr *domain.APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "endpoint_safety_rejected" || len(apiErr.Diagnostics) == 0 {
			t.Fatalf("expected endpoint_safety_rejected diagnostics, got %#v", err)
		}
	}
	req.EndpointURL = "http://172.17.0.1:18081/hooks?token=secret-value&foo=bar"
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", req); err == nil {
		t.Fatal("expected non-local http endpoint with token query to be rejected without allowlist")
	} else {
		var apiErr *domain.APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "endpoint_safety_rejected" {
			t.Fatalf("expected endpoint_safety_rejected diagnostics, got %#v", err)
		}
		raw, _ := json.Marshal(apiErr)
		if strings.Contains(string(raw), "secret-value") {
			t.Fatalf("diagnostics should redact token query, got %s", string(raw))
		}
	}

	t.Setenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST", "http://172.17.0.1:18081")
	req.EndpointURL = "http://172.17.0.1:18081"
	cfg, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", req)
	if err != nil {
		t.Fatalf("expected allowlisted http endpoint to be accepted: %v", err)
	}
	if cfg.EndpointURL != req.EndpointURL {
		t.Fatalf("unexpected endpoint URL %q", cfg.EndpointURL)
	}

	t.Setenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST", "")
	res, err := svc.RunPluginExternalServiceHealthCheck(PluginExternalServiceOperator{Name: "tester"}, "qa")
	if err != nil {
		t.Fatalf("runtime safety rejection should be returned as health response: %v", err)
	}
	if res.ErrorCode != "endpoint_safety_rejected" || len(res.Diagnostics) == 0 {
		t.Fatalf("expected runtime endpoint_safety_rejected diagnostics, got %#v", res)
	}

	t.Setenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST", "http://172.17.0.1:18081")
	req.EndpointURL = "http://172.17.0.2:18081"
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", req); err == nil {
		t.Fatal("expected non-allowlisted http endpoint to be rejected")
	}
}

func TestExternalServiceAdminHTTPAllowlist(t *testing.T) {
	t.Setenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST", "")

	repo := store.NewMemoryStore()
	svc := New(repo)
	enabled := true
	req := domain.PluginExternalServiceUpdateRequest{
		EndpointURL:     "http://172.17.0.1:18081",
		HealthCheckPath: "/health",
		FailurePolicy:   "warn",
		AuthType:        "none",
		Enabled:         &enabled,
	}
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", req); err == nil {
		t.Fatal("expected non-local http endpoint to be rejected before admin allowlist")
	}
	resp, err := svc.AddExternalServiceHTTPAllowlistOrigin(PluginExternalServiceOperator{Name: "admin", ID: 1}, domain.PluginExternalServiceHTTPAllowlistUpdateRequest{
		Origin:        "http://172.17.0.1:18081",
		Usage:         "local receiver",
		RiskConfirmed: true,
	})
	if err != nil {
		t.Fatalf("add admin allowlist: %v", err)
	}
	if len(resp.AdminAllowlist) != 1 || len(resp.EffectiveAllowlist) < 4 {
		t.Fatalf("unexpected allowlist response: %#v", resp)
	}
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", req); err != nil {
		t.Fatalf("expected admin allowlisted endpoint to be accepted: %v", err)
	}
	if _, err := svc.DeleteExternalServiceHTTPAllowlistOrigin(PluginExternalServiceOperator{Name: "admin", ID: 1}, resp.AdminAllowlist[0].ID); err != nil {
		t.Fatalf("delete admin allowlist: %v", err)
	}
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "qa", req); err == nil {
		t.Fatal("expected endpoint to be rejected after deleting admin allowlist")
	}
	logs := svc.AdminLogs("admin")
	foundCreate, foundDelete := false, false
	for _, log := range logs {
		foundCreate = foundCreate || log.Action == "system.external_service.http_allowlist.created"
		foundDelete = foundDelete || log.Action == "system.external_service.http_allowlist.deleted"
	}
	if !foundCreate || !foundDelete {
		t.Fatalf("expected create/delete audit logs, got %#v", logs)
	}
}

func TestExternalServiceHTTPAllowlistOriginValidation(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	bad := []string{
		"",
		"*",
		"http://*",
		"http://0.0.0.0:18081",
		"http://172.17.0.0/16",
		"http://172.17.0.1:18081/path",
		"http://172.17.0.1:18081?x=1",
		"http://172.17.0.1:18081#frag",
		"https://172.17.0.1:18081",
		"172.17.0.1:18081",
	}
	for _, origin := range bad {
		if _, err := svc.AddExternalServiceHTTPAllowlistOrigin(PluginExternalServiceOperator{Name: "admin"}, domain.PluginExternalServiceHTTPAllowlistUpdateRequest{Origin: origin, Usage: "bad", RiskConfirmed: true}); err == nil {
			t.Fatalf("expected origin %q to be rejected", origin)
		}
	}
	if _, err := svc.AddExternalServiceHTTPAllowlistOrigin(PluginExternalServiceOperator{Name: "admin"}, domain.PluginExternalServiceHTTPAllowlistUpdateRequest{Origin: "http://172.17.0.1:18081", RiskConfirmed: false}); err == nil {
		t.Fatal("expected missing risk confirmation to be rejected")
	}
}

func TestExternalServiceBearerTokenRefDeliveryAndRedaction(t *testing.T) {
	t.Setenv("CMS_STORE", "memory")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", "")

	repo := store.NewMemoryStore()
	svc := New(repo)
	code := "fixture_ext_tokenref"
	categoryID := installExternalServiceFixturePlugin(t, svc, code, code+".note", code+".note.create", []domain.HookDefinition{
		{
			Name:          pluginregistry.HookAfterCreateContent,
			Mode:          string(pluginregistry.HookNonBlocking),
			ServiceType:   "external_service",
			Path:          "/hooks/content.after_create",
			Method:        "POST",
			TimeoutMS:     800,
			RetryEnabled:  false,
			MaxAttempts:   1,
			FailurePolicy: "warn",
		},
	})

	const token = "s14-token-ref-secret"
	received := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	enabled := true
	cfg, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester", ID: 42}, code, domain.PluginExternalServiceUpdateRequest{
		EndpointURL:      server.URL,
		HealthCheckPath:  "/health",
		TimeoutMS:        800,
		FailurePolicy:    "warn",
		AuthType:         "bearer",
		Token:            token,
		Enabled:          &enabled,
		WarningThreshold: 1,
		ErrorThreshold:   2,
	})
	if err != nil {
		t.Fatalf("save bearer external_service config: %v", err)
	}
	wantRef := "secret://external_service/" + code + "/token"
	if cfg.TokenRef != wantRef || cfg.TokenCiphertext != "" || cfg.TokenHash != "" {
		t.Fatalf("expected token_ref only, got %#v", cfg)
	}
	rawCfg, _ := json.Marshal(cfg)
	if strings.Contains(string(rawCfg), token) {
		t.Fatalf("external_service response leaked token: %s", string(rawCfg))
	}
	secret, ok := repo.SecretRefByRef(wantRef)
	if !ok || secret.EncryptedValue == "" || strings.Contains(secret.EncryptedValue, token) {
		t.Fatalf("expected encrypted secret_ref storage, got ok=%v secret=%#v", ok, secret)
	}

	if res, err := svc.RunPluginExternalServiceHealthCheck(PluginExternalServiceOperator{Name: "tester"}, code); err != nil || res.HealthStatus != "healthy" {
		t.Fatalf("bearer health check should resolve token_ref, res=%#v err=%v", res, err)
	}
	select {
	case auth := <-received:
		if auth != "Bearer "+token {
			t.Fatalf("unexpected health auth header: %q", auth)
		}
	case <-time.After(time.Second):
		t.Fatal("expected health request")
	}

	if _, err := svc.CreateTopic(domain.CreateTopicRequest{
		UserID:       1,
		CommunityID:  1,
		CategoryID:   categoryID,
		Title:        "external service token ref",
		ContentType:  code + ".note",
		Content:      "hello",
		ActorContext: domain.ActorContext{UserID: 1, Permissions: []string{code + ".note.create"}},
	}); err != nil {
		t.Fatalf("create topic: %v", err)
	}
	select {
	case auth := <-received:
		if auth != "Bearer "+token {
			t.Fatalf("unexpected delivery auth header: %q", auth)
		}
	case <-time.After(time.Second):
		t.Fatal("expected delivery request")
	}
	rows := waitExternalServiceExecutions(t, repo, code, func(rows []domain.HookExecution) bool {
		return hasHookExecutionStatus(rows, "success")
	})
	assertNoSensitiveExecutionData(t, rows, token)
	assertNoSensitiveAdminLogData(t, svc, token)

	secret.Status = domain.SecretRefStatusRevoked
	secret.UpdatedAt = Now()
	if _, err := repo.SaveSecretRef(secret); err != nil {
		t.Fatalf("revoke fixture token_ref: %v", err)
	}
	if err := svc.DispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterCreateContent,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  code,
			ContentType: code + ".note",
			CommunityID: 1,
			ContentID:   999,
			RequestID:   "req-token-ref-revoked",
			ActorType:   pluginregistry.HookActorSystem,
		},
	}); err != nil {
		t.Fatalf("dispatch revoked token_ref hook: %v", err)
	}
	rows = waitExternalServiceExecutions(t, repo, code, func(rows []domain.HookExecution) bool {
		for _, row := range rows {
			if row.Status == "failed" && row.ErrorCode == "secret_ref_revoked" {
				return true
			}
		}
		return false
	})
	assertNoSensitiveExecutionData(t, rows, token)
	assertNoSensitiveAdminLogData(t, svc, token)
}

func TestExternalServiceNonBlockingHookDeliverySuccess(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	categoryID := installExternalServiceFixturePlugin(t, svc, "fixture_ext_success", "fixture_ext_success.note", "fixture_ext_success.note.create", []domain.HookDefinition{
		{
			Name:          pluginregistry.HookAfterCreateContent,
			Mode:          string(pluginregistry.HookNonBlocking),
			ServiceType:   "external_service",
			Path:          "/hooks/content.after_create",
			Method:        "POST",
			TimeoutMS:     800,
			RetryEnabled:  true,
			MaxAttempts:   2,
			FailurePolicy: "warn",
		},
	})
	received := make(chan *http.Request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("auth_type=none should not send Authorization header")
		}
		if r.Header.Get("X-DevHub-Delivery-Mode") != string(pluginregistry.HookNonBlocking) {
			t.Fatalf("unexpected delivery mode header: %q", r.Header.Get("X-DevHub-Delivery-Mode"))
		}
		received <- r.Clone(r.Context())
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	enabled := true
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "fixture_ext_success", domain.PluginExternalServiceUpdateRequest{
		EndpointURL:      server.URL,
		HealthCheckPath:  "/health",
		TimeoutMS:        800,
		FailurePolicy:    "warn",
		AuthType:         "none",
		Enabled:          &enabled,
		WarningThreshold: 1,
		ErrorThreshold:   2,
	}); err != nil {
		t.Fatalf("save external_service config: %v", err)
	}

	started := time.Now()
	topic, err := svc.CreateTopic(domain.CreateTopicRequest{
		UserID:       1,
		CommunityID:  1,
		CategoryID:   categoryID,
		Title:        "external service delivery",
		ContentType:  "fixture_ext_success.note",
		Content:      "hello",
		ActorContext: domain.ActorContext{UserID: 1, Permissions: []string{"fixture_ext_success.note.create"}},
	})
	if err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if topic == nil || topic.ID == 0 {
		t.Fatal("expected topic to be created")
	}
	if elapsed := time.Since(started); elapsed > 500*time.Millisecond {
		t.Fatalf("non-blocking hook should not wait for remote endpoint, elapsed=%s", elapsed)
	}
	select {
	case req := <-received:
		if req.URL.Path != "/hooks/content.after_create" {
			t.Fatalf("unexpected request path: %s", req.URL.Path)
		}
	case <-time.After(time.Second):
		t.Fatal("expected external_service hook delivery")
	}
	rows := waitExternalServiceExecutions(t, repo, "fixture_ext_success", func(rows []domain.HookExecution) bool {
		for _, row := range rows {
			if row.Status == "success" {
				return true
			}
		}
		return false
	})
	assertNoSensitiveExecutionData(t, rows)
	cfg, _ := repo.PluginExternalServiceConfig("fixture_ext_success")
	if cfg.LastHealthStatus != "healthy" || cfg.FailureCount != 0 {
		t.Fatalf("expected healthy external_service after success, got %#v", cfg)
	}
}

func TestExternalServiceSubscriberReceivesCoreArticleCreate(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	code := "fixture_article_audit_logger"
	manifest := domain.PluginManifest{
		Code:        code,
		Name:        "Fixture Article Audit Logger",
		Version:     "1.0.0",
		Description: "external_service subscriber fixture",
		Permissions: []domain.PermissionDefinition{
			{PluginCode: code, Code: code + ".manage", Name: "管理测试审计插件", Scope: "global"},
		},
		Hooks: []domain.HookDefinition{
			{
				PluginCode:    code,
				Name:          pluginregistry.HookAfterCreateContent,
				Mode:          string(pluginregistry.HookNonBlocking),
				ServiceType:   "external_service",
				Path:          "/hooks/content.after_create",
				Method:        "POST",
				TimeoutMS:     800,
				RetryEnabled:  true,
				MaxAttempts:   1,
				FailurePolicy: "warn",
			},
		},
		ExternalService: &domain.PluginExternalService{
			Endpoint:        "https://example.com",
			SubscribedHooks: []string{pluginregistry.HookAfterCreateContent},
			TimeoutMS:       800,
			FailurePolicy:   "warn",
			RetryPolicy:     "linear",
		},
		Migrations: []domain.PluginMigrationDefinition{},
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal subscriber manifest: %v", err)
	}
	if _, validation, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("install subscriber manifest: %v validation=%#v", err, validation)
	}
	if _, err := svc.SetPluginStatus(code, pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable subscriber plugin: %v", err)
	}
	if _, err := svc.SetCommunityPluginStatus(1, code, pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable subscriber plugin for community: %v", err)
	}

	received := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		received <- body
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	enabled := true
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, code, domain.PluginExternalServiceUpdateRequest{
		EndpointURL:      server.URL,
		HealthCheckPath:  "/health",
		TimeoutMS:        800,
		FailurePolicy:    "warn",
		AuthType:         "none",
		Enabled:          &enabled,
		WarningThreshold: 1,
		ErrorThreshold:   2,
	}); err != nil {
		t.Fatalf("save subscriber external_service config: %v", err)
	}

	topic, err := svc.CreateTopic(domain.CreateTopicRequest{
		UserID:       1,
		CommunityID:  1,
		Title:        "core article should notify subscriber",
		ContentType:  "article",
		Content:      "hello from core article",
		ActorContext: domain.ActorContext{UserID: 1, Permissions: []string{"core.topic.create"}},
	})
	if err != nil {
		t.Fatalf("create core article: %v", err)
	}
	if topic.PluginCode != pluginregistry.CoreCode {
		t.Fatalf("expected core topic, got plugin_code=%s", topic.PluginCode)
	}

	select {
	case body := <-received:
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode webhook payload: %v", err)
		}
		if payload["plugin_code"] != code {
			t.Fatalf("expected subscriber plugin_code, got %#v", payload["plugin_code"])
		}
		data, _ := payload["data"].(map[string]any)
		if data["content_type"] != "article" {
			t.Fatalf("expected article payload, got %#v", data)
		}
		metadata, _ := data["metadata"].(map[string]any)
		if metadata["source_plugin_code"] != pluginregistry.CoreCode {
			t.Fatalf("expected source core metadata, got %#v", metadata)
		}
	case <-time.After(time.Second):
		t.Fatal("expected subscriber external_service webhook delivery")
	}
}

func TestExternalServiceNonBlockingHookRetryWarningAndSkipped(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	categoryID := installExternalServiceFixturePlugin(t, svc, "fixture_ext_retry", "fixture_ext_retry.note", "fixture_ext_retry.note.create", []domain.HookDefinition{
		{
			Name:          pluginregistry.HookAfterCreateContent,
			Mode:          string(pluginregistry.HookNonBlocking),
			ServiceType:   "external_service",
			Path:          "/hooks/content.after_create",
			Method:        "POST",
			TimeoutMS:     800,
			RetryEnabled:  true,
			MaxAttempts:   2,
			FailurePolicy: "error",
		},
	})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "temporary failure", http.StatusInternalServerError)
	}))
	defer server.Close()

	enabled := true
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "fixture_ext_retry", domain.PluginExternalServiceUpdateRequest{
		EndpointURL:      server.URL,
		TimeoutMS:        800,
		FailurePolicy:    "error",
		AuthType:         "none",
		Enabled:          &enabled,
		WarningThreshold: 1,
		ErrorThreshold:   2,
	}); err != nil {
		t.Fatalf("save external_service config: %v", err)
	}
	if _, err := svc.CreateTopic(domain.CreateTopicRequest{
		UserID:       1,
		CommunityID:  1,
		CategoryID:   categoryID,
		Title:        "external service retry",
		ContentType:  "fixture_ext_retry.note",
		Content:      "hello",
		ActorContext: domain.ActorContext{UserID: 1, Permissions: []string{"fixture_ext_retry.note.create"}},
	}); err != nil {
		t.Fatalf("create topic: %v", err)
	}
	rows := waitExternalServiceExecutions(t, repo, "fixture_ext_retry", func(rows []domain.HookExecution) bool {
		return hasHookExecutionStatus(rows, "retry_scheduled") && hasHookExecutionStatus(rows, "retry_exhausted")
	})
	assertNoSensitiveExecutionData(t, rows)
	cfg, _ := repo.PluginExternalServiceConfig("fixture_ext_retry")
	if cfg.LastHealthStatus != "error" {
		t.Fatalf("expected health error after exhausted retries, got %#v", cfg)
	}

	if _, err := svc.SetPluginStatus("fixture_ext_retry", pluginregistry.StatusDisabled); err != nil {
		t.Fatalf("disable fixture plugin: %v", err)
	}
	if err := svc.DispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterCreateContent,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  "fixture_ext_retry",
			ContentType: "fixture_ext_retry.note",
			CommunityID: 1,
			ContentID:   99,
			ActorType:   pluginregistry.HookActorSystem,
		},
	}); err != nil {
		t.Fatalf("dispatch disabled hook: %v", err)
	}
	rows = waitExternalServiceExecutions(t, repo, "fixture_ext_retry", func(rows []domain.HookExecution) bool {
		for _, row := range rows {
			if row.Status == "skipped" && row.ErrorCode == "PLUGIN_DISABLED" {
				return true
			}
		}
		return false
	})
	assertNoSensitiveExecutionData(t, rows)
}

func TestExternalServiceManifestRejectsBlockingHook(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	raw := []byte(`{
		"code": "fixture_ext_invalid",
		"name": "Invalid External Service",
		"version": "1.0.0",
		"content_types": [],
		"permissions": [],
		"hooks": [
			{"name": "AfterCreateContent", "mode": "blocking", "service_type": "external_service", "path": "/hooks/content.after_create", "method": "POST", "failure_policy": "warn"}
		]
	}`)
	validation, err := svc.ValidatePluginManifestJSON(raw)
	if err != nil {
		t.Fatalf("validate manifest: %v", err)
	}
	if validation.Valid {
		t.Fatalf("expected external_service blocking Hook to be rejected: %#v", validation)
	}
	found := false
	for _, msg := range validation.Errors {
		if strings.Contains(msg, "只支持 non_blocking") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected non_blocking validation error, got %#v", validation.Errors)
	}
}

func TestExternalServiceManualRetrySuccessAndForbiddenStates(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	_ = installExternalServiceFixturePlugin(t, svc, "fixture_ext_manual", "fixture_ext_manual.note", "fixture_ext_manual.note.create", []domain.HookDefinition{
		{
			Name:          pluginregistry.HookAfterCreateContent,
			Mode:          string(pluginregistry.HookNonBlocking),
			ServiceType:   "external_service",
			Path:          "/hooks/content.after_create",
			Method:        "POST",
			TimeoutMS:     800,
			RetryEnabled:  true,
			MaxAttempts:   2,
			FailurePolicy: "warn",
		},
	})
	fail := true
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if strings.Contains(strings.ToLower(r.Header.Get("Authorization")), "bearer") {
			t.Fatal("auth_type=none should not send Authorization")
		}
		if fail {
			http.Error(w, "temporary failure", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	enabled := true
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "fixture_ext_manual", domain.PluginExternalServiceUpdateRequest{
		EndpointURL:      server.URL,
		TimeoutMS:        800,
		FailurePolicy:    "warn",
		AuthType:         "none",
		Enabled:          &enabled,
		WarningThreshold: 1,
		ErrorThreshold:   2,
	}); err != nil {
		t.Fatalf("save external_service config: %v", err)
	}
	if err := svc.DispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterCreateContent,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  "fixture_ext_manual",
			ContentType: "fixture_ext_manual.note",
			CommunityID: 1,
			ContentID:   100,
			RequestID:   "req-manual-retry",
			ActorType:   pluginregistry.HookActorSystem,
		},
	}); err != nil {
		t.Fatalf("dispatch hook: %v", err)
	}
	rows := waitExternalServiceExecutions(t, repo, "fixture_ext_manual", func(rows []domain.HookExecution) bool {
		return hasHookExecutionStatus(rows, "retry_exhausted")
	})
	var failed domain.HookExecution
	for _, row := range rows {
		if row.Status == "retry_exhausted" {
			failed = row
			break
		}
	}
	if failed.ID == 0 {
		t.Fatal("expected failed source execution")
	}

	fail = false
	resp, err := svc.ManualRetryExternalServiceHookExecution(PluginExternalServiceOperator{Name: "operator", ID: 7}, "fixture_ext_manual", failed.ID)
	if err != nil {
		t.Fatalf("manual retry should succeed: %v", err)
	}
	if resp.Status != "success" || resp.RetryExecutionID == "" || resp.SourceExecutionID == "" {
		t.Fatalf("unexpected retry response: %#v", resp)
	}
	rows = waitExternalServiceExecutions(t, repo, "fixture_ext_manual", func(rows []domain.HookExecution) bool {
		return hasHookExecutionStatus(rows, "success")
	})
	assertNoSensitiveExecutionData(t, rows)
	if requests == 0 {
		t.Fatal("expected retry to call mock receiver")
	}

	var success domain.HookExecution
	for _, row := range rows {
		if row.Status == "success" {
			success = row
			break
		}
	}
	if _, err := svc.ManualRetryExternalServiceHookExecution(PluginExternalServiceOperator{Name: "operator"}, "fixture_ext_manual", success.ID); err == nil {
		t.Fatal("success execution should not be retryable")
	}
	if _, err := svc.ManualRetryExternalServiceHookExecution(PluginExternalServiceOperator{Name: "operator"}, "qa", failed.ID); err == nil {
		t.Fatal("cross-plugin retry should be rejected")
	}
}

func TestExternalServiceManualRetryRejectsSkippedAndDisabledPlugin(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)
	_ = installExternalServiceFixturePlugin(t, svc, "fixture_ext_manual_guard", "fixture_ext_manual_guard.note", "fixture_ext_manual_guard.note.create", []domain.HookDefinition{
		{
			Name:          pluginregistry.HookAfterCreateContent,
			Mode:          string(pluginregistry.HookNonBlocking),
			ServiceType:   "external_service",
			Path:          "/hooks/content.after_create",
			Method:        "POST",
			TimeoutMS:     800,
			RetryEnabled:  true,
			MaxAttempts:   1,
			FailurePolicy: "warn",
		},
	})
	enabled := true
	if _, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{Name: "tester"}, "fixture_ext_manual_guard", domain.PluginExternalServiceUpdateRequest{
		EndpointURL:      "http://127.0.0.1:65535",
		TimeoutMS:        800,
		FailurePolicy:    "warn",
		AuthType:         "none",
		Enabled:          &enabled,
		WarningThreshold: 1,
		ErrorThreshold:   2,
	}); err != nil {
		t.Fatalf("save external_service config: %v", err)
	}
	if _, err := svc.SetPluginStatus("fixture_ext_manual_guard", pluginregistry.StatusDisabled); err != nil {
		t.Fatalf("disable plugin: %v", err)
	}
	if err := svc.DispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterCreateContent,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode:  "fixture_ext_manual_guard",
			ContentType: "fixture_ext_manual_guard.note",
			CommunityID: 1,
			ContentID:   101,
			ActorType:   pluginregistry.HookActorSystem,
		},
	}); err != nil {
		t.Fatalf("dispatch disabled hook: %v", err)
	}
	rows := waitExternalServiceExecutions(t, repo, "fixture_ext_manual_guard", func(rows []domain.HookExecution) bool {
		return hasHookExecutionStatus(rows, "skipped")
	})
	var skipped domain.HookExecution
	for _, row := range rows {
		if row.Status == "skipped" {
			skipped = row
			break
		}
	}
	if _, err := svc.ManualRetryExternalServiceHookExecution(PluginExternalServiceOperator{Name: "operator"}, "fixture_ext_manual_guard", skipped.ID); err == nil {
		t.Fatal("skipped execution should not be retryable")
	}

	if _, err := svc.SetPluginStatus("fixture_ext_manual_guard", pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable plugin: %v", err)
	}
	failed, err := repo.AppendHookExecution(domain.HookExecution{
		HookName:    pluginregistry.HookAfterCreateContent,
		PluginCode:  "fixture_ext_manual_guard",
		ServiceType: "external_service",
		Mode:        string(pluginregistry.HookNonBlocking),
		Status:      "failed",
		Success:     false,
		Metadata:    `{"execution_id":"source-disabled"}`,
	})
	if err != nil {
		t.Fatalf("append failed execution: %v", err)
	}
	if _, err := svc.SetPluginStatus("fixture_ext_manual_guard", pluginregistry.StatusDisabled); err != nil {
		t.Fatalf("disable plugin again: %v", err)
	}
	before := len(rows)
	_, err = svc.ManualRetryExternalServiceHookExecution(PluginExternalServiceOperator{Name: "operator"}, "fixture_ext_manual_guard", failed.ID)
	if err == nil {
		t.Fatal("disabled plugin should reject manual retry")
	}
	afterRows, _, err := repo.HookExecutionsByFilter(domain.HookExecutionFilter{PluginCode: "fixture_ext_manual_guard", ServiceType: "external_service", Page: 1, PageSize: 100})
	if err != nil {
		t.Fatalf("query executions: %v", err)
	}
	if len(afterRows) != before+1 {
		t.Fatalf("disabled retry should not create a new delivery record, before=%d after=%d", before, len(afterRows))
	}
}

func installExternalServiceFixturePlugin(t *testing.T, svc *Service, code, contentType, permission string, hooks []domain.HookDefinition) int64 {
	t.Helper()
	manifest := domain.PluginManifest{
		Code:        code,
		Name:        code,
		Version:     "1.0.0",
		Description: "external_service non-blocking fixture",
		ContentTypes: []string{
			contentType,
		},
		ContentTypeDefs: []domain.ContentTypeDefinition{
			{Type: contentType, Name: contentType, PluginCode: code, CreatePermission: permission, AllowComment: true},
		},
		Permissions: []domain.PermissionDefinition{
			{PluginCode: code, Code: permission, Name: "创建测试内容", Scope: "community"},
		},
		Hooks:      hooks,
		Migrations: []domain.PluginMigrationDefinition{},
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal fixture manifest: %v", err)
	}
	if _, validation, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("install fixture manifest: %v validation=%#v", err, validation)
	}
	if _, err := svc.SetPluginStatus(code, pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable fixture plugin: %v", err)
	}
	if _, err := svc.SetCommunityPluginStatus(1, code, pluginregistry.StatusEnabled); err != nil {
		t.Fatalf("enable fixture community plugin: %v", err)
	}
	category, err := svc.CreateCategory(1, domain.CategoryRequest{
		Name:                code,
		Slug:                code,
		ContentType:         contentType,
		PluginCode:          code,
		AllowedContentTypes: []string{contentType},
	})
	if err != nil {
		t.Fatalf("create fixture category: %v", err)
	}
	return category.ID
}

func waitExternalServiceExecutions(t *testing.T, repo *store.MemoryStore, pluginCode string, ok func([]domain.HookExecution) bool) []domain.HookExecution {
	t.Helper()
	var rows []domain.HookExecution
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		items, _, err := repo.HookExecutionsByFilter(domain.HookExecutionFilter{PluginCode: pluginCode, ServiceType: "external_service", Page: 1, PageSize: 100})
		if err != nil {
			t.Fatalf("query hook executions: %v", err)
		}
		rows = items
		if ok(items) {
			return items
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for external_service executions, got %#v", rows)
	return rows
}

func hasHookExecutionStatus(rows []domain.HookExecution, status string) bool {
	for _, row := range rows {
		if row.Status == status {
			return true
		}
	}
	return false
}

func assertNoSensitiveExecutionData(t *testing.T, rows []domain.HookExecution, extraForbidden ...string) {
	t.Helper()
	forbidden := []string{"authorization", "bearer ", "webhook secret", "callback token"}
	for _, item := range extraForbidden {
		if item = strings.ToLower(strings.TrimSpace(item)); item != "" {
			forbidden = append(forbidden, item)
		}
	}
	for _, row := range rows {
		raw := strings.ToLower(row.Metadata + row.ErrorMessage + row.ResponseBodyExcerpt)
		for _, marker := range forbidden {
			if strings.Contains(raw, marker) {
				t.Fatalf("hook_execution leaked sensitive marker %q in %#v", marker, row)
			}
		}
	}
}

func assertNoSensitiveAdminLogData(t *testing.T, svc *Service, extraForbidden ...string) {
	t.Helper()
	forbidden := []string{"authorization", "bearer ", "webhook secret", "callback token", "encrypted_value"}
	for _, item := range extraForbidden {
		if item = strings.ToLower(strings.TrimSpace(item)); item != "" {
			forbidden = append(forbidden, item)
		}
	}
	logs := svc.AdminLogs("admin")
	for _, log := range logs {
		raw := strings.ToLower(log.Action + log.Target + log.OldValue + log.NewValue + log.Metadata)
		for _, marker := range forbidden {
			if strings.Contains(raw, marker) {
				t.Fatalf("admin_log leaked sensitive marker %q in %#v", marker, log)
			}
		}
	}
}
