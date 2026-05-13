package service

import (
	"strings"

	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginReadinessCheck struct {
	Key            string         `json:"key"`
	Title          string         `json:"title"`
	Status         string         `json:"status"` // pass|warning|blocked
	Reason         string         `json:"reason,omitempty"`
	Suggestion     string         `json:"suggestion,omitempty"`
	Code           string         `json:"code,omitempty"` // plugin error code
	PluginCode     string         `json:"plugin_code,omitempty"`
	DependencyCode string         `json:"dependency_code,omitempty"`
	PermissionCode string         `json:"permission_code,omitempty"`
	Details        map[string]any `json:"details,omitempty"`
}

type PluginReadinessResult struct {
	PluginCode string                 `json:"plugin_code"`
	Action     string                 `json:"action"`
	Status     string                 `json:"status"` // pass|warning|blocked
	Checks     []PluginReadinessCheck `json:"checks"`
}

func readinessWorse(a, b string) string {
	order := map[string]int{"pass": 0, "warning": 1, "blocked": 2}
	if order[b] > order[a] {
		return b
	}
	return a
}

func (s *Service) PluginReadiness(code string, action string) (PluginReadinessResult, error) {
	code = strings.TrimSpace(code)
	action = strings.TrimSpace(action)
	if action == "" {
		action = "enable"
	}
	plugin, ok := s.repo.PluginByCode(code)
	if !ok || plugin.Code == "" {
		return PluginReadinessResult{}, pluginNotFound(code)
	}
	out := PluginReadinessResult{PluginCode: code, Action: action, Status: "pass", Checks: []PluginReadinessCheck{}}

	// Status check
	statusCheck := PluginReadinessCheck{Key: "status", Title: "插件状态", Status: "pass", PluginCode: code}
	switch strings.TrimSpace(plugin.Status) {
	case pluginregistry.StatusArchived:
		statusCheck.Status = "blocked"
		statusCheck.Reason = "插件已归档"
		statusCheck.Suggestion = "请先恢复插件后重试"
		statusCheck.Code = PluginErrArchived
	case pluginregistry.StatusDiscovered:
		statusCheck.Status = "blocked"
		statusCheck.Reason = "插件尚未安装"
		statusCheck.Suggestion = "请先安装插件后重试"
		statusCheck.Code = PluginErrNotInstalled
	case pluginregistry.StatusMigrationFailed:
		statusCheck.Status = "blocked"
		statusCheck.Reason = "插件存在失败迁移"
		statusCheck.Suggestion = "请先处理失败迁移后重试"
		statusCheck.Code = PluginErrMigrationFailed
	}
	out.Status = readinessWorse(out.Status, statusCheck.Status)
	out.Checks = append(out.Checks, statusCheck)

	// Core compatibility
	compat := pluginregistry.CheckPluginVersionCompatibility(plugin.PluginManifest, currentCoreVersion())
	compatCheck := PluginReadinessCheck{
		Key:        "core_version",
		Title:      "Core 版本兼容",
		Status:     "pass",
		PluginCode: code,
		Details: map[string]any{
			"core_version":            compat.CoreVersion,
			"min_core_version":        compat.MinCoreVersion,
			"compatible_core_version": compat.CompatibleCoreVersion,
			"messages":                compat.Messages,
		},
	}
	if compat.Status == pluginregistry.CompatibilityWarning {
		compatCheck.Status = "warning"
		compatCheck.Reason = strings.Join(compat.Messages, "；")
		compatCheck.Suggestion = "建议补充 min_core_version 或 compatible_core_version"
		compatCheck.Code = PluginErrCoreVersionIncompat
	} else if compat.Status == pluginregistry.CompatibilityIncompatible {
		compatCheck.Status = "blocked"
		compatCheck.Reason = strings.Join(compat.Messages, "；")
		compatCheck.Suggestion = "请升级 Core 或选择兼容版本的插件"
		compatCheck.Code = PluginErrCoreVersionIncompat
	}
	out.Status = readinessWorse(out.Status, compatCheck.Status)
	out.Checks = append(out.Checks, compatCheck)

	// Config schema check
	if err := pluginregistry.ValidateConfigJSON(plugin, plugin.ConfigJSON); err != nil {
		msg := strings.TrimSpace(err.Error())
		path := ""
		if strings.HasPrefix(msg, "$") {
			parts := strings.Fields(msg)
			if len(parts) > 0 {
				path = parts[0]
			}
		}
		check := PluginReadinessCheck{
			Key:        "config_schema",
			Title:      "config_schema 校验",
			Status:     "blocked",
			Reason:     msg,
			Suggestion: "请修复配置后再重试",
			Code:       PluginErrConfigSchemaInvalid,
			PluginCode: code,
			Details:    map[string]any{"path": path},
		}
		out.Status = readinessWorse(out.Status, check.Status)
		out.Checks = append(out.Checks, check)
	} else {
		out.Checks = append(out.Checks, PluginReadinessCheck{Key: "config_schema", Title: "config_schema 校验", Status: "pass", PluginCode: code})
	}

	// Migrations check
	migrations, migErr := s.pluginMigrationsWithDefinitions(code)
	migCheck := PluginReadinessCheck{Key: "migrations", Title: "迁移状态", Status: "pass", PluginCode: code}
	if migErr != nil {
		migCheck.Status = "warning"
		migCheck.Reason = "迁移状态不可用"
		migCheck.Suggestion = "请稍后重试或检查存储状态"
		migCheck.Details = map[string]any{"error": migErr.Error()}
		out.Status = readinessWorse(out.Status, migCheck.Status)
		out.Checks = append(out.Checks, migCheck)
	} else {
		failed := ""
		pending := 0
		for _, item := range migrations {
			if strings.TrimSpace(item.Status) == "failed" && failed == "" {
				failed = item.MigrationName
			}
			if strings.TrimSpace(item.Status) == "pending" || strings.TrimSpace(item.Status) == "running" {
				pending++
			}
		}
		if failed != "" {
			migCheck.Status = "blocked"
			migCheck.Reason = "存在失败迁移：" + failed
			migCheck.Suggestion = "请先重试或处理迁移错误"
			migCheck.Code = PluginErrMigrationFailed
			migCheck.Details = map[string]any{"failed_migration": failed, "pending_count": pending}
		} else if pending > 0 {
			migCheck.Status = "warning"
			migCheck.Reason = "存在待处理迁移"
			migCheck.Suggestion = "建议先执行 pending migration 再启用"
			migCheck.Details = map[string]any{"pending_count": pending}
		} else {
			migCheck.Details = map[string]any{"pending_count": 0}
		}
		out.Status = readinessWorse(out.Status, migCheck.Status)
		out.Checks = append(out.Checks, migCheck)
	}

	// Dependency checks
	checks, summary := pluginregistry.ResolvePluginDependencies(plugin.PluginManifest, s.repo.Plugins())
	for _, dep := range checks {
		item := PluginReadinessCheck{
			Key:            "dependency." + dep.Code,
			Title:          "依赖插件 " + dep.Code,
			Status:         "pass",
			PluginCode:     code,
			DependencyCode: dep.Code,
			Details: map[string]any{
				"required":         dep.Required,
				"required_version": dep.Version,
				"current_version":  dep.CurrentVersion,
				"current_status":   dep.CurrentStatus,
				"dependency_chain": dep.Chain,
			},
		}
		if dep.Satisfied {
			out.Checks = append(out.Checks, item)
			continue
		}
		blocking := dep.Required || dep.Status == pluginregistry.DependencySelfDependency || dep.Status == pluginregistry.DependencyCircularDependency
		if !blocking {
			item.Status = "warning"
		} else {
			item.Status = "blocked"
		}
		item.Reason = firstNonBlank(dep.Message, "依赖未满足")
		apiErr := dependencyAPIError(code, dep)
		item.Code = apiErr.Code
		item.Suggestion = apiErr.Suggestion
		out.Status = readinessWorse(out.Status, item.Status)
		out.Checks = append(out.Checks, item)
	}

	// Summarize dependency stats as a check for quick scan.
	depSummary := PluginReadinessCheck{
		Key:        "dependency_summary",
		Title:      "依赖总览",
		Status:     "pass",
		PluginCode: code,
		Details: map[string]any{
			"summary":  summary,
			"total":    summary.Total,
			"blocking": summary.Blocking,
			"warnings": summary.Warnings,
		},
	}
	if summary.Blocking > 0 {
		depSummary.Status = "blocked"
		depSummary.Reason = "存在未满足的强依赖"
		depSummary.Suggestion = "请先修复依赖插件状态或版本后重试"
		depSummary.Code = PluginErrDependencyMissing
	} else if summary.Warnings > 0 {
		depSummary.Status = "warning"
		depSummary.Reason = "存在可选依赖缺失"
		depSummary.Suggestion = "可选依赖缺失不阻断，但建议补齐依赖"
		depSummary.Code = PluginErrDependencyMissing
	}
	out.Status = readinessWorse(out.Status, depSummary.Status)
	out.Checks = append(out.Checks, depSummary)

	return out, nil
}
