package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginUninstallOperator struct {
	ID   int64
	Name string
}

func (s *Service) PluginUninstallImpact(code string) (map[string]any, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, domain.NewPluginError("plugin_uninstall_invalid_request", "plugin_code 不能为空").WithStatus(400)
	}
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return nil, pluginNotFound(code)
	}
	impact, _ := s.repo.PluginImpact(code)
	dependents := s.dependentPluginsForUninstall(code)
	warnings := []string{}
	for _, dep := range dependents.RequiredEnabled {
		warnings = append(warnings, "存在启用中的 required 依赖插件："+dep)
	}
	sort.Strings(warnings)
	return map[string]any{
		"plugin_code": plugin.Code,
		"version":     plugin.Version,
		"is_system":   plugin.IsSystem,
		"source_type": plugin.SourceType,
		"status":      plugin.Status,
		"impact":      impact,
		"dependents":  dependents,
		"warnings":    warnings,
	}, nil
}

type dependentPluginsResult struct {
	RequiredEnabled []string `json:"required_enabled,omitempty"`
	OptionalEnabled []string `json:"optional_enabled,omitempty"`
}

func (s *Service) dependentPluginsForUninstall(code string) dependentPluginsResult {
	code = strings.TrimSpace(code)
	if code == "" {
		return dependentPluginsResult{}
	}
	existing := s.repo.Plugins()
	required := []string{}
	optional := []string{}
	for _, p := range existing {
		if strings.TrimSpace(p.Code) == "" || strings.TrimSpace(p.Code) == code {
			continue
		}
		if strings.TrimSpace(p.Status) != pluginregistry.StatusEnabled {
			continue
		}
		for _, dep := range p.Dependencies {
			if strings.TrimSpace(dep.Code) != code {
				continue
			}
			if dep.Required {
				required = append(required, p.Code)
			} else {
				optional = append(optional, p.Code)
			}
		}
	}
	required = uniqueStrings(required)
	optional = uniqueStrings(optional)
	sort.Strings(required)
	sort.Strings(optional)
	return dependentPluginsResult{RequiredEnabled: required, OptionalEnabled: optional}
}

