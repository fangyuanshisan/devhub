package plugins

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
)

const (
	PluginPackageUploadMaxZipSize    int64 = 20 * 1024 * 1024
	PluginPackageUploadMaxTotalSize  int64 = 50 * 1024 * 1024
	PluginPackageUploadMaxFileSize   int64 = 5 * 1024 * 1024
	PluginPackageUploadMaxFiles            = 300
	PluginPackageUploadMaxDepth            = 8
	PluginPackageUploadMaxNameLength       = 240
)

type ZipExtractResult struct {
	ZipScan       domain.PluginPackageZipScan
	PackageRelDir string
	ManifestCount int
}

var nestedArchiveExts = map[string]bool{
	".zip": true,
	".tar": true,
	".gz":  true,
	".tgz": true,
	".rar": true,
	".7z":  true,
	".bz2": true,
	".xz":  true,
}

func ExtractPluginPackageZip(zipPath string, stagingAbs string, stagingClean string) (ZipExtractResult, error) {
	info, err := os.Stat(zipPath)
	if err != nil {
		return ZipExtractResult{}, fmt.Errorf("读取上传 zip 失败：%w", err)
	}
	if info.Size() > PluginPackageUploadMaxZipSize {
		return ZipExtractResult{}, domain.NewPluginError("plugin_package_upload_too_large", "上传 zip 超过大小限制").
			WithStatus(400).
			WithDetail("max_bytes", PluginPackageUploadMaxZipSize).
			WithSuggestion("请将 zip 控制在 20MB 以内。")
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return ZipExtractResult{}, domain.NewPluginError("plugin_package_zip_invalid", "zip 文件无法解析").
			WithStatus(400).
			WithSuggestion("请确认上传的是合法 .zip 插件包。")
	}
	defer reader.Close()

	scan := domain.PluginPackageZipScan{}
	seen := map[string]bool{}
	manifestPaths := []string{}
	filesWritten := 0
	totalUncompressed := int64(0)

	stagingAbs = filepath.Clean(stagingAbs)
	for _, item := range reader.File {
		scan.TotalEntries++
		scan.CompressedSize += int64(item.CompressedSize64)
		scan.UncompressedSize += int64(item.UncompressedSize64)

		entryName, err := normalizeZipEntryName(item.Name)
		if err != nil {
			return ZipExtractResult{ZipScan: scan}, err
		}
		if seen[entryName] {
			return ZipExtractResult{ZipScan: scan}, domain.NewPluginError("plugin_package_zip_bomb_detected", "zip 包含重复 entry，禁止解压").
				WithStatus(400).
				WithDetail("entry", entryName).
				WithSuggestion("请移除重复路径后重新上传。")
		}
		seen[entryName] = true

		if item.FileInfo().Mode()&os.ModeSymlink != 0 {
			return ZipExtractResult{ZipScan: scan}, domain.NewPluginError("plugin_package_zip_symlink_forbidden", "zip 包含 symlink，禁止解压").
				WithStatus(400).
				WithDetail("entry", entryName).
				WithSuggestion("请移除软链接后重新打包。")
		}
		if !item.FileInfo().Mode().IsRegular() && !item.FileInfo().IsDir() {
			return ZipExtractResult{ZipScan: scan}, domain.NewPluginError("plugin_package_zip_entry_path_invalid", "zip 包含非常规文件，禁止解压").
				WithStatus(400).
				WithDetail("entry", entryName).
				WithSuggestion("插件包 zip 只能包含普通文件和目录。")
		}
		if strings.Count(entryName, "/")+1 > PluginPackageUploadMaxDepth {
			return ZipExtractResult{ZipScan: scan}, domain.NewPluginError("plugin_package_zip_bomb_detected", "zip 目录层级过深，禁止解压").
				WithStatus(400).
				WithDetail("entry", entryName).
				WithDetail("max_depth", PluginPackageUploadMaxDepth).
				WithSuggestion("请减少目录层级后重新上传。")
		}

		if item.FileInfo().IsDir() {
			if err := safeMkdirAll(stagingAbs, entryName); err != nil {
				return ZipExtractResult{ZipScan: scan}, err
			}
			continue
		}

		filesWritten++
		if filesWritten > PluginPackageUploadMaxFiles {
			return ZipExtractResult{ZipScan: scan}, domain.NewPluginError("plugin_package_zip_too_many_files", "zip 解压文件数量超过限制").
				WithStatus(400).
				WithDetail("max_files", PluginPackageUploadMaxFiles).
				WithSuggestion("请减少插件包文件数量后重新上传。")
		}
		if int64(item.UncompressedSize64) > PluginPackageUploadMaxFileSize {
			return ZipExtractResult{ZipScan: scan}, domain.NewPluginError("plugin_package_zip_file_too_large", "zip 内单文件超过限制").
				WithStatus(400).
				WithDetail("entry", entryName).
				WithDetail("max_bytes", PluginPackageUploadMaxFileSize).
				WithSuggestion("请拆分或移除过大的文件后重新上传。")
		}
		totalUncompressed += int64(item.UncompressedSize64)
		if totalUncompressed > PluginPackageUploadMaxTotalSize {
			return ZipExtractResult{ZipScan: scan}, domain.NewPluginError("plugin_package_zip_total_size_exceeded", "zip 解压后总大小超过限制").
				WithStatus(400).
				WithDetail("max_bytes", PluginPackageUploadMaxTotalSize).
				WithSuggestion("请减小插件包体积后重新上传。")
		}
		if nestedArchiveExts[strings.ToLower(filepath.Ext(entryName))] {
			return ZipExtractResult{ZipScan: scan}, domain.NewPluginError("plugin_package_zip_nested_archive_forbidden", "zip 内嵌压缩包被禁止").
				WithStatus(400).
				WithDetail("entry", entryName).
				WithSuggestion("请移除内嵌 zip/tar/rar/7z 等压缩包后重新上传。")
		}
		if path.Base(entryName) == "manifest.json" {
			manifestPaths = append(manifestPaths, entryName)
		}
		if err := writeZipFile(stagingAbs, item, entryName); err != nil {
			return ZipExtractResult{ZipScan: scan}, err
		}
	}
	scan.TotalEntries = filesWritten
	scan.UncompressedSize = totalUncompressed

	packageRel, err := identifyZipPluginPackageRoot(manifestPaths)
	if err != nil {
		return ZipExtractResult{ZipScan: scan, ManifestCount: len(manifestPaths)}, err
	}
	if _, err := os.Stat(filepath.Join(stagingAbs, filepath.FromSlash(packageRel), "manifest.json")); err != nil {
		return ZipExtractResult{ZipScan: scan, ManifestCount: len(manifestPaths)}, domain.NewPluginError("plugin_package_zip_manifest_missing", "解压后未找到 manifest.json").
			WithStatus(400).
			WithSuggestion("请确认 zip 根目录或单一顶层目录内包含 manifest.json。")
	}
	return ZipExtractResult{
		ZipScan:       scan,
		PackageRelDir: filepath.ToSlash(filepath.Join(stagingClean, packageRel)),
		ManifestCount: len(manifestPaths),
	}, nil
}

