package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) validateAdminPluginManifest(c *gin.Context) {
	result, err := s.parseAndValidatePluginManifest(c)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) dryRunAdminPluginManifest(c *gin.Context) {
	result, err := s.parseAndValidatePluginManifest(c)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) dryRunAdminPluginUpgrade(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	raw, err := c.GetRawData()
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	manifestJSON, err := extractManifestJSON(raw)
	if err != nil {
		failAPIError(c, err)
		return
	}
	result, err := s.svc.PluginUpgradeDryRun(code, manifestJSON)
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.upgrade.previewed", fmt.Sprintf("plugins#%s", code), nil,
		gin.H{"status": "previewed"},
		gin.H{"plugin_code": code, "operation": "plugin_upgrade_preview", "compatibility_status": result.CompatibilityStatus, "changed_keys": result.ChangedKeys})
	c.JSON(http.StatusOK, result)
}

func (s *Server) upgradeAdminPlugin(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	raw, err := c.GetRawData()
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	manifestJSON, err := extractManifestJSON(raw)
	if err != nil {
		failAPIError(c, err)
		return
	}
	executor := auditActor(c)
	preview, err := s.svc.PluginUpgradeDryRun(code, manifestJSON)
	if err != nil {
		s.auditStructured(c, "system", "plugin.upgrade.failed", fmt.Sprintf("plugins#%s", code), nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"plugin_code": code, "operation": "plugin_upgrade", "error": err.Error(), "actor": executor}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.upgrade.started", fmt.Sprintf("plugins#%s", code), nil,
		gin.H{"status": "pending"},
		gin.H{"plugin_code": code, "operation": "plugin_upgrade", "actor": executor, "compatibility_status": preview.CompatibilityStatus, "changed_keys": preview.ChangedKeys})
	result, err := s.svc.UpgradePluginManifest(code, manifestJSON)
	if err != nil {
		s.auditStructured(c, "system", "plugin.upgrade.failed", fmt.Sprintf("plugins#%s", code), nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"plugin_code": code, "operation": "plugin_upgrade", "actor": executor, "error": err.Error(), "compatibility_status": preview.CompatibilityStatus}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.upgraded", fmt.Sprintf("plugins#%s", code),
		gin.H{"status": "pending"},
		gin.H{"status": result.Plugin.Status},
		gin.H{"plugin_code": code, "operation": "plugin_upgraded", "actor": executor, "current_version": preview.CurrentVersion, "new_version": preview.NewVersion, "compatibility_status": preview.CompatibilityStatus, "changed_keys": preview.ChangedKeys})
	c.JSON(http.StatusOK, result)
}

func (s *Server) installAdminPluginManifest(c *gin.Context) {
	raw, err := c.GetRawData()
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	manifestJSON, err := extractManifestJSON(raw)
	if err != nil {
		failAPIError(c, err)
		return
	}
	executor := auditActor(c)
	validation, _ := s.svc.ValidatePluginManifestJSON(manifestJSON)
	beforeCount := len(s.svc.Plugins())
	s.auditStructured(c, "system", "plugin.install.started", "plugins#manifest", nil,
		gin.H{"status": "pending", "source_type": firstNonEmpty(validation.NormalizedManifest.SourceType, "manifest")},
		gin.H{"plugin_code": validation.NormalizedManifest.Code, "operation": "plugin_install", "actor": executor, "impact_summary": validation.ImpactSummary, "before_count": beforeCount})
	plugin, result, err := s.svc.InstallPluginManifest(manifestJSON)
	if err != nil {
		s.auditStructured(c, "system", "plugin.install.failed", "plugins#manifest", nil, gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"operation": "plugin_install", "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	impact := result.ImpactSummary
	s.auditStructured(c, "system", "plugin.installed", fmt.Sprintf("plugins#%s", plugin.Code),
		gin.H{"status": "pending"},
		gin.H{"status": plugin.Status},
		gin.H{"plugin_code": plugin.Code, "operation": "plugin_installed", "impact_summary": impact})
	c.JSON(http.StatusCreated, gin.H{"plugin": plugin, "validation": result})
}
