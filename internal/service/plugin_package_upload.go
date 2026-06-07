package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

const (
	pluginPackageUploadsRootDefault    = "storage/plugins/uploads"
	pluginPackageStagingRootDefault    = "storage/plugins/staging"
	pluginPackageQuarantineRootDefault = "storage/plugins/quarantine"
	pluginPackagePromoteRoot           = "storage/plugins/packages"
	pluginPackageUploadTTL             = 7 * 24 * time.Hour
)

type PluginPackageUploadOperator struct {
	ID   int64
	Name string
}

func pluginPackageUploadsRoot() string {
	if v := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_UPLOADS_ROOT")); v != "" {
		return filepath.ToSlash(v)
	}
	return pluginPackageUploadsRootDefault
}

func pluginPackageStagingRoot() string {
	if v := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_STAGING_ROOT")); v != "" {
		return filepath.ToSlash(v)
	}
	return pluginPackageStagingRootDefault
}

func pluginPackageQuarantineRoot() string {
	if v := strings.TrimSpace(os.Getenv("DEVHUB_PLUGIN_PACKAGE_QUARANTINE_ROOT")); v != "" {
		return filepath.ToSlash(v)
	}
	return pluginPackageQuarantineRootDefault
}

// UploadPluginPackageZip stores a zip in uploads/, extracts it into a per-upload
// staging sandbox, and reuses package dry-run for scan/checksum/signature/risk.
func (s *Service) UploadPluginPackageZip(filename string, size int64, r io.Reader) (domain.PluginPackageUploadResult, error) {
	return s.UploadPluginPackageZipAs(PluginPackageUploadOperator{}, filename, size, r)
}

