package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

func (s *Server) listAdminPluginCallbackRequests(c *gin.Context) {
	filter := domain.PluginCallbackRequestFilter{
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		TokenRef:   strings.TrimSpace(c.Query("token_ref")),
		Status:     strings.TrimSpace(c.Query("status")),
		RequestID:  strings.TrimSpace(c.Query("request_id")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	items, total, err := s.svc.PluginCallbackRequests(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	f := filter.Normalize()
	c.JSON(http.StatusOK, domain.PluginCallbackRequestListResponse{
		Items: items,
		Pagination: domain.Pagination{
			Page:     f.Page,
			PageSize: f.PageSize,
			Total:    total,
		},
	})
}
