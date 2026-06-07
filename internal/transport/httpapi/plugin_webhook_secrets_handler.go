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

func (s *Server) listAdminWebhookSecrets(c *gin.Context) {
	filter := domain.PluginWebhookSecretFilter{
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Status:     strings.TrimSpace(c.Query("status")),
		SecretRef:  strings.TrimSpace(c.Query("secret_ref")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	res, err := s.svc.ListPluginWebhookSecrets(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminWebhookSecretDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404))
		return
	}
	out, err := s.svc.GetPluginWebhookSecret(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

type createWebhookSecretRequest struct {
	PluginCode string `json:"plugin_code"`
	TargetURL  string `json:"target_url"`
}

func (s *Server) createAdminWebhookSecret(c *gin.Context) {
	var req createWebhookSecretRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("webhook_secret_invalid", "请求参数不合法").WithStatus(400))
		return
	}
	actor := auditActor(c)
	operator := service.WebhookSecretOperator{ID: 0, Name: actor}
	s.auditStructured(c, "system", "plugin.webhook.secret.created", "webhook-secrets", nil, gin.H{"status": "requested"}, gin.H{
		"plugin_code": strings.TrimSpace(req.PluginCode),
		"target_url":  strings.TrimSpace(req.TargetURL),
		"actor":       actor,
	})

	out, err := s.svc.CreatePluginWebhookSecret(operator, service.CreateWebhookSecretRequest{
		PluginCode: req.PluginCode,
		TargetURL:  req.TargetURL,
	})
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.secret.created", "webhook-secrets", nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"plugin_code": strings.TrimSpace(req.PluginCode),
			"target_url":  strings.TrimSpace(req.TargetURL),
			"actor":       actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) rotateAdminWebhookSecret(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	operator := service.WebhookSecretOperator{ID: 0, Name: actor}
	s.auditStructured(c, "system", "plugin.webhook.secret.rotated", fmt.Sprintf("webhook-secrets#%d", id), nil, gin.H{"status": "requested"}, gin.H{
		"secret_id": id,
		"actor":     actor,
	})
	out, err := s.svc.RotatePluginWebhookSecret(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.secret.rotated", fmt.Sprintf("webhook-secrets#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"secret_id": id,
			"actor":     actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) disableAdminWebhookSecret(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	operator := service.WebhookSecretOperator{ID: 0, Name: actor}
	s.auditStructured(c, "system", "plugin.webhook.secret.disabled", fmt.Sprintf("webhook-secrets#%d", id), nil, gin.H{"status": "requested"}, gin.H{
		"secret_id": id,
		"actor":     actor,
	})
	out, err := s.svc.DisablePluginWebhookSecret(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.secret.disabled", fmt.Sprintf("webhook-secrets#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"secret_id": id,
			"actor":     actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) enableAdminWebhookSecret(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	operator := service.WebhookSecretOperator{ID: 0, Name: actor}
	s.auditStructured(c, "system", "plugin.webhook.secret.enabled", fmt.Sprintf("webhook-secrets#%d", id), nil, gin.H{"status": "requested"}, gin.H{
		"secret_id": id,
		"actor":     actor,
	})
	out, err := s.svc.EnablePluginWebhookSecret(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.secret.enabled", fmt.Sprintf("webhook-secrets#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"secret_id": id,
			"actor":     actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) revokeAdminWebhookSecret(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_secret_not_found", "Webhook Secret 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	operator := service.WebhookSecretOperator{ID: 0, Name: actor}
	s.auditStructured(c, "system", "plugin.webhook.secret.revoked", fmt.Sprintf("webhook-secrets#%d", id), nil, gin.H{"status": "requested"}, gin.H{
		"secret_id": id,
		"actor":     actor,
	})
	out, err := s.svc.RevokePluginWebhookSecret(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.secret.revoked", fmt.Sprintf("webhook-secrets#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"secret_id": id,
			"actor":     actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