func (s *Service) UploadPluginPackageZipAs(operator PluginPackageUploadOperator, filename string, size int64, r io.Reader) (domain.PluginPackageUploadResult, error) {
	filename = sanitizeUploadFilename(filename)
	if filename == "" || !strings.EqualFold(filepath.Ext(filename), ".zip") {
		return domain.PluginPackageUploadResult{}, domain.NewPluginError("plugin_package_upload_invalid_type", "只允许上传 .zip 插件包").
			WithStatus(400).
			WithDetail("filename", filename).
			WithSuggestion("请上传扩展名为 .zip 的插件包；不支持 tar/gz/rar/7z。")
	}
	if size > pluginregistry.PluginPackageUploadMaxZipSize {
		return domain.PluginPackageUploadResult{}, domain.NewPluginError("plugin_package_upload_too_large", "上传 zip 超过大小限制").
			WithStatus(400).
			WithDetail("max_bytes", pluginregistry.PluginPackageUploadMaxZipSize).
			WithSuggestion("请将 zip 控制在 20MB 以内。")
	}

	root, err := serviceProjectRoot()
	if err != nil {
		return domain.PluginPackageUploadResult{}, err
	}
	uploadID := newPluginPackageUploadID()
	uploadsAbs := filepath.Join(root, filepath.FromSlash(pluginPackageUploadsRoot()))
	stagingAbs := filepath.Join(root, filepath.FromSlash(pluginPackageStagingRoot()), uploadID)
	stagingClean := filepath.ToSlash(filepath.Join(pluginPackageStagingRoot(), uploadID))
	uploadPathClean := filepath.ToSlash(filepath.Join(pluginPackageUploadsRoot(), uploadID+"_"+filename))
	record := domain.PluginPackageUploadRecord{
		UploadID:         uploadID,
		OriginalFilename: filename,
		UploadedBy:       operator.ID,
		UploadedByName:   strings.TrimSpace(operator.Name),
		UploadedAt:       Now(),
		Status:           domain.PluginPackageUploadStatusUploaded,
		UploadPath:       uploadPathClean,
		StagingPath:      stagingClean,
		ExpiresAt:        timeNow().Add(pluginPackageUploadTTL).Format("2006-01-02 15:04:05"),
	}
	record, _ = s.repo.AppendPluginPackageUpload(record)
	if err := os.MkdirAll(uploadsAbs, 0o755); err != nil {
		record.Status = domain.PluginPackageUploadStatusFailed
		record.ErrorCode = "plugin_package_upload_failed"
		record.ErrorMessage = "创建上传目录失败"
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadResult{}, fmt.Errorf("创建上传目录失败：%w", err)
	}
	if err := os.MkdirAll(stagingAbs, 0o755); err != nil {
		record.Status = domain.PluginPackageUploadStatusFailed
		record.ErrorCode = "plugin_package_upload_failed"
		record.ErrorMessage = "创建 staging 目录失败"
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadResult{}, fmt.Errorf("创建 staging 目录失败：%w", err)
	}

	zipAbs := filepath.Join(root, filepath.FromSlash(uploadPathClean))
	dst, err := os.OpenFile(zipAbs, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		_ = os.RemoveAll(stagingAbs)
		record.Status = domain.PluginPackageUploadStatusFailed
		record.ErrorCode = "plugin_package_upload_failed"
		record.ErrorMessage = "保存上传 zip 失败"
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadResult{}, fmt.Errorf("保存上传 zip 失败：%w", err)
	}
	_, copyErr := io.CopyN(dst, r, pluginregistry.PluginPackageUploadMaxZipSize+1)
	closeErr := dst.Close()
	if copyErr != nil && copyErr != io.EOF {
		_ = os.Remove(zipAbs)
		_ = os.RemoveAll(stagingAbs)
		record.Status = domain.PluginPackageUploadStatusFailed
		record.ErrorCode = "plugin_package_upload_failed"
		record.ErrorMessage = "保存上传 zip 失败"
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadResult{}, fmt.Errorf("保存上传 zip 失败：%w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(zipAbs)
		_ = os.RemoveAll(stagingAbs)
		record.Status = domain.PluginPackageUploadStatusFailed
		record.ErrorCode = "plugin_package_upload_failed"
		record.ErrorMessage = "保存上传 zip 失败"
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadResult{}, fmt.Errorf("保存上传 zip 失败：%w", closeErr)
	}
	if info, err := os.Stat(zipAbs); err == nil && info.Size() > pluginregistry.PluginPackageUploadMaxZipSize {
		_ = os.Remove(zipAbs)
		_ = os.RemoveAll(stagingAbs)
		record.Status = domain.PluginPackageUploadStatusFailed
		record.ErrorCode = "plugin_package_upload_too_large"
		record.ErrorMessage = "上传 zip 超过大小限制"
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadResult{}, domain.NewPluginError("plugin_package_upload_too_large", "上传 zip 超过大小限制").
			WithStatus(400).
			WithDetail("max_bytes", pluginregistry.PluginPackageUploadMaxZipSize).
			WithSuggestion("请将 zip 控制在 20MB 以内。")
	}

	extracted, err := pluginregistry.ExtractPluginPackageZip(zipAbs, stagingAbs, stagingClean)
	if err != nil {
		_ = os.RemoveAll(stagingAbs)
		record.Status = domain.PluginPackageUploadStatusBlocked
		record.ErrorCode, record.ErrorMessage = uploadErrorCodeMessage(err)
		record.ZipScanJSON = mustJSON(scrubAnyForSnapshot(extracted.ZipScan))
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadResult{}, err
	}
	record.Status = domain.PluginPackageUploadStatusScanned
	record.ZipScanJSON = mustJSON(scrubAnyForSnapshot(extracted.ZipScan))
	record.CompressedSize = extracted.ZipScan.CompressedSize
	record.UncompressedSize = extracted.ZipScan.UncompressedSize
	packagePath := extracted.PackageRelDir
	dry, dryErr := s.DryRunPluginPackage(packagePath)
	if dryErr != nil {
		_ = os.RemoveAll(stagingAbs)
		record.Status = domain.PluginPackageUploadStatusFailed
		record.ErrorCode, record.ErrorMessage = uploadErrorCodeMessage(dryErr)
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadResult{}, dryErr
	}
	if strings.EqualFold(dry.Status, "blocked") {
		quarantineClean, quarantineErr := moveUploadToQuarantine(root, uploadID)
		if quarantineErr != nil {
			record.Status = domain.PluginPackageUploadStatusFailed
			record.ErrorCode = "plugin_package_upload_blocked"
			record.ErrorMessage = quarantineErr.Error()
			_, _ = s.repo.SavePluginPackageUpload(record)
			return domain.PluginPackageUploadResult{}, quarantineErr
		}
		packagePath = filepath.ToSlash(filepath.Join(quarantineClean, strings.TrimPrefix(strings.TrimPrefix(extracted.PackageRelDir, stagingClean), "/")))
		dry, dryErr = s.DryRunPluginPackage(packagePath)
		if dryErr != nil {
			record.Status = domain.PluginPackageUploadStatusFailed
			record.ErrorCode, record.ErrorMessage = uploadErrorCodeMessage(dryErr)
			_, _ = s.repo.SavePluginPackageUpload(record)
			return domain.PluginPackageUploadResult{}, dryErr
		}
	}

	record = uploadRecordFromDryRun(record, packagePath, uploadContainerPath(packagePath), extracted.ZipScan, dry)
	record, _ = s.repo.SavePluginPackageUpload(record)
	res := uploadResultFromDryRun(uploadID, filename, packagePath, uploadContainerPath(packagePath), extracted.ZipScan, dry)
	res.Record = &record
	res.Actions = uploadActions(record)
	return res, nil
}

func (s *Service) GetPluginPackageUpload(uploadID string) (domain.PluginPackageUploadResult, error) {
	uploadID = strings.TrimSpace(uploadID)
	if uploadID == "" || !strings.HasPrefix(uploadID, "pkg_upload_") || strings.Contains(uploadID, "/") || strings.Contains(uploadID, "\\") {
		return domain.PluginPackageUploadResult{}, domain.NewPluginError("plugin_package_upload_not_found", "上传包不存在").
			WithStatus(404).
			WithSuggestion("请刷新上传记录后重试。")
	}
	record, err := s.getPluginPackageUploadRecord(uploadID)
	if err != nil {
		return domain.PluginPackageUploadResult{}, err
	}
	packagePath := strings.TrimSpace(record.PackagePath)
	containerPath := strings.TrimSpace(record.StagingPath)
	if packagePath == "" {
		var findErr error
		packagePath, containerPath, findErr = findUploadedPackagePath(uploadID)
		if findErr != nil {
			return domain.PluginPackageUploadResult{}, findErr
		}
	}
	dry, err := s.DryRunPluginPackage(packagePath)
	if err != nil {
		return domain.PluginPackageUploadResult{}, err
	}
	res := uploadResultFromDryRun(uploadID, record.OriginalFilename, packagePath, containerPath, parseZipScan(record.ZipScanJSON), dry)
	res.Status = record.Status
	res.Record = &record
	res.Actions = uploadActions(record)
	return res, nil
}

func (s *Service) PromotePluginPackageUpload(uploadID string, force bool) (domain.PluginPackagePromoteResponse, error) {
	record, err := s.getPluginPackageUploadRecord(strings.TrimSpace(uploadID))
	if err != nil {
		return domain.PluginPackagePromoteResponse{}, err
	}
	if record.Status == domain.PluginPackageUploadStatusBlocked {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_blocked", "上传包未通过校验，禁止转入本地仓库").
			WithStatus(400).
			WithDetail("upload_id", uploadID).
			WithDetail("status", record.Status).
			WithSuggestion("请先修复 blocked 风险、checksum 或 manifest 错误后重新上传。")
	}
	if record.Status == domain.PluginPackageUploadStatusPromoted {
		if strings.TrimSpace(record.PromotedPath) != "" {
			return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_target_exists", "本地插件仓库目标目录已存在").
				WithStatus(409).
				WithDetail("path", record.PromotedPath).
				WithSuggestion("该上传包已经 promote；请刷新本地插件仓库列表。")
		}
	}
	if record.Status != domain.PluginPackageUploadStatusStaged && record.Status != domain.PluginPackageUploadStatusApproved {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_upload_invalid_status", "当前上传包状态不允许转入本地仓库").
			WithStatus(400).
			WithDetail("upload_id", uploadID).
			WithDetail("status", record.Status).
			WithSuggestion("只有 staged / approved 上传包可以 promote；blocked、deleted、expired、canceled 均不可 promote。")
	}
	packagePath := strings.TrimSpace(record.PackagePath)
	if packagePath == "" {
		var container string
		packagePath, container, err = findUploadedPackagePath(record.UploadID)
		if err != nil {
			return domain.PluginPackagePromoteResponse{}, err
		}
		record.StagingPath = container
		record.PackagePath = packagePath
	}
	dry, err := s.DryRunPluginPackage(packagePath)
	if err != nil {
		return domain.PluginPackagePromoteResponse{}, err
	}
	if dry.Status == "blocked" || !canPromoteDryRun(dry, packagePath) {
		record.Status = domain.PluginPackageUploadStatusBlocked
		record.ErrorCode = "plugin_package_promote_blocked"
		record.ErrorMessage = "promote 前重新校验发现阻断项"
		record = uploadRecordFromDryRun(record, packagePath, record.StagingPath, parseZipScan(record.ZipScanJSON), dry)
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_blocked", "上传包未通过校验，禁止转入本地仓库").
			WithStatus(400).
			WithDetail("upload_id", uploadID).
			WithDetail("status", dry.Status).
			WithSuggestion("请先修复 blocked 风险、checksum 或 manifest 错误后重新上传。")
	}
	code := strings.TrimSpace(dry.Package.Code)
	if code == "" {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_blocked", "上传包缺少 manifest.code，禁止转入本地仓库").
			WithStatus(400).
			WithSuggestion("请修复 manifest.json 后重新上传。")
	}
	root, err := serviceProjectRoot()
	if err != nil {
		return domain.PluginPackagePromoteResponse{}, err
	}
	targetClean := filepath.ToSlash(filepath.Join(pluginPackagePromoteRoot, code))
	targetAbs := filepath.Join(root, pluginPackagePromoteRoot, code)
	if _, statErr := os.Stat(targetAbs); statErr == nil && !force {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_target_exists", "本地插件仓库目标目录已存在").
			WithStatus(409).
			WithDetail("path", targetClean).
			WithSuggestion("请删除或更名已有目录后重试；当前默认不覆盖。")
	}
	if force {
		if err := os.RemoveAll(targetAbs); err != nil {
			return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_failed", "清理已有目标目录失败").
				WithStatus(500).
				WithDetail("path", targetClean)
		}
	}
	packageAbs := filepath.Join(root, filepath.FromSlash(packagePath))
	if err := copyPackageTree(packageAbs, targetAbs); err != nil {
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_failed", "转入本地插件仓库失败").
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
		record.Status = domain.PluginPackageUploadStatusBlocked
		record.ErrorCode = "plugin_package_promote_blocked"
		record.ErrorMessage = "转入后复检未通过"
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackagePromoteResponse{}, domain.NewPluginError("plugin_package_promote_blocked", "转入后复检未通过，已回滚本地仓库目录").
			WithStatus(400).
			WithDetail("status", dry.Status).
			WithSuggestion("请修复上传包后重新上传。")
	}
	record.Status = domain.PluginPackageUploadStatusPromoted
	record.PromotedPath = targetClean
	record.ErrorCode = ""
	record.ErrorMessage = ""
	record = uploadRecordFromDryRun(record, packagePath, record.StagingPath, parseZipScan(record.ZipScanJSON), dry)
	record.PromotedPath = targetClean
	_, _ = s.repo.SavePluginPackageUpload(record)
	return domain.PluginPackagePromoteResponse{
		Message:     "上传包已转入本地插件仓库；仍未安装插件",
		UploadID:    uploadID,
		PackagePath: targetClean,
		Status:      dry.Status,
		DryRun:      dry,
		Warnings:    dry.Warnings,
	}, nil
}

func (s *Service) ListPluginPackageUploads(filter domain.PluginPackageUploadFilter) (domain.PluginPackageUploadListResponse, error) {
	filter.Status = normalizeUploadFilter(filter.Status)
	filter.RiskLevel = normalizeUploadFilter(filter.RiskLevel)
	filter.TrustStatus = normalizeUploadFilter(filter.TrustStatus)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	items, total, err := s.repo.PluginPackageUploads(filter)
	if err != nil {
		return domain.PluginPackageUploadListResponse{}, err
	}
	summaryItems := items
	if total > len(items) {
		summaryItems = make([]domain.PluginPackageUploadRecord, 0, total)
		summaryFilter := filter
		summaryFilter.Page = 1
		summaryFilter.PageSize = 100
		for {
			chunk, _, err := s.repo.PluginPackageUploads(summaryFilter)
			if err != nil {
				return domain.PluginPackageUploadListResponse{}, err
			}
			summaryItems = append(summaryItems, chunk...)
			if len(summaryItems) >= total || len(chunk) == 0 {
				break
			}
			summaryFilter.Page++
		}
	}
	out := make([]domain.PluginPackageUploadListItem, 0, len(items))
	summary := domain.PluginPackageUploadSummary{Total: total}
	for _, it := range summaryItems {
		addUploadSummary(&summary, it.Status)
	}
	for _, it := range items {
		risk := parseRiskReport(it.RiskReportJSON)
		out = append(out, domain.PluginPackageUploadListItem{
			UploadID:          it.UploadID,
			OriginalFilename:  it.OriginalFilename,
			PackageCode:       it.PackageCode,
			PackageName:       it.PackageName,
			PackageVersion:    it.PackageVersion,
			Status:            it.Status,
			RiskLevel:         it.RiskLevel,
			ChecksumStatus:    it.ChecksumStatus,
			SignatureStatus:   it.SignatureStatus,
			TrustStatus:       it.TrustStatus,
			UploadedBy:        it.UploadedBy,
			UploadedByName:    it.UploadedByName,
			UploadedAt:        it.UploadedAt,
			ExpiresAt:         it.ExpiresAt,
			PromotedPath:      it.PromotedPath,
			ApprovalID:        it.ApprovalID,
			InstallApprovalID: it.InstallApprovalID,
			ErrorCode:         it.ErrorCode,
			RiskSummary:       risk.Summary,
		})
	}
	return domain.PluginPackageUploadListResponse{Items: out, Pagination: domain.Pagination{Page: filter.Page, PageSize: filter.PageSize, Total: total}, Summary: summary}, nil
}

func normalizeUploadFilter(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "all" {
		return ""
	}
	return value
}

func (s *Service) GetPluginPackageUploadDetail(uploadID string) (domain.PluginPackageUploadDetailResponse, error) {
	record, err := s.getPluginPackageUploadRecord(uploadID)
	if err != nil {
		return domain.PluginPackageUploadDetailResponse{}, err
	}
	var approval *domain.PluginApprovalRequest
	if record.ApprovalID > 0 {
		if it, ok := s.repo.PluginApprovalRequestByID(record.ApprovalID); ok {
			approval = &it
		}
	}
	var installApproval *domain.PluginApprovalRequest
	if record.InstallApprovalID > 0 {
		if it, ok := s.repo.PluginApprovalRequestByID(record.InstallApprovalID); ok {
			installApproval = &it
		}
	}
	return domain.PluginPackageUploadDetailResponse{
		Record:             record,
		ZipScan:            parseZipScan(record.ZipScanJSON),
		FileScan:           parseFileScan(record.FileScanJSON),
		Checksum:           domain.PluginPackageChecksumResult{Status: record.ChecksumStatus},
		Signature:          domain.PluginPackageSignatureResult{VerificationStatus: record.SignatureStatus, PublisherID: record.PublisherID, TrustStatus: record.TrustStatus},
		RiskReport:         parseRiskReport(record.RiskReportJSON),
		ManifestValidation: parseManifestValidation(record.ManifestValidationJSON),
		InstallDryRun:      parseInstallDryRun(record.InstallDryRunJSON),
		Approval:           approval,
		InstallApproval:    installApproval,
		Actions:            uploadActions(record),
	}, nil
}

func (s *Service) RescanPluginPackageUpload(uploadID string) (domain.PluginPackageUploadDetailResponse, error) {
	record, err := s.getPluginPackageUploadRecord(uploadID)
	if err != nil {
		return domain.PluginPackageUploadDetailResponse{}, err
	}
	if !uploadStatusIn(record.Status, domain.PluginPackageUploadStatusUploaded, domain.PluginPackageUploadStatusScanned, domain.PluginPackageUploadStatusStaged, domain.PluginPackageUploadStatusBlocked, domain.PluginPackageUploadStatusFailed) {
		return domain.PluginPackageUploadDetailResponse{}, domain.NewPluginError("plugin_package_upload_action_not_allowed", "当前状态不允许重新扫描").
			WithStatus(400).
			WithDetail("status", record.Status).
			WithSuggestion("仅 uploaded / scanned / staged / blocked / failed 可以 rescan；deleted / expired 不允许。")
	}
	packagePath := strings.TrimSpace(record.PackagePath)
	container := strings.TrimSpace(record.StagingPath)
	if packagePath == "" {
		packagePath, container, err = findUploadedPackagePath(record.UploadID)
		if err != nil {
			return domain.PluginPackageUploadDetailResponse{}, err
		}
	}
	dry, err := s.DryRunPluginPackage(packagePath)
	if err != nil {
		record.Status = domain.PluginPackageUploadStatusFailed
		record.ErrorCode, record.ErrorMessage = uploadErrorCodeMessage(err)
		_, _ = s.repo.SavePluginPackageUpload(record)
		return domain.PluginPackageUploadDetailResponse{}, err
	}
	record = uploadRecordFromDryRun(record, packagePath, container, parseZipScan(record.ZipScanJSON), dry)
	record, _ = s.repo.SavePluginPackageUpload(record)
	return s.GetPluginPackageUploadDetail(record.UploadID)
}

func (s *Service) CancelPluginPackageUpload(uploadID string) (domain.PluginPackageUploadDetailResponse, error) {
	record, err := s.getPluginPackageUploadRecord(uploadID)
	if err != nil {
		return domain.PluginPackageUploadDetailResponse{}, err
	}
	if uploadStatusIn(record.Status, domain.PluginPackageUploadStatusDeleted, domain.PluginPackageUploadStatusExpired, domain.PluginPackageUploadStatusInstalled, domain.PluginPackageUploadStatusPromoted) {
		return domain.PluginPackageUploadDetailResponse{}, domain.NewPluginError("plugin_package_upload_action_not_allowed", "当前状态不允许取消").
			WithStatus(400).WithDetail("status", record.Status).WithSuggestion("promoted / installed / deleted / expired 上传包不能取消。")
	}
	record.Status = domain.PluginPackageUploadStatusCanceled
	record.ErrorCode = ""
	record.ErrorMessage = ""
	record, _ = s.repo.SavePluginPackageUpload(record)
	return s.GetPluginPackageUploadDetail(record.UploadID)
}

func (s *Service) DeletePluginPackageUpload(uploadID string) (domain.PluginPackageCleanupResponse, error) {
	record, err := s.getPluginPackageUploadRecord(uploadID)
	if err != nil {
		return domain.PluginPackageCleanupResponse{}, err
	}
	item := uploadCleanupItem(record)
	if !uploadDeleteStatusAllowed(record.Status) {
		return domain.PluginPackageCleanupResponse{}, domain.NewPluginError("plugin_package_upload_delete_not_allowed", "当前上传包状态不允许删除").
			WithStatus(400).
			WithDetail("status", record.Status).
			WithSuggestion("仅 uploaded / staged / blocked / failed / expired 上传包可删除；promoted / installed 需保留追溯记录。")
	}
	if err := removeUploadRecordFiles(record); err != nil {
		return domain.PluginPackageCleanupResponse{}, err
	}
	if err := s.repo.DeletePluginPackageUpload(record.UploadID); err != nil {
		return domain.PluginPackageCleanupResponse{}, err
	}
	item.CanDelete = true
	return domain.PluginPackageCleanupResponse{DryRun: false, WillDeleteCount: 1, WillFreeBytes: item.Bytes, DeletedCount: 1, FreedBytes: item.Bytes, Items: []domain.PluginPackageCleanupItem{item}}, nil
}

func (s *Service) CleanupPluginPackageUploads(req domain.PluginPackageCleanupRequest) (domain.PluginPackageCleanupResponse, error) {
	items := []domain.PluginPackageUploadRecord{}
	for page := 1; ; page++ {
		chunk, total, err := s.repo.PluginPackageUploads(domain.PluginPackageUploadFilter{Status: "all", Page: page, PageSize: 50})
		if err != nil {
			return domain.PluginPackageCleanupResponse{}, err
		}
		items = append(items, chunk...)
		if len(items) >= total || len(chunk) == 0 {
			break
		}
	}
	now := timeNow()
	statuses := normalizedCleanupStatuses(req.Statuses)
	if len(statuses) == 0 {
		statuses = []string{domain.PluginPackageUploadStatusBlocked, domain.PluginPackageUploadStatusFailed, domain.PluginPackageUploadStatusExpired}
	}
	statusSet := map[string]bool{}
	for _, status := range statuses {
		statusSet[status] = true
	}
	res := domain.PluginPackageCleanupResponse{DryRun: req.DryRun, ConfirmRequired: true}
	for _, record := range items {
		if record.ExpiresAt != "" && !uploadStatusIn(record.Status, domain.PluginPackageUploadStatusDeleted, domain.PluginPackageUploadStatusExpired, domain.PluginPackageUploadStatusPromoted, domain.PluginPackageUploadStatusInstalled) {
			if ts, err := time.ParseInLocation("2006-01-02 15:04:05", record.ExpiresAt, time.Local); err == nil && now.After(ts) {
				record.Status = domain.PluginPackageUploadStatusExpired
				if !req.DryRun {
					record, _ = s.repo.SavePluginPackageUpload(record)
				}
			}
		}
		if !statusSet[record.Status] || !uploadDeleteStatusAllowed(record.Status) || !cleanupOlderThanAllowed(record.UpdatedAt, req.OlderThanDays, now) {
			continue
		}
		item := uploadCleanupItem(record)
		item.CanDelete = true
		res.Items = append(res.Items, item)
		res.WillDeleteCount++
		res.WillFreeBytes += item.Bytes
	}
	if req.DryRun {
		res.ConfirmToken = s.signPluginPackageCleanupToken("plugin_package_upload", req, res.Items)
		return res, nil
	}
	if err := s.verifyPluginPackageCleanupToken("plugin_package_upload", req, res.Items); err != nil {
		return domain.PluginPackageCleanupResponse{}, err
	}
	for _, item := range res.Items {
		record, ok := s.repo.PluginPackageUploadByUploadID(item.ID)
		if !ok {
			continue
		}
		if !uploadDeleteStatusAllowed(record.Status) {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: status %s not allowed", record.UploadID, record.Status))
			continue
		}
		if err := removeUploadRecordFiles(record); err != nil {
			res.Errors = append(res.Errors, err.Error())
			continue
		}
		if err := s.repo.DeletePluginPackageUpload(record.UploadID); err != nil {
			res.Errors = append(res.Errors, err.Error())
			continue
		}
		res.DeletedCount++
		res.FreedBytes += item.Bytes
	}
	return res, nil
}

func uploadDeleteStatusAllowed(status string) bool {
	return uploadStatusIn(status,
		domain.PluginPackageUploadStatusUploaded,
		domain.PluginPackageUploadStatusStaged,
		domain.PluginPackageUploadStatusBlocked,
		domain.PluginPackageUploadStatusFailed,
		domain.PluginPackageUploadStatusExpired,
	)
}

func uploadCleanupItem(record domain.PluginPackageUploadRecord) domain.PluginPackageCleanupItem {
	bytes := safePathSize(record.UploadPath) + safePathSize(record.StagingPath)
	return domain.PluginPackageCleanupItem{
		Kind:       "upload",
		ID:         record.UploadID,
		Path:       firstNonEmpty(record.PackagePath, record.StagingPath, record.UploadPath),
		Status:     record.Status,
		PluginCode: record.PackageCode,
		Version:    record.PackageVersion,
		Bytes:      bytes,
		Reason:     firstNonEmpty(record.ErrorCode, record.RiskLevel, record.Status),
		CanDelete:  uploadDeleteStatusAllowed(record.Status),
	}
}

func cleanupOlderThanAllowed(updatedAt string, days int, now time.Time) bool {
	if days <= 0 {
		return true
	}
	ts, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(updatedAt), time.Local)
	if err != nil {
		return false
	}
	return now.Sub(ts) >= time.Duration(days)*24*time.Hour
}

