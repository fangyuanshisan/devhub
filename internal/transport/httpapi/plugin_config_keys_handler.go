package httpapi

import (
	"net/http"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (s *Server) adminPluginConfigKeyStatus(c *gin.Context) {
	resp, err := s.svc.PluginConfigKeyStatus()
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "查看插件配置密钥状态", "plugin_config_keys",
		gin.H{},
		gin.H{"current_key_id": resp.CurrentKeyID, "key_count": resp.KeyCount, "status": resp.Status},
		gin.H{"operation": "plugin.config_key.status"})
	c.JSON(http.StatusOK, resp)
}

func (s *Server) dryRunAdminPluginConfigKeyRotation(c *gin.Context) {
	var req domain.PluginConfigKeyRotationDryRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := s.svc.PluginConfigKeyRotationDryRun(req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "插件配置密钥轮换预检", "plugin_config_keys",
		gin.H{"scope": strings.TrimSpace(req.Scope), "plugin_code": strings.TrimSpace(req.PluginCode), "community_id": req.CommunityID, "include_config_versions": req.IncludeConfigVersions},
		gin.H{"status": resp.Status, "needs_reencrypt": resp.Summary.NeedsReencrypt, "decrypt_failed": resp.Summary.DecryptFailed, "missing_key": resp.Summary.MissingKey},
		gin.H{"operation": "plugin.config_key.rotation.dry_run"})
	c.JSON(http.StatusOK, resp)
}

func (s *Server) reencryptAdminPluginConfigKeys(c *gin.Context) {
	var req domain.PluginConfigKeyRotationReencryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}

	operator := service.PluginConfigVersionOperator{Type: "admin_user", ID: 0, Name: auditActor(c)}
	if adminCtx, ok := currentAdminContext(c); ok && adminCtx.CurrentUser.ID > 0 {
		operator.ID = adminCtx.CurrentUser.ID
		if adminCtx.CurrentUser.Username != "" {
			operator.Name = adminCtx.CurrentUser.Username
		} else if adminCtx.CurrentUser.Nickname != "" {
			operator.Name = adminCtx.CurrentUser.Nickname
		}
	}

	resp, err := s.svc.PluginConfigKeyRotationReencrypt(req, operator)
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "执行插件配置密钥轮换 re-encrypt", "plugin_config_keys",
		gin.H{"scope": strings.TrimSpace(req.Scope), "plugin_code": strings.TrimSpace(req.PluginCode), "community_id": req.CommunityID, "include_config_versions": req.IncludeConfigVersions},
		gin.H{"status": resp.Status, "updated_count": resp.UpdatedCount, "current_key_id": resp.CurrentKeyID},
		gin.H{"operation": "plugin.config_key.rotation.re_encrypt"})
	c.JSON(http.StatusOK, resp)
}
