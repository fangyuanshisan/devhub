package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

type checksumsWire struct {
	Algorithm string `json:"algorithm"`
	Files     []struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"files"`
}

// VerifyPluginPackageChecksums verifies checksums.json (if present) and returns a checksum result.
//
// - Missing checksums.json returns Status=missing (warning, not blocked).
// - Invalid JSON, unsupported algorithm, duplicate paths, referenced file missing, mismatch => blocked via returned APIError.
func VerifyPluginPackageChecksums(packageDir string, scan domain.PluginPackageFileScan) (domain.PluginPackageChecksumResult, error) {
	checksumPath := filepath.Join(packageDir, "checksums.json")
	result := domain.PluginPackageChecksumResult{
		Algorithm:  "sha256",
		Status:     "missing",
		Matched:    []domain.PluginPackageChecksumFile{},
		Mismatched: []domain.PluginPackageChecksumMismatch{},
		Missing:    []string{},
		Extra:      []string{},
		Warnings:   []string{},
		Errors:     []string{},
	}
	if _, err := os.Stat(checksumPath); err != nil {
		result.Warnings = append(result.Warnings, "未找到 checksums.json（建议提供以便校验文件完整性）")
		return result, nil
	}

	raw, err := os.ReadFile(checksumPath)
	if err != nil {
		return result, fmt.Errorf("读取 checksums.json 失败：%w", err)
	}
	var wire checksumsWire
	if err := json.Unmarshal(raw, &wire); err != nil {
		return result, domain.NewPluginError("plugin_package_checksum_invalid", "checksums.json 不是合法 JSON").
			WithStatus(400).
			WithDetail("path", "checksums.json").
			WithDetail("reason", strings.TrimSpace(err.Error())).
			WithSuggestion("请修复 checksums.json 后重试。")
	}
	algo := strings.TrimSpace(strings.ToLower(wire.Algorithm))
	if algo == "" {
		algo = "sha256"
	}
	if algo != "sha256" {
		return result, domain.NewPluginError("plugin_package_checksum_unsupported_algorithm", "不支持的 checksum algorithm").
			WithStatus(400).
			WithDetail("algorithm", wire.Algorithm).
			WithSuggestion("当前仅支持 sha256。")
	}
	result.Algorithm = algo
	result.Status = "ok"

	declared := map[string]string{}
	order := []string{}
	for _, f := range wire.Files {
		p := filepath.ToSlash(filepath.Clean(strings.TrimSpace(f.Path)))
		if p == "" || p == "." {
			continue
		}
		if strings.HasPrefix(p, "/") || filepath.IsAbs(p) {
			return result, domain.NewPluginError("plugin_package_checksum_invalid", "checksums.json path 不允许为绝对路径").
				WithStatus(400).
				WithDetail("path", f.Path).
				WithSuggestion("请使用插件包内的相对路径。")
		}
		if strings.HasPrefix(p, "../") || p == ".." || strings.Contains(p, "/../") {
			return result, domain.NewPluginError("plugin_package_checksum_invalid", "checksums.json path 不允许包含 ..").
				WithStatus(400).
				WithDetail("path", f.Path).
				WithSuggestion("请移除路径穿越段。")
		}
		if declared[p] != "" {
			return result, domain.NewPluginError("plugin_package_checksum_duplicate_path", "checksums.json 存在重复 path").
				WithStatus(400).
				WithDetail("path", p).
				WithSuggestion("请在 checksums.json 中移除重复项。")
		}
		sum := strings.TrimSpace(strings.ToLower(f.SHA256))
		if sum == "" {
			return result, domain.NewPluginError("plugin_package_checksum_invalid", "checksums.json 缺少 sha256 值").
				WithStatus(400).
				WithDetail("path", p).
				WithSuggestion("请补齐 sha256 值。")
		}
		declared[p] = sum
		order = append(order, p)
	}
	sort.Strings(order)

	// Verify referenced files exist and match.
	for _, p := range order {
		full := filepath.Join(packageDir, filepath.FromSlash(p))
		info, err := os.Lstat(full)
		if err != nil {
			return result, domain.NewPluginError("plugin_package_checksum_file_missing", "checksums.json 声明的文件不存在").
				WithStatus(400).
				WithDetail("path", p).
				WithSuggestion("请修复 checksums.json 的 files 列表，或补齐缺失文件。")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return result, domain.NewPluginError("plugin_package_symlink_forbidden", "checksums.json 引用的文件是软链接，禁止").
				WithStatus(400).
				WithDetail("path", p).
				WithSuggestion("请移除软链接文件，改为普通文件。")
		}
		raw, err := os.ReadFile(full)
		if err != nil {
			return result, fmt.Errorf("读取文件失败：%s: %w", p, err)
		}
		sum := sha256.Sum256(raw)
		actual := hex.EncodeToString(sum[:])
		expected := declared[p]
		if actual != expected {
			result.Status = "failed"
			result.Mismatched = append(result.Mismatched, domain.PluginPackageChecksumMismatch{Path: p, Expected: expected, Actual: actual})
			continue
		}
		result.Matched = append(result.Matched, domain.PluginPackageChecksumFile{Path: p, SHA256: expected})
	}

	// Extra/uncovered files: warn only in this stage.
	all := packageFilesFromScan(scan)
	for _, p := range all {
		if p == "checksums.json" {
			continue
		}
		if _, ok := declared[p]; !ok {
			result.Extra = append(result.Extra, p)
		}
	}
	sort.Strings(result.Extra)

	if len(result.Mismatched) > 0 {
		return result, domain.NewPluginError("plugin_package_checksum_mismatch", "checksum 校验失败").
			WithStatus(400).
			WithDetail("mismatched", result.Mismatched).
			WithSuggestion("请重新生成 checksums.json 或修复文件内容后重试。")
	}

	if len(result.Extra) > 0 {
		result.Status = "warning"
		result.Warnings = append(result.Warnings, fmt.Sprintf("存在 %d 个未被 checksums.json 覆盖的文件（extra）", len(result.Extra)))
	}
	return result, nil
}

func packageFilesFromScan(scan domain.PluginPackageFileScan) []string {
	out := []string{}
	add := func(items []domain.PluginPackageFileEntry) {
		for _, item := range items {
			p := filepath.ToSlash(strings.TrimSpace(item.Path))
			if p != "" {
				out = append(out, p)
			}
		}
	}
	add(scan.AllowedFiles)
	add(scan.UnknownFiles)
	add(scan.DangerousFiles)
	sort.Strings(out)
	return uniqueStrings(out)
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
