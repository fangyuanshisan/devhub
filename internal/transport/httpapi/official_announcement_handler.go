package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

const officialAnnouncementPluginCode = "official_announcement"

type officialAnnouncementContextResponse struct {
	PluginCode string         `json:"plugin_code"`
	MountID    string         `json:"mount_id"`
	Area       string         `json:"area"`
	Visible    bool           `json:"visible"`
	Reason     string         `json:"reason,omitempty"`
	Community  map[string]any `json:"community,omitempty"`
	Viewer     map[string]any `json:"viewer,omitempty"`
	Context    map[string]any `json:"context"`
	Config     map[string]any `json:"config"`
}

func (s *Server) officialAnnouncementContext(c *gin.Context) {
	mountID := strings.TrimSpace(c.Query("mount_id"))
	area := strings.TrimSpace(c.DefaultQuery("area", "frontend"))
	communitySlug := strings.TrimSpace(c.Query("community_slug"))

	if mountID == "" {
		failAPIError(c, domain.NewAPIError("mount_id_required", "mount_id 必填").WithStatus(http.StatusBadRequest))
		return
	}
	if area != "frontend" && area != "admin" {
		failAPIError(c, domain.NewAPIError("area_invalid", "area 不合法").WithStatus(http.StatusBadRequest))
		return
	}

	communityID := int64(0)
	communityInfo := map[string]any{}
	visible := true
	reason := ""
	pluginEnabled := s.svc.IsPluginEnabled(officialAnnouncementPluginCode)
	plugin, pluginOK := s.svc.PluginByCode(officialAnnouncementPluginCode)
	pluginArchived := !pluginOK || strings.TrimSpace(plugin.Status) == pluginregistry.StatusArchived
	if !pluginEnabled || pluginArchived {
		visible = false
		reason = "plugin_disabled"
		if pluginArchived {
			reason = "plugin_soft_uninstalled"
		}
	}
	if communitySlug != "" {
		comm, ok := s.svc.CommunityBySlug(communitySlug)
		if !ok || comm.ID <= 0 {
			failAPIError(c, domain.NewAPIError("community_not_found", "子站不存在").WithStatus(http.StatusNotFound))
			return
		}
		communityID = comm.ID
		communityInfo = map[string]any{"id": comm.ID, "slug": comm.Slug, "name": comm.Name}
		if !s.svc.IsPluginEnabledForCommunity(comm.ID, officialAnnouncementPluginCode) {
			visible = false
			if reason == "" {
				reason = "community_plugin_disabled"
			}
		}
	}

	viewer := map[string]any{"logged_in": false, "role": "anonymous"}
	user, ok := currentUser(c)
	if ok {
		viewer["logged_in"] = true
		viewer["role"] = strings.TrimSpace(firstNonEmpty(user.RoleCode, user.TokenType, "user"))
		viewer["user_id"] = user.ID
	}
	if area == "admin" {
		if !ok || strings.TrimSpace(user.TokenType) != "admin" {
			failAPIError(c, domain.NewAPIError("admin_required", "需要后台管理员身份").WithStatus(http.StatusForbidden))
			return
		}
		if !hasPermission(user.Permissions, "plugin.read") && !hasPermission(user.Permissions, "plugin.manage") {
			failAPIError(c, domain.NewAPIError("permission_denied", "权限不足").WithStatus(http.StatusForbidden).
				WithDetail("permission_code", "plugin.read"))
			return
		}
	}

	if !visible && area == "admin" {
		// admin preview should still be accessible to authorized operators, but it can explicitly render as invisible.
		// Do not force a 403 here; the UI can show the empty/invisible state.
	}

	config := s.officialAnnouncementEffectiveConfig(communityID)
	if !truthy(config["enabled"]) || strings.TrimSpace(asString(config["message"])) == "" {
		visible = false
		if reason == "" {
			reason = "config_disabled"
		}
	}

	ctx := map[string]any{
		"plugin_code":  officialAnnouncementPluginCode,
		"mount_id":     mountID,
		"community_id": communityID,
		"viewer":       viewer,
		"capabilities": []any{"config.read", "audit.write"},
	}

	c.JSON(http.StatusOK, officialAnnouncementContextResponse{
		PluginCode: officialAnnouncementPluginCode,
		MountID:    mountID,
		Area:       area,
		Visible:    visible,
		Reason:     reason,
		Community:  communityInfo,
		Viewer:     viewer,
		Context:    ctx,
		Config:     config,
	})
}

