package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestRemoteIndex_CreateFetchListAndNoPackageDownload(t *testing.T) {
	t.Setenv("DEVHUB_ALLOW_LOCAL_REMOTE_INDEX", "1")
	packageHits := 0
	indexDoc := testRemoteIndexDocument("remote_demo", "devhub-official", "devhub-official-2026")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/packages/remote_demo.zip" {
			packageHits++
			http.Error(w, "package should not be fetched", http.StatusTeapot)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(indexDoc)
	}))
	defer server.Close()

	repo := store.NewMemoryStore()
	svc := New(repo)
	operator := RemoteIndexOperator{ID: 1, Name: "admin"}
	created, err := svc.CreatePluginRemoteIndex(operator, domain.PluginRemoteIndexSource{
		SourceID: "fixture-index",
		Name:     "Fixture Index",
		IndexURL: server.URL + "/index.json",
	})
	if err != nil {
		t.Fatalf("CreatePluginRemoteIndex: %v", err)
	}
	fetched, err := svc.FetchPluginRemoteIndex(operator, created.ID)
	if err != nil {
		t.Fatalf("FetchPluginRemoteIndex: %v", err)
	}
	if fetched.Source.LastFetchStatus != "ok" || fetched.IndexHash == "" {
		t.Fatalf("unexpected fetch result: %#v", fetched.Source)
	}
	if packageHits != 0 {
		t.Fatalf("remote index fetch must not download package_url, hits=%d", packageHits)
	}
	list, err := svc.ListRemoteIndexPlugins(created.ID, "", 1, 20)
	if err != nil {
		t.Fatalf("ListRemoteIndexPlugins: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].Code != "remote_demo" {
		t.Fatalf("unexpected plugin list: %#v", list.Items)
	}
	if list.Items[0].PublisherTrustStatus != "trusted" {
		t.Fatalf("expected trusted publisher, got %q", list.Items[0].PublisherTrustStatus)
	}
	detail, err := svc.GetRemoteIndexPlugin(created.ID, "remote_demo")
	if err != nil {
		t.Fatalf("GetRemoteIndexPlugin: %v", err)
	}
	if !detail.Readonly || !strings.Contains(detail.ReadonlyMessage, "不会下载") {
		t.Fatalf("expected readonly detail: %#v", detail)
	}
}

func TestRemoteIndex_URLSafetyAndDisabled(t *testing.T) {
	svc := New(store.NewMemoryStore())
	operator := RemoteIndexOperator{ID: 1, Name: "admin"}
	cases := []string{"file:///tmp/index.json", "http://127.0.0.1/index.json", "http://localhost/index.json"}
	for _, rawURL := range cases {
		_, err := svc.CreatePluginRemoteIndex(operator, domain.PluginRemoteIndexSource{SourceID: strings.ReplaceAll(rawURL, "/", "_"), Name: "Bad", IndexURL: rawURL})
		if err == nil {
			t.Fatalf("expected unsafe url rejected: %s", rawURL)
		}
	}

	t.Setenv("DEVHUB_ALLOW_LOCAL_REMOTE_INDEX", "1")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(testRemoteIndexDocument("disabled_demo", "devhub-official", "devhub-official-2026"))
	}))
	defer server.Close()
	created, err := svc.CreatePluginRemoteIndex(operator, domain.PluginRemoteIndexSource{SourceID: "disabled-index", Name: "Disabled", IndexURL: server.URL, Status: domain.PluginRemoteIndexStatusDisabled})
	if err != nil {
		t.Fatalf("create disabled: %v", err)
	}
	if _, err := svc.FetchPluginRemoteIndex(operator, created.ID); err == nil {
		t.Fatalf("expected disabled source fetch rejected")
	} else if api, ok := err.(*domain.APIError); !ok || api.Code != "plugin_remote_index_disabled" {
		t.Fatalf("unexpected disabled error: %T %v", err, err)
	}
}

