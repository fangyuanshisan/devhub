package plugins

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"devhub-gin-backend/internal/domain"
)

const (
	pluginPackageMaxFileSize     int64 = 2 * 1024 * 1024  // 2MB
	pluginPackageMaxTotalSize    int64 = 10 * 1024 * 1024 // 10MB
	pluginPackageMaxFiles              = 100
	pluginPackageMaxManifestSize int64 = 256 * 1024
	pluginPackageMaxReadmeSize   int64 = 512 * 1024
	pluginPackageMaxConfigSize   int64 = 256 * 1024
)

var pluginPackageDangerousExt = map[string]bool{
	".sh":    true,
	".bash":  true,
	".zsh":   true,
	".ps1":   true,
	".bat":   true,
	".cmd":   true,
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
	".php":   true,
	".go":    true,
	".js":    true,
	".mjs":   true,
	".ts":    true,
	".tsx":   true,
	".jsx":   true,
	".wasm":  true,
	".lua":   true,
	".py":    true,
	".rb":    true,
	".jar":   true,
	".class": true,
	".sql":   true,
	".env":   true,
}

func pluginPackageAllowedRoots(workdir string) []string {
	roots := []string{
		filepath.Join(workdir, "examples", "plugins"),
		filepath.Join(workdir, "plugins-local"),
		filepath.Join(workdir, "storage", "plugins", "packages"),
		filepath.Join(workdir, "storage", "plugins", "exports"),
		filepath.Join(workdir, "storage", "plugins", "staging"),
		filepath.Join(workdir, "storage", "plugins", "quarantine"),
		filepath.Join(workdir, ".devhub", "plugins"),
	}
	out := make([]string, 0, len(roots))
	for _, root := range roots {
		out = append(out, filepath.Clean(root))
	}
	return out
}

func findProjectRoot(start string) (string, error) {
	dir := filepath.Clean(start)
	for i := 0; i < 12; i++ {
		if dir == "" || dir == string(filepath.Separator) {
			break
		}
		if fileExists(filepath.Join(dir, "go.mod")) && fileExists(filepath.Join(dir, "VERSION")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("unable to locate project root from %s", start)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// NormalizePluginPackagePath validates and resolves a user-supplied package path into an absolute path.
// It rejects path traversal, absolute paths outside allowlisted roots, and empty paths.
func NormalizePluginPackagePath(input string) (abs string, clean string, err error) {
	clean = filepath.Clean(strings.TrimSpace(input))
	if clean == "" || clean == "." {
		return "", "", domain.NewPluginError("plugin_package_path_invalid", "插件包路径不能为空").
			WithStatus(400).
			WithDetail("path", input).
			WithSuggestion("请提供一个允许目录内的相对路径，例如 examples/plugins/demo_notice。")
	}

	if strings.Contains(clean, "\x00") {
		return "", "", domain.NewPluginError("plugin_package_path_invalid", "插件包路径不合法").
			WithStatus(400).
			WithDetail("path", input).
			WithSuggestion("请检查路径是否包含非法字符。")
	}

	if strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." || strings.Contains(clean, string(filepath.Separator)+".."+string(filepath.Separator)) {
		return "", "", domain.NewPluginError("plugin_package_path_invalid", "不允许使用 ../ 路径穿越").
			WithStatus(400).
			WithDetail("path", input).
			WithSuggestion("请使用允许目录内的相对路径。")
	}

	workdir, werr := os.Getwd()
	if werr != nil {
		return "", "", fmt.Errorf("读取工作目录失败：%w", werr)
	}
	root, rerr := findProjectRoot(workdir)
	if rerr != nil {
		return "", "", fmt.Errorf("读取项目根目录失败：%w", rerr)
	}
	roots := pluginPackageAllowedRoots(root)

	var candidate string
	if filepath.IsAbs(clean) {
		candidate = clean
	} else {
		candidate = filepath.Join(root, clean)
	}
	abs, err = filepath.Abs(candidate)
	if err != nil {
		return "", "", domain.NewPluginError("plugin_package_path_invalid", "插件包路径不合法").
			WithStatus(400).
			WithDetail("path", input).
			WithSuggestion("请检查路径格式。")
	}
	abs = filepath.Clean(abs)

	if !isUnderAllowedRoots(abs, roots) {
		return "", "", domain.NewPluginError("plugin_package_path_invalid", "插件包路径不在允许目录内").
			WithStatus(400).
			WithDetail("path", input).
			WithDetail("allowed_roots", roots).
			WithSuggestion("请把插件包放到 examples/plugins、plugins-local、storage/plugins/packages、storage/plugins/exports、storage/plugins/staging、storage/plugins/quarantine、.devhub/plugins 目录下后重试。")
	}
	return abs, clean, nil
}

func isUnderAllowedRoots(path string, roots []string) bool {
	for _, root := range roots {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			continue
		}
		if rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..") {
			return true
		}
	}
	return false
}

