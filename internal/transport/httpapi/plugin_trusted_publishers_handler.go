package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (s *Server) listAdminPluginTrustedPublishers(c *gin.Context) {
	filter := domain.PluginTrustedPublisherFilter{
		Status:   strings.TrimSpace(c.Query("status")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Page:     1,
		PageSize: 20,
	}
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			filter.Page = n
		}
	}
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			filter.PageSize = n
		}
	}
	res, err := s.svc.ListPluginTrustedPublishers(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginTrustedPublisherDetail(c *gin.Context) {
	id, ok := trustedPublisherIDParam(c)
	if !ok {
		return
	}
	res, err := s.svc.GetPluginTrustedPublisher(id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) createAdminPluginTrustedPublisher(c *gin.Context) {
	var req domain.PluginTrustedPublisher
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	operator := trustedPublisherOperator(c)
	res, err := s.svc.CreatePluginTrustedPublisher(operator, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) updateAdminPluginTrustedPublisher(c *gin.Context) {
	id, ok := trustedPublisherIDParam(c)
	if !ok {
		return
	}
	var req domain.PluginTrustedPublisher
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.UpdatePluginTrustedPublisher(trustedPublisherOperator(c), id, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) blockAdminPluginTrustedPublisher(c *gin.Context) {
	s.setAdminPluginTrustedPublisherStatus(c, "blocked")
}

func (s *Server) revokeAdminPluginTrustedPublisher(c *gin.Context) {
	s.setAdminPluginTrustedPublisherStatus(c, "revoked")
}

func (s *Server) restoreAdminPluginTrustedPublisher(c *gin.Context) {
	s.setAdminPluginTrustedPublisherStatus(c, "trusted")
}

func (s *Server) setAdminPluginTrustedPublisherStatus(c *gin.Context, status string) {
	id, ok := trustedPublisherIDParam(c)
	if !ok {
		return
	}
	var req struct {
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&req)
	res, err := s.svc.SetPluginTrustedPublisherStatus(trustedPublisherOperator(c), id, status, req.Comment)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) deleteAdminPluginTrustedPublisher(c *gin.Context) {
	id, ok := trustedPublisherIDParam(c)
	if !ok {
		return
	}
	if err := s.svc.DeletePluginTrustedPublisher(trustedPublisherOperator(c), id); err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "id": id})
}

func trustedPublisherIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_trusted_publisher_not_found", "可信发布者不存在").WithStatus(404))
		return 0, false
	}
	return id, true
}

func trustedPublisherOperator(c *gin.Context) service.TrustedPublisherOperator {
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		return service.TrustedPublisherOperator{}
	}
	return service.TrustedPublisherOperator{
		ID:   adminCtx.CurrentUser.ID,
		Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname),
	}
}
