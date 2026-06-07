package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultAddr        = ":18110"
	defaultHookPath    = "/hooks/content.after_create"
	defaultContentType = "article"
	maxBodyBytes       = 2 * 1024 * 1024
)

type hookPayload struct {
	SchemaVersion string         `json:"schema_version"`
	PluginCode    string         `json:"plugin_code"`
	HookName      string         `json:"hook_name"`
	EventID       string         `json:"event_id"`
	ExecutionID   string         `json:"execution_id"`
	ResourceType  string         `json:"resource_type"`
	ResourceID    int64          `json:"resource_id"`
	CommunityID   int64          `json:"community_id"`
	Actor         hookActor      `json:"actor"`
	OccurredAt    string         `json:"occurred_at"`
	Data          map[string]any `json:"data"`
}

type hookActor struct {
	Type string `json:"type"`
	ID   int64  `json:"id"`
}

type response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func main() {
	addr := getenv("ARTICLE_AUDIT_LOGGER_ADDR", defaultAddr)
	hookPath := getenv("ARTICLE_AUDIT_LOGGER_HOOK_PATH", defaultHookPath)
	onlyContentType := getenv("ARTICLE_AUDIT_LOGGER_CONTENT_TYPE", defaultContentType)
	token := strings.TrimSpace(os.Getenv("ARTICLE_AUDIT_LOGGER_TOKEN"))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, response{OK: true, Message: "ok"})
	})
	mux.HandleFunc(hookPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, response{OK: false, Error: "method_not_allowed"})
			return
		}
		if err := verifyBearer(r, token); err != nil {
			writeJSON(w, http.StatusUnauthorized, response{OK: false, Error: "unauthorized", Message: err.Error()})
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
		_ = r.Body.Close()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, response{OK: false, Error: "read_failed", Message: err.Error()})
			return
		}
		var payload hookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, response{OK: false, Error: "invalid_json", Message: err.Error()})
			return
		}

		contentType := payloadContentType(payload)
		if onlyContentType != "" && contentType != onlyContentType {
			log.Printf("audit_logger skipped content_type=%s event_id=%s resource_id=%d\n", contentType, payload.EventID, payload.ResourceID)
			writeJSON(w, http.StatusAccepted, response{OK: true, Message: "skipped"})
			return
		}

		log.Printf(
			"article_published event_id=%s execution_id=%s topic_id=%d community_id=%d actor_type=%s actor_id=%d title=%q status=%s source_plugin_code=%s occurred_at=%s\n",
			payload.EventID,
			payload.ExecutionID,
			payload.ResourceID,
			payload.CommunityID,
			payload.Actor.Type,
			payload.Actor.ID,
			stringFromData(payload.Data, "title"),
			stringFromData(payload.Data, "status"),
			stringFromData(payload.Data, "metadata.source_plugin_code"),
			firstNonEmpty(payload.OccurredAt, time.Now().UTC().Format(time.RFC3339)),
		)
		writeJSON(w, http.StatusAccepted, response{OK: true, Message: "logged"})
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           logRequest(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("article audit logger listening on %s hook_path=%s content_type=%s auth=%t\n", addr, hookPath, onlyContentType, token != "")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func verifyBearer(r *http.Request, token string) error {
	if token == "" {
		return nil
	}
	got := strings.TrimSpace(r.Header.Get("Authorization"))
	want := "Bearer " + token
	if got != want {
		return errors.New("invalid bearer token")
	}
	return nil
}

func payloadContentType(payload hookPayload) string {
	if v := stringFromData(payload.Data, "content_type"); v != "" {
		return v
	}
	return strings.TrimSpace(payload.ResourceType)
}

func stringFromData(data map[string]any, key string) string {
	if len(data) == 0 {
		return ""
	}
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		var current any = data
		for _, part := range parts {
			m, ok := current.(map[string]any)
			if !ok {
				return ""
			}
			current = m[part]
		}
		return stringify(current)
	}
	return stringify(data[key])
}

func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return fmt.Sprintf("%.0f", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s in %dms\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(started).Milliseconds())
	})
}

func writeJSON(w http.ResponseWriter, status int, payload response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
