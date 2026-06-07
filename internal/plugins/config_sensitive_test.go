package plugins

import (
	"encoding/json"
	"testing"
)

func TestEncryptDecryptStringV1_Roundtrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	enc, err := EncryptStringV1(key, "hello")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if enc == "hello" || enc[:7] != "enc:v1:" {
		t.Fatalf("expected enc:v1 prefix, got %q", enc)
	}
	dec, err := DecryptStringV1(key, enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if dec != "hello" {
		t.Fatalf("expected hello, got %q", dec)
	}
}

func TestEncryptDecryptStringV2_Roundtrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	enc, err := EncryptStringV2("key-1", key, "hello")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if enc == "hello" || enc[:7] != "enc:v2:" {
		t.Fatalf("expected enc:v2 prefix, got %q", enc)
	}
	kr := &PluginConfigKeyring{CurrentKeyID: "key-1", Keys: map[string][]byte{"key-1": key}, LegacyV1Supported: true}
	dec, info, err := DecryptStringWithKeyring(kr, enc)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if info.Version != "v2" || info.KeyID != "key-1" {
		t.Fatalf("unexpected cipher info: %+v", info)
	}
	if dec != "hello" {
		t.Fatalf("expected hello, got %q", dec)
	}
}

func TestMergeSensitivePlaceholders_KeepCiphertext(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"token": map[string]any{"type": "string", "x-sensitive": true},
			"flag":  map[string]any{"type": "boolean"},
		},
	}
	stored := map[string]any{"token": "enc:v1:abc:def", "flag": true}
	submitted := map[string]any{"token": "[ENCRYPTED]", "flag": false}
	merged := MergeSensitivePlaceholders(schema, stored, submitted)
	buf, _ := json.Marshal(merged)
	if string(buf) == "" {
		t.Fatalf("unexpected empty")
	}
	m := merged.(map[string]any)
	if m["token"] != "enc:v1:abc:def" {
		t.Fatalf("expected keep ciphertext, got %#v", m["token"])
	}
	if m["flag"] != false {
		t.Fatalf("expected updated flag=false, got %#v", m["flag"])
	}
}
