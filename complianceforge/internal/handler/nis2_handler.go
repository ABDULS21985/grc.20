package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// NIS2Handler handles HTTP requests for the NIS2 Compliance Automation Module.
// It covers entity assessment, 3-phase incident reporting (Article 23),
// security measures (Article 21), management accountability (Article 20),
// and the NIS2 compliance dashboard.
type NIS2Handler struct {
	svc *service.NIS2Service
}

// NewNIS2Handler creates a new NIS2Handler with the given service.
func NewNIS2Handler(svc *service.NIS2Service) *NIS2Handler {
	return &NIS2Handler{svc: svc}
}

// ============================================================
// ENTITY ASSESSMENT
// ============================================================

// GetAssessment returns the most recent NIS2 entity assessment.
// GET /api/v1/nis2/assessment
func (h *NIS2Handler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	assessment, err := h.svc.GetEntityAssessment(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "No NIS2 entity assessment found. Submit one to determine your entity classification.")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: assessment})
}

// SubmitAssessment creates a new entity categorisation assessment.
// POST /api/v1/nis2/assessment
func (h *NIS2Handler) SubmitAssessment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var input service.AssessEntityTypeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.Sector == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "sector is required")
		return
	}

	assessment, err := h.svc.AssessEntityType(r.Context(), orgID, userID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to perform entity assessment")
		return
	}

	// Build response with regulatory guidance
	resp := map[string]interface{}{
		"assessment": assessment,
		"guidance":   buildEntityGuidance(assessment.EntityType, assessment.IsInScope),
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: resp})
}

// buildEntityGuidance returns regulatory guidance based on entity classification.
func buildEntityGuidance(entityType string, inScope bool) map[string]interface{} {
	guidance := map[string]interface{}{
		"entity_type": entityType,
		"in_scope":    inScope,
	}

	switch entityType {
	case "essential":
		guidance["supervision_regime"] = "Proactive (ex-ante) supervision by competent authority"
		guidance["penalties_max_fine"] = "EUR 10,000,000 or 2% of worldwide annual turnover"
		guidance["incident_reporting"] = "Mandatory 3-phase reporting to CSIRT (24h/72h/1 month)"
		guidance["obligations"] = []string{
			"Implement all 10 Article 21 cybersecurity measures",
			"Report significant incidents within 24 hours (early warning)",
			"Board members must undergo cybersecurity training (Article 20)",
			"Board must approve cybersecurity risk-management measures",
			"Subject to regular security audits",
			"Register with competent authority",
		}
	case "important":
		guidance["supervision_regime"] = "Reactive (ex-post) supervision by competent authority"
		guidance["penalties_max_fine"] = "EUR 7,000,000 or 1.4% of worldwide annual turnover"
		guidance["incident_reporting"] = "Mandatory 3-phase reporting to CSIRT (24h/72h/1 month)"
		guidance["obligations"] = []string{
			"Implement all 10 Article 21 cybersecurity measures",
			"Report significant incidents within 24 hours (early warning)",
			"Board members must undergo cybersecurity training (Article 20)",
			"Board must approve cybersecurity risk-management measures",
			"Register with competent authority",
		}
	default:
		guidance["supervision_regime"] = "Not subject to NIS2 supervision"
		guidance["penalties_max_fine"] = "N/A"
		guidance["incident_reporting"] = "Not required under NIS2"
		guidance["obligations"] = []string{
			"Consider voluntary adoption of NIS2 measures as best practice",
		}
	}

	return guidance
}

// ============================================================
// INCIDENT REPORTS (3-Phase)
// ============================================================

// ListIncidentReports returns all NIS2 incident reports for the organisation.
// GET /api/v1/nis2/incidents
func (h *NIS2Handler) ListIncidentReports(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	reports, err := h.svc.ListIncidentReports(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve NIS2 incident reports")
		return
	}

	if reports == nil {
		reports = []service.NIS2IncidentReport{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: reports})
}

// GetIncidentReport returns a single NIS2 3-phase incident report with all phase details.
// GET /api/v1/nis2/incidents/{id}
func (h *NIS2Handler) GetIncidentReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid NIS2 report ID")
		return
	}

	report, err := h.svc.GetIncidentReport(r.Context(), orgID, reportID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "NIS2 incident report not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: report})
}

