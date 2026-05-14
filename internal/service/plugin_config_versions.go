package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginConfigVersionOperator struct {
	Type string
	ID   int64
	Name string
}

// RecordPluginConfigVersion records a new config version when config changes.
// It stores diff_json in redacted form, and returns (version, created=false) when no change.
func (s *Service) RecordPluginConfigVersion(pluginCode string, scope string, communityID int64, beforeRaw string, afterRaw string, source string, operator PluginConfigVersionOperator) (domain.PluginConfigVersion, bool, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = domain.PluginConfigScopeGlobal
	}
	if pluginCode == "" {
		return domain.PluginConfigVersion{}, false, fmt.Errorf("plugin_code 不能为空")
	}

	beforeHash := pluginregistry.ConfigHash(beforeRaw)
	afterHash := pluginregistry.ConfigHash(afterRaw)
	if beforeHash != "" && beforeHash == afterHash {
		return domain.PluginConfigVersion{}, false, nil
	}
	if beforeHash == "" && strings.TrimSpace(beforeRaw) == "" && afterHash == "" && strings.TrimSpace(afterRaw) == "" {
		return domain.PluginConfigVersion{}, false, nil
	}

	def, _ := s.PluginByCode(pluginCode)
	schema := def.ConfigSchema

	changedKeys, diffItems := pluginregistry.DiffPluginConfig(schema, beforeRaw, afterRaw)
	changedKeysJSON, _ := json.Marshal(changedKeys)
	diffJSON, _ := json.Marshal(diffItems)

	record := domain.PluginConfigVersion{
		PluginCode:      pluginCode,
		Scope:           scope,
		CommunityID:     communityID,
		ConfigJSON:      strings.TrimSpace(afterRaw),
		ConfigHash:      afterHash,
		ChangedKeysJSON: string(changedKeysJSON),
		DiffJSON:        string(diffJSON),
		Source:          firstNonBlank(strings.TrimSpace(source), "manual"),
		OperatorType:    firstNonBlank(strings.TrimSpace(operator.Type), "admin_user"),
		OperatorID:      operator.ID,
		OperatorName:    strings.TrimSpace(operator.Name),
	}
	saved, err := s.repo.AppendPluginConfigVersion(record)
	if err != nil {
		return domain.PluginConfigVersion{}, false, err
	}
	return saved, true, nil
}