func safePathSize(rel string) int64 {
	abs, ok := safePluginPackageDeleteAbs(rel)
	if !ok {
		return 0
	}
	var total int64
	_ = filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		info, ierr := d.Info()
		if ierr == nil && !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func safePathFileCount(rel string) int {
	abs, ok := safePluginPackageDeleteAbs(rel)
	if !ok {
		return 0
	}
	total := 0
	_ = filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if !d.IsDir() {
			total++
		}
		return nil
	})
	return total
}

func safePluginPackageDeleteAbs(rel string) (string, bool) {
	rel = strings.TrimSpace(filepath.ToSlash(rel))
	if rel == "" || strings.Contains(rel, "\x00") || strings.HasPrefix(rel, "../") || rel == ".." || strings.Contains(rel, "/../") {
		return "", false
	}
	root, err := serviceProjectRoot()
	if err != nil {
		return "", false
	}
	abs, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return "", false
	}
	allowed := []string{
		pluginPackageUploadsRoot(),
		pluginPackageStagingRoot(),
		pluginPackageQuarantineRoot(),
		pluginPackagePromoteRoot,
		pluginPackageTemplateRoot,
	}
	for _, base := range allowed {
		baseAbs, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(base)))
		if err != nil {
			continue
		}
		baseAbs = filepath.Clean(baseAbs)
		abs = filepath.Clean(abs)
		sub, err := filepath.Rel(baseAbs, abs)
		if err == nil && sub != "." && !strings.HasPrefix(sub, ".."+string(filepath.Separator)) && sub != ".." {
			return abs, true
		}
	}
	return "", false
}

