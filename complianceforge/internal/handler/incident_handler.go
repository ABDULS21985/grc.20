package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/repository"
)

type IncidentHandler struct {
	repo *repository.IncidentRepo
	db   *database.DB
}

func NewIncidentHandler(repo *repository.IncidentRepo, db *database.DB) *IncidentHandler {
	return &IncidentHandler{repo: repo, db: db}
}

// ListIncidents returns paginated incidents.
// GET /api/v1/incidents
func (h *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	incidents, total, err := h.repo.List(r.Context(), orgID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve incidents")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: incidents,
		Pagination: models.Pagination{
			Page: params.Page, PageSize: params.PageSize,
			TotalItems: total, TotalPages: totalPages,
			HasNext: params.Page < totalPages, HasPrev: params.Page > 1,
		},
	})
}

// GetIncident returns full incident detail.
// GET /api/v1/incidents/{id}
func (h *IncidentHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	incID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid incident ID")
		return
	}

	incident, err := h.repo.GetByID(r.Context(), orgID, incID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Incident not found")
		return
	}

	// Calculate breach notification status
	if incident.IsDataBreach && incident.NotificationDeadline != nil {
		remaining := time.Until(*incident.NotificationDeadline)
		extraData := map[string]interface{}{
			"incident":                     incident,
			"breach_notification_hours_remaining": remaining.Hours(),
			"breach_notification_overdue":         remaining < 0,
		}
		if incident.DPANotifiedAt != nil {
			extraData["dpa_notification_status"] = "completed"
		} else {
			extraData["dpa_notification_status"] = "pending"
		}
		writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: extraData})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: incident})
}

// ReportIncident creates a new incident.
// POST /api/v1/incidents
func (h *IncidentHandler) ReportIncident(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Title                  string   `json:"title"`
		Description            string   `json:"description"`
		IncidentType           string   `json:"incident_type"`
		Severity               string   `json:"severity"`
		IsDataBreach           bool     `json:"is_data_breach"`
		DataSubjectsAffected   int      `json:"data_subjects_affected"`
		DataCategoriesAffected []string `json:"data_categories_affected"`
		IsNIS2Reportable       bool     `json:"is_nis2_reportable"`
		ImpactDescription      string   `json:"impact_description"`
		AssignedTo             *uuid.UUID `json:"assigned_to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" || req.Description == "" || req.IncidentType == "" || req.Severity == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title, description, incident_type, severity required")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), orgID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Transaction failed")
		return
	}
	defer tx.Rollback(r.Context())

	now := time.Now()
	incident := &models.Incident{
		BaseModel:              models.BaseModel{OrganizationID: orgID},
		Title:                  req.Title,
		Description:            req.Description,
		IncidentType:           req.IncidentType,
		Severity:               models.IncidentSeverity(req.Severity),
		Status:                 models.IncidentStatusOpen,
		ReportedBy:             userID,
		ReportedAt:             now,
		AssignedTo:             req.AssignedTo,
		IsDataBreach:           req.IsDataBreach,
		DataSubjectsAffected:   req.DataSubjectsAffected,
		DataCategoriesAffected: req.DataCategoriesAffected,
		IsNIS2Reportable:       req.IsNIS2Reportable,
		ImpactDescription:      req.ImpactDescription,
		Metadata:               models.JSONB("{}"),
	}

	if err := h.repo.Create(r.Context(), tx, incident); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to report incident")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Commit failed")
		return
	}

	// Build response with regulatory context
	resp := map[string]interface{}{
		"incident": incident,
	}
	if incident.IsDataBreach && incident.NotificationDeadline != nil {
		resp["gdpr_breach_alert"] = map[string]interface{}{
			"notification_deadline":    incident.NotificationDeadline,
			"hours_remaining":          time.Until(*incident.NotificationDeadline).Hours(),
			"action_required":          "Notify supervisory authority within 72 hours per GDPR Article 33",
			"data_subjects_affected":   incident.DataSubjectsAffected,
		}
	}
	if incident.IsNIS2Reportable {
		resp["nis2_alert"] = map[string]interface{}{
			"early_warning_deadline": "24 hours from detection",
			"notification_deadline":  "72 hours from detection",
			"final_report_deadline":  "1 month from notification",
			"action_required":        "Submit early warning to CSIRT within 24 hours per NIS2 Article 23",
		}
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: resp})
}

// NotifyDPA records that the supervisory authority has been notified of a data breach.
// POST /api/v1/incidents/{id}/notify-dpa
func (h *IncidentHandler) NotifyDPA(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	incID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid incident ID")
		return
	}

	if err := h.repo.RecordDPANotification(r.Context(), orgID, incID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record DPA notification")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"incident_id":    incID,
			"dpa_notified":   true,
			"notified_at":    time.Now(),
			"message":        "DPA notification recorded per GDPR Article 33",
		},
	})
}

// SubmitNIS2EarlyWarning records the NIS2 24-hour early warning submission.
// POST /api/v1/incidents/{id}/nis2-early-warning
func (h *IncidentHandler) SubmitNIS2EarlyWarning(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	incID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid incident ID")
		return
	}

	if err := h.repo.RecordNIS2EarlyWarning(r.Context(), orgID, incID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record NIS2 early warning")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"incident_id":       incID,
			"early_warning":     true,
			"submitted_at":      time.Now(),
			"next_step":         "Submit full NIS2 incident notification within 72 hours",
			"message":           "NIS2 early warning recorded per Article 23(4)(a)",
		},
	})
}

// GetBreachesNearDeadline returns data breaches approaching the 72-hour GDPR deadline.
// GET /api/v1/incidents/breaches/urgent
func (h *IncidentHandler) GetBreachesNearDeadline(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	breaches, err := h.repo.GetBreachesNearingDeadline(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check breaches")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: breaches})
}

// GetIncidentStats returns incident dashboard metrics.
// GET /api/v1/incidents/stats
func (h *IncidentHandler) GetIncidentStats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	stats, err := h.repo.GetDashboardStats(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get stats")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: stats})
}
