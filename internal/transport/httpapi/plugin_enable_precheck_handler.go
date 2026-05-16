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

func (s *Server) runAdminPluginEnablePrecheck(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	if code == "" {
		failAPIError(c, domain.NewPluginError("plugin_enable_precheck_invalid_request", "plugin_code 不能为空").WithStatus(400))
		return
	}

	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.enable_precheck.requested", fmt.Sprintf("plugins#%s", code), nil, gin.H{"status": "requested"}, gin.H{
		"plugin_code": code,
		"actor":       actor,
	})

	operator := service.PluginEnablePrecheckOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}

	res, err := s.svc.RunPluginEnablePrecheckAs(operator, code)
	if err != nil {
		s.auditStructured(c, "system", "plugin.enable_precheck.failed", fmt.Sprintf("plugins#%s", code), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"plugin_code": code,
			"actor":       actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}

	action := "plugin.enable_precheck.success"
	switch res.Status {
	case domain.PluginEnablePrecheckStatusConfigInvalid:
		action = "plugin.enable_precheck.config_invalid"
	case domain.PluginEnablePrecheckStatusMigrationPending:
		action = "plugin.enable_precheck.migration_pending"
	case domain.PluginEnablePrecheckStatusFileIntegrityFailed, domain.PluginEnablePrecheckStatusManifestChanged:
		action = "plugin.enable_precheck.file_integrity_failed"
	case domain.PluginEnablePrecheckStatusBlocked, domain.PluginEnablePrecheckStatusFailed, domain.PluginEnablePrecheckStatusDependencyMissing, domain.PluginEnablePrecheckStatusConflictDetected:
		action = "plugin.enable_precheck.blocked"
	}
	s.auditStructured(c, "system", action, fmt.Sprintf("plugin-enable-prechecks#%d", res.ID), nil, gin.H{"status": res.Status}, gin.H{
		"enable_precheck_id": res.ID,
		"plugin_code":        res.PluginCode,
		"version":            res.Version,
		"status":             res.Status,
		"can_enable":         res.CanEnable,
		"warnings":           res.Warnings,
		"errors":             res.Errors,
		"actor":              actor,
	})
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginEnablePrechecks(c *gin.Context) {
	filter := domain.PluginEnablePrecheckFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	res, err := s.svc.ListPluginEnablePrechecks(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginEnablePrecheckDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_enable_precheck_not_found", "启用前检查记录不存在").WithStatus(404))
		return
	}
	res, err := s.svc.GetPluginEnablePrecheck(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) deleteAdminPluginEnablePrecheck(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_enable_precheck_not_found", "启用前检查记录不存在").WithStatus(404))
		return
	}
	operator := service.PluginEnablePrecheckOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.DeletePluginEnablePrecheckAs(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.enable_precheck.delete.failed", fmt.Sprintf("plugin-enable-prechecks#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"id":    id,
			"actor": auditActor(c),
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.enable_precheck.deleted", fmt.Sprintf("plugin-enable-prechecks#%d", res.ID), nil, gin.H{"status": res.Status}, gin.H{
		"id":          res.ID,
		"plugin_code": res.PluginCode,
		"version":     res.Version,
		"status":      res.Status,
		"can_enable":  res.CanEnable,
		"deleted":     true,
		"actor":       auditActor(c),
	})
	c.JSON(http.StatusOK, res)
}
