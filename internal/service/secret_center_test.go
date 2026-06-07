package service

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestSecretCenter_CreateResolveDisableRevoke(t *testing.T) {
	t.Setenv("CMS_STORE", "memory")
	// Ensure no real root key is needed for this unit test.
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", "")

	repo := store.NewMemoryStore()
	svc := New(repo)

	op := SecretOperator{Type: "admin_user", ID: 100, Name: "tester"}
	created, err := svc.CreateSecret(op, domain.SecretCreateRequest{
		Namespace:   "external_service",
		Name:        "feishu_link/token",
		Value:       "secret-value",
		Description: "unit secret",
	})
	if err != nil {
		t.Fatalf("create secret: %v", err)
	}
	if created.Ref == "" || created.KeyID == "" || created.Status != domain.SecretRefStatusActive {
		t.Fatalf("unexpected created record: %#v", created)
	}
	if created.EncryptedValue != "" {
		t.Fatalf("api record should not include encrypted_value, got %#v", created)
	}

	plain, meta, err := svc.ResolveSecretInternal(created.Ref)
	if err != nil {
		t.Fatalf("resolve secret: %v", err)
	}
	if plain != "secret-value" {
		t.Fatalf("unexpected plain: %q", plain)
	}
	if meta.Ref != created.Ref || meta.EncryptedValue != "" {
		t.Fatalf("unexpected meta: %#v", meta)
	}

	// Disable should block resolve.
	if _, err := svc.DisableSecret(op, domain.SecretStatusRequest{Ref: created.Ref}); err != nil {
		t.Fatalf("disable secret: %v", err)
	}
	if _, _, err := svc.ResolveSecretInternal(created.Ref); err == nil {
		t.Fatal("expected resolve to fail after disable")
	} else {
		var apiErr *domain.APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "secret_ref_disabled" {
			t.Fatalf("expected secret_ref_disabled, got %#v", err)
		}
	}

	// Update should re-activate.
	if _, err := svc.UpdateSecret(op, domain.SecretUpdateRequest{Ref: created.Ref, Value: "rotated"}); err != nil {
		t.Fatalf("update secret: %v", err)
	}
	plain, _, err = svc.ResolveSecretInternal(created.Ref)
	if err != nil {
		t.Fatalf("resolve after update: %v", err)
	}
	if plain != "rotated" {
		t.Fatalf("unexpected rotated value: %q", plain)
	}

	// Revoke should block resolve.
	if _, err := svc.RevokeSecret(op, domain.SecretStatusRequest{Ref: created.Ref, ConfirmRef: created.Ref}); err != nil {
		t.Fatalf("revoke secret: %v", err)
	}
	if _, _, err := svc.ResolveSecretInternal(created.Ref); err == nil {
		t.Fatal("expected resolve to fail after revoke")
	} else {
		var apiErr *domain.APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "secret_ref_revoked" {
			t.Fatalf("expected secret_ref_revoked, got %#v", err)
		}
	}
}

