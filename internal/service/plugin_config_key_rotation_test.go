package service

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/store"
)

func TestPluginConfigKeyRotation_DryRunAndReencrypt_V1ToV2(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	manifest := map[string]any{
		"code":    "cfg_rotate",
		"name":    "Config Rotate",
		"version": "1.0.0",
		"config_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"enabled": map[string]any{"type": "boolean"},
				"token":   map[string]any{"type": "string", "x-sensitive": true},
			},
		},
	}
	raw, _ := json.Marshal(manifest)
	if _, _, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("InstallPluginManifest: %v", err)
	}

	// Prepare keyring env: current key plus an old key used by v1 ciphertext.
	oldKey := make([]byte, 32)
	for i := range oldKey {
		oldKey[i] = byte(i + 1)
	}
	currentKey := make([]byte, 32)
	for i := range currentKey {
		currentKey[i] = byte(i + 100)
	}
	oldB64 := encodeB64(oldKey)
	curB64 := encodeB64(currentKey)

	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "key-new")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", curB64)
	t.Setenv("DEVHUB_PLUGIN_CONFIG_OLD_KEYS", `{"key-old":"`+oldB64+`"}`)

	// Store an enc:v1 value encrypted by old key (no key_id).
	encV1, err := pluginregistry.EncryptStringV1(oldKey, "secret-token")
	if err != nil {
		t.Fatalf("EncryptStringV1: %v", err)
	}
	beforeJSON := `{"enabled":true,"token":"` + encV1 + `"}` // keep as JSON
	if _, err := repo.SetPluginConfig("cfg_rotate", beforeJSON); err != nil {
		t.Fatalf("SetPluginConfig: %v", err)
	}

	dry, err := svc.PluginConfigKeyRotationDryRun(domain.PluginConfigKeyRotationDryRunRequest{Scope: "plugin", PluginCode: "cfg_rotate"})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if dry.Status != "warning" {
		t.Fatalf("expected warning, got %q summary=%+v", dry.Status, dry.Summary)
	}
	if dry.Summary.LegacyV1 == 0 || dry.Summary.NeedsReencrypt == 0 {
		t.Fatalf("expected legacy_v1/needs_reencrypt > 0, got %+v", dry.Summary)
	}

	op := PluginConfigVersionOperator{Type: "admin_user", ID: 1, Name: "admin#1"}
	_, err = svc.PluginConfigKeyRotationReencrypt(domain.PluginConfigKeyRotationReencryptRequest{
		Scope:               "plugin",
		PluginCode:          "cfg_rotate",
		ConfirmCurrentKeyID: "key-new",
	}, op)
	if err != nil {
		t.Fatalf("re-encrypt: %v", err)
	}

	after, _ := repo.PluginByCode("cfg_rotate")
	if !strings.Contains(after.ConfigJSON, "enc:v2:key-new:") {
		t.Fatalf("expected enc:v2 with current key id, got %s", after.ConfigJSON)
	}
}

func encodeB64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