func normalizeZipEntryName(raw string) (string, error) {
	name := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if name == "" || name == "." {
		return "", domain.NewPluginError("plugin_package_zip_entry_path_invalid", "zip entry 路径为空").
			WithStatus(400).
			WithSuggestion("请重新打包插件 zip。")
	}
	if len(name) > PluginPackageUploadMaxNameLength {
		return "", domain.NewPluginError("plugin_package_zip_entry_path_invalid", "zip entry 路径过长").
			WithStatus(400).
			WithDetail("entry", name).
			WithSuggestion("请缩短文件名或目录名后重新上传。")
	}
	lower := strings.ToLower(name)
	if strings.Contains(name, "\x00") || strings.Contains(name, "..") || strings.HasPrefix(name, "/") || filepath.IsAbs(name) || hasWindowsDrivePrefix(name) {
		return "", domain.NewPluginError("plugin_package_zip_slip_detected", "检测到 zip 路径穿越风险").
			WithStatus(400).
			WithDetail("entry", name).
			WithSuggestion("请移除 ../、绝对路径或 Windows 盘符路径后重新打包。")
	}
	if strings.HasPrefix(lower, "file:") || strings.Contains(lower, "%2e") || strings.Contains(lower, "%2f") || strings.Contains(lower, "%5c") {
		return "", domain.NewPluginError("plugin_package_zip_entry_path_invalid", "zip entry 路径包含可疑编码").
			WithStatus(400).
			WithDetail("entry", name).
			WithSuggestion("请使用普通相对路径重新打包。")
	}
	clean := path.Clean(name)
	clean = strings.TrimPrefix(clean, "./")
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." || strings.HasPrefix(clean, "/") {
		return "", domain.NewPluginError("plugin_package_zip_slip_detected", "检测到 zip slip 路径穿越").
			WithStatus(400).
			WithDetail("entry", name).
			WithSuggestion("请使用包内相对路径。")
	}
	return clean, nil
}