func TestSecretCenter_DetailUsagesImpactAndSourceCredentials(t *testing.T) {
	t.Setenv("CMS_STORE", "memory")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", "")

	repo := store.NewMemoryStore()
	svc := New(repo)
	op := SecretOperator{Type: "admin_user", ID: 101, Name: "tester"}

	enabled := true
	cfg, err := svc.UpdatePluginExternalServiceConfig(PluginExternalServiceOperator{ID: op.ID, Name: op.Name}, "qa", domain.PluginExternalServiceUpdateRequest{
		EndpointURL:     "http://127.0.0.1:18081",
		HealthCheckPath: "/health",
		AuthType:        "bearer",
		Token:           "external-token-plain",
		Enabled:         &enabled,
	})
	if err != nil {
		t.Fatalf("save external service: %v", err)
	}
	if cfg.TokenRef == "" || strings.Contains(mustJSONForTest(cfg), "external-token-plain") {
		t.Fatalf("expected token_ref without plaintext, got %#v", cfg)
	}

	detail, err := svc.GetSecretDetail(cfg.TokenRef)
	if err != nil {
		t.Fatalf("get external secret detail: %v", err)
	}
	if detail.Record.EncryptedValue != "" {
		t.Fatalf("detail must not expose encrypted_value: %#v", detail.Record)
	}
	if detail.Source.Type != "external_service" || detail.Source.PluginCode != "qa" {
		t.Fatalf("unexpected source: %#v", detail.Source)
	}
	if detail.Record.UsageType != "external_service_token" || detail.Record.SourceType != "external_service" || detail.Record.SourceID != "qa" || detail.Record.SourceCode != "qa" {
		t.Fatalf("unexpected external secret metadata: %#v", detail.Record)
	}
	if len(detail.Usages) == 0 || detail.Usages[0].Type != "external_service" || detail.Usages[0].PluginCode != "qa" {
		t.Fatalf("unexpected usages: %#v", detail.Usages)
	}
	if detail.Usages[0].UsageType != "external_service_token" || detail.Usages[0].SourceID != "qa" {
		t.Fatalf("unexpected usage source metadata: %#v", detail.Usages[0])
	}
	if raw := mustJSONForTest(detail); strings.Contains(raw, "external-token-plain") || strings.Contains(raw, "encrypted_value") {
		t.Fatalf("detail leaked sensitive content: %s", raw)
	}

	preview, err := svc.SecretImpactPreview(cfg.TokenRef, "disable")
	if err != nil {
		t.Fatalf("impact preview: %v", err)
	}
	if !preview.Allowed || preview.AffectedPlugins != 1 || preview.AffectedExternalServices != 1 || preview.PossibleFailedHealthChecks != 1 {
		t.Fatalf("unexpected impact preview: %#v", preview)
	}
	effective, err := svc.SystemEffectiveConfig()
	if err != nil {
		t.Fatalf("system effective config: %v", err)
	}
	if len(effective.ExternalServices) == 0 || effective.ExternalServices[0].ConfigSource != "plugin runtime config" {
		t.Fatalf("expected config source in effective config: %#v", effective.ExternalServices)
	}
	if raw := mustJSONForTest(effective); strings.Contains(raw, "external-token-plain") || strings.Contains(raw, `"encrypted_value":`) || strings.Contains(effective.DiagnosticText, "encrypted_value") {
		t.Fatalf("effective config leaked sensitive content: %s", raw)
	}
	if _, err := svc.DisableSecret(op, domain.SecretStatusRequest{Ref: cfg.TokenRef}); err != nil {
		t.Fatalf("disable external service secret: %v", err)
	}
	effective, err = svc.SystemEffectiveConfig()
	if err != nil {
		t.Fatalf("system effective config after disable: %v", err)
	}
	if len(effective.ExternalServices) == 0 || effective.ExternalServices[0].TokenStatus != domain.SecretRefStatusDisabled || len(effective.ExternalServices[0].NextSteps) == 0 {
		t.Fatalf("expected disabled secret next steps: %#v", effective.ExternalServices)
	}

	webhook, err := svc.CreatePluginWebhookSecret(WebhookSecretOperator{ID: op.ID, Name: op.Name}, CreateWebhookSecretRequest{
		PluginCode: "qa",
		TargetURL:  "https://plugin.example.test/hooks?token=should-redact",
	})
	if err != nil {
		t.Fatalf("create webhook secret: %v", err)
	}
	whDetail, err := svc.GetSecretDetail(webhook.Secret.SecretRef)
	if err != nil {
		t.Fatalf("get webhook secret detail: %v", err)
	}
	if whDetail.Source.Type != "webhook" || len(whDetail.Usages) == 0 || whDetail.Usages[0].Type != "webhook" {
		t.Fatalf("unexpected webhook detail: %#v", whDetail)
	}
	if whDetail.Record.UsageType != "webhook_secret" || whDetail.Source.SourceCode != "qa" || whDetail.Usages[0].SourceCode != "qa" {
		t.Fatalf("unexpected webhook source metadata: %#v", whDetail)
	}
	if raw := mustJSONForTest(whDetail); strings.Contains(raw, webhook.SecretPlaintext) || strings.Contains(raw, "should-redact") {
		t.Fatalf("webhook detail leaked sensitive content: %s", raw)
	}
	if _, err := svc.RevokeSecret(op, domain.SecretStatusRequest{Ref: webhook.Secret.SecretRef, ConfirmRef: webhook.Secret.SecretRef}); err != nil {
		t.Fatalf("revoke webhook source secret: %v", err)
	}
	logs, total := repo.AdminLogsByFilter(domain.AdminLogFilter{Action: "secret_center.secret.revoked", Target: "secret_refs#" + webhook.Secret.SecretRef, Page: 1, PageSize: 10})
	if total == 0 || len(logs) == 0 {
		t.Fatalf("expected secret_center revoke audit log, total=%d logs=%#v", total, logs)
	}
	if !strings.Contains(logs[0].Metadata, `"source_type":"webhook"`) || !strings.Contains(logs[0].Metadata, `"source_code":"qa"`) {
		t.Fatalf("expected source metadata in secret audit log: %#v", logs[0])
	}

	callback, err := svc.CreatePluginCallbackToken(CallbackTokenOperator{Type: "admin_user", ID: op.ID, Name: op.Name}, CreatePluginCallbackTokenRequest{
		PluginCode:     "qa",
		Name:           "qa callback",
		Scopes:         []string{"config.read"},
		CommunityScope: []int64{1},
	})
	if err != nil {
		t.Fatalf("create callback token: %v", err)
	}
	cbDetail, err := svc.GetSecretDetail(callback.TokenRef)
	if err != nil {
		t.Fatalf("get callback token detail: %v", err)
	}
	if cbDetail.Source.Type != "callback" || len(cbDetail.Usages) == 0 || cbDetail.Usages[0].Type != "callback" {
		t.Fatalf("unexpected callback detail: %#v", cbDetail)
	}
	if cbDetail.Record.UsageType != "callback_token" || cbDetail.Source.SourceCode != "qa" || cbDetail.Usages[0].SourceCode != "qa" {
		t.Fatalf("unexpected callback source metadata: %#v", cbDetail)
	}
	if raw := mustJSONForTest(cbDetail); strings.Contains(raw, callback.Token) || strings.Contains(raw, "token_hash") {
		t.Fatalf("callback detail leaked sensitive content: %s", raw)
	}

	pluginConfigSecret, err := svc.CreateSecret(op, domain.SecretCreateRequest{
		Namespace: "plugin_config",
		Name:      "qa/api_key",
		Value:     "plugin-config-secret-plain",
	})
	if err != nil {
		t.Fatalf("create plugin config secret: %v", err)
	}
	pcDetail, err := svc.GetSecretDetail(pluginConfigSecret.Ref)
	if err != nil {
		t.Fatalf("get plugin config secret detail: %v", err)
	}
	if pcDetail.Source.Type != "plugin_config" || len(pcDetail.Usages) == 0 || pcDetail.Usages[0].Type != "plugin_config_sensitive_field" {
		t.Fatalf("unexpected plugin config detail: %#v", pcDetail)
	}
	if pcDetail.Record.UsageType != "plugin_config_sensitive_field" || pcDetail.Source.SourceID != "qa" || pcDetail.Usages[0].SourceID != "qa" {
		t.Fatalf("unexpected plugin config source metadata: %#v", pcDetail)
	}
	if raw := mustJSONForTest(pcDetail); strings.Contains(raw, "plugin-config-secret-plain") || strings.Contains(raw, "encrypted_value") {
		t.Fatalf("plugin config detail leaked sensitive content: %s", raw)
	}

	seedSecret, err := svc.CreateSecret(op, domain.SecretCreateRequest{
		Namespace: "seed",
		Name:      "demo/token",
		Value:     "seed-secret-plain",
	})
	if err != nil {
		t.Fatalf("create seed secret: %v", err)
	}
	seedDetail, err := svc.GetSecretDetail(seedSecret.Ref)
	if err != nil {
		t.Fatalf("get seed secret detail: %v", err)
	}
	if seedDetail.Record.SourceType != "test" || seedDetail.Record.UsageType != "test_fixture_seed" || len(seedDetail.Usages) == 0 || seedDetail.Usages[0].Type != "test_fixture_seed" {
		t.Fatalf("unexpected seed secret usage: %#v", seedDetail)
	}
	if preview, err := svc.SecretImpactPreview(seedSecret.Ref, "disable"); err != nil || preview.Allowed {
		t.Fatalf("expected seed secret dangerous action to be blocked, preview=%#v err=%v", preview, err)
	}
	if raw := mustJSONForTest(seedDetail); strings.Contains(raw, "seed-secret-plain") || strings.Contains(raw, "encrypted_value") {
		t.Fatalf("seed detail leaked sensitive content: %s", raw)
	}
}

