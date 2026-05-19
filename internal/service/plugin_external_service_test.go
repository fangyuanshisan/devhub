package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
