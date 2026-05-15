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

func (s *Server) runAdminPluginPackageCompatCheck(c *gin.Context) {
	precheckID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || precheckID <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_package_precheck_not_found", "插件包预检记录不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.package.compat_check.requested", fmt.Sprintf("plugin-package-prechecks#%d", precheckID), nil, gin.H{"status": "requested"}, gin.H{
		"package_precheck_id": precheckID,
		"actor":               actor,
	})
	operator := service.PluginPackageCompatOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.RunPluginPackageCompatCheckAs(operator, precheckID)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.compat_check.failed", fmt.Sprintf("plugin-package-prechecks#%d", precheckID), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"package_precheck_id": precheckID,
			"actor":               actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	action := "plugin.package.compat_check.success"
	switch res.Status {
	case domain.PluginPackageCompatCheckStatusIncompatible:
		action = "plugin.package.compat_check.incompatible"
	case domain.PluginPackageCompatCheckStatusDependencyMissing, domain.PluginPackageCompatCheckStatusDependencyVersionMismatch:
		action = "plugin.package.compat_check.dependency_missing"
	case domain.PluginPackageCompatCheckStatusConflictDetected:
		action = "plugin.package.compat_check.conflict_detected"
	case domain.PluginPackageCompatCheckStatusFailed:
		action = "plugin.package.compat_check.failed"
	}
	s.auditStructured(c, "system", action, fmt.Sprintf("plugin-package-compat-checks#%d", res.ID), nil, gin.H{"status": res.Status}, gin.H{
		"package_download_id":     res.PackageDownloadID,
		"package_precheck_id":     res.PackagePrecheckID,
		"plugin_code":             res.PluginCode,
		"version":                 res.Version,
		"status":                  res.Status,
		"can_install":             res.CanInstall,
		"core_version":            res.CoreVersion,
		"compatible_core_version": res.CompatibleCoreVersion,
		"blockers":                res.Errors,
		"warnings":                res.Warnings,
		"actor":                   actor,
	})
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginPackageCompatChecks(c *gin.Context) {
	filter := domain.PluginPackageCompatCheckFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	if raw := strings.TrimSpace(c.Query("package_precheck_id")); raw != "" {
		filter.PackagePrecheckID, _ = strconv.ParseInt(raw, 10, 64)
	}
	res, err := s.svc.ListPluginPackageCompatChecks(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginPackageCompatCheckDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_package_compat_check_not_found", "兼容性检查记录不存在").WithStatus(404))
		return
	}
	res, err := s.svc.GetPluginPackageCompatCheck(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) deleteAdminPluginPackageCompatCheck(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_package_compat_check_not_found", "兼容性检查记录不存在").WithStatus(404))
		return
	}
	res, err := s.svc.DeletePluginPackageCompatCheck(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.compat_check.deleted", fmt.Sprintf("plugin-package-compat-checks#%d", res.ID), nil, gin.H{"status": res.Status}, gin.H{
		"id":                  res.ID,
		"package_precheck_id": res.PackagePrecheckID,
		"plugin_code":         res.PluginCode,
		"version":             res.Version,
		"actor":               auditActor(c),
	})
	c.JSON(http.StatusOK, res)
}
