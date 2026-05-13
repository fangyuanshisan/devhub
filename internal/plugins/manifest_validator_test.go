package plugins

import (
	"strings"
	"testing"
)

func TestValidatePluginManifestJSON(t *testing.T) {
	valid := []byte(`{
		"code": "notes",
		"name": "Notes Plugin",
		"version": "1.0.0",
		"description": "manifest-only test plugin",
		"compatible_core_version": ">=1.3.0",
		"content_types": [
			{
				"type": "note",
				"name": "Note",
				"create_permission": "notes.note.create",
				"edit_permission": "notes.note.edit",
				"delete_permission": "notes.note.delete",
				"allow_comment": true,
				"allow_like": true,
				"allow_favorite": true
			}
		],
		"permissions": [
			{"code": "notes.note.create", "name": "Create Note", "scope": "community"},
			{"code": "notes.note.edit", "name": "Edit Note", "scope": "own"},
			{"code": "notes.note.delete", "name": "Delete Note", "scope": "own"}
		],
		"menus": [
			{"code": "notes.frontend", "title": "Notes", "path": "/notes", "location": "frontend", "permission": "notes.note.create"}
		],
		"routes": [
			{"area": "frontend", "method": "GET", "path": "/notes", "permission": "notes.note.create"}
		],
		"hooks": [
			{"name": "BeforeCreateContent", "mode": "blocking", "failure_policy": "block", "timeout_ms": 1000}
		],
		"config_schema": {
			"type": "object",
			"properties": {
				"enabled": {"type": "boolean", "default": true}
			}
		},
		"migrations": [
			{"migration_version": "1.0.0", "migration_name": "notes_init", "direction": "up", "checksum": "sha256:test"}
		],
		"dependencies": ["qa"],
		"assets": ["assets/icon.svg"]
	}`)
	result := ValidatePluginManifestJSON(valid, Definitions(), "v1.3.4")
	if !result.Valid {
		t.Fatalf("expected valid manifest, errors=%v warnings=%v", result.Errors, result.Warnings)
	}
	if result.NormalizedManifest.Code != "notes" || result.Checksum == "" {
		t.Fatalf("manifest was not normalized/checksummed: %#v", result)
	}
	if result.ImpactSummary.ContentTypesCount != 1 || result.ImpactSummary.MigrationsCount != 1 {
		t.Fatalf("unexpected impact summary: %#v", result.ImpactSummary)
	}

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "code conflict",
			raw:  strings.Replace(string(valid), `"code": "notes"`, `"code": "qa"`, 1),
			want: "code 已存在",
		},
		{
			name: "content type conflict",
			raw:  strings.Replace(string(valid), `"type": "note"`, `"type": "question"`, 1),
			want: "content_type 与插件 qa 冲突",
		},
		{
			name: "invalid hook",
			raw:  strings.Replace(string(valid), `"BeforeCreateContent"`, `"DangerousHook"`, 1),
			want: "Hook 名称不支持",
		},
		{
			name: "invalid schema",
			raw:  strings.Replace(string(valid), `"type": "boolean"`, `"type": "function"`, 1),
			want: "config_schema 不合法",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidatePluginManifestJSON([]byte(tc.raw), Definitions(), "v1.3.4")
			if result.Valid {
				t.Fatalf("expected manifest to fail")
			}
			if !strings.Contains(strings.Join(result.Errors, "\n"), tc.want) {
				t.Fatalf("expected error %q, got %v", tc.want, result.Errors)
			}
		})
	}
}

func TestValidatePluginManifestJSONRequiredDependencyMissingBlocks(t *testing.T) {
	raw := []byte(`{
		"code": "depwarn",
		"name": "Dependency Warning",
		"version": "1.0.0",
		"content_types": [{"type":"depwarn_item","create_permission":"depwarn.item.create"}],
		"permissions": [{"code":"depwarn.item.create","name":"Create","scope":"community"}],
		"dependencies": ["missing_plugin"]
	}`)
	result := ValidatePluginManifestJSON(raw, Definitions(), "v1.3.4")
	if result.Valid {
		t.Fatalf("missing required dependency should invalidate manifest")
	}
	if !strings.Contains(strings.Join(result.Errors, "\n"), "required 依赖插件不存在") {
		t.Fatalf("expected missing dependency error, got %v", result.Errors)
	}
}

func TestValidatePluginManifestJSONOptionalDependencyMissingWarns(t *testing.T) {
	raw := []byte(`{
		"code": "depwarn",
		"name": "Dependency Warning",
		"version": "1.0.0",
		"content_types": [{"type":"depwarn_item","create_permission":"depwarn.item.create"}],
		"permissions": [{"code":"depwarn.item.create","name":"Create","scope":"community"}],
		"dependencies": [{"code":"missing_plugin","required":false,"reason":"optional integration"}]
	}`)
	result := ValidatePluginManifestJSON(raw, Definitions(), "v1.3.4")
	if !result.Valid {
		t.Fatalf("missing optional dependency warning should not invalidate manifest: %v", result.Errors)
	}
	if !strings.Contains(strings.Join(result.Warnings, "\n"), "optional 依赖插件不存在") {
		t.Fatalf("expected missing dependency warning, got %v", result.Warnings)
	}
}
