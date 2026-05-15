package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

func (s *Service) ListPluginVersionRepository(filter domain.PluginVersionFilter) (domain.PluginVersionRepositoryListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	records := s.collectPluginVersionRecords(filter)
	grouped := map[string]*domain.PluginVersionRepositoryItem{}
	for _, record := range records {
		if !matchPluginVersionFilter(record, filter) {
			continue
		}
		item := grouped[record.PluginCode]
		if item == nil {
			item = &domain.PluginVersionRepositoryItem{PluginCode: record.PluginCode, PluginName: record.PluginName}
			grouped[record.PluginCode] = item
		}
		if item.PluginName == "" {
			item.PluginName = record.PluginName
		}
		item.Versions = append(item.Versions, record)
		item.Sources = appendUnique(item.Sources, record.Source)
		if record.IsInstalled {
			item.InstalledVersion = record.Version
		}
		if record.Source == string(domain.PluginVersionSourceLocalPackage) {
			item.LatestLocalVersion = maxVersion(item.LatestLocalVersion, record.Version)
		}
		if record.Source == string(domain.PluginVersionSourceRemoteIndex) {
			item.LatestRemoteVersion = maxVersion(item.LatestRemoteVersion, record.Version)
		}
		if record.IsUpgradeCandidate {
			item.UpdateAvailable = true
		}
		if item.RiskSummary == "" && record.RiskSummary != "" {
			item.RiskSummary = record.RiskSummary
		}
	}

	items := make([]domain.PluginVersionRepositoryItem, 0, len(grouped))
	for _, item := range grouped {
		sort.Slice(item.Versions, func(i, j int) bool {
			cmp := pluginregistry.CompareVersionStrings(item.Versions[i].Version, item.Versions[j].Version)
			if cmp == 0 {
				return item.Versions[i].Source < item.Versions[j].Source
			}
			return cmp > 0
		})
		sort.Strings(item.Sources)
		items = append(items, *item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].PluginCode < items[j].PluginCode })

	summary := domain.PluginVersionRepositorySummary{Total: len(items)}
	for _, record := range records {
		switch record.Source {
		case string(domain.PluginVersionSourceInstalled):
			summary.Installed++
		case string(domain.PluginVersionSourceLocalPackage):
			summary.LocalPackages++
		case string(domain.PluginVersionSourceUploadedPackage):
			summary.UploadedPackage++
		case string(domain.PluginVersionSourceRemoteIndex):
			summary.RemoteIndex++
		}
		if record.IsUpgradeCandidate {
			summary.UpdateAvailable++
		}
		if record.Readonly {
			summary.Readonly++
		}
	}

	total := len(items)
	start := (filter.Page - 1) * filter.PageSize
	if start > total {
		start = total
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	return domain.PluginVersionRepositoryListResponse{
		Items:      append([]domain.PluginVersionRepositoryItem(nil), items[start:end]...),
		Pagination: domain.Pagination{Page: filter.Page, PageSize: filter.PageSize, Total: total},
		Summary:    summary,
	}, nil
}

func (s *Service) PluginVersions(code string) (domain.PluginVersionDetailResponse, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.PluginVersionDetailResponse{}, domain.NewPluginError("plugin_version_not_found", "插件版本不存在").WithStatus(404)
	}
	records := s.collectPluginVersionRecords(domain.PluginVersionFilter{PluginCode: code})
	out := domain.PluginVersionDetailResponse{PluginCode: code}
	for _, record := range records {
		if record.PluginCode != code {
			continue
		}
		if out.PluginName == "" {
			out.PluginName = record.PluginName
		}
		if record.IsInstalled {
			out.InstalledVersion = record.Version
		}
		out.Versions = append(out.Versions, record)
	}
	if len(out.Versions) == 0 {
		return domain.PluginVersionDetailResponse{}, domain.NewPluginError("plugin_version_not_found", "插件版本不存在").WithStatus(404).WithDetail("plugin_code", code)
	}
	sort.Slice(out.Versions, func(i, j int) bool {
		cmp := pluginregistry.CompareVersionStrings(out.Versions[i].Version, out.Versions[j].Version)
		if cmp == 0 {
			return out.Versions[i].Source < out.Versions[j].Source
		}
		return cmp > 0
	})
	return out, nil
}

