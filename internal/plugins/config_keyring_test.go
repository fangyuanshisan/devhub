package plugins

import "testing"

func TestLoadPluginConfigKeyring_LegacySingleKey(t *testing.T) {
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY", "devhub-test-key")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEY_ID", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", "")
	t.Setenv("DEVHUB_PLUGIN_CONFIG_OLD_KEYS", "")

	kr, ok, err := LoadPluginConfigKeyring()
	if err != nil || !ok || kr == nil {
		t.Fatalf("expected ok keyring, ok=%v err=%v", ok, err)
	}
	if kr.CurrentKeyID != "legacy" {
		t.Fatalf("expected current=legacy, got %q", kr.CurrentKeyID)
	}
	if _, ok := kr.Keys["legacy"]; !ok {
		t.Fatalf("expected legacy key in keyring")
	}
}

func TestLoadPluginConfigKeyring_JSONForm(t *testing.T) {
	// base64 of 32 bytes "0123456789abcdef0123456789abcdef"
	t.Setenv("DEVHUB_PLUGIN_CONFIG_KEYS", `{"current":"key-1","keys":{"key-1":"MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="}}`)
	kr, ok, err := LoadPluginConfigKeyring()
	if err != nil || !ok || kr == nil {
		t.Fatalf("expected ok keyring, ok=%v err=%v", ok, err)
	}
	if kr.CurrentKeyID != "key-1" {
		t.Fatalf("unexpected current: %q", kr.CurrentKeyID)
	}
	if len(kr.Keys["key-1"]) != 32 {
		t.Fatalf("expected 32 bytes key")
	}
}
