package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// RegulatoryHandler handles HTTP requests for the Regulatory Change
// Management & Horizon Scanning module, including browsing changes,
// managing sources and subscriptions, impact assessments, and dashboards.
type RegulatoryHandler struct {
	svc *service.RegulatoryService
}

// NewRegulatoryHandler creates a new RegulatoryHandler.
func NewRegulatoryHandler(svc *service.RegulatoryService) *RegulatoryHandler {
	return &RegulatoryHandler{svc: svc}
}

// RegisterRoutes mounts all regulatory change management routes on the router.
func (h *RegulatoryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/changes", h.ListChanges)
	r.Get("/changes/{id}", h.GetChange)
	r.Post("/changes/{id}/assess", h.AssessImpact)
	r.Get("/changes/{id}/assessment", h.GetAssessment)
	r.Post("/changes/{id}/respond", h.CreateResponsePlan)

	r.Get("/sources", h.ListSources)
	r.Post("/sources", h.CreateSource)

	r.Get("/subscriptions", h.GetSubscriptions)
	r.Post("/subscriptions", h.Subscribe)
	r.Delete("/subscriptions/{id}", h.Unsubscribe)

	r.Get("/dashboard", h.Dashboard)
	r.Get("/timeline", h.Timeline)
}

// ============================================================
// CHANGES
// ============================================================

// ListChanges returns a filtered, paginated list of regulatory changes.
// GET /changes
func (h *RegulatoryHandler) ListChanges(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filters := service.ChangeFilters{
		Severity:   q.Get("severity"),
		Framework:  q.Get("framework"),
		Region:     q.Get("region"),
		Status:     q.Get("status"),
		ChangeType: q.Get("change_type"),
		Search:     q.Get("search"),
		Page:       1,
		PageSize:   20,
	}

	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filters.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filters.PageSize = ps
	}
	if sid := q.Get("source_id"); sid != "" {
		if parsed, err := uuid.Parse(sid); err == nil {
			filters.SourceID = &parsed
		}
	}
	if df := q.Get("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			filters.DateFrom = &t
		}
	}
	if dt := q.Get("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			filters.DateTo = &t
		}
	}

	changes, total, err := h.svc.ListChanges(r.Context(), filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list regulatory changes")
		return
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: changes,
		Pagination: models.Pagination{
			Page:       filters.Page,
			PageSize:   filters.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    filters.Page < totalPages,
			HasPrev:    filters.Page > 1,
		},
	})
}

// GetChange returns a single regulatory change with full details.
// GET /changes/{id}
func (h *RegulatoryHandler) GetChange(w http.ResponseWriter, r *http.Request) {
	changeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid regulatory change ID")
		return
	}

	change, err := h.svc.GetChange(r.Context(), changeID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Regulatory change not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: change})
}

// ============================================================
// IMPACT ASSESSMENT
// ============================================================

// AssessImpact creates or updates the impact assessment for a regulatory change.
// POST /changes/{id}/assess
func (h *RegulatoryHandler) AssessImpact(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	changeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid regulatory change ID")
		return
	}

	var req service.AssessReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	assessment, err := h.svc.AssessImpact(r.Context(), orgID, changeID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to assess impact: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: assessment})
}

// GetAssessment returns the impact assessment for a regulatory change.
// GET /changes/{id}/assessment
func (h *RegulatoryHandler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	changeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid regulatory change ID")
		return
	}

	assessment, err := h.svc.GetAssessment(r.Context(), orgID, changeID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Impact assessment not found for this change")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: assessment})
}

// CreateResponsePlan creates a placeholder response plan for a regulatory change.
// POST /changes/{id}/respond
func (h *RegulatoryHandler) CreateResponsePlan(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	changeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid regulatory change ID")
		return
	}

	planID, err := h.svc.CreateResponsePlan(r.Context(), orgID, changeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create response plan")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"response_plan_id": planID,
			"change_id":        changeID,
			"message":          "Response plan created. Use the plan ID to link remediation actions.",
		},
	})
}

// ============================================================
// SOURCES
// ============================================================

// ListSources returns all registered regulatory sources.
// GET /sources
func (h *RegulatoryHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.svc.ListSources(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list regulatory sources")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: sources})
}

// CreateSource registers a new regulatory source for scanning.
// POST /sources
func (h *RegulatoryHandler) CreateSource(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSourceReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	source, err := h.svc.CreateSource(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create regulatory source")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: source})
}

// ============================================================
// SUBSCRIPTIONS
// ============================================================

// GetSubscriptions returns all subscriptions for the current organisation.
// GET /subscriptions
func (h *RegulatoryHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	subs, err := h.svc.GetRegSubscriptions(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list subscriptions")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: subs})
}

// Subscribe creates a new subscription to a regulatory source.
// POST /subscriptions
func (h *RegulatoryHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.SubscribeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.SourceID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "source_id is required")
		return
	}

	sub, err := h.svc.Subscribe(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to subscribe: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: sub})
}

// Unsubscribe removes a subscription.
// DELETE /subscriptions/{id}
func (h *RegulatoryHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	subID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid subscription ID")
		return
	}

	if err := h.svc.Unsubscribe(r.Context(), orgID, subID); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Subscription not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Unsubscribed successfully"},
	})
}

// ============================================================
// DASHBOARD & TIMELINE
// ============================================================

// Dashboard returns aggregated regulatory change metrics.
// GET /dashboard
func (h *RegulatoryHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dash, err := h.svc.GetDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate regulatory dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dash})
}

// Timeline returns a chronological view of regulatory events.
// GET /timeline
func (h *RegulatoryHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	months := 6
	if m, err := strconv.Atoi(r.URL.Query().Get("months")); err == nil && m > 0 && m <= 24 {
		months = m
	}

	entries, err := h.svc.GetTimeline(r.Context(), orgID, months)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate timeline")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: entries})
}