func (s *Service) PluginVersionDetail(code, version string, req domain.PluginUpgradeDiffRequest) (domain.PluginVersionRecord, error) {
	detail, err := s.PluginVersions(code)
	if err != nil {
		return domain.PluginVersionRecord{}, err
	}
	source := strings.TrimSpace(req.Source)
	packagePath := strings.TrimSpace(req.PackagePath)
	for _, record := range detail.Versions {
		if record.Version != strings.TrimSpace(version) {
			continue
		}
		if source != "" && record.Source != source {
			continue
		}
		if packagePath != "" && record.PackagePath != packagePath {
			continue
		}
		if req.RemoteIndexID > 0 && record.RemoteIndexID != req.RemoteIndexID {
			continue
		}
		return record, nil
	}
	return domain.PluginVersionRecord{}, domain.NewPluginError("plugin_version_not_found", "插件目标版本不存在").WithStatus(404).
		WithDetail("plugin_code", code).WithDetail("version", version).WithSuggestion("请先确认该版本来自本地仓库、上传包或远程只读索引。")
}

func (s *Service) PluginVersionUpgradeDiff(code, version string, req domain.PluginUpgradeDiffRequest) (domain.PluginUpgradeDiffResult, error) {
	code = strings.TrimSpace(code)
	version = strings.TrimSpace(version)
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = string(domain.PluginVersionSourceLocalPackage)
	}
	if source == string(domain.PluginVersionSourceRemoteIndex) {
		return domain.PluginUpgradeDiffResult{}, domain.NewPluginError("plugin_version_remote_readonly", "远程索引版本仅可只读展示，不能直接升级对比").
			WithStatus(409).WithDetail("plugin_code", code).WithDetail("version", version).
			WithSuggestion("请先通过受控上传 / promote 将插件包纳入本地仓库，再执行升级差异对比。")
	}
	current, ok := s.repo.PluginByCode(code)
	if !ok {
		return domain.PluginUpgradeDiffResult{}, pluginNotFound(code)
	}
	cmp := pluginregistry.CompareVersionStrings(version, current.Version)
	if cmp == 0 {
		return domain.PluginUpgradeDiffResult{}, domain.NewPluginError("plugin_version_same_version", "目标版本与当前版本相同，不能普通升级").
			WithStatus(409).WithDetail("current_version", current.Version).WithDetail("target_version", version)
	}
	if cmp < 0 {
		return domain.PluginUpgradeDiffResult{}, domain.NewPluginError("plugin_version_downgrade_forbidden", "目标版本低于当前版本，本轮不支持降级").
			WithStatus(409).WithDetail("current_version", current.Version).WithDetail("target_version", version)
	}

	manifest, packageDryRun, err := s.resolveUpgradeTargetManifest(code, version, req)
	if err != nil {
		return domain.PluginUpgradeDiffResult{}, err
	}
	if manifest.Code != code || manifest.Version != version {
		return domain.PluginUpgradeDiffResult{}, domain.NewPluginError("plugin_version_package_missing", "目标插件包与请求版本不一致").
			WithStatus(400).WithDetail("expected_code", code).WithDetail("expected_version", version).
			WithDetail("actual_code", manifest.Code).WithDetail("actual_version", manifest.Version)
	}

	sections, summary := buildPluginManifestDiff(current.PluginManifest, manifest)
	compat := pluginregistry.CheckPluginVersionCompatibility(manifest, currentCoreVersion())
	status := "ok"
	warnings := []string{}
	errorsList := []string{}
	riskItems := []domain.PluginPackageRiskItem{}
	if packageDryRun.Status != "" {
		if packageDryRun.RiskReport.Level == "blocked" || packageDryRun.Status == "blocked" {
			status = "blocked"
			riskItems = append(riskItems, domain.PluginPackageRiskItem{Code: "plugin_upgrade_diff_blocked", Level: "blocked", Message: "目标插件包风险报告已阻断升级", Suggestion: "请先修复 checksum、签名、危险文件或 manifest 错误。"})
		} else if packageDryRun.RiskReport.Level == "high" || packageDryRun.Status == "warning" {
			status = "warning"
		}
	}
	if compat.Status == pluginregistry.CompatibilityIncompatible {
		status = "blocked"
		summary.Blocked++
		riskItems = append(riskItems, domain.PluginPackageRiskItem{Code: "plugin_upgrade_target_core_incompatible", Level: "blocked", Message: "目标版本 Core 兼容性不满足", Suggestion: "请升级 DevHub Core 或选择兼容版本。"})
	}
	if summary.HighRisk > 0 && status == "ok" {
		status = "warning"
	}
	for _, section := range sections {
		for _, item := range section.Items {
			if item.RiskLevel == "high" {
				riskItems = append(riskItems, domain.PluginPackageRiskItem{Code: "plugin_upgrade_diff_high_risk", Level: "high", Path: item.Path, Message: item.Message, Suggestion: "请在审批详情中确认该变更影响。"})
			}
		}
	}
	riskLevel := "low"
	score := 5
	if status == "warning" {
		riskLevel = "high"
		score = 70
	}
	if status == "blocked" {
		riskLevel = "blocked"
		score = 100
	}
	riskReport := domain.PluginPackageRiskReport{
		Level:   riskLevel,
		Score:   score,
		Summary: upgradeDiffRiskSummary(status, summary),
		Items:   riskItems,
	}
	if status == "blocked" {
		errorsList = append(errorsList, "升级差异存在阻断项")
	} else if status == "warning" {
		warnings = append(warnings, "升级差异包含高风险变更，建议走审批确认")
	}
	return domain.PluginUpgradeDiffResult{
		PluginCode:        code,
		CurrentVersion:    current.Version,
		TargetVersion:     manifest.Version,
		Source:            source,
		Status:            status,
		Summary:           summary,
		DiffSections:      sections,
		RiskReport:        riskReport,
		Compatibility:     compat,
		Dependencies:      packageDryRun.InstallDryRun.DependencySummary,
		PackageRiskReport: packageDryRun.RiskReport,
		Warnings:          uniqueStrings(warnings),
		Errors:            uniqueStrings(errorsList),
	}, nil
}

