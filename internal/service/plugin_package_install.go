package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

// InstallPluginPackage installs a plugin from a local package directory.
//
// It always re-runs dry-run (scan/checksum/risk/manifest validate/install preview) server-side,
// and it never executes plugin code / SQL / frontend assets.
//
// The installed plugin status is always disabled.
func (s *Service) InstallPluginPackage(req domain.PluginPackageInstallRequest) (domain.PluginPackageInstallResponse, error) {
	path := strings.TrimSpace(req.Path)
	if path == "" {
		return domain.PluginPackageInstallResponse{}, domain.NewPluginError("plugin_package_path_invalid", "缺少插件包路径").
			WithStatus(400).
			WithSuggestion("请提供 path，例如 storage/plugins/packages/demo_notice。")
	}

	// Always re-run dry-run on backend; do not trust previous UI results.
	dry, err := s.DryRunPluginPackage(path)
	if err != nil {
		return domain.PluginPackageInstallResponse{}, err
	}

	// Reject duplicate install by code; suggest upgrade.
	code := strings.TrimSpace(dry.Package.Code)
	if code != "" {
		if _, ok := s.repo.PluginByCode(code); ok {
			return domain.PluginPackageInstallResponse{}, domain.NewPluginError("plugin_package_already_installed", "同编码插件已安装，不能重复安装").
				WithStatus(400).
				WithDetail("plugin_code", code).
				WithSuggestion("请使用升级流程（upgrade）更新插件版本。")
		}
	}

	// Hard blocks.
	if strings.ToLower(dry.Status) == "blocked" || strings.ToLower(dry.RiskReport.Level) == "blocked" {
		code := "plugin_package_install_blocked"
		switch strings.TrimSpace(dry.BlockedCode) {
		case "plugin_package_dangerous_file":
			code = "plugin_package_dangerous_file"
		case "plugin_package_checksum_invalid", "plugin_package_checksum_mismatch", "plugin_package_checksum_file_missing", "plugin_package_checksum_duplicate_path", "plugin_package_checksum_unsupported_algorithm":
			code = strings.TrimSpace(dry.BlockedCode)
		case "plugin_package_manifest_invalid", "plugin_package_manifest_missing":
			code = strings.TrimSpace(dry.BlockedCode)
		}
		return domain.PluginPackageInstallResponse{}, domain.NewPluginError(code, "插件包风险校验未通过，禁止安装").
			WithStatus(400).
			WithDetail("path", dry.Package.Path).
			WithDetail("risk_level", dry.RiskReport.Level).
			WithDetail("blocked_code", dry.BlockedCode).
			WithDetail("blocked_reasons", dry.BlockedReasons).
			WithSuggestion("请先根据风险报告修复阻断项，再重新扫描并安装。")
	}

	// checksums.json exists => checksum must be ok.
	if dry.Package.ChecksumFound {
		if strings.ToLower(dry.Checksum.Status) != "ok" {
			code := "plugin_package_checksum_mismatch"
			if strings.ToLower(dry.Checksum.Status) == "missing" {
				code = "plugin_package_checksum_missing"
			}
			return domain.PluginPackageInstallResponse{}, domain.NewPluginError(code, "插件包 checksum 校验未通过，禁止安装").
				WithStatus(400).
				WithDetail("path", dry.Package.Path).
				WithDetail("checksum_status", dry.Checksum.Status).
				WithSuggestion("请修复 checksums.json 或移除被篡改文件后重试。")
		}
	}

	// manifest validation + install preview must pass.
	if !dry.ManifestValidation.Valid || !dry.InstallDryRun.Valid {
		if strings.ToLower(strings.TrimSpace(dry.InstallDryRun.Compatibility.Status)) == pluginregistry.CompatibilityIncompatible {
			return domain.PluginPackageInstallResponse{}, domain.NewPluginError("plugin_package_core_incompatible", "插件要求的 Core 版本不兼容，无法安装").
				WithStatus(400).
				WithDetail("path", dry.Package.Path).
				WithDetail("core_version", dry.InstallDryRun.Compatibility.CoreVersion).
				WithDetail("min_core_version", dry.InstallDryRun.Compatibility.MinCoreVersion).
				WithDetail("compatible_core_version", dry.InstallDryRun.Compatibility.CompatibleCoreVersion).
				WithDetail("messages", dry.InstallDryRun.Compatibility.Messages).
				WithSuggestion("请升级 Core 或选择兼容的插件包版本后重试。")
		}
		requiredDeps := []string{}
		for _, dep := range dry.InstallDryRun.Dependencies {
			if dep.Required && strings.ToLower(dep.Status) != pluginregistry.DependencySatisfied {
				requiredDeps = append(requiredDeps, dep.Code)
			}
		}
		if len(requiredDeps) > 0 {
			return domain.PluginPackageInstallResponse{}, domain.NewPluginError("plugin_package_dependency_missing", "插件包 required 依赖未满足，无法安装").
				WithStatus(400).
				WithDetail("path", dry.Package.Path).
				WithDetail("dependencies", requiredDeps).
				WithSuggestion("请先安装并启用 required 依赖插件后重试。")
		}
		return domain.PluginPackageInstallResponse{}, domain.NewPluginError("plugin_package_manifest_invalid", "插件包 manifest 校验未通过，无法安装").
			WithStatus(400).
			WithDetail("path", dry.Package.Path).
			WithDetail("errors", append([]string(nil), dry.ManifestValidation.Errors...)).
			WithSuggestion("请修复 manifest.json 后重试。")
	}

	// Require explicit confirmation when risk level is not low.
	actualRisk := strings.ToLower(strings.TrimSpace(dry.RiskReport.Level))
	confirm := strings.ToLower(strings.TrimSpace(req.ConfirmRiskLevel))
	if actualRisk != "" && actualRisk != "low" {
		if confirm == "" || confirm != actualRisk {
			return domain.PluginPackageInstallResponse{}, domain.NewPluginError("plugin_package_install_blocked", "插件包存在风险项，需要确认风险等级后才能安装").
				WithStatus(400).
				WithDetail("path", dry.Package.Path).
				WithDetail("risk_level", actualRisk).
				WithSuggestion(fmt.Sprintf("请在确认弹窗中选择 confirm_risk_level=%s 后重试。", actualRisk))
		}
	}

	if code == "" {
		return domain.PluginPackageInstallResponse{}, domain.NewPluginError("plugin_package_manifest_invalid", "插件包 manifest 缺少 code，无法安装").
			WithStatus(400).
			WithDetail("path", dry.Package.Path).
			WithSuggestion("请在 manifest.json 中补充 code 后重试。")
	}

	// Load raw manifest.json and install via existing manifest install logic.
	abs, _, nerr := pluginregistry.NormalizePluginPackagePath(path)
	if nerr != nil {
		return domain.PluginPackageInstallResponse{}, nerr
	}
	manifestRaw, rerr := os.ReadFile(filepath.Join(abs, "manifest.json"))
	if rerr != nil {
		return domain.PluginPackageInstallResponse{}, fmt.Errorf("读取 manifest.json 失败：%w", rerr)
	}

	packageManifestChecksum := findChecksumSHA256(dry.Checksum.Matched, "manifest.json")
	plugin, validation, ierr := s.installPluginManifestInternal(manifestRaw, "local_package", packageManifestChecksum)
	if ierr != nil {
		// Ensure a stable error code for UI.
		if apiErr, ok := ierr.(*domain.APIError); ok && apiErr != nil {
			return domain.PluginPackageInstallResponse{}, apiErr
		}
		return domain.PluginPackageInstallResponse{}, domain.NewPluginError("plugin_package_install_failed", "插件安装失败").
			WithStatus(500).
			WithDetail("plugin_code", code).
			WithSuggestion("请查看后台日志，修复后重试。")
	}

	createdMigrations := len(validation.NormalizedManifest.Migrations)
	resp := domain.PluginPackageInstallResponse{
		Message:   "插件已从本地插件包安装完成（默认 disabled）",
		Plugin:    plugin,
		Package:   dry.Package,
		Checksum:  dry.Checksum,
		RiskLevel: dry.RiskReport.Level,
		InstallResult: domain.PluginPackageInstallResult{
			Installed:          true,
			CreatedConfig:      true,
			CreatedMigrations:  createdMigrations,
			CreatedPermissions: len(validation.NormalizedManifest.Permissions),
			CreatedMenus:       len(validation.NormalizedManifest.Menus),
			CreatedRoutes:      len(validation.NormalizedManifest.Routes),
		},
		Warnings: dry.Warnings,
	}
	_ = validation
	return resp, nil
}

func findChecksumSHA256(files []domain.PluginPackageChecksumFile, path string) string {
	for _, it := range files {
		if strings.TrimSpace(it.Path) == path && strings.TrimSpace(it.SHA256) != "" {
			return strings.TrimSpace(it.SHA256)
		}
	}
	return ""
}
