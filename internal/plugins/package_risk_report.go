package plugins

import (
	"fmt"
	"strings"

	"devhub-gin-backend/internal/domain"
)

// BuildPluginPackageRiskReport evaluates a dry-run result and produces a risk report.
//
// Risk levels: low | medium | high | blocked.
// It is derived from backend scan/check results; frontend must not fake it.
func BuildPluginPackageRiskReport(
	info domain.PluginPackageInfo,
	scan domain.PluginPackageFileScan,
	checksum domain.PluginPackageChecksumResult,
	signature domain.PluginPackageSignatureResult,
	manifestValidation domain.PluginPackageManifestValidation,
	installDryRun domain.PluginManifestValidationResult,
	blockedCode string,
) domain.PluginPackageRiskReport {
	items := make([]domain.PluginPackageRiskItem, 0, 8)

	add := func(code, level, path, message, suggestion string) {
		items = append(items, domain.PluginPackageRiskItem{
			Code:       strings.TrimSpace(code),
			Level:      strings.TrimSpace(level),
			Path:       strings.TrimSpace(path),
			Message:    strings.TrimSpace(message),
			Suggestion: strings.TrimSpace(suggestion),
		})
	}

	// Hard-blocking conditions.
	if !manifestValidation.Valid {
		add("plugin_package_manifest_invalid", "blocked", "manifest.json", "manifest 校验失败", "请修复 manifest.json 后重新 dry-run。")
	}
	for _, f := range scan.DangerousFiles {
		code, msg, sug := riskFromDangerousFileEntry(f)
		add(code, "blocked", f.Path, msg, sug)
	}
	for _, msg := range scan.Errors {
		trim := strings.TrimSpace(msg)
		if trim == "" {
			continue
		}
		add("plugin_package_dry_run_blocked", "blocked", "", trim, "请移除危险文件或修复超限/路径问题后重新 dry-run。")
	}

	switch checksum.Status {
	case "failed":
		add("plugin_package_checksum_mismatch", "blocked", "checksums.json", "checksum 校验失败", "请重新生成 checksums.json 或修复文件内容后重试。")
	}
	for _, e := range checksum.Errors {
		if strings.TrimSpace(e) == "" {
			continue
		}
		add("plugin_package_checksum_invalid", "blocked", "checksums.json", e, "请修复 checksums.json 后重试。")
	}

	// High/medium warnings.
	if strings.TrimSpace(blockedCode) != "" {
		add(blockedCode, "blocked", "", "dry-run 被阻断", "请根据阻断原因修复后重试。")
	}

	if checksum.Status == "missing" {
		add("plugin_package_checksum_missing", "medium", "checksums.json", "未提供 checksums.json（无法校验文件完整性）", "建议补充 checksums.json 并使用 sha256。")
	}
	if checksum.Status == "warning" && len(checksum.Extra) > 0 {
		add("plugin_package_file_not_covered", "high", "", fmt.Sprintf("存在 %d 个未被 checksum 覆盖的文件", len(checksum.Extra)), "建议将所有文件加入 checksums.json，避免导入前后内容不一致。")
	}

	if len(scan.UnknownFiles) > 0 {
		add("plugin_package_unknown_files", "medium", "", fmt.Sprintf("存在 %d 个未知文件（unknown_files）", len(scan.UnknownFiles)), "请确认这些文件是否必要；若不需要建议移除。")
	}

	// Signature / publisher signals.
	if !signature.SignatureFound {
		add("plugin_package_signature_missing", "medium", "signature.json", "未提供 signature.json（未签名）", "建议为插件包生成签名并配置 trusted_publishers。")
	} else {
		switch strings.TrimSpace(strings.ToLower(signature.VerificationStatus)) {
		case "unsupported":
			add("plugin_package_signature_unsupported_algorithm", "blocked", "signature.json", "签名算法不受支持，阻断导入预览", "请使用 ed25519 重新生成签名。")
		case "failed":
			add("plugin_package_signature_verification_failed", "blocked", "signature.json", "签名验签失败，可能文件已被篡改", "请修复文件与签名后重试。")
		case "publisher_unknown":
			add("plugin_package_signature_publisher_unknown", "high", "signature.json", "缺少可信或包内公钥，无法完成真实验签", "请在后台添加 trusted publisher，或补齐 publisher.json 公钥后重试。")
		}
	}
	if !signature.PublisherFound {
		add("plugin_package_publisher_missing", "medium", "publisher.json", "未提供 publisher.json（发布者信息缺失）", "建议补充 publisher.json 并在 trusted_publishers 中配置可信发布者。")
	}
	switch strings.TrimSpace(strings.ToLower(signature.TrustStatus)) {
	case "unknown":
		if signature.SignatureFound && strings.TrimSpace(strings.ToLower(signature.VerificationStatus)) == "verified" {
			add("plugin_package_signature_unknown_publisher", "high", "", "插件包签名发布者未在本地可信来源中", "确认 publisher 后加入 trusted_publishers，或仅在测试环境使用。")
		}
	case "blocked":
		add("plugin_package_signature_publisher_blocked", "blocked", "", "插件包发布者被本地策略 blocked", "请移除该插件包或调整 trusted publishers 配置。")
	case "revoked":
		add("plugin_package_signature_publisher_revoked", "blocked", "", "插件包发布者被本地策略 revoked", "请更换可信发布者签名或移除该插件包。")
	}
	if len(signature.UnsignedFiles) > 0 && signature.SignatureFound {
		add("plugin_package_signature_missing", "high", "", fmt.Sprintf("存在 %d 个文件未被签名覆盖", len(signature.UnsignedFiles)), "建议将所有 checksums.json 覆盖文件纳入签名范围。")
	}

	// Dependency / compatibility signals from existing dry-run logic.
	for _, dep := range installDryRun.Dependencies {
		if dep.Satisfied {
			continue
		}
		if dep.Required {
			add("plugin_dependency_missing", "blocked", dep.Code, "required 依赖未满足", "请先安装并启用依赖插件后再导入。")
		} else {
			add("plugin_dependency_missing", "high", dep.Code, "optional 依赖未满足", "建议补齐依赖插件，或确认该依赖确实可选。")
		}
	}
	if strings.TrimSpace(installDryRun.Compatibility.Status) == "warning" {
		add("plugin_core_version_incompatible", "high", "", "Core 版本兼容存在警告", "请检查 min_core_version / compatible_core_version。")
	}
	if strings.TrimSpace(installDryRun.Compatibility.Status) == "blocked" {
		add("plugin_core_version_incompatible", "blocked", "", "Core 版本不兼容，阻断导入预览", "请修复 Core 版本约束后重试。")
	}

	level := "low"
	for _, item := range items {
		level = maxRiskLevel(level, item.Level)
	}
	score := riskScore(level)
	summary := riskSummary(level, items)
	return domain.PluginPackageRiskReport{
		Level:   level,
		Score:   score,
		Summary: summary,
		Items:   items,
	}
}

