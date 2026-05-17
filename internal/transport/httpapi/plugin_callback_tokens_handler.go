package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type createPluginCallbackTokenRequest struct {
	PluginCode     string   `json:"plugin_code"`
	Name           string   `json:"name"`
	Scopes         []string `json:"scopes"`
	CommunityScope []int64  `json:"community_scope"`
	ExpiresAt      string   `json:"expires_at,omitempty"`
}

func (s *Server) listAdminPluginCallbackTokens(c *gin.Context) {
	filter := domain.PluginCallbackTokenFilter{
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Status:     strings.TrimSpace(c.Query("status")),
		Scope:      strings.TrimSpace(c.Query("scope")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	res, err := s.svc.ListPluginCallbackTokens(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginCallbackTokenDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404))
		return
	}
	out, err := s.svc.GetPluginCallbackToken(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) createAdminPluginCallbackToken(c *gin.Context) {
	var req createPluginCallbackTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("callback_token_invalid", "请求参数不合法").WithStatus(400))
		return
	}
	actor := auditActor(c)
	operator := service.CallbackTokenOperator{Type: "admin_user", ID: 0, Name: actor}
	if adminCtx, ok := currentAdminContext(c); ok && adminCtx.CurrentUser.ID > 0 {
		operator.ID = adminCtx.CurrentUser.ID
		if adminCtx.CurrentUser.Username != "" {
			operator.Name = adminCtx.CurrentUser.Username
		} else if adminCtx.CurrentUser.Nickname != "" {
			operator.Name = adminCtx.CurrentUser.Nickname
		}
	}

	s.auditStructured(c, "system", "plugin.callback_token.created", "callback-tokens", nil, gin.H{"status": "requested"}, gin.H{
		"plugin_code":     strings.TrimSpace(req.PluginCode),
		"scopes":          req.Scopes,
		"community_scope": req.CommunityScope,
		"expires_at":      strings.TrimSpace(req.ExpiresAt),
		"actor":           actor,
	})

	out, err := s.svc.CreatePluginCallbackToken(operator, service.CreatePluginCallbackTokenRequest{
		PluginCode:       req.PluginCode,
		Name:             req.Name,
		Scopes:           req.Scopes,
		CommunityScope:   req.CommunityScope,
		ExpiresAtRFC3339: req.ExpiresAt,
	})
	if err != nil {
		s.auditStructured(c, "system", "plugin.callback_token.created", "callback-tokens", nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"plugin_code": strings.TrimSpace(req.PluginCode),
			"actor":       actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) disableAdminPluginCallbackToken(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	operator := service.CallbackTokenOperator{Type: "admin_user", ID: 0, Name: actor}
	if adminCtx, ok := currentAdminContext(c); ok && adminCtx.CurrentUser.ID > 0 {
		operator.ID = adminCtx.CurrentUser.ID
	}
	s.auditStructured(c, "system", "plugin.callback_token.disabled", "callback-tokens#"+strconv.FormatInt(id, 10), nil, gin.H{"status": "requested"}, gin.H{"id": id, "actor": actor})
	out, err := s.svc.DisablePluginCallbackToken(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.callback_token.disabled", "callback-tokens#"+strconv.FormatInt(id, 10), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"id": id, "actor": actor}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) enableAdminPluginCallbackToken(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	operator := service.CallbackTokenOperator{Type: "admin_user", ID: 0, Name: actor}
	if adminCtx, ok := currentAdminContext(c); ok && adminCtx.CurrentUser.ID > 0 {
		operator.ID = adminCtx.CurrentUser.ID
	}
	s.auditStructured(c, "system", "plugin.callback_token.enabled", "callback-tokens#"+strconv.FormatInt(id, 10), nil, gin.H{"status": "requested"}, gin.H{"id": id, "actor": actor})
	out, err := s.svc.EnablePluginCallbackToken(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.callback_token.enabled", "callback-tokens#"+strconv.FormatInt(id, 10), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"id": id, "actor": actor}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

type revokePluginCallbackTokenRequest struct {
	Reason string `json:"reason"`
}

func (s *Server) revokeAdminPluginCallbackToken(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404))
		return
	}
	var req revokePluginCallbackTokenRequest
	_ = c.ShouldBindJSON(&req)
	actor := auditActor(c)
	operator := service.CallbackTokenOperator{Type: "admin_user", ID: 0, Name: actor}
	if adminCtx, ok := currentAdminContext(c); ok && adminCtx.CurrentUser.ID > 0 {
		operator.ID = adminCtx.CurrentUser.ID
	}
	s.auditStructured(c, "system", "plugin.callback_token.revoked", "callback-tokens#"+strconv.FormatInt(id, 10), nil, gin.H{"status": "requested"}, gin.H{"id": id, "actor": actor})
	out, err := s.svc.RevokePluginCallbackToken(operator, id, req.Reason)
	if err != nil {
		s.auditStructured(c, "system", "plugin.callback_token.revoked", "callback-tokens#"+strconv.FormatInt(id, 10), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"id": id, "actor": actor}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) rotateAdminPluginCallbackToken(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("callback_token_not_found", "Callback Token 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	operator := service.CallbackTokenOperator{Type: "admin_user", ID: 0, Name: actor}
	if adminCtx, ok := currentAdminContext(c); ok && adminCtx.CurrentUser.ID > 0 {
		operator.ID = adminCtx.CurrentUser.ID
	}
	s.auditStructured(c, "system", "plugin.callback_token.rotated", "callback-tokens#"+strconv.FormatInt(id, 10), nil, gin.H{"status": "requested"}, gin.H{"id": id, "actor": actor})
	out, err := s.svc.RotatePluginCallbackToken(operator, id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.callback_token.rotated", "callback-tokens#"+strconv.FormatInt(id, 10), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{"id": id, "actor": actor}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