func (s *Service) collectPluginVersionRecords(filter domain.PluginVersionFilter) []domain.PluginVersionRecord {
	installed := map[string]domain.Plugin{}
	out := []domain.PluginVersionRecord{}
	for _, plugin := range s.repo.Plugins() {
		installed[plugin.Code] = plugin
		record := domain.PluginVersionRecord{
			PluginCode:            plugin.Code,
			PluginName:            plugin.Name,
			Version:               plugin.Version,
			Source:                string(domain.PluginVersionSourceInstalled),
			Status:                "installed",
			RiskLevel:             "low",
			SignatureStatus:       firstNonBlank(plugin.PackageChecksum, plugin.ManifestChecksum),
			InstalledVersion:      plugin.Version,
			IsInstalled:           true,
			CreatedAt:             plugin.InstalledAt,
			UpdatedAt:             plugin.UpdatedAt,
			CompatibleCoreVersion: plugin.CompatibleCoreVersion,
			MinCoreVersion:        plugin.MinCoreVersion,
			CoreCompatibility:     pluginregistry.CheckPluginVersionCompatibility(plugin.PluginManifest, currentCoreVersion()),
		}
		out = append(out, record)
	}

	if repo, err := s.ListPluginPackages("storage/plugins/packages", PluginPackageRepositoryFilter{Page: 1, PageSize: 100}); err == nil {
		for _, item := range repo.Items {
			record := domain.PluginVersionRecord{
				PluginCode:       item.Code,
				PluginName:       item.Name,
				Version:          item.Version,
				Source:           string(domain.PluginVersionSourceLocalPackage),
				Status:           packageVersionStatus(item.Status),
				PackagePath:      item.Path,
				RiskLevel:        item.RiskLevel,
				RiskSummary:      item.RiskSummary,
				ChecksumStatus:   item.ChecksumStatus,
				SignatureStatus:  signatureVerificationStatus(item.Signature),
				PublisherID:      signaturePublisherID(item.Signature),
				PublicKeyID:      signaturePublicKeyID(item.Signature),
				TrustStatus:      signatureTrustStatus(item.Signature),
				UpdatedAt:        fmt.Sprintf("%d", item.UpdatedAt),
				InstalledVersion: installedVersion(installed, item.Code),
				CoreCompatibility: domain.PluginCoreCompatibility{
					CoreVersion: currentCoreVersion(),
					Status:      pluginregistry.CompatibilityCompatible,
				},
			}
			record.IsInstalled = record.InstalledVersion == record.Version && record.InstalledVersion != ""
			record.IsUpgradeCandidate = record.InstalledVersion != "" && pluginregistry.CompareVersionStrings(record.Version, record.InstalledVersion) > 0 && record.Status != "blocked"
			out = append(out, record)
		}
	}

	if uploads, _, err := s.repo.PluginPackageUploads(domain.PluginPackageUploadFilter{Page: 1, PageSize: 1000}); err == nil {
		for _, upload := range uploads {
			if upload.PackageCode == "" || upload.PackageVersion == "" {
				continue
			}
			record := domain.PluginVersionRecord{
				PluginCode:        upload.PackageCode,
				PluginName:        upload.PackageName,
				Version:           upload.PackageVersion,
				Source:            string(domain.PluginVersionSourceUploadedPackage),
				Status:            upload.Status,
				PackagePath:       firstNonBlank(upload.PromotedPath, upload.PackagePath),
				RiskLevel:         upload.RiskLevel,
				RiskSummary:       upload.ErrorMessage,
				ChecksumStatus:    upload.ChecksumStatus,
				SignatureStatus:   upload.SignatureStatus,
				PublisherID:       upload.PublisherID,
				TrustStatus:       upload.TrustStatus,
				InstalledVersion:  installedVersion(installed, upload.PackageCode),
				CreatedAt:         upload.CreatedAt,
				UpdatedAt:         upload.UpdatedAt,
				CoreCompatibility: domain.PluginCoreCompatibility{CoreVersion: currentCoreVersion(), Status: pluginregistry.CompatibilityCompatible},
			}
			record.IsInstalled = record.Status == domain.PluginPackageUploadStatusInstalled
			record.IsUpgradeCandidate = record.InstalledVersion != "" && pluginregistry.CompareVersionStrings(record.Version, record.InstalledVersion) > 0 && record.Status != domain.PluginPackageUploadStatusBlocked
			out = append(out, record)
		}
	}

	if indexes, _, err := s.repo.PluginRemoteIndexes(domain.PluginRemoteIndexFilter{Page: 1, PageSize: 1000}); err == nil {
		for _, index := range indexes {
			var payload struct {
				Document domain.PluginRemoteIndexDocument `json:"document"`
			}
			if strings.TrimSpace(index.MetadataJSON) == "" || json.Unmarshal([]byte(index.MetadataJSON), &payload) != nil {
				continue
			}
			for _, plugin := range payload.Document.Plugins {
				for _, ver := range plugin.Versions {
					compat := pluginregistry.CheckPluginVersionCompatibility(domain.PluginManifest{MinCoreVersion: ver.MinCoreVersion, CompatibleCoreVersion: ver.CompatibleCoreVersion}, currentCoreVersion())
					riskLevel, risks := s.remoteVersionRisks(ver, compat)
					record := domain.PluginVersionRecord{
						PluginCode:            plugin.Code,
						PluginName:            plugin.Name,
						Version:               ver.Version,
						Source:                string(domain.PluginVersionSourceRemoteIndex),
						Status:                "readonly",
						RemoteIndexID:         index.ID,
						RemoteSourceID:        index.SourceID,
						RiskLevel:             riskLevel,
						RiskSummary:           firstRiskSummary(risks),
						ChecksumStatus:        "metadata",
						SignatureStatus:       boolStatus(ver.SignatureURL != "", "declared", "missing"),
						PublisherID:           ver.PublisherID,
						PublicKeyID:           ver.PublicKeyID,
						TrustStatus:           s.remotePublisherTrust(ver.PublisherID, ver.PublicKeyID),
						InstalledVersion:      installedVersion(installed, plugin.Code),
						IsUpgradeCandidate:    installedVersion(installed, plugin.Code) != "" && pluginregistry.CompareVersionStrings(ver.Version, installedVersion(installed, plugin.Code)) > 0,
						Readonly:              true,
						ReadonlyMessage:       "远程索引版本仅展示元数据；不会下载、安装或执行代码。",
						CreatedAt:             ver.CreatedAt,
						UpdatedAt:             ver.UpdatedAt,
						PackageSHA256:         ver.PackageSHA256,
						SignatureURL:          ver.SignatureURL,
						CompatibleCoreVersion: ver.CompatibleCoreVersion,
						MinCoreVersion:        ver.MinCoreVersion,
						CoreCompatibility:     compat,
					}
					out = append(out, record)
				}
			}
		}
	}
	return out
}

