package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginEnableOperator struct {
	ID   int64
	Name string
}

func pluginEnablePrecheckTTL() time.Duration {
	v := strings.TrimSpace(getEnv("DEVHUB_PLUGIN_ENABLE_PRECHECK_TTL_SECONDS", "600"))
	secs, err := time.ParseDuration(v + "s")
	if err != nil || secs < 0 {
		return 10 * time.Minute
	}
	return secs
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(strings.TrimSpace(getenv(key))); v != "" {
		return v
	}
	return fallback
}

// getenv is a seam for tests.
var getenv = os.Getenv

// ParseTimeLayout parses datetime in '2006-01-02 15:04:05' (local time).
// It is used across plugin governance flows for persisted DATETIME strings.
func ParseTimeLayout(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (s *Service) EnablePluginFromEnablePrecheckAs(operator PluginEnableOperator, enablePrecheckID int64) (domain.PluginEnableTaskResponse, error) {
	if enablePrecheckID <= 0 {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_invalid_request", "enable_precheck_id 不合法").WithStatus(400)
	}
	pre, ok := s.repo.PluginEnablePrecheckByID(enablePrecheckID)
	if !ok || pre.ID <= 0 {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_precheck_not_found", "启用前检查记录不存在").WithStatus(404)
	}
	if !pre.CanEnable {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_precheck_blocked", "启用前检查未通过，禁止启用").
			WithStatus(400).
			WithDetail("enable_precheck_id", pre.ID).
			WithDetail("status", pre.Status).
			WithSuggestion("请先修复 enable-precheck 的 blockers 后重试。")
	}
	if pre.Status != domain.PluginEnablePrecheckStatusPassed && pre.Status != domain.PluginEnablePrecheckStatusWarning {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_precheck_blocked", "启用前检查状态不允许启用").
			WithStatus(400).
			WithDetail("enable_precheck_id", pre.ID).
			WithDetail("status", pre.Status)
	}
	plugin, ok := s.repo.PluginByCode(strings.TrimSpace(pre.PluginCode))
	if !ok || plugin.Code == "" {
		return domain.PluginEnableTaskResponse{}, pluginNotFound(pre.PluginCode)
	}
	if strings.TrimSpace(plugin.Status) == pluginregistry.StatusEnabled {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_already_enabled", "插件已启用，无需重复启用").
			WithStatus(400).
			WithDetail("plugin_code", plugin.Code)
	}
	if strings.TrimSpace(plugin.Status) == pluginregistry.StatusArchived {
		return domain.PluginEnableTaskResponse{}, pluginArchived(plugin.Code)
	}

	// TTL check (best-effort, can be disabled by setting TTL=0).
	if ttl := pluginEnablePrecheckTTL(); ttl > 0 {
		if finished, ok := ParseTimeLayout(pre.FinishedAt); ok && time.Since(finished) > ttl {
			return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_precheck_expired", "启用前检查已过期，请重新执行启用前检查后再启用").
				WithStatus(400).
				WithDetail("enable_precheck_id", pre.ID).
				WithDetail("finished_at", pre.FinishedAt).
				WithDetail("ttl_seconds", int(ttl.Seconds())).
				WithSuggestion("请先重新运行 enable-precheck 获取最新 can_enable 结论。")
		}
	}

	startedAt := Now()
	task := domain.PluginEnableTask{
		PluginCode:             plugin.Code,
		Version:                plugin.Version,
		PluginEnablePrecheckID: pre.ID,
		Status:                 domain.PluginEnableTaskStatusEnabling,
		PreviousStatus:         strings.TrimSpace(plugin.Status),
		NewStatus:              pluginregistry.StatusEnabled,
		StartedAt:              startedAt,
		EnabledBy:              operator.ID,
		CreatedAt:              startedAt,
	}
	task, _ = s.repo.AppendPluginEnableTask(task)

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.enable.requested",
		Target:    fmt.Sprintf("plugins#%s", plugin.Code),
		Metadata: mustJSON(map[string]any{
			"plugin_code":        plugin.Code,
			"version":            plugin.Version,
			"enable_precheck_id": pre.ID,
			"enable_task_id":     task.ID,
		}),
		CreatedAt: Now(),
	})
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.enable.started",
		Target:    fmt.Sprintf("plugin-enable-tasks#%d", task.ID),
		Metadata:  mustJSON(map[string]any{"plugin_code": plugin.Code, "version": plugin.Version, "previous_status": task.PreviousStatus, "new_status": task.NewStatus}),
		CreatedAt: Now(),
	})

	// Quick re-check (TOCTOU): reuse existing readiness checks + key conflict checks.
	if err := s.validatePluginEnableReadiness(plugin.Code); err != nil {
		return s.failEnableTask(task, operator, "plugin_enable_failed", "启用前快速校验失败", err)
	}
	// Block pending migrations (policy for v1.7 P0-06).
	migrations, _ := s.pluginMigrationsWithDefinitions(plugin.Code)
	pending := 0
	for _, it := range migrations {
		if strings.TrimSpace(it.Status) == "pending" {
			pending++
		}
	}
	if pending > 0 {
		return s.failEnableTask(task, operator, "plugin_enable_migration_pending", "存在未完成迁移，禁止启用", errors.New("migration pending"))
	}

	// Re-check file integrity if local_package (same as enable-precheck).
	precheck, _, chainErr := s.latestEnablePrecheckChainForPlugin(plugin.Code, plugin.Version)
	if chainErr != nil {
		return s.failEnableTask(task, operator, "plugin_enable_failed", "缺少预检链路，禁止启用", chainErr)
	}
	if strings.TrimSpace(plugin.SourceType) == "local_package" {
		_, ferr := s.enablePrecheckFileIntegrity(plugin, precheck)
		if ferr != nil {
			return s.failEnableTask(task, operator, ferr.Code, "文件完整性复检失败，禁止启用", errors.New(ferr.Message))
		}
	}
	// Content type conflicts against enabled plugins.
	existing := s.repo.Plugins()
	_, ctIssues := s.enablePrecheckContentTypes(plugin, existing)
	if len(ctIssues) > 0 {
		raw, _ := json.Marshal(ctIssues)
		return s.failEnableTask(task, operator, "plugin_enable_conflict_detected", "content_type 冲突或声明不合法，禁止启用", errors.New(string(raw)))
	}

	// Enable plugin status (DB-as-runtime-source) and emit lifecycle hook.
	enabled, err := s.repo.SetPluginStatus(plugin.Code, pluginregistry.StatusEnabled)
	if err != nil {
		return s.failEnableTask(task, operator, "plugin_enable_failed", "更新插件状态失败", err)
	}

	registered := map[string]int{
		"content_types": len(enabled.ContentTypes) + len(enabled.ContentTypeDefs),
		"permissions":   len(enabled.Permissions),
		"menus":         len(enabled.Menus),
		"routes":        len(enabled.Routes),
		"hooks":         len(enabled.Hooks),
	}
	task.RegisteredContentTypesJSON = mustJSON(enabled.ContentTypes)
	task.RegisteredPermissionsJSON = mustJSON(enabled.Permissions)
	task.RegisteredMenusJSON = mustJSON(enabled.Menus)
	task.RegisteredRoutesJSON = mustJSON(enabled.Routes)
	task.RegisteredHooksJSON = mustJSON(enabled.Hooks)

	// Effective config snapshot (governance-only; does not reveal secrets).
	eff := pluginregistry.ResolvePluginConfig(enabled, enabled.ConfigJSON, "")
	task.EffectiveConfigJSON = mustJSON(scrubAnyForSnapshot(eff))

	task.Status = domain.PluginEnableTaskStatusEnabled
	task.FinishedAt = Now()
	if start, ok := ParseTimeLayout(task.StartedAt); ok {
		task.DurationMS = int64(time.Since(start).Milliseconds())
	}
	task, _ = s.repo.SavePluginEnableTask(task)

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.runtime.registered",
		Target:    fmt.Sprintf("plugins#%s", enabled.Code),
		Metadata:  mustJSON(map[string]any{"plugin_code": enabled.Code, "version": enabled.Version, "registered": registered, "enable_task_id": task.ID}),
		CreatedAt: Now(),
	})
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.enable.success",
		Target:    fmt.Sprintf("plugins#%s", enabled.Code),
		Metadata:  mustJSON(map[string]any{"plugin_code": enabled.Code, "version": enabled.Version, "enable_task_id": task.ID, "registered": registered}),
		CreatedAt: Now(),
	})
	_ = s.DispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterPluginEnabled,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode: enabled.Code,
			ActorType:  pluginregistry.HookActorAdmin,
			ActorID:    operator.ID,
			Actor:      ActorContextFromOperator(operator),
			Metadata:   map[string]any{"scope": "global", "enable_task_id": task.ID},
		},
	})

	return enableTaskResponse(task, registered), nil
}

