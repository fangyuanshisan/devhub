package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	pluginCallbackCtxToken = "plugin_callback_token"
)

type pluginCallbackIdentity struct {
	Token       domain.PluginCallbackToken
	Scopes      []string
	Communities []int64
}

func (s *Server) pluginCallbackAuthRequired(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := strings.TrimSpace(c.GetHeader("X-DevHub-Request-ID"))
		if requestID == "" {
			requestID = "cbreq_" + strconv.FormatInt(time.Now().UnixNano(), 10)
		}
		pluginCodeHeader := strings.TrimSpace(c.GetHeader("X-DevHub-Plugin-Code"))
		tokenRefHeader := strings.TrimSpace(c.GetHeader("X-DevHub-Token-Ref"))
		communityIDHeader := strings.TrimSpace(c.GetHeader("X-DevHub-Community-ID"))

		tokenPlain := bearerToken(c.GetHeader("Authorization"))
		if tokenPlain == "" {
			s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
				RequestID:      requestID,
				PluginCode:     firstNonEmpty(pluginCodeHeader, "unknown"),
				TokenRef:       tokenRefHeader,
				APIPath:        c.FullPath(),
				Method:         c.Request.Method,
				ScopeRequired:  requiredScope,
				Status:         domain.PluginCallbackRequestStatusRejected,
				ResponseStatus: http.StatusUnauthorized,
				ErrorCode:      "TOKEN_MISSING",
				ErrorMessage:   "缺少 token",
				CommunityID:    parseInt64OrZero(communityIDHeader),
				ActorType:      "plugin_service",
				ActorID:        0,
				IPAddress:      c.ClientIP(),
				UserAgent:      truncateString(c.Request.UserAgent(), 255),
				DurationMS:     int64(time.Since(start).Milliseconds()),
			})
			failAPIError(c, domain.NewPluginError("TOKEN_MISSING", "缺少 token").WithStatus(http.StatusUnauthorized))
			c.Abort()
			return
		}
		if !strings.HasPrefix(strings.TrimSpace(tokenPlain), "cbsk_") {
			s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
				RequestID:      requestID,
				PluginCode:     firstNonEmpty(pluginCodeHeader, "unknown"),
				TokenRef:       tokenRefHeader,
				APIPath:        c.FullPath(),
				Method:         c.Request.Method,
				ScopeRequired:  requiredScope,
				Status:         domain.PluginCallbackRequestStatusRejected,
				ResponseStatus: http.StatusUnauthorized,
				ErrorCode:      "TOKEN_INVALID",
				ErrorMessage:   "token 格式不合法",
				CommunityID:    parseInt64OrZero(communityIDHeader),
				ActorType:      "plugin_service",
				ActorID:        0,
				IPAddress:      c.ClientIP(),
				UserAgent:      truncateString(c.Request.UserAgent(), 255),
				DurationMS:     int64(time.Since(start).Milliseconds()),
			})
			failAPIError(c, domain.NewPluginError("TOKEN_INVALID", "token 不合法").WithStatus(http.StatusUnauthorized))
			c.Abort()
			return
		}

		tokenHash := s.svc.HashCallbackBearerToken(tokenPlain)
		rec, ok := s.svc.PluginCallbackTokenByHash(tokenHash)
		if !ok || rec.ID == 0 {
			s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
				RequestID:      requestID,
				PluginCode:     firstNonEmpty(pluginCodeHeader, "unknown"),
				TokenRef:       tokenRefHeader,
				APIPath:        c.FullPath(),
				Method:         c.Request.Method,
				ScopeRequired:  requiredScope,
				Status:         domain.PluginCallbackRequestStatusRejected,
				ResponseStatus: http.StatusUnauthorized,
				ErrorCode:      "TOKEN_INVALID",
				ErrorMessage:   "token 不存在",
				CommunityID:    parseInt64OrZero(communityIDHeader),
				ActorType:      "plugin_service",
				ActorID:        0,
				IPAddress:      c.ClientIP(),
				UserAgent:      truncateString(c.Request.UserAgent(), 255),
				DurationMS:     int64(time.Since(start).Milliseconds()),
			})
			failAPIError(c, domain.NewPluginError("TOKEN_INVALID", "token 不存在").WithStatus(http.StatusUnauthorized))
			c.Abort()
			return
		}
		// Never expose token_hash downstream.
		rec.TokenHash = ""

		// status / expires_at checks.
		status := strings.TrimSpace(rec.Status)
		if status == domain.PluginCallbackTokenStatusDisabled {
			s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
				RequestID:      requestID,
				PluginCode:     rec.PluginCode,
				TokenRef:       rec.TokenRef,
				APIPath:        c.FullPath(),
				Method:         c.Request.Method,
				ScopeRequired:  requiredScope,
				Status:         domain.PluginCallbackRequestStatusRejected,
				ResponseStatus: http.StatusUnauthorized,
				ErrorCode:      "TOKEN_DISABLED",
				ErrorMessage:   "token 已禁用",
				CommunityID:    parseInt64OrZero(communityIDHeader),
				ActorType:      "plugin_service",
				ActorID:        0,
				IPAddress:      c.ClientIP(),
				UserAgent:      truncateString(c.Request.UserAgent(), 255),
				DurationMS:     int64(time.Since(start).Milliseconds()),
			})
			s.svc.AppendAdminLog(domain.AdminLog{
				Site:      "admin",
				Actor:     "system",
				ActorType: "system",
				ActorID:   0,
				Action:    "plugin.callback.request.rejected",
				Target:    "plugin-callback#" + rec.PluginCode,
				Metadata:  auditJSON(gin.H{"plugin_code": rec.PluginCode, "token_ref": rec.TokenRef, "error_code": "TOKEN_DISABLED", "api_path": c.FullPath(), "method": c.Request.Method, "request_id": requestID}),
				CreatedAt: service.Now(),
			})
			failAPIError(c, domain.NewPluginError("TOKEN_DISABLED", "token 已禁用").WithStatus(http.StatusUnauthorized))
			c.Abort()
			return
		}
		if status == domain.PluginCallbackTokenStatusRevoked {
			s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
				RequestID:      requestID,
				PluginCode:     rec.PluginCode,
				TokenRef:       rec.TokenRef,
				APIPath:        c.FullPath(),
				Method:         c.Request.Method,
				ScopeRequired:  requiredScope,
				Status:         domain.PluginCallbackRequestStatusRejected,
				ResponseStatus: http.StatusUnauthorized,
				ErrorCode:      "TOKEN_REVOKED",
				ErrorMessage:   "token 已吊销",
				CommunityID:    parseInt64OrZero(communityIDHeader),
				ActorType:      "plugin_service",
				ActorID:        0,
				IPAddress:      c.ClientIP(),
				UserAgent:      truncateString(c.Request.UserAgent(), 255),
				DurationMS:     int64(time.Since(start).Milliseconds()),
			})
			s.svc.AppendAdminLog(domain.AdminLog{
				Site:      "admin",
				Actor:     "system",
				ActorType: "system",
				ActorID:   0,
				Action:    "plugin.callback.request.rejected",
				Target:    "plugin-callback#" + rec.PluginCode,
				Metadata:  auditJSON(gin.H{"plugin_code": rec.PluginCode, "token_ref": rec.TokenRef, "error_code": "TOKEN_REVOKED", "api_path": c.FullPath(), "method": c.Request.Method, "request_id": requestID}),
				CreatedAt: service.Now(),
			})
			failAPIError(c, domain.NewPluginError("TOKEN_REVOKED", "token 已吊销").WithStatus(http.StatusUnauthorized))
			c.Abort()
			return
		}
		if status == domain.PluginCallbackTokenStatusExpired {
			s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
				RequestID:      requestID,
				PluginCode:     rec.PluginCode,
				TokenRef:       rec.TokenRef,
				APIPath:        c.FullPath(),
				Method:         c.Request.Method,
				ScopeRequired:  requiredScope,
				Status:         domain.PluginCallbackRequestStatusRejected,
				ResponseStatus: http.StatusUnauthorized,
				ErrorCode:      "TOKEN_EXPIRED",
				ErrorMessage:   "token 已过期",
				CommunityID:    parseInt64OrZero(communityIDHeader),
				ActorType:      "plugin_service",
				ActorID:        0,
				IPAddress:      c.ClientIP(),
				UserAgent:      truncateString(c.Request.UserAgent(), 255),
				DurationMS:     int64(time.Since(start).Milliseconds()),
			})
			s.svc.AppendAdminLog(domain.AdminLog{
				Site:      "admin",
				Actor:     "system",
				ActorType: "system",
				ActorID:   0,
				Action:    "plugin.callback.request.rejected",
				Target:    "plugin-callback#" + rec.PluginCode,
				Metadata:  auditJSON(gin.H{"plugin_code": rec.PluginCode, "token_ref": rec.TokenRef, "error_code": "TOKEN_EXPIRED", "api_path": c.FullPath(), "method": c.Request.Method, "request_id": requestID}),
				CreatedAt: service.Now(),
			})
			failAPIError(c, domain.NewPluginError("TOKEN_EXPIRED", "token 已过期").WithStatus(http.StatusUnauthorized))
			c.Abort()
			return
		}
		if strings.TrimSpace(rec.ExpiresAt) != "" {
			if t, ok := service.ParseTimeLayout(rec.ExpiresAt); ok && time.Now().After(t) {
				s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
					RequestID:      requestID,
					PluginCode:     rec.PluginCode,
					TokenRef:       rec.TokenRef,
					APIPath:        c.FullPath(),
					Method:         c.Request.Method,
					ScopeRequired:  requiredScope,
					Status:         domain.PluginCallbackRequestStatusRejected,
					ResponseStatus: http.StatusUnauthorized,
					ErrorCode:      "TOKEN_EXPIRED",
					ErrorMessage:   "token 已过期",
					CommunityID:    parseInt64OrZero(communityIDHeader),
					ActorType:      "plugin_service",
					ActorID:        0,
					IPAddress:      c.ClientIP(),
					UserAgent:      truncateString(c.Request.UserAgent(), 255),
					DurationMS:     int64(time.Since(start).Milliseconds()),
				})
				s.svc.AppendAdminLog(domain.AdminLog{
					Site:      "admin",
					Actor:     "system",
					ActorType: "system",
					ActorID:   0,
					Action:    "plugin.callback.request.rejected",
					Target:    "plugin-callback#" + rec.PluginCode,
					Metadata:  auditJSON(gin.H{"plugin_code": rec.PluginCode, "token_ref": rec.TokenRef, "error_code": "TOKEN_EXPIRED", "api_path": c.FullPath(), "method": c.Request.Method, "request_id": requestID}),
					CreatedAt: service.Now(),
				})
				failAPIError(c, domain.NewPluginError("TOKEN_EXPIRED", "token 已过期").WithStatus(http.StatusUnauthorized))
				c.Abort()
				return
			}
		}

		// Plugin must be globally enabled.
		if err := s.svc.ShouldAllowCallbackForPlugin(rec.PluginCode); err != nil {
			s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
				RequestID:      requestID,
				PluginCode:     rec.PluginCode,
				TokenRef:       rec.TokenRef,
				APIPath:        c.FullPath(),
				Method:         c.Request.Method,
				ScopeRequired:  requiredScope,
				Status:         domain.PluginCallbackRequestStatusRejected,
				ResponseStatus: http.StatusForbidden,
				ErrorCode:      "PLUGIN_DISABLED",
				ErrorMessage:   "插件未启用",
				CommunityID:    parseInt64OrZero(communityIDHeader),
				ActorType:      "plugin_service",
				ActorID:        0,
				IPAddress:      c.ClientIP(),
				UserAgent:      truncateString(c.Request.UserAgent(), 255),
				DurationMS:     int64(time.Since(start).Milliseconds()),
			})
			s.svc.AppendAdminLog(domain.AdminLog{
				Site:      "admin",
				Actor:     "system",
				ActorType: "system",
				ActorID:   0,
				Action:    "plugin.callback.plugin_disabled",
				Target:    "plugin-callback#" + rec.PluginCode,
				Metadata:  auditJSON(gin.H{"plugin_code": rec.PluginCode, "token_ref": rec.TokenRef, "api_path": c.FullPath(), "method": c.Request.Method, "request_id": requestID}),
				CreatedAt: service.Now(),
			})
			failAPIError(c, err)
			c.Abort()
			return
		}

		// Update token last_used.
		_ = s.svc.TouchPluginCallbackTokenUsage(rec.ID, strings.TrimSpace(c.ClientIP()))

		// Parse scopes/community scope.
		scopes := []string{}
		_ = json.Unmarshal([]byte(firstNonEmpty(strings.TrimSpace(rec.ScopesJSON), "[]")), &scopes)
		communities := []int64{}
		_ = json.Unmarshal([]byte(firstNonEmpty(strings.TrimSpace(rec.CommunityScopeJSON), "[]")), &communities)

		// required scope.
		if strings.TrimSpace(requiredScope) != "" {
			has := false
			for _, sc := range scopes {
				if strings.TrimSpace(sc) == requiredScope {
					has = true
					break
				}
			}
			if !has {
				s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
					RequestID:      requestID,
					PluginCode:     rec.PluginCode,
					TokenRef:       rec.TokenRef,
					APIPath:        c.FullPath(),
					Method:         c.Request.Method,
					ScopeRequired:  requiredScope,
					Status:         domain.PluginCallbackRequestStatusRejected,
					ResponseStatus: http.StatusForbidden,
					ErrorCode:      "SCOPE_DENIED",
					ErrorMessage:   "scope 不足",
					CommunityID:    parseInt64OrZero(communityIDHeader),
					ActorType:      "plugin_service",
					ActorID:        0,
					IPAddress:      c.ClientIP(),
					UserAgent:      truncateString(c.Request.UserAgent(), 255),
					DurationMS:     int64(time.Since(start).Milliseconds()),
				})
				// Audit (no token plaintext/hash).
				s.svc.AppendAdminLog(domain.AdminLog{
					Site:      "admin",
					Actor:     "system",
					ActorType: "system",
					ActorID:   0,
					Action:    "plugin.callback.scope.denied",
					Target:    "plugin-callback#" + rec.PluginCode,
					Metadata:  auditJSON(gin.H{"plugin_code": rec.PluginCode, "token_ref": rec.TokenRef, "scope_required": requiredScope, "api_path": c.FullPath(), "method": c.Request.Method, "request_id": requestID}),
					CreatedAt: service.Now(),
				})
				failAPIError(c, domain.NewPluginError("SCOPE_DENIED", "scope 不足").WithStatus(http.StatusForbidden))
				c.Abort()
				return
			}
		}

		// Community scope (optional). If provided in request, it must match token communities.
		if rawCID := strings.TrimSpace(firstNonEmpty(communityIDHeader, c.Query("community_id"))); rawCID != "" {
			cid, _ := strconv.ParseInt(rawCID, 10, 64)
			if cid > 0 {
				allowed := false
				for _, id := range communities {
					if id == cid {
						allowed = true
						break
					}
				}
				if !allowed {
					s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
						RequestID:      requestID,
						PluginCode:     rec.PluginCode,
						TokenRef:       rec.TokenRef,
						APIPath:        c.FullPath(),
						Method:         c.Request.Method,
						ScopeRequired:  requiredScope,
						Status:         domain.PluginCallbackRequestStatusRejected,
						ResponseStatus: http.StatusForbidden,
						ErrorCode:      "COMMUNITY_SCOPE_DENIED",
						ErrorMessage:   "community_scope 不匹配",
						CommunityID:    cid,
						ActorType:      "plugin_service",
						ActorID:        0,
						IPAddress:      c.ClientIP(),
						UserAgent:      truncateString(c.Request.UserAgent(), 255),
						DurationMS:     int64(time.Since(start).Milliseconds()),
					})
					s.svc.AppendAdminLog(domain.AdminLog{
						Site:      "admin",
						Actor:     "system",
						ActorType: "system",
						ActorID:   0,
						Action:    "plugin.callback.community_scope.denied",
						Target:    "plugin-callback#" + rec.PluginCode,
						Metadata:  auditJSON(gin.H{"plugin_code": rec.PluginCode, "token_ref": rec.TokenRef, "community_id": cid, "api_path": c.FullPath(), "method": c.Request.Method, "request_id": requestID}),
						CreatedAt: service.Now(),
					})
					failAPIError(c, domain.NewPluginError("COMMUNITY_SCOPE_DENIED", "community_scope 不匹配").WithStatus(http.StatusForbidden))
					c.Abort()
					return
				}
				// Additionally require plugin enabled for that community.
				if !s.svc.IsPluginEnabledForCommunity(cid, rec.PluginCode) {
					s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
						RequestID:      requestID,
						PluginCode:     rec.PluginCode,
						TokenRef:       rec.TokenRef,
						APIPath:        c.FullPath(),
						Method:         c.Request.Method,
						ScopeRequired:  requiredScope,
						Status:         domain.PluginCallbackRequestStatusRejected,
						ResponseStatus: http.StatusForbidden,
						ErrorCode:      "PLUGIN_DISABLED",
						ErrorMessage:   "插件在该社区未启用",
						CommunityID:    cid,
						ActorType:      "plugin_service",
						ActorID:        0,
						IPAddress:      c.ClientIP(),
						UserAgent:      truncateString(c.Request.UserAgent(), 255),
						DurationMS:     int64(time.Since(start).Milliseconds()),
					})
					s.svc.AppendAdminLog(domain.AdminLog{
						Site:      "admin",
						Actor:     "system",
						ActorType: "system",
						ActorID:   0,
						Action:    "plugin.callback.plugin_disabled",
						Target:    "plugin-callback#" + rec.PluginCode,
						Metadata:  auditJSON(gin.H{"plugin_code": rec.PluginCode, "token_ref": rec.TokenRef, "community_id": cid, "api_path": c.FullPath(), "method": c.Request.Method, "request_id": requestID}),
						CreatedAt: service.Now(),
					})
					failAPIError(c, domain.NewPluginError("PLUGIN_DISABLED", "插件未启用").WithStatus(http.StatusForbidden))
					c.Abort()
					return
				}
			}
		}

		c.Set(pluginCallbackCtxToken, pluginCallbackIdentity{Token: rec, Scopes: scopes, Communities: communities})
		c.Set("plugin_callback_request_id", requestID)
		c.Next()
	}
}

