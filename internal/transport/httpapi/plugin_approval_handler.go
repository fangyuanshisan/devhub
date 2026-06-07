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

// ===== Plugin approvals (v1.5.0-P1-07) =====

func (s *Server) createAdminPluginApproval(c *gin.Context) {
	var req domain.PluginApprovalCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	operator := service.PluginApprovalOperator{ID: adminCtx.CurrentUser.ID, Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)}
	res, err := s.svc.CreatePluginApproval(operator, req)
	if err != nil {
		s.auditStructured(c, "system", "plugin.approval.create.failed", "plugin-approvals", nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"action": req.Action, "plugin_code": req.PluginCode, "package_path": req.PackagePath, "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.approval.created", fmt.Sprintf("plugin-approvals#%d", res.ID), nil, gin.H{"status": res.Status},
		gin.H{"approval_id": res.ID, "action": res.Action, "plugin_code": res.PluginCode, "package_path": res.PackagePath, "risk_level": res.PackageRiskLevel, "checksum_status": res.PackageChecksumStatus, "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginApprovals(c *gin.Context) {
	filter := domain.PluginApprovalFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		Action:     strings.TrimSpace(c.Query("action")),
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	if v := strings.TrimSpace(c.Query("requested_by")); v != "" {
		filter.RequestedBy, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := strings.TrimSpace(c.Query("reviewed_by")); v != "" {
		filter.ReviewedBy, _ = strconv.ParseInt(v, 10, 64)
	}
	res, err := s.svc.ListPluginApprovals(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginApprovalDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_approval_not_found", "审批记录不存在").WithStatus(404))
		return
	}
	it, err := s.svc.GetPluginApproval(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.PluginApprovalDetailResponse{Request: it})
}

func (s *Server) approveAdminPluginApproval(c *gin.Context) {
	id, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	var req domain.PluginApprovalReviewRequest
	_ = c.ShouldBindJSON(&req)
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	operator := service.PluginApprovalOperator{ID: adminCtx.CurrentUser.ID, Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)}
	res, err := s.svc.ApprovePluginApproval(operator, id, req.Comment)
	if err != nil {
		s.auditStructured(c, "system", "plugin.approval.approve.failed", fmt.Sprintf("plugin-approvals#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"approval_id": id, "comment": strings.TrimSpace(req.Comment), "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.approval.approved", fmt.Sprintf("plugin-approvals#%d", id), nil, gin.H{"status": res.Status},
		gin.H{"approval_id": id, "action": res.Action, "plugin_code": res.PluginCode, "comment": strings.TrimSpace(req.Comment), "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) rejectAdminPluginApproval(c *gin.Context) {
	id, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	var req domain.PluginApprovalReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	operator := service.PluginApprovalOperator{ID: adminCtx.CurrentUser.ID, Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)}
	res, err := s.svc.RejectPluginApproval(operator, id, req.Comment)
	if err != nil {
		s.auditStructured(c, "system", "plugin.approval.reject.failed", fmt.Sprintf("plugin-approvals#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"approval_id": id, "comment": strings.TrimSpace(req.Comment), "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.approval.rejected", fmt.Sprintf("plugin-approvals#%d", id), nil, gin.H{"status": res.Status},
		gin.H{"approval_id": id, "action": res.Action, "plugin_code": res.PluginCode, "comment": strings.TrimSpace(req.Comment), "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) cancelAdminPluginApproval(c *gin.Context) {
	id, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	var req domain.PluginApprovalCancelRequest
	_ = c.ShouldBindJSON(&req)
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	operator := service.PluginApprovalOperator{ID: adminCtx.CurrentUser.ID, Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)}
	allowManage := hasPermission(adminCtx.CurrentUser.Permissions, "plugin.approve")
	res, err := s.svc.CancelPluginApproval(operator, id, req.Comment, allowManage)
	if err != nil {
		s.auditStructured(c, "system", "plugin.approval.cancel.failed", fmt.Sprintf("plugin-approvals#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"approval_id": id, "comment": strings.TrimSpace(req.Comment), "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.approval.canceled", fmt.Sprintf("plugin-approvals#%d", id), nil, gin.H{"status": res.Status},
		gin.H{"approval_id": id, "action": res.Action, "plugin_code": res.PluginCode, "comment": strings.TrimSpace(req.Comment), "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}

func (s *Server) executeAdminPluginApproval(c *gin.Context) {
	id, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		fail(c, http.StatusUnauthorized, "未登录")
		return
	}
	operator := service.PluginApprovalOperator{ID: adminCtx.CurrentUser.ID, Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)}
	res, err := s.svc.ExecutePluginApproval(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.approval.execute.failed", fmt.Sprintf("plugin-approvals#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"approval_id": id, "actor": auditActor(c), "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.approval.executed", fmt.Sprintf("plugin-approvals#%d", id), nil, gin.H{"status": res.Status},
		gin.H{"approval_id": id, "action": res.Action, "plugin_code": res.PluginCode, "package_path": res.PackagePath, "risk_level": res.PackageRiskLevel, "actor": auditActor(c)})
	c.JSON(http.StatusOK, res)
}
