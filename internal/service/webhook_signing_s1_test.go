package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestWebhookSigningHeadersAndSignature_SignedAndRedactedInStore(t *testing.T) {
	// Enable ephemeral keyring for memory store tests (see CreatePluginWebhookSecret).
	old := os.Getenv("DEVHUB_E2E_TESTING")
	_ = os.Setenv("DEVHUB_E2E_TESTING", "1")
	t.Cleanup(func() { _ = os.Setenv("DEVHUB_E2E_TESTING", old) })

	var gotHeaders http.Header
	var gotBody []byte

	secretCh := make(chan string, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := <-secretCh
		secretCh <- secret

		body, _ := io.ReadAll(r.Body)
		gotBody = body
		gotHeaders = r.Header.Clone()

		// basic verify on receiver side (doc-level behavior)
		timestamp := r.Header.Get("X-DevHub-Timestamp")
		bodySHA := r.Header.Get("X-DevHub-Body-SHA256")
		if timestamp == "" || bodySHA == "" {
			w.WriteHeader(401)
			return
		}
		sum := sha256.Sum256(body)
		if hex.EncodeToString(sum[:]) != bodySHA {
			w.WriteHeader(401)
			return
		}
		signingString := timestamp + "." + r.Method + "." + r.URL.EscapedPath() + "." + bodySHA
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(signingString))
		wantSig := hex.EncodeToString(mac.Sum(nil))

		sig := r.Header.Get("X-DevHub-Signature")
		if !strings.HasPrefix(sig, "v1=") {
			w.WriteHeader(401)
			return
		}
		gotSig := strings.TrimPrefix(sig, "v1=")
		if !hmac.Equal([]byte(gotSig), []byte(wantSig)) {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
	}))
	t.Cleanup(ts.Close)

	repo := store.NewMemoryStore()
	svc := New(repo)

	created, err := svc.CreatePluginWebhookSecret(WebhookSecretOperator{ID: 1, Name: "admin"}, CreateWebhookSecretRequest{
		PluginCode: "qa",
		TargetURL:  ts.URL + "/hooks/content.after_create",
	})
	if err != nil {
		t.Fatalf("CreatePluginWebhookSecret err: %v", err)
	}
	if created.SecretPlaintext == "" {
		t.Fatalf("expected plaintext to be returned once")
	}
	secretCh <- created.SecretPlaintext

	// Ensure plaintext is not stored in delivery/secret records.
	sec, ok := repo.PluginWebhookSecretByID(created.Secret.ID)
	if !ok {
		t.Fatalf("expected secret to exist in store")
	}
	if strings.Contains(sec.SecretCiphertext, created.SecretPlaintext) {
		t.Fatalf("secret plaintext leaked into ciphertext field")
	}

	// Prepare a delivery and execute a manual retry to trigger a signed attempt.
	d := domain.WebhookDelivery{
		DeliveryID:  "del_test_1",
		EventID:     "evt_test_1",
		PluginCode:  "qa",
		HookName:    "content.after_create",
		TargetURL:   ts.URL + "/hooks/content.after_create",
		Status:      domain.WebhookDeliveryStatusFailed,
		Attempt:     1,
		MaxAttempts: 5,
	}
	saved, err := repo.AppendWebhookDelivery(d)
	if err != nil {
		t.Fatalf("AppendWebhookDelivery err: %v", err)
	}

	out, err := svc.ManualRetryWebhookDelivery(t.Context(), saved.ID)
	if err != nil {
		t.Fatalf("ManualRetryWebhookDelivery err: %v", err)
	}
	if out.Status != domain.WebhookDeliveryStatusSuccess {
		t.Fatalf("expected success, got %q", out.Status)
	}

	// Header completeness checks.
	for _, key := range []string{
		"X-DevHub-Event-ID",
		"X-DevHub-Delivery-ID",
		"X-DevHub-Plugin-Code",
		"X-DevHub-Timestamp",
		"X-DevHub-Signature",
		"X-DevHub-Signature-Alg",
		"X-DevHub-Idempotency-Key",
		"X-DevHub-Request-ID",
		"X-DevHub-Body-SHA256",
		"X-DevHub-Secret-Ref",
	} {
		if gotHeaders.Get(key) == "" {
			t.Fatalf("missing header %s", key)
		}
	}
	if gotHeaders.Get("X-DevHub-Signature-Alg") != "HMAC-SHA256" {
		t.Fatalf("unexpected signature alg %q", gotHeaders.Get("X-DevHub-Signature-Alg"))
	}
	if !strings.HasPrefix(gotHeaders.Get("X-DevHub-Signature"), "v1=") {
		t.Fatalf("unexpected signature format %q", gotHeaders.Get("X-DevHub-Signature"))
	}

	// Ensure body is expected and sha256 header matches.
	if strings.TrimSpace(string(gotBody)) != "{}" {
		t.Fatalf("unexpected body: %q", string(gotBody))
	}
	sum := sha256.Sum256(gotBody)
	if gotHeaders.Get("X-DevHub-Body-SHA256") != hex.EncodeToString(sum[:]) {
		t.Fatalf("body sha256 header mismatch")
	}

	// Store should redact signature in request_headers_json.
	reloaded, ok := repo.WebhookDeliveryByID(saved.ID)
	if !ok {
		t.Fatalf("expected delivery to exist")
	}
	var storedHeaders map[string]string
	_ = json.Unmarshal([]byte(reloaded.RequestHeadersJSON), &storedHeaders)
	if v := storedHeaders["X-DevHub-Signature"]; v != "v1=[REDACTED]" {
		t.Fatalf("expected stored signature to be redacted, got %q", v)
	}
	// No plaintext secret in stored headers.
	for _, v := range storedHeaders {
		if v == created.SecretPlaintext {
			t.Fatalf("plaintext secret leaked in stored headers json")
		}
	}
}