// ScanPluginPackage scans files under a local plugin package directory and returns a file scan report.
// It never executes code; it only reads file metadata (size/name) and does basic path safety checks.
func ScanPluginPackage(dir string) (domain.PluginPackageFileScan, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.PluginPackageFileScan{}, domain.NewPluginError("plugin_package_not_found", "插件包目录不存在").
				WithStatus(404).
				WithDetail("path", dir).
				WithSuggestion("请检查路径是否正确，或先把插件包放到允许目录内。")
		}
		return domain.PluginPackageFileScan{}, fmt.Errorf("读取插件包目录失败：%w", err)
	}
	if !info.IsDir() {
		return domain.PluginPackageFileScan{}, domain.NewPluginError("plugin_package_path_invalid", "插件包路径必须为目录").
			WithStatus(400).
			WithDetail("path", dir).
			WithSuggestion("请提供包含 manifest.json 的插件包目录路径。")
	}

	var scan domain.PluginPackageFileScan
	scan.AllowedFiles = []domain.PluginPackageFileEntry{}
	scan.UnknownFiles = []domain.PluginPackageFileEntry{}
	scan.DangerousFiles = []domain.PluginPackageFileEntry{}

	fileCount := 0
	totalSize := int64(0)
	var tooMany bool

	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			scan.Errors = append(scan.Errors, fmt.Sprintf("读取文件失败：%s", walkErr.Error()))
			return nil
		}
		if path == dir {
			return nil
		}
		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			scan.Errors = append(scan.Errors, fmt.Sprintf("解析相对路径失败：%s", rerr.Error()))
			return nil
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "../") || rel == ".." {
			scan.Errors = append(scan.Errors, fmt.Sprintf("发现路径穿越：%s", rel))
			scan.DangerousFiles = append(scan.DangerousFiles, domain.PluginPackageFileEntry{Path: rel, Rule: "path_traversal"})
			return nil
		}

		// Never follow symlinks: they can escape the package root.
		if d.Type()&os.ModeSymlink != 0 {
			scan.DangerousFiles = append(scan.DangerousFiles, domain.PluginPackageFileEntry{Path: rel, Rule: "symlink"})
			return nil
		}
		if d.IsDir() {
			// Treat internal vendored dirs as dangerous.
			if isDangerousDir(rel) {
				scan.DangerousFiles = append(scan.DangerousFiles, domain.PluginPackageFileEntry{Path: rel + "/", Rule: "dangerous_dir"})
				return filepath.SkipDir
			}
			return nil
		}

		fi, ierr := d.Info()
		if ierr != nil {
			scan.Errors = append(scan.Errors, fmt.Sprintf("读取文件信息失败：%s", ierr.Error()))
			return nil
		}
		if isHardLink(fi) {
			scan.DangerousFiles = append(scan.DangerousFiles, domain.PluginPackageFileEntry{Path: rel, Rule: "hardlink"})
			scan.Errors = append(scan.Errors, fmt.Sprintf("发现硬链接文件（禁止）：%s", rel))
			return nil
		}

		fileCount++
		if fileCount > pluginPackageMaxFiles {
			tooMany = true
			return fs.SkipAll
		}

		size := fi.Size()
		totalSize += size
		entry := domain.PluginPackageFileEntry{Path: rel, Size: size}

		// Size limits.
		if size > pluginPackageMaxFileSize {
			entry.Rule = "file_too_large"
			scan.DangerousFiles = append(scan.DangerousFiles, entry)
			scan.Errors = append(scan.Errors, fmt.Sprintf("单文件超过限制：%s", rel))
			return nil
		}
		base := filepath.Base(rel)
		switch base {
		case "manifest.json":
			if size > pluginPackageMaxManifestSize {
				entry.Rule = "manifest_too_large"
				scan.DangerousFiles = append(scan.DangerousFiles, entry)
				scan.Errors = append(scan.Errors, "manifest.json 超过大小限制")
				return nil
			}
		case "README.md":
			if size > pluginPackageMaxReadmeSize {
				entry.Rule = "readme_too_large"
				scan.UnknownFiles = append(scan.UnknownFiles, entry)
				scan.Warnings = append(scan.Warnings, "README.md 超过推荐大小限制")
				return nil
			}
		case "config.example.json":
			if size > pluginPackageMaxConfigSize {
				entry.Rule = "config_example_too_large"
				scan.UnknownFiles = append(scan.UnknownFiles, entry)
				scan.Warnings = append(scan.Warnings, "config.example.json 超过推荐大小限制")
				return nil
			}
		}

		// File safety rules.
		category, rule := classifyPluginPackageFile(rel, fi)
		entry.Rule = rule
		switch category {
		case "dangerous":
			scan.DangerousFiles = append(scan.DangerousFiles, entry)
		case "unknown":
			scan.UnknownFiles = append(scan.UnknownFiles, entry)
		default:
			scan.AllowedFiles = append(scan.AllowedFiles, entry)
		}
		return nil
	})

	scan.TotalFiles = fileCount
	scan.TotalSize = totalSize

	if tooMany {
		scan.Errors = append(scan.Errors, fmt.Sprintf("文件数量超过限制（>%d）", pluginPackageMaxFiles))
	}
	if totalSize > pluginPackageMaxTotalSize {
		scan.Errors = append(scan.Errors, fmt.Sprintf("插件包总大小超过限制（>%d bytes）", pluginPackageMaxTotalSize))
	}
	if walkErr != nil && !errors.Is(walkErr, fs.SkipAll) {
		return scan, fmt.Errorf("扫描插件包失败：%w", walkErr)
	}
	return scan, nil
}

