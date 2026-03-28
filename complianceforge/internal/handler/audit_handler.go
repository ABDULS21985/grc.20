package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/repository"
)

type AuditHandler struct {
	repo *repository.AuditRepo
	db   *database.DB
}

func NewAuditHandler(repo *repository.AuditRepo, db *database.DB) *AuditHandler {
	return &AuditHandler{repo: repo, db: db}
}

// ListAudits returns paginated audits.
// GET /api/v1/audits
func (h *AuditHandler) ListAudits(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	audits, total, err := h.repo.List(r.Context(), orgID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve audits")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: audits,
		Pagination: models.Pagination{
			Page: params.Page, PageSize: params.PageSize,
			TotalItems: total, TotalPages: totalPages,
			HasNext: params.Page < totalPages, HasPrev: params.Page > 1,
		},
	})
}

// GetAudit returns a single audit with findings.
// GET /api/v1/audits/{id}
func (h *AuditHandler) GetAudit(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	auditID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid audit ID")
		return
	}

	audit, err := h.repo.GetByID(r.Context(), orgID, auditID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Audit not found")
		return
	}

	findings, _ := h.repo.ListFindings(r.Context(), orgID, auditID)
	audit.Findings = findings

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: audit})
}

// CreateAudit plans a new audit engagement.
// POST /api/v1/audits
func (h *AuditHandler) CreateAudit(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req struct {
		Title            string      `json:"title"`
		AuditType        string      `json:"audit_type"`
		Description      string      `json:"description"`
		Scope            string      `json:"scope"`
		Methodology      string      `json:"methodology"`
		LeadAuditorID    *uuid.UUID  `json:"lead_auditor_id"`
		FrameworkIDs     []uuid.UUID `json:"framework_ids"`
		PlannedStartDate string      `json:"planned_start_date"`
		PlannedEndDate   string      `json:"planned_end_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" || req.AuditType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Title and audit_type are required")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), orgID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Transaction failed")
		return
	}
	defer tx.Rollback(r.Context())

	audit := &models.Audit{
		BaseModel:          models.BaseModel{OrganizationID: orgID},
		Title:              req.Title,
		AuditType:          req.AuditType,
		Status:             models.AuditStatusPlanned,
		Description:        req.Description,
		Scope:              req.Scope,
		Methodology:        req.Methodology,
		LeadAuditorID:      req.LeadAuditorID,
		LinkedFrameworkIDs: req.FrameworkIDs,
		Metadata:           models.JSONB("{}"),
	}

	if err := h.repo.Create(r.Context(), tx, audit); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create audit")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Commit failed")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: audit})
}

// CreateFinding adds a finding to an audit.
// POST /api/v1/audits/{id}/findings
func (h *AuditHandler) CreateFinding(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	auditID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid audit ID")
		return
	}

	var req struct {
		FindingRef        string     `json:"finding_ref"`
		Title             string     `json:"title"`
		Description       string     `json:"description"`
		Severity          string     `json:"severity"`
		FindingType       string     `json:"finding_type"`
		ControlID         *uuid.UUID `json:"control_id"`
		RootCause         string     `json:"root_cause"`
		Recommendation    string     `json:"recommendation"`
		ResponsibleUserID *uuid.UUID `json:"responsible_user_id"`
		DueDate           string     `json:"due_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), orgID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Transaction failed")
		return
	}
	defer tx.Rollback(r.Context())

	finding := &models.AuditFinding{
		BaseModel:         models.BaseModel{OrganizationID: orgID},
		AuditID:           auditID,
		FindingRef:        req.FindingRef,
		Title:             req.Title,
		Description:       req.Description,
		Severity:          models.FindingSeverity(req.Severity),
		Status:            "open",
		FindingType:       req.FindingType,
		ControlID:         req.ControlID,
		RootCause:         req.RootCause,
		Recommendation:    req.Recommendation,
		ResponsibleUserID: req.ResponsibleUserID,
		Metadata:          models.JSONB("{}"),
	}

	if err := h.repo.CreateFinding(r.Context(), tx, finding); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create finding")
		return
	}

	// Update audit finding counts
	tx.Exec(r.Context(), `
		UPDATE audits SET
			total_findings = total_findings + 1,
			critical_findings = critical_findings + CASE WHEN $1 = 'critical' THEN 1 ELSE 0 END,
			high_findings = high_findings + CASE WHEN $1 = 'high' THEN 1 ELSE 0 END,
			medium_findings = medium_findings + CASE WHEN $1 = 'medium' THEN 1 ELSE 0 END,
			low_findings = low_findings + CASE WHEN $1 = 'low' THEN 1 ELSE 0 END
		WHERE id = $2 AND organization_id = $3`,
		string(finding.Severity), auditID, orgID)

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Commit failed")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: finding})
}

// ListFindings returns all findings for an audit.
// GET /api/v1/audits/{id}/findings
func (h *AuditHandler) ListFindings(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	auditID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid audit ID")
		return
	}

	findings, err := h.repo.ListFindings(r.Context(), orgID, auditID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve findings")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: findings})
}

// GetFindingsStats returns finding metrics for the dashboard.
// GET /api/v1/audits/findings/stats
func (h *AuditHandler) GetFindingsStats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	stats, err := h.repo.GetFindingsStats(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get stats")
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: stats})
}