func removeSafePluginPackagePath(rel string) error {
	abs, ok := safePluginPackageDeleteAbs(rel)
	if !ok {
		if strings.TrimSpace(rel) == "" {
			return nil
		}
		return domain.NewPluginError("plugin_package_delete_path_forbidden", "插件包删除路径不在允许目录内").
			WithStatus(400).
			WithDetail("path", rel).
			WithSuggestion("只能删除 storage/plugins/uploads、staging、quarantine、packages、drafts 下的包。")
	}
	if err := os.RemoveAll(abs); err != nil {
		return err
	}
	cleanupEmptySafePluginPackageParents(abs)
	return nil
}

func cleanupEmptySafePluginPackageParents(abs string) {
	root, err := serviceProjectRoot()
	if err != nil {
		return
	}
	abs = filepath.Clean(abs)
	for _, base := range []string{
		pluginPackageUploadsRoot(),
		pluginPackageStagingRoot(),
		pluginPackageQuarantineRoot(),
		pluginPackagePromoteRoot,
	} {
		baseAbs, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(base)))
		if err != nil {
			continue
		}
		baseAbs = filepath.Clean(baseAbs)
		sub, err := filepath.Rel(baseAbs, abs)
		if err != nil || sub == "." || sub == ".." || strings.HasPrefix(sub, ".."+string(filepath.Separator)) {
			continue
		}
		for dir := filepath.Dir(abs); dir != baseAbs && dir != "." && dir != string(filepath.Separator); dir = filepath.Dir(dir) {
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) > 0 {
				return
			}
			if err := os.Remove(dir); err != nil {
				return
			}
		}
		return
	}
}

