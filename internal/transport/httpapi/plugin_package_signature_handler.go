package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"
)

func (s *Server) verifyAdminPluginPackageSignature(c *gin.Context) {
	raw := strings.TrimSpace(c.Param("id"))
	precheckID, _ := strconv.ParseInt(raw, 10, 64)
	if precheckID <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_package_signature_invalid_request", "precheck_id 不合法").WithStatus(400))
		return
	}

	operator := service.PluginPackageSignatureOperator{}
	if adminCtx, ok := currentAdminContext(c); ok {
		operator.ID = adminCtx.CurrentUser.ID
		operator.Name = firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname)
	}

	res, err := s.svc.VerifyPluginPackageSignatureForPrecheckAs(operator, precheckID)
	action := "plugin.signature.verify.success"
	status := domain.PluginPackageSignatureStatusFailed
	if err != nil {
		action = "plugin.signature.verify.failed"
	} else {
		status = res.Status
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.signature.verify.requested", fmt.Sprintf("plugin-package-prechecks#%d", precheckID), nil, gin.H{"status": "requested"}, gin.H{
		"package_precheck_id": precheckID,
		"actor":               actor,
	})
	if err != nil {
		s.auditStructured(c, "system", action, fmt.Sprintf("plugin-package-prechecks#%d", precheckID), nil, gin.H{"status": status}, mergeAuditMeta(gin.H{
			"package_precheck_id": precheckID,
			"signature_status":    res.Status,
			"plugin_code":         res.PluginCode,
			"version":             res.Version,
			"publisher_id":        res.PublisherID,
			"key_id":              res.KeyID,
			"actor":               actor,
		}, gin.H{"error": err.Error()}))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", action, fmt.Sprintf("plugin-package-prechecks#%d", precheckID), nil, gin.H{"status": status}, gin.H{
		"package_precheck_id": precheckID,
		"signature_id":        res.ID,
		"signature_status":    res.Status,
		"plugin_code":         res.PluginCode,
		"version":             res.Version,
		"publisher_id":        res.PublisherID,
		"key_id":              res.KeyID,
		"actor":               actor,
	})
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginPackageSignatures(c *gin.Context) {
	filter := domain.PluginPackageSignatureFilter{
		Status:     strings.TrimSpace(c.Query("status")),
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Keyword:    strings.TrimSpace(c.Query("keyword")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	if raw := strings.TrimSpace(c.Query("package_download_id")); raw != "" {
		filter.PackageDownloadID, _ = strconv.ParseInt(raw, 10, 64)
	}
	if raw := strings.TrimSpace(c.Query("package_precheck_id")); raw != "" {
		filter.PackagePrecheckID, _ = strconv.ParseInt(raw, 10, 64)
	}
	res, err := s.svc.ListPluginPackageSignatures(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginPackageSignatureDetail(c *gin.Context) {
	raw := strings.TrimSpace(c.Param("id"))
	id, _ := strconv.ParseInt(raw, 10, 64)
	if id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_package_signature_not_found", "验签记录不存在").WithStatus(404))
		return
	}
	it, err := s.svc.GetPluginPackageSignature(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

func (s *Server) deleteAdminPluginPackageSignature(c *gin.Context) {
	raw := strings.TrimSpace(c.Param("id"))
	id, _ := strconv.ParseInt(raw, 10, 64)
	if id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_package_signature_not_found", "验签记录不存在").WithStatus(404))
		return
	}
	it, err := s.svc.DeletePluginPackageSignature(id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.signature.deleted.failed", fmt.Sprintf("plugin-package-signatures#%d", id), nil, gin.H{"status": "failed"}, gin.H{
			"signature_id": id,
			"error":        err.Error(),
		})
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.signature.deleted", fmt.Sprintf("plugin-package-signatures#%d", id), nil, gin.H{"status": it.Status}, gin.H{
		"signature_id":     it.ID,
		"plugin_code":      it.PluginCode,
		"version":          it.Version,
		"signature_status": it.Status,
	})
	c.JSON(http.StatusOK, it)
}
