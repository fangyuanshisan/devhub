package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (s *Server) uploadAdminPluginPackageZip(c *gin.Context) {
	actor := auditActor(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 21*1024*1024)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		apiErr := domain.NewPluginError("plugin_package_upload_invalid_type", "缺少插件包 zip 文件").
			WithStatus(400).
			WithSuggestion("请使用 multipart/form-data，并在 file 字段上传 .zip 插件包。")
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			apiErr = domain.NewPluginError("plugin_package_upload_too_large", "上传 zip 超过大小限制").
				WithStatus(400).
				WithDetail("max_bytes", pluginregistry.PluginPackageUploadMaxZipSize).
				WithSuggestion("请将 zip 控制在 20MB 以内。")
		}
		s.auditStructured(c, "system", "plugin.package.upload.failed", "plugin-package-uploads", nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"operation": "plugin_package_upload", "actor": actor}, auditAPIErrorFields(apiErr)))
		failAPIError(c, apiErr)
		return
	}
	defer file.Close()

	filename := ""
	size := int64(0)
	if header != nil {
		filename = header.Filename
		size = header.Size
	}
	s.auditStructured(c, "system", "plugin.package.upload.started", "plugin-package-uploads", nil,
		gin.H{"status": "started"},
		gin.H{"operation": "plugin_package_upload", "filename": filename, "size": size, "actor": actor})
	adminCtx, hasAdmin := currentAdminContext(c)
	operator := service.PluginPackageUploadOperator{}
	if hasAdmin {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.UploadPluginPackageZipAs(operator, filename, size, file)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.upload.failed", "plugin-package-uploads", nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"operation": "plugin_package_upload", "filename": filename, "actor": actor, "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.uploaded", fmt.Sprintf("plugin-package-uploads#%s", res.UploadID), nil,
		gin.H{"status": res.Status},
		gin.H{
			"operation":           "plugin_package_upload",
			"upload_id":           res.UploadID,
			"filename":            res.Filename,
			"package_path":        res.PackagePath,
			"staging_path":        res.StagingPath,
			"zip_scan":            res.ZipScan,
			"risk_level":          res.RiskReport.Level,
			"checksum_status":     res.Checksum.Status,
			"manifest_valid":      res.ManifestValidation.Valid,
			"can_promote":         res.CanPromote,
			"can_submit_approval": res.CanSubmitApproval,
			"actor":               actor,
		})
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginPackageUploads(c *gin.Context) {
	filter := domain.PluginPackageUploadFilter{
		Status:      strings.TrimSpace(c.Query("status")),
		RiskLevel:   strings.TrimSpace(c.Query("risk_level")),
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		PackageCode: strings.TrimSpace(c.Query("package_code")),
		PublisherID: strings.TrimSpace(c.Query("publisher_id")),
		TrustStatus: strings.TrimSpace(c.Query("trust_status")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	if raw := strings.TrimSpace(c.Query("uploaded_by")); raw != "" {
		filter.UploadedBy, _ = strconv.ParseInt(raw, 10, 64)
	}
	res, err := s.svc.ListPluginPackageUploads(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginPackageUploadDetail(c *gin.Context) {
	uploadID := strings.TrimSpace(c.Param("upload_id"))
	res, err := s.svc.GetPluginPackageUploadDetail(uploadID)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) rescanAdminPluginPackageUpload(c *gin.Context) {
	uploadID := strings.TrimSpace(c.Param("upload_id"))
	res, err := s.svc.RescanPluginPackageUpload(uploadID)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.upload.rescan.failed", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"upload_id": uploadID, "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.upload.rescanned", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": res.Record.Status}, gin.H{"upload_id": uploadID, "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) cancelAdminPluginPackageUpload(c *gin.Context) {
	uploadID := strings.TrimSpace(c.Param("upload_id"))
	res, err := s.svc.CancelPluginPackageUpload(uploadID)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.upload.cancel.failed", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"upload_id": uploadID, "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.upload.canceled", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": res.Record.Status}, gin.H{"upload_id": uploadID, "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) deleteAdminPluginPackageUpload(c *gin.Context) {
	uploadID := strings.TrimSpace(c.Param("upload_id"))
	res, err := s.svc.DeletePluginPackageUpload(uploadID)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.upload.delete.failed", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"upload_id": uploadID, "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.upload.deleted", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": res.Record.Status}, gin.H{"upload_id": uploadID, "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) cleanupAdminPluginPackageUploads(c *gin.Context) {
	res, err := s.svc.CleanupPluginPackageUploads()
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.upload.cleanup.failed", "plugin-package-uploads", nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.upload.cleaned", "plugin-package-uploads", nil, gin.H{"status": "ok"}, gin.H{"cleaned": res.Cleaned, "expired": res.Expired, "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) submitAdminPluginPackageUploadApproval(c *gin.Context) {
	uploadID := strings.TrimSpace(c.Param("upload_id"))
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	operator := service.PluginApprovalOperator{ID: adminCtx.CurrentUser.ID, Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)}
	res, err := s.svc.SubmitPluginPackageUploadApproval(operator, uploadID, req.Reason)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.upload.approval.failed", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"upload_id": uploadID, "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.upload.approval.submitted", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": res.Record.Status}, gin.H{"upload_id": uploadID, "approval_id": res.Record.ApprovalID, "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) approveAdminPluginPackageUpload(c *gin.Context) {
	s.reviewAdminPluginPackageUpload(c, true)
}

func (s *Server) rejectAdminPluginPackageUpload(c *gin.Context) {
	s.reviewAdminPluginPackageUpload(c, false)
}

func (s *Server) reviewAdminPluginPackageUpload(c *gin.Context, approve bool) {
	uploadID := strings.TrimSpace(c.Param("upload_id"))
	var req domain.PluginApprovalReviewRequest
	_ = c.ShouldBindJSON(&req)
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	operator := service.PluginApprovalOperator{ID: adminCtx.CurrentUser.ID, Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)}
	res, err := s.svc.ReviewPluginPackageUploadApproval(operator, uploadID, approve, req.Comment)
	action := "approve"
	if !approve {
		action = "reject"
	}
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.upload."+action+".failed", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"upload_id": uploadID, "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.upload."+action+"ed", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil, gin.H{"status": res.Record.Status}, gin.H{"upload_id": uploadID, "approval_id": res.Record.ApprovalID, "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) promoteAdminPluginPackageUpload(c *gin.Context) {
	uploadID := strings.TrimSpace(c.Param("upload_id"))
	var req struct {
		Force bool `json:"force"`
	}
	_ = c.ShouldBindJSON(&req)
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.package.promote.started", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil,
		gin.H{"status": "started"},
		gin.H{"operation": "plugin_package_promote", "upload_id": uploadID, "force": req.Force, "actor": actor})
	res, err := s.svc.PromotePluginPackageUpload(uploadID, req.Force)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.promote.failed", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"operation": "plugin_package_promote", "upload_id": uploadID, "actor": actor, "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.promoted", fmt.Sprintf("plugin-package-uploads#%s", uploadID), nil,
		gin.H{"status": res.Status},
		gin.H{"operation": "plugin_package_promote", "upload_id": uploadID, "package_path": res.PackagePath, "risk_level": res.DryRun.RiskReport.Level, "actor": actor})
	c.JSON(http.StatusOK, res)
}

func (s *Server) previewAdminPluginPackageTemplate(c *gin.Context) {
	var req domain.PluginPackageTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.PreviewPluginPackageTemplate(req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) createAdminPluginPackageTemplate(c *gin.Context) {
	var req domain.PluginPackageTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.package.template.create.started", "plugins", nil,
		gin.H{"status": "started"},
		gin.H{"operation": "plugin_package_template_create", "plugin_code": strings.TrimSpace(req.Code), "actor": actor})
	res, err := s.svc.CreatePluginPackageTemplate(req)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.template.create.failed", "plugins", nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"operation": "plugin_package_template_create", "plugin_code": strings.TrimSpace(req.Code), "actor": actor, "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.template.created", fmt.Sprintf("plugin-packages#%s", res.Template.Code), nil,
		gin.H{"status": res.Status},
		gin.H{
			"operation":           "plugin_package_template_create",
			"plugin_code":         res.Template.Code,
			"package_path":        res.Template.PackagePath,
			"dry_run_status":      res.DryRun.Status,
			"risk_level":          res.DryRun.RiskReport.Level,
			"manifest_valid":      res.DryRun.ManifestValidation.Valid,
			"generated_files":     res.Template.Files,
			"registry_go_omitted": true,
			"actor":               actor,
		})
	c.JSON(http.StatusOK, res)
}

