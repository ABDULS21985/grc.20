package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// CalendarHandler handles HTTP requests for the Compliance Calendar
// & Deadline Management module.
type CalendarHandler struct {
	svc *service.CalendarService
}

// NewCalendarHandler creates a new CalendarHandler.
func NewCalendarHandler(svc *service.CalendarService) *CalendarHandler {
	return &CalendarHandler{svc: svc}
}

// RegisterRoutes mounts all calendar routes on the router.
func (h *CalendarHandler) RegisterRoutes(r chi.Router) {
	// Events
	r.Get("/events", h.ListEvents)
	r.Post("/events", h.CreateEvent)
	r.Get("/events/{id}", h.GetEvent)
	r.Put("/events/{id}/complete", h.CompleteEvent)
	r.Put("/events/{id}/reschedule", h.RescheduleEvent)
	r.Put("/events/{id}/assign", h.AssignEvent)

	// Dashboards
	r.Get("/deadlines", h.GetUpcomingDeadlines)
	r.Get("/overdue", h.GetOverdueItems)
	r.Get("/summary", h.GetSummary)

	// Subscriptions
	r.Get("/subscriptions", h.GetSubscription)
	r.Put("/subscriptions", h.UpdateSubscription)

	// Sync
	r.Get("/sync/status", h.GetSyncStatus)
	r.Post("/sync/trigger", h.TriggerSync)
	r.Put("/sync/configs", h.UpdateSyncConfigs)
}

// RegisterPublicRoutes mounts the public iCal feed route.
// This should be mounted outside JWT-protected routes.
func (h *CalendarHandler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/ical/{token}", h.ExportICalFeed)
}

// ============================================================
// EVENTS
// ============================================================

// ListEvents returns a filtered, paginated list of calendar events.
// GET /events?start_date=...&end_date=...&category=...&priority=...&status=...&search=...&page=1&page_size=50
func (h *CalendarHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	q := r.URL.Query()

	filter := service.CalendarViewFilter{
		StartDate: q.Get("start_date"),
		EndDate:   q.Get("end_date"),
		Search:    q.Get("search"),
		Page:      1,
		PageSize:  50,
	}

	if cats := q.Get("category"); cats != "" {
		filter.Categories = strings.Split(cats, ",")
	}
	if pris := q.Get("priority"); pris != "" {
		filter.Priorities = strings.Split(pris, ",")
	}
	if statuses := q.Get("status"); statuses != "" {
		filter.Statuses = strings.Split(statuses, ",")
	}
	if types := q.Get("event_type"); types != "" {
		filter.EventTypes = strings.Split(types, ",")
	}
	if assigned := q.Get("assigned_to"); assigned != "" {
		if parsed, err := uuid.Parse(assigned); err == nil {
			filter.AssignedTo = &parsed
		}
	}
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}

	// Default date range: current month if not specified
	if filter.StartDate == "" {
		now := time.Now()
		filter.StartDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}
	if filter.EndDate == "" {
		now := time.Now()
		lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC)
		filter.EndDate = lastDay.Format("2006-01-02")
	}

	result, err := h.svc.GetCalendarView(r.Context(), orgID, userID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list calendar events")
		return
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: result.Events,
		Pagination: models.Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalItems: result.Total,
			TotalPages: result.TotalPages,
			HasNext:    result.Page < result.TotalPages,
			HasPrev:    result.Page > 1,
		},
	})
}

// GetEvent returns a single calendar event by ID.
// GET /events/{id}
func (h *CalendarHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid event ID")
		return
	}

	event, err := h.svc.GetEvent(r.Context(), orgID, eventID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Calendar event not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: event})
}

// CreateEvent creates a new calendar event.
// POST /events
func (h *CalendarHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateCalendarEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title is required")
		return
	}
	if req.DueDate == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "due_date is required")
		return
	}
	if req.EventType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "event_type is required")
		return
	}
	if req.Category == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "category is required")
		return
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", req.DueDate); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_DATE", "due_date must be YYYY-MM-DD format")
		return
	}

	// Default source entity for custom events
	if req.SourceEntityType == "" {
		req.SourceEntityType = "custom"
	}
	if req.SourceEntityID == uuid.Nil {
		req.SourceEntityID = uuid.New()
	}

	event, err := h.svc.CreateEvent(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create calendar event: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: event})
}

