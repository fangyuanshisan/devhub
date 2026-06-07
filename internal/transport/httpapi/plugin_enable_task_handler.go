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

func (s *Server) enablePluginFromPrecheck(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_enable_precheck_not_found", "启用前检查记录不存在").WithStatus(404))
		return
	}

	operator := service.PluginEnableOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}

	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.enable.requested", fmt.Sprintf("plugin-enable-prechecks#%d", id), nil, gin.H{"status": "requested"}, gin.H{
		"enable_precheck_id": id,
		"actor":              actor,
	})

	res, err := s.svc.EnablePluginFromEnablePrecheckAs(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.enable.failed", fmt.Sprintf("plugin-enable-prechecks#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"enable_precheck_id": id,
			"actor":              actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginEnableTasks(c *gin.Context) {
	filter := domain.PluginEnableTaskFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	res, err := s.svc.ListPluginEnableTasks(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginEnableTaskDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_enable_task_not_found", "启用任务不存在").WithStatus(404))
		return
	}
	res, err := s.svc.GetPluginEnableTask(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) retryAdminPluginEnableTask(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_enable_task_not_found", "启用任务不存在").WithStatus(404))
		return
	}
	operator := service.PluginEnableOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.RetryPluginEnableTaskAs(operator, id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) deleteAdminPluginEnableTask(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_enable_task_not_found", "启用任务不存在").WithStatus(404))
		return
	}
	operator := service.PluginEnableOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.DeletePluginEnableTaskAs(operator, id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
