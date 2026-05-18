package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

func (s *Server) listAdminWebhookEvents(c *gin.Context) {
	filter := domain.WebhookEventFilter{
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		HookName:   strings.TrimSpace(c.Query("hook_name")),
		Mode:       strings.TrimSpace(c.Query("mode")),
		Status:     strings.TrimSpace(c.Query("status")),
		ActorType:  strings.TrimSpace(c.Query("actor_type")),
		RequestID:  strings.TrimSpace(c.Query("request_id")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	filter.CommunityID, _ = strconv.ParseInt(strings.TrimSpace(c.Query("community_id")), 10, 64)
	filter.ActorID, _ = strconv.ParseInt(strings.TrimSpace(c.Query("actor_id")), 10, 64)
	res, err := s.svc.WebhookEventsAdmin(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

