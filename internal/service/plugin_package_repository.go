package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	pluginregistry "devhub-gin-backend/internal/plugins"

	"devhub-gin-backend/internal/domain"
)

type PluginPackageRepositoryFilter struct {
	Status         string
	Keyword        string
	RiskLevel      string
	ChecksumStatus string
	ManifestValid  string
	Page           int
	PageSize       int
}

type pluginPackageRepositoryScanRoot struct {
	Abs   string
	Clean string
}

func (s *Service) ListPluginPackages(root string, filter PluginPackageRepositoryFilter) (domain.PluginPackageRepositoryListResponse, error) {
	scanRoots, err := pluginPackageRepositoryScanRoots(root)
	if err != nil {
		return domain.PluginPackageRepositoryListResponse{}, err
	}

	items := []pluginregistry.PluginPackageRepositoryScanItem{}
	for _, scanRoot := range scanRoots {
		rootItems, scanErr := pluginregistry.ScanPluginRepository(scanRoot.Abs, scanRoot.Clean)
		if scanErr != nil {
			if len(scanRoots) > 1 && pluginPackageRepositoryRootMissing(scanErr) {
				continue
			}
			return domain.PluginPackageRepositoryListResponse{}, scanErr
		}
		items = append(items, rootItems...)
	}

	statusFilter := strings.TrimSpace(strings.ToLower(filter.Status))
	if statusFilter == "" {
		statusFilter = "all"
	}
	keyword := strings.TrimSpace(strings.ToLower(filter.Keyword))
	riskFilter := strings.TrimSpace(strings.ToLower(filter.RiskLevel))
	checksumFilter := strings.TrimSpace(strings.ToLower(filter.ChecksumStatus))
	manifestValidFilter := strings.TrimSpace(strings.ToLower(filter.ManifestValid))

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	all := make([]domain.PluginPackageRepositoryListItem, 0, len(items))
	summary := domain.PluginPackageRepositorySummary{}
	promotedUploads := s.promotedPackageUploadIndex()

	for _, it := range items {
		manifestPath := filepath.Join(it.AbsPath, "manifest.json")
		readmePath := filepath.Join(it.AbsPath, "README.md")
		checksumPath := filepath.Join(it.AbsPath, "checksums.json")

		row := domain.PluginPackageRepositoryListItem{
			Path:            it.CleanPath,
			RepositoryRoot:  pluginPackageRepositoryRootForPath(it.CleanPath),
			RepositoryScope: pluginPackageRepositoryScopeForPath(it.CleanPath),
			ManifestFound:   fileExists(manifestPath),
			ReadmeFound:     fileExists(readmePath),
			ChecksumFound:   fileExists(checksumPath),
			SignatureFound:  fileExists(filepath.Join(it.AbsPath, "signature.json")),
			PublisherFound:  fileExists(filepath.Join(it.AbsPath, "publisher.json")),
			UpdatedAt:       it.UpdatedAt.Unix(),
		}
		if upload, ok := promotedUploads[filepath.ToSlash(it.CleanPath)]; ok {
			row.SourceUploadID = upload.UploadID
			row.PromotedAt = upload.UpdatedAt
		}

		if !row.ManifestFound {
			row.Status = "invalid"
			row.RiskLevel = "blocked"
			row.RiskSummary = "缺少 manifest.json"
			row.Errors = append(row.Errors, "缺少 manifest.json")
		} else {
			// Reuse existing dry-run logic for detail summary.
			res, derr := s.DryRunPluginPackage(it.CleanPath)
			if derr != nil {
				// Non-fatal: treat as invalid entry with reason.
				row.Status = "invalid"
				row.RiskLevel = "blocked"
				row.RiskSummary = "dry-run 失败"
				row.Errors = append(row.Errors, fmt.Sprintf("dry-run 失败：%s", derr.Error()))
			} else {
				row.Code = res.Package.Code
				row.Name = res.Package.Name
				row.Version = res.Package.Version
				row.TotalFiles = res.FileScan.TotalFiles
				row.TotalSize = res.FileScan.TotalSize
				row.Warnings = res.Warnings
				row.Errors = res.Errors

				row.Status = normalizeRepoStatus(res.Status, res.ManifestValidation.Valid)
				row.RiskLevel = res.RiskReport.Level
				row.RiskSummary = res.RiskReport.Summary
				row.ChecksumStatus = res.Checksum.Status
				row.Signature = &res.Signature
				mv := res.ManifestValidation.Valid
				row.ManifestValid = &mv
			}
		}
		row.IsInstalled = s.repositoryPackageInstalled(row)
		row.CanDelete = s.repositoryPackageCanDelete(row)
		if row.IsInstalled {
			row.DeleteDisabledReason = "已安装包不能直接删除，请先归档 / 软卸载插件。"
		} else if !row.CanDelete {
			row.DeleteDisabledReason = "仅 storage/plugins/packages 或 storage/plugins/drafts 下的包可删除。"
		}

		// Summary counts.
		summary.Total++
		switch row.Status {
		case "ok":
			summary.OK++
		case "warning":
			summary.Warning++
		case "blocked":
			summary.Blocked++
		default:
			summary.Invalid++
		}

		// Filters.
		if statusFilter != "all" && row.Status != statusFilter {
			continue
		}
		if riskFilter != "" && strings.ToLower(row.RiskLevel) != riskFilter {
			continue
		}
		if checksumFilter != "" && strings.ToLower(row.ChecksumStatus) != checksumFilter {
			continue
		}
		if manifestValidFilter == "true" {
			if row.ManifestValid == nil || !*row.ManifestValid {
				continue
			}
		}
		if manifestValidFilter == "false" {
			if row.ManifestValid != nil && *row.ManifestValid {
				continue
			}
		}
		if keyword != "" {
			hay := strings.ToLower(strings.Join([]string{row.Path, row.RepositoryRoot, row.RepositoryScope, row.Code, row.Name, row.Version}, " "))
			if !strings.Contains(hay, keyword) {
				continue
			}
		}

		all = append(all, row)
	}

	// Stable ordering: updated desc, then path asc.
	sort.Slice(all, func(i, j int) bool {
		if all[i].UpdatedAt == all[j].UpdatedAt {
			return all[i].Path < all[j].Path
		}
		return all[i].UpdatedAt > all[j].UpdatedAt
	})

	total := len(all)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	paged := all[start:end]

	return domain.PluginPackageRepositoryListResponse{
		Items: paged,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
		Summary: summary,
	}, nil
}