func (s *Service) resolveUpgradeTargetManifest(code, version string, req domain.PluginUpgradeDiffRequest) (domain.PluginManifest, domain.PluginPackageDryRunResult, error) {
	source := strings.TrimSpace(req.Source)
	path := strings.TrimSpace(req.PackagePath)
	if path == "" {
		detail, err := s.PluginVersions(code)
		if err != nil {
			return domain.PluginManifest{}, domain.PluginPackageDryRunResult{}, err
		}
		for _, record := range detail.Versions {
			if record.Version == version && record.Source == source && record.PackagePath != "" {
				path = record.PackagePath
				break
			}
		}
	}
	if path == "" {
		return domain.PluginManifest{}, domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_version_package_missing", "目标版本缺少本地插件包路径").
			WithStatus(400).WithDetail("plugin_code", code).WithDetail("version", version).WithSuggestion("请选择 local_package 或已 promote 的 uploaded_package。")
	}
	dryRun, err := s.DryRunPluginPackage(path)
	if err != nil {
		return domain.PluginManifest{}, domain.PluginPackageDryRunResult{}, err
	}
	abs, _, err := pluginregistry.NormalizePluginPackagePath(path)
	if err != nil {
		return domain.PluginManifest{}, domain.PluginPackageDryRunResult{}, err
	}
	raw, err := os.ReadFile(filepath.Join(abs, "manifest.json"))
	if err != nil {
		return domain.PluginManifest{}, domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_version_package_missing", "目标插件包 manifest.json 不存在或不可读").WithStatus(404)
	}
	manifest, _, err := pluginregistry.DecodePluginManifestJSON(raw)
	if err != nil {
		return domain.PluginManifest{}, domain.PluginPackageDryRunResult{}, domain.NewPluginError("plugin_export_manifest_invalid", "目标插件包 manifest 无法解析").WithStatus(400)
	}
	return manifest, dryRun, nil
}

