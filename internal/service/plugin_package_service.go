package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pluginregistry "devhub-gin-backend/internal/plugins"

	"devhub-gin-backend/internal/domain"
)

// DryRunPluginPackage scans a local plugin package directory, validates its files,
// loads manifest.json, and reuses existing manifest validation/dry-run logic.
//
// It never installs the plugin, never writes plugin/config/migration tables, and never executes code/SQL.
func (s *Service) DryRunPluginPackage(inputPath string) (domain.PluginPackageDryRunResult, error) {
	abs, clean, err := pluginregistry.NormalizePluginPackagePath(inputPath)
	if err != nil {
		return domain.PluginPackageDryRunResult{}, err
	}

	scan, err := pluginregistry.ScanPluginPackage(abs)
	if err != nil {
		return domain.PluginPackageDryRunResult{}, err
	}

	manifestPath := filepath.Join(abs, "manifest.json")
	readmePath := filepath.Join(abs, "README.md")
	configExamplePath := filepath.Join(abs, "config.example.json")
	checksumPath := filepath.Join(abs, "checksums.json")

	info := domain.PluginPackageInfo{
		Path:               clean,
		DirName:            filepath.Base(abs),
		ManifestFound:      fileExists(manifestPath),
		ReadmeFound:        fileExists(readmePath),
		ConfigExampleFound: fileExists(configExamplePath),
		ChecksumFound:      fileExists(checksumPath),
	}

	// If manifest is missing, return an explicit error code (blocked).
	if !info.ManifestFound {
		return domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_package_manifest_missing", "插件包缺少 manifest.json").
			WithStatus(400).
			WithDetail("path", clean).
			WithSuggestion("请在插件包根目录放置 manifest.json（唯一主声明文件）。")
	}

	manifestRaw, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		return domain.PluginPackageDryRunResult{}, fmt.Errorf("读取 manifest.json 失败：%w", readErr)
	}

	manifest, checksum, decodeErr := pluginregistry.DecodePluginManifestJSON(manifestRaw)
	if decodeErr != nil {
		return domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_package_manifest_invalid", "manifest.json 不是合法的插件声明").
			WithStatus(400).
			WithDetail("path", clean).
			WithDetail("reason", strings.TrimSpace(decodeErr.Error())).
			WithSuggestion("请修复 manifest.json 后重试。")
	}
	_ = checksum
	info.Code = manifest.Code
	info.Name = manifest.Name
	info.Version = manifest.Version

	status := "ok"
	blockedCode := ""
	blockedReasons := []string{}
	warnings := []string{}
	errorsList := []string{}

	// Directory name suggestion.
	if info.DirName != "" && info.Code != "" && info.DirName != info.Code {
		warnings = append(warnings, fmt.Sprintf("目录名建议与 manifest.code 保持一致：dir=%s code=%s", info.DirName, info.Code))
		status = "warning"
	}

	// Scan-derived status.
	if len(scan.DangerousFiles) > 0 || len(scan.Errors) > 0 {
		status = "blocked"
		blockedCode = scanBlockedCode(scan)
		if blockedCode != "" {
			blockedReasons = append(blockedReasons, blockedCode)
		}
		errorsList = append(errorsList, scan.Errors...)
	}
	if len(scan.UnknownFiles) > 0 || len(scan.Warnings) > 0 {
		if status != "blocked" {
			status = "warning"
		}
		if len(scan.UnknownFiles) > 0 {
			warnings = append(warnings, fmt.Sprintf("发现 %d 个未知文件（unknown_files）", len(scan.UnknownFiles)))
		}
		warnings = append(warnings, scan.Warnings...)
	}

	// checksums.json verification (optional, but mismatch/invalid blocks dry-run).
	checksumResult, chkErr := pluginregistry.VerifyPluginPackageChecksums(abs, scan)
	if chkErr != nil {
		if apiErr, ok := chkErr.(*domain.APIError); ok {
			status = "blocked"
			if blockedCode == "" {
				blockedCode = apiErr.Code
			}
			blockedReasons = append(blockedReasons, apiErr.Code)
			errorsList = append(errorsList, fmt.Sprintf("[%s] %s", apiErr.Code, apiErr.Message))
			if apiErr.Suggestion != "" {
				warnings = append(warnings, fmt.Sprintf("建议：%s", apiErr.Suggestion))
			}
			if checksumResult.Status == "" {
				checksumResult.Status = "failed"
			}
			if checksumResult.Algorithm == "" {
				checksumResult.Algorithm = "sha256"
			}
			checksumResult.Errors = append(checksumResult.Errors, apiErr.Message)
		} else {
			return domain.PluginPackageDryRunResult{}, chkErr
		}
	} else {
		if len(checksumResult.Warnings) > 0 {
			warnings = append(warnings, checksumResult.Warnings...)
			if status == "ok" {
				status = "warning"
			}
		}
	}

	// config.example.json sanity (optional, warning only).
	if info.ConfigExampleFound {
		raw, cfgErr := os.ReadFile(configExamplePath)
		if cfgErr != nil {
			warnings = append(warnings, fmt.Sprintf("读取 config.example.json 失败：%s", cfgErr.Error()))
			if status == "ok" {
				status = "warning"
			}
		} else {
			var v any
			if err := json.Unmarshal(raw, &v); err != nil {
				warnings = append(warnings, fmt.Sprintf("config.example.json 不是合法 JSON：%s", err.Error()))
				if status == "ok" {
					status = "warning"
				}
			}
		}
	} else {
		warnings = append(warnings, "建议提供 config.example.json（示例配置，不应包含真实 secret）")
		if status == "ok" {
			status = "warning"
		}
	}

	// README missing is a warning.
	if !info.ReadmeFound {
		warnings = append(warnings, "建议提供 README.md（插件说明）")
		if status == "ok" {
			status = "warning"
		}
	}

	validation, _ := s.ValidatePluginManifestJSON(manifestRaw)
	manifestValidation := domain.PluginPackageManifestValidation{
		Valid:    validation.Valid,
		Errors:   append([]string(nil), validation.Errors...),
		Warnings: append([]string(nil), validation.Warnings...),
	}
	if !manifestValidation.Valid {
		status = "blocked"
		blockedCode = "plugin_package_manifest_invalid"
		blockedReasons = append(blockedReasons, blockedCode)
		errorsList = append(errorsList, "manifest 校验失败")
	}

	// In v1.4, "install dry-run" shares the same payload as manifest validation result.
	installDryRun := validation

	risk := pluginregistry.BuildPluginPackageRiskReport(info, scan, checksumResult, manifestValidation, installDryRun, blockedCode)
	if risk.Level == "blocked" {
		status = "blocked"
	}
	if (risk.Level == "medium" || risk.Level == "high") && status == "ok" {
		status = "warning"
	}

	return domain.PluginPackageDryRunResult{
		Package:            info,
		FileScan:           scan,
		Checksum:           checksumResult,
		ManifestValidation: manifestValidation,
		InstallDryRun:      installDryRun,
		RiskReport:         risk,
		Status:             status,
		BlockedCode:        blockedCode,
		BlockedReasons:     uniqueStrings(blockedReasons),
		Warnings:           uniqueStrings(warnings),
		Errors:             uniqueStrings(errorsList),
	}, nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func uniqueStrings(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func scanBlockedCode(scan domain.PluginPackageFileScan) string {
	for _, item := range scan.DangerousFiles {
		switch strings.TrimSpace(item.Rule) {
		case "file_too_large", "manifest_too_large":
			return "plugin_package_file_too_large"
		}
	}
	for _, msg := range scan.Errors {
		if strings.Contains(msg, "文件数量超过限制") {
			return "plugin_package_too_many_files"
		}
		if strings.Contains(msg, "插件包总大小超过限制") {
			return "plugin_package_too_large"
		}
		if strings.Contains(msg, "单文件超过限制") {
			return "plugin_package_file_too_large"
		}
	}
	if len(scan.DangerousFiles) > 0 {
		return "plugin_package_dangerous_file"
	}
	if len(scan.Errors) > 0 {
		return "plugin_package_dry_run_blocked"
	}
	return ""
}