func (s *Service) SubmitPluginPackageUploadApproval(operator PluginApprovalOperator, uploadID, reason string) (domain.PluginPackageUploadDetailResponse, error) {
	record, err := s.getPluginPackageUploadRecord(uploadID)
	if err != nil {
		return domain.PluginPackageUploadDetailResponse{}, err
	}
	if record.Status != domain.PluginPackageUploadStatusStaged {
		return domain.PluginPackageUploadDetailResponse{}, domain.NewPluginError("plugin_package_upload_approval_blocked", "当前状态不能提交导入审批").
			WithStatus(400).WithDetail("status", record.Status).WithSuggestion("只有 staged 上传包可以提交导入审批。")
	}
	approval := domain.PluginApprovalRequest{
		Action:                domain.PluginApprovalActionPackagePromote,
		PluginCode:            record.PackageCode,
		PluginName:            record.PackageName,
		TargetVersion:         record.PackageVersion,
		PackagePath:           record.PackagePath,
		PackageChecksumStatus: record.ChecksumStatus,
		PackageRiskLevel:      record.RiskLevel,
		Status:                domain.PluginApprovalStatusPending,
		Reason:                strings.TrimSpace(reason),
		RequestedBy:           operator.ID,
		RequestedByName:       strings.TrimSpace(operator.Name),
		RequestedAt:           Now(),
		DryRunJSON:            record.InstallDryRunJSON,
		RiskReportJSON:        record.RiskReportJSON,
		MetadataJSON: mustJSON(map[string]any{
			"upload_id":        record.UploadID,
			"approval_scope":   "plugin_package_upload",
			"zip_scan_json":    record.ZipScanJSON,
			"file_scan_json":   record.FileScanJSON,
			"signature_status": record.SignatureStatus,
			"trust_status":     record.TrustStatus,
		}),
	}
	approval, err = s.repo.AppendPluginApprovalRequest(approval)
	if err != nil {
		return domain.PluginPackageUploadDetailResponse{}, err
	}
	record.Status = domain.PluginPackageUploadStatusApprovalPending
	record.ApprovalID = approval.ID
	record, _ = s.repo.SavePluginPackageUpload(record)
	return s.GetPluginPackageUploadDetail(record.UploadID)
}

