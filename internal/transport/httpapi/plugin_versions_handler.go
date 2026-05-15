package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

func (s *Server) listAdminPluginVersions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	res, err := s.svc.ListPluginVersionRepository(domain.PluginVersionFilter{
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Source:     strings.TrimSpace(c.Query("source")),
		Status:     strings.TrimSpace(c.Query("status")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginCodeVersions(c *gin.Context) {
	res, err := s.svc.PluginVersions(strings.TrimSpace(c.Param("code")))
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginVersionDetail(c *gin.Context) {
	remoteIndexID, _ := strconv.ParseInt(c.Query("remote_index_id"), 10, 64)
	res, err := s.svc.PluginVersionDetail(strings.TrimSpace(c.Param("code")), strings.TrimSpace(c.Param("version")), domain.PluginUpgradeDiffRequest{
		Source:        strings.TrimSpace(c.Query("source")),
		PackagePath:   strings.TrimSpace(c.Query("package_path")),
		RemoteIndexID: remoteIndexID,
	})
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) dryRunAdminPluginVersionUpgradeDiff(c *gin.Context) {
	var req domain.PluginUpgradeDiffRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.PluginVersionUpgradeDiff(strings.TrimSpace(c.Param("code")), strings.TrimSpace(c.Param("version")), req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.version.upgrade_diff.previewed", "plugins#"+res.PluginCode, nil, gin.H{"status": res.Status}, gin.H{
		"plugin_code":     res.PluginCode,
		"current_version": res.CurrentVersion,
		"target_version":  res.TargetVersion,
		"source":          res.Source,
		"summary":         res.Summary,
	})
	c.JSON(http.StatusOK, res)
}
