package plugins

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

const (
	DependencySatisfied          = "satisfied"
	DependencyMissing            = "missing"
	DependencyDisabled           = "disabled"
	DependencyArchived           = "archived"
	DependencyMigrationFailed    = "migration_failed"
	DependencyConfigInvalid      = "config_invalid"
	DependencyVersionMismatch    = "version_mismatch"
	DependencyCircularDependency = "circular_dependency"
	DependencySelfDependency     = "self_dependency"
	DependencyOptionalMissing    = "optional_missing"

	CompatibilityCompatible   = "compatible"
	CompatibilityWarning      = "warning"
	CompatibilityIncompatible = "incompatible"
)

var numericVersionPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+){0,2}$`)

func CheckPluginVersionCompatibility(manifest domain.PluginManifest, currentCoreVersion string) domain.PluginCoreCompatibility {
	core := strings.TrimPrefix(strings.TrimSpace(currentCoreVersion), "v")
	result := domain.PluginCoreCompatibility{
		CoreVersion:           firstNonBlank(currentCoreVersion, core),
		MinCoreVersion:        strings.TrimSpace(manifest.MinCoreVersion),
		CompatibleCoreVersion: strings.TrimSpace(manifest.CompatibleCoreVersion),
		Status:                CompatibilityCompatible,
	}
	if result.MinCoreVersion == "" {
		result.Status = CompatibilityWarning
		result.Messages = append(result.Messages, "min_core_version 未声明")
	} else if !numericVersionPattern.MatchString(strings.TrimPrefix(result.MinCoreVersion, "v")) {
		result.Status = CompatibilityIncompatible
		result.Messages = append(result.Messages, "min_core_version 仅支持 x.y.z 数字版本")
	} else if versionCompare(core, result.MinCoreVersion) < 0 {
		result.Status = CompatibilityIncompatible
		result.Messages = append(result.Messages, fmt.Sprintf("当前 Core %s 低于最低要求 %s", currentCoreVersion, result.MinCoreVersion))
	}
	if result.CompatibleCoreVersion != "" {
		ok, err := CheckVersionConstraint(core, result.CompatibleCoreVersion)
		if err != nil {
			result.Status = CompatibilityIncompatible
			result.Messages = append(result.Messages, "compatible_core_version 语法不支持："+err.Error())
		} else if !ok {
			result.Status = CompatibilityIncompatible
			result.Messages = append(result.Messages, fmt.Sprintf("当前 Core %s 不满足兼容范围 %s", currentCoreVersion, result.CompatibleCoreVersion))
		}
	}
	if len(result.Messages) == 0 {
		result.Messages = append(result.Messages, "当前 Core 版本兼容")
	}
	return result
}

func CheckVersionConstraint(version, constraint string) (bool, error) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	constraint = strings.TrimSpace(constraint)
	if constraint == "" {
		return true, nil
	}
	if !numericVersionPattern.MatchString(version) {
		return false, fmt.Errorf("版本 %s 不是 x.y.z 数字版本", version)
	}
	parts := strings.Fields(constraint)
	if len(parts) == 0 {
		return true, nil
	}
	for _, part := range parts {
		if part == "" {
			continue
		}
		op := "="
		want := part
		for _, candidate := range []string{">=", "<=", ">", "<"} {
			if strings.HasPrefix(part, candidate) {
				op = candidate
				want = strings.TrimSpace(strings.TrimPrefix(part, candidate))
				break
			}
		}
		want = strings.TrimPrefix(want, "v")
		if !numericVersionPattern.MatchString(want) {
			return false, fmt.Errorf("约束 %s 仅支持 x.y.z 数字版本和 >= / < / <= / > / 精确版本", part)
		}
		cmp := versionCompare(version, want)
		switch op {
		case "=":
			if cmp != 0 {
				return false, nil
			}
		case ">=":
			if cmp < 0 {
				return false, nil
			}
		case ">":
			if cmp <= 0 {
				return false, nil
			}
		case "<=":
			if cmp > 0 {
				return false, nil
			}
		case "<":
			if cmp >= 0 {
				return false, nil
			}
		}
	}
	return true, nil
}

func ResolvePluginDependencies(manifest domain.PluginManifest, existing []domain.Plugin) ([]domain.PluginDependencyCheck, domain.PluginDependencySummary) {
	manifest = NormalizeManifest(manifest)
	byCode := pluginsByCode(existing)
	checks := make([]domain.PluginDependencyCheck, 0, len(manifest.Dependencies))
	for _, dep := range manifest.Dependencies {
		check := resolveOneDependency(manifest.Code, dep, byCode)
		checks = append(checks, check)
	}
	for i, dep := range manifest.Dependencies {
		if checks[i].Status == DependencySelfDependency {
			continue
		}
		if chain := dependencyCycleChain(manifest.Code, dep.Code, byCode, []string{manifest.Code}, map[string]bool{manifest.Code: true}); len(chain) > 0 {
			checks[i].Status = DependencyCircularDependency
			checks[i].Satisfied = false
			checks[i].Chain = chain
			checks[i].Message = "存在循环依赖：" + strings.Join(chain, " -> ")
		}
	}
	return checks, summarizeDependencies(checks)
}

func ValidatePluginDependencies(manifest domain.PluginManifest, existing []domain.Plugin) ([]domain.PluginDependencyCheck, domain.PluginDependencySummary, []string, []string) {
	checks, summary := ResolvePluginDependencies(manifest, existing)
	errors := []string{}
	warnings := []string{}
	for _, check := range checks {
		if check.Satisfied {
			continue
		}
		message := firstNonBlank(check.Message, fmt.Sprintf("依赖 %s 未满足：%s", check.Code, check.Status))
		if check.Required || check.Status == DependencySelfDependency || check.Status == DependencyCircularDependency {
			errors = append(errors, message)
		} else {
			warnings = append(warnings, message)
		}
	}
	sort.Strings(errors)
	sort.Strings(warnings)
	return checks, summary, errors, warnings
}

func DependencyDiff(current, next []domain.PluginDependency) domain.PluginDependencyDiff {
	currentMap := map[string]domain.PluginDependency{}
	nextMap := map[string]domain.PluginDependency{}
	for _, dep := range current {
		currentMap[dep.Code] = dep
	}
	for _, dep := range next {
		nextMap[dep.Code] = dep
	}
	diff := domain.PluginDependencyDiff{}
	changed := map[string]bool{}
	for code, dep := range nextMap {
		old, ok := currentMap[code]
		if !ok {
			diff.Added = append(diff.Added, dep)
			changed[code] = true
			continue
		}
		if strings.TrimSpace(old.Version) != strings.TrimSpace(dep.Version) {
			diff.VersionChanged = append(diff.VersionChanged, dep)
			changed[code] = true
		}
		if old.Required != dep.Required {
			diff.RequiredChanged = append(diff.RequiredChanged, dep)
			changed[code] = true
		}
	}
	for code, dep := range currentMap {
		if _, ok := nextMap[code]; !ok {
			diff.Removed = append(diff.Removed, dep)
			changed[code] = true
		}
	}
	for code := range changed {
		diff.ChangedDependencies = append(diff.ChangedDependencies, code)
	}
	sort.Slice(diff.Added, func(i, j int) bool { return diff.Added[i].Code < diff.Added[j].Code })
	sort.Slice(diff.Removed, func(i, j int) bool { return diff.Removed[i].Code < diff.Removed[j].Code })
	sort.Slice(diff.VersionChanged, func(i, j int) bool { return diff.VersionChanged[i].Code < diff.VersionChanged[j].Code })
	sort.Slice(diff.RequiredChanged, func(i, j int) bool { return diff.RequiredChanged[i].Code < diff.RequiredChanged[j].Code })
	sort.Strings(diff.ChangedDependencies)
	return diff
}

func resolveOneDependency(owner string, dep domain.PluginDependency, byCode map[string]domain.Plugin) domain.PluginDependencyCheck {
	check := domain.PluginDependencyCheck{
		Code:      strings.TrimSpace(dep.Code),
		Version:   strings.TrimSpace(dep.Version),
		Required:  dep.Required,
		Reason:    strings.TrimSpace(dep.Reason),
		Status:    DependencySatisfied,
		Satisfied: true,
	}
	if check.Code == owner {
		check.Status = DependencySelfDependency
		check.Satisfied = false
		check.Message = "插件不能依赖自身"
		return check
	}
	plugin, ok := byCode[check.Code]
	if !ok || plugin.Code == "" {
		check.Satisfied = false
		if check.Required {
			check.Status = DependencyMissing
			check.Message = "required 依赖插件不存在或未安装：" + check.Code
		} else {
			check.Status = DependencyOptionalMissing
			check.Message = "optional 依赖插件不存在或未安装：" + check.Code
		}
		return check
	}
	check.PluginName = plugin.Name
	check.CurrentVersion = plugin.Version
	check.CurrentStatus = plugin.Status
	switch strings.TrimSpace(plugin.Status) {
	case StatusEnabled:
	case StatusArchived:
		check.Status = DependencyArchived
	case StatusMigrationFailed:
		check.Status = DependencyMigrationFailed
	case StatusConfigInvalid:
		check.Status = DependencyConfigInvalid
	default:
		check.Status = DependencyDisabled
	}
	if check.Status != DependencySatisfied {
		check.Satisfied = false
		check.Message = fmt.Sprintf("依赖插件 %s 当前状态为 %s", check.Code, plugin.Status)
		return check
	}
	if check.Version != "" {
		ok, err := CheckVersionConstraint(plugin.Version, check.Version)
		if err != nil {
			check.Status = DependencyVersionMismatch
			check.Satisfied = false
			check.Message = fmt.Sprintf("依赖插件 %s 版本约束不支持：%s", check.Code, err.Error())
			return check
		}
		if !ok {
			check.Status = DependencyVersionMismatch
			check.Satisfied = false
			check.Message = fmt.Sprintf("依赖插件 %s 当前版本 %s 不满足 %s", check.Code, plugin.Version, check.Version)
		}
	}
	return check
}

func summarizeDependencies(checks []domain.PluginDependencyCheck) domain.PluginDependencySummary {
	summary := domain.PluginDependencySummary{Total: len(checks)}
	for _, check := range checks {
		if check.Required {
			summary.Required++
		} else {
			summary.Optional++
		}
		if check.Satisfied {
			summary.Satisfied++
			continue
		}
		if check.Required || check.Status == DependencySelfDependency || check.Status == DependencyCircularDependency {
			summary.Blocking++
		} else {
			summary.Warnings++
		}
		switch check.Status {
		case DependencyMissing, DependencyOptionalMissing:
			summary.Missing++
		case DependencyDisabled:
			summary.Disabled++
		case DependencyArchived:
			summary.Archived++
		case DependencyVersionMismatch:
			summary.VersionIssues++
		case DependencyCircularDependency, DependencySelfDependency:
			summary.Cycles++
		}
	}
	return summary
}

func dependencyCycleChain(root, code string, byCode map[string]domain.Plugin, chain []string, seen map[string]bool) []string {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	if code == root {
		return append(chain, code)
	}
	if seen[code] {
		return nil
	}
	plugin, ok := byCode[code]
	if !ok {
		return nil
	}
	seen[code] = true
	nextChain := append(chain, code)
	for _, dep := range plugin.Dependencies {
		if dep.Code == "" {
			continue
		}
		if found := dependencyCycleChain(root, dep.Code, byCode, nextChain, cloneSeen(seen)); len(found) > 0 {
			return found
		}
	}
	return nil
}

func cloneSeen(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func pluginsByCode(items []domain.Plugin) map[string]domain.Plugin {
	out := map[string]domain.Plugin{}
	for _, plugin := range items {
		out[strings.TrimSpace(plugin.Code)] = plugin
	}
	return out
}