func (s *Server) dryRunAdminPluginPackage(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.svc.DryRunPluginPackage(req.Path)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) installAdminPluginPackage(c *gin.Context) {
	var req domain.PluginPackageInstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	user, _ := currentUser(c)
	opName := strings.TrimSpace(user.Nickname)
	if opName == "" {
		opName = strings.TrimSpace(user.Username)
	}
	operator := service.PluginOperationOperator{ID: user.ID, Name: opName}
	executor := auditActor(c)
	s.auditStructured(c, "system", "plugin.package.install.started", "plugins", nil,
		gin.H{"status": "started"},
		mergeAuditMeta(gin.H{"operation": "plugin_package_install", "path": strings.TrimSpace(req.Path), "actor": executor}, nil))

	res, err := s.svc.InstallPluginPackage(operator, req)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.install.failed", "plugins", nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"operation": "plugin_package_install", "path": strings.TrimSpace(req.Path), "error": err.Error(), "actor": executor}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.installed", fmt.Sprintf("plugins#%s", res.Plugin.Code), nil,
		gin.H{"status": "installed"},
		mergeAuditMeta(gin.H{
			"operation":       "plugin_package_install",
			"plugin_code":     res.Plugin.Code,
			"install_source":  res.Plugin.SourceType,
			"package_path":    res.Package.Path,
			"risk_level":      res.RiskLevel,
			"checksum_status": res.Checksum.Status,
			"actor":           executor,
		}, nil))
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginPackages(c *gin.Context) {
	root := strings.TrimSpace(c.Query("root"))
	status := strings.TrimSpace(c.DefaultQuery("status", "all"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	risk := strings.TrimSpace(c.Query("risk_level"))
	checksum := strings.TrimSpace(c.Query("checksum_status"))
	manifestValid := strings.TrimSpace(c.Query("manifest_valid"))

	page := 1
	pageSize := 20
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			pageSize = n
		}
	}

	resp, err := s.svc.ListPluginPackages(root, service.PluginPackageRepositoryFilter{
		Status:         status,
		Keyword:        keyword,
		RiskLevel:      risk,
		ChecksumStatus: checksum,
		ManifestValid:  manifestValid,
		Page:           page,
		PageSize:       pageSize,
	})
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminPluginPackageDetail(c *gin.Context) {
	path := strings.TrimSpace(c.Query("path"))
	if path == "" {
		failAPIError(c, domain.NewPluginError("plugin_package_path_invalid", "缺少插件包路径").WithStatus(400).WithSuggestion("请提供 query 参数 path，例如 storage/plugins/packages/demo_notice。"))
		return
	}
	res, err := s.svc.GetPluginPackageDetail(path)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) dryRunAdminPluginExport(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	var req domain.PluginPackageExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.DryRunPluginPackageExport(code, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) exportAdminPluginPackage(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	var req domain.PluginPackageExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.export.started", fmt.Sprintf("plugins#%s", code), nil,
		gin.H{"status": "started"},
		gin.H{"plugin_code": code, "operation": "plugin_export", "actor": actor})
	res, err := s.svc.ExportPluginPackage(code, req)
	if err != nil {
		s.auditStructured(c, "system", "plugin.export.failed", fmt.Sprintf("plugins#%s", code), nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"plugin_code": code, "operation": "plugin_export", "actor": actor, "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.exported", fmt.Sprintf("plugins#%s", code), nil,
		gin.H{"status": "exported"},
		gin.H{
			"plugin_code":            res.PluginCode,
			"operation":              "plugin_export",
			"actor":                  actor,
			"output_dir":             res.OutputDir,
			"files":                  res.Files,
			"checksum_status":        res.ChecksumStatus,
			"package_dry_run_status": res.PackageDryRunStatus,
		})
	c.JSON(http.StatusOK, res)
}
