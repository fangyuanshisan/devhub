package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	pluginregistry "devhub-gin-backend/internal/plugins"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (s *Server) updateAdminPluginConfig(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	before, _ := s.svc.PluginByCode(code)
	var req struct {
		ConfigJSON any `json:"config_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	payload := ""
	if req.ConfigJSON != nil {
		raw, err := json.Marshal(req.ConfigJSON)
		if err != nil || !json.Valid(raw) {
			fail(c, http.StatusBadRequest, "config_json 必须是合法 JSON")
			return
		}
		payload = string(raw)
	}
	plugin, err := s.svc.SetPluginConfig(code, payload)
	if err != nil {
		failAPIError(c, err)
		return
	}
	redactedBefore := pluginregistry.RedactConfig(plugin.ConfigSchema, before.ConfigJSON)
	redactedAfter := pluginregistry.RedactConfig(plugin.ConfigSchema, plugin.ConfigJSON)
	// Record config version history (redacted diff) and link it in audit metadata.
	operator := service.PluginConfigVersionOperator{Type: "admin_user", ID: 0, Name: auditActor(c)}
	if adminCtx, ok := currentAdminContext(c); ok && adminCtx.CurrentUser.ID > 0 {
		operator.ID = adminCtx.CurrentUser.ID
		if adminCtx.CurrentUser.Username != "" {
			operator.Name = adminCtx.CurrentUser.Username
		} else if adminCtx.CurrentUser.Nickname != "" {
			operator.Name = adminCtx.CurrentUser.Nickname
		}
	}
	ver, created, _ := s.svc.RecordPluginConfigVersion(plugin.Code, domain.PluginConfigScopeGlobal, 0, before.ConfigJSON, plugin.ConfigJSON, "manual", operator)
	s.auditStructured(c, "system", "更新插件全局配置", fmt.Sprintf("plugins#%s", plugin.Code),
		gin.H{"config_json": redactedBefore},
		gin.H{"config_json": redactedAfter},
		gin.H{
			"scope":        "global",
			"plugin_code":  plugin.Code,
			"operation":    "plugin_config",
			"changed_keys": configChangedKeys(before.ConfigJSON, plugin.ConfigJSON),
			"config_version_id": func() int64 {
				if created {
					return ver.ID
				}
				return 0
			}(),
		})
	c.JSON(http.StatusOK, plugin)
}

func (s *Server) listAdminPluginConfigVersions(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	page := 1
	pageSize := 20
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			pageSize = n
		}
	}
	resp, err := s.svc.ListPluginConfigVersions(code, domain.PluginConfigScopeGlobal, 0, page, pageSize)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminPluginConfigVersionDetail(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	rawID := strings.TrimSpace(c.Param("version_id"))
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_config_version_not_found", "配置版本不存在").WithStatus(404))
		return
	}
	resp, err := s.svc.GetPluginConfigVersionDetail(code, domain.PluginConfigScopeGlobal, 0, id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) dryRunAdminPluginConfigRollback(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	rawID := strings.TrimSpace(c.Param("version_id"))
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_config_version_not_found", "配置版本不存在").WithStatus(404))
		return
	}
	plugin, ok := s.svc.PluginByCode(code)
	if !ok || plugin.Code == "" {
		failAPIError(c, domain.NewPluginError("plugin_not_found", "插件不存在").WithStatus(404).WithDetail("plugin_code", code))
		return
	}
	resp, err := s.svc.PluginConfigRollbackDryRun(code, domain.PluginConfigScopeGlobal, 0, id, plugin.ConfigJSON)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) listAdminCommunityPluginConfigVersions(c *gin.Context) {
	communityID, ok := idParam(c, "id")
	if !ok {
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	page := 1
	pageSize := 20
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			pageSize = n
		}
	}
	resp, err := s.svc.ListPluginConfigVersions(code, domain.PluginConfigScopeCommunity, communityID, page, pageSize)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminCommunityPluginConfigVersionDetail(c *gin.Context) {
	communityID, ok := idParam(c, "id")
	if !ok {
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	rawID := strings.TrimSpace(c.Param("version_id"))
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_config_version_not_found", "配置版本不存在").WithStatus(404))
		return
	}
	resp, err := s.svc.GetPluginConfigVersionDetail(code, domain.PluginConfigScopeCommunity, communityID, id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) dryRunAdminCommunityPluginConfigRollback(c *gin.Context) {
	communityID, ok := idParam(c, "id")
	if !ok {
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	rawID := strings.TrimSpace(c.Param("version_id"))
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_config_version_not_found", "配置版本不存在").WithStatus(404))
		return
	}
	items, _ := s.svc.CommunityPlugins(communityID)
	current := ""
	for _, it := range items {
		if it.Code == code {
			current = it.ConfigJSON
			break
		}
	}
	resp, err := s.svc.PluginConfigRollbackDryRun(code, domain.PluginConfigScopeCommunity, communityID, id, current)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) updateAdminCommunityPluginConfig(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if !s.canManageCommunityConfig(c, id) {
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	beforeItems, _ := s.svc.CommunityPlugins(id)
	beforeConfig := ""
	for _, item := range beforeItems {
		if item.Code == code {
			beforeConfig = item.ConfigJSON
			break
		}
	}
	var req struct {
		ConfigJSON any `json:"config_json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	payload := ""
	if req.ConfigJSON != nil {
		raw, err := json.Marshal(req.ConfigJSON)
		if err != nil || !json.Valid(raw) {
			fail(c, http.StatusBadRequest, "config_json 必须是合法 JSON")
			return
		}
		payload = string(raw)
	}
	plugin, err := s.svc.SetCommunityPluginConfig(id, code, payload)
	if err != nil {
		failAPIError(c, err)
		return
	}
	redactedBefore := pluginregistry.RedactConfig(plugin.ConfigSchema, beforeConfig)
	redactedAfter := pluginregistry.RedactConfig(plugin.ConfigSchema, plugin.ConfigJSON)
	operator := service.PluginConfigVersionOperator{Type: "admin_user", ID: 0, Name: auditActor(c)}
	if adminCtx, ok := currentAdminContext(c); ok && adminCtx.CurrentUser.ID > 0 {
		operator.ID = adminCtx.CurrentUser.ID
		if adminCtx.CurrentUser.Username != "" {
			operator.Name = adminCtx.CurrentUser.Username
		} else if adminCtx.CurrentUser.Nickname != "" {
			operator.Name = adminCtx.CurrentUser.Nickname
		}
	}
	ver, created, _ := s.svc.RecordPluginConfigVersion(plugin.Code, domain.PluginConfigScopeCommunity, id, beforeConfig, plugin.ConfigJSON, "manual", operator)
	s.auditStructured(c, "system", "更新子站插件配置", fmt.Sprintf("community_plugins#%d:%s", id, plugin.Code),
		gin.H{"config_json": redactedBefore},
		gin.H{"config_json": redactedAfter},
		gin.H{
			"scope":        "community",
			"community_id": id,
			"plugin_code":  plugin.Code,
			"operation":    "community_plugin_config",
			"changed_keys": configChangedKeys(beforeConfig, plugin.ConfigJSON),
			"config_version_id": func() int64 {
				if created {
					return ver.ID
				}
				return 0
			}(),
		})
	c.JSON(http.StatusOK, plugin)
}