func buildPluginManifestDiff(current, target domain.PluginManifest) ([]domain.PluginManifestDiffSection, domain.PluginUpgradeDiffSummary) {
	builder := manifestDiffBuilder{}
	builder.compareObjectFields("basic", "基础信息", map[string]any{
		"name": current.Name, "version": current.Version, "description": current.Description, "author": current.Author, "license": current.License,
		"min_core_version": current.MinCoreVersion, "compatible_core_version": current.CompatibleCoreVersion,
	}, map[string]any{
		"name": target.Name, "version": target.Version, "description": target.Description, "author": target.Author, "license": target.License,
		"min_core_version": target.MinCoreVersion, "compatible_core_version": target.CompatibleCoreVersion,
	})
	builder.compareStringSet("content_types", "内容类型", current.ContentTypes, target.ContentTypes, "high")
	builder.compareKeyed("content_type_definitions", "内容类型定义", keyContentTypeDefs(current.ContentTypeDefs), keyContentTypeDefs(target.ContentTypeDefs), highRiskRemovedOrChanged)
	builder.compareKeyed("permissions", "权限变化", keyPermissions(current.Permissions), keyPermissions(target.Permissions), permissionRisk)
	builder.compareKeyed("menus", "菜单变化", keyMenus(current.Menus), keyMenus(target.Menus), highRiskRemoved)
	builder.compareKeyed("routes", "路由变化", keyRoutes(current.Routes), keyRoutes(target.Routes), highRiskRemoved)
	builder.compareConfigSchema(current.ConfigSchema, target.ConfigSchema)
	builder.compareAnyMap("default_config", "默认配置", pluginManifestDefaultConfig(current), pluginManifestDefaultConfig(target), "low")
	builder.compareKeyed("dependencies", "依赖变化", keyDependencies(current.Dependencies), keyDependencies(target.Dependencies), dependencyRisk)
	builder.compareKeyed("hooks", "Hook 变化", keyHooks(current.Hooks), keyHooks(target.Hooks), hookRisk)
	builder.compareKeyed("migrations", "迁移声明", keyMigrations(current.Migrations), keyMigrations(target.Migrations), migrationRisk)
	return builder.sections(), builder.summary
}

type manifestDiffBuilder struct {
	bySection map[string]*domain.PluginManifestDiffSection
	order     []string
	summary   domain.PluginUpgradeDiffSummary
}

func (b *manifestDiffBuilder) add(section, title, path, typ, risk string, before, after any, message string) {
	if b.bySection == nil {
		b.bySection = map[string]*domain.PluginManifestDiffSection{}
	}
	sec := b.bySection[section]
	if sec == nil {
		sec = &domain.PluginManifestDiffSection{Section: section, Title: title, RiskLevel: "low"}
		b.bySection[section] = sec
		b.order = append(b.order, section)
	}
	if risk == "" {
		risk = "low"
	}
	sec.Items = append(sec.Items, domain.PluginManifestDiffItem{Section: section, Path: path, Type: typ, RiskLevel: risk, Before: redactSensitiveValue(path, before), After: redactSensitiveValue(path, after), Message: message})
	sec.RiskLevel = maxRisk(sec.RiskLevel, risk)
	switch typ {
	case "added":
		b.summary.Added++
	case "removed":
		b.summary.Removed++
	default:
		b.summary.Changed++
	}
	if risk == "high" {
		b.summary.HighRisk++
	}
	if risk == "blocked" {
		b.summary.Blocked++
	}
}

