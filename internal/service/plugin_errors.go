package service

import (
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const (
	PluginErrNotFound                = "plugin_not_found"
	PluginErrNotInstalled            = "plugin_not_installed"
	PluginErrArchived                = "plugin_archived"
	PluginErrDisabled                = "plugin_disabled"
	PluginErrConfigInvalid           = "plugin_config_invalid"
	PluginErrMigrationFailed         = "plugin_migration_failed"
	PluginErrDependencyMissing       = "plugin_dependency_missing"
	PluginErrDependencyDisabled      = "plugin_dependency_disabled"
	PluginErrDependencyArchived      = "plugin_dependency_archived"
	PluginErrDependencyVersion       = "plugin_dependency_version_mismatch"
	PluginErrDependencyCycle         = "plugin_dependency_cycle"
	PluginErrCoreVersionIncompat     = "plugin_core_version_incompatible"
	PluginErrPermissionDenied        = "plugin_permission_denied"
	PluginErrConfigPermissionDenied  = "plugin_config_permission_denied"
	PluginErrContentPermissionDenied = "plugin_content_permission_denied"
	PluginErrHookBlocked             = "plugin_hook_blocked"
	PluginErrHookFailed              = "plugin_hook_failed"
	PluginErrConfigSchemaInvalid     = "plugin_config_schema_invalid"
	PluginErrManifestInvalid         = "plugin_manifest_invalid"
)

func pluginNotFound(code string) *domain.APIError {
	return domain.NewPluginError(PluginErrNotFound, "插件不存在").
		WithStatus(404).
		WithDetail("plugin_code", strings.TrimSpace(code)).
		WithSuggestion("请检查插件编码是否正确，或先执行 manifest 安装。")
}

func pluginNotInstalled(code string) *domain.APIError {
	return domain.NewPluginError(PluginErrNotInstalled, "插件尚未安装，无法执行该操作").
		WithStatus(400).
		WithDetail("plugin_code", strings.TrimSpace(code)).
		WithSuggestion("请先在“安装”向导中安装该插件。")
}

func pluginArchived(code string) *domain.APIError {
	return domain.NewPluginError(PluginErrArchived, "插件已归档，无法执行该操作").
		WithStatus(400).
		WithDetail("plugin_code", strings.TrimSpace(code)).
		WithSuggestion("请先恢复插件后再重试。")
}

func pluginDisabled(code string) *domain.APIError {
	return domain.NewPluginError(PluginErrDisabled, "插件未启用，无法执行该操作").
		WithStatus(400).
		WithDetail("plugin_code", strings.TrimSpace(code)).
		WithSuggestion("请先启用该插件后再重试。")
}

func pluginConfigInvalid(code string, reason string) *domain.APIError {
	err := domain.NewPluginError(PluginErrConfigInvalid, "插件配置无效，无法执行该操作").
		WithStatus(400).
		WithDetail("plugin_code", strings.TrimSpace(code)).
		WithSuggestion("请先修复配置后再重试。")
	if strings.TrimSpace(reason) != "" {
		err.WithDetail("reason", strings.TrimSpace(reason))
	}
	return err
}

func pluginConfigSchemaInvalid(code string, path string, reason string) *domain.APIError {
	err := domain.NewPluginError(PluginErrConfigSchemaInvalid, "插件配置未通过 config_schema 校验").
		WithStatus(400).
		WithDetail("plugin_code", strings.TrimSpace(code)).
		WithSuggestion("请根据字段路径修复配置后再保存。")
	if strings.TrimSpace(path) != "" {
		err.WithDetail("path", path)
	}
	if strings.TrimSpace(reason) != "" {
		err.WithDetail("reason", strings.TrimSpace(reason))
	}
	return err
}

func pluginMigrationFailed(code string, migrationName string) *domain.APIError {
	err := domain.NewPluginError(PluginErrMigrationFailed, "插件迁移失败，无法执行该操作").
		WithStatus(400).
		WithDetail("plugin_code", strings.TrimSpace(code)).
		WithSuggestion("请先在“迁移”Tab 重试或处理失败原因。")
	if strings.TrimSpace(migrationName) != "" {
		err.WithDetail("migration_name", strings.TrimSpace(migrationName))
	}
	return err
}

func dependencyAPIError(owner string, check domain.PluginDependencyCheck) *domain.APIError {
	base := domain.NewPluginError(PluginErrDependencyMissing, "插件依赖未满足，无法执行该操作").
		WithStatus(400).
		WithDetail("plugin_code", strings.TrimSpace(owner)).
		WithDetail("dependency_code", strings.TrimSpace(check.Code)).
		WithDetail("required", check.Required).
		WithDetail("required_version", strings.TrimSpace(check.Version)).
		WithDetail("current_version", strings.TrimSpace(check.CurrentVersion)).
		WithDetail("current_status", strings.TrimSpace(check.CurrentStatus)).
		WithDetail("dependency_status", strings.TrimSpace(check.Status)).
		WithDetail("dependency_chain", check.Chain)

	switch check.Status {
	case pluginregistry.DependencyMissing, pluginregistry.DependencyOptionalMissing:
		base.Code = PluginErrDependencyMissing
		base.Message = "插件依赖缺失，无法执行该操作"
		base.WithSuggestion("请先安装并启用依赖插件后重试。")
	case pluginregistry.DependencyDisabled:
		base.Code = PluginErrDependencyDisabled
		base.Message = "依赖插件未启用，无法执行该操作"
		base.WithSuggestion("请先启用依赖插件后重试。")
	case pluginregistry.DependencyArchived:
		base.Code = PluginErrDependencyArchived
		base.Message = "依赖插件已归档，无法执行该操作"
		base.WithSuggestion("请先恢复依赖插件后重试。")
	case pluginregistry.DependencyVersionMismatch:
		base.Code = PluginErrDependencyVersion
		base.Message = "依赖插件版本不满足，无法执行该操作"
		base.WithSuggestion("请升级或降级依赖插件到兼容版本后重试。")
	case pluginregistry.DependencyCircularDependency, pluginregistry.DependencySelfDependency:
		base.Code = PluginErrDependencyCycle
		base.Message = "存在循环依赖，无法执行该操作"
		base.WithSuggestion("请调整插件依赖关系，避免循环依赖。")
	default:
		base.Code = PluginErrDependencyMissing
	}
	if strings.TrimSpace(check.Message) != "" {
		base.WithDetail("reason", strings.TrimSpace(check.Message))
	}
	return base
}

func coreVersionIncompatible(owner string, compatibility domain.PluginCoreCompatibility) *domain.APIError {
	err := domain.NewPluginError(PluginErrCoreVersionIncompat, "插件要求的 Core 版本不兼容，无法执行该操作").
		WithStatus(400).
		WithDetail("plugin_code", strings.TrimSpace(owner)).
		WithDetail("core_version", strings.TrimSpace(compatibility.CoreVersion)).
		WithDetail("min_core_version", strings.TrimSpace(compatibility.MinCoreVersion)).
		WithDetail("compatible_core_version", strings.TrimSpace(compatibility.CompatibleCoreVersion)).
		WithDetail("messages", compatibility.Messages).
		WithSuggestion("请升级 Core 或选择兼容版本的插件后重试。")
	return err
}