func hasWindowsDrivePrefix(name string) bool {
	if len(name) < 3 {
		return false
	}
	first := name[0]
	return ((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) && name[1] == ':' && (name[2] == '/' || name[2] == '\\')
}

func safeMkdirAll(rootAbs string, rel string) error {
	target := filepath.Clean(filepath.Join(rootAbs, filepath.FromSlash(rel)))
	if !isPathUnderRoot(rootAbs, target) {
		return domain.NewPluginError("plugin_package_zip_slip_detected", "解压路径逃逸 staging 目录").
			WithStatus(400).
			WithDetail("entry", rel).
			WithSuggestion("请移除路径穿越 entry 后重新上传。")
	}
	return os.MkdirAll(target, 0o755)
}

func writeZipFile(rootAbs string, item *zip.File, rel string) error {
	target := filepath.Clean(filepath.Join(rootAbs, filepath.FromSlash(rel)))
	if !isPathUnderRoot(rootAbs, target) {
		return domain.NewPluginError("plugin_package_zip_slip_detected", "解压路径逃逸 staging 目录").
			WithStatus(400).
			WithDetail("entry", rel).
			WithSuggestion("请移除路径穿越 entry 后重新上传。")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	src, err := item.Open()
	if err != nil {
		return domain.NewPluginError("plugin_package_upload_extract_failed", "读取 zip entry 失败").
			WithStatus(400).
			WithDetail("entry", rel).
			WithSuggestion("请重新打包后上传。")
	}
	defer src.Close()

	dst, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return domain.NewPluginError("plugin_package_zip_bomb_detected", "zip 包含重复输出路径").
				WithStatus(400).
				WithDetail("entry", rel).
				WithSuggestion("请移除重复路径后重新上传。")
		}
		return err
	}
	defer dst.Close()
	if _, err := io.CopyN(dst, src, PluginPackageUploadMaxFileSize+1); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if info, err := dst.Stat(); err == nil && info.Size() > PluginPackageUploadMaxFileSize {
		return domain.NewPluginError("plugin_package_zip_file_too_large", "解压后单文件超过限制").
			WithStatus(400).
			WithDetail("entry", rel).
			WithSuggestion("请移除过大的文件后重新上传。")
	}
	return nil
}

func isPathUnderRoot(rootAbs, targetAbs string) bool {
	rel, err := filepath.Rel(filepath.Clean(rootAbs), filepath.Clean(targetAbs))
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}

func identifyZipPluginPackageRoot(manifestPaths []string) (string, error) {
	if len(manifestPaths) == 0 {
		return "", domain.NewPluginError("plugin_package_zip_manifest_missing", "zip 中未找到 manifest.json").
			WithStatus(400).
			WithSuggestion("请确认 zip 根目录或单一顶层目录内包含 manifest.json。")
	}
	sort.Strings(manifestPaths)
	if len(manifestPaths) > 1 {
		return "", domain.NewPluginError("plugin_package_zip_multiple_manifests", "zip 中发现多个 manifest.json，本轮不支持批量插件包上传").
			WithStatus(400).
			WithDetail("manifests", manifestPaths).
			WithSuggestion("请每次只上传一个插件包。")
	}
	manifest := manifestPaths[0]
	if manifest == "manifest.json" {
		return ".", nil
	}
	if path.Base(manifest) == "manifest.json" && strings.Count(manifest, "/") == 1 {
		return path.Dir(manifest), nil
	}
	return "", domain.NewPluginError("plugin_package_zip_manifest_missing", "manifest.json 不在 zip 根目录或单一顶层目录内").
		WithStatus(400).
		WithDetail("manifest", manifest).
		WithSuggestion("请使用 manifest.json 位于根目录或 demo_plugin/manifest.json 的结构。")
}