func ActorContextFromOperator(op PluginEnableOperator) domain.ActorContext {
	return domain.ActorContext{IsAdmin: true, AdminID: op.ID}
}

func (s *Service) failEnableTask(task domain.PluginEnableTask, operator PluginEnableOperator, code, message string, err error) (domain.PluginEnableTaskResponse, error) {
	task.Status = domain.PluginEnableTaskStatusFailed
	task.FinishedAt = Now()
	task.ErrorsJSON = mustJSON([]string{firstNonEmpty(code, "plugin_enable_failed") + ": " + firstNonEmpty(message, err.Error())})
	task, _ = s.repo.SavePluginEnableTask(task)
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.enable.failed",
		Target:    fmt.Sprintf("plugin-enable-tasks#%d", task.ID),
		Metadata: mustJSON(map[string]any{
			"plugin_code": task.PluginCode,
			"version":     task.Version,
			"status":      task.Status,
			"error_code":  code,
			"error":       err.Error(),
		}),
		CreatedAt: Now(),
	})
	return domain.PluginEnableTaskResponse{}, domain.NewPluginError(code, message).
		WithStatus(400).
		WithDetail("plugin_code", task.PluginCode).
		WithDetail("enable_task_id", task.ID).
		WithSuggestion("请先重新执行 enable-precheck 和修复阻断项后重试。")
}