func pluginPackageRepositoryScanRoots(root string) ([]pluginPackageRepositoryScanRoot, error) {
	clean := strings.TrimSpace(root)
	if strings.EqualFold(clean, "__all__") || strings.EqualFold(clean, "all") {
		candidates := []string{
			pluginPackagePromoteRoot,
			pluginPackageTemplateRoot,
			"plugins-local",
			"examples/plugins",
			"storage/plugins/exports",
			"storage/plugins/staging",
			"storage/plugins/quarantine",
			".devhub/plugins",
		}
		out := make([]pluginPackageRepositoryScanRoot, 0, len(candidates))
		seen := map[string]bool{}
		for _, candidate := range candidates {
			abs, normalized, err := pluginregistry.NormalizePluginPackageRepositoryRoot(candidate)
			if err != nil {
				return nil, err
			}
			key := filepath.ToSlash(normalized)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, pluginPackageRepositoryScanRoot{Abs: abs, Clean: normalized})
		}
		return out, nil
	}

	abs, normalized, err := pluginregistry.NormalizePluginPackageRepositoryRoot(clean)
	if err != nil {
		return nil, err
	}
	return []pluginPackageRepositoryScanRoot{{Abs: abs, Clean: normalized}}, nil
}

func pluginPackageRepositoryRootMissing(err error) bool {
	apiErr, ok := err.(*domain.APIError)
	return ok && apiErr != nil && apiErr.Code == "plugin_package_repository_not_found"
}

