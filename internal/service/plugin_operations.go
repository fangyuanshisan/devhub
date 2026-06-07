package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginOperationOperator struct {
	ID   int64
	Name string
}

func newPluginOperationID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("op_%s", strings.ReplaceAll(Now(), " ", "_"))
	}
	return "op_" + hex.EncodeToString(b[:])
}

func (s *Service) ListPluginOperations(filter domain.PluginOperationFilter) (domain.PluginOperationListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	records, total, err := s.repo.PluginOperationSnapshots(filter)
	if err != nil {
		return domain.PluginOperationListResponse{}, err
	}
	items := make([]domain.PluginOperationListItem, 0, len(records))
	for _, it := range records {
		items = append(items, domain.PluginOperationListItem{
			OperationID:   it.OperationID,
			OperationType: it.OperationType,
			PluginCode:    it.PluginCode,
			FromVersion:   it.FromVersion,
			ToVersion:     it.ToVersion,
			PackageSource: it.PackageSource,
			Status:        it.Status,
			CreatedBy:     it.CreatedBy,
			CreatedAt:     it.CreatedAt,
			ErrorCode:     it.ErrorCode,
		})
	}
	return domain.PluginOperationListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     filter.Page,
			PageSize: filter.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) GetPluginOperation(operationID string) (domain.PluginOperationSnapshot, error) {
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return domain.PluginOperationSnapshot{}, domain.NewPluginError("plugin_operation_not_found", "操作记录不存在").
			WithStatus(404).
			WithSuggestion("请检查 operation_id 是否正确。")
	}
	it, ok := s.repo.PluginOperationSnapshotByOperationID(operationID)
	if !ok || it.OperationID == "" {
		return domain.PluginOperationSnapshot{}, domain.NewPluginError("plugin_operation_not_found", "操作记录不存在").
			WithStatus(404).
			WithDetail("operation_id", operationID).
			WithSuggestion("请刷新操作历史页面后重试。")
	}
	return it, nil
}

func (s *Service) RecoverPluginOperationDryRun(operationID string) (domain.PluginOperationRecoverDryRunResponse, error) {
	op, err := s.GetPluginOperation(operationID)
	if err != nil {
		return domain.PluginOperationRecoverDryRunResponse{}, err
	}
	resp := domain.PluginOperationRecoverDryRunResponse{
		Operation: op,
		Status:    "ok",
	}
	if op.Status != domain.PluginOperationStatusFailed {
		resp.Status = "blocked"
		resp.Errors = append(resp.Errors, "仅 failed 状态允许恢复预览")
		resp.AllowedActions = []string{}
		return resp, nil
	}

	detected := []string{}
	plan := []string{}
	if op.OperationType == domain.PluginOperationTypeInstall {
		if p, ok := s.repo.PluginByCode(op.PluginCode); ok && p.Code != "" {
			detected = append(detected, "plugins:plugin_present")
			plan = append(plan, "删除 plugins/community_plugins/plugin_migrations/plugin_config_versions 中与该 plugin_code 相关的残留记录（仅限本次失败安装）")
		}
	} else if op.OperationType == domain.PluginOperationTypeUpgrade {
		if strings.TrimSpace(op.ToVersion) != "" {
			detected = append(detected, "plugin_migrations:possible_new_records")
			plan = append(plan, "清理 plugin_migrations 中 plugin_code+to_version 的迁移残留记录（不执行 migration down）")
		}
	}

	if len(detected) == 0 {
		resp.Status = "warning"
		resp.Summary = "未检测到可清理残留"
		resp.Detected = []string{}
		resp.CleanupPlan = []string{}
		resp.AllowedActions = []string{}
		return resp, nil
	}

	resp.Summary = "检测到失败操作可能残留，可执行 cleanup"
	resp.Detected = detected
	resp.CleanupPlan = plan
	resp.AllowedActions = []string{"cleanup"}
	return resp, nil
}