func riskFromDangerousFileEntry(f domain.PluginPackageFileEntry) (code, message, suggestion string) {
	switch strings.TrimSpace(f.Rule) {
	case "symlink":
		return "plugin_package_symlink_forbidden", "发现软链接文件（禁止）", "请移除软链接文件，改为普通文件。"
	case "hardlink":
		return "plugin_package_dangerous_file", "发现硬链接文件（禁止）", "请移除硬链接文件，改为普通文件。"
	case "path_traversal":
		return "plugin_package_path_invalid", "发现路径穿越（禁止）", "请移除 ../ 或异常路径后重试。"
	case "executable_file", "hidden_executable":
		return "plugin_package_dangerous_file", "发现可执行权限文件（禁止）", "请移除可执行文件或去除可执行权限后重试。"
	case "dangerous_dir", "git_dir", "node_modules", "vendor_dir":
		return "plugin_package_dangerous_file", "发现禁止目录（禁止）", "请移除 node_modules/.git/vendor 等目录后重试。"
	case "file_too_large", "manifest_too_large":
		return "plugin_package_size_limit_exceeded", "文件大小超过限制（禁止）", "请缩小文件体积或移除大文件后重试。"
	case "dangerous_ext":
		return "plugin_package_dangerous_file", "发现禁止扩展名文件（禁止）", "请移除脚本/代码/SQL 等文件后重试。"
	default:
		return "plugin_package_dangerous_file", "发现危险文件（禁止）", "请移除危险文件后重试。"
	}
}

func maxRiskLevel(a, b string) string {
	return []string{a, b}[boolToInt(riskLevelRank(b) > riskLevelRank(a))]
}

func riskLevelRank(level string) int {
	switch strings.TrimSpace(strings.ToLower(level)) {
	case "blocked":
		return 3
	case "high":
		return 2
	case "medium":
		return 1
	default:
		return 0
	}
}

func riskScore(level string) int {
	switch strings.TrimSpace(strings.ToLower(level)) {
	case "blocked":
		return 95
	case "high":
		return 70
	case "medium":
		return 40
	default:
		return 0
	}
}

func riskSummary(level string, items []domain.PluginPackageRiskItem) string {
	switch strings.TrimSpace(strings.ToLower(level)) {
	case "blocked":
		return "发现阻断项（危险文件 / checksum / manifest 等），禁止导入预览"
	case "high":
		return "存在高风险项（依赖/兼容/完整性警告），建议修复后再继续"
	case "medium":
		return "存在告警项（未知文件 / 缺少说明或 checksum），建议补齐"
	default:
		if len(items) > 0 {
			return "通过校验，但仍有轻微提示"
		}
		return "通过校验"
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
