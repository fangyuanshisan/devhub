package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (s *Server) downloadAdminPluginPackageToStaging(c *gin.Context) {
	var req domain.PluginPackageDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("plugin_package_download_invalid_request", "下载请求格式无效").
			WithStatus(400).
			WithSuggestion("请提交 plugin_code、version、package_url 和可选 sha256。"))
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.package.download.requested", "plugin-package-downloads", nil, gin.H{"status": "requested"}, gin.H{
		"plugin_code":   req.PluginCode,
		"version":       req.Version,
		"source_url":    req.PackageURL,
		"signature_url": strings.TrimSpace(req.SignatureURL),
		"actor":         actor,
	})
	adminCtx, hasAdmin := currentAdminContext(c)
	operator := service.PluginPackageDownloadOperator{}
	if hasAdmin {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	record, err := s.svc.DownloadPluginPackageToStagingAs(operator, req)
	if err != nil {
		action := "plugin.package.download.failed"
		if record.Status == domain.PluginPackageDownloadStatusRejected {
			action = "plugin.package.download.rejected"
		}
		if record.Status == domain.PluginPackageDownloadStatusChecksumFailed {
			action = "plugin.package.checksum.failed"
		}
		s.auditStructured(c, "system", action, fmt.Sprintf("plugin-package-downloads#%d", record.ID), nil, gin.H{"status": record.Status}, mergeAuditMeta(gin.H{
			"plugin_code":     req.PluginCode,
			"version":         req.Version,
			"source_url":      req.PackageURL,
			"signature_url":   strings.TrimSpace(req.SignatureURL),
			"final_url":       record.FinalURL,
			"file_size":       record.FileSize,
			"sha256_expected": req.SHA256,
			"sha256_actual":   record.SHA256Actual,
			"status":          record.Status,
			"error_code":      record.ErrorCode,
			"error_message":   record.ErrorMessage,
			"actor":           actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.download.success", fmt.Sprintf("plugin-package-downloads#%d", record.ID), nil, gin.H{"status": record.Status}, gin.H{
		"plugin_code":     record.PluginCode,
		"version":         record.Version,
		"source_url":      record.SourceURL,
		"final_url":       record.FinalURL,
		"signature_url":   strings.TrimSpace(record.SignatureURL),
		"file_size":       record.FileSize,
		"sha256_expected": record.SHA256Expected,
		"sha256_actual":   record.SHA256Actual,
		"status":          record.Status,
		"actor":           actor,
	})
	c.JSON(http.StatusOK, record)
}

func (s *Server) listAdminPluginPackageStaging(c *gin.Context) {
	filter := domain.PluginPackageDownloadFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	res, err := s.svc.ListPluginPackageDownloads(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginPackageStagingDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_package_staging_not_found", "staging 插件包不存在").WithStatus(404))
		return
	}
	record, err := s.svc.GetPluginPackageDownload(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, record)
}

func (s *Server) deleteAdminPluginPackageStaging(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_package_staging_not_found", "staging 插件包不存在").WithStatus(404))
		return
	}
	record, err := s.svc.DeletePluginPackageDownload(id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.staging.delete.failed", fmt.Sprintf("plugin-package-downloads#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"id": id, "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.staging.deleted", fmt.Sprintf("plugin-package-downloads#%d", record.ID), nil, gin.H{"status": record.Status}, gin.H{
		"id":           record.ID,
		"plugin_code":  record.PluginCode,
		"version":      record.Version,
		"staging_path": record.StagingPath,
		"actor":        auditActor(c),
	})
	c.JSON(http.StatusOK, record)
}