func TestRemoteIndex_InvalidJSONSchemaAndPublisherRisk(t *testing.T) {
	t.Setenv("DEVHUB_ALLOW_LOCAL_REMOTE_INDEX", "1")
	svc := New(store.NewMemoryStore())
	operator := RemoteIndexOperator{ID: 1, Name: "admin"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()
	created, err := svc.CreatePluginRemoteIndex(operator, domain.PluginRemoteIndexSource{SourceID: "bad-json", Name: "Bad JSON", IndexURL: server.URL})
	if err != nil {
		t.Fatalf("create bad-json: %v", err)
	}
	if _, err := svc.FetchPluginRemoteIndex(operator, created.ID); err == nil {
		t.Fatalf("expected invalid json")
	} else if api, ok := err.(*domain.APIError); !ok || api.Code != "plugin_remote_index_invalid_json" {
		t.Fatalf("unexpected invalid json error: %T %v", err, err)
	}

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(testRemoteIndexDocument("unknown_demo", "unknown-publisher", "unknown-key"))
	})
	created.IndexURL = server.URL
	created.Status = domain.PluginRemoteIndexStatusEnabled
	created, _ = repoSaveRemoteIndexForTest(t, svc, created)
	fetched, err := svc.FetchPluginRemoteIndex(operator, created.ID)
	if err != nil {
		t.Fatalf("FetchPluginRemoteIndex unknown: %v", err)
	}
	if fetched.Source.LastFetchStatus != "ok" {
		t.Fatalf("expected ok fetch, got %q", fetched.Source.LastFetchStatus)
	}
	list, err := svc.ListRemoteIndexPlugins(created.ID, "", 1, 20)
	if err != nil {
		t.Fatalf("ListRemoteIndexPlugins unknown: %v", err)
	}
	if len(list.Items) != 1 || list.Items[0].PublisherTrustStatus != "unknown" || list.Items[0].RiskLevel == "low" {
		t.Fatalf("expected unknown publisher warning/high risk: %#v", list.Items)
	}
}

func repoSaveRemoteIndexForTest(t *testing.T, svc *Service, record domain.PluginRemoteIndexSource) (domain.PluginRemoteIndexSource, error) {
	t.Helper()
	return svc.repo.SavePluginRemoteIndex(record)
}

func TestRemoteIndex_ResponseTooLarge(t *testing.T) {
	t.Setenv("DEVHUB_ALLOW_LOCAL_REMOTE_INDEX", "1")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", remoteIndexMaxBodyBytes+2)))
	}))
	defer server.Close()
	svc := New(store.NewMemoryStore())
	created, err := svc.CreatePluginRemoteIndex(RemoteIndexOperator{ID: 1, Name: "admin"}, domain.PluginRemoteIndexSource{SourceID: "too-large", Name: "Too Large", IndexURL: server.URL})
	if err != nil {
		t.Fatalf("create too-large: %v", err)
	}
	if _, err := svc.FetchPluginRemoteIndex(RemoteIndexOperator{ID: 1, Name: "admin"}, created.ID); err == nil {
		t.Fatalf("expected response too large")
	} else if api, ok := err.(*domain.APIError); !ok || api.Code != "plugin_remote_index_response_too_large" {
		t.Fatalf("unexpected too-large error: %T %v", err, err)
	}
}

func testRemoteIndexDocument(code, publisherID, publicKeyID string) domain.PluginRemoteIndexDocument {
	return domain.PluginRemoteIndexDocument{
		SchemaVersion: "1",
		GeneratedAt:   "2026-01-01T00:00:00Z",
		Source: domain.PluginRemoteIndexSourceMeta{
			SourceID: "fixture-index",
			Name:     "Fixture Plugin Index",
			Homepage: "https://example.com/devhub/plugins",
		},
		Plugins: []domain.PluginRemoteIndexPluginDoc{
			{
				Code:          code,
				Name:          "Remote Demo",
				Description:   "只读远程索引测试插件",
				LatestVersion: "1.0.0",
				Versions: []domain.PluginRemoteIndexVersionDoc{
					{
						Version:               "1.0.0",
						MinCoreVersion:        "v1.6.0",
						CompatibleCoreVersion: ">=1.6.0 <1.7.0",
						PackageURL:            "https://example.com/packages/" + code + ".zip",
						PackageSHA256:         strings.Repeat("a", 64),
						ManifestSHA256:        strings.Repeat("b", 64),
						SignatureURL:          "https://example.com/packages/" + code + ".signature.json",
						PublisherID:           publisherID,
						PublicKeyID:           publicKeyID,
						License:               "MIT",
						Tags:                  []string{"notice", "content"},
						CreatedAt:             "2026-01-01T00:00:00Z",
						UpdatedAt:             "2026-01-01T00:00:00Z",
					},
				},
			},
		},
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
