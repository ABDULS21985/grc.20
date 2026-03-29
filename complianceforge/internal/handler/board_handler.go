package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// BoardHandler handles HTTP requests for the Executive Board Reporting
// Portal & Governance Dashboards module.
type BoardHandler struct {
	svc *service.BoardService
}

// NewBoardHandler creates a new BoardHandler.
func NewBoardHandler(svc *service.BoardService) *BoardHandler {
	return &BoardHandler{svc: svc}
}

// RegisterRoutes mounts all board reporting routes on the router.
func (h *BoardHandler) RegisterRoutes(r chi.Router) {
	// Members
	r.Get("/members", h.ListBoardMembers)
	r.Post("/members", h.CreateBoardMember)
	r.Put("/members/{id}", h.UpdateBoardMember)

	// Meetings
	r.Get("/meetings", h.ListMeetings)
	r.Post("/meetings", h.CreateMeeting)
	r.Get("/meetings/{id}", h.GetMeeting)
	r.Put("/meetings/{id}", h.UpdateMeeting)
	r.Post("/meetings/{id}/generate-pack", h.GenerateBoardPack)
	r.Get("/meetings/{id}/download-pack", h.DownloadBoardPack)

	// Decisions
	r.Post("/decisions", h.RecordDecision)
	r.Get("/decisions", h.ListDecisions)
	r.Put("/decisions/{id}/action", h.UpdateDecisionAction)

	// Reports
	r.Get("/reports", h.ListReports)
	r.Post("/reports/generate", h.GenerateReport)

	// Dashboard & NIS2
	r.Get("/dashboard", h.GetBoardDashboard)
	r.Get("/nis2-governance", h.GetNIS2GovernanceReport)
}

// ============================================================
// MEMBERS
// ============================================================

// ListBoardMembers returns all board members for the organisation.
// GET /members
func (h *BoardHandler) ListBoardMembers(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	members, err := h.svc.ListBoardMembers(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list board members")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: members})
}

// CreateBoardMember creates a new board member.
// POST /members
func (h *BoardHandler) CreateBoardMember(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	member, err := h.svc.CreateBoardMember(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create board member: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: member})
}

// UpdateBoardMember updates an existing board member.
// PUT /members/{id}
func (h *BoardHandler) UpdateBoardMember(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	memberID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid member ID")
		return
	}

	var req service.UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.svc.UpdateBoardMember(r.Context(), orgID, memberID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update board member: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Board member updated successfully"},
	})
}

// ============================================================
// MEETINGS
// ============================================================

// ListMeetings returns a filtered list of board meetings.
// GET /meetings?status=planned&meeting_type=full_board&year=2026&page=1&page_size=20
func (h *BoardHandler) ListMeetings(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.MeetingFilter{
		Status:      q.Get("status"),
		MeetingType: q.Get("meeting_type"),
		Page:        1,
		PageSize:    20,
	}

	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}
	if yr, err := strconv.Atoi(q.Get("year")); err == nil && yr > 0 {
		filter.Year = yr
	}

	meetings, err := h.svc.ListMeetings(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list meetings")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: meetings})
}

// CreateMeeting schedules a new board meeting.
// POST /meetings
func (h *BoardHandler) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" || req.Date == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title and date are required")
		return
	}

	meeting, err := h.svc.CreateMeeting(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create meeting: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: meeting})
}

// GetMeeting returns a single board meeting.
// GET /meetings/{id}
func (h *BoardHandler) GetMeeting(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid meeting ID")
		return
	}

	meeting, err := h.svc.GetMeeting(r.Context(), orgID, meetingID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Meeting not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: meeting})
}

// UpdateMeeting updates a board meeting.
// PUT /meetings/{id}
func (h *BoardHandler) UpdateMeeting(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid meeting ID")
		return
	}

	var req service.UpdateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.svc.UpdateMeeting(r.Context(), orgID, meetingID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update meeting: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Meeting updated successfully"},
	})
}

// GenerateBoardPack generates a board pack for a meeting.
// POST /meetings/{id}/generate-pack
func (h *BoardHandler) GenerateBoardPack(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid meeting ID")
		return
	}

	report, err := h.svc.GenerateBoardPack(r.Context(), orgID, meetingID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate board pack: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: report})
}

// DownloadBoardPack returns the file path for downloading the board pack.
// GET /meetings/{id}/download-pack
func (h *BoardHandler) DownloadBoardPack(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid meeting ID")
		return
	}

	path, err := h.svc.GetBoardPackPath(r.Context(), orgID, meetingID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Board pack not found: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]string{
			"file_path":  path,
			"meeting_id": meetingID.String(),
		},
	})
}

// ============================================================
// DECISIONS
// ============================================================

// RecordDecision records a board decision for a meeting.
// POST /decisions
func (h *BoardHandler) RecordDecision(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var body struct {
		MeetingID uuid.UUID                   `json:"meeting_id"`
		service.RecordDecisionRequest
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title is required")
		return
	}
	if body.MeetingID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "meeting_id is required")
		return
	}

	decision, err := h.svc.RecordDecision(r.Context(), orgID, body.MeetingID, body.RecordDecisionRequest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record decision: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: decision})
}

// ListDecisions returns a filtered list of board decisions.
// GET /decisions?meeting_id=...&decision_type=...&action_status=...&page=1&page_size=20
func (h *BoardHandler) ListDecisions(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	q := r.URL.Query()

	filter := service.DecisionFilter{
		DecisionType: q.Get("decision_type"),
		ActionStatus: q.Get("action_status"),
		Page:         1,
		PageSize:     20,
	}

	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		filter.Page = p
	}
	if ps, err := strconv.Atoi(q.Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		filter.PageSize = ps
	}
	if mid := q.Get("meeting_id"); mid != "" {
		if parsed, err := uuid.Parse(mid); err == nil {
			filter.MeetingID = &parsed
		}
	}

	decisions, err := h.svc.ListDecisions(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list decisions")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: decisions})
}

// UpdateDecisionAction updates the action status of a board decision.
// PUT /decisions/{id}/action
func (h *BoardHandler) UpdateDecisionAction(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	decisionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid decision ID")
		return
	}

	var req service.BoardUpdateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.ActionStatus == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "action_status is required")
		return
	}

	if err := h.svc.UpdateDecisionAction(r.Context(), orgID, decisionID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update decision action: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Decision action updated successfully"},
	})
}

// ============================================================
// REPORTS
// ============================================================

// ListReports returns all board reports.
// GET /reports
func (h *BoardHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	reports, err := h.svc.ListReports(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list reports")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: reports})
}

// GenerateReport generates a new board report.
// POST /reports/generate
func (h *BoardHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.BoardGenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	report, err := h.svc.GenerateReport(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate report: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: report})
}

// ============================================================
// DASHBOARD & NIS2
// ============================================================

// GetBoardDashboard returns the aggregated governance dashboard.
// GET /dashboard
func (h *BoardHandler) GetBoardDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dashboard, err := h.svc.GetBoardDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate board dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}

// GetNIS2GovernanceReport generates and returns the NIS2 governance report.
// GET /nis2-governance
func (h *BoardHandler) GetNIS2GovernanceReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	report, err := h.svc.GenerateNIS2GovernanceReport(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate NIS2 governance report: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: report})
}