func (b *manifestDiffBuilder) sections() []domain.PluginManifestDiffSection {
	out := make([]domain.PluginManifestDiffSection, 0, len(b.order))
	for _, key := range b.order {
		section := b.bySection[key]
		sort.Slice(section.Items, func(i, j int) bool { return section.Items[i].Path < section.Items[j].Path })
		out = append(out, *section)
	}
	return out
}

func (b *manifestDiffBuilder) compareObjectFields(section, title string, before, after map[string]any) {
	keys := sortedUnionKeys(before, after)
	for _, key := range keys {
		if jsonEqual(before[key], after[key]) {
			continue
		}
		b.add(section, title, section+"."+key, "changed", "low", before[key], after[key], "基础信息发生变化")
	}
}

func (b *manifestDiffBuilder) compareStringSet(section, title string, before, after []string, removeRisk string) {
	beforeSet := stringSet(before)
	afterSet := stringSet(after)
	for _, key := range sortedUnionStringSet(beforeSet, afterSet) {
		_, hasBefore := beforeSet[key]
		_, hasAfter := afterSet[key]
		switch {
		case !hasBefore && hasAfter:
			b.add(section, title, section+"."+key, "added", "low", nil, key, "新增内容类型")
		case hasBefore && !hasAfter:
			b.add(section, title, section+"."+key, "removed", removeRisk, key, nil, "删除内容类型可能影响历史内容治理")
		}
	}
}

func (b *manifestDiffBuilder) compareKeyed(section, title string, before, after map[string]any, riskFn func(string, string, any, any) (string, string)) {
	for _, key := range sortedUnionKeys(before, after) {
		_, hasBefore := before[key]
		_, hasAfter := after[key]
		typ := "changed"
		if !hasBefore {
			typ = "added"
		}
		if !hasAfter {
			typ = "removed"
		}
		if typ == "changed" && jsonEqual(before[key], after[key]) {
			continue
		}
		risk, msg := riskFn(key, typ, before[key], after[key])
		b.add(section, title, section+"."+key, typ, risk, before[key], after[key], msg)
	}
}

func (b *manifestDiffBuilder) compareAnyMap(section, title string, before, after map[string]any, risk string) {
	for _, key := range sortedUnionKeys(before, after) {
		_, hasBefore := before[key]
		_, hasAfter := after[key]
		typ := "changed"
		if !hasBefore {
			typ = "added"
		}
		if !hasAfter {
			typ = "removed"
		}
		if typ == "changed" && jsonEqual(before[key], after[key]) {
			continue
		}
		b.add(section, title, section+"."+key, typ, risk, before[key], after[key], "默认配置发生变化")
	}
}

func (b *manifestDiffBuilder) compareConfigSchema(before, after any) {
	beforeMap := schemaProperties(before)
	afterMap := schemaProperties(after)
	for _, key := range sortedUnionKeys(beforeMap, afterMap) {
		_, hasBefore := beforeMap[key]
		_, hasAfter := afterMap[key]
		typ := "changed"
		risk := "low"
		msg := "配置 schema 字段发生变化"
		if !hasBefore {
			typ = "added"
			msg = "新增配置字段"
		}
		if !hasAfter {
			typ = "removed"
			risk = "high"
			msg = "删除配置字段可能影响现有配置"
		}
		if typ == "changed" && jsonEqual(beforeMap[key], afterMap[key]) {
			continue
		}
		if typ == "changed" && schemaFieldType(beforeMap[key]) != schemaFieldType(afterMap[key]) {
			risk = "high"
			msg = "配置字段类型变化可能导致现有配置不兼容"
		}
		b.add("config_schema", "配置 schema", "config_schema."+key, typ, risk, beforeMap[key], afterMap[key], msg)
	}
	for _, key := range addedRequiredFields(before, after) {
		b.add("config_schema", "配置 schema", "config_schema.required."+key, "changed", "high", false, true, "新增必填配置字段")
	}
}