func TestSecretCenter_RevokeRequiresConfirmAndAuditsFailure(t *testing.T) {
	t.Setenv("CMS_STORE", "memory")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", "")

	repo := store.NewMemoryStore()
	svc := New(repo)
	op := SecretOperator{Type: "admin_user", ID: 102, Name: "tester"}
	created, err := svc.CreateSecret(op, domain.SecretCreateRequest{
		Namespace: "custom",
		Name:      "prod/token",
		Value:     "secret-value",
	})
	if err != nil {
		t.Fatalf("create secret: %v", err)
	}
	if _, err := svc.RevokeSecret(op, domain.SecretStatusRequest{Ref: created.Ref}); err == nil {
		t.Fatal("expected revoke without confirm to fail")
	}
	_, total := repo.AdminLogsByFilter(domain.AdminLogFilter{Action: "secret_center.secret.revoke.failed", Target: "secret_refs#" + created.Ref, Page: 1, PageSize: 10})
	if total == 0 {
		t.Fatal("expected failed revoke audit log")
	}
}

func mustJSONForTest(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(raw)
}

func TestSecretCenter_CreateRequiresStartupKeyOutsideMemoryMode(t *testing.T) {
	t.Setenv("CMS_STORE", "")
	t.Setenv("DEVHUB_E2E_TESTING", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", "")

	svc := New(store.NewMemoryStore())
	_, err := svc.CreateSecret(SecretOperator{Type: "admin_user", ID: 1, Name: "tester"}, domain.SecretCreateRequest{
		Namespace: "external_service",
		Name:      "feishu_link/token",
		Value:     "secret-value",
	})
	if err == nil {
		t.Fatal("expected missing startup key to block secret creation")
	}
	var apiErr *domain.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T %[1]v", err)
	}
	if apiErr.Code != "secret_center_encryption_key_missing" {
		t.Fatalf("unexpected code: %s", apiErr.Code)
	}
	for _, want := range []string{"DEVHUB_PLUGIN_CONFIG_KEYS", "DEVHUB_PLUGIN_CONFIG_KEY_ID", "DEVHUB_PLUGIN_CONFIG_KEY", "后台页面中创建或保存"} {
		if !strings.Contains(apiErr.Suggestion, want) {
			t.Fatalf("missing %q in suggestion: %s", want, apiErr.Suggestion)
		}
	}
}