// SubmitEarlyWarning records Phase 1 (24-hour) early warning submission.
// POST /api/v1/nis2/incidents/{id}/early-warning
func (h *NIS2Handler) SubmitEarlyWarning(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid NIS2 report ID")
		return
	}

	var input service.EarlyWarningInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	report, err := h.svc.SubmitEarlyWarning(r.Context(), orgID, reportID, userID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to submit early warning")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"report":  report,
			"phase":   "early_warning",
			"message": "Phase 1 Early Warning submitted per NIS2 Article 23(4)(a). Next: submit incident notification within 72 hours.",
		},
	})
}

// SubmitNotification records Phase 2 (72-hour) incident notification.
// POST /api/v1/nis2/incidents/{id}/notification
func (h *NIS2Handler) SubmitNotification(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid NIS2 report ID")
		return
	}

	var input service.NotificationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	report, err := h.svc.SubmitNotification(r.Context(), orgID, reportID, userID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to submit notification")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"report":  report,
			"phase":   "notification",
			"message": "Phase 2 Incident Notification submitted per NIS2 Article 23(4)(b). Next: submit final report within 1 month.",
		},
	})
}

// SubmitFinalReport records Phase 3 (1-month) final report.
// POST /api/v1/nis2/incidents/{id}/final-report
func (h *NIS2Handler) SubmitFinalReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid NIS2 report ID")
		return
	}

	var input service.FinalReportInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	report, err := h.svc.SubmitFinalReport(r.Context(), orgID, reportID, userID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to submit final report")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"report":  report,
			"phase":   "final_report",
			"message": "Phase 3 Final Report submitted per NIS2 Article 23(4)(d). All NIS2 reporting phases complete.",
		},
	})
}

// ============================================================
// SECURITY MEASURES (Article 21)
// ============================================================

// ListMeasures returns the status of all 10 NIS2 Article 21 security measures.
// GET /api/v1/nis2/measures
func (h *NIS2Handler) ListMeasures(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	measures, err := h.svc.GetSecurityMeasuresStatus(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve security measures")
		return
	}

	if measures == nil {
		measures = []service.NIS2SecurityMeasure{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: measures})
}

// UpdateMeasure updates the implementation status of a single security measure.
// PUT /api/v1/nis2/measures/{id}
func (h *NIS2Handler) UpdateMeasure(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	measureID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid measure ID")
		return
	}

	var input service.UpdateMeasureInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	validStatuses := map[string]bool{
		"not_started": true, "in_progress": true,
		"implemented": true, "verified": true,
	}
	if !validStatuses[input.ImplementationStatus] {
		writeError(w, http.StatusBadRequest, "INVALID_STATUS",
			"implementation_status must be one of: not_started, in_progress, implemented, verified")
		return
	}

	measure, err := h.svc.UpdateSecurityMeasure(r.Context(), orgID, measureID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update security measure")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: measure})
}

// ============================================================
// MANAGEMENT ACCOUNTABILITY (Article 20)
// ============================================================

// ListManagement returns all management accountability records.
// GET /api/v1/nis2/management
func (h *NIS2Handler) ListManagement(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	records, err := h.svc.ListManagementRecords(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve management records")
		return
	}

	if records == nil {
		records = []service.NIS2ManagementRecord{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: records})
}

// RecordManagement creates a management training or approval record.
// POST /api/v1/nis2/management
func (h *NIS2Handler) RecordManagement(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var input service.ManagementTrainingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if input.BoardMemberName == "" || input.BoardMemberRole == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "board_member_name and board_member_role are required")
		return
	}

	record, err := h.svc.RecordManagementTraining(r.Context(), orgID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record management training")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"record":  record,
			"message": "Management cybersecurity training recorded per NIS2 Article 20(2).",
		},
	})
}

// ============================================================
// DASHBOARD
// ============================================================

// GetDashboard returns aggregated NIS2 compliance metrics.
// GET /api/v1/nis2/dashboard
func (h *NIS2Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dashboard, err := h.svc.GetComplianceDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate NIS2 dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}
