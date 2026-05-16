package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginUpgradeOperator struct {
	ID   int64
	Name string
}

func (s *Service) PluginUpgradeImpact(operator PluginUpgradeOperator, code string, targetCompatCheckID int64) (domain.PluginUpgradeImpactResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" || targetCompatCheckID <= 0 {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_invalid_request", "plugin_code 或 target_compat_check_id 不合法").
			WithStatus(400)
	}
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return domain.PluginUpgradeImpactResponse{}, pluginNotFound(code)
	}
	if plugin.IsSystem || strings.TrimSpace(plugin.SourceType) == "builtin" {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_system_forbidden", "内置系统插件不支持远程包升级").
			WithStatus(400).
			WithDetail("plugin_code", code)
	}

	compat, ok := s.repo.PluginPackageCompatCheckByID(targetCompatCheckID)
	if !ok || compat.ID <= 0 {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_package_compat_check_not_found", "兼容性检查记录不存在").
			WithStatus(404).
			WithDetail("compat_check_id", targetCompatCheckID)
	}
	if strings.TrimSpace(compat.PluginCode) != code {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_code_mismatch", "目标包的 plugin_code 与当前插件不一致，不能升级").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithDetail("target_plugin_code", compat.PluginCode)
	}
	if !compat.CanInstall {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_not_installable", "目标包 compat-check 未通过，不能升级").
			WithStatus(400).
			WithDetail("compat_check_id", compat.ID).
			WithDetail("status", compat.Status)
	}

	precheck, ok := s.repo.PluginPackagePrecheckByID(compat.PackagePrecheckID)
	if !ok || precheck.ID <= 0 {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_package_precheck_not_found", "目标包预检记录不存在").
			WithStatus(404).
			WithDetail("package_precheck_id", compat.PackagePrecheckID)
	}
	if precheck.Status != domain.PluginPackagePrecheckStatusPassed {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_precheck_not_passed", "目标包预检未通过，不能升级").
			WithStatus(400).
			WithDetail("package_precheck_id", precheck.ID).
			WithDetail("precheck_status", precheck.Status)
	}
	if strings.TrimSpace(precheck.ManifestJSON) == "" {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_manifest_missing", "目标包预检缺少 manifest_json").
			WithStatus(400).
			WithDetail("package_precheck_id", precheck.ID)
	}
	if err := s.ensureCompatPrecheckSourceUsable(precheck); err != nil {
		return domain.PluginUpgradeImpactResponse{}, err
	}
	if precheck.PackageDownloadID > 0 {
		download, ok := s.repo.PluginPackageDownloadByID(precheck.PackageDownloadID)
		if !ok {
			return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_package_download_not_found", "目标包 download 记录不存在").WithStatus(404)
		}
		if download.Status != domain.PluginPackageDownloadStatusDownloaded {
			return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_download_invalid", "目标包下载状态不允许升级").
				WithStatus(400).
				WithDetail("package_download_id", download.ID).
				WithDetail("download_status", download.Status).
				WithSuggestion("请先完成下载与 sha256 校验。")
		}
		if strings.TrimSpace(download.SHA256Expected) == "" || strings.TrimSpace(download.SHA256Actual) == "" {
			return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_checksum_missing", "目标包缺少 sha256 校验信息，默认不允许进入升级流程").
				WithStatus(400).
				WithDetail("package_download_id", download.ID).
				WithSuggestion("请提供 sha256 并重新下载后再升级。")
		}
		if strings.TrimSpace(download.SHA256Expected) != strings.TrimSpace(download.SHA256Actual) {
			return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_checksum_invalid", "目标包 sha256 不一致，禁止升级").
				WithStatus(400).
				WithDetail("package_download_id", download.ID).
				WithDetail("sha256_expected", download.SHA256Expected).
				WithDetail("sha256_actual", download.SHA256Actual)
		}
	}

	if pluginregistry.CompareVersionStrings(strings.TrimSpace(compat.Version), strings.TrimSpace(plugin.Version)) <= 0 {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_version_same_version", "目标版本必须高于当前版本，本轮不支持降级或重复升级").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithDetail("current_version", plugin.Version).
			WithDetail("target_version", compat.Version)
	}

	manifest, _, err := pluginregistry.DecodePluginManifestJSON([]byte(precheck.ManifestJSON))
	if err != nil {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_manifest_invalid", "目标 manifest_json 不合法").
			WithStatus(400).
			WithDetail("reason", err.Error())
	}
	manifest = pluginregistry.NormalizeManifest(manifest)
	sections, summary := buildPluginManifestDiff(plugin.PluginManifest, manifest)

	// Config compatibility: existing config must pass new schema.
	nextDef := domain.Plugin{PluginManifest: manifest}
	if err := pluginregistry.ValidateConfigJSON(nextDef, plugin.ConfigJSON); err != nil {
		return domain.PluginUpgradeImpactResponse{}, domain.NewPluginError("plugin_upgrade_target_config_incompatible", "现有配置不兼容目标版本 config_schema，禁止升级").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithDetail("current_version", plugin.Version).
			WithDetail("target_version", manifest.Version).
			WithDetail("reason", err.Error()).
			WithSuggestion("请先调整现有配置以满足新 schema，或选择兼容的插件版本。")
	}

	// Migration policy: do not auto-run migrations. If target declares migrations and plugin is enabled, we allow upgrade
	// but will force status to migration_pending (later) unless your policy blocks.
	migDiff := map[string]any{"declared": len(manifest.Migrations), "note": "升级流程本轮不执行 migration；新增迁移将落入 pending。"}

	impact, _ := s.repo.PluginImpact(code)
	resp := domain.PluginUpgradeImpactResponse{
		PluginCode:           code,
		OldVersion:           plugin.Version,
		NewVersion:           manifest.Version,
		TargetCompatCheckID:  targetCompatCheckID,
		PackageDownloadID:    precheck.PackageDownloadID,
		PackagePrecheckID:    precheck.ID,
		Status:               "ok",
		CanUpgrade:           true,
		Impact:               impact,
		ManifestDiffSections: sections,
		ManifestDiffSummary:  summary,
		DependencyDiff:       pluginregistry.DependencyDiff(plugin.Dependencies, manifest.Dependencies),
		ConfigDiff:           map[string]any{"note": "本轮仅做兼容校验；diff 细化留到后续增强", "current_config_valid": true},
		MigrationDiff:        migDiff,
	}

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.upgrade.impact.viewed",
		Target:    fmt.Sprintf("plugins#%s", code),
		Metadata: mustJSON(map[string]any{
			"plugin_code":            code,
			"old_version":            plugin.Version,
			"new_version":            manifest.Version,
			"target_compat_check_id": targetCompatCheckID,
		}),
		CreatedAt: Now(),
	})
	return resp, nil
}