func TestSystemEffectiveConfigRootKeyMissingDiagnosticsAreRedacted(t *testing.T) {
	t.Setenv("CMS_STORE", "")
	t.Setenv("DEVHUB_E2E_TESTING", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", "")

	svc := New(store.NewMemoryStore())
	resp, err := svc.SystemEffectiveConfig()
	if err != nil {
		t.Fatalf("system effective config: %v", err)
	}
	if len(resp.NextSteps) == 0 {
		t.Fatalf("expected root-key next step: %#v", resp)
	}
	if raw := mustJSONForTest(resp); strings.Contains(raw, `"encrypted_value":`) || strings.Contains(resp.DiagnosticText, "encrypted_value") || strings.Contains(raw, "DEVHUB_PLUGIN_CONFIG_KEYS='[{") {
		t.Fatalf("effective config diagnostic should stay redacted: %s", raw)
	}
}

func TestSystemEffectiveConfigTroubleshootingS3RedactionAndAllowlist(t *testing.T) {
	t.Setenv("CMS_STORE", "memory")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", "")
	t.Setenv("DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST", "")

	repo := store.NewMemoryStore()
	svc := New(repo)
	op := PluginExternalServiceOperator{ID: 103, Name: "tester"}
	if _, err := svc.AddExternalServiceHTTPAllowlistOrigin(op, domain.PluginExternalServiceHTTPAllowlistUpdateRequest{
		Origin:        "http://172.17.0.1:18081",
		Usage:         "unit test receiver",
		RiskConfirmed: true,
	}); err != nil {
		t.Fatalf("add allowlist: %v", err)
	}
	enabled := true
	cfg, err := svc.UpdatePluginExternalServiceConfig(op, "docs", domain.PluginExternalServiceUpdateRequest{
		EndpointURL:     "http://172.17.0.1:18081?token=diagnostic-plain-token",
		HealthCheckPath: "/health",
		AuthType:        "bearer",
		Token:           "external-token-plain-s3",
		TimeoutMS:       2500,
		FailurePolicy:   "warn",
		Enabled:         &enabled,
	})
	if err != nil {
		t.Fatalf("save external service: %v", err)
	}
	if _, err := svc.CreatePluginWebhookSecret(WebhookSecretOperator{ID: 103, Name: "tester"}, CreateWebhookSecretRequest{
		PluginCode: "docs",
		TargetURL:  "https://plugin.example.test/hooks",
	}); err != nil {
		t.Fatalf("create webhook secret: %v", err)
	}
	if _, err := svc.CreatePluginCallbackToken(CallbackTokenOperator{Type: "admin_user", ID: 103, Name: "tester"}, CreatePluginCallbackTokenRequest{
		PluginCode:     "docs",
		Name:           "docs callback",
		Scopes:         []string{"config.read"},
		CommunityScope: []int64{1},
	}); err != nil {
		t.Fatalf("create callback token: %v", err)
	}
	resp, err := svc.SystemEffectiveConfig()
	if err != nil {
		t.Fatalf("system effective config: %v", err)
	}
	if resp.HTTPAllowlistSource != "admin_setting" {
		t.Fatalf("expected admin_setting allowlist source, got %q", resp.HTTPAllowlistSource)
	}
	if resp.SecretCenterStatus.SecretRefCount == 0 || resp.WebhookCallbackSecurity.WebhookSecretTotal == 0 || resp.WebhookCallbackSecurity.CallbackTokenTotal == 0 {
		t.Fatalf("expected secret/webhook/callback summaries: %#v", resp)
	}
	var docsSvc domain.PluginExternalServiceEffectiveConfig
	for _, item := range resp.ExternalServices {
		if item.PluginCode == "docs" {
			docsSvc = item
			break
		}
	}
	if docsSvc.PluginCode == "" {
		t.Fatalf("docs external_service not found: %#v", resp.ExternalServices)
	}
	if docsSvc.AuthType != "bearer" || docsSvc.TokenRef != cfg.TokenRef || docsSvc.TokenNamespace != "external_service" || docsSvc.TokenUsageType != "external_service_token" {
		t.Fatalf("unexpected token metadata: %#v", docsSvc)
	}
	if !docsSvc.AllowlistMatched || docsSvc.AllowlistSource != "admin_setting" || docsSvc.EndpointOrigin != "http://172.17.0.1:18081" {
		t.Fatalf("unexpected allowlist match: %#v", docsSvc)
	}
	if docsSvc.TokenKeyID == "" || docsSvc.TokenMasked == "" || docsSvc.LastHealthStatus == "" {
		t.Fatalf("expected redacted token and health metadata: %#v", docsSvc)
	}
	raw := mustJSONForTest(resp)
	if strings.Contains(resp.DiagnosticText, "diagnostic-plain-token") || strings.Contains(resp.DiagnosticText, "external-token-plain-s3") || strings.Contains(resp.DiagnosticText, "encrypted_value") || strings.Contains(resp.DiagnosticText, "Authorization") {
		t.Fatalf("diagnostic_text leaked sensitive content: %s", resp.DiagnosticText)
	}
	if strings.Contains(raw, "external-token-plain-s3") || strings.Contains(raw, `"encrypted_value":`) || strings.Contains(raw, "token_hash") {
		t.Fatalf("effective config leaked sensitive content: %s", raw)
	}
}
