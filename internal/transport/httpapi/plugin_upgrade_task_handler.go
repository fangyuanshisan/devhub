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

func (s *Server) adminPluginUpgradeImpact(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	target := strings.TrimSpace(c.Query("target_compat_check_id"))
	id, _ := strconv.ParseInt(target, 10, 64)
	op := service.PluginUpgradeOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		op.ID = adminCtx.CurrentUser.ID
		op.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.PluginUpgradeImpact(op, code, id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) upgradeAdminPluginFromCompat(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	var req struct {
		TargetCompatCheckID int64  `json:"target_compat_check_id"`
		Reason              string `json:"reason,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "request json 不合法")
		return
	}
	op := service.PluginUpgradeOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		op.ID = adminCtx.CurrentUser.ID
		op.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.UpgradePluginFromCompatCheckAs(op, code, service.PluginUpgradeRequest{
		TargetCompatCheckID: req.TargetCompatCheckID,
		Reason:              req.Reason,
	})
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.upgrade.success", fmt.Sprintf("plugins#%s", code), nil, gin.H{"status": "upgraded"},
		gin.H{"plugin_code": code, "operation": "plugin_upgrade", "upgrade_task_id": res.ID, "old_version": res.OldVersion, "new_version": res.NewVersion})
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginUpgradeTasks(c *gin.Context) {
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	filter := domain.PluginUpgradeTaskFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		Page:       page,
		PageSize:   pageSize,
	}
	res, err := s.svc.ListPluginUpgradeTasks(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginUpgradeTaskDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	res, err := s.svc.GetPluginUpgradeTask(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) retryAdminPluginUpgradeTask(c *gin.Context) {
	id, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	op := service.PluginUpgradeOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		op.ID = adminCtx.CurrentUser.ID
		op.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}
	res, err := s.svc.RetryPluginUpgradeTaskAs(op, id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) deleteAdminPluginUpgradeTask(c *gin.Context) {
	id, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err := s.svc.DeletePluginUpgradeTask(id); err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