func (s *Service) SoftUninstallPluginAs(operator PluginUninstallOperator, code string, req map[string]any) (domain.PluginUninstallTaskResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.PluginUninstallTaskResponse{}, domain.NewPluginError("plugin_uninstall_invalid_request", "plugin_code 不能为空").WithStatus(400)
	}
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return domain.PluginUninstallTaskResponse{}, pluginNotFound(code)
	}
	if plugin.IsSystem || strings.TrimSpace(plugin.SourceType) == "builtin" {
		return domain.PluginUninstallTaskResponse{}, domain.NewPluginError("plugin_soft_uninstall_system_forbidden", "内置系统插件不允许软卸载").
			WithStatus(400).
			WithDetail("plugin_code", plugin.Code).
			WithSuggestion("如需限制入口请使用 disable；软卸载仅用于远程/包安装插件。")
	}
	if strings.TrimSpace(plugin.Status) == "deleted" {
		return domain.PluginUninstallTaskResponse{}, domain.NewPluginError("plugin_soft_uninstall_invalid_status", "插件状态不允许软卸载").
			WithStatus(400).
			WithDetail("plugin_code", plugin.Code).
			WithDetail("status", plugin.Status)
	}

	version := ""
	reason := ""
	if v, ok := req["version"].(string); ok {
		version = strings.TrimSpace(v)
	}
	if v, ok := req["reason"].(string); ok {
		reason = strings.TrimSpace(v)
	}
	if version != "" && strings.TrimSpace(plugin.Version) != "" && version != strings.TrimSpace(plugin.Version) {
		return domain.PluginUninstallTaskResponse{}, domain.NewPluginError("plugin_soft_uninstall_version_mismatch", "插件版本不匹配，禁止软卸载").
			WithStatus(400).
			WithDetail("plugin_code", plugin.Code).
			WithDetail("expected_version", plugin.Version).
			WithDetail("request_version", version).
			WithSuggestion("请刷新插件详情后重试。")
	}

	startedAt := Now()
	task := domain.PluginUninstallTask{
		PluginCode:     plugin.Code,
		Version:        plugin.Version,
		Status:         domain.PluginUninstallTaskStatusUninstalling,
		UninstallType:  domain.PluginUninstallTypeSoft,
		PreviousStatus: strings.TrimSpace(plugin.Status),
		NewStatus:      pluginregistry.StatusArchived,
		Reason:         reason,
		StartedAt:      startedAt,
		RequestedBy:    operator.ID,
		CreatedAt:      startedAt,
	}

	impact, _ := s.repo.PluginImpact(code)
	task.AffectedContentsCount = impact.ExistingContentsCount
	task.AffectedCommunitiesCount = impact.EnabledCommunitiesCount

	dependents := s.dependentPluginsForUninstall(code)
	task.DependentPluginsJSON = mustJSON(dependents)
	if len(dependents.RequiredEnabled) > 0 {
		task.Status = domain.PluginUninstallTaskStatusFailed
		task.ErrorsJSON = mustJSON([]string{"plugin_soft_uninstall_dependency_blocked: 存在启用中的 required 依赖插件：" + strings.Join(dependents.RequiredEnabled, ",")})
		task.FinishedAt = Now()
		if start, ok := parseTimeLayout(task.StartedAt); ok {
			task.DurationMS = int64(time.Since(start).Milliseconds())
		}
		task, _ = s.repo.AppendPluginUninstallTask(task)
		s.repo.AppendAdminLog(domain.AdminLog{
			Site:      "admin",
			Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
			ActorType: "admin_user",
			ActorID:   operator.ID,
			Action:    "plugin.soft_uninstall.failed",
			Target:    fmt.Sprintf("plugins#%s", plugin.Code),
			Metadata: mustJSON(map[string]any{
				"plugin_code":          plugin.Code,
				"version":              plugin.Version,
				"uninstall_task_id":    task.ID,
				"status":               task.Status,
				"dependent_plugins":    dependents,
				"impact_summary":       impact,
				"error_code":           "plugin_soft_uninstall_dependency_blocked",
				"reason":               reason,
				"previous_status":      task.PreviousStatus,
				"new_status":           task.NewStatus,
				"uninstall_type":       task.UninstallType,
				"affected_contents":    task.AffectedContentsCount,
				"affected_communities": task.AffectedCommunitiesCount,
			}),
			CreatedAt: Now(),
		})
		return uninstallTaskResponse(task), domain.NewPluginError("plugin_soft_uninstall_dependency_blocked", "存在启用中的 required 依赖插件，禁止软卸载").
			WithStatus(400).
			WithDetail("plugin_code", plugin.Code).
			WithDetail("dependent_plugins", dependents.RequiredEnabled).
			WithSuggestion("请先禁用依赖该插件的其他插件后重试。")
	}
	if len(dependents.OptionalEnabled) > 0 {
		task.WarningsJSON = mustJSON([]string{"存在启用中的 optional 依赖插件：" + strings.Join(dependents.OptionalEnabled, ",")})
	}

	task, _ = s.repo.AppendPluginUninstallTask(task)
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.soft_uninstall.requested",
		Target:    fmt.Sprintf("plugins#%s", plugin.Code),
		Metadata: mustJSON(map[string]any{
			"plugin_code":       plugin.Code,
			"version":           plugin.Version,
			"uninstall_task_id": task.ID,
			"reason":            reason,
			"impact_summary":    impact,
		}),
		CreatedAt: Now(),
	})
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.soft_uninstall.started",
		Target:    fmt.Sprintf("plugin-uninstall-tasks#%d", task.ID),
		Metadata:  mustJSON(map[string]any{"plugin_code": plugin.Code, "version": plugin.Version, "previous_status": task.PreviousStatus, "new_status": task.NewStatus}),
		CreatedAt: Now(),
	})

	// Best-effort transactional protection (store handles persistence).
	archived, err := s.repo.SetPluginStatus(plugin.Code, pluginregistry.StatusArchived)
	if err != nil {
		return s.failSoftUninstallTask(task, operator, "plugin_soft_uninstall_failed", "更新插件状态失败", err)
	}

	// Record unregistered snapshot (runtime gating is derived from status).
	task.UnregisteredContentTypesJSON = mustJSON(map[string]any{"count": len(archived.ContentTypes) + len(archived.ContentTypeDefs)})
	task.UnregisteredPermissionsJSON = mustJSON(map[string]any{"count": len(archived.Permissions)})
	task.UnregisteredMenusJSON = mustJSON(map[string]any{"count": len(archived.Menus)})
	task.UnregisteredRoutesJSON = mustJSON(map[string]any{"count": len(archived.Routes)})
	task.UnregisteredHooksJSON = mustJSON(map[string]any{"count": len(archived.Hooks)})
	task.PreservedFilesJSON = mustJSON(map[string]any{"note": "软卸载不删除插件文件"})
	task.PreservedConfigsJSON = mustJSON(map[string]any{"note": "软卸载保留配置，不清空"})
	task.PreservedMigrationsJSON = mustJSON(map[string]any{"note": "软卸载保留迁移记录，不执行 migration down"})

	task.Status = domain.PluginUninstallTaskStatusSoftDone
	task.FinishedAt = Now()
	if start, ok := parseTimeLayout(task.StartedAt); ok {
		task.DurationMS = int64(time.Since(start).Milliseconds())
	}
	task, _ = s.repo.SavePluginUninstallTask(task)

	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.runtime.unregistered",
		Target:    fmt.Sprintf("plugins#%s", archived.Code),
		Metadata: mustJSON(map[string]any{
			"plugin_code":       archived.Code,
			"version":           archived.Version,
			"uninstall_task_id": task.ID,
			"previous_status":   task.PreviousStatus,
			"new_status":        archived.Status,
		}),
		CreatedAt: Now(),
	})
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.soft_uninstall.success",
		Target:    fmt.Sprintf("plugins#%s", archived.Code),
		Metadata: mustJSON(map[string]any{
			"plugin_code":       archived.Code,
			"version":           archived.Version,
			"uninstall_task_id": task.ID,
			"previous_status":   task.PreviousStatus,
			"new_status":        archived.Status,
		}),
		CreatedAt: Now(),
	})

	// Platform governance hook (non-blocking).
	_ = s.DispatchHook(pluginregistry.HookEvent{
		Name: pluginregistry.HookAfterPluginDisabled,
		Mode: pluginregistry.HookNonBlocking,
		Ctx: pluginregistry.HookContext{
			PluginCode: archived.Code,
			ActorType:  pluginregistry.HookActorAdmin,
			ActorID:    operator.ID,
			Actor: domain.ActorContext{
				AdminID: operator.ID,
				IsAdmin: true,
			},
			Metadata: map[string]any{"scope": "global", "status": pluginregistry.StatusArchived},
		},
	})

	return uninstallTaskResponse(task), nil
}