type PluginUpgradeRequest struct {
	TargetCompatCheckID int64  `json:"target_compat_check_id"`
	Reason              string `json:"reason,omitempty"`
}

func (s *Service) UpgradePluginFromCompatCheckAs(operator PluginUpgradeOperator, code string, req PluginUpgradeRequest) (domain.PluginUpgradeTaskResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" || req.TargetCompatCheckID <= 0 {
		return domain.PluginUpgradeTaskResponse{}, domain.NewPluginError("plugin_upgrade_invalid_request", "plugin_code 或 target_compat_check_id 不合法").
			WithStatus(400)
	}

	impact, err := s.PluginUpgradeImpact(operator, code, req.TargetCompatCheckID)
	if err != nil {
		return domain.PluginUpgradeTaskResponse{}, err
	}
	if !impact.CanUpgrade {
		return domain.PluginUpgradeTaskResponse{}, domain.NewPluginError("plugin_upgrade_blocked", "升级前置校验未通过，禁止升级").
			WithStatus(400).
			WithDetail("plugin_code", code).
			WithDetail("target_compat_check_id", req.TargetCompatCheckID)
	}

	plugin, _ := s.repo.PluginByCode(code)
	compat, _ := s.repo.PluginPackageCompatCheckByID(req.TargetCompatCheckID)
	precheck, _ := s.repo.PluginPackagePrecheckByID(compat.PackagePrecheckID)

	startedAt := Now()
	task := domain.PluginUpgradeTask{
		PluginCode:              code,
		OldVersion:              strings.TrimSpace(plugin.Version),
		NewVersion:              strings.TrimSpace(compat.Version),
		NewPackageDownloadID:    precheck.PackageDownloadID,
		NewPackagePrecheckID:    precheck.ID,
		NewPackageCompatCheckID: compat.ID,
		Status:                  domain.PluginUpgradeTaskStatusUpgrading,
		PreviousPluginStatus:    strings.TrimSpace(plugin.Status),
		NewPluginStatus:         pluginregistry.StatusDisabled,
		Reason:                  strings.TrimSpace(req.Reason),
		StartedAt:               startedAt,
		RequestedBy:             operator.ID,
		CreatedAt:               startedAt,
	}

	// Persist initial task record.
	task, _ = s.repo.AppendPluginUpgradeTask(task)

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.upgrade.requested",
		Target:    fmt.Sprintf("plugins#%s", code),
		Metadata: mustJSON(map[string]any{
			"plugin_code":            code,
			"old_version":            plugin.Version,
			"new_version":            compat.Version,
			"target_compat_check_id": compat.ID,
			"upgrade_task_id":        task.ID,
			"reason":                 req.Reason,
		}),
		CreatedAt: Now(),
	})
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.upgrade.started",
		Target:    fmt.Sprintf("plugin-upgrade-tasks#%d", task.ID),
		Metadata:  mustJSON(map[string]any{"plugin_code": code, "old_version": plugin.Version, "new_version": compat.Version}),
		CreatedAt: Now(),
	})

	// Ensure project root is resolvable (used by package path normalization).
	if _, rerr := serviceProjectRoot(); rerr != nil {
		return s.failUpgradeTask(task, operator, "plugin_upgrade_failed", "读取项目根目录失败", rerr)
	}
	// Determine package source path to upgrade from.
	relPkg := strings.TrimSpace(precheck.PackagePath)
	if relPkg == "" {
		return s.failUpgradeTask(task, operator, "plugin_upgrade_target_package_missing", "目标包缺少 package_path，不能升级", errors.New("missing package_path"))
	}
	absPkg, cleanPkg, nerr := pluginregistry.NormalizePluginPackagePath(relPkg)
	if nerr != nil {
		return s.failUpgradeTask(task, operator, nerr.(*domain.APIError).Code, "目标包路径不合法，不能升级", nerr)
	}
	_ = cleanPkg
	manifestRaw, readErr := os.ReadFile(filepath.Join(absPkg, "manifest.json"))
	if readErr != nil {
		return s.failUpgradeTask(task, operator, "plugin_upgrade_target_manifest_missing", "读取目标 manifest.json 失败", readErr)
	}
	manifest, _, derr := pluginregistry.DecodePluginManifestJSON(manifestRaw)
	if derr != nil {
		return s.failUpgradeTask(task, operator, "plugin_upgrade_target_manifest_invalid", "目标 manifest.json 不合法", derr)
	}
	manifest = pluginregistry.NormalizeManifest(manifest)
	if strings.TrimSpace(manifest.Code) != code {
		return s.failUpgradeTask(task, operator, "plugin_upgrade_target_code_mismatch", "目标包 plugin_code 不匹配，禁止升级", errors.New("code mismatch"))
	}

	// Ensure file integrity for local_package: re-run scan+checksum and block if dangerous/mismatch.
	// This reuses v1.5/1.6 package dry-run pipeline.
	dry, dryErr := s.DryRunPluginPackage(relPkg)
	if dryErr != nil {
		return s.failUpgradeTask(task, operator, "plugin_upgrade_target_dry_run_failed", "目标包 dry-run 失败，禁止升级", dryErr)
	}
	if strings.ToLower(dry.Status) == "blocked" || strings.ToLower(dry.RiskReport.Level) == "blocked" {
		return s.failUpgradeTask(task, operator, firstNonEmpty(dry.BlockedCode, "plugin_upgrade_target_blocked"), "目标包风险阻断，禁止升级", errors.New(strings.Join(dry.BlockedReasons, "; ")))
	}
	if dry.Package.ChecksumFound && strings.ToLower(dry.Checksum.Status) != "ok" {
		return s.failUpgradeTask(task, operator, "plugin_package_checksum_mismatch", "目标包 checksum 校验未通过，禁止升级", errors.New(dry.Checksum.Status))
	}

	// Diff snapshot.
	sections, summary := buildPluginManifestDiff(plugin.PluginManifest, manifest)
	task.ManifestDiffJSON = mustJSON(map[string]any{"diff_sections": sections, "diff_summary": summary})
	task.PermissionDiffJSON = mustJSON(map[string]any{"note": "权限 diff 统一来自 manifest diff sections"})
	task.MenuDiffJSON = mustJSON(map[string]any{"note": "菜单 diff 统一来自 manifest diff sections"})
	task.RouteDiffJSON = mustJSON(map[string]any{"note": "路由 diff 统一来自 manifest diff sections"})
	task.HookDiffJSON = mustJSON(map[string]any{"note": "hook diff 统一来自 manifest diff sections"})
	task.ContentTypeDiffJSON = mustJSON(map[string]any{"note": "content_type diff 统一来自 manifest diff sections"})
	task.MigrationDiffJSON = mustJSON(map[string]any{"declared": len(manifest.Migrations), "will_execute": false})
	task.ImpactJSON = mustJSON(impact.Impact)

	// Config compatibility again (must be safe).
	nextDef := domain.Plugin{PluginManifest: manifest}
	if err := pluginregistry.ValidateConfigJSON(nextDef, plugin.ConfigJSON); err != nil {
		return s.failUpgradeTask(task, operator, "plugin_upgrade_target_config_incompatible", "现有配置不兼容目标版本 schema，禁止升级", err)
	}
	task.ConfigDiffJSON = mustJSON(map[string]any{"note": "仅做兼容校验；旧配置可通过新 schema", "current_config_valid": true})

	// Migration compatibility: if current has failed migration, block.
	records, _ := s.repo.PluginMigrations(code)
	for _, it := range records {
		if strings.TrimSpace(it.Status) == "failed" {
			return s.failUpgradeTask(task, operator, "plugin_migration_failed", "存在失败迁移，禁止升级", errors.New(it.MigrationName))
		}
	}

	// Apply upgrade by reusing existing UpgradePluginManifest (DB-only manifest upgrade with operation snapshot).
	// This keeps existing behavior: status remains same or becomes disabled; it does not auto-enable.
	// NOTE: This repo does not persist installed filesystem snapshots; upgrade is governance-level manifest swap.
	res, uerr := s.UpgradePluginManifestWithOperation(PluginOperationOperator{ID: operator.ID, Name: operator.Name}, 0, "", code, manifestRaw)
	if uerr != nil {
		return s.failUpgradeTask(task, operator, "plugin_upgrade_failed", "升级写入失败", uerr)
	}
	_ = res

	// Ensure upgraded plugin is disabled or migration_pending for safety; do not auto-enable.
	updated, _ := s.repo.PluginByCode(code)
	nextStatus := pluginregistry.StatusDisabled
	// If target declares migrations, record pending migrations and force migration_pending.
	if len(manifest.Migrations) > 0 {
		nextStatus = pluginregistry.StatusMigrationPending
	}
	if strings.TrimSpace(updated.Status) == pluginregistry.StatusArchived {
		// Keep archived (soft-uninstalled) as-is.
		nextStatus = updated.Status
	}
	if strings.TrimSpace(updated.Status) == pluginregistry.StatusEnabled {
		// Do not keep enabled after package upgrade.
		_, _ = s.repo.SetPluginStatus(code, nextStatus)
	} else {
		// Ensure status aligns with policy if current is disabled-ish.
		_, _ = s.repo.SetPluginStatus(code, nextStatus)
	}

	task.Status = domain.PluginUpgradeTaskStatusUpgraded
	task.NewPluginStatus = nextStatus
	task.FinishedAt = Now()
	if start, ok := parseTimeLayout(task.StartedAt); ok {
		task.DurationMS = int64(time.Since(start).Milliseconds())
	}
	task, _ = s.repo.SavePluginUpgradeTask(task)

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.upgrade.success",
		Target:    fmt.Sprintf("plugins#%s", code),
		Metadata: mustJSON(map[string]any{
			"plugin_code":       code,
			"old_version":       task.OldVersion,
			"new_version":       task.NewVersion,
			"upgrade_task_id":   task.ID,
			"new_plugin_status": task.NewPluginStatus,
		}),
		CreatedAt: Now(),
	})

	return upgradeTaskResponse(task), nil
}