func pluginPackageRepositoryRootForPath(path string) string {
	clean := strings.Trim(strings.TrimSpace(filepath.ToSlash(path)), "/")
	idx := strings.LastIndex(clean, "/")
	if idx <= 0 {
		return ""
	}
	return clean[:idx]
}

func pluginPackageRepositoryScopeForPath(path string) string {
	clean := strings.Trim(strings.TrimSpace(filepath.ToSlash(path)), "/")
	switch {
	case strings.HasPrefix(clean, pluginPackagePromoteRoot+"/"):
		return "packages"
	case strings.HasPrefix(clean, pluginPackageTemplateRoot+"/"):
		return "drafts"
	case strings.HasPrefix(clean, "plugins-local/"):
		return "plugins_local"
	case strings.HasPrefix(clean, "examples/plugins/"):
		return "examples"
	case strings.HasPrefix(clean, "storage/plugins/exports/"):
		return "exports"
	case strings.HasPrefix(clean, "storage/plugins/staging/"):
		return "staging"
	case strings.HasPrefix(clean, "storage/plugins/quarantine/"):
		return "quarantine"
	case strings.HasPrefix(clean, ".devhub/plugins/"):
		return "devhub"
	default:
		return "custom"
	}
}

func (s *Service) PromotePluginPackageDraft(path string, force bool) (domain.PluginPackagePromoteResponse, error) {
	sourceAbs, sourceClean, err := pluginregistry.NormalizePluginPackagePath(path)
	if err != nil {
		return domain.PluginPackagePromoteResponse{}, err
	}
	sourceClean = filepath.ToSlash(sourceClean)
	if !strings.HasPrefix(sourceClean, pluginPackageTemplateRoot+"/") {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_draft_promote_source_invalid", "仅草稿插件包允许转入正式仓库").
			WithStatus(400).
			WithDetail("path", sourceClean).
			WithSuggestion("请使用 storage/plugins/drafts/{code} 下的草稿包，或通过上传 zip promote。")
	}

	dry, err := s.DryRunPluginPackage(sourceClean)
	if err != nil {
		return domain.PluginPackagePromoteResponse{}, err
	}
	if dry.Status == "blocked" || dry.RiskReport.Level == "blocked" || dry.Checksum.Status == "failed" || !dry.ManifestValidation.Valid {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_draft_promote_blocked", "草稿包预检未通过，禁止转入正式仓库").
			WithStatus(400).
			WithDetail("path", sourceClean).
			WithDetail("status", dry.Status).
			WithDetail("risk_level", dry.RiskReport.Level).
			WithSuggestion("请先修复 blocked 风险、checksum 或 manifest 错误后重新预检。")
	}
	code := strings.TrimSpace(dry.Package.Code)
	if code == "" {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_draft_promote_blocked", "草稿包缺少 manifest.code，禁止转入正式仓库").
			WithStatus(400).
			WithSuggestion("请修复 manifest.json 后重试。")
	}

	root, err := serviceProjectRoot()
	if err != nil {
		return domain.PluginPackagePromoteResponse{}, err
	}
	targetClean := filepath.ToSlash(filepath.Join(pluginPackagePromoteRoot, code))
	targetAbs := filepath.Join(root, filepath.FromSlash(targetClean))
	if _, statErr := os.Stat(targetAbs); statErr == nil && !force {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_target_exists", "本地插件仓库目标目录已存在").
			WithStatus(409).
			WithDetail("path", targetClean).
			WithSuggestion("请删除或更名已有正式仓库目录后重试；当前默认不覆盖。")
	}
	if force {
		if err := os.RemoveAll(targetAbs); err != nil {
			return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_failed", "清理已有目标目录失败").
				WithStatus(500).
				WithDetail("path", targetClean)
		}
	}
	if err := copyPackageTree(sourceAbs, targetAbs); err != nil {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_failed", "草稿包转入正式仓库失败").
			WithStatus(500).
			WithDetail("path", targetClean).
			WithSuggestion("请检查 storage/plugins/packages 权限和磁盘空间。")
	}
	dry, err = s.DryRunPluginPackage(targetClean)
	if err != nil {
		_ = os.RemoveAll(targetAbs)
		return domain.PluginPackagePromoteResponse{}, err
	}
	if dry.Status == "blocked" || dry.Checksum.Status == "failed" || !dry.ManifestValidation.Valid {
		_ = os.RemoveAll(targetAbs)
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_draft_promote_blocked", "转入后复检未通过，已回滚正式仓库目录").
			WithStatus(400).
			WithDetail("status", dry.Status).
			WithSuggestion("请修复草稿包后重新转入。")
	}
	return domain.PluginPackagePromoteResponse{
		Message:     "草稿包已转入正式本地仓库；仍未安装插件",
		SourcePath:  sourceClean,
		PackagePath: targetClean,
		Status:      dry.Status,
		DryRun:      dry,
		Warnings:    dry.Warnings,
	}, nil
}