func (s *Service) failSoftUninstallTask(task domain.PluginUninstallTask, operator PluginUninstallOperator, code, message string, err error) (domain.PluginUninstallTaskResponse, error) {
	task.Status = domain.PluginUninstallTaskStatusFailed
	task.FinishedAt = Now()
	task.ErrorsJSON = mustJSON([]string{firstNonEmpty(code, "plugin_soft_uninstall_failed") + ": " + firstNonEmpty(message, err.Error())})
	task, _ = s.repo.SavePluginUninstallTask(task)
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.soft_uninstall.failed",
		Target:    fmt.Sprintf("plugin-uninstall-tasks#%d", task.ID),
		Metadata: mustJSON(map[string]any{
			"plugin_code":       task.PluginCode,
			"version":           task.Version,
			"status":            task.Status,
			"error_code":        code,
			"error":             err.Error(),
			"uninstall_task_id": task.ID,
		}),
		CreatedAt: Now(),
	})
	return domain.PluginUninstallTaskResponse{}, domain.NewPluginError(code, message).
		WithStatus(400).
		WithDetail("plugin_code", task.PluginCode).
		WithDetail("uninstall_task_id", task.ID).
		WithSuggestion("请查看错误详情并修复阻断项后重试。")
}

func uninstallTaskResponse(task domain.PluginUninstallTask) domain.PluginUninstallTaskResponse {
	parseStrings := func(raw string) []string {
		out := []string{}
		_ = json.Unmarshal([]byte(raw), &out)
		return out
	}
	resp := domain.PluginUninstallTaskResponse{
		PluginUninstallTask: task,
		Errors:              parseStrings(task.ErrorsJSON),
		Warnings:            parseStrings(task.WarningsJSON),
		Rollback:            parseStrings(task.RollbackLogJSON),
	}
	return resp
}

