package service

import (
	"encoding/json"
	"testing"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/store"
)

func TestPluginConfigVersions_RecordAndList_Global(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	manifest := map[string]any{
		"code":    "cfg_demo",
		"name":    "Config Demo",
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

	op := PluginConfigVersionOperator{Type: "admin_user", ID: 1, Name: "admin#1"}

	v1, created, err := svc.RecordPluginConfigVersion("cfg_demo", domain.PluginConfigScopeGlobal, 0, "", `{"enabled":true,"token":"aaa"}`, "manual", op)
	if err != nil || !created {
		t.Fatalf("Record v1: created=%v err=%v", created, err)
	}
	if v1.VersionNo != 1 {
		t.Fatalf("expected version_no=1, got %d", v1.VersionNo)
	}

	v2, created, err := svc.RecordPluginConfigVersion("cfg_demo", domain.PluginConfigScopeGlobal, 0, `{"enabled":true,"token":"aaa"}`, `{"enabled":false,"token":"bbb"}`, "manual", op)
	if err != nil || !created {
		t.Fatalf("Record v2: created=%v err=%v", created, err)
	}
	if v2.VersionNo != 2 {
		t.Fatalf("expected version_no=2, got %d", v2.VersionNo)
	}
	if v2.PreviousVersionID != v1.ID {
		t.Fatalf("expected previous_version_id=%d, got %d", v1.ID, v2.PreviousVersionID)
	}

	list, err := svc.ListPluginConfigVersions("cfg_demo", domain.PluginConfigScopeGlobal, 0, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 items, got %#v", list.Items)
	}
}

func TestPluginConfigRollbackDryRun_BlockedWhenSchemaInvalid(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	manifest := map[string]any{
		"code":    "cfg_demo2",
		"name":    "Config Demo2",
		"version": "1.0.0",
		"config_schema": map[string]any{
			"type":     "object",
			"required": []any{"enabled"},
			"properties": map[string]any{
				"enabled": map[string]any{"type": "boolean"},
			},
		},
	}
	raw, _ := json.Marshal(manifest)
	if _, _, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("InstallPluginManifest: %v", err)
	}
	op := PluginConfigVersionOperator{Type: "admin_user", ID: 1, Name: "admin#1"}

	v1, created, err := svc.RecordPluginConfigVersion("cfg_demo2", domain.PluginConfigScopeGlobal, 0, "", `{}`, "manual", op)
	if err != nil || !created {
		t.Fatalf("Record v1: %v %v", created, err)
	}

	preview, err := svc.PluginConfigRollbackDryRun("cfg_demo2", domain.PluginConfigScopeGlobal, 0, v1.ID, `{"enabled":true}`)
	if err != nil {
		t.Fatalf("DryRun: %v", err)
	}
	if preview.Status != "blocked" {
		t.Fatalf("expected blocked, got %q", preview.Status)
	}
	if preview.SchemaValidation.Valid {
		t.Fatalf("expected schema invalid")
	}
	if preview.BlockedCode != "plugin_config_version_schema_invalid" {
		t.Fatalf("expected blocked_code=plugin_config_version_schema_invalid, got %q", preview.BlockedCode)
	}
}

func TestPluginConfigVersions_NoChangeDoesNotCreateNewVersion(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	manifest := map[string]any{
		"code":    "cfg_no_change",
		"name":    "Config No Change",
		"version": "1.0.0",
		"config_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"enabled": map[string]any{"type": "boolean"},
			},
		},
	}
	raw, _ := json.Marshal(manifest)
	if _, _, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("InstallPluginManifest: %v", err)
	}
	op := PluginConfigVersionOperator{Type: "admin_user", ID: 1, Name: "admin#1"}

	_, created, err := svc.RecordPluginConfigVersion("cfg_no_change", domain.PluginConfigScopeGlobal, 0, "", `{"enabled":true}`, "manual", op)
	if err != nil || !created {
		t.Fatalf("Record v1: created=%v err=%v", created, err)
	}

	_, created, err = svc.RecordPluginConfigVersion("cfg_no_change", domain.PluginConfigScopeGlobal, 0, `{"enabled":true}`, `{"enabled":true}`, "manual", op)
	if err != nil {
		t.Fatalf("Record v2: %v", err)
	}
	if created {
		t.Fatalf("expected created=false when config unchanged")
	}

	list, err := svc.ListPluginConfigVersions("cfg_no_change", domain.PluginConfigScopeGlobal, 0, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 item, got %#v", list.Items)
	}
}

func TestPluginConfigVersions_GlobalAndCommunityVersionNoIndependent(t *testing.T) {
	repo := store.NewMemoryStore()
	svc := New(repo)

	manifest := map[string]any{
		"code":    "cfg_scope",
		"name":    "Config Scope",
		"version": "1.0.0",
		"config_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"enabled": map[string]any{"type": "boolean"},
			},
		},
	}
	raw, _ := json.Marshal(manifest)
	if _, _, err := svc.InstallPluginManifest(raw); err != nil {
		t.Fatalf("InstallPluginManifest: %v", err)
	}
	op := PluginConfigVersionOperator{Type: "admin_user", ID: 1, Name: "admin#1"}

	vg1, created, err := svc.RecordPluginConfigVersion("cfg_scope", domain.PluginConfigScopeGlobal, 0, "", `{"enabled":true}`, "manual", op)
	if err != nil || !created || vg1.VersionNo != 1 {
		t.Fatalf("global v1: created=%v ver=%+v err=%v", created, vg1, err)
	}
	vc1, created, err := svc.RecordPluginConfigVersion("cfg_scope", domain.PluginConfigScopeCommunity, 1, "", `{"enabled":true}`, "manual", op)
	if err != nil || !created || vc1.VersionNo != 1 {
		t.Fatalf("community v1: created=%v ver=%+v err=%v", created, vc1, err)
	}
	vc2, created, err := svc.RecordPluginConfigVersion("cfg_scope", domain.PluginConfigScopeCommunity, 1, `{"enabled":true}`, `{"enabled":false}`, "manual", op)
	if err != nil || !created || vc2.VersionNo != 2 {
		t.Fatalf("community v2: created=%v ver=%+v err=%v", created, vc2, err)
	}
}