func (s *Service) DeletePluginPackageRepositoryPackage(path string) (domain.PluginPackageCleanupResponse, error) {
	item, err := s.repositoryCleanupItem(path, true)
	if err != nil {
		return domain.PluginPackageCleanupResponse{}, err
	}
	if !item.CanDelete {
		return domain.PluginPackageCleanupResponse{}, domain.NewPluginError(firstNonEmpty(item.BlockedCode, "plugin_package_repository_delete_not_allowed"), "本地插件包不能删除").
			WithStatus(400).
			WithDetail("path", path).
			WithDetail("plugin_code", item.PluginCode).
			WithSuggestion(firstNonEmpty(item.Reason, "仅 promoted 且未安装的本地仓库包可删除。"))
	}
	if err := removeSafePluginPackagePath(item.Path); err != nil {
		return domain.PluginPackageCleanupResponse{}, err
	}
	return domain.PluginPackageCleanupResponse{DryRun: false, WillDeleteCount: 1, WillFreeBytes: item.Bytes, DeletedCount: 1, FreedBytes: item.Bytes, Items: []domain.PluginPackageCleanupItem{item}}, nil
}

func (s *Service) CleanupPluginPackageRepository(req domain.PluginPackageCleanupRequest) (domain.PluginPackageCleanupResponse, error) {
	return s.cleanupPluginPackageRepository(req)
}

func (s *Service) PreviewPluginPackageRepositoryCleanup(req domain.PluginPackageCleanupRequest) (domain.PluginPackageCleanupResponse, error) {
	req.DryRun = true
	req.ConfirmToken = ""
	return s.cleanupPluginPackageRepository(req)
}

