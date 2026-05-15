package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
)

type PluginApprovalOperator struct {
	ID   int64
	Name string
}

func (s *Service) CreatePluginApproval(operator PluginApprovalOperator, req domain.PluginApprovalCreateRequest) (domain.PluginApprovalRequest, error) {
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action != domain.PluginApprovalActionInstall && action != domain.PluginApprovalActionUpgrade {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_invalid_action", "不支持的审批动作").
			WithStatus(400).
			WithDetail("action", action).
			WithSuggestion("action 仅支持 install / upgrade。")
	}

	record := domain.PluginApprovalRequest{
		Action:          action,
		Status:          domain.PluginApprovalStatusPending,
		Reason:          strings.TrimSpace(req.Reason),
		RequestedBy:     operator.ID,
		RequestedByName: strings.TrimSpace(operator.Name),
	}

	switch action {
	case domain.PluginApprovalActionInstall:
		path := strings.TrimSpace(req.PackagePath)
		if path == "" {
			return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_package_path_invalid", "缺少插件包路径").
				WithStatus(400).
				WithSuggestion("请提供 package_path，例如 storage/plugins/packages/demo_notice。")
		}
		dry, err := s.DryRunPluginPackage(path)
		if err != nil {
			return domain.PluginApprovalRequest{}, err
		}
		if strings.ToLower(strings.TrimSpace(dry.Status)) == "blocked" || strings.ToLower(strings.TrimSpace(dry.RiskReport.Level)) == "blocked" {
			return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_create_blocked", "插件包风险校验未通过，禁止提交审批").
				WithStatus(400).
				WithDetail("path", dry.Package.Path).
				WithDetail("risk_level", dry.RiskReport.Level).
				WithDetail("blocked_code", dry.BlockedCode).
				WithDetail("blocked_reasons", dry.BlockedReasons).
				WithSuggestion("请先根据风险报告修复阻断项，再重新 dry-run 并提交审批。")
		}
		if dry.Package.ChecksumFound && strings.ToLower(strings.TrimSpace(dry.Checksum.Status)) != "ok" {
			return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_create_blocked", "插件包 checksum 校验未通过，禁止提交审批").
				WithStatus(400).
				WithDetail("path", dry.Package.Path).
				WithDetail("checksum_status", dry.Checksum.Status).
				WithSuggestion("请修复 checksums.json 或移除被篡改文件后重试。")
		}
		if !dry.ManifestValidation.Valid || !dry.InstallDryRun.Valid {
			return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_create_blocked", "插件包安装预览未通过，禁止提交审批").
				WithStatus(400).
				WithDetail("path", dry.Package.Path).
				WithDetail("manifest_errors", dry.ManifestValidation.Errors).
				WithDetail("install_errors", dry.InstallDryRun.Errors).
				WithSuggestion("请先修复依赖/版本/声明冲突等问题后再提交审批。")
		}

		abs, _, nerr := pluginregistry.NormalizePluginPackagePath(path)
		if nerr != nil {
			return domain.PluginApprovalRequest{}, nerr
		}
		manifestRaw, rerr := os.ReadFile(filepath.Join(abs, "manifest.json"))
		if rerr != nil {
			return domain.PluginApprovalRequest{}, fmt.Errorf("读取 manifest.json 失败：%w", rerr)
		}

		record.PluginCode = strings.TrimSpace(dry.Package.Code)
		record.PluginName = strings.TrimSpace(dry.Package.Name)
		record.TargetVersion = strings.TrimSpace(dry.Package.Version)
		record.PackagePath = strings.TrimSpace(dry.Package.Path)
		record.PackageChecksumStatus = strings.TrimSpace(dry.Checksum.Status)
		record.PackageRiskLevel = strings.TrimSpace(dry.RiskReport.Level)
		record.ManifestJSON = scrubManifestJSONForSnapshot(string(manifestRaw))

		record.RiskReportJSON = mustJSON(scrubAnyForSnapshot(dry.RiskReport))
		record.DependencySummaryJSON = mustJSON(scrubAnyForSnapshot(dry.InstallDryRun.Dependencies))
		record.CompatibilityJSON = mustJSON(scrubAnyForSnapshot(dry.InstallDryRun.Compatibility))
		record.DryRunJSON = mustJSON(scrubAnyForSnapshot(map[string]any{
			"package":             dry.Package,
			"file_scan":           dry.FileScan,
			"checksum":            dry.Checksum,
			"manifest_validation": dry.ManifestValidation,
			"install_dry_run":     dry.InstallDryRun,
			"risk_report":         dry.RiskReport,
			"status":              dry.Status,
			"blocked_code":        dry.BlockedCode,
			"blocked_reasons":     dry.BlockedReasons,
			"warnings":            dry.Warnings,
			"errors":              dry.Errors,
		}))

	case domain.PluginApprovalActionUpgrade:
		code := strings.TrimSpace(req.PluginCode)
		if code == "" {
			return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_not_found", "缺少 plugin_code").
				WithStatus(400).
				WithSuggestion("请提供 plugin_code，例如 qa。")
		}
		var manifestRaw []byte
		if strings.TrimSpace(req.PackagePath) != "" {
			// Upgrade based on a local package: reuse package dry-run + risk report.
			dry, err := s.DryRunPluginPackage(req.PackagePath)
			if err != nil {
				return domain.PluginApprovalRequest{}, err
			}
			if strings.ToLower(strings.TrimSpace(dry.Status)) == "blocked" || strings.ToLower(strings.TrimSpace(dry.RiskReport.Level)) == "blocked" {
				return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_create_blocked", "插件包风险校验未通过，禁止提交升级审批").
					WithStatus(400).
					WithDetail("path", dry.Package.Path).
					WithDetail("risk_level", dry.RiskReport.Level).
					WithDetail("blocked_code", dry.BlockedCode).
					WithDetail("blocked_reasons", dry.BlockedReasons).
					WithSuggestion("请先修复阻断项后重试。")
			}
			abs, _, nerr := pluginregistry.NormalizePluginPackagePath(req.PackagePath)
			if nerr != nil {
				return domain.PluginApprovalRequest{}, nerr
			}
			raw, rerr := os.ReadFile(filepath.Join(abs, "manifest.json"))
			if rerr != nil {
				return domain.PluginApprovalRequest{}, fmt.Errorf("读取 manifest.json 失败：%w", rerr)
			}
			manifestRaw = raw
			record.PackagePath = strings.TrimSpace(dry.Package.Path)
			record.PackageChecksumStatus = strings.TrimSpace(dry.Checksum.Status)
			record.PackageRiskLevel = strings.TrimSpace(dry.RiskReport.Level)
			record.RiskReportJSON = mustJSON(scrubAnyForSnapshot(dry.RiskReport))
		} else {
			manifestRaw = []byte(mustMarshalManifestInput(req.ManifestJSON))
		}

		preview, err := s.PluginUpgradeDryRun(code, manifestRaw)
		if err != nil {
			return domain.PluginApprovalRequest{}, err
		}
		if !preview.Validation.Valid || strings.ToLower(strings.TrimSpace(preview.CompatibilityStatus)) != "compatible" {
			return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_create_blocked", "升级预览未通过，禁止提交审批").
				WithStatus(400).
				WithDetail("plugin_code", code).
				WithDetail("errors", preview.Validation.Errors).
				WithDetail("compatibility_status", preview.CompatibilityStatus).
				WithSuggestion("请先修复依赖/版本/兼容问题后再提交审批。")
		}

		record.PluginCode = code
		record.PluginName = strings.TrimSpace(preview.Validation.NormalizedManifest.Name)
		record.CurrentVersion = strings.TrimSpace(preview.CurrentVersion)
		record.TargetVersion = strings.TrimSpace(preview.NewVersion)
		record.ManifestJSON = scrubManifestJSONForSnapshot(string(manifestRaw))
		record.DryRunJSON = mustJSON(scrubAnyForSnapshot(preview))
		record.DependencySummaryJSON = mustJSON(scrubAnyForSnapshot(preview.DependencyDiff))
		record.CompatibilityJSON = mustJSON(scrubAnyForSnapshot(map[string]any{
			"status":                  preview.CompatibilityStatus,
			"core_version":            preview.CurrentCoreVersion,
			"compatible_core_version": preview.CompatibleCoreVersion,
		}))
		record.ChangedKeysJSON = mustJSON(scrubAnyForSnapshot(preview.ChangedKeys))
		record.DiffJSON = mustJSON(scrubAnyForSnapshot(preview.Diff))
	}

	saved, err := s.repo.AppendPluginApprovalRequest(record)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	return saved, nil
}

