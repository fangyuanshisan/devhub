package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (s *Server) listAdminPluginOperations(c *gin.Context) {
	pluginCode := strings.TrimSpace(c.Query("plugin_code"))
	typ := strings.TrimSpace(c.Query("operation_type"))
	status := strings.TrimSpace(c.Query("status"))
	page, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page_size", "20")))
	resp, err := s.svc.ListPluginOperations(domain.PluginOperationFilter{
		PluginCode:    pluginCode,
		OperationType: typ,
		Status:        status,
		Page:          page,
		PageSize:      pageSize,
	})
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminPluginOperationDetail(c *gin.Context) {
	operationID := strings.TrimSpace(c.Param("operation_id"))
	op, err := s.svc.GetPluginOperation(operationID)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, op)
}

func (s *Server) recoverDryRunAdminPluginOperation(c *gin.Context) {
	operationID := strings.TrimSpace(c.Param("operation_id"))
	resp, err := s.svc.RecoverPluginOperationDryRun(operationID)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) cleanupAdminPluginOperation(c *gin.Context) {
	operationID := strings.TrimSpace(c.Param("operation_id"))
	user, _ := currentUser(c)
	opName := strings.TrimSpace(user.Nickname)
	if opName == "" {
		opName = strings.TrimSpace(user.Username)
	}
	resp, err := s.svc.CleanupPluginOperation(service.PluginOperationOperator{ID: user.ID, Name: opName}, operationID)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) rollbackDryRunAdminPluginUpgrade(c *gin.Context) {
	pluginCode := strings.TrimSpace(c.Param("code"))
	var req domain.PluginUpgradeRollbackDryRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := s.svc.PluginUpgradeRollbackDryRun(pluginCode, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
