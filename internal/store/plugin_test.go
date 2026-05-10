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

	if _, err := s.SetPluginConfig("qa", `{"allow_anonymous_answer":true,"default_question_status":"publish"}`); err != nil {
		t.Fatalf("set global plugin config: %v", err)
	}
	item, err := s.SetCommunityPluginConfig(1, "qa", `{"allow_anonymous_answer":true,"default_question_status":"publish","require_accept_permission":true}`)
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
	if effective["allow_anonymous_answer"] != true || effective["default_question_status"] != "publish" || effective["require_accept_permission"] != true {
		t.Fatalf("unexpected effective config: %#v", effective)
	}
}