func (s *Service) ListPluginConfigVersions(pluginCode string, scope string, communityID int64, page, pageSize int) (domain.PluginConfigVersionListResponse, error) {
	items, total, err := s.repo.PluginConfigVersions(pluginCode, scope, communityID, page, pageSize)
	if err != nil {
		return domain.PluginConfigVersionListResponse{}, err
	}
	out := make([]domain.PluginConfigVersionListItem, 0, len(items))
	for _, it := range items {
		keys := []string{}
		_ = json.Unmarshal([]byte(strings.TrimSpace(it.ChangedKeysJSON)), &keys)
		out = append(out, domain.PluginConfigVersionListItem{
			ID:           it.ID,
			PluginCode:   it.PluginCode,
			Scope:        it.Scope,
			CommunityID:  it.CommunityID,
			VersionNo:    it.VersionNo,
			ChangedKeys:  keys,
			Source:       it.Source,
			OperatorName: it.OperatorName,
			ConfigHash:   it.ConfigHash,
			CreatedAt:    it.CreatedAt,
		})
	}
	return domain.PluginConfigVersionListResponse{
		Items: out,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) GetPluginConfigVersionDetail(pluginCode string, scope string, communityID int64, versionID int64) (domain.PluginConfigVersionDetailResponse, error) {
	v, ok := s.repo.PluginConfigVersionByID(versionID)
	if !ok || v.ID == 0 {
		return domain.PluginConfigVersionDetailResponse{}, domain.NewPluginError("plugin_config_version_not_found", "配置版本不存在").
			WithStatus(404).
			WithDetail("version_id", versionID).
			WithSuggestion("请刷新版本列表后重试。")
	}
	if pluginCode != "" && v.PluginCode != pluginCode {
		return domain.PluginConfigVersionDetailResponse{}, domain.NewPluginError("plugin_config_version_not_found", "配置版本不存在").
			WithStatus(404).
			WithDetail("version_id", versionID).
			WithSuggestion("请检查 plugin_code 与版本是否匹配。")
	}
	if scope != "" && v.Scope != scope {
		return domain.PluginConfigVersionDetailResponse{}, domain.NewPluginError("plugin_config_version_invalid_scope", "配置版本 scope 不匹配").
			WithStatus(400).
			WithDetail("scope", scope).
			WithDetail("actual_scope", v.Scope).
			WithSuggestion("请从正确的 scope 入口打开该版本详情。")
	}
	if v.Scope == domain.PluginConfigScopeCommunity && communityID > 0 && v.CommunityID != communityID {
		return domain.PluginConfigVersionDetailResponse{}, domain.NewPluginError("plugin_config_version_invalid_scope", "配置版本不属于该子站").
			WithStatus(400).
			WithDetail("community_id", communityID).
			WithSuggestion("请从正确的子站配置入口打开该版本详情。")
	}

	def, _ := s.PluginByCode(v.PluginCode)
	schema := def.ConfigSchema
	redacted := pluginregistry.RedactConfig(schema, v.ConfigJSON)

	diff := []domain.PluginConfigDiffItem{}
	_ = json.Unmarshal([]byte(strings.TrimSpace(v.DiffJSON)), &diff)
	keys := []string{}
	_ = json.Unmarshal([]byte(strings.TrimSpace(v.ChangedKeysJSON)), &keys)
	sort.Strings(keys)

	return domain.PluginConfigVersionDetailResponse{
		Version: domain.PluginConfigVersionListItem{
			ID:           v.ID,
			PluginCode:   v.PluginCode,
			Scope:        v.Scope,
			CommunityID:  v.CommunityID,
			VersionNo:    v.VersionNo,
			ChangedKeys:  keys,
			Source:       v.Source,
			OperatorName: v.OperatorName,
			ConfigHash:   v.ConfigHash,
			CreatedAt:    v.CreatedAt,
		},
		ConfigJSON:  redacted,
		Diff:        diff,
		ChangedKeys: keys,
		RawScopeInfo: map[string]any{
			"community_id": v.CommunityID,
		},
	}, nil
}

func (s *Service) PluginConfigRollbackDryRun(pluginCode string, scope string, communityID int64, versionID int64, currentConfigJSON string) (domain.PluginConfigRollbackDryRunResponse, error) {
	// Load target version.
	target, ok := s.repo.PluginConfigVersionByID(versionID)
	if !ok || target.ID == 0 {
		return domain.PluginConfigRollbackDryRunResponse{}, domain.NewPluginError("plugin_config_version_not_found", "配置版本不存在").
			WithStatus(404).
			WithDetail("version_id", versionID).
			WithSuggestion("请刷新版本列表后重试。")
	}
	if pluginCode != "" && target.PluginCode != pluginCode {
		return domain.PluginConfigRollbackDryRunResponse{}, domain.NewPluginError("plugin_config_version_not_found", "配置版本不存在").
			WithStatus(404).
			WithDetail("version_id", versionID).
			WithSuggestion("请检查 plugin_code 与版本是否匹配。")
	}
	if scope != "" && target.Scope != scope {
		return domain.PluginConfigRollbackDryRunResponse{}, domain.NewPluginError("plugin_config_version_invalid_scope", "配置版本 scope 不匹配").
			WithStatus(400).
			WithDetail("scope", scope).
			WithDetail("actual_scope", target.Scope).
			WithSuggestion("请从正确的 scope 入口打开该版本。")
	}
	if target.Scope == domain.PluginConfigScopeCommunity && communityID > 0 && target.CommunityID != communityID {
		return domain.PluginConfigRollbackDryRunResponse{}, domain.NewPluginError("plugin_config_version_invalid_scope", "配置版本不属于该子站").
			WithStatus(400).
			WithDetail("community_id", communityID).
			WithSuggestion("请从正确的子站配置入口打开该版本。")
	}

	def, _ := s.PluginByCode(target.PluginCode)
	schema := def.ConfigSchema

	// Schema validation for target config under current schema.
	status := "ok"
	schemaErrors := []string{}
	if err := pluginregistry.ValidateConfigJSON(def, target.ConfigJSON); err != nil {
		status = "blocked"
		schemaErrors = append(schemaErrors, err.Error())
	}

	changedKeys, diff := pluginregistry.DiffPluginConfig(schema, currentConfigJSON, target.ConfigJSON)
	if status == "ok" && (len(schemaErrors) > 0) {
		status = "blocked"
	}

	// Current version snapshot (best-effort: use hash match to find latest).
	currentHash := pluginregistry.ConfigHash(currentConfigJSON)
	current := domain.PluginConfigVersionListItem{
		ID:          0,
		PluginCode:  target.PluginCode,
		Scope:       target.Scope,
		CommunityID: target.CommunityID,
		VersionNo:   0,
		ConfigHash:  currentHash,
	}

	resp := domain.PluginConfigRollbackDryRunResponse{
		PluginCode:  target.PluginCode,
		Scope:       target.Scope,
		Status:      status,
		BlockedCode: "",
		Suggestion:  "",
		TargetVersion: domain.PluginConfigVersionListItem{
			ID:           target.ID,
			PluginCode:   target.PluginCode,
			Scope:        target.Scope,
			CommunityID:  target.CommunityID,
			VersionNo:    target.VersionNo,
			Source:       target.Source,
			OperatorName: target.OperatorName,
			ConfigHash:   target.ConfigHash,
			CreatedAt:    target.CreatedAt,
		},
		CurrentVersion: current,
		ChangedKeys:    changedKeys,
		Diff:           diff,
		Warnings:       []string{},
		Errors:         []string{},
	}
	resp.SchemaValidation.Valid = len(schemaErrors) == 0
	resp.SchemaValidation.Errors = schemaErrors
	if status == "blocked" {
		resp.Errors = append(resp.Errors, "目标配置未通过当前 config_schema 校验，回滚预览被阻断")
		resp.BlockedCode = "plugin_config_version_schema_invalid"
		resp.Suggestion = "请修复配置（按当前 schema）或选择其他版本后重试。"
	}
	return resp, nil
}
