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

func (s *Server) previewAdminPluginPackageTemplate(c *gin.Context) {
	var req domain.PluginPackageTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.PreviewPluginPackageTemplate(req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) createAdminPluginPackageTemplate(c *gin.Context) {
	var req domain.PluginPackageTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.package.template.create.started", "plugins", nil,
		gin.H{"status": "started"},
		gin.H{"operation": "plugin_package_template_create", "plugin_code": strings.TrimSpace(req.Code), "actor": actor})
	res, err := s.svc.CreatePluginPackageTemplate(req)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.template.create.failed", "plugins", nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"operation": "plugin_package_template_create", "plugin_code": strings.TrimSpace(req.Code), "actor": actor, "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.template.created", fmt.Sprintf("plugin-packages#%s", res.Template.Code), nil,
		gin.H{"status": res.Status},
		gin.H{
			"operation":           "plugin_package_template_create",
			"plugin_code":         res.Template.Code,
			"package_path":        res.Template.PackagePath,
			"dry_run_status":      res.DryRun.Status,
			"risk_level":          res.DryRun.RiskReport.Level,
			"manifest_valid":      res.DryRun.ManifestValidation.Valid,
			"generated_files":     res.Template.Files,
			"registry_go_omitted": true,
			"actor":               actor,
		})
	c.JSON(http.StatusOK, res)
}

func (s *Server) dryRunAdminPluginPackage(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := s.svc.DryRunPluginPackage(req.Path)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) installAdminPluginPackage(c *gin.Context) {
	var req domain.PluginPackageInstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	executor := auditActor(c)
	s.auditStructured(c, "system", "plugin.package.install.started", "plugins", nil,
		gin.H{"status": "started"},
		mergeAuditMeta(gin.H{"operation": "plugin_package_install", "path": strings.TrimSpace(req.Path), "actor": executor}, nil))

	res, err := s.svc.InstallPluginPackage(req)
	if err != nil {
		s.auditStructured(c, "system", "plugin.package.install.failed", "plugins", nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"operation": "plugin_package_install", "path": strings.TrimSpace(req.Path), "error": err.Error(), "actor": executor}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.package.installed", fmt.Sprintf("plugins#%s", res.Plugin.Code), nil,
		gin.H{"status": "installed"},
		mergeAuditMeta(gin.H{
			"operation":       "plugin_package_install",
			"plugin_code":     res.Plugin.Code,
			"install_source":  res.Plugin.SourceType,
			"package_path":    res.Package.Path,
			"risk_level":      res.RiskLevel,
			"checksum_status": res.Checksum.Status,
			"actor":           executor,
		}, nil))
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginPackages(c *gin.Context) {
	root := strings.TrimSpace(c.Query("root"))
	status := strings.TrimSpace(c.DefaultQuery("status", "all"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	risk := strings.TrimSpace(c.Query("risk_level"))
	checksum := strings.TrimSpace(c.Query("checksum_status"))
	manifestValid := strings.TrimSpace(c.Query("manifest_valid"))

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

	resp, err := s.svc.ListPluginPackages(root, service.PluginPackageRepositoryFilter{
		Status:         status,
		Keyword:        keyword,
		RiskLevel:      risk,
		ChecksumStatus: checksum,
		ManifestValid:  manifestValid,
		Page:           page,
		PageSize:       pageSize,
	})
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminPluginPackageDetail(c *gin.Context) {
	path := strings.TrimSpace(c.Query("path"))
	if path == "" {
		failAPIError(c, domain.NewPluginError("plugin_package_path_invalid", "缺少插件包路径").WithStatus(400).WithSuggestion("请提供 query 参数 path，例如 storage/plugins/packages/demo_notice。"))
		return
	}
	res, err := s.svc.GetPluginPackageDetail(path)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) dryRunAdminPluginExport(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	var req domain.PluginPackageExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.DryRunPluginPackageExport(code, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) exportAdminPluginPackage(c *gin.Context) {
	code := strings.TrimSpace(c.Param("code"))
	var req domain.PluginPackageExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.export.started", fmt.Sprintf("plugins#%s", code), nil,
		gin.H{"status": "started"},
		gin.H{"plugin_code": code, "operation": "plugin_export", "actor": actor})
	res, err := s.svc.ExportPluginPackage(code, req)
	if err != nil {
		s.auditStructured(c, "system", "plugin.export.failed", fmt.Sprintf("plugins#%s", code), nil,
			gin.H{"status": "failed"},
			mergeAuditMeta(gin.H{"plugin_code": code, "operation": "plugin_export", "actor": actor, "error": err.Error()}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.exported", fmt.Sprintf("plugins#%s", code), nil,
		gin.H{"status": "exported"},
		gin.H{
			"plugin_code":            res.PluginCode,
			"operation":              "plugin_export",
			"actor":                  actor,
			"output_dir":             res.OutputDir,
			"files":                  res.Files,
			"checksum_status":        res.ChecksumStatus,
			"package_dry_run_status": res.PackageDryRunStatus,
		})
	c.JSON(http.StatusOK, res)
}
