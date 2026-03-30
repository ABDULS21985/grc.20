package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// DEVELOPER HANDLER
// HTTP layer for the Developer Portal. Provides API key
// management, webhook subscriptions, sandbox environments,
// and developer documentation endpoints.
// ============================================================

// DeveloperHandler handles developer portal HTTP requests.
type DeveloperHandler struct {
	portal   *service.DeveloperPortalService
	webhooks *service.WebhookService
}

// NewDeveloperHandler creates a new DeveloperHandler.
func NewDeveloperHandler(portal *service.DeveloperPortalService, webhooks *service.WebhookService) *DeveloperHandler {
	return &DeveloperHandler{portal: portal, webhooks: webhooks}
}

// RegisterRoutes mounts all developer portal routes on the given chi router.
// The caller is expected to wrap these under /developer.
func (h *DeveloperHandler) RegisterRoutes(r chi.Router) {
	// ── API Keys ──
	r.Get("/api-keys", h.ListAPIKeys)
	r.Post("/api-keys", h.CreateAPIKey)
	r.Put("/api-keys/{id}", h.UpdateAPIKey)
	r.Delete("/api-keys/{id}", h.RevokeAPIKey)
	r.Get("/api-keys/{id}/usage", h.GetAPIKeyUsage)

	// ── Webhooks ──
	r.Get("/webhooks", h.ListWebhooks)
	r.Post("/webhooks", h.CreateWebhook)
	r.Put("/webhooks/{id}", h.UpdateWebhook)
	r.Delete("/webhooks/{id}", h.DeleteWebhook)
	r.Post("/webhooks/{id}/test", h.TestWebhook)
	r.Get("/webhooks/{id}/deliveries", h.GetWebhookDeliveries)
	r.Post("/webhooks/deliveries/{id}/replay", h.ReplayDelivery)

	// ── Sandbox ──
	r.Post("/sandbox", h.CreateSandbox)
	r.Get("/sandbox", h.GetSandbox)
	r.Delete("/sandbox", h.DestroySandbox)

	// ── Documentation ──
	r.Get("/events", h.ListEventTypes)
	r.Get("/scopes", h.ListScopes)
}

// ============================================================
// API KEY ENDPOINTS
// ============================================================

// ListAPIKeys handles GET /developer/api-keys
func (h *DeveloperHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	keys, err := h.portal.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list API keys: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: keys})
}

// CreateAPIKey handles POST /developer/api-keys
func (h *DeveloperHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var req service.GenerateAPIKeyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	key, err := h.portal.GenerateAPIKey(r.Context(), orgID, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "KEY_GENERATION_FAILED"
		if strings.Contains(err.Error(), "required") {
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    key,
	})
}

// UpdateAPIKey handles PUT /developer/api-keys/{id}
func (h *DeveloperHandler) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid API key ID")
		return
	}

	var req service.UpdateAPIKeyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	updated, err := h.portal.UpdateAPIKey(r.Context(), orgID, keyID, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "UPDATE_FAILED"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: updated})
}

// RevokeAPIKey handles DELETE /developer/api-keys/{id}
func (h *DeveloperHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid API key ID")
		return
	}

	if err := h.portal.RevokeAPIKey(r.Context(), orgID, keyID); err != nil {
		status := http.StatusInternalServerError
		code := "REVOKE_FAILED"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "API key revoked successfully"},
	})
}

// GetAPIKeyUsage handles GET /developer/api-keys/{id}/usage
func (h *DeveloperHandler) GetAPIKeyUsage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid API key ID")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}

	stats, err := h.portal.GetAPIUsageStats(r.Context(), orgID, keyID, period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get usage stats: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: stats})
}

// ============================================================
// WEBHOOK ENDPOINTS
// ============================================================

// ListWebhooks handles GET /developer/webhooks
func (h *DeveloperHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	subs, err := h.webhooks.ListSubscriptions(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list webhooks: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: subs})
}