func keyContentTypeDefs(items []domain.ContentTypeDefinition) map[string]any {
	out := map[string]any{}
	for _, item := range items {
		out[firstNonBlank(item.Type, item.Name)] = item
	}
	return out
}

func keyPermissions(items []domain.PermissionDefinition) map[string]any {
	out := map[string]any{}
	for _, item := range items {
		out[firstNonBlank(item.Code, item.Name)] = item
	}
	return out
}

func keyMenus(items []domain.MenuDefinition) map[string]any {
	out := map[string]any{}
	for _, item := range items {
		out[firstNonBlank(item.Code, item.Key, item.Path, item.Title)] = item
	}
	return out
}

func keyRoutes(items []domain.RouteDefinition) map[string]any {
	out := map[string]any{}
	for _, item := range items {
		out[strings.Join([]string{item.Area, item.Method, item.Path}, " ")] = item
	}
	return out
}

func keyDependencies(items []domain.PluginDependency) map[string]any {
	out := map[string]any{}
	for _, item := range items {
		out[item.Code] = item
	}
	return out
}

func keyHooks(items []domain.HookDefinition) map[string]any {
	out := map[string]any{}
	for _, item := range items {
		out[item.Name] = item
	}
	return out
}

func keyMigrations(items []domain.PluginMigrationDefinition) map[string]any {
	out := map[string]any{}
	for _, item := range items {
		out[firstNonBlank(item.MigrationName, item.MigrationVersion)] = item
	}
	return out
}

func highRiskRemoved(_ string, typ string, _, _ any) (string, string) {
	if typ == "removed" {
		return "high", "删除声明项可能影响现有入口或能力"
	}
	return "low", "声明项发生变化"
}

func highRiskRemovedOrChanged(_ string, typ string, _, _ any) (string, string) {
	if typ == "removed" || typ == "changed" {
		return "high", "内容类型定义删除或变更可能影响历史内容"
	}
	return "low", "新增内容类型定义"
}

func permissionRisk(key, typ string, _, after any) (string, string) {
	if typ == "removed" {
		return "high", "删除权限可能影响既有角色和菜单可见性"
	}
	raw, _ := json.Marshal(after)
	lower := strings.ToLower(string(raw) + " " + key)
	if typ == "added" && (strings.Contains(lower, "high") || strings.Contains(lower, "manage") || strings.Contains(lower, "admin")) {
		return "high", "新增高危插件权限"
	}
	return "low", "权限声明发生变化"
}

func dependencyRisk(_ string, typ string, before, after any) (string, string) {
	if typ == "added" {
		if dep, ok := after.(domain.PluginDependency); ok && dep.Required {
			return "high", "新增 required 依赖会阻断缺失环境升级"
		}
	}
	if typ == "changed" {
		oldDep, oldOK := before.(domain.PluginDependency)
		newDep, newOK := after.(domain.PluginDependency)
		if oldOK && newOK && (!oldDep.Required && newDep.Required || oldDep.Version != newDep.Version) {
			return "high", "依赖约束收紧或 required 状态变化"
		}
	}
	return "low", "依赖声明发生变化"
}

func hookRisk(_ string, typ string, before, after any) (string, string) {
	if typ == "changed" {
		oldHook, oldOK := before.(domain.HookDefinition)
		newHook, newOK := after.(domain.HookDefinition)
		if oldOK && newOK && !oldHook.Blocking && newHook.Blocking {
			return "high", "Hook 从 non-blocking 改为 blocking"
		}
	}
	return "low", "Hook 声明发生变化"
}

func migrationRisk(_ string, typ string, _, _ any) (string, string) {
	if typ == "added" {
		return "high", "新增 migration 声明需要升级前确认"
	}
	return "low", "迁移声明发生变化"
}

func pluginManifestDefaultConfig(manifest domain.PluginManifest) map[string]any {
	raw, _ := json.Marshal(manifest)
	var data map[string]any
	_ = json.Unmarshal(raw, &data)
	if cfg, ok := data["default_config"].(map[string]any); ok {
		return cfg
	}
	return map[string]any{}
}

func schemaProperties(schema any) map[string]any {
	raw, _ := json.Marshal(schema)
	var data map[string]any
	_ = json.Unmarshal(raw, &data)
	if props, ok := data["properties"].(map[string]any); ok {
		return props
	}
	return map[string]any{}
}

func schemaFieldType(value any) string {
	if m, ok := value.(map[string]any); ok {
		return fmt.Sprint(m["type"])
	}
	raw, _ := json.Marshal(value)
	var data map[string]any
	_ = json.Unmarshal(raw, &data)
	return fmt.Sprint(data["type"])
}

