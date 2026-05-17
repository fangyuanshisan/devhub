package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"devhub-gin-backend/internal/domain"

	"github.com/gin-gonic/gin"
)

func (s *Server) listAdminWebhookDeliveries(c *gin.Context) {
	filter := domain.WebhookDeliveryFilter{
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		HookName:   strings.TrimSpace(c.Query("hook_name")),
		Status:     strings.TrimSpace(c.Query("status")),
		EventID:    strings.TrimSpace(c.Query("event_id")),
		DeliveryID: strings.TrimSpace(c.Query("delivery_id")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	res, err := s.svc.WebhookDeliveriesAdmin(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminWebhookDeliveryDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_delivery_not_found", "delivery 不存在").WithStatus(404))
		return
	}
	out, ok := s.svc.WebhookDeliveryByID(id)
	if !ok || out.ID == 0 {
		failAPIError(c, domain.NewPluginError("webhook_delivery_not_found", "delivery 不存在").WithStatus(404))
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) retryAdminWebhookDelivery(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_delivery_not_found", "delivery 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.webhook.delivery.manual_retry", fmt.Sprintf("webhook-deliveries#%d", id), nil, gin.H{"status": "requested"}, gin.H{
		"delivery_id": id,
		"actor":       actor,
	})

	out, err := s.svc.ManualRetryWebhookDelivery(c.Request.Context(), id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.delivery.retry_failed", fmt.Sprintf("webhook-deliveries#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"delivery_id": id,
			"actor":       actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.webhook.delivery.retry_success", fmt.Sprintf("webhook-deliveries#%d", id), nil, gin.H{"status": out.Status}, gin.H{
		"delivery_id": id,
		"actor":       actor,
	})
	c.JSON(http.StatusOK, out)
}

type retryDueWebhookDeliveriesRequest struct {
	Limit int `json:"limit"`
}

func (s *Server) retryDueAdminWebhookDeliveries(c *gin.Context) {
	var req retryDueWebhookDeliveriesRequest
	_ = c.ShouldBindJSON(&req)
	if req.Limit <= 0 || req.Limit > 200 {
		req.Limit = 50
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.webhook.delivery.retry_started", "webhook-deliveries", nil, gin.H{"status": "started"}, gin.H{
		"limit": req.Limit,
		"actor": actor,
	})
	res, err := s.svc.RetryDueWebhookDeliveries(c.Request.Context(), req.Limit)
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.delivery.retry_failed", "webhook-deliveries", nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"limit": req.Limit,
			"actor": actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.webhook.delivery.retry_success", "webhook-deliveries", nil, gin.H{"status": "success"}, gin.H{
		"limit":     req.Limit,
		"processed": res.Processed,
		"success":   res.Success,
		"failed":    res.Failed,
		"skipped":   res.Skipped,
		"actor":     actor,
	})
	c.JSON(http.StatusOK, res)
}

func (s *Server) listAdminWebhookCircuitBreakers(c *gin.Context) {
	filter := domain.WebhookCircuitBreakerFilter{
		PluginCode: strings.TrimSpace(c.Query("plugin_code")),
		Status:     strings.TrimSpace(c.Query("status")),
	}
	filter.Page, _ = strconv.Atoi(strings.TrimSpace(c.Query("page")))
	filter.PageSize, _ = strconv.Atoi(strings.TrimSpace(c.Query("page_size")))
	res, err := s.svc.WebhookCircuitBreakersAdmin(filter)
	if err != nil {
		failAPIError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (s *Server) adminWebhookCircuitBreakerDetail(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_circuit_breaker_not_found", "circuit breaker 不存在").WithStatus(404))
		return
	}
	out, ok := s.svc.WebhookCircuitBreakerByID(id)
	if !ok || out.ID == 0 {
		failAPIError(c, domain.NewPluginError("webhook_circuit_breaker_not_found", "circuit breaker 不存在").WithStatus(404))
		return
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) closeAdminWebhookCircuitBreaker(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_circuit_breaker_not_found", "circuit breaker 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.webhook.circuit.manual_closed", fmt.Sprintf("webhook-circuit-breakers#%d", id), nil, gin.H{"status": "requested"}, gin.H{
		"circuit_breaker_id": id,
		"actor":              actor,
	})
	out, err := s.svc.CloseWebhookCircuitBreaker(c.Request.Context(), id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.circuit.closed", fmt.Sprintf("webhook-circuit-breakers#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"circuit_breaker_id": id,
			"actor":              actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.webhook.circuit.closed", fmt.Sprintf("webhook-circuit-breakers#%d", id), nil, gin.H{"status": out.Status}, gin.H{
		"circuit_breaker_id": id,
		"actor":              actor,
	})
	c.JSON(http.StatusOK, out)
}

func (s *Server) openAdminWebhookCircuitBreaker(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		failAPIError(c, domain.NewPluginError("webhook_circuit_breaker_not_found", "circuit breaker 不存在").WithStatus(404))
		return
	}
	actor := auditActor(c)
	s.auditStructured(c, "system", "plugin.webhook.circuit.manual_opened", fmt.Sprintf("webhook-circuit-breakers#%d", id), nil, gin.H{"status": "requested"}, gin.H{
		"circuit_breaker_id": id,
		"actor":              actor,
	})
	out, err := s.svc.OpenWebhookCircuitBreaker(c.Request.Context(), id)
	if err != nil {
		s.auditStructured(c, "system", "plugin.webhook.circuit.opened", fmt.Sprintf("webhook-circuit-breakers#%d", id), nil, gin.H{"status": "failed"}, mergeAuditMeta(gin.H{
			"circuit_breaker_id": id,
			"actor":              actor,
		}, auditAPIErrorFields(err)))
		failAPIError(c, err)
		return
	}
	s.auditStructured(c, "system", "plugin.webhook.circuit.opened", fmt.Sprintf("webhook-circuit-breakers#%d", id), nil, gin.H{"status": out.Status}, gin.H{
		"circuit_breaker_id": id,
		"actor":              actor,
	})
	c.JSON(http.StatusOK, out)
}
