package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"
	"devhub-gin-backend/internal/service"

	"github.com/gin-gonic/gin"
)

func (s *Server) listAdminPluginRemoteIndexes(c *gin.Context) {
	filter := domain.PluginRemoteIndexFilter{
		Status:   strings.TrimSpace(c.Query("status")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Page:     queryInt(c, "page", 1),
		PageSize: queryInt(c, "page_size", 20),
	}
	res, err := s.svc.ListPluginRemoteIndexes(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) createAdminPluginRemoteIndex(c *gin.Context) {
	var req domain.PluginRemoteIndexSource
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.CreatePluginRemoteIndex(remoteIndexOperator(c), req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) updateAdminPluginRemoteIndex(c *gin.Context) {
	id, ok := remoteIndexIDParam(c)
	if !ok {
		return
	}
	var req domain.PluginRemoteIndexSource
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	res, err := s.svc.UpdatePluginRemoteIndex(remoteIndexOperator(c), id, req)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) enableAdminPluginRemoteIndex(c *gin.Context) {
	s.setAdminPluginRemoteIndexStatus(c, domain.PluginRemoteIndexStatusEnabled)
}

func (s *Server) disableAdminPluginRemoteIndex(c *gin.Context) {
	s.setAdminPluginRemoteIndexStatus(c, domain.PluginRemoteIndexStatusDisabled)
}

func (s *Server) setAdminPluginRemoteIndexStatus(c *gin.Context, status string) {
	id, ok := remoteIndexIDParam(c)
	if !ok {
		return
	}
	res, err := s.svc.SetPluginRemoteIndexStatus(remoteIndexOperator(c), id, status)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) deleteAdminPluginRemoteIndex(c *gin.Context) {
	id, ok := remoteIndexIDParam(c)
	if !ok {
		return
	}
	if err := s.svc.DeletePluginRemoteIndex(remoteIndexOperator(c), id); err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "id": id})
}

func (s *Server) fetchAdminPluginRemoteIndex(c *gin.Context) {
	id, ok := remoteIndexIDParam(c)
	if !ok {
		return
	}
	res, err := s.svc.FetchPluginRemoteIndex(remoteIndexOperator(c), id)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminPluginRemoteIndexPlugins(c *gin.Context) {
	id, ok := remoteIndexIDParam(c)
	if !ok {
		return
	}
	res, err := s.svc.ListRemoteIndexPlugins(id, strings.TrimSpace(c.Query("keyword")), queryInt(c, "page", 1), queryInt(c, "page_size", 20))
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminPluginRemoteIndexPluginDetail(c *gin.Context) {
	id, ok := remoteIndexIDParam(c)
	if !ok {
		return
	}
	res, err := s.svc.GetRemoteIndexPlugin(id, strings.TrimSpace(c.Param("code")))
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func remoteIndexIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("plugin_remote_index_not_found", "远程索引源不存在").WithStatus(404))
		return 0, false
	}
	return id, true
}

func remoteIndexOperator(c *gin.Context) service.RemoteIndexOperator {
	adminCtx, ok := currentAdminContext(c)
	if !ok {
		return service.RemoteIndexOperator{}
	}
	return service.RemoteIndexOperator{
		ID:   adminCtx.CurrentUser.ID,
		Name: firstNonEmpty(adminCtx.CurrentUser.Username, adminCtx.CurrentUser.Nickname),
	}
}

func queryInt(c *gin.Context, key string, fallback int) int {
	if raw := strings.TrimSpace(c.Query(key)); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
