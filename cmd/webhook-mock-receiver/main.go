package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Official mock receiver used for DevHub webhook end-to-end verification.
//
// Safety boundary:
// - This binary does NOT execute any third-party plugin code.
// - This binary only verifies headers/signature and responds with configured status codes.
//
// Usage (example):
//   DEVHUB_WEBHOOK_SECRETS="whsec_xxx=secret1,whsec_yyy=secret2" \
//   DEVHUB_WEBHOOK_TIME_SKEW_SECONDS=300 \
//   DEVHUB_WEBHOOK_MOCK_MODE=ok \
//   DEVHUB_WEBHOOK_MOCK_RETRY_AFTER_SECONDS=60 \
//   go run ./cmd/webhook-mock-receiver
//
// Request override:
// - Header `X-DevHub-Mock-Mode`: ok|500|429|401
// - Query `?mode=ok|500|429|401`
//
// Signature rule (must match DevHub sender):
//   signing_string = timestamp + "." + method + "." + path + "." + body_sha256
//   signature = "v1=" + hex(hmac_sha256(secret, signing_string))

const (
	defaultAddr             = ":18090"
	defaultTimeSkewSeconds  = 300
	defaultMockMode         = "ok"
	defaultRetryAfterSecond = 60
)

type response struct {
	OK      bool              `json:"ok"`
	Message string            `json:"message,omitempty"`
	Error   string            `json:"error,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

func main() {
	addr := getenv("DEVHUB_WEBHOOK_MOCK_ADDR", defaultAddr)
	secrets := parseSecretsEnv(getenv("DEVHUB_WEBHOOK_SECRETS", ""))
	timeSkew := atoi(getenv("DEVHUB_WEBHOOK_TIME_SKEW_SECONDS", ""), defaultTimeSkewSeconds)
	defaultMode := getenv("DEVHUB_WEBHOOK_MOCK_MODE", defaultMockMode)
	retryAfter := atoi(getenv("DEVHUB_WEBHOOK_MOCK_RETRY_AFTER_SECONDS", ""), defaultRetryAfterSecond)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, response{OK: true, Message: "ok"})
	})
	mux.HandleFunc("/hooks/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
		_ = r.Body.Close()

		mode := strings.TrimSpace(r.URL.Query().Get("mode"))
		if mode == "" {
			mode = strings.TrimSpace(r.Header.Get("X-DevHub-Mock-Mode"))
		}
		if mode == "" {
			mode = defaultMode
		}

		if err := verifyRequest(r, body, secrets, timeSkew); err != nil {
			writeJSON(w, http.StatusUnauthorized, response{OK: false, Error: "signature_failed", Message: err.Error()})
			return
		}

		switch mode {
		case "ok", "200":
			writeJSON(w, http.StatusOK, response{OK: true, Message: "accepted"})
		case "500":
			writeJSON(w, http.StatusInternalServerError, response{OK: false, Error: "mock_500", Message: "mock failure"})
		case "429":
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeJSON(w, http.StatusTooManyRequests, response{OK: false, Error: "mock_429", Message: "rate limited"})
		case "401":
			writeJSON(w, http.StatusUnauthorized, response{OK: false, Error: "mock_401", Message: "mock unauthorized"})
		default:
			writeJSON(w, http.StatusBadRequest, response{OK: false, Error: "mock_mode_invalid", Message: "unsupported mode"})
		}
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           logRequest(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("devhub webhook mock receiver listening on %s (secrets=%d)\n", addr, len(secrets))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s in %dms\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start).Milliseconds())
	})
}

func verifyRequest(r *http.Request, body []byte, secrets map[string]string, timeSkewSeconds int) error {
	alg := strings.TrimSpace(r.Header.Get("X-DevHub-Signature-Alg"))
	if alg != "HMAC-SHA256" {
		return fmt.Errorf("unexpected signature alg: %q", alg)
	}

	ts := strings.TrimSpace(r.Header.Get("X-DevHub-Timestamp"))
	if ts == "" {
		return errors.New("missing X-DevHub-Timestamp")
	}
	tsUnix, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || tsUnix <= 0 {
		return errors.New("invalid X-DevHub-Timestamp")
	}
	now := time.Now().Unix()
	if abs64(now-tsUnix) > int64(timeSkewSeconds) {
		return errors.New("timestamp expired")
	}

	bodySHA := strings.TrimSpace(r.Header.Get("X-DevHub-Body-SHA256"))
	if bodySHA == "" {
		return errors.New("missing X-DevHub-Body-SHA256")
	}
	gotBodySHA := sha256Hex(body)
	if !strings.EqualFold(gotBodySHA, bodySHA) {
		return errors.New("body hash mismatch")
	}

	secretRef := strings.TrimSpace(r.Header.Get("X-DevHub-Secret-Ref"))
	if secretRef == "" {
		return errors.New("missing X-DevHub-Secret-Ref")
	}
	secret := secrets[secretRef]
	if secret == "" {
		return errors.New("secret missing")
	}

	sig := strings.TrimSpace(r.Header.Get("X-DevHub-Signature"))
	if !strings.HasPrefix(sig, "v1=") {
		return errors.New("invalid signature format")
	}
	path := r.URL.Path
	method := strings.ToUpper(r.Method)
	signing := ts + "." + method + "." + path + "." + bodySHA
	expected := "v1=" + hmacSHA256Hex([]byte(secret), signing)
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return errors.New("signature mismatch")
	}

	return nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256Hex(secret []byte, signing string) string {
	m := hmac.New(sha256.New, secret)
	_, _ = m.Write([]byte(signing))
	return hex.EncodeToString(m.Sum(nil))
}

func parseSecretsEnv(v string) map[string]string {
	// Format: "ref1=secret1,ref2=secret2"
	out := map[string]string{}
	for _, part := range strings.Split(v, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		ref, sec, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		ref = strings.TrimSpace(ref)
		sec = strings.TrimSpace(sec)
		if ref == "" || sec == "" {
			continue
		}
		out[ref] = sec
	}
	return out
}

func getenv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func atoi(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}
