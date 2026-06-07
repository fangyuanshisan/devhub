package plugins

import (
	"testing"
)

func TestDiffPluginConfig_RedactsSensitive(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"token": map[string]any{"type": "string", "x-sensitive": true},
			"name":  map[string]any{"type": "string"},
		},
	}
	before := `{"token":"aaa","name":"a"}`
	after := `{"token":"bbb","name":"b"}`

	_, items := DiffPluginConfig(schema, before, after)
	foundToken := false
	for _, it := range items {
		if it.Path == "token" {
			foundToken = true
			if it.Before != RedactedValue || it.After != RedactedValue {
				t.Fatalf("expected token redacted, got before=%v after=%v", it.Before, it.After)
			}
		}
	}
	if !foundToken {
		t.Fatalf("expected token diff item")
	}
}