func (s *Service) failUpgradeTask(task domain.PluginUpgradeTask, operator PluginUpgradeOperator, code, message string, err error) (domain.PluginUpgradeTaskResponse, error) {
	task.Status = domain.PluginUpgradeTaskStatusFailed
	task.FinishedAt = Now()
	if start, ok := parseTimeLayout(task.StartedAt); ok {
		task.DurationMS = int64(time.Since(start).Milliseconds())
	}
	if strings.TrimSpace(task.ErrorsJSON) == "" && err != nil {
		task.ErrorsJSON = mustJSON([]string{code + ": " + err.Error()})
	}
	task, _ = s.repo.SavePluginUpgradeTask(task)
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.upgrade.failed",
		Target:    fmt.Sprintf("plugin-upgrade-tasks#%d", task.ID),
		Metadata: mustJSON(map[string]any{
			"plugin_code":     task.PluginCode,
			"old_version":     task.OldVersion,
			"new_version":     task.NewVersion,
			"upgrade_task_id": task.ID,
			"error_code":      code,
			"error":           firstNonEmpty(errString(err), message),
		}),
		CreatedAt: Now(),
	})
	if err == nil {
		err = errors.New(message)
	}
	if api, ok := err.(*domain.APIError); ok && api != nil {
		return upgradeTaskResponse(task), api
	}
	return upgradeTaskResponse(task), domain.NewPluginError(code, message).
		WithStatus(400).
		WithDetail("plugin_code", task.PluginCode).
		WithDetail("upgrade_task_id", task.ID).
		WithSuggestion("请先查看 upgrade impact / compat-check 结果，并修复 blockers 后重试。")
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}