func (s *Service) cleanupPluginPackageRepository(req domain.PluginPackageCleanupRequest) (domain.PluginPackageCleanupResponse, error) {
	rows := []domain.PluginPackageRepositoryListItem{}
	for page := 1; ; page++ {
		resp, err := s.ListPluginPackages(pluginPackagePromoteRoot, PluginPackageRepositoryFilter{Status: "all", Page: page, PageSize: 100})
		if err != nil {
			return domain.PluginPackageCleanupResponse{}, err
		}
		rows = append(rows, resp.Items...)
		if len(rows) >= resp.Pagination.Total || len(resp.Items) == 0 {
			break
		}
	}
	rules := normalizePluginPackageCleanupRules(req)
	now := timeNow()
	out := domain.PluginPackageCleanupResponse{
		DryRun:          req.DryRun,
		ConfirmRequired: true,
		StatusCounts:    map[string]int{},
		SkippedCounts:   map[string]int{},
	}
	for _, row := range rows {
		if !cleanupRepoOlderThanAllowed(row.UpdatedAt, req.OlderThanDays, now) {
			continue
		}
		item, err := s.repositoryCleanupItem(row.Path, true)
		if err != nil {
			out.Errors = append(out.Errors, err.Error())
			continue
		}
		item.Status = firstNonEmpty(item.Status, row.Status)
		item.PluginCode = firstNonEmpty(item.PluginCode, row.Code)
		item.Name = firstNonEmpty(item.Name, row.Name)
		item.Version = firstNonEmpty(item.Version, row.Version)
		item.FileCount = safePathFileCount(item.Path)

		canMatch := cleanupRepositoryRowMatches(row, item, rules)
		if !canMatch && item.CanDelete {
			continue
		}
		if canMatch && !item.CanDelete {
			out.Items = append(out.Items, item)
			out.SkippedCount++
			out.SkippedCounts[firstNonEmpty(item.BlockedCode, "skipped")]++
			continue
		}
		if !canMatch {
			continue
		}
		if active, reason := s.repositoryPackageHasActiveTask(item.PluginCode); active {
			item.CanDelete = false
			item.BlockedCode = "plugin_package_cleanup_active_task"
			item.Reason = reason
			out.Items = append(out.Items, item)
			out.SkippedCount++
			out.SkippedCounts[item.BlockedCode]++
			continue
		}
		out.Items = append(out.Items, item)
		out.WillDeleteCount++
		out.WillFreeBytes += item.Bytes
		out.WillDeleteFiles += item.FileCount
		out.StatusCounts[firstNonEmpty(item.Status, "unknown")]++
	}
	if req.DryRun {
		out.ConfirmToken = s.signPluginPackageCleanupToken("plugin_package_repository", req, out.Items)
		return out, nil
	}
	if err := s.verifyPluginPackageCleanupToken("plugin_package_repository", req, out.Items); err != nil {
		return domain.PluginPackageCleanupResponse{}, err
	}
	for _, item := range out.Items {
		if !item.CanDelete {
			continue
		}
		latest, err := s.repositoryCleanupItem(item.Path, true)
		if err != nil {
			out.Errors = append(out.Errors, err.Error())
			out.FailedCount++
			continue
		}
		if !cleanupRepositoryItemMatches(latest, rules) {
			continue
		}
		if !latest.CanDelete {
			if latest.BlockedCode == "plugin_package_repository_not_found" {
				out.Warnings = append(out.Warnings, fmt.Sprintf("%s: 本地仓库包已不存在，按幂等清理跳过。", item.Path))
				continue
			}
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %s", item.Path, latest.Reason))
			out.SkippedCount++
			out.SkippedCounts[firstNonEmpty(latest.BlockedCode, "skipped")]++
			continue
		}
		if active, reason := s.repositoryPackageHasActiveTask(latest.PluginCode); active {
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %s", item.Path, reason))
			out.SkippedCount++
			out.SkippedCounts["plugin_package_cleanup_active_task"]++
			continue
		}
		if err := removeSafePluginPackagePath(latest.Path); err != nil {
			out.Errors = append(out.Errors, err.Error())
			out.FailedCount++
			continue
		}
		if upload, ok := s.promotedPackageUploadIndex()[filepath.ToSlash(strings.TrimSpace(latest.Path))]; ok {
			if err := s.repo.DeletePluginPackageUpload(upload.UploadID); err != nil {
				out.Warnings = append(out.Warnings, fmt.Sprintf("%s: 本地仓库目录已删除，但上传记录删除失败：%v", latest.Path, err))
				out.FailedCount++
			}
		}
		out.DeletedCount++
		out.FreedBytes += latest.Bytes
		out.DeletedFiles += latest.FileCount
	}
	return out, nil
}