func isDangerousDir(rel string) bool {
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
		switch part {
		case ".git", "node_modules", "vendor", "dist", "build":
			return true
		}
	}
	return false
}

func classifyPluginPackageFile(rel string, info fs.FileInfo) (category string, rule string) {
	rel = strings.TrimPrefix(filepath.ToSlash(rel), "/")
	lower := strings.ToLower(rel)
	base := strings.ToLower(filepath.Base(rel))
	dir := strings.ToLower(filepath.Dir(rel))
	if dir != "" && dir != "." {
		for _, part := range strings.Split(filepath.ToSlash(dir), "/") {
			if strings.HasPrefix(part, ".") && part != "." && part != ".." {
				return "dangerous", "hidden_dir"
			}
		}
	}

	if strings.Contains(lower, "/.git/") || strings.HasPrefix(lower, ".git/") {
		return "dangerous", "git_dir"
	}
	if strings.Contains(lower, "/node_modules/") || strings.HasPrefix(lower, "node_modules/") {
		return "dangerous", "node_modules"
	}
	if strings.Contains(lower, "/vendor/") || strings.HasPrefix(lower, "vendor/") {
		return "dangerous", "vendor_dir"
	}

	ext := strings.ToLower(filepath.Ext(base))
	if pluginPackageDangerousExt[ext] {
		if (info.Mode() & 0111) != 0 {
			return "dangerous", "executable_file"
		}
		if lower == "001_schema.sql" {
			return "unknown", "deprecated_root_schema_sql"
		}
		if strings.HasPrefix(lower, "migrations/") && ext == ".sql" {
			return "allowed", "allowed_migration_sql"
		}
		return "dangerous", "dangerous_ext"
	}

	// Executable bit on any file is forbidden (best-effort).
	if (info.Mode() & 0111) != 0 {
		return "dangerous", "executable_file"
	}

	// Hidden executable file (best-effort).
	if strings.HasPrefix(base, ".") && (info.Mode()&0111) != 0 {
		return "dangerous", "hidden_executable"
	}

	// Allow list.
	switch base {
	case "manifest.json", "readme.md", "license", "config.example.json", "checksums.json",
		"publisher.json", "signature.json", "packaging.md", "package.example.md", "receiver.example.md":
		return "allowed", "allowed"
	}
	if strings.HasPrefix(lower, "docs/") && strings.HasSuffix(lower, ".md") {
		return "allowed", "allowed_docs"
	}
	if strings.HasPrefix(lower, "examples/") && (strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".json")) {
		return "allowed", "allowed_examples"
	}
	if strings.HasPrefix(lower, "migrations/") && strings.HasSuffix(lower, ".json") {
		return "allowed", "allowed_migrations"
	}
	if strings.HasPrefix(lower, "assets/") {
		switch ext {
		case ".png", ".jpg", ".jpeg", ".webp":
			return "allowed", "allowed_assets"
		case ".svg":
			return "unknown", "svg_not_verified"
		default:
			return "unknown", "unknown_assets"
		}
	}

	return "unknown", "unknown_file"
}

func isHardLink(info fs.FileInfo) bool {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st == nil {
		return false
	}
	return st.Nlink > 1
}