func enableTaskResponse(task domain.PluginEnableTask, registered map[string]int) domain.PluginEnableTaskResponse {
	parseStrings := func(raw string) []string {
		out := []string{}
		_ = json.Unmarshal([]byte(raw), &out)
		return out
	}
	resp := domain.PluginEnableTaskResponse{
		PluginEnableTask: task,
		Registered:       registered,
		Errors:           parseStrings(task.ErrorsJSON),
		Warnings:         parseStrings(task.WarningsJSON),
		Rollback:         parseStrings(task.RollbackLogJSON),
	}
	return resp
}

func (s *Service) ListPluginEnableTasks(filter domain.PluginEnableTaskFilter) (domain.PluginEnableTaskListResponse, error) {
	items, total, err := s.repo.PluginEnableTasks(filter)
	if err != nil {
		return domain.PluginEnableTaskListResponse{}, err
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	out := make([]domain.PluginEnableTaskResponse, 0, len(items))
	for _, it := range items {
		out = append(out, enableTaskResponse(it, nil))
	}
	return domain.PluginEnableTaskListResponse{
		Items: out,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) GetPluginEnableTask(id int64) (domain.PluginEnableTaskResponse, error) {
	it, ok := s.repo.PluginEnableTaskByID(id)
	if !ok || it.ID <= 0 {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_task_not_found", "启用任务不存在").WithStatus(404)
	}
	return enableTaskResponse(it, nil), nil
}

func (s *Service) RetryPluginEnableTaskAs(operator PluginEnableOperator, id int64) (domain.PluginEnableTaskResponse, error) {
	task, ok := s.repo.PluginEnableTaskByID(id)
	if !ok || task.ID <= 0 {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_task_not_found", "启用任务不存在").WithStatus(404)
	}
	if task.Status != domain.PluginEnableTaskStatusFailed && task.Status != domain.PluginEnableTaskStatusRollbackFail && task.Status != domain.PluginEnableTaskStatusRolledBack {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_task_invalid_status", "该启用任务状态不允许重试").
			WithStatus(400).
			WithDetail("status", task.Status)
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.enable.retry",
		Target:    fmt.Sprintf("plugin-enable-tasks#%d", task.ID),
		Metadata:  mustJSON(map[string]any{"plugin_code": task.PluginCode, "version": task.Version, "enable_precheck_id": task.PluginEnablePrecheckID}),
		CreatedAt: Now(),
	})
	return s.EnablePluginFromEnablePrecheckAs(operator, task.PluginEnablePrecheckID)
}

func (s *Service) DeletePluginEnableTaskAs(operator PluginEnableOperator, id int64) (domain.PluginEnableTaskResponse, error) {
	it, ok := s.repo.PluginEnableTaskByID(id)
	if !ok || it.ID <= 0 {
		return domain.PluginEnableTaskResponse{}, domain.NewPluginError("plugin_enable_task_not_found", "启用任务不存在").WithStatus(404)
	}
	it.Status = domain.PluginEnableTaskStatusDeleted
	it.UpdatedAt = Now()
	it, _ = s.repo.SavePluginEnableTask(it)
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.enable_task.deleted",
		Target:    fmt.Sprintf("plugin-enable-tasks#%d", it.ID),
		Metadata:  mustJSON(map[string]any{"plugin_code": it.PluginCode, "version": it.Version}),
		CreatedAt: Now(),
	})
	return enableTaskResponse(it, nil), nil
}