func (s *Service) ReviewPluginPackageUploadApproval(operator PluginApprovalOperator, uploadID string, approve bool, comment string) (domain.PluginPackageUploadDetailResponse, error) {
	record, err := s.getPluginPackageUploadRecord(uploadID)
	if err != nil {
		return domain.PluginPackageUploadDetailResponse{}, err
	}
	if record.Status != domain.PluginPackageUploadStatusApprovalPending || record.ApprovalID <= 0 {
		return domain.PluginPackageUploadDetailResponse{}, domain.NewPluginError("plugin_package_upload_lifecycle_invalid", "当前上传包没有待审批导入申请").
			WithStatus(400).WithDetail("status", record.Status)
	}
	if approve {
		if _, err := s.ApprovePluginApproval(operator, record.ApprovalID, comment); err != nil {
			return domain.PluginPackageUploadDetailResponse{}, err
		}
		record.Status = domain.PluginPackageUploadStatusApproved
	} else {
		if _, err := s.RejectPluginApproval(operator, record.ApprovalID, comment); err != nil {
			return domain.PluginPackageUploadDetailResponse{}, err
		}
		record.Status = domain.PluginPackageUploadStatusApprovalRejected
	}
	record, _ = s.repo.SavePluginPackageUpload(record)
	return s.GetPluginPackageUploadDetail(record.UploadID)
}

func (s *Service) getPluginPackageUploadRecord(uploadID string) (domain.PluginPackageUploadRecord, error) {
	uploadID = strings.TrimSpace(uploadID)
	if uploadID == "" {
		return domain.PluginPackageUploadRecord{}, domain.NewPluginError("plugin_package_upload_not_found", "上传包不存在").
			WithStatus(404).WithSuggestion("请刷新上传包列表后重试。")
	}
	record, ok := s.repo.PluginPackageUploadByUploadID(uploadID)
	if !ok || record.UploadID == "" {
		return domain.PluginPackageUploadRecord{}, domain.NewPluginError("plugin_package_upload_not_found", "上传包不存在").
			WithStatus(404).WithDetail("upload_id", uploadID).WithSuggestion("请刷新上传包列表后重试。")
	}
	return record, nil
}

func uploadResultFromDryRun(uploadID, filename, packagePath, containerPath string, zipScan domain.PluginPackageZipScan, dry domain.PluginPackageDryRunResult) domain.PluginPackageUploadResult {
	return domain.PluginPackageUploadResult{
		UploadID:           uploadID,
		Filename:           filename,
		Status:             dry.Status,
		StagingPath:        containerPath,
		PackagePath:        packagePath,
		ZipScan:            zipScan,
		FileScan:           dry.FileScan,
		Checksum:           dry.Checksum,
		Signature:          dry.Signature,
		ManifestValidation: dry.ManifestValidation,
		InstallDryRun:      dry.InstallDryRun,
		RiskReport:         dry.RiskReport,
		CanPromote:         canPromoteDryRun(dry, packagePath),
		CanSubmitApproval:  canPromoteDryRun(dry, packagePath),
		Warnings:           dry.Warnings,
		Errors:             dry.Errors,
	}
}