func (s *Service) repositoryCleanupItem(path string, _ bool) (domain.PluginPackageCleanupItem, error) {
	abs, clean, err := pluginregistry.NormalizePluginPackagePath(path)
	if err != nil {
		return domain.PluginPackageCleanupItem{}, err
	}
	clean = filepath.ToSlash(clean)
	isPromotedPackage := strings.HasPrefix(clean, pluginPackagePromoteRoot+"/")
	isDraftPackage := strings.HasPrefix(clean, pluginPackageTemplateRoot+"/")
	if !isPromotedPackage && !isDraftPackage {
		return domain.PluginPackageCleanupItem{Kind: "repository", Path: clean, CanDelete: false, Reason: "仅 storage/plugins/packages 或 storage/plugins/drafts 下的插件包可删除。", BlockedCode: "plugin_package_repository_path_forbidden"}, nil
	}
	dry, err := s.DryRunPluginPackage(clean)
	item := domain.PluginPackageCleanupItem{Kind: "repository", Path: clean, Bytes: safePathSize(clean), FileCount: safePathFileCount(clean)}
	if err == nil {
		item.Status = normalizeRepoStatus(dry.Status, dry.ManifestValidation.Valid)
		item.PluginCode = strings.TrimSpace(dry.Package.Code)
		item.Name = strings.TrimSpace(dry.Package.Name)
		item.Version = strings.TrimSpace(dry.Package.Version)
	} else {
		item.Status = "invalid"
		item.Reason = err.Error()
	}
	if item.PluginCode == "" {
		item.PluginCode = filepath.Base(clean)
	}
	if isPromotedPackage {
		if plugin, installed := s.repositoryPackageBinding(domain.PluginPackageRepositoryListItem{Path: clean, Code: item.PluginCode, Version: item.Version}); installed {
			item.CanDelete = false
			if strings.TrimSpace(strings.ToLower(plugin.Status)) == "enabled" || strings.TrimSpace(strings.ToLower(plugin.Status)) == "running" {
				item.BlockedCode = "plugin_package_repository_enabled"
				item.Reason = "已启用插件当前使用包不能清理。"
			} else {
				item.BlockedCode = "plugin_package_repository_installed"
				item.Reason = "已安装包不能直接删除，请先归档 / 软卸载插件。"
			}
			return item, nil
		}
	}
	if info, statErr := os.Stat(abs); statErr != nil || !info.IsDir() {
		item.CanDelete = false
		item.BlockedCode = "plugin_package_repository_not_found"
		item.Reason = "本地仓库包不存在。"
		return item, nil
	}
	item.CanDelete = true
	return item, nil
}

type pluginPackageCleanupRules struct {
	Scope        string
	Statuses     map[string]bool
	Prefixes     []string
	TestOnly     bool
	AnyUninstall bool
}

func normalizePluginPackageCleanupRules(req domain.PluginPackageCleanupRequest) pluginPackageCleanupRules {
	scope := normalizePluginPackageCleanupScope(req.Scope)
	prefixes := normalizedCleanupPrefixes(req.Prefixes)
	statuses := normalizedCleanupStatuses(req.Statuses)
	if len(statuses) == 0 {
		statuses = defaultCleanupStatuses(scope, req)
	}
	statusSet := map[string]bool{}
	for _, status := range statuses {
		statusSet[status] = true
	}
	return pluginPackageCleanupRules{
		Scope:        scope,
		Statuses:     statusSet,
		Prefixes:     prefixes,
		TestOnly:     scope == "test_packages" || scope == "expired_dry_runs",
		AnyUninstall: scope == "uninstalled",
	}
}

func normalizePluginPackageCleanupScope(scope string) string {
	switch strings.TrimSpace(strings.ToLower(scope)) {
	case "test_packages", "uninstalled", "blocked_invalid", "expired_dry_runs":
		return strings.TrimSpace(strings.ToLower(scope))
	default:
		return "blocked_invalid"
	}
}

func defaultCleanupStatuses(scope string, req domain.PluginPackageCleanupRequest) []string {
	switch scope {
	case "test_packages":
		return []string{"ok", "warning", "blocked", "invalid", "uploaded", "prechecked", "promoted", "dry_run_required", "dry_run_failed", "install_failed"}
	case "uninstalled":
		return []string{"ok", "warning", "blocked", "invalid", "uploaded", "prechecked", "promoted", "dry_run_required", "dry_run_failed", "install_failed"}
	case "expired_dry_runs":
		return []string{"warning", "blocked", "invalid", "dry_run_failed", "dry_run_required"}
	default:
		out := []string{}
		if req.IncludeBlocked || len(req.Statuses) == 0 {
			out = append(out, "blocked")
		}
		if req.IncludeInvalid || len(req.Statuses) == 0 {
			out = append(out, "invalid")
		}
		if req.IncludeWarningUninstalled {
			out = append(out, "warning")
		}
		if req.IncludePromotedUninstalled {
			out = append(out, "ok")
		}
		if req.IncludeDryRunFailed {
			out = append(out, "dry_run_failed", "dry_run_required")
		}
		return out
	}
}

