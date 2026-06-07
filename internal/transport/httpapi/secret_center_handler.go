package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (s *Server) adminListSecretRefs(c *gin.Context) {
	filter := domain.SecretRefFilter{
		Namespace: strings.TrimSpace(c.Query("namespace")),
	}
	if p := strings.TrimSpace(c.DefaultQuery("page", "1")); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			filter.Page = v
		}
	}
	if p := strings.TrimSpace(c.DefaultQuery("page_size", "20")); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			filter.PageSize = v
		}
	}
	resp, err := s.svc.ListSecretMetadata(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "查看 SecretCenter 列表", "secret_center",
		gin.H{"namespace": filter.Namespace},
		gin.H{"count": len(resp.Items)},
		gin.H{"operation": "secret_center.secret_refs.list"})
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminGetSecretRefMetadata(c *gin.Context) {
	ref := strings.TrimSpace(c.Query("ref"))
	it, err := s.svc.GetSecretMetadata(ref)
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "查看 SecretCenter 明细", "secret_center",
		gin.H{"ref": ref},
		gin.H{"status": it.Status, "key_id": it.KeyID},
		gin.H{"operation": "secret_center.secret_refs.get"})
	c.JSON(http.StatusOK, it)
}

func (s *Server) adminGetSecretRefDetail(c *gin.Context) {
	ref := strings.TrimSpace(c.Query("ref"))
	it, err := s.svc.GetSecretDetail(ref)
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "查看 SecretCenter 详情", "secret_center",
		gin.H{"ref": ref},
		gin.H{"status": it.Record.Status, "source_type": it.Source.Type, "plugin_code": it.Source.PluginCode},
		gin.H{"operation": "secret_center.secret_refs.detail"})
	c.JSON(http.StatusOK, it)
}

func (s *Server) adminGetSecretRefUsages(c *gin.Context) {
	ref := strings.TrimSpace(c.Query("ref"))
	items := s.svc.SecretUsages(ref)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) adminCreateSecretRef(c *gin.Context) {
	var req domain.SecretCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("secret_center_invalid", "请求参数不合法").WithStatus(http.StatusBadRequest))
		return
	}
	op := secretOperatorFromContext(c)
	out, err := s.svc.CreateSecret(op, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) adminUpdateSecretRef(c *gin.Context) {
	var req domain.SecretUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("secret_center_invalid", "请求参数不合法").WithStatus(http.StatusBadRequest))
		return
	}
	op := secretOperatorFromContext(c)
	out, err := s.svc.UpdateSecret(op, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) adminDisableSecretRef(c *gin.Context) {
	var req domain.SecretStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("secret_center_invalid", "请求参数不合法").WithStatus(http.StatusBadRequest))
		return
	}
	op := secretOperatorFromContext(c)
	out, err := s.svc.DisableSecret(op, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) adminPreviewDisableSecretRef(c *gin.Context) {
	var req domain.SecretStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("secret_center_invalid", "请求参数不合法").WithStatus(http.StatusBadRequest))
		return
	}
	out, err := s.svc.SecretImpactPreview(req.Ref, "disable")
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) adminRevokeSecretRef(c *gin.Context) {
	var req domain.SecretStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("secret_center_invalid", "请求参数不合法").WithStatus(http.StatusBadRequest))
		return
	}
	op := secretOperatorFromContext(c)
	out, err := s.svc.RevokeSecret(op, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) adminPreviewRevokeSecretRef(c *gin.Context) {
	var req domain.SecretStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("secret_center_invalid", "请求参数不合法").WithStatus(http.StatusBadRequest))
		return
	}
	out, err := s.svc.SecretImpactPreview(req.Ref, "revoke")
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func secretOperatorFromContext(c *gin.Context) service.SecretOperator {
	op := service.SecretOperator{Type: "admin_user", Name: auditActor(c)}
	if adminCtx, ok := currentAdminContext(c); ok {
		op.ID = adminCtx.CurrentUser.ID
		if adminCtx.CurrentUser.Username != "" {
			op.Name = adminCtx.CurrentUser.Username
		} else if adminCtx.CurrentUser.Nickname != "" {
			op.Name = adminCtx.CurrentUser.Nickname
		}
	}
	if strings.TrimSpace(op.Name) == "" {
		op.Name = "system"
	}
	return op
}

func requireSecretManage() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			fail(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}
		if user.TokenType == "admin" && (user.RoleCode == "super_admin" || hasPermission(user.Permissions, "*") || hasPermission(user.Permissions, "secret.manage") || hasPermission(user.Permissions, "system.write") || hasPermission(user.Permissions, "plugin.manage")) {
			c.Next()
			return
		}
		failAPIError(c, domain.NewAPIError("secret_center_permission_denied", "缺少 SecretCenter 危险操作权限").
			WithStatus(http.StatusForbidden).
			WithDetail("permission_code", "secret.manage").
			WithSuggestion("禁用或吊销 Secret 需要 super_admin、secret.manage、system.write 或 plugin.manage 权限。"))
		c.Abort()
	}
}