func upgradeTaskResponse(task domain.PluginUpgradeTask) domain.PluginUpgradeTaskResponse {
	parseStrings := func(raw string) []string {
		out := []string{}
		if strings.TrimSpace(raw) != "" {
			_ = json.Unmarshal([]byte(raw), &out)
		}
		return out
	}
	parseAny := func(raw string) any {
		if strings.TrimSpace(raw) == "" {
			return nil
		}
		var out any
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return nil
		}
		return out
	}
	resp := domain.PluginUpgradeTaskResponse{
		PluginUpgradeTask: task,
		Errors:            parseStrings(task.ErrorsJSON),
		Warnings:          parseStrings(task.WarningsJSON),
		Rollback:          parseStrings(task.RollbackLogJSON),
		ManifestDiff:      parseAny(task.ManifestDiffJSON),
		ConfigDiff:        parseAny(task.ConfigDiffJSON),
		Impact:            parseAny(task.ImpactJSON),
	}
	sort.Strings(resp.Errors)
	sort.Strings(resp.Warnings)
	return resp
}

func (s *Service) ListPluginUpgradeTasks(filter domain.PluginUpgradeTaskFilter) (domain.PluginUpgradeTaskListResponse, error) {
	items, total, err := s.repo.PluginUpgradeTasks(filter)
	if err != nil {
		return domain.PluginUpgradeTaskListResponse{}, err
	}
	page := filter.Page
	pageSize := filter.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	out := make([]domain.PluginUpgradeTaskResponse, 0, len(items))
	for _, it := range items {
		out = append(out, upgradeTaskResponse(it))
	}
	return domain.PluginUpgradeTaskListResponse{
		Items:      out,
		Pagination: domain.Pagination{Page: page, PageSize: pageSize, Total: total},
	}, nil
}