// CreateWebhook handles POST /developer/webhooks
func (h *DeveloperHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	var req service.CreateSubscriptionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	sub, err := h.webhooks.Subscribe(r.Context(), orgID, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "SUBSCRIPTION_FAILED"
		if strings.Contains(err.Error(), "HTTPS") {
			status = http.StatusBadRequest
			code = "HTTPS_REQUIRED"
		} else if strings.Contains(err.Error(), "at least one") {
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    sub,
	})
}

// UpdateWebhook handles PUT /developer/webhooks/{id}
func (h *DeveloperHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	subID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid webhook ID")
		return
	}

	var req service.UpdateSubscriptionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	sub, err := h.webhooks.UpdateSubscription(r.Context(), orgID, subID, req)
	if err != nil {
		status := http.StatusInternalServerError
		code := "UPDATE_FAILED"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		} else if strings.Contains(err.Error(), "HTTPS") {
			status = http.StatusBadRequest
			code = "HTTPS_REQUIRED"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: sub})
}

// DeleteWebhook handles DELETE /developer/webhooks/{id}
func (h *DeveloperHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	subID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid webhook ID")
		return
	}

	if err := h.webhooks.DeleteSubscription(r.Context(), orgID, subID); err != nil {
		status := http.StatusInternalServerError
		code := "DELETE_FAILED"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Webhook subscription deleted successfully"},
	})
}

// TestWebhook handles POST /developer/webhooks/{id}/test
func (h *DeveloperHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	subID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid webhook ID")
		return
	}

	if err := h.webhooks.PingWebhook(r.Context(), orgID, subID); err != nil {
		status := http.StatusInternalServerError
		code := "PING_FAILED"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Ping event queued for delivery"},
	})
}

// GetWebhookDeliveries handles GET /developer/webhooks/{id}/deliveries
func (h *DeveloperHandler) GetWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	subID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid webhook ID")
		return
	}

	page := 1
	pageSize := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	deliveries, total, err := h.webhooks.GetDeliveryHistory(r.Context(), orgID, subID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get deliveries: "+err.Error())
		return
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: deliveries,
		Pagination: models.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: int64(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	})
}

// ReplayDelivery handles POST /developer/webhooks/deliveries/{id}/replay
func (h *DeveloperHandler) ReplayDelivery(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	deliveryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid delivery ID")
		return
	}

	if err := h.webhooks.ReplayDelivery(r.Context(), orgID, deliveryID); err != nil {
		status := http.StatusInternalServerError
		code := "REPLAY_FAILED"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Delivery replayed successfully"},
	})
}

// ============================================================
// SANDBOX ENDPOINTS
// ============================================================

// CreateSandbox handles POST /developer/sandbox
func (h *DeveloperHandler) CreateSandbox(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	sb, err := h.portal.CreateSandbox(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SANDBOX_FAILED", "Failed to create sandbox: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    sb,
	})
}

// GetSandbox handles GET /developer/sandbox
func (h *DeveloperHandler) GetSandbox(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	sb, err := h.portal.GetSandbox(r.Context(), orgID)
	if err != nil {
		status := http.StatusInternalServerError
		code := "INTERNAL_ERROR"
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: sb})
}

// DestroySandbox handles DELETE /developer/sandbox
func (h *DeveloperHandler) DestroySandbox(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Organisation context required")
		return
	}

	if err := h.portal.DestroySandbox(r.Context(), orgID); err != nil {
		status := http.StatusInternalServerError
		code := "DESTROY_FAILED"
		if strings.Contains(err.Error(), "no active") {
			status = http.StatusNotFound
			code = "NOT_FOUND"
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Sandbox destroyed successfully"},
	})
}

// ============================================================
// DOCUMENTATION ENDPOINTS
// ============================================================

// ListEventTypes handles GET /developer/events
func (h *DeveloperHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	events := h.portal.ListWebhookEventTypes()
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: events})
}

// ListScopes handles GET /developer/scopes
func (h *DeveloperHandler) ListScopes(w http.ResponseWriter, r *http.Request) {
	scopes := h.portal.ListAPIScopes()
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: scopes})
}