// CompleteEvent marks a calendar event as completed.
// PUT /events/{id}/complete
func (h *CalendarHandler) CompleteEvent(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid event ID")
		return
	}

	var req service.CompleteEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Body is optional for completion
		req = service.CompleteEventRequest{}
	}

	if err := h.svc.CompleteEvent(r.Context(), orgID, eventID, userID, req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Calendar event not found or already completed")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to complete event: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Event completed successfully"},
	})
}

// RescheduleEvent changes the due date of a calendar event.
// PUT /events/{id}/reschedule
func (h *CalendarHandler) RescheduleEvent(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid event ID")
		return
	}

	var req service.RescheduleEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.NewDueDate == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "new_due_date is required")
		return
	}
	if _, err := time.Parse("2006-01-02", req.NewDueDate); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_DATE", "new_due_date must be YYYY-MM-DD format")
		return
	}

	if err := h.svc.RescheduleEvent(r.Context(), orgID, eventID, req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Calendar event not found or already completed")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reschedule event: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Event rescheduled successfully"},
	})
}

// AssignEvent assigns a calendar event to a user.
// PUT /events/{id}/assign
func (h *CalendarHandler) AssignEvent(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid event ID")
		return
	}

	var req service.AssignEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.AssignedToUserID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "assigned_to_user_id is required")
		return
	}

	if err := h.svc.AssignEvent(r.Context(), orgID, eventID, req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Calendar event not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to assign event: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Event assigned successfully"},
	})
}

// ============================================================
// DASHBOARD / QUERIES
// ============================================================

// GetUpcomingDeadlines returns events due within a specified number of days.
// GET /deadlines?days=30&limit=20
func (h *CalendarHandler) GetUpcomingDeadlines(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	days := 30
	if d, err := strconv.Atoi(r.URL.Query().Get("days")); err == nil && d > 0 && d <= 365 {
		days = d
	}

	limit := 20
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	deadlines, err := h.svc.GetUpcomingDeadlines(r.Context(), orgID, userID, days, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get upcoming deadlines")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: deadlines})
}

// GetOverdueItems returns all overdue calendar events.
// GET /overdue
func (h *CalendarHandler) GetOverdueItems(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	items, err := h.svc.GetOverdueItems(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get overdue items")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: items})
}

// GetSummary returns aggregated compliance calendar statistics for a month.
// GET /summary?month=2026-03
func (h *CalendarHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	// Validate month format
	if _, err := time.Parse("2006-01", month); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_MONTH", "month must be YYYY-MM format")
		return
	}

	summary, err := h.svc.GetComplianceCalendarSummary(r.Context(), orgID, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get calendar summary")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: summary})
}

// ============================================================
// SUBSCRIPTIONS
// ============================================================

// GetSubscription returns the current user's calendar subscription.
// GET /subscriptions
func (h *CalendarHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	sub, err := h.svc.GetSubscription(r.Context(), orgID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get calendar subscription")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: sub})
}

// UpdateSubscription updates the current user's calendar subscription.
// PUT /subscriptions
func (h *CalendarHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req service.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	sub, err := h.svc.UpdateSubscription(r.Context(), orgID, userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update subscription: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: sub})
}

// ============================================================
// iCAL FEED (PUBLIC — token-auth, no JWT)
// ============================================================

// ExportICalFeed exports an iCal feed for a user identified by token.
// GET /ical/{token}
func (h *CalendarHandler) ExportICalFeed(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	ical, err := h.svc.ExportICalFeed(r.Context(), token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=complianceforge-calendar.ics")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(ical))
}

// ============================================================
// SYNC
// ============================================================

// GetSyncStatus returns the sync status across all modules.
// GET /sync/status
func (h *CalendarHandler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	status, err := h.svc.GetSyncStatus(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get sync status")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: status})
}

// TriggerSync manually triggers a full calendar sync for all enabled modules.
// POST /sync/trigger
func (h *CalendarHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	results, err := h.svc.SyncAllModules(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to trigger sync: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":      "Sync completed",
			"sync_results": results,
		},
	})
}

// UpdateSyncConfigs updates the sync configuration for one or more modules.
// PUT /sync/configs
func (h *CalendarHandler) UpdateSyncConfigs(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.UpdateSyncConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if len(req.Configs) == 0 {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "configs array is required")
		return
	}

	if err := h.svc.UpdateSyncConfigs(r.Context(), orgID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update sync configs: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Sync configurations updated successfully"},
	})
}