func (s *Service) GetPluginUpgradeTask(id int64) (domain.PluginUpgradeTaskResponse, error) {
	it, ok := s.repo.PluginUpgradeTaskByID(id)
	if !ok || it.ID <= 0 {
		return domain.PluginUpgradeTaskResponse{}, domain.NewPluginError("plugin_upgrade_task_not_found", "升级任务不存在").WithStatus(404)
	}
	return upgradeTaskResponse(it), nil
}

func (s *Service) RetryPluginUpgradeTaskAs(operator PluginUpgradeOperator, id int64) (domain.PluginUpgradeTaskResponse, error) {
	task, ok := s.repo.PluginUpgradeTaskByID(id)
	if !ok || task.ID <= 0 {
		return domain.PluginUpgradeTaskResponse{}, domain.NewPluginError("plugin_upgrade_task_not_found", "升级任务不存在").WithStatus(404)
	}
	if strings.TrimSpace(task.Status) != domain.PluginUpgradeTaskStatusFailed {
		return domain.PluginUpgradeTaskResponse{}, domain.NewPluginError("plugin_upgrade_task_invalid_status", "仅允许重试 failed 的升级任务").
			WithStatus(400).
			WithDetail("status", task.Status)
	}
	// Re-run upgrade using original compat-check id; keep reason.
	return s.UpgradePluginFromCompatCheckAs(operator, task.PluginCode, PluginUpgradeRequest{
		TargetCompatCheckID: task.NewPackageCompatCheckID,
		Reason:              firstNonEmpty(strings.TrimSpace(task.Reason), "retry"),
	})
}

func (s *Service) DeletePluginUpgradeTask(id int64) error {
	task, ok := s.repo.PluginUpgradeTaskByID(id)
	if !ok || task.ID <= 0 {
		return domain.NewPluginError("plugin_upgrade_task_not_found", "升级任务不存在").WithStatus(404)
	}
	task.Status = domain.PluginUpgradeTaskStatusDeleted
	_, _ = s.repo.SavePluginUpgradeTask(task)
	return nil
}