func uploadRecordFromDryRun(record domain.PluginPackageUploadRecord, packagePath, containerPath string, zipScan domain.PluginPackageZipScan, dry domain.PluginPackageDryRunResult) domain.PluginPackageUploadRecord {
	status := domain.PluginPackageUploadStatusStaged
	if strings.EqualFold(dry.Status, "blocked") || strings.EqualFold(dry.RiskReport.Level, "blocked") {
		status = domain.PluginPackageUploadStatusBlocked
	}
	if uploadStatusIn(record.Status, domain.PluginPackageUploadStatusPromoted, domain.PluginPackageUploadStatusApproved, domain.PluginPackageUploadStatusApprovalPending, domain.PluginPackageUploadStatusApprovalRejected, domain.PluginPackageUploadStatusCanceled, domain.PluginPackageUploadStatusDeleted, domain.PluginPackageUploadStatusExpired, domain.PluginPackageUploadStatusInstalled) {
		status = record.Status
	}
	record.Status = status
	record.PackagePath = filepath.ToSlash(packagePath)
	record.StagingPath = filepath.ToSlash(containerPath)
	record.PackageCode = strings.TrimSpace(dry.Package.Code)
	record.PackageName = strings.TrimSpace(dry.Package.Name)
	record.PackageVersion = strings.TrimSpace(dry.Package.Version)
	record.CompressedSize = zipScan.CompressedSize
	record.UncompressedSize = zipScan.UncompressedSize
	record.FileCount = dry.FileScan.TotalFiles
	record.ChecksumStatus = strings.TrimSpace(dry.Checksum.Status)
	record.SignatureStatus = firstNonEmpty(strings.TrimSpace(dry.Signature.VerificationStatus), "missing")
	record.PublisherID = strings.TrimSpace(dry.Signature.PublisherID)
	record.TrustStatus = strings.TrimSpace(dry.Signature.TrustStatus)
	record.RiskLevel = strings.TrimSpace(dry.RiskReport.Level)
	record.ZipScanJSON = mustJSON(scrubAnyForSnapshot(zipScan))
	record.FileScanJSON = mustJSON(scrubAnyForSnapshot(dry.FileScan))
	record.RiskReportJSON = mustJSON(scrubAnyForSnapshot(dry.RiskReport))
	record.ManifestValidationJSON = mustJSON(scrubAnyForSnapshot(dry.ManifestValidation))
	record.InstallDryRunJSON = mustJSON(scrubAnyForSnapshot(dry.InstallDryRun))
	if status == domain.PluginPackageUploadStatusBlocked {
		record.ErrorCode = firstNonEmpty(strings.TrimSpace(dry.BlockedCode), "plugin_package_upload_blocked")
		record.ErrorMessage = firstNonEmpty(strings.Join(dry.BlockedReasons, "; "), "上传包风险校验未通过")
	} else if status == domain.PluginPackageUploadStatusStaged {
		record.ErrorCode = ""
		record.ErrorMessage = ""
	}
	return record
}

