package store

import "testing"

func TestMemoryPluginConfigValidationAndMerge(t *testing.T) {
	s := NewMemoryStore()

	if _, err := s.SetPluginConfig("qa", "{"); err == nil {
		t.Fatal("expected invalid global plugin config to fail")
	}
	if _, err := s.SetCommunityPluginConfig(1, "qa", "{"); err == nil {
		t.Fatal("expected invalid community plugin config to fail")
	}

	if _, err := s.SetPluginConfig("qa", `{"limit":3,"mode":"global"}`); err != nil {
		t.Fatalf("set global plugin config: %v", err)
	}
	item, err := s.SetCommunityPluginConfig(1, "qa", `{"mode":"community"}`)
	if err != nil {
		t.Fatalf("set community plugin config: %v", err)
	}
	resolved, ok := item.ResolvedConfig.(map[string]any)
	if !ok {
		t.Fatalf("expected resolved config map, got %#v", item.ResolvedConfig)
	}
	effective, ok := resolved["effective"].(map[string]any)
	if !ok {
		t.Fatalf("expected effective config map, got %#v", resolved["effective"])
	}
	if effective["limit"] != float64(3) || effective["mode"] != "community" {
		t.Fatalf("unexpected effective config: %#v", effective)
	}
}