func (s *Service) ListPluginUninstallTasks(filter domain.PluginUninstallTaskFilter) (domain.PluginUninstallTaskListResponse, error) {
	items, total, err := s.repo.PluginUninstallTasks(filter)
	if err != nil {
		return domain.PluginUninstallTaskListResponse{}, err
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	out := make([]domain.PluginUninstallTaskResponse, 0, len(items))
	for _, it := range items {
		out = append(out, uninstallTaskResponse(it))
	}
	return domain.PluginUninstallTaskListResponse{
		Items: out,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) GetPluginUninstallTask(id int64) (domain.PluginUninstallTaskResponse, error) {
	it, ok := s.repo.PluginUninstallTaskByID(id)
	if !ok || it.ID <= 0 {
		return domain.PluginUninstallTaskResponse{}, domain.NewPluginError("plugin_uninstall_task_not_found", "软卸载任务不存在").WithStatus(404)
	}
	return uninstallTaskResponse(it), nil
}

func (s *Service) RetryPluginUninstallTaskAs(operator PluginUninstallOperator, id int64) (domain.PluginUninstallTaskResponse, error) {
	task, ok := s.repo.PluginUninstallTaskByID(id)
	if !ok || task.ID <= 0 {
		return domain.PluginUninstallTaskResponse{}, domain.NewPluginError("plugin_uninstall_task_not_found", "软卸载任务不存在").WithStatus(404)
	}
	if task.Status != domain.PluginUninstallTaskStatusFailed && task.Status != domain.PluginUninstallTaskStatusRollbackFail && task.Status != domain.PluginUninstallTaskStatusRolledBack {
		return domain.PluginUninstallTaskResponse{}, domain.NewPluginError("plugin_uninstall_task_invalid_status", "该软卸载任务状态不允许重试").
			WithStatus(400).
			WithDetail("status", task.Status)
	}
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.soft_uninstall.retry",
		Target:    fmt.Sprintf("plugin-uninstall-tasks#%d", task.ID),
		Metadata:  mustJSON(map[string]any{"plugin_code": task.PluginCode, "version": task.Version}),
		CreatedAt: Now(),
	})
	// Use stored version if request does not pass it.
	return s.SoftUninstallPluginAs(operator, task.PluginCode, map[string]any{"version": task.Version, "reason": task.Reason})
}

func (s *Service) DeletePluginUninstallTaskAs(operator PluginUninstallOperator, id int64) (domain.PluginUninstallTaskResponse, error) {
	it, ok := s.repo.PluginUninstallTaskByID(id)
	if !ok || it.ID <= 0 {
		return domain.PluginUninstallTaskResponse{}, domain.NewPluginError("plugin_uninstall_task_not_found", "软卸载任务不存在").WithStatus(404)
	}
	it.Status = domain.PluginUninstallTaskStatusDeleted
	it.UpdatedAt = Now()
	it, _ = s.repo.SavePluginUninstallTask(it)
	s.repo.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     firstNonEmpty(strings.TrimSpace(operator.Name), "system"),
		ActorType: "admin_user",
		ActorID:   operator.ID,
		Action:    "plugin.soft_uninstall.deleted",
		Target:    fmt.Sprintf("plugin-uninstall-tasks#%d", it.ID),
		Metadata:  mustJSON(map[string]any{"plugin_code": it.PluginCode, "version": it.Version}),
		CreatedAt: Now(),
	})
	return uninstallTaskResponse(it), nil
}