func parseZipScan(raw string) domain.PluginPackageZipScan {
	var out domain.PluginPackageZipScan
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func parseFileScan(raw string) domain.PluginPackageFileScan {
	var out domain.PluginPackageFileScan
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func parseRiskReport(raw string) domain.PluginPackageRiskReport {
	var out domain.PluginPackageRiskReport
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func parseManifestValidation(raw string) domain.PluginPackageManifestValidation {
	var out domain.PluginPackageManifestValidation
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func parseInstallDryRun(raw string) domain.PluginManifestValidationResult {
	var out domain.PluginManifestValidationResult
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func uploadActions(record domain.PluginPackageUploadRecord) []domain.PluginPackageUploadAction {
	action := func(name string, enabled bool, code, reason string) domain.PluginPackageUploadAction {
		return domain.PluginPackageUploadAction{Action: name, Enabled: enabled, ReasonCode: code, Reason: reason}
	}
	canRescan := uploadStatusIn(record.Status, domain.PluginPackageUploadStatusUploaded, domain.PluginPackageUploadStatusScanned, domain.PluginPackageUploadStatusStaged, domain.PluginPackageUploadStatusBlocked, domain.PluginPackageUploadStatusFailed)
	canSubmitApproval := record.Status == domain.PluginPackageUploadStatusStaged
	canPromote := record.Status == domain.PluginPackageUploadStatusStaged || record.Status == domain.PluginPackageUploadStatusApproved
	canCancel := !uploadStatusIn(record.Status, domain.PluginPackageUploadStatusDeleted, domain.PluginPackageUploadStatusExpired, domain.PluginPackageUploadStatusPromoted, domain.PluginPackageUploadStatusInstalled)
	canDelete := uploadDeleteStatusAllowed(record.Status)
	return []domain.PluginPackageUploadAction{
		action("rescan", canRescan, disabledReason(canRescan, "plugin_package_upload_action_not_allowed"), disabledText(canRescan, "当前状态不允许重新扫描")),
		action("submit_approval", canSubmitApproval, disabledReason(canSubmitApproval, "plugin_package_upload_approval_blocked"), disabledText(canSubmitApproval, "只有 staged 上传包可以提交导入审批")),
		action("approve", record.Status == domain.PluginPackageUploadStatusApprovalPending, disabledReason(record.Status == domain.PluginPackageUploadStatusApprovalPending, "plugin_package_upload_action_not_allowed"), disabledText(record.Status == domain.PluginPackageUploadStatusApprovalPending, "仅 approval_pending 可审批通过")),
		action("reject", record.Status == domain.PluginPackageUploadStatusApprovalPending, disabledReason(record.Status == domain.PluginPackageUploadStatusApprovalPending, "plugin_package_upload_action_not_allowed"), disabledText(record.Status == domain.PluginPackageUploadStatusApprovalPending, "仅 approval_pending 可拒绝")),
		action("promote", canPromote, disabledReason(canPromote, "plugin_package_upload_invalid_status"), disabledText(canPromote, "只有 staged / approved 上传包可以 promote")),
		action("submit_install_approval", record.Status == domain.PluginPackageUploadStatusPromoted, disabledReason(record.Status == domain.PluginPackageUploadStatusPromoted, "plugin_package_upload_action_not_allowed"), disabledText(record.Status == domain.PluginPackageUploadStatusPromoted, "promoted 后才可进入安装审批流程")),
		action("cancel", canCancel, disabledReason(canCancel, "plugin_package_upload_action_not_allowed"), disabledText(canCancel, "当前状态不可取消")),
		action("delete", canDelete, disabledReason(canDelete, "plugin_package_upload_delete_not_allowed"), disabledText(canDelete, "仅 uploaded / staged / blocked / failed / expired 可删除")),
		action("cleanup", uploadStatusIn(record.Status, domain.PluginPackageUploadStatusBlocked, domain.PluginPackageUploadStatusExpired, domain.PluginPackageUploadStatusFailed), "", ""),
	}
}

func disabledReason(enabled bool, code string) string {
	if enabled {
		return ""
	}
	return code
}

func disabledText(enabled bool, text string) string {
	if enabled {
		return ""
	}
	return text
}

func uploadStatusIn(status string, values ...string) bool {
	for _, value := range values {
		if status == value {
			return true
		}
	}
	return false
}

func addUploadSummary(summary *domain.PluginPackageUploadSummary, status string) {
	switch status {
	case domain.PluginPackageUploadStatusUploaded:
		summary.Uploaded++
	case domain.PluginPackageUploadStatusScanned:
		summary.Scanned++
	case domain.PluginPackageUploadStatusStaged:
		summary.Staged++
	case domain.PluginPackageUploadStatusBlocked:
		summary.Blocked++
	case domain.PluginPackageUploadStatusApprovalPending:
		summary.ApprovalPending++
	case domain.PluginPackageUploadStatusApprovalRejected:
		summary.ApprovalRejected++
	case domain.PluginPackageUploadStatusApproved:
		summary.Approved++
	case domain.PluginPackageUploadStatusPromoted:
		summary.Promoted++
	case domain.PluginPackageUploadStatusInstallApprovalPending:
		summary.InstallApprovalPending++
	case domain.PluginPackageUploadStatusInstalled:
		summary.Installed++
	case domain.PluginPackageUploadStatusCanceled:
		summary.Canceled++
	case domain.PluginPackageUploadStatusExpired:
		summary.Expired++
	case domain.PluginPackageUploadStatusDeleted:
		summary.Deleted++
	case domain.PluginPackageUploadStatusFailed:
		summary.Failed++
	}
}

func uploadErrorCodeMessage(err error) (string, string) {
	if apiErr, ok := err.(*domain.APIError); ok && apiErr != nil {
		return apiErr.Code, apiErr.Message
	}
	return "plugin_package_upload_failed", err.Error()
}

func removeUploadRecordFiles(record domain.PluginPackageUploadRecord) error {
	for _, rel := range []string{record.UploadPath, record.StagingPath} {
		rel = strings.TrimSpace(rel)
		if rel == "" || strings.HasPrefix(rel, pluginPackagePromoteRoot+"/") {
			continue
		}
		if err := removeSafePluginPackagePath(rel); err != nil {
			return err
		}
	}
	return nil
}

func canPromoteDryRun(dry domain.PluginPackageDryRunResult, packagePath string) bool {
	if !strings.HasPrefix(packagePath, pluginPackageStagingRoot()+"/") {
		return false
	}
	if dry.Status == "blocked" || dry.RiskReport.Level == "blocked" || dry.Checksum.Status == "failed" || !dry.ManifestValidation.Valid {
		return false
	}
	return strings.TrimSpace(dry.Package.Code) != ""
}

func sanitizeUploadFilename(name string) string {
	name = strings.TrimSpace(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, name)
	return strings.Trim(name, "._-")
}

func newPluginPackageUploadID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("pkg_upload_%d", time.Now().UnixNano())
	}
	return "pkg_upload_" + hex.EncodeToString(b[:])
}

func moveUploadToQuarantine(root, uploadID string) (string, error) {
	src := filepath.Join(root, filepath.FromSlash(pluginPackageStagingRoot()), uploadID)
	dst := filepath.Join(root, filepath.FromSlash(pluginPackageQuarantineRoot()), uploadID)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	_ = os.RemoveAll(dst)
	if err := os.Rename(src, dst); err != nil {
		return "", err
	}
	return filepath.ToSlash(filepath.Join(pluginPackageQuarantineRoot(), uploadID)), nil
}

func uploadContainerPath(packagePath string) string {
	parts := strings.Split(filepath.ToSlash(packagePath), "/")
	if len(parts) >= 4 && parts[0] == "storage" && parts[1] == "plugins" {
		return strings.Join(parts[:4], "/")
	}
	return ""
}

func findUploadedPackagePath(uploadID string) (string, string, error) {
	root, err := serviceProjectRoot()
	if err != nil {
		return "", "", err
	}
	for _, base := range []string{pluginPackageStagingRoot(), pluginPackageQuarantineRoot()} {
		containerClean := filepath.ToSlash(filepath.Join(base, uploadID))
		containerAbs := filepath.Join(root, filepath.FromSlash(containerClean))
		if info, err := os.Stat(containerAbs); err == nil && info.IsDir() {
			rel, err := detectUploadedPackageRel(containerAbs, containerClean)
			return rel, containerClean, err
		}
	}
	return "", "", domain.NewPluginError("plugin_package_upload_not_found", "上传包不存在").
		WithStatus(404).
		WithDetail("upload_id", uploadID).
		WithSuggestion("请确认 upload_id 是否正确，或重新上传插件包。")
}

func detectUploadedPackageRel(containerAbs, containerClean string) (string, error) {
	direct := filepath.Join(containerAbs, "manifest.json")
	if fileExists(direct) {
		return containerClean, nil
	}
	entries, err := os.ReadDir(containerAbs)
	if err != nil {
		return "", err
	}
	matches := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if fileExists(filepath.Join(containerAbs, entry.Name(), "manifest.json")) {
			matches = append(matches, filepath.ToSlash(filepath.Join(containerClean, entry.Name())))
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", domain.NewPluginError("plugin_package_zip_multiple_manifests", "上传包中发现多个 manifest.json").
			WithStatus(400).
			WithDetail("manifests", matches).
			WithSuggestion("本轮只支持单插件包 zip。")
	}
	return "", domain.NewPluginError("plugin_package_zip_manifest_missing", "上传包中未找到 manifest.json").
		WithStatus(400).
		WithSuggestion("请确认 zip 根目录或单一顶层目录内包含 manifest.json。")
}

func copyPackageTree(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink forbidden: %s", rel)
		}
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("non-regular file forbidden: %s", rel)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			_ = in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeOutErr := out.Close()
		closeInErr := in.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeOutErr != nil {
			return closeOutErr
		}
		return closeInErr
	})
}
