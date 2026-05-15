package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginPackageCompatOperator struct {
	ID   int64
	Name string
}

type compatIssue struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Section    string         `json:"section,omitempty"`
	Path       string         `json:"path,omitempty"`
	Suggestion string         `json:"suggestion,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

func (s *Service) RunPluginPackageCompatCheckAs(operator PluginPackageCompatOperator, precheckID int64) (domain.PluginPackageCompatCheckResponse, error) {
	precheck, ok := s.repo.PluginPackagePrecheckByID(precheckID)
	if !ok || precheck.ID <= 0 {
		return domain.PluginPackageCompatCheckResponse{}, domain.NewPluginError("plugin_package_precheck_not_found", "插件包预检记录不存在").
			WithStatus(404).
			WithSuggestion("请先完成插件包解压安全检查与 manifest 预校验。")
	}
	if precheck.Status != domain.PluginPackagePrecheckStatusPassed {
		return domain.PluginPackageCompatCheckResponse{}, domain.NewPluginError("plugin_package_compat_precheck_not_passed", "只有预检通过的插件包才能执行兼容性检查").
			WithStatus(400).
			WithDetail("precheck_id", precheck.ID).
			WithDetail("precheck_status", precheck.Status).
			WithSuggestion("请修复预检错误后重新执行预检。")
	}
	if strings.TrimSpace(precheck.ManifestJSON) == "" {
		return domain.PluginPackageCompatCheckResponse{}, domain.NewPluginError("plugin_package_compat_manifest_missing", "预检记录缺少 manifest_json").
			WithStatus(400).
			WithDetail("precheck_id", precheck.ID).
			WithSuggestion("请重新执行 manifest 预校验。")
	}
	if err := s.ensureCompatPrecheckSourceUsable(precheck); err != nil {
		return domain.PluginPackageCompatCheckResponse{}, err
	}

	started := Now()
	record := domain.PluginPackageCompatCheckRecord{
		PackageDownloadID: precheck.PackageDownloadID,
		PackagePrecheckID: precheck.ID,
		PluginCode:        precheck.PluginCode,
		Version:           precheck.Version,
		Status:            domain.PluginPackageCompatCheckStatusChecking,
		CreatedBy:         operator.ID,
		StartedAt:         started,
		CreatedAt:         started,
	}
	record, _ = s.repo.AppendPluginPackageCompatCheck(record)

	manifest, _, err := pluginregistry.DecodePluginManifestJSON([]byte(precheck.ManifestJSON))
	if err != nil {
		record.Status = domain.PluginPackageCompatCheckStatusFailed
		record.ErrorsJSON = mustJSON([]string{"manifest_json 不合法：" + err.Error()})
		record.FinishedAt = Now()
		record, _ = s.repo.SavePluginPackageCompatCheck(record)
		return compatRecordResponse(record), nil
	}
	if record.PluginCode == "" {
		record.PluginCode = manifest.Code
	}
	if record.Version == "" {
		record.Version = manifest.Version
	}

	evaluation := s.evaluatePluginPackageCompatibility(manifest, []byte(precheck.ManifestJSON), precheck)
	record.Status = evaluation.status
	record.CanInstall = evaluation.canInstall
	record.CoreVersion = evaluation.coreVersion
	record.CompatibleCoreVersion = manifest.CompatibleCoreVersion
	record.DependencyResultJSON = mustJSON(scrubAnyForSnapshot(evaluation.dependencyResult))
	record.ConflictResultJSON = mustJSON(scrubAnyForSnapshot(evaluation.conflictResult))
	record.PermissionResultJSON = mustJSON(scrubAnyForSnapshot(evaluation.permissionResult))
	record.RouteResultJSON = mustJSON(scrubAnyForSnapshot(evaluation.routeResult))
	record.MenuResultJSON = mustJSON(scrubAnyForSnapshot(evaluation.menuResult))
	record.HookResultJSON = mustJSON(scrubAnyForSnapshot(evaluation.hookResult))
	record.ConfigSchemaResultJSON = mustJSON(scrubAnyForSnapshot(evaluation.configSchemaResult))
	record.MigrationResultJSON = mustJSON(scrubAnyForSnapshot(evaluation.migrationResult))
	record.WarningsJSON = mustJSON(evaluation.warningMessages())
	record.ErrorsJSON = mustJSON(evaluation.errorMessages())
	record.SummaryJSON = mustJSON(scrubAnyForSnapshot(evaluation.summary))
	record.FinishedAt = Now()
	record, _ = s.repo.SavePluginPackageCompatCheck(record)
	return compatRecordResponse(record), nil
}

func (s *Service) ensureCompatPrecheckSourceUsable(precheck domain.PluginPackagePrecheckRecord) error {
	if precheck.PackageDownloadID > 0 {
		download, ok := s.repo.PluginPackageDownloadByID(precheck.PackageDownloadID)
		if !ok {
			return domain.NewPluginError("plugin_package_download_not_found", "预检关联的 staging 下载记录不存在").
				WithStatus(404).
				WithDetail("package_download_id", precheck.PackageDownloadID)
		}
		switch download.Status {
		case domain.PluginPackageDownloadStatusChecksumFailed, domain.PluginPackageDownloadStatusRejected, domain.PluginPackageDownloadStatusFailed, domain.PluginPackageDownloadStatusDeleted:
			return domain.NewPluginError("plugin_package_compat_source_invalid", "预检关联的 staging 包状态不允许兼容性检查").
				WithStatus(400).
				WithDetail("package_download_id", download.ID).
				WithDetail("download_status", download.Status)
		}
		if strings.TrimSpace(download.StagingPath) != "" {
			root, err := serviceProjectRoot()
			if err != nil {
				return domain.NewPluginError("plugin_package_compat_source_invalid", "读取项目根目录失败").WithStatus(500)
			}
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(download.StagingPath))); err != nil {
				return domain.NewPluginError("plugin_package_compat_source_missing", "staging 文件已经不存在").
					WithStatus(400).
					WithDetail("package_download_id", download.ID).
					WithSuggestion("请重新下载并执行预检。")
			}
		}
	}
	return nil
}

type compatEvaluation struct {
	status             string
	canInstall         bool
	coreVersion        string
	errors             []compatIssue
	warnings           []compatIssue
	dependencyResult   map[string]any
	conflictResult     map[string]any
	permissionResult   map[string]any
	routeResult        map[string]any
	menuResult         map[string]any
	hookResult         map[string]any
	configSchemaResult map[string]any
	migrationResult    map[string]any
	summary            map[string]any
}

func (ev compatEvaluation) errorMessages() []string {
	out := make([]string, 0, len(ev.errors))
	for _, it := range ev.errors {
		out = append(out, it.Code+": "+it.Message)
	}
	sort.Strings(out)
	return out
}

func (ev compatEvaluation) warningMessages() []string {
	out := make([]string, 0, len(ev.warnings))
	for _, it := range ev.warnings {
		out = append(out, it.Code+": "+it.Message)
	}
	sort.Strings(out)
	return out
}

func (s *Service) evaluatePluginPackageCompatibility(manifest domain.PluginManifest, rawManifest []byte, precheck domain.PluginPackagePrecheckRecord) compatEvaluation {
	coreVersion := currentCoreVersion()
	existing := s.repo.Plugins()
	validation := pluginregistry.ValidatePluginManifest(manifest, existing, coreVersion)
	ev := compatEvaluation{
		status:      domain.PluginPackageCompatCheckStatusPassed,
		canInstall:  true,
		coreVersion: coreVersion,
		errors:      []compatIssue{},
		warnings:    []compatIssue{},
	}
	addError := func(code, section, path, message string, details map[string]any) {
		ev.errors = append(ev.errors, compatIssue{Code: code, Section: section, Path: path, Message: message, Details: details})
	}
	addWarning := func(code, section, path, message string, details map[string]any) {
		ev.warnings = append(ev.warnings, compatIssue{Code: code, Section: section, Path: path, Message: message, Details: details})
	}
	permissionSet := map[string]bool{}
	for _, perm := range manifest.Permissions {
		permissionSet[strings.TrimSpace(perm.Code)] = true
	}

	if strings.TrimSpace(manifest.CompatibleCoreVersion) == "" {
		addError("plugin_package_compat_core_constraint_missing", "core", "compatible_core_version", "manifest 必须声明 compatible_core_version", nil)
	} else if validation.Compatibility.Status == pluginregistry.CompatibilityIncompatible {
		addError("plugin_upgrade_target_core_incompatible", "core", "compatible_core_version", "当前 Core 版本不满足插件兼容范围", map[string]any{
			"core_version": coreVersion, "compatible_core_version": manifest.CompatibleCoreVersion, "messages": validation.Compatibility.Messages,
		})
	} else if validation.Compatibility.Status == pluginregistry.CompatibilityWarning {
		addWarning("plugin_package_compat_core_warning", "core", "compatible_core_version", "Core 兼容声明存在警告", map[string]any{"messages": validation.Compatibility.Messages})
	}

	dependencyChecks, dependencySummary, dependencyErrors, dependencyWarnings := pluginregistry.ValidatePluginDependencies(manifest, existing)
	for _, msg := range dependencyErrors {
		code := "plugin_upgrade_target_dependency_missing"
		if strings.Contains(msg, "不满足") {
			code = "plugin_package_compat_dependency_version_mismatch"
		}
		addError(code, "dependencies", "", msg, nil)
	}
	for _, msg := range dependencyWarnings {
		addWarning("plugin_package_compat_optional_dependency_missing", "dependencies", "", msg, nil)
	}
	ev.dependencyResult = map[string]any{"checks": dependencyChecks, "summary": dependencySummary}

	conflictErrors, conflictWarnings, conflictResult := s.checkCompatConflicts(manifest, precheck)
	for _, it := range conflictErrors {
		ev.errors = append(ev.errors, it)
	}
	for _, it := range conflictWarnings {
		ev.warnings = append(ev.warnings, it)
	}
	ev.conflictResult = conflictResult

	permErrors, permResult := checkCompatPermissions(manifest, permissionSet, existing)
	ev.errors = append(ev.errors, permErrors...)
	ev.permissionResult = permResult

	menuErrors, menuWarnings, menuResult := checkCompatMenus(manifest, permissionSet, existing)
	ev.errors = append(ev.errors, menuErrors...)
	ev.warnings = append(ev.warnings, menuWarnings...)
	ev.menuResult = menuResult

	routeErrors, routeResult := checkCompatRoutes(manifest, permissionSet, existing)
	ev.errors = append(ev.errors, routeErrors...)
	ev.routeResult = routeResult

	hookErrors, hookResult := checkCompatHooks(manifest)
	ev.errors = append(ev.errors, hookErrors...)
	ev.hookResult = hookResult

	configErrors, configWarnings, configResult := checkCompatConfigSchema(manifest, rawManifest)
	ev.errors = append(ev.errors, configErrors...)
	ev.warnings = append(ev.warnings, configWarnings...)
	ev.configSchemaResult = configResult

	migrationErrors, migrationWarnings, migrationResult := checkCompatMigrations(manifest)
	ev.errors = append(ev.errors, migrationErrors...)
	ev.warnings = append(ev.warnings, migrationWarnings...)
	ev.migrationResult = migrationResult

	if strings.EqualFold(strings.TrimSpace(precheck.ChecksumStatus), domain.PluginPackageDownloadStatusChecksumMissing) || strings.EqualFold(strings.TrimSpace(precheck.ChecksumStatus), "checksum_missing") {
		addError("plugin_package_compat_checksum_missing", "checksum", "sha256", "没有 sha256 的 staging 包不能进入安装链路", nil)
	}
	for _, msg := range validation.Errors {
		addError("plugin_package_compat_manifest_invalid", "manifest", "", msg, nil)
	}
	for _, msg := range validation.Warnings {
		addWarning("plugin_package_compat_manifest_warning", "manifest", "", msg, nil)
	}

	ev.canInstall = len(ev.errors) == 0
	ev.summary = map[string]any{
		"core_compatible":     noIssue(ev.errors, "core"),
		"dependencies_ok":     len(dependencyErrors) == 0,
		"conflicts_ok":        noIssue(ev.errors, "conflicts") && noIssue(ev.errors, "content_types") && noIssue(ev.errors, "plugin_code"),
		"permissions_ok":      noIssue(ev.errors, "permissions"),
		"routes_ok":           noIssue(ev.errors, "routes"),
		"menus_ok":            noIssue(ev.errors, "menus"),
		"hooks_ok":            noIssue(ev.errors, "hooks"),
		"config_schema_ok":    noIssue(ev.errors, "config_schema"),
		"migrations_ok":       noIssue(ev.errors, "migrations"),
		"errors_count":        len(ev.errors),
		"warnings_count":      len(ev.warnings),
		"can_install":         ev.canInstall,
		"install_blocked":     !ev.canInstall,
		"does_not_install":    true,
		"does_not_execute":    true,
		"does_not_register":   true,
		"precheck_status":     precheck.Status,
		"package_precheck_id": precheck.ID,
	}
	ev.status = compatStatus(ev)
	return ev
}

func noIssue(items []compatIssue, section string) bool {
	for _, it := range items {
		if it.Section == section {
			return false
		}
	}
	return true
}

func compatStatus(ev compatEvaluation) string {
	if len(ev.errors) == 0 {
		if len(ev.warnings) > 0 {
			return domain.PluginPackageCompatCheckStatusWarning
		}
		return domain.PluginPackageCompatCheckStatusPassed
	}
	for _, it := range ev.errors {
		switch it.Code {
		case "plugin_upgrade_target_core_incompatible", "plugin_package_compat_core_constraint_missing":
			return domain.PluginPackageCompatCheckStatusIncompatible
		case "plugin_upgrade_target_dependency_missing":
			return domain.PluginPackageCompatCheckStatusDependencyMissing
		case "plugin_package_compat_dependency_version_mismatch":
			return domain.PluginPackageCompatCheckStatusDependencyVersionMismatch
		case "plugin_package_compat_plugin_code_conflict", "plugin_package_compat_content_type_conflict", "plugin_package_compat_route_conflict", "plugin_package_compat_permission_conflict":
			return domain.PluginPackageCompatCheckStatusConflictDetected
		}
	}
	return domain.PluginPackageCompatCheckStatusFailed
}

func (s *Service) checkCompatConflicts(manifest domain.PluginManifest, precheck domain.PluginPackagePrecheckRecord) ([]compatIssue, []compatIssue, map[string]any) {
	errors := []compatIssue{}
	warnings := []compatIssue{}
	existingCodes := []string{}
	existingTypes := []string{}
	for _, plugin := range s.repo.Plugins() {
		if plugin.Code == manifest.Code {
			existingCodes = append(existingCodes, plugin.Code)
		}
		for _, typ := range plugin.ContentTypes {
			if pluginregistry.NormalizeContentType(typ) != "" && pluginregistry.NormalizeContentType(typ) == pluginregistry.NormalizeContentType(typ) {
				for _, own := range manifest.ContentTypes {
					if pluginregistry.NormalizeContentType(own) == pluginregistry.NormalizeContentType(typ) {
						existingTypes = append(existingTypes, own)
					}
				}
			}
		}
		for _, def := range plugin.ContentTypeDefs {
			for _, own := range manifest.ContentTypeDefs {
				if pluginregistry.NormalizeContentType(def.Type) == pluginregistry.NormalizeContentType(own.Type) {
					existingTypes = append(existingTypes, own.Type)
				}
			}
		}
	}
	if len(existingCodes) > 0 {
		errors = append(errors, compatIssue{Code: "plugin_package_compat_plugin_code_conflict", Section: "plugin_code", Path: "code", Message: "plugin_code 已与内置或已安装插件冲突", Details: map[string]any{"plugin_code": manifest.Code}})
	}
	seenType := map[string]bool{}
	duplicates := []string{}
	for _, typ := range manifest.ContentTypes {
		n := pluginregistry.NormalizeContentType(typ)
		if n == "" {
			continue
		}
		if seenType[n] {
			duplicates = append(duplicates, n)
		}
		seenType[n] = true
	}
	for _, def := range manifest.ContentTypeDefs {
		if def.PluginCode != "" && def.PluginCode != manifest.Code {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_content_type_owner_invalid", Section: "content_types", Path: def.Type + ".plugin_code", Message: "content_type.plugin_code 必须等于 manifest code"})
		}
	}
	if len(existingTypes) > 0 {
		errors = append(errors, compatIssue{Code: "plugin_package_compat_content_type_conflict", Section: "content_types", Message: "content_type 与现有 Core / 插件声明冲突", Details: map[string]any{"content_types": uniqueStrings(existingTypes)}})
	}
	if len(duplicates) > 0 {
		errors = append(errors, compatIssue{Code: "plugin_package_compat_content_type_duplicate", Section: "content_types", Message: "manifest 内重复声明 content_type", Details: map[string]any{"content_types": uniqueStrings(duplicates)}})
	}
	if strings.EqualFold(strings.TrimSpace(precheck.ChecksumStatus), domain.PluginPackageDownloadStatusChecksumMissing) {
		warnings = append(warnings, compatIssue{Code: "plugin_package_compat_checksum_missing_warning", Section: "checksum", Path: "sha256", Message: "staging 包缺少 sha256，后续安装链路默认阻断"})
	}
	return errors, warnings, map[string]any{
		"plugin_code_conflicts":   uniqueStrings(existingCodes),
		"content_type_conflicts":  uniqueStrings(existingTypes),
		"duplicate_content_types": uniqueStrings(duplicates),
	}
}

func checkCompatPermissions(manifest domain.PluginManifest, permissionSet map[string]bool, existing []domain.Plugin) ([]compatIssue, map[string]any) {
	errors := []compatIssue{}
	duplicates := []string{}
	invalidScopes := []string{}
	conflicts := []string{}
	seen := map[string]bool{}
	allowedScopes := map[string]bool{"global": true, "community": true, "channel": true, "content": true}
	existingPerms := map[string]string{}
	for _, plugin := range existing {
		for _, perm := range plugin.Permissions {
			existingPerms[strings.TrimSpace(perm.Code)] = plugin.Code
		}
	}
	for _, perm := range manifest.Permissions {
		code := strings.TrimSpace(perm.Code)
		if seen[code] {
			duplicates = append(duplicates, code)
		}
		seen[code] = true
		if code == "" || !strings.HasPrefix(code, manifest.Code+".") || strings.HasPrefix(code, "core.") || strings.HasPrefix(code, "admin.") || strings.HasPrefix(code, "system.") {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_permission_forbidden", Section: "permissions", Path: code, Message: "权限码必须以 plugin_code. 开头，且不能声明 core/admin/system 或其他插件权限"})
		}
		if perm.Scope != "" && !allowedScopes[strings.TrimSpace(perm.Scope)] {
			invalidScopes = append(invalidScopes, code)
			errors = append(errors, compatIssue{Code: "plugin_package_compat_permission_scope_invalid", Section: "permissions", Path: code + ".scope", Message: "权限 scope 不在允许范围内"})
		}
		if owner := existingPerms[code]; owner != "" && owner != manifest.Code {
			conflicts = append(conflicts, code)
			errors = append(errors, compatIssue{Code: "plugin_package_compat_permission_conflict", Section: "permissions", Path: code, Message: "权限码与已存在插件冲突", Details: map[string]any{"owner": owner}})
		}
	}
	for _, def := range manifest.ContentTypeDefs {
		if def.CreatePermission != "" && !permissionSet[def.CreatePermission] {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_permission_missing", Section: "permissions", Path: def.Type + ".create_permission", Message: "content_type 引用的 create_permission 未声明"})
		}
	}
	return errors, map[string]any{"duplicates": uniqueStrings(duplicates), "invalid_scopes": uniqueStrings(invalidScopes), "conflicts": uniqueStrings(conflicts)}
}

func checkCompatMenus(manifest domain.PluginManifest, permissionSet map[string]bool, existing []domain.Plugin) ([]compatIssue, []compatIssue, map[string]any) {
	errors := []compatIssue{}
	warnings := []compatIssue{}
	duplicates := []string{}
	conflicts := []string{}
	seen := map[string]bool{}
	allowedArea := map[string]bool{"frontend": true, "admin": true, "moderator": true, "user_center": true}
	existingPaths := map[string]string{}
	for _, plugin := range existing {
		for _, menu := range plugin.Menus {
			existingPaths[strings.TrimSpace(menu.Path)] = plugin.Code
		}
	}
	for _, menu := range manifest.Menus {
		path := strings.TrimSpace(menu.Path)
		area := strings.TrimSpace(firstNonBlank(menu.Area, menu.Location))
		if area != "" && !allowedArea[area] {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_menu_area_invalid", Section: "menus", Path: path, Message: "菜单 area 不在允许范围内"})
		}
		if !safeSitePath(path) || strings.HasPrefix(strings.ToLower(path), "javascript:") || strings.HasPrefix(strings.ToLower(path), "data:") {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_menu_path_invalid", Section: "menus", Path: path, Message: "菜单 path 必须是站内路径，不能是外链/javascript/data"})
		}
		if sensitiveMenuPath(path) {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_menu_sensitive_path", Section: "menus", Path: path, Message: "菜单 path 不能覆盖核心敏感入口"})
		}
		if menu.Permission != "" && !permissionSet[menu.Permission] {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_menu_permission_missing", Section: "menus", Path: path, Message: "菜单引用了未声明权限"})
		}
		if seen[path] {
			duplicates = append(duplicates, path)
			errors = append(errors, compatIssue{Code: "plugin_package_compat_menu_duplicate", Section: "menus", Path: path, Message: "同一插件内菜单 path 重复"})
		}
		seen[path] = true
		if owner := existingPaths[path]; path != "" && owner != "" && owner != manifest.Code {
			conflicts = append(conflicts, path)
			warnings = append(warnings, compatIssue{Code: "plugin_package_compat_menu_path_conflict", Section: "menus", Path: path, Message: "菜单 path 与已有插件重复，当前作为 warning 处理", Details: map[string]any{"owner": owner}})
		}
	}
	return errors, warnings, map[string]any{"duplicates": uniqueStrings(duplicates), "conflicts": uniqueStrings(conflicts)}
}

func checkCompatRoutes(manifest domain.PluginManifest, permissionSet map[string]bool, existing []domain.Plugin) ([]compatIssue, map[string]any) {
	errors := []compatIssue{}
	duplicates := []string{}
	conflicts := []string{}
	seen := map[string]bool{}
	allowedArea := map[string]bool{"frontend": true, "admin": true, "moderator": true, "user_center": true, "api": true}
	allowedMethod := map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true}
	existingRoutes := map[string]string{}
	for _, plugin := range existing {
		for _, route := range plugin.Routes {
			existingRoutes[strings.ToUpper(route.Method)+" "+route.Path] = plugin.Code
		}
	}
	for _, route := range manifest.Routes {
		path := strings.TrimSpace(route.Path)
		method := strings.ToUpper(strings.TrimSpace(route.Method))
		key := method + " " + path
		if route.Area != "" && !allowedArea[route.Area] {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_route_area_invalid", Section: "routes", Path: path, Message: "route area 不在允许范围内"})
		}
		if !allowedMethod[method] {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_route_method_invalid", Section: "routes", Path: path, Message: "route method 不支持"})
		}
		if !safeSitePath(path) {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_route_path_invalid", Section: "routes", Path: path, Message: "route path 必须是站内路径"})
		}
		if sensitiveRoutePath(path) {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_route_sensitive_path", Section: "routes", Path: path, Message: "route path 不能覆盖核心敏感 API 或前台页面"})
		}
		if route.Permission != "" && !permissionSet[route.Permission] {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_route_permission_missing", Section: "routes", Path: path, Message: "route 引用了未声明权限"})
		}
		if seen[key] {
			duplicates = append(duplicates, key)
			errors = append(errors, compatIssue{Code: "plugin_package_compat_route_duplicate", Section: "routes", Path: path, Message: "同一插件内 route method+path 重复"})
		}
		seen[key] = true
		if owner := existingRoutes[key]; owner != "" && owner != manifest.Code {
			conflicts = append(conflicts, key)
			errors = append(errors, compatIssue{Code: "plugin_package_compat_route_conflict", Section: "routes", Path: path, Message: "route 与已有插件冲突", Details: map[string]any{"owner": owner, "method": method}})
		}
	}
	return errors, map[string]any{"duplicates": uniqueStrings(duplicates), "conflicts": uniqueStrings(conflicts)}
}

func checkCompatHooks(manifest domain.PluginManifest) ([]compatIssue, map[string]any) {
	errors := []compatIssue{}
	unsupported := []string{}
	allowedHooks := map[string]bool{
		pluginregistry.HookBeforeCreateContent:   true,
		pluginregistry.HookAfterCreateContent:    true,
		pluginregistry.HookBeforeUpdateContent:   true,
		pluginregistry.HookAfterUpdateContent:    true,
		pluginregistry.HookBeforeModerateContent: true,
		pluginregistry.HookAfterModerateContent:  true,
		pluginregistry.HookBeforeBuildSEO:        true,
		pluginregistry.HookAfterBuildSEO:         true,
		pluginregistry.HookAfterCreateComment:    true,
		pluginregistry.HookOnSearchIndex:         true,
		pluginregistry.HookOnNotificationBuild:   true,
		pluginregistry.HookOnSEOBuild:            true,
		pluginregistry.HookAfterPluginEnabled:    true,
		pluginregistry.HookAfterPluginDisabled:   true,
	}
	for _, hook := range manifest.Hooks {
		if !allowedHooks[hook.Name] {
			unsupported = append(unsupported, hook.Name)
			errors = append(errors, compatIssue{Code: "plugin_package_compat_hook_unknown", Section: "hooks", Path: hook.Name, Message: "HookBus 不支持该 Hook"})
		}
		if hook.Mode != "" && hook.Mode != string(pluginregistry.HookBlocking) && hook.Mode != string(pluginregistry.HookNonBlocking) {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_hook_mode_invalid", Section: "hooks", Path: hook.Name + ".mode", Message: "Hook mode 只能是 blocking 或 non_blocking"})
		}
		if hook.FailurePolicy != "" && hook.FailurePolicy != "block" && hook.FailurePolicy != "log" && hook.FailurePolicy != "ignore" {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_hook_failure_policy_invalid", Section: "hooks", Path: hook.Name + ".failure_policy", Message: "failure_policy 只能是 block/log/ignore"})
		}
		if hook.Mode == string(pluginregistry.HookBlocking) && hook.FailurePolicy == "ignore" {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_hook_blocking_ignore_forbidden", Section: "hooks", Path: hook.Name + ".failure_policy", Message: "blocking Hook 不允许 failure_policy=ignore"})
		}
	}
	return errors, map[string]any{"unsupported_hooks": uniqueStrings(unsupported), "count": len(manifest.Hooks)}
}

func checkCompatConfigSchema(manifest domain.PluginManifest, rawManifest []byte) ([]compatIssue, []compatIssue, map[string]any) {
	errors := []compatIssue{}
	warnings := []compatIssue{}
	result := map[string]any{"schema_present": manifest.ConfigSchema != nil, "default_config_present": false}
	var raw map[string]any
	_ = json.Unmarshal(rawManifest, &raw)
	defaultConfig, hasDefault := raw["default_config"]
	result["default_config_present"] = hasDefault
	if hasDefault && manifest.ConfigSchema != nil {
		if msg := validateDefaultConfigAgainstSimpleSchema(defaultConfig, manifest.ConfigSchema); msg != "" {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_config_default_invalid", Section: "config_schema", Path: "default_config", Message: msg})
		}
	}
	if schemaMap, ok := manifest.ConfigSchema.(map[string]any); ok {
		if unsupported := collectUnsupportedSchemaFeatures(schemaMap, "$"); len(unsupported) > 0 {
			warnings = append(warnings, compatIssue{Code: "plugin_package_compat_config_schema_feature_unsupported", Section: "config_schema", Path: "$", Message: "config_schema 使用了当前简化 schema 能力之外的字段", Details: map[string]any{"features": unsupported}})
			result["unsupported_features"] = unsupported
		}
	}
	return errors, warnings, result
}

func validateDefaultConfigAgainstSimpleSchema(config any, schema any) string {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return ""
	}
	if strings.TrimSpace(asCompatString(schemaMap["type"])) == "object" {
		obj, ok := config.(map[string]any)
		if !ok {
			return "default_config 必须是 object"
		}
		props, _ := schemaMap["properties"].(map[string]any)
		required := map[string]bool{}
		if reqRaw, ok := schemaMap["required"].([]any); ok {
			for _, item := range reqRaw {
				required[asCompatString(item)] = true
			}
		}
		for key := range required {
			if _, ok := obj[key]; !ok {
				return "default_config 缺少 required 字段：" + key
			}
		}
		if schemaMap["additionalProperties"] == false {
			for key := range obj {
				if _, ok := props[key]; !ok {
					return "default_config 包含未知字段：" + key
				}
			}
		}
		for key, specRaw := range props {
			spec, _ := specRaw.(map[string]any)
			if spec == nil {
				continue
			}
			value, ok := obj[key]
			if !ok || value == nil {
				continue
			}
			if !simpleSchemaTypeMatches(value, asCompatString(spec["type"])) {
				return fmt.Sprintf("default_config 字段 %s 类型不匹配", key)
			}
		}
	}
	return ""
}

func simpleSchemaTypeMatches(value any, typ string) bool {
	if typ == "" {
		return true
	}
	switch typ {
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "number", "integer":
		_, ok := value.(float64)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "object":
		_, ok := value.(map[string]any)
		return ok
	default:
		return false
	}
}

func collectUnsupportedSchemaFeatures(schema map[string]any, path string) []string {
	supported := map[string]bool{"type": true, "properties": true, "required": true, "default": true, "enum": true, "format": true, "writeOnly": true, "x-sensitive": true, "additionalProperties": true, "items": true, "description": true, "title": true}
	out := []string{}
	for key, raw := range schema {
		if !supported[key] {
			out = append(out, path+"."+key)
		}
		if key == "properties" {
			if props, ok := raw.(map[string]any); ok {
				for prop, specRaw := range props {
					if spec, ok := specRaw.(map[string]any); ok {
						out = append(out, collectUnsupportedSchemaFeatures(spec, path+".properties."+prop)...)
					}
				}
			}
		}
	}
	sort.Strings(out)
	return out
}

func checkCompatMigrations(manifest domain.PluginManifest) ([]compatIssue, []compatIssue, map[string]any) {
	errors := []compatIssue{}
	warnings := []compatIssue{}
	versions := map[string]bool{}
	names := map[string]bool{}
	duplicates := []string{}
	for _, migration := range manifest.Migrations {
		if migration.Direction != "" && migration.Direction != "up" {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_migration_direction_unsupported", Section: "migrations", Path: migration.MigrationName + ".direction", Message: "本轮只允许 direction=up，且不会执行 migration"})
		}
		if strings.Contains(migration.MigrationName, "..") || filepath.IsAbs(migration.MigrationName) || strings.ContainsAny(migration.MigrationName, `/\\`) {
			errors = append(errors, compatIssue{Code: "plugin_package_compat_migration_name_invalid", Section: "migrations", Path: migration.MigrationName, Message: "migration name 不能包含路径穿越或路径分隔符"})
		}
		if versions[migration.MigrationVersion] {
			duplicates = append(duplicates, migration.MigrationVersion)
		}
		if names[migration.MigrationName] {
			duplicates = append(duplicates, migration.MigrationName)
		}
		versions[migration.MigrationVersion] = true
		names[migration.MigrationName] = true
	}
	if len(duplicates) > 0 {
		errors = append(errors, compatIssue{Code: "plugin_package_compat_migration_duplicate", Section: "migrations", Message: "migration version/name 重复", Details: map[string]any{"duplicates": uniqueStrings(duplicates)}})
	}
	if len(manifest.Migrations) > 20 {
		warnings = append(warnings, compatIssue{Code: "plugin_package_compat_migration_many", Section: "migrations", Message: "migration 数量较多，本轮仅记录 pending 不执行"})
	}
	return errors, warnings, map[string]any{"count": len(manifest.Migrations), "duplicates": uniqueStrings(duplicates), "will_execute": false}
}

func safeSitePath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "..") {
		return false
	}
	if u, err := url.Parse(path); err == nil && u.IsAbs() {
		return false
	}
	lower := strings.ToLower(path)
	return !strings.HasPrefix(lower, "javascript:") && !strings.HasPrefix(lower, "data:")
}

func sensitiveMenuPath(path string) bool {
	for _, item := range []string{"/admin-next", "/login", "/register", "/topics", "/c"} {
		if path == item || strings.HasPrefix(path, item+"/") {
			return true
		}
	}
	return false
}

func sensitiveRoutePath(path string) bool {
	if path == "/" || path == "/topics/:id" || path == "/c/:slug" || path == "/tags/:slug" {
		return true
	}
	for _, item := range []string{"/api/v1/auth/", "/api/v1/admin/auth/", "/api/v1/admin/plugins/", "/api/v1/admin/users/", "/api/v1/admin/system/"} {
		if strings.HasPrefix(path, item) {
			return true
		}
	}
	return false
}

func asCompatString(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func compatRecordResponse(record domain.PluginPackageCompatCheckRecord) domain.PluginPackageCompatCheckResponse {
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
	parseStrings := func(raw string) []string {
		out := []string{}
		if strings.TrimSpace(raw) != "" {
			_ = json.Unmarshal([]byte(raw), &out)
		}
		return out
	}
	return domain.PluginPackageCompatCheckResponse{
		PluginPackageCompatCheckRecord: record,
		DependencyResult:               parseAny(record.DependencyResultJSON),
		ConflictResult:                 parseAny(record.ConflictResultJSON),
		PermissionResult:               parseAny(record.PermissionResultJSON),
		RouteResult:                    parseAny(record.RouteResultJSON),
		MenuResult:                     parseAny(record.MenuResultJSON),
		HookResult:                     parseAny(record.HookResultJSON),
		ConfigSchemaResult:             parseAny(record.ConfigSchemaResultJSON),
		MigrationResult:                parseAny(record.MigrationResultJSON),
		Warnings:                       parseStrings(record.WarningsJSON),
		Errors:                         parseStrings(record.ErrorsJSON),
		Summary:                        parseAny(record.SummaryJSON),
	}
}

func (s *Service) ListPluginPackageCompatChecks(filter domain.PluginPackageCompatCheckFilter) (domain.PluginPackageCompatCheckListResponse, error) {
	items, total, err := s.repo.PluginPackageCompatChecks(filter)
	if err != nil {
		return domain.PluginPackageCompatCheckListResponse{}, err
	}
	res := make([]domain.PluginPackageCompatCheckResponse, 0, len(items))
	summary := map[string]int{}
	for _, item := range items {
		summary[item.Status]++
		res = append(res, compatRecordResponse(item))
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	return domain.PluginPackageCompatCheckListResponse{Items: res, Pagination: domain.Pagination{Page: filter.Page, PageSize: filter.PageSize, Total: total}, Summary: summary}, nil
}

func (s *Service) GetPluginPackageCompatCheck(id int64) (domain.PluginPackageCompatCheckResponse, error) {
	record, ok := s.repo.PluginPackageCompatCheckByID(id)
	if !ok {
		return domain.PluginPackageCompatCheckResponse{}, domain.NewPluginError("plugin_package_compat_check_not_found", "兼容性检查记录不存在").WithStatus(404)
	}
	return compatRecordResponse(record), nil
}

func (s *Service) DeletePluginPackageCompatCheck(id int64) (domain.PluginPackageCompatCheckResponse, error) {
	record, ok := s.repo.PluginPackageCompatCheckByID(id)
	if !ok {
		return domain.PluginPackageCompatCheckResponse{}, domain.NewPluginError("plugin_package_compat_check_not_found", "兼容性检查记录不存在").WithStatus(404)
	}
	record.Status = domain.PluginPackageCompatCheckStatusDeleted
	record, _ = s.repo.SavePluginPackageCompatCheck(record)
	return compatRecordResponse(record), nil
}