func addedRequiredFields(before, after any) []string {
	oldSet := requiredSet(before)
	newSet := requiredSet(after)
	out := []string{}
	for key := range newSet {
		if _, ok := oldSet[key]; !ok {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

func requiredSet(schema any) map[string]struct{} {
	raw, _ := json.Marshal(schema)
	var data map[string]any
	_ = json.Unmarshal(raw, &data)
	out := map[string]struct{}{}
	if arr, ok := data["required"].([]any); ok {
		for _, item := range arr {
			out[fmt.Sprint(item)] = struct{}{}
		}
	}
	return out
}

func matchPluginVersionFilter(record domain.PluginVersionRecord, filter domain.PluginVersionFilter) bool {
	if filter.PluginCode != "" && record.PluginCode != strings.TrimSpace(filter.PluginCode) {
		return false
	}
	if filter.Source != "" && record.Source != strings.TrimSpace(filter.Source) {
		return false
	}
	if filter.Status != "" && record.Status != strings.TrimSpace(filter.Status) {
		return false
	}
	if kw := strings.ToLower(strings.TrimSpace(filter.Keyword)); kw != "" {
		hay := strings.ToLower(strings.Join([]string{record.PluginCode, record.PluginName, record.Version, record.Source, record.PackagePath, record.PublisherID}, " "))
		return strings.Contains(hay, kw)
	}
	return true
}

func sortedUnionKeys(a, b map[string]any) []string {
	seen := map[string]struct{}{}
	for key := range a {
		seen[key] = struct{}{}
	}
	for key := range b {
		seen[key] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func jsonEqual(a, b any) bool {
	aa, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return string(aa) == string(bb)
}

func stringSet(items []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out[item] = struct{}{}
		}
	}
	return out
}

func sortedUnionStringSet(a, b map[string]struct{}) []string {
	seen := map[string]struct{}{}
	for key := range a {
		seen[key] = struct{}{}
	}
	for key := range b {
		seen[key] = struct{}{}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func appendUnique(items []string, item string) []string {
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}

func maxVersion(current, next string) string {
	if current == "" || pluginregistry.CompareVersionStrings(next, current) > 0 {
		return next
	}
	return current
}

func packageVersionStatus(status string) string {
	switch status {
	case "ok", "warning":
		return "available"
	case "blocked", "invalid":
		return "blocked"
	default:
		return firstNonBlank(status, "available")
	}
}

func installedVersion(installed map[string]domain.Plugin, code string) string {
	if plugin, ok := installed[code]; ok {
		return plugin.Version
	}
	return ""
}

func signatureVerificationStatus(signature *domain.PluginPackageSignatureResult) string {
	if signature == nil {
		return ""
	}
	return signature.VerificationStatus
}

func signaturePublisherID(signature *domain.PluginPackageSignatureResult) string {
	if signature == nil {
		return ""
	}
	return signature.PublisherID
}

func signaturePublicKeyID(signature *domain.PluginPackageSignatureResult) string {
	if signature == nil {
		return ""
	}
	return signature.PublicKeyID
}

func signatureTrustStatus(signature *domain.PluginPackageSignatureResult) string {
	if signature == nil {
		return ""
	}
	return signature.TrustStatus
}

func boolStatus(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}

func firstRiskSummary(items []domain.PluginPackageRiskItem) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].Message
}

func maxRisk(a, b string) string {
	rank := map[string]int{"low": 1, "medium": 2, "warning": 2, "high": 3, "blocked": 4}
	if rank[b] > rank[a] {
		return b
	}
	return a
}

func redactSensitiveValue(path string, value any) any {
	lower := strings.ToLower(path)
	for _, token := range []string{"password", "passwd", "secret", "token", "api_key", "credential", "app_secret", "aes_key", "private_key", "client_secret", "key"} {
		if strings.Contains(lower, token) {
			if value == nil {
				return nil
			}
			return "[REDACTED]"
		}
	}
	return value
}

func upgradeDiffRiskSummary(status string, summary domain.PluginUpgradeDiffSummary) string {
	if status == "blocked" {
		return "升级差异包含阻断项，禁止继续升级"
	}
	if summary.HighRisk > 0 {
		return fmt.Sprintf("升级差异包含 %d 个高风险变更", summary.HighRisk)
	}
	return "升级差异未发现高风险变更"
}
