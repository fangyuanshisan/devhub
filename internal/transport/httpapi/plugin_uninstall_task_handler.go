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

func (s *Server) adminPluginUninstallImpact(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	impact, err := s.svc.PluginUninstallImpact(code)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, impact)
}

func (s *Server) softUninstallAdminPlugin(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	var req map[string]any
	_ = c.ShouldBindJSON(&req)

	operator := service.PluginUninstallOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	actor := auditActor(c)

	s.auditStructured(c, "system", "plugin.soft_uninstall.requested", fmt.Sprintf("plugins#%s", code), nil, gin.H{"status": "requested"}, gin.H{
		"plugin_code": code,
		"actor":       actor,
	})

	res, err := s.svc.SoftUninstallPluginAs(operator, code, req)
	if err != nil {
		s.auditStructured(c, "system", "plugin.soft_uninstall.failed", fmt.Sprintf("plugins#%s", code), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"plugin_code": code,
			"actor":       actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginUninstallTasks(c *gin.Context) {
	filter := domain.PluginUninstallTaskFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	res, err := s.svc.ListPluginUninstallTasks(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginUninstallTaskDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_uninstall_task_not_found", "软卸载任务不存在").WithStatus(404))
		return
	}
	res, err := s.svc.GetPluginUninstallTask(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) retryAdminPluginUninstallTask(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_uninstall_task_not_found", "软卸载任务不存在").WithStatus(404))
		return
	}
	operator := service.PluginUninstallOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.RetryPluginUninstallTaskAs(operator, id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) deleteAdminPluginUninstallTask(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_uninstall_task_not_found", "软卸载任务不存在").WithStatus(404))
		return
	}
	operator := service.PluginUninstallOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.DeletePluginUninstallTaskAs(operator, id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