func (s *Service) ListPluginApprovals(filter domain.PluginApprovalFilter) (domain.PluginApprovalListResponse, error) {
	items, total, err := s.repo.PluginApprovalRequests(filter)
	if err != nil {
		return domain.PluginApprovalListResponse{}, err
	}
	out := make([]domain.PluginApprovalListItem, 0, len(items))
	for _, it := range items {
		out = append(out, domain.PluginApprovalListItem{
			ID:              it.ID,
			RequestNo:       it.RequestNo,
			Action:          it.Action,
			PluginCode:      it.PluginCode,
			PluginName:      it.PluginName,
			CurrentVersion:  it.CurrentVersion,
			TargetVersion:   it.TargetVersion,
			Status:          it.Status,
			PackagePath:     it.PackagePath,
			RiskLevel:       it.PackageRiskLevel,
			ChecksumStatus:  it.PackageChecksumStatus,
			RequestedByName: it.RequestedByName,
			RequestedAt:     it.RequestedAt,
			ReviewedByName:  it.ReviewedByName,
			ReviewedAt:      it.ReviewedAt,
		})
	}
	return domain.PluginApprovalListResponse{
		Items: out,
		Pagination: domain.Pagination{
			Page:     filter.Page,
			PageSize: filter.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *Service) GetPluginApproval(id int64) (domain.PluginApprovalRequest, error) {
	it, ok := s.repo.PluginApprovalRequestByID(id)
	if !ok || it.ID == 0 {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_not_found", "审批记录不存在").
			WithStatus(404).
			WithDetail("id", id).
			WithSuggestion("请刷新审批列表后重试。")
	}
	return it, nil
}

func (s *Service) ApprovePluginApproval(operator PluginApprovalOperator, id int64, comment string) (domain.PluginApprovalRequest, error) {
	it, err := s.GetPluginApproval(id)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	if it.Status != domain.PluginApprovalStatusPending {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_invalid_status", "当前状态不允许审批通过").
			WithStatus(400).
			WithDetail("status", it.Status).
			WithSuggestion("仅 pending 状态可以审批。")
	}
	it.Status = domain.PluginApprovalStatusApproved
	it.ReviewedBy = operator.ID
	it.ReviewedByName = strings.TrimSpace(operator.Name)
	it.ReviewComment = strings.TrimSpace(comment)
	it.ReviewedAt = Now()
	saved, err := s.repo.SavePluginApprovalRequest(it)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	return saved, nil
}

func (s *Service) RejectPluginApproval(operator PluginApprovalOperator, id int64, comment string) (domain.PluginApprovalRequest, error) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_reject_reason_required", "拒绝原因不能为空").
			WithStatus(400).
			WithSuggestion("请填写 comment（拒绝原因）后重试。")
	}
	it, err := s.GetPluginApproval(id)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	if it.Status != domain.PluginApprovalStatusPending {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_invalid_status", "当前状态不允许拒绝").
			WithStatus(400).
			WithDetail("status", it.Status).
			WithSuggestion("仅 pending 状态可以拒绝。")
	}
	it.Status = domain.PluginApprovalStatusRejected
	it.ReviewedBy = operator.ID
	it.ReviewedByName = strings.TrimSpace(operator.Name)
	it.ReviewedAt = Now()
	it.ReviewComment = comment
	saved, err := s.repo.SavePluginApprovalRequest(it)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	return saved, nil
}