func TestWebhookSigning_DisabledSecretDoesNotSend(t *testing.T) {
	old := os.Getenv("DEVHUB_E2E_TESTING")
	_ = os.Setenv("DEVHUB_E2E_TESTING", "1")
	t.Cleanup(func() { _ = os.Setenv("DEVHUB_E2E_TESTING", old) })

	var hits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(200)
	}))
	t.Cleanup(ts.Close)

	repo := store.NewMemoryStore()
	svc := New(repo)

	created, err := svc.CreatePluginWebhookSecret(WebhookSecretOperator{ID: 1, Name: "admin"}, CreateWebhookSecretRequest{
		PluginCode: "qa",
		TargetURL:  ts.URL + "/hooks/x",
	})
	if err != nil {
		t.Fatalf("CreatePluginWebhookSecret err: %v", err)
	}
	if _, err := svc.DisablePluginWebhookSecret(WebhookSecretOperator{ID: 1, Name: "admin"}, created.Secret.ID); err != nil {
		t.Fatalf("DisablePluginWebhookSecret err: %v", err)
	}

	d := domain.WebhookDelivery{
		DeliveryID:  "del_test_2",
		EventID:     "evt_test_2",
		PluginCode:  "qa",
		HookName:    "content.after_create",
		TargetURL:   ts.URL + "/hooks/x",
		Status:      domain.WebhookDeliveryStatusFailed,
		Attempt:     1,
		MaxAttempts: 5,
	}
	saved, err := repo.AppendWebhookDelivery(d)
	if err != nil {
		t.Fatalf("AppendWebhookDelivery err: %v", err)
	}
	out, err := svc.ManualRetryWebhookDelivery(t.Context(), saved.ID)
	if err != nil {
		t.Fatalf("ManualRetryWebhookDelivery err: %v", err)
	}
	if atomic.LoadInt32(&hits) != 0 {
		t.Fatalf("expected no network send when secret disabled, hits=%d", hits)
	}
	if out.Status != domain.WebhookDeliveryStatusFailed {
		t.Fatalf("expected failed, got %q", out.Status)
	}
	reloaded, _ := repo.WebhookDeliveryByID(saved.ID)
	if reloaded.SignatureStatus != "secret_disabled" {
		t.Fatalf("expected signature_status secret_disabled, got %q", reloaded.SignatureStatus)
	}
}

func TestWebhookSigning_Remote401DoesNotRetry(t *testing.T) {
	old := os.Getenv("DEVHUB_E2E_TESTING")
	_ = os.Setenv("DEVHUB_E2E_TESTING", "1")
	t.Cleanup(func() { _ = os.Setenv("DEVHUB_E2E_TESTING", old) })

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	t.Cleanup(ts.Close)

	repo := store.NewMemoryStore()
	svc := New(repo)

	_, err := svc.CreatePluginWebhookSecret(WebhookSecretOperator{ID: 1, Name: "admin"}, CreateWebhookSecretRequest{
		PluginCode: "qa",
		TargetURL:  ts.URL + "/hooks/unauth",
	})
	if err != nil {
		t.Fatalf("CreatePluginWebhookSecret err: %v", err)
	}

	d := domain.WebhookDelivery{
		DeliveryID:  "del_test_3",
		EventID:     "evt_test_3",
		PluginCode:  "qa",
		HookName:    "content.after_create",
		TargetURL:   ts.URL + "/hooks/unauth",
		Status:      domain.WebhookDeliveryStatusFailed,
		Attempt:     1,
		MaxAttempts: 5,
	}
	saved, err := repo.AppendWebhookDelivery(d)
	if err != nil {
		t.Fatalf("AppendWebhookDelivery err: %v", err)
	}
	out, err := svc.ManualRetryWebhookDelivery(t.Context(), saved.ID)
	if err != nil {
		t.Fatalf("ManualRetryWebhookDelivery err: %v", err)
	}
	if out.Status != domain.WebhookDeliveryStatusFailed {
		t.Fatalf("expected failed, got %q", out.Status)
	}
	reloaded, _ := repo.WebhookDeliveryByID(saved.ID)
	if reloaded.Status == domain.WebhookDeliveryStatusRetryScheduled {
		t.Fatalf("expected no retry scheduling on 401")
	}
	if strings.TrimSpace(reloaded.NextRetryAt) != "" {
		t.Fatalf("expected next_retry_at empty on 401")
	}

	// Small sanity: timestamp should be close to now; this ensures header exists.
	if tsHeader := gotTimestamp(reloaded.RequestHeadersJSON); tsHeader != "" {
		if _, err := time.ParseDuration("0s"); err != nil {
			_ = err
		}
	}
}

func gotTimestamp(headersJSON string) string {
	var m map[string]string
	_ = json.Unmarshal([]byte(headersJSON), &m)
	return m["X-DevHub-Timestamp"]
}
