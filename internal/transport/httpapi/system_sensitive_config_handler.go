package httpapi

import (
	"net/http"

	"devhub-gin-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

func (s *Server) adminSystemSensitiveConfigStatus(c *gin.Context) {
	resp, err := s.svc.SystemSensitiveConfigStatus()
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "查看系统敏感配置状态", "system_sensitive_config",
		gin.H{},
		gin.H{
			"key_status":        resp.PluginConfigKeyring.Status,
			"current_key_id":    resp.PluginConfigKeyring.CurrentKeyID,
			"allowlist_origins": len(resp.ExternalServiceHTTP.AllowlistOrigins),
			"secret_ref_count":  resp.SecretCenter.SecretRefCount,
		},
		gin.H{"operation": "system.sensitive_config.status"})
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminExternalServiceHTTPAllowlist(c *gin.Context) {
	resp, err := s.svc.ExternalServiceHTTPAllowlist()
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminSystemEffectiveConfig(c *gin.Context) {
	resp, err := s.svc.SystemEffectiveConfig()
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "查看当前生效配置", "system_effective_config",
		gin.H{},
		gin.H{
			"external_service_count": len(resp.ExternalServices),
			"secret_count":           len(resp.Secrets),
			"allowlist_count":        len(resp.ExternalServiceHTTPAllowlist.EffectiveAllowlist),
		},
		gin.H{"operation": "system.effective_config.view"})
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminPluginExternalServiceEffectiveConfig(c *gin.Context) {
	resp, err := s.svc.PluginExternalServiceEffectiveConfig(c.Param("code"))
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) createAdminExternalServiceHTTPAllowlist(c *gin.Context) {
	var req domain.PluginExternalServiceHTTPAllowlistUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("external_service_http_allowlist_invalid", "请求参数不合法").WithStatus(http.StatusBadRequest))
		return
	}
	resp, err := s.svc.AddExternalServiceHTTPAllowlistOrigin(pluginExternalServiceOperator(c), req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) deleteAdminExternalServiceHTTPAllowlist(c *gin.Context) {
	if c.Query("confirmed") != "true" {
		failAPIError(c, domain.NewPluginError("external_service_http_allowlist_confirm_required", "删除 HTTP allowlist origin 需要确认").WithStatus(http.StatusBadRequest))
		return
	}
	resp, err := s.svc.DeleteExternalServiceHTTPAllowlistOrigin(pluginExternalServiceOperator(c), c.Param("id"))
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