func (s *Server) currentPluginCallbackIdentity(c *gin.Context) (pluginCallbackIdentity, bool) {
	v, ok := c.Get(pluginCallbackCtxToken)
	if !ok {
		return pluginCallbackIdentity{}, false
	}
	out, ok := v.(pluginCallbackIdentity)
	return out, ok
}

func parseInt64OrZero(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func (s *Server) appendPluginCallbackRequestLog(record domain.PluginCallbackRequest) {
	record.ErrorMessage = truncateString(record.ErrorMessage, 1000)
	record.UserAgent = truncateString(record.UserAgent, 255)
	_, _ = s.svc.AppendPluginCallbackRequest(record)
}

func truncateString(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func (s *Server) pluginCallbackConfig(c *gin.Context) {
	ident, ok := s.currentPluginCallbackIdentity(c)
	if !ok || ident.Token.PluginCode == "" {
		failAPIError(c, domain.NewPluginError("TOKEN_INVALID", "token 不存在").WithStatus(http.StatusUnauthorized))
		return
	}
	communityID := parseInt64OrZero(firstNonEmpty(strings.TrimSpace(c.Query("community_id")), strings.TrimSpace(c.GetHeader("X-DevHub-Community-ID"))))
	if communityID <= 0 {
		failAPIError(c, domain.NewPluginError("COMMUNITY_ID_REQUIRED", "community_id 必填").WithStatus(400))
		return
	}

	// Use community view (global+community merged) for effective config.
	items, err := s.svc.CommunityPlugins(communityID)
	if err != nil {
		failAPIError(c, domain.NewPluginError("community_not_found", "子站不存在").WithStatus(404))
		return
	}
	var plugin domain.Plugin
	for _, it := range items {
		if strings.TrimSpace(it.Code) == strings.TrimSpace(ident.Token.PluginCode) {
			plugin = it
			break
		}
	}
	if plugin.Code == "" {
		// fallback to global
		plugin, _ = s.svc.PluginByCode(ident.Token.PluginCode)
	}
	if plugin.Code == "" {
		failAPIError(c, domain.NewPluginError("plugin_not_found", "插件不存在").WithStatus(404))
		return
	}

	effective := map[string]any{}
	if m, ok := plugin.ResolvedConfig.(map[string]any); ok {
		if eff, ok := m["effective"]; ok {
			if mm, ok := eff.(map[string]any); ok {
				effective = mm
			}
		}
	}
	raw, _ := json.Marshal(effective)
	redacted := pluginregistry.RedactConfig(plugin.ConfigSchema, string(raw))

	// Audit + request log.
	reqID, _ := c.Get("plugin_callback_request_id")
	requestID, _ := reqID.(string)
	s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
		RequestID:      requestID,
		PluginCode:     ident.Token.PluginCode,
		TokenRef:       ident.Token.TokenRef,
		APIPath:        c.FullPath(),
		Method:         c.Request.Method,
		ScopeRequired:  "config.read",
		Status:         domain.PluginCallbackRequestStatusAccepted,
		ResponseStatus: http.StatusOK,
		CommunityID:    communityID,
		ActorType:      "plugin_service",
		ActorID:        0,
		IPAddress:      c.ClientIP(),
		UserAgent:      truncateString(c.Request.UserAgent(), 255),
		DurationMS:     0,
	})
	s.svc.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     "system",
		ActorType: "system",
		ActorID:   0,
		Action:    "plugin.callback.config.read",
		Target:    "plugin-callback#" + ident.Token.PluginCode,
		Metadata:  auditJSON(gin.H{"plugin_code": ident.Token.PluginCode, "token_ref": ident.Token.TokenRef, "community_id": communityID, "request_id": requestID}),
		CreatedAt: service.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"plugin_code":      ident.Token.PluginCode,
		"community_id":     communityID,
		"effective_config": redacted,
	})
}