func (s *Service) CancelPluginApproval(operator PluginApprovalOperator, id int64, comment string, allowManage bool) (domain.PluginApprovalRequest, error) {
	it, err := s.GetPluginApproval(id)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	if it.Status != domain.PluginApprovalStatusPending {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_invalid_status", "当前状态不允许撤销").
			WithStatus(400).
			WithDetail("status", it.Status).
			WithSuggestion("仅 pending 状态可以撤销。")
	}
	if !allowManage && it.RequestedBy > 0 && operator.ID > 0 && it.RequestedBy != operator.ID {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_permission_denied", "只有申请人可以撤销该申请").
			WithStatus(403).
			WithDetail("requested_by", it.RequestedBy).
			WithSuggestion("请联系申请人撤销，或使用具备 plugin.manage 权限的账号处理。")
	}
	it.Status = domain.PluginApprovalStatusCanceled
	it.ReviewComment = strings.TrimSpace(comment)
	it.UpdatedAt = Now()
	saved, err := s.repo.SavePluginApprovalRequest(it)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	return saved, nil
}

func (s *Service) ExecutePluginApproval(operator PluginApprovalOperator, id int64) (domain.PluginApprovalRequest, error) {
	it, err := s.GetPluginApproval(id)
	if err != nil {
		return domain.PluginApprovalRequest{}, err
	}
	if it.Status == domain.PluginApprovalStatusExecuted {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_already_executed", "审批已执行，无需重复执行").
			WithStatus(400)
	}
	if it.Status == domain.PluginApprovalStatusCanceled {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_already_canceled", "审批已撤销，不能执行").
			WithStatus(400)
	}
	if it.Status != domain.PluginApprovalStatusApproved {
		return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_invalid_status", "审批未通过，不能执行").
			WithStatus(400).
			WithDetail("status", it.Status).
			WithSuggestion("请先审批通过后再执行。")
	}

	it.ExecutedBy = operator.ID
	it.ExecutedAt = Now()

	switch it.Action {
	case domain.PluginApprovalActionInstall:
		// Re-run dry-run at execution time.
		dry, derr := s.DryRunPluginPackage(it.PackagePath)
		if derr != nil {
			it.Status = domain.PluginApprovalStatusFailed
			it.ErrorCode = "plugin_approval_execute_failed"
			it.ErrorMessage = derr.Error()
			_, _ = s.repo.SavePluginApprovalRequest(it)
			return domain.PluginApprovalRequest{}, derr
		}
		if strings.ToLower(strings.TrimSpace(dry.Status)) == "blocked" || strings.ToLower(strings.TrimSpace(dry.RiskReport.Level)) == "blocked" {
			it.Status = domain.PluginApprovalStatusFailed
			it.ErrorCode = "plugin_approval_execute_blocked"
			it.ErrorMessage = "执行前重新校验发现阻断项，禁止执行"
			it.DryRunJSON = mustJSON(scrubAnyForSnapshot(map[string]any{
				"package":         dry.Package,
				"checksum":        dry.Checksum,
				"risk_report":     dry.RiskReport,
				"status":          dry.Status,
				"blocked_code":    dry.BlockedCode,
				"blocked_reasons": dry.BlockedReasons,
				"errors":          dry.Errors,
				"warnings":        dry.Warnings,
			}))
			_, _ = s.repo.SavePluginApprovalRequest(it)
			return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_execute_blocked", "执行前校验未通过，禁止执行").
				WithStatus(400).
				WithDetail("risk_level", dry.RiskReport.Level).
				WithDetail("blocked_code", dry.BlockedCode).
				WithDetail("blocked_reasons", dry.BlockedReasons).
				WithSuggestion("请先修复阻断项后重新提交审批。")
		}
		opID := newPluginOperationID()
		it.MetadataJSON = mustJSON(scrubAnyForSnapshot(map[string]any{
			"operation_id": opID,
			"approval_id":  it.ID,
		}))
		_, _ = s.repo.SavePluginApprovalRequest(it)
		installRes, ierr := s.InstallPluginPackage(PluginOperationOperator{ID: operator.ID, Name: operator.Name}, domain.PluginPackageInstallRequest{
			Path:             it.PackagePath,
			ConfirmRiskLevel: strings.ToLower(strings.TrimSpace(dry.RiskReport.Level)),
			ApprovalID:       it.ID,
			OperationID:      opID,
		})
		if ierr != nil {
			it.Status = domain.PluginApprovalStatusFailed
			if apiErr, ok := ierr.(*domain.APIError); ok && apiErr != nil {
				it.ErrorCode = apiErr.Code
				it.ErrorMessage = apiErr.Message
			} else {
				it.ErrorCode = "plugin_approval_execute_failed"
				it.ErrorMessage = ierr.Error()
			}
			_, _ = s.repo.SavePluginApprovalRequest(it)
			return domain.PluginApprovalRequest{}, ierr
		}
		it.Status = domain.PluginApprovalStatusExecuted
		it.ExecuteResultJSON = mustJSON(scrubAnyForSnapshot(installRes))
		it.ErrorCode = ""
		it.ErrorMessage = ""
		it.PluginCode = installRes.Plugin.Code
		it.PluginName = installRes.Plugin.Name
		it.TargetVersion = installRes.Plugin.Version
		it.PackageChecksumStatus = installRes.Checksum.Status
		it.PackageRiskLevel = installRes.RiskLevel
		_, _ = s.repo.SavePluginApprovalRequest(it)
		return it, nil

	case domain.PluginApprovalActionUpgrade:
		raw := []byte(it.ManifestJSON)
		preview, perr := s.PluginUpgradeDryRun(it.PluginCode, raw)
		if perr != nil {
			it.Status = domain.PluginApprovalStatusFailed
			it.ErrorCode = "plugin_approval_execute_failed"
			it.ErrorMessage = perr.Error()
			_, _ = s.repo.SavePluginApprovalRequest(it)
			return domain.PluginApprovalRequest{}, perr
		}
		if !preview.Validation.Valid || strings.ToLower(strings.TrimSpace(preview.CompatibilityStatus)) != "compatible" {
			it.Status = domain.PluginApprovalStatusFailed
			it.ErrorCode = "plugin_approval_execute_blocked"
			it.ErrorMessage = "执行前重新 dry-run 发现阻断项"
			it.DryRunJSON = mustJSON(scrubAnyForSnapshot(preview))
			_, _ = s.repo.SavePluginApprovalRequest(it)
			return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_execute_blocked", "执行前升级预览未通过，禁止执行").
				WithStatus(400).
				WithDetail("plugin_code", it.PluginCode).
				WithDetail("errors", preview.Validation.Errors).
				WithDetail("compatibility_status", preview.CompatibilityStatus).
				WithSuggestion("请修复阻断项后重新提交审批。")
		}
		opID := newPluginOperationID()
		it.MetadataJSON = mustJSON(scrubAnyForSnapshot(map[string]any{
			"operation_id": opID,
			"approval_id":  it.ID,
		}))
		_, _ = s.repo.SavePluginApprovalRequest(it)
		upRes, uerr := s.UpgradePluginManifestWithOperation(PluginOperationOperator{ID: operator.ID, Name: operator.Name}, it.ID, opID, it.PluginCode, raw)
		if uerr != nil {
			it.Status = domain.PluginApprovalStatusFailed
			if apiErr, ok := uerr.(*domain.APIError); ok && apiErr != nil {
				it.ErrorCode = apiErr.Code
				it.ErrorMessage = apiErr.Message
			} else {
				it.ErrorCode = "plugin_approval_execute_failed"
				it.ErrorMessage = uerr.Error()
			}
			_, _ = s.repo.SavePluginApprovalRequest(it)
			return domain.PluginApprovalRequest{}, uerr
		}
		it.Status = domain.PluginApprovalStatusExecuted
		it.ExecuteResultJSON = mustJSON(scrubAnyForSnapshot(upRes))
		it.CurrentVersion = preview.CurrentVersion
		it.TargetVersion = preview.NewVersion
		it.PluginName = upRes.Plugin.Name
		_, _ = s.repo.SavePluginApprovalRequest(it)
		return it, nil
	}

	return domain.PluginApprovalRequest{}, domain.NewPluginError("plugin_approval_invalid_action", "不支持的审批动作").
		WithStatus(400).
		WithDetail("action", it.Action)
}

func mustJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func mustMarshalManifestInput(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		// allow passing a JSON string.
		return strings.TrimSpace(t)
	default:
		raw, _ := json.Marshal(v)
		return string(raw)
	}
}

func Now() string {
	return nowString()
}

func nowString() string {
	return timeNow().Format("2006-01-02 15:04:05")
}

// timeNow is a seam for tests.
var timeNow = func() (t time.Time) { return time.Now() }
