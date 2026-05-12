package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

var (
	manifestCodePattern       = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
	manifestPermissionPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*){1,5}$`)
	manifestVersionPattern    = regexp.MustCompile(`^[0-9]+(\.[0-9]+){0,2}([-.+][A-Za-z0-9_.-]+)?$`)
	manifestAllowedHooks      = map[string]bool{
		HookBeforeCreateContent:   true,
		HookAfterCreateContent:    true,
		HookBeforeUpdateContent:   true,
		HookAfterUpdateContent:    true,
		HookBeforeModerateContent: true,
		HookAfterModerateContent:  true,
		HookBeforeBuildSEO:        true,
		HookAfterBuildSEO:         true,
		"BeforeDeleteContent":     true,
		"AfterDeleteContent":      true,
		HookAfterCreateComment:    true,
		HookOnSearchIndex:         true,
		HookOnNotificationBuild:   true,
		HookOnSEOBuild:            true,
		HookAfterPluginEnabled:    true,
		HookAfterPluginDisabled:   true,
	}
	manifestFailurePolicies = map[string]bool{"": true, "block": true, "log": true, "retry_later": true}
	manifestHookModes       = map[string]bool{"": true, string(HookBlocking): true, string(HookNonBlocking): true}
)

// DecodePluginManifestJSON accepts DevHub's manifest wire format. It supports
// both legacy content_types:["question"] and SDK-style content_types:[{...}].
func DecodePluginManifestJSON(raw []byte) (domain.PluginManifest, string, error) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return domain.PluginManifest{}, "", fmt.Errorf("manifest 不能为空")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return domain.PluginManifest{}, "", fmt.Errorf("manifest 必须是合法 JSON: %w", err)
	}
	contentTypesRaw := fields["content_types"]
	delete(fields, "content_types")
	normalizedRaw, _ := json.Marshal(fields)
	var manifest domain.PluginManifest
	if err := json.Unmarshal(normalizedRaw, &manifest); err != nil {
		return domain.PluginManifest{}, "", err
	}
	if len(contentTypesRaw) > 0 {
		var names []string
		if err := json.Unmarshal(contentTypesRaw, &names); err == nil {
			manifest.ContentTypes = names
		} else {
			var defs []domain.ContentTypeDefinition
			if err := json.Unmarshal(contentTypesRaw, &defs); err != nil {
				return domain.PluginManifest{}, "", fmt.Errorf("content_types 必须是字符串数组或内容类型对象数组")
			}
			manifest.ContentTypeDefs = append(manifest.ContentTypeDefs, defs...)
			for _, item := range defs {
				if item.Type != "" {
					manifest.ContentTypes = append(manifest.ContentTypes, item.Type)
				}
			}
		}
	}
	manifest = NormalizeManifest(manifest)
	checksum := ManifestChecksum(manifest)
	return manifest, checksum, nil
}

func ManifestChecksum(manifest domain.PluginManifest) string {
	raw, _ := json.Marshal(manifest)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func NormalizeManifest(manifest domain.PluginManifest) domain.PluginManifest {
	manifest.Code = strings.TrimSpace(firstNonBlank(manifest.Code, manifest.PluginCode))
	manifest.PluginCode = manifest.Code
	manifest.Name = strings.TrimSpace(manifest.Name)
	manifest.Version = strings.TrimSpace(manifest.Version)
	manifest.SourceType = firstNonBlank(manifest.SourceType, "manifest")
	seenTypes := map[string]bool{}
	normalizedTypes := []string{}
	for _, item := range manifest.ContentTypeDefs {
		typ := NormalizeContentType(item.Type)
		if typ != "" && !seenTypes[typ] {
			seenTypes[typ] = true
			normalizedTypes = append(normalizedTypes, typ)
		}
	}
	for _, typ := range manifest.ContentTypes {
		typ = NormalizeContentType(typ)
		if typ != "" && !seenTypes[typ] {
			seenTypes[typ] = true
			normalizedTypes = append(normalizedTypes, typ)
		}
	}
	sort.Strings(normalizedTypes)
	manifest.ContentTypes = normalizedTypes
	defs := make([]domain.ContentTypeDefinition, 0, len(normalizedTypes))
	existingDef := map[string]domain.ContentTypeDefinition{}
	for _, item := range manifest.ContentTypeDefs {
		item.Type = NormalizeContentType(item.Type)
		if item.Type == "" {
			continue
		}
		existingDef[item.Type] = item
	}
	for _, typ := range normalizedTypes {
		item := existingDef[typ]
		item.Type = typ
		item.PluginCode = firstNonBlank(item.PluginCode, manifest.Code)
		if item.Name == "" {
			item.Name = typ
		}
		defs = append(defs, item)
	}
	manifest.ContentTypeDefs = defs
	for i := range manifest.Permissions {
		manifest.Permissions[i].PluginCode = firstNonBlank(manifest.Permissions[i].PluginCode, manifest.Code)
	}
	for i := range manifest.Menus {
		manifest.Menus[i].PluginCode = firstNonBlank(manifest.Menus[i].PluginCode, manifest.Code)
	}
	for i := range manifest.Routes {
		manifest.Routes[i].PluginCode = firstNonBlank(manifest.Routes[i].PluginCode, manifest.Code)
		manifest.Routes[i].Method = strings.ToUpper(strings.TrimSpace(manifest.Routes[i].Method))
	}
	for i := range manifest.Hooks {
		manifest.Hooks[i].PluginCode = firstNonBlank(manifest.Hooks[i].PluginCode, manifest.Code)
	}
	for i := range manifest.Migrations {
		manifest.Migrations[i].PluginCode = firstNonBlank(manifest.Migrations[i].PluginCode, manifest.Code)
		if manifest.Migrations[i].Direction == "" {
			manifest.Migrations[i].Direction = "up"
		}
	}
	return manifest
}

func ValidatePluginManifest(manifest domain.PluginManifest, existing []domain.Plugin, currentCoreVersion string) domain.PluginManifestValidationResult {
	manifest = NormalizeManifest(manifest)
	result := domain.PluginManifestValidationResult{
		Valid:              true,
		Errors:             []string{},
		Warnings:           []string{},
		NormalizedManifest: manifest,
		Checksum:           ManifestChecksum(manifest),
		Dependencies:       append([]string(nil), manifest.Dependencies...),
		MigrationPlan:      append([]domain.PluginMigrationDefinition(nil), manifest.Migrations...),
		InstallPreview: map[string]any{
			"initial_status": "disabled",
			"source_type":    firstNonBlank(manifest.SourceType, "manifest"),
			"executes_code":  false,
		},
	}
	result.ImpactSummary = domain.PluginInstallImpact{
		ContentTypesCount: len(manifest.ContentTypes),
		PermissionsCount:  len(manifest.Permissions),
		MenusCount:        len(manifest.Menus),
		RoutesCount:       len(manifest.Routes),
		HooksCount:        len(manifest.Hooks),
		MigrationsCount:   len(manifest.Migrations),
		Dependencies:      append([]string(nil), manifest.Dependencies...),
	}
	addError := func(format string, args ...any) {
		result.Errors = append(result.Errors, fmt.Sprintf(format, args...))
		result.Valid = false
	}
	addWarning := func(format string, args ...any) {
		result.Warnings = append(result.Warnings, fmt.Sprintf(format, args...))
	}

	if !manifestCodePattern.MatchString(manifest.Code) {
		addError("code 必填且只能使用小写字母、数字和下划线，并以字母开头")
	}
	if strings.TrimSpace(manifest.Name) == "" {
		addError("name 必填")
	}
	if !manifestVersionPattern.MatchString(manifest.Version) {
		addError("version 必填且应为合法版本号")
	}
	if manifest.CompatibleCoreVersion != "" && currentCoreVersion != "" && !compatibleCoreVersion(manifest.CompatibleCoreVersion, currentCoreVersion) {
		addError("compatible_core_version 与当前 Core 版本不兼容")
	}

	existingCodes := map[string]bool{}
	existingTypes := map[string]string{}
	existingPerms := map[string]string{}
	for _, plugin := range existing {
		existingCodes[plugin.Code] = true
		for _, typ := range plugin.ContentTypes {
			existingTypes[NormalizeContentType(typ)] = plugin.Code
		}
		for _, def := range plugin.ContentTypeDefs {
			existingTypes[NormalizeContentType(def.Type)] = plugin.Code
		}
		for _, perm := range plugin.Permissions {
			existingPerms[strings.TrimSpace(perm.Code)] = plugin.Code
		}
	}
	if existingCodes[manifest.Code] {
		addError("插件 code 已存在：%s", manifest.Code)
	}

	permissionSet := map[string]bool{}
	for _, perm := range manifest.Permissions {
		code := strings.TrimSpace(perm.Code)
		if !manifestPermissionPattern.MatchString(code) {
			addError("权限码不合法：%s", code)
			continue
		}
		if owner := existingPerms[code]; owner != "" && owner != manifest.Code {
			result.PermissionConflicts = append(result.PermissionConflicts, code)
			addError("权限码与插件 %s 冲突：%s", owner, code)
		}
		permissionSet[code] = true
	}
	for _, item := range manifest.ContentTypeDefs {
		if item.Type == "" {
			addError("content_type 不能为空")
			continue
		}
		if item.PluginCode != manifest.Code {
			addError("content_type %s 的 plugin_code 必须为 %s", item.Type, manifest.Code)
		}
		if owner := existingTypes[item.Type]; owner != "" && owner != manifest.Code {
			result.ContentTypeConflicts = append(result.ContentTypeConflicts, item.Type)
			addError("content_type 与插件 %s 冲突：%s", owner, item.Type)
		}
		if strings.TrimSpace(item.CreatePermission) == "" {
			addError("content_type %s 缺少 create_permission", item.Type)
		} else if !permissionSet[item.CreatePermission] {
			addError("content_type %s 引用了未声明的 create_permission：%s", item.Type, item.CreatePermission)
		}
	}
	for _, menu := range manifest.Menus {
		if menu.Path == "" || !strings.HasPrefix(menu.Path, "/") {
			addError("菜单 %s path 必须以 / 开头", firstNonBlank(menu.Code, menu.Title))
		}
		if menu.Permission != "" && !permissionSet[menu.Permission] {
			addError("菜单 %s 引用了未声明权限：%s", firstNonBlank(menu.Code, menu.Title), menu.Permission)
		}
	}
	for _, route := range manifest.Routes {
		if route.Path == "" || !strings.HasPrefix(route.Path, "/") {
			addError("路由 path 必须以 / 开头：%s", route.Path)
		}
		if route.Method == "" {
			addError("路由 method 不能为空：%s", route.Path)
		}
		if route.Permission != "" && !permissionSet[route.Permission] {
			addError("路由 %s 引用了未声明权限：%s", route.Path, route.Permission)
		}
	}
	for _, hook := range manifest.Hooks {
		if !manifestAllowedHooks[hook.Name] {
			addError("Hook 名称不支持：%s", hook.Name)
		}
		if !manifestHookModes[hook.Mode] {
			addError("Hook mode 不合法：%s", hook.Mode)
		}
		if !manifestFailurePolicies[hook.FailurePolicy] {
			addError("Hook failure_policy 不合法：%s", hook.FailurePolicy)
		}
		if hook.TimeoutMS < 0 || hook.TimeoutMS > 30000 {
			addError("Hook timeout_ms 必须在 0-30000 范围内")
		}
	}
	if schema, ok := manifest.ConfigSchema.(map[string]any); ok {
		if err := validateSchemaShape(schema, "$"); err != nil {
			addError("config_schema 不合法：%s", err.Error())
		}
	}
	for _, migration := range manifest.Migrations {
		if strings.TrimSpace(migration.MigrationName) == "" {
			addError("migration_name 不能为空")
		}
		if migration.Direction != "" && migration.Direction != "up" {
			addError("当前只支持 up migration：%s", migration.MigrationName)
		}
		if migration.Checksum == "" {
			addWarning("migration %s 未声明 checksum", migration.MigrationName)
		}
	}
	for _, asset := range manifest.Assets {
		if strings.Contains(asset, "..") || filepath.IsAbs(asset) {
			addError("asset 路径不安全：%s", asset)
		}
		if isDangerousAsset(asset) {
			addError("asset 不允许使用危险可执行文件：%s", asset)
		}
	}
	if manifest.ExternalService != nil {
		validateExternalService(manifest.ExternalService, addError, addWarning)
		result.ImpactSummary.SecurityWarnings = append(result.ImpactSummary.SecurityWarnings, "外部服务型插件只允许通过 Webhook 元信息接入，不执行本地第三方代码")
	}
	for _, dep := range manifest.Dependencies {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		if dep == manifest.Code {
			addError("插件不能依赖自身")
			continue
		}
		if !existingCodes[dep] {
			addWarning("依赖插件缺失：%s", dep)
		}
	}
	sort.Strings(result.Errors)
	sort.Strings(result.Warnings)
	return result
}

func ValidatePluginManifestJSON(raw []byte, existing []domain.Plugin, currentCoreVersion string) domain.PluginManifestValidationResult {
	manifest, checksum, err := DecodePluginManifestJSON(raw)
	if err != nil {
		return domain.PluginManifestValidationResult{Valid: false, Errors: []string{err.Error()}}
	}
	result := ValidatePluginManifest(manifest, existing, currentCoreVersion)
	result.Checksum = checksum
	return result
}

func validateSchemaShape(schema map[string]any, path string) error {
	typ := strings.TrimSpace(asString(schema["type"]))
	if typ != "" {
		switch typ {
		case "object", "string", "number", "integer", "boolean", "array":
		default:
			return fmt.Errorf("%s type 不支持：%s", path, typ)
		}
	}
	if props, ok := schema["properties"].(map[string]any); ok {
		for key, raw := range props {
			sub, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("%s.properties.%s 必须是 object", path, key)
			}
			if err := validateSchemaShape(sub, path+"."+key); err != nil {
				return err
			}
		}
	}
	if rawItems, ok := schema["items"]; ok {
		if _, ok := rawItems.(map[string]any); !ok {
			return fmt.Errorf("%s.items 必须是 object", path)
		}
	}
	return nil
}

func validateExternalService(service *domain.PluginExternalService, addError func(string, ...any), addWarning func(string, ...any)) {
	if strings.TrimSpace(service.Endpoint) == "" {
		addError("external_service.endpoint 不能为空")
		return
	}
	parsed, err := url.Parse(service.Endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		addError("external_service.endpoint 必须是合法 URL")
		return
	}
	if parsed.Scheme != "https" {
		addWarning("external_service.endpoint 推荐使用 https")
	}
	if service.TimeoutMS <= 0 {
		addWarning("external_service.timeout_ms 未设置，将使用系统默认超时")
	}
	if service.TimeoutMS > 10000 {
		addError("external_service.timeout_ms 不能超过 10000")
	}
	if !manifestFailurePolicies[service.FailurePolicy] {
		addError("external_service.failure_policy 不合法：%s", service.FailurePolicy)
	}
}

func compatibleCoreVersion(requirement, current string) bool {
	requirement = strings.TrimSpace(requirement)
	current = strings.TrimPrefix(strings.TrimSpace(current), "v")
	if requirement == "" || current == "" {
		return true
	}
	if strings.HasPrefix(requirement, ">=") {
		return versionCompare(current, strings.TrimSpace(strings.TrimPrefix(requirement, ">="))) >= 0
	}
	if strings.HasPrefix(requirement, ">") {
		return versionCompare(current, strings.TrimSpace(strings.TrimPrefix(requirement, ">"))) > 0
	}
	if strings.HasPrefix(requirement, "<=") {
		return versionCompare(current, strings.TrimSpace(strings.TrimPrefix(requirement, "<="))) <= 0
	}
	if strings.HasPrefix(requirement, "<") {
		return versionCompare(current, strings.TrimSpace(strings.TrimPrefix(requirement, "<"))) < 0
	}
	return versionCompare(current, strings.TrimPrefix(requirement, "v")) == 0
}

// CompareVersionStrings compares semantic-ish version strings.
// Returns 1 if a > b, -1 if a < b, and 0 if equal.
func CompareVersionStrings(a, b string) int {
	return versionCompare(a, b)
}

func versionCompare(a, b string) int {
	pa := strings.Split(strings.TrimPrefix(a, "v"), ".")
	pb := strings.Split(strings.TrimPrefix(b, "v"), ".")
	for len(pa) < 3 {
		pa = append(pa, "0")
	}
	for len(pb) < 3 {
		pb = append(pb, "0")
	}
	for i := 0; i < 3; i++ {
		ai := atoiLoose(pa[i])
		bi := atoiLoose(pb[i])
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return 0
}

func atoiLoose(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func isDangerousAsset(path string) bool {
	lower := strings.ToLower(path)
	for _, suffix := range []string{".sh", ".bash", ".exe", ".dll", ".so", ".dylib", ".bat", ".cmd"} {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}