type pluginCallbackAuditEventRequest struct {
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Metadata     map[string]any `json:"metadata"`
}

func (s *Server) pluginCallbackAuditEvents(c *gin.Context) {
	ident, ok := s.currentPluginCallbackIdentity(c)
	if !ok || ident.Token.PluginCode == "" {
		failAPIError(c, domain.NewPluginError("TOKEN_INVALID", "token 不存在").WithStatus(http.StatusUnauthorized))
		return
	}
	var req pluginCallbackAuditEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewPluginError("callback_request_invalid", "请求参数不合法").WithStatus(400))
		return
	}
	req.Action = strings.TrimSpace(req.Action)
	if req.Action == "" {
		failAPIError(c, domain.NewPluginError("callback_request_invalid", "action 必填").WithStatus(400))
		return
	}
	// Require plugin_code prefix to prevent forging core/admin actions.
	if !strings.HasPrefix(req.Action, ident.Token.PluginCode+".") {
		failAPIError(c, domain.NewPluginError("callback_action_invalid", "action 必须以 plugin_code 前缀开头").WithStatus(400).
			WithDetail("plugin_code", ident.Token.PluginCode).
			WithSuggestion("例如："+ident.Token.PluginCode+".received_event"))
		return
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}
	// Limit metadata size.
	metaRaw, _ := json.Marshal(req.Metadata)
	if len(metaRaw) > 8*1024 {
		failAPIError(c, domain.NewPluginError("callback_metadata_too_large", "metadata 过大").WithStatus(400))
		return
	}

	reqID, _ := c.Get("plugin_callback_request_id")
	requestID, _ := reqID.(string)
	s.appendPluginCallbackRequestLog(domain.PluginCallbackRequest{
		RequestID:      requestID,
		PluginCode:     ident.Token.PluginCode,
		TokenRef:       ident.Token.TokenRef,
		APIPath:        c.FullPath(),
		Method:         c.Request.Method,
		ScopeRequired:  "audit.write",
		Status:         domain.PluginCallbackRequestStatusAccepted,
		ResponseStatus: http.StatusOK,
		CommunityID:    parseInt64OrZero(strings.TrimSpace(c.GetHeader("X-DevHub-Community-ID"))),
		ActorType:      "plugin_service",
		ActorID:        0,
		IPAddress:      c.ClientIP(),
		UserAgent:      truncateString(c.Request.UserAgent(), 255),
		DurationMS:     0,
	})
	s.svc.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     ident.Token.PluginCode,
		ActorType: "plugin_service",
		ActorID:   0,
		Action:    "plugin.callback.audit.write",
		Target:    strings.TrimSpace(req.ResourceType) + "#" + strings.TrimSpace(req.ResourceID),
		Metadata: auditJSON(gin.H{
			"plugin_code": ident.Token.PluginCode,
			"token_ref":   ident.Token.TokenRef,
			"request_id":  requestID,
			"action":      req.Action,
			"metadata":    req.Metadata,
		}),
		IP: c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