func (s *Service) CleanupPluginOperation(operator PluginOperationOperator, operationID string) (domain.PluginOperationCleanupResponse, error) {
	op, err := s.GetPluginOperation(operationID)
	if err != nil {
		return domain.PluginOperationCleanupResponse{}, err
	}
	if op.Status != domain.PluginOperationStatusFailed {
		return domain.PluginOperationCleanupResponse{}, domain.NewPluginError("plugin_operation_invalid_status", "当前状态不允许 cleanup").
			WithStatus(400).
			WithDetail("status", op.Status).
			WithSuggestion("仅 failed 状态允许 cleanup。")
	}
	cleaned := []string{}
	warnings := []string{}

	switch op.OperationType {
	case domain.PluginOperationTypeInstall:
		// Only allow cleanup for failed install where the plugin did not exist before.
		if strings.TrimSpace(op.BeforeManifestJSON) != "" || strings.TrimSpace(op.FromVersion) != "" {
			return domain.PluginOperationCleanupResponse{}, domain.NewPluginError("plugin_operation_cleanup_failed", "该操作快照不是可清理的失败安装残留").
				WithStatus(400).
				WithSuggestion("请通过恢复预览确认残留，再联系管理员手动处理。")
		}
		if err := s.repo.DeletePluginByCode(op.PluginCode); err == nil {
			cleaned = append(cleaned, "plugins")
		}
		if n, err := s.repo.DeleteCommunityPluginsByCode(op.PluginCode); err == nil && n > 0 {
			cleaned = append(cleaned, "community_plugins")
		}
		if n, err := s.repo.DeletePluginMigrationsByPlugin(op.PluginCode); err == nil && n > 0 {
			cleaned = append(cleaned, "plugin_migrations")
		}
		if n, err := s.repo.DeletePluginConfigVersionsByPlugin(op.PluginCode); err == nil && n > 0 {
			cleaned = append(cleaned, "plugin_config_versions")
		}
	case domain.PluginOperationTypeUpgrade:
		if strings.TrimSpace(op.ToVersion) == "" {
			warnings = append(warnings, "to_version 缺失，无法定位迁移残留")
			break
		}
		if n, err := s.repo.DeletePluginMigrationsByVersion(op.PluginCode, op.ToVersion); err == nil && n > 0 {
			cleaned = append(cleaned, "plugin_migrations")
		}
	default:
		return domain.PluginOperationCleanupResponse{}, domain.NewPluginError("plugin_operation_cleanup_failed", "不支持的 operation_type").
			WithStatus(400).
			WithDetail("operation_type", op.OperationType)
	}

	op.Status = domain.PluginOperationStatusRecovered
	op.ErrorCode = ""
	op.ErrorMessage = ""
	op.MetadataJSON = scrubJSONStringForSnapshot(op.MetadataJSON)
	op.CreatedBy = operator.ID
	_, _ = s.repo.SavePluginOperationSnapshot(op)

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.operation.cleanup.executed",
		Target:    op.PluginCode,
		Metadata: mustJSON(map[string]any{
			"operation_id":   op.OperationID,
			"operation_type": op.OperationType,
			"plugin_code":    op.PluginCode,
			"from_version":   op.FromVersion,
			"to_version":     op.ToVersion,
			"package_source": op.PackageSource,
			"cleaned":        cleaned,
		}),
		CreatedAt: Now(),
	})

	return domain.PluginOperationCleanupResponse{
		OperationID: op.OperationID,
		Status:      "ok",
		Cleaned:     cleaned,
		Warnings:    warnings,
	}, nil
}

func (s *Service) PluginUpgradeRollbackDryRun(pluginCode string, req domain.PluginUpgradeRollbackDryRunRequest) (domain.PluginUpgradeRollbackDryRunResponse, error) {
	pluginCode = strings.TrimSpace(pluginCode)
	req.OperationID = strings.TrimSpace(req.OperationID)
	if pluginCode == "" || req.OperationID == "" {
		return domain.PluginUpgradeRollbackDryRunResponse{}, errors.New("plugin_code 和 operation_id 不能为空")
	}
	op, err := s.GetPluginOperation(req.OperationID)
	if err != nil {
		return domain.PluginUpgradeRollbackDryRunResponse{}, err
	}
	if op.PluginCode != pluginCode {
		return domain.PluginUpgradeRollbackDryRunResponse{}, domain.NewPluginError("plugin_operation_rollback_blocked", "operation_id 与 plugin_code 不匹配").
			WithStatus(400).
			WithDetail("operation_id", op.OperationID).
			WithDetail("plugin_code", pluginCode)
	}
	if op.OperationType != domain.PluginOperationTypeUpgrade {
		return domain.PluginUpgradeRollbackDryRunResponse{}, domain.NewPluginError("plugin_upgrade_rollback_not_supported", "仅支持对 upgrade 操作做回滚预览").
			WithStatus(400).
			WithDetail("operation_type", op.OperationType)
	}
	p, ok := s.repo.PluginByCode(pluginCode)
	if !ok || p.Code == "" {
		return domain.PluginUpgradeRollbackDryRunResponse{}, pluginNotFound(pluginCode)
	}

	var before domain.PluginManifest
	if strings.TrimSpace(op.BeforeManifestJSON) == "" || !json.Valid([]byte(op.BeforeManifestJSON)) {
		return domain.PluginUpgradeRollbackDryRunResponse{
			PluginCode:  pluginCode,
			OperationID: op.OperationID,
			Status:      "blocked",
			Errors:      []string{"快照缺少 before_manifest_json，无法回滚预览"},
		}, nil
	}
	_ = json.Unmarshal([]byte(op.BeforeManifestJSON), &before)

	current := p.PluginManifest
	diffSections, diffSummary := buildPluginManifestDiff(before, current)

	resp := domain.PluginUpgradeRollbackDryRunResponse{
		PluginCode:     pluginCode,
		OperationID:    op.OperationID,
		FromVersion:    op.FromVersion,
		ToVersion:      op.ToVersion,
		CurrentVersion: p.Version,
		Status:         "ok",
		DiffSections:   diffSections,
		DiffSummary:    diffSummary,
		Warnings: []string{
			"本接口仅提供回滚 dry-run 预览：不写入数据、不支持 migration down、不回滚业务内容。",
		},
	}
	return resp, nil
}

func buildOperationSnapshotPluginJSON(p domain.Plugin) string {
	// Store only a scrubbed subset to avoid leaking ciphertext.
	out := map[string]any{
		"plugin_code": p.Code,
		"name":        p.Name,
		"version":     p.Version,
		"status":      p.Status,
		"source_type": p.SourceType,
	}
	raw, _ := json.Marshal(out)
	return string(raw)
}

func buildOperationSnapshotConfigJSON(p domain.Plugin) string {
	// ConfigJSON may contain ciphertext; always redact for snapshots.
	redacted := pluginregistry.RedactConfig(p.ConfigSchema, p.ConfigJSON)
	raw, _ := json.Marshal(redacted)
	return string(raw)
}