func normalizedCleanupPrefixes(prefixes []string) []string {
	if len(prefixes) == 0 {
		prefixes = []string{"e2e_", "fixture_", "test_", "demo_"}
	}
	out := []string{}
	seen := map[string]bool{}
	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(strings.ToLower(prefix))
		if prefix == "" || seen[prefix] {
			continue
		}
		seen[prefix] = true
		out = append(out, prefix)
	}
	sort.Strings(out)
	return out
}

func cleanupRepositoryRowMatches(row domain.PluginPackageRepositoryListItem, item domain.PluginPackageCleanupItem, rules pluginPackageCleanupRules) bool {
	if row.IsInstalled || !item.CanDelete {
		return rules.Statuses[strings.ToLower(strings.TrimSpace(item.Status))] || rules.AnyUninstall || cleanupLooksLikeTestPackage(item, rules.Prefixes)
	}
	return cleanupRepositoryItemMatches(item, rules)
}

func cleanupRepositoryItemMatches(item domain.PluginPackageCleanupItem, rules pluginPackageCleanupRules) bool {
	status := strings.ToLower(strings.TrimSpace(item.Status))
	statusMatch := rules.Statuses[status] || (status == "ok" && rules.Statuses["promoted"]) || (status == "warning" && rules.Statuses["prechecked"])
	if rules.AnyUninstall {
		statusMatch = true
	}
	testMatch := cleanupLooksLikeTestPackage(item, rules.Prefixes)
	if rules.TestOnly {
		return testMatch && statusMatch
	}
	return statusMatch || testMatch
}

func cleanupLooksLikeTestPackage(item domain.PluginPackageCleanupItem, prefixes []string) bool {
	hay := strings.ToLower(strings.Join([]string{item.PluginCode, item.Name, item.Path}, " "))
	for _, prefix := range prefixes {
		if prefix != "" && strings.HasPrefix(strings.ToLower(item.PluginCode), prefix) {
			return true
		}
	}
	for _, marker := range []string{"e2e_upload_", "e2e_upload_promote_", "e2e_upload_lifecycle_", "fixture_"} {
		if strings.Contains(hay, marker) {
			return true
		}
	}
	return false
}

func (s *Service) repositoryPackageHasActiveTask(code string) (bool, string) {
	code = strings.TrimSpace(code)
	if code == "" {
		return false, ""
	}
	enableTasks, _, _ := s.repo.PluginEnableTasks(domain.PluginEnableTaskFilter{PluginCode: code, Page: 1, PageSize: 100})
	for _, task := range enableTasks {
		switch strings.TrimSpace(task.Status) {
		case domain.PluginEnableTaskStatusPending, domain.PluginEnableTaskStatusEnabling:
			return true, "存在正在执行或待执行的启用任务，已跳过。"
		}
	}
	upgradeTasks, _, _ := s.repo.PluginUpgradeTasks(domain.PluginUpgradeTaskFilter{PluginCode: code, Page: 1, PageSize: 100})
	for _, task := range upgradeTasks {
		switch strings.TrimSpace(task.Status) {
		case domain.PluginUpgradeTaskStatusPending, domain.PluginUpgradeTaskStatusAnalyzing, domain.PluginUpgradeTaskStatusUpgrading:
			return true, "存在正在执行或待执行的升级任务，已跳过。"
		}
	}
	uninstallTasks, _, _ := s.repo.PluginUninstallTasks(domain.PluginUninstallTaskFilter{PluginCode: code, Page: 1, PageSize: 100})
	for _, task := range uninstallTasks {
		switch strings.TrimSpace(task.Status) {
		case domain.PluginUninstallTaskStatusPending, domain.PluginUninstallTaskStatusUninstalling:
			return true, "存在正在执行或待执行的卸载任务，已跳过。"
		}
	}
	return false, ""
}

