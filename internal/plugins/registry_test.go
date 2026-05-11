package plugins

import "testing"

func TestManifestContractAndContentTypeMappings(t *testing.T) {
	defs := Definitions()
	if len(defs) == 0 {
		t.Fatal("expected built-in plugin definitions")
	}

	seen := map[string]bool{}
	for _, def := range defs {
		if def.Code == "" || def.PluginCode == "" || def.Name == "" || def.Version == "" {
			t.Fatalf("plugin manifest missing identity fields: %#v", def)
		}
		if !def.IsSystem {
			t.Fatalf("%s should be declared as a system plugin", def.Code)
		}
		if seen[def.Code] {
			t.Fatalf("duplicated plugin code %q", def.Code)
		}
		seen[def.Code] = true
		if len(def.ContentTypes) == 0 || len(def.ContentTypeDefs) == 0 {
			t.Fatalf("%s should declare content types and definitions", def.Code)
		}
		if len(def.Permissions) == 0 || len(def.Menus) == 0 || len(def.Routes) == 0 {
			t.Fatalf("%s should declare permissions, menus and routes", def.Code)
		}
		if def.ConfigSchema == nil {
			t.Fatalf("%s should declare config_schema placeholder", def.Code)
		}
		if def.MinCoreVersion == "" {
			t.Fatalf("%s should declare min_core_version", def.Code)
		}
		for _, ct := range ContentTypeDefinitions(def.Code) {
			if ct.Type == "" || ct.PluginCode != def.Code || ct.CreatePermission == "" {
				t.Fatalf("%s has invalid content type definition: %#v", def.Code, ct)
			}
			if got := PluginCodeForContentType(ct.Type); got != def.Code {
				t.Fatalf("content type %s mapped to %s, want %s", ct.Type, got, def.Code)
			}
			if resolved, ok := ContentTypeDefinitionByType(ct.Type); !ok || resolved.CreatePermission != ct.CreatePermission {
				t.Fatalf("content type %s definition did not roundtrip: %#v ok=%v", ct.Type, resolved, ok)
			}
		}
	}

	if got := NormalizeContentType("doc"); got != "document" {
		t.Fatalf("doc alias normalized to %q", got)
	}
	if got := NormalizeContentType("wiki"); got != "wiki_page" {
		t.Fatalf("wiki alias normalized to %q", got)
	}
	if got := PluginCodeForContentType("question"); got != "qa" {
		t.Fatalf("question mapped to %q", got)
	}
	if got := PluginCodeForContentType("document"); got != "docs" {
		t.Fatalf("document mapped to %q", got)
	}
	if got := PluginCodeForContentType("wiki_page"); got != "wiki" {
		t.Fatalf("wiki_page mapped to %q", got)
	}
	if got := PluginCodeForContentType("project"); got != "projects" {
		t.Fatalf("project mapped to %q", got)
	}
	if got := PluginCodeForContentType("job"); got != "jobs" {
		t.Fatalf("job mapped to %q", got)
	}
	if got := PluginCodeForContentType("ai_work"); got != "ai_works" {
		t.Fatalf("ai_work mapped to %q", got)
	}
}

func TestResolvePluginConfigPrecedence(t *testing.T) {
	def, ok := DefinitionByCode("qa")
	if !ok {
		t.Fatal("qa definition missing")
	}
	resolved := ResolvePluginConfig(def, `{"allow_anonymous_answer":true,"default_question_status":"review"}`, `{"default_question_status":"publish"}`)
	effective, ok := resolved["effective"].(map[string]any)
	if !ok {
		t.Fatalf("expected effective map, got %#v", resolved["effective"])
	}
	if effective["allow_anonymous_answer"] != true ||
		effective["default_question_status"] != "publish" ||
		effective["require_accept_permission"] != true {
		t.Fatalf("unexpected merged config: %#v", effective)
	}
}

func TestValidateConfigJSONSchemaRules(t *testing.T) {
	def, ok := DefinitionByCode("docs")
	if !ok {
		t.Fatal("docs definition missing")
	}
	cases := []struct {
		name string
		raw  string
	}{
		{name: "required", raw: `{}`},
		{name: "boolean type", raw: `{"allow_public_spaces":"yes"}`},
		{name: "integer type", raw: `{"allow_public_spaces":true,"max_tree_depth":1.5}`},
		{name: "integer min", raw: `{"allow_public_spaces":true,"max_tree_depth":0}`},
		{name: "unknown field", raw: `{"allow_public_spaces":true,"unknown":true}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateConfigJSON(def, tc.raw); err == nil {
				t.Fatalf("expected %s config to fail", tc.name)
			}
		})
	}
	if err := ValidateConfigJSON(def, `{"allow_public_spaces":true,"max_tree_depth":10}`); err != nil {
		t.Fatalf("expected valid docs config: %v", err)
	}

	qaDef, ok := DefinitionByCode("qa")
	if !ok {
		t.Fatal("qa definition missing")
	}
	if err := ValidateConfigJSON(qaDef, `{"allow_anonymous_answer":true,"default_question_status":"invalid"}`); err == nil {
		t.Fatal("expected invalid enum to fail")
	}
}