type officialAnnouncementAuditEventRequest struct {
	MountID       string         `json:"mount_id"`
	Area          string         `json:"area"`
	CommunitySlug string         `json:"community_slug"`
	Action        string         `json:"action"`
	Metadata      map[string]any `json:"metadata"`
	RequestID     string         `json:"request_id"`
}

func (s *Server) officialAnnouncementAuditEvents(c *gin.Context) {
	var req officialAnnouncementAuditEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failAPIError(c, domain.NewAPIError("request_invalid", "请求参数不合法").WithStatus(http.StatusBadRequest))
		return
	}
	req.MountID = strings.TrimSpace(req.MountID)
	req.Area = strings.TrimSpace(firstNonEmpty(req.Area, "frontend"))
	req.CommunitySlug = strings.TrimSpace(req.CommunitySlug)
	req.Action = strings.TrimSpace(req.Action)
	req.RequestID = strings.TrimSpace(req.RequestID)

	if req.MountID == "" {
		failAPIError(c, domain.NewAPIError("mount_id_required", "mount_id 必填").WithStatus(http.StatusBadRequest))
		return
	}
	if req.Area != "frontend" && req.Area != "admin" {
		failAPIError(c, domain.NewAPIError("area_invalid", "area 不合法").WithStatus(http.StatusBadRequest))
		return
	}
	if req.Action == "" || !strings.HasPrefix(req.Action, officialAnnouncementPluginCode+".") {
		failAPIError(c, domain.NewAPIError("action_invalid", "action 必须以 official_announcement. 前缀开头").WithStatus(http.StatusBadRequest))
		return
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}
	raw, _ := json.Marshal(req.Metadata)
	if len(raw) > 8*1024 {
		failAPIError(c, domain.NewAPIError("metadata_too_large", "metadata 过大").WithStatus(http.StatusBadRequest))
		return
	}
	// Never allow any credential-like values to be stored accidentally.
	for k := range req.Metadata {
		lk := strings.ToLower(strings.TrimSpace(k))
		if strings.Contains(lk, "token") || strings.Contains(lk, "secret") || strings.Contains(lk, "authorization") {
			delete(req.Metadata, k)
		}
	}
	if !s.svc.IsPluginEnabled(officialAnnouncementPluginCode) {
		failAPIError(c, domain.NewAPIError("plugin_disabled", "插件未启用").WithStatus(http.StatusForbidden))
		return
	}
	if plugin, ok := s.svc.PluginByCode(officialAnnouncementPluginCode); !ok || strings.TrimSpace(plugin.Status) == pluginregistry.StatusArchived {
		failAPIError(c, domain.NewAPIError("plugin_soft_uninstalled", "插件已软卸载").WithStatus(http.StatusForbidden))
		return
	}
	if req.CommunitySlug != "" {
		comm, ok := s.svc.CommunityBySlug(req.CommunitySlug)
		if !ok || comm.ID <= 0 {
			failAPIError(c, domain.NewAPIError("community_not_found", "子站不存在").WithStatus(http.StatusNotFound))
			return
		}
		if !s.svc.IsPluginEnabledForCommunity(comm.ID, officialAnnouncementPluginCode) {
			failAPIError(c, domain.NewAPIError("community_plugin_disabled", "插件未在子站启用").WithStatus(http.StatusForbidden))
			return
		}
	}

	user, ok := currentUser(c)
	actorType := "anonymous"
	actor := "anonymous"
	actorID := int64(0)
	if ok {
		actor = auditActor(c)
		actorID = user.ID
		if strings.TrimSpace(user.TokenType) == "admin" {
			actorType = "admin_user"
		} else if user.IsModerator {
			actorType = "moderator"
		} else {
			actorType = "user"
		}
	}
	if req.Area == "admin" {
		if !ok || strings.TrimSpace(user.TokenType) != "admin" {
			failAPIError(c, domain.NewAPIError("admin_required", "需要后台管理员身份").WithStatus(http.StatusForbidden))
			return
		}
		if !hasPermission(user.Permissions, "plugin.read") && !hasPermission(user.Permissions, "plugin.manage") {
			failAPIError(c, domain.NewAPIError("permission_denied", "权限不足").WithStatus(http.StatusForbidden).
				WithDetail("permission_code", "plugin.read"))
			return
		}
	}

	metadata := gin.H{
		"plugin_code":    officialAnnouncementPluginCode,
		"mount_id":       req.MountID,
		"request_id":     req.RequestID,
		"area":           req.Area,
		"community_slug": req.CommunitySlug,
		"metadata":       req.Metadata,
	}
	s.svc.AppendAdminLog(domain.AdminLog{
		Site:      "admin",
		Actor:     actor,
		ActorType: actorType,
		ActorID:   actorID,
		Action:    req.Action,
		Target:    "official-announcement#" + req.MountID,
		Metadata:  auditJSON(metadata),
		IP:        c.ClientIP(),
		CreatedAt: service.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) officialAnnouncementEffectiveConfig(communityID int64) map[string]any {
	plugin, _ := s.svc.PluginByCode(officialAnnouncementPluginCode)
	if plugin.Code == "" {
		return map[string]any{}
	}
	// For effective config, prefer community merged view when communityID > 0.
	if communityID > 0 {
		if items, err := s.svc.CommunityPlugins(communityID); err == nil {
			for _, it := range items {
				if strings.TrimSpace(it.Code) == officialAnnouncementPluginCode {
					plugin = it
					break
				}
			}
		}
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
	cfg, _ := redacted.(map[string]any)
	if cfg == nil {
		cfg = map[string]any{}
	}
	// Hardening: ensure link_url is safe and strictly internal.
	cfg["link_url"] = sanitizeInternalLink(asString(cfg["link_url"]))
	cfg["message"] = strings.TrimSpace(asString(cfg["message"]))
	cfg["link_text"] = strings.TrimSpace(asString(cfg["link_text"]))
	return cfg
}

func (s *Server) officialAnnouncementIframe(c *gin.Context) {
	// Serve a minimal built-in iframe page for the official plugin.
	// The iframe performs postMessage handshake only; it does not call Core APIs directly.
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, officialAnnouncementIframeHTML())
}

func officialAnnouncementIframeHTML() string {
	// NOTE: Keep this HTML self-contained to avoid depending on frontend build outputs.
	// Security: no remote JS, no inline secrets/tokens.
	return `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Official Announcement</title>
  <style>
    body{margin:0;font-family:ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial;}
    .banner{display:flex;gap:10px;align-items:center;padding:10px 12px;border:1px solid #e5e7eb;border-radius:10px;background:#fff;line-height:1.4}
    .msg{flex:1;min-width:0}
    .link{color:#2563eb;text-decoration:none;white-space:nowrap}
    .muted{color:#6b7280}
  </style>
</head>
<body>
  <div id="root"></div>
  <script>
    (() => {
      const root = document.getElementById('root');
      const state = { mountId: '', pluginCode: 'official_announcement', requestId: '', area: 'frontend', context: null, config: null, loading: true };
      const escapeHTML = (v) => String(v ?? '').replace(/[&<>"']/g, (c) => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#39;'}[c]));
      const uid = (prefix) => prefix + '_' + Math.random().toString(16).slice(2) + Date.now().toString(16);

      const renderLoading = () => {
        root.innerHTML = '<div class="muted" style="padding:10px 12px">加载中...</div>';
      };

      const renderError = () => {
        root.innerHTML = '<div class="muted" style="padding:10px 12px">公告加载失败</div>';
      };

      const render = () => {
        const cfg = state.config || {};
        const enabled = !!cfg.enabled;
        const message = String(cfg.message || '').trim();
        if (!enabled || !message) {
          root.innerHTML = '<div class="muted" style="padding:10px 12px">（公告未启用）</div>';
          return;
        }
        const linkText = String(cfg.link_text || '').trim();
        const linkURL = String(cfg.link_url || '').trim();
        const safeLinkURL = linkURL && linkURL.startsWith('/') ? escapeHTML(linkURL) : '';
        const link = linkText && safeLinkURL
          ? '<a id="announcement-link" class="link" href="' + safeLinkURL + '" target="_top" rel="noreferrer noopener">' + escapeHTML(linkText) + '</a>'
          : '';
        const dismiss = cfg.dismissible
          ? '<button id="announcement-dismiss" type="button" style="border:none;background:#f3f4f6;color:#64748b;cursor:pointer;padding:0 8px;border-radius:6px;">关闭</button>'
          : '';
        root.innerHTML = '<div class="banner"><div class="msg">' + escapeHTML(message) + '</div>' + link + dismiss + '</div>';
        if (link && safeLinkURL) {
          const linkEl = document.getElementById('announcement-link');
          if (linkEl) {
            linkEl.addEventListener('click', () => {
              post('devhub.plugin.audit.write', { action: 'official_announcement.clicked', metadata: { mount_id: state.mountId, link_url: safeLinkURL } });
            });
          }
        }
        if (cfg.dismissible) {
          const dismissEl = document.getElementById('announcement-dismiss');
          if (dismissEl) {
            dismissEl.addEventListener('click', () => {
              root.innerHTML = '<div class="muted" style="padding:10px 12px">公告已关闭</div>';
              post('devhub.plugin.audit.write', { action: 'official_announcement.dismissed', metadata: { mount_id: state.mountId } });
            });
          }
        }
      };

      const post = (type, payload) => {
        window.parent && window.parent.postMessage({
          schema_version: '1',
          type,
          plugin_code: state.pluginCode,
          mount_id: state.mountId,
          request_id: state.requestId,
          payload: payload || {}
        }, '*');
      };

      window.addEventListener('message', (event) => {
        const msg = event && event.data;
        if (!msg || msg.schema_version !== '1' || msg.plugin_code !== state.pluginCode) return;
        if (state.mountId && msg.mount_id !== state.mountId) return;
        if (msg.type === 'devhub.plugin.context') {
          state.context = msg.payload && msg.payload.context || null;
          state.config = msg.payload && msg.payload.config || null;
          state.loading = false;
          render();
          if (state.area === 'admin') {
            post('devhub.plugin.audit.write', { action: 'official_announcement.admin_previewed', metadata: { mount_id: state.mountId } });
          }
          post('devhub.plugin.audit.write', { action: 'official_announcement.rendered', metadata: { mount_id: state.mountId } });
          return;
        }
        if (msg.type === 'devhub.plugin.config.result') {
          state.config = msg.payload && msg.payload.config || null;
          render();
          return;
        }
        if (msg.type === 'devhub.plugin.error') {
          root.innerHTML = '<div class="muted" style="padding:10px 12px">公告加载失败</div>';
          return;
        }
      });

      state.mountId = String(new URLSearchParams(location.search).get('mount_id') || '').trim() || uid('mnt');
      state.area = String(new URLSearchParams(location.search).get('area') || 'frontend').trim() || 'frontend';
      state.requestId = uid('req');
      renderLoading();
      post('devhub.plugin.ready', {});
      post('devhub.plugin.config.read', {});
    })();
  </script>
</body>
</html>`
}

func truthy(v any) bool {
	switch vv := v.(type) {
	case bool:
		return vv
	case string:
		vv = strings.TrimSpace(strings.ToLower(vv))
		return vv == "true" || vv == "1" || vv == "yes" || vv == "on"
	case float64:
		return vv != 0
	default:
		return false
	}
}

func asString(v any) string {
	switch vv := v.(type) {
	case string:
		return vv
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), "\"")
	}
}

func sanitizeInternalLink(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "/"
	}
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") {
		return "/"
	}
	// Only allow internal absolute paths.
	if !strings.HasPrefix(raw, "/") {
		return "/"
	}
	// Avoid scheme-relative URLs like //evil.com
	if strings.HasPrefix(raw, "//") {
		return "/"
	}
	return raw
}
