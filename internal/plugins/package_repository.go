package plugins

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
)

const defaultPluginPackageRepositoryRoot = "storage/plugins/packages"

// NormalizePluginPackageRepositoryRoot validates and resolves a repository root path.
// It only allows roots under the project allowlist (same as plugin package allowlist).
//
// - Empty input => default repo root (storage/plugins/packages).
// - Rejects path traversal.
func NormalizePluginPackageRepositoryRoot(input string) (abs string, clean string, err error) {
	clean = strings.TrimSpace(input)
	if clean == "" {
		clean = defaultPluginPackageRepositoryRoot
	}
	abs, cleanOut, err := NormalizePluginPackagePath(clean)
	if err == nil {
		return abs, cleanOut, nil
	}
	if apiErr, ok := err.(*domain.APIError); ok && apiErr != nil {
		// For repository roots, "not under allowlist" is treated as forbidden.
		if apiErr.Code == "plugin_package_path_invalid" && strings.Contains(apiErr.Message, "不在允许目录内") {
			return "", "", domain.NewPluginError("plugin_package_repository_forbidden", "插件包仓库路径不在允许目录内").
				WithStatus(400).
				WithDetail("root", clean).
				WithDetail("allowed_roots", apiErr.Details["allowed_roots"]).
				WithSuggestion("请使用允许目录内的仓库路径，例如 storage/plugins/packages。")
		}
	}
	return "", "", err
}

type PluginPackageRepositoryScanItem struct {
	AbsPath   string
	CleanPath string
	DirName   string
	UpdatedAt time.Time
}

// ScanPluginRepository lists 1-level subdirectories under repository root.
// It never executes code; it only inspects file metadata.
func ScanPluginRepository(repoAbs string, repoClean string) ([]PluginPackageRepositoryScanItem, error) {
	st, err := os.Stat(repoAbs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, domain.NewPluginError("plugin_package_repository_not_found", "插件包仓库目录不存在").
				WithStatus(404).
				WithDetail("root", repoClean).
				WithSuggestion("请检查仓库目录路径，或先创建仓库目录。")
		}
		return nil, fmt.Errorf("读取插件包仓库目录失败：%w", err)
	}
	if !st.IsDir() {
		return nil, domain.NewPluginError("plugin_package_repository_forbidden", "插件包仓库路径必须为目录").
			WithStatus(400).
			WithDetail("root", repoClean).
			WithSuggestion("请提供允许目录内的仓库目录路径。")
	}

	entries, err := os.ReadDir(repoAbs)
	if err != nil {
		return nil, domain.NewPluginError("plugin_package_scan_failed", "插件包仓库扫描失败").
			WithStatus(500).
			WithDetail("root", repoClean).
			WithDetail("reason", strings.TrimSpace(err.Error())).
			WithSuggestion("请检查仓库目录权限或文件状态后重试。")
	}
	out := make([]PluginPackageRepositoryScanItem, 0, len(entries))
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		name := strings.TrimSpace(ent.Name())
		if name == "" {
			continue
		}
		abs := filepath.Join(repoAbs, name)
		info, err := ent.Info()
		if err != nil {
			continue
		}
		out = append(out, PluginPackageRepositoryScanItem{
			AbsPath:   abs,
			CleanPath: filepath.ToSlash(filepath.Join(repoClean, name)),
			DirName:   name,
			UpdatedAt: info.ModTime(),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].CleanPath < out[j].CleanPath
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}