func (s *Service) repositoryPackageCanDelete(row domain.PluginPackageRepositoryListItem) bool {
	path := filepath.ToSlash(strings.TrimSpace(row.Path))
	if strings.HasPrefix(path, pluginPackageTemplateRoot+"/") {
		return true
	}
	if !strings.HasPrefix(path, pluginPackagePromoteRoot+"/") || row.IsInstalled {
		return false
	}
	return true
}

func (s *Service) repositoryPackageInstalled(row domain.PluginPackageRepositoryListItem) bool {
	path := filepath.ToSlash(strings.TrimSpace(row.Path))
	if !strings.HasPrefix(path, pluginPackagePromoteRoot+"/") {
		return false
	}
	_, ok := s.repositoryPackageBinding(row)
	return ok
}

func (s *Service) repositoryPackageBinding(row domain.PluginPackageRepositoryListItem) (domain.Plugin, bool) {
	code := strings.TrimSpace(row.Code)
	if code == "" {
		return domain.Plugin{}, false
	}
	plugin, ok := s.repo.PluginByCode(code)
	if !ok {
		return domain.Plugin{}, false
	}
	if strings.TrimSpace(row.Version) != "" && strings.TrimSpace(plugin.Version) != strings.TrimSpace(row.Version) {
		return domain.Plugin{}, false
	}
	status := strings.TrimSpace(strings.ToLower(plugin.Status))
	return plugin, status != ""
}

func (s *Service) repositoryPackagePromoted(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	for _, record := range s.promotedPackageUploadIndex() {
		if filepath.ToSlash(strings.TrimSpace(record.PromotedPath)) == path {
			return true
		}
	}
	return false
}

func cleanupRepoOlderThanAllowed(updatedAt int64, days int, now time.Time) bool {
	if days <= 0 || updatedAt <= 0 {
		return true
	}
	return now.Sub(time.Unix(updatedAt, 0)) >= time.Duration(days)*24*time.Hour
}

func (s *Service) promotedPackageUploadIndex() map[string]domain.PluginPackageUploadRecord {
	out := map[string]domain.PluginPackageUploadRecord{}
	for page := 1; ; page++ {
		rows, total, err := s.repo.PluginPackageUploads(domain.PluginPackageUploadFilter{
			Status:   domain.PluginPackageUploadStatusPromoted,
			Page:     page,
			PageSize: 100,
		})
		if err != nil {
			return nil
		}
		for _, row := range rows {
			path := filepath.ToSlash(strings.TrimSpace(row.PromotedPath))
			if path == "" {
				continue
			}
			out[path] = row
		}
		if len(out) >= total || len(rows) == 0 {
			break
		}
	}
	return out
}

func (s *Service) GetPluginPackageDetail(path string) (domain.PluginPackageDryRunResult, error) {
	// Reuse dry-run end-to-end validation; it already enforces allowlist and never writes.
	res, err := s.DryRunPluginPackage(path)
	if err == nil {
		return res, nil
	}
	if apiErr, ok := err.(*domain.APIError); ok && apiErr != nil {
		if apiErr.Code == "plugin_package_not_found" {
			return domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_package_detail_not_found", "插件包详情不存在").
				WithStatus(404).
				WithDetail("path", path).
				WithSuggestion("请检查 path 是否正确，或先扫描仓库确认该包存在。")
		}
	}
	return domain.PluginPackageDryRunResult{}, err
}

func normalizeRepoStatus(status string, manifestValid bool) string {
	v := strings.TrimSpace(strings.ToLower(status))
	if !manifestValid {
		return "invalid"
	}
	switch v {
	case "blocked":
		return "blocked"
	case "warning":
		return "warning"
	default:
		return "ok"
	}
}
