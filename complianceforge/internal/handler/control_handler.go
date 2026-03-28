package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
)

type ControlHandler struct {
	db *database.DB
}

func NewControlHandler(db *database.DB) *ControlHandler {
	return &ControlHandler{db: db}
}

// ListControlImplementations returns control implementations for an adopted framework.
// GET /api/v1/frameworks/{fwId}/implementations?page=1&status=not_implemented
func (h *ControlHandler) ListImplementations(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	fwID, err := uuid.Parse(chi.URLParam(r, "fwId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid framework ID")
		return
	}

	params := parsePagination(r)
	statusFilter := r.URL.Query().Get("status")

	query := `
		SELECT ci.id, ci.organization_id, ci.framework_control_id, ci.org_framework_id,
			ci.status, ci.implementation_status, ci.maturity_level,
			ci.owner_user_id, ci.implementation_description, ci.gap_description,
			ci.remediation_plan, ci.remediation_due_date, ci.test_frequency,
			ci.last_tested_at, ci.last_test_result, ci.effectiveness_score,
			ci.risk_if_not_implemented, ci.automation_level, ci.tags,
			fc.code AS control_code, fc.title AS control_title,
			fc.control_type, fc.implementation_type,
			COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned') AS owner_name
		FROM control_implementations ci
		JOIN framework_controls fc ON ci.framework_control_id = fc.id
		JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
		LEFT JOIN users u ON ci.owner_user_id = u.id
		WHERE ci.organization_id = $1 AND of2.framework_id = $2 AND ci.deleted_at IS NULL`

	args := []interface{}{orgID, fwID}
	argIdx := 3

	if statusFilter != "" {
		query += " AND ci.status = $" + string(rune('0'+argIdx))
		args = append(args, statusFilter)
		argIdx++
	}

	query += " ORDER BY fc.sort_order LIMIT $" + string(rune('0'+argIdx)) + " OFFSET $" + string(rune('0'+argIdx+1))
	args = append(args, params.PageSize, params.Offset())

	rows, err := h.db.Pool.Query(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list implementations")
		return
	}
	defer rows.Close()

	type ImplRow struct {
		models.ControlImplementation
		ControlCode    string `json:"control_code"`
		ControlTitle   string `json:"control_title"`
		ControlType    string `json:"control_type"`
		ImplType       string `json:"implementation_type"`
		OwnerName      string `json:"owner_name"`
	}

	var results []ImplRow
	for rows.Next() {
		var row ImplRow
		if err := rows.Scan(
			&row.ID, &row.OrganizationID, &row.FrameworkControlID, &row.OrgFrameworkID,
			&row.Status, &row.ImplementationStatus, &row.MaturityLevel,
			&row.OwnerUserID, &row.ImplementationDescription, &row.GapDescription,
			&row.RemediationPlan, &row.RemediationDueDate, &row.TestFrequency,
			&row.LastTestedAt, &row.LastTestResult, &row.EffectivenessScore,
			&row.RiskIfNotImplemented, &row.AutomationLevel, &row.Tags,
			&row.ControlCode, &row.ControlTitle, &row.ControlType, &row.ImplType,
			&row.OwnerName,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan row")
			return
		}
		results = append(results, row)
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: results})
}

// UpdateImplementation updates a control's implementation status, maturity, and details.
// PUT /api/v1/controls/{id}
func (h *ControlHandler) UpdateImplementation(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	implID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid implementation ID")
		return
	}

	var req models.UpdateControlImplementationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Build dynamic UPDATE
	sets := "updated_at = NOW()"
	args := []interface{}{}
	argIdx := 1

	if req.Status != nil {
		sets += ", status = $" + itoa(argIdx)
		args = append(args, string(*req.Status))
		argIdx++
	}
	if req.MaturityLevel != nil {
		sets += ", maturity_level = $" + itoa(argIdx)
		args = append(args, *req.MaturityLevel)
		argIdx++
	}
	if req.OwnerUserID != nil {
		sets += ", owner_user_id = $" + itoa(argIdx)
		args = append(args, *req.OwnerUserID)
		argIdx++
	}
	if req.ImplementationDescription != nil {
		sets += ", implementation_description = $" + itoa(argIdx)
		args = append(args, *req.ImplementationDescription)
		argIdx++
	}
	if req.GapDescription != nil {
		sets += ", gap_description = $" + itoa(argIdx)
		args = append(args, *req.GapDescription)
		argIdx++
	}
	if req.RemediationPlan != nil {
		sets += ", remediation_plan = $" + itoa(argIdx)
		args = append(args, *req.RemediationPlan)
		argIdx++
	}
	if req.CompensatingControlDescription != nil {
		sets += ", compensating_control_description = $" + itoa(argIdx)
		args = append(args, *req.CompensatingControlDescription)
		argIdx++
	}
	if req.AutomationLevel != nil {
		sets += ", automation_level = $" + itoa(argIdx)
		args = append(args, *req.AutomationLevel)
		argIdx++
	}

	query := "UPDATE control_implementations SET " + sets +
		" WHERE id = $" + itoa(argIdx) + " AND organization_id = $" + itoa(argIdx+1) +
		" AND deleted_at IS NULL"
	args = append(args, implID, orgID)

	result, err := h.db.Pool.Exec(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update implementation")
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Control implementation not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":      implID,
			"message": "Control implementation updated",
		},
	})
}

// GetImplementation returns a single control implementation with evidence.
// GET /api/v1/controls/{id}
func (h *ControlHandler) GetImplementation(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	implID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid implementation ID")
		return
	}

	query := `
		SELECT ci.id, ci.organization_id, ci.framework_control_id, ci.org_framework_id,
			ci.status, ci.implementation_status, ci.maturity_level,
			ci.owner_user_id, ci.reviewer_user_id,
			ci.implementation_description, ci.implementation_notes,
			ci.compensating_control_description, ci.gap_description,
			ci.remediation_plan, ci.remediation_due_date,
			ci.test_frequency, ci.last_tested_at, ci.last_tested_by,
			ci.last_test_result, ci.effectiveness_score,
			ci.risk_if_not_implemented, ci.automation_level,
			ci.tags, ci.created_at, ci.updated_at,
			fc.code, fc.title, fc.description, fc.guidance,
			fc.control_type, fc.implementation_type, fc.is_mandatory
		FROM control_implementations ci
		JOIN framework_controls fc ON ci.framework_control_id = fc.id
		WHERE ci.id = $1 AND ci.organization_id = $2 AND ci.deleted_at IS NULL`

	var impl struct {
		models.ControlImplementation
		ControlCode    string `json:"control_code"`
		ControlTitle   string `json:"control_title"`
		ControlDesc    string `json:"control_description"`
		ControlGuidance string `json:"control_guidance"`
		CtrlType       string `json:"control_type"`
		ImplType       string `json:"implementation_type"`
		IsMandatory    bool   `json:"is_mandatory"`
	}

	err = h.db.Pool.QueryRow(r.Context(), query, implID, orgID).Scan(
		&impl.ID, &impl.OrganizationID, &impl.FrameworkControlID, &impl.OrgFrameworkID,
		&impl.Status, &impl.ImplementationStatus, &impl.MaturityLevel,
		&impl.OwnerUserID, &impl.ReviewerUserID,
		&impl.ImplementationDescription, &impl.ImplementationNotes,
		&impl.CompensatingControlDescription, &impl.GapDescription,
		&impl.RemediationPlan, &impl.RemediationDueDate,
		&impl.TestFrequency, &impl.LastTestedAt, &impl.LastTestedBy,
		&impl.LastTestResult, &impl.EffectivenessScore,
		&impl.RiskIfNotImplemented, &impl.AutomationLevel,
		&impl.Tags, &impl.CreatedAt, &impl.UpdatedAt,
		&impl.ControlCode, &impl.ControlTitle, &impl.ControlDesc, &impl.ControlGuidance,
		&impl.CtrlType, &impl.ImplType, &impl.IsMandatory,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Control implementation not found")
		return
	}

	// Load evidence
	evidenceRows, _ := h.db.Pool.Query(r.Context(), `
		SELECT id, title, evidence_type, file_name, file_size_bytes, collection_method,
			collected_at, is_current, review_status
		FROM control_evidence
		WHERE control_implementation_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY collected_at DESC`, implID, orgID)

	var evidence []models.ControlEvidence
	if evidenceRows != nil {
		defer evidenceRows.Close()
		for evidenceRows.Next() {
			var e models.ControlEvidence
			evidenceRows.Scan(&e.ID, &e.Title, &e.EvidenceType, &e.FileName,
				&e.FileSizeBytes, &e.CollectionMethod, &e.CollectedAt,
				&e.IsCurrent, &e.ReviewStatus)
			evidence = append(evidence, e)
		}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"implementation": impl,
			"evidence":       evidence,
			"evidence_count": len(evidence),
		},
	})
}

// RecordTestResult records a control test result.
// POST /api/v1/controls/{id}/test
func (h *ControlHandler) RecordTestResult(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	implID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid implementation ID")
		return
	}

	var req struct {
		TestType      string `json:"test_type"`
		TestProcedure string `json:"test_procedure"`
		Result        string `json:"result"` // pass, fail, partial, inconclusive
		Findings      string `json:"findings"`
		Recommendations string `json:"recommendations"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	// Insert test result
	var testID uuid.UUID
	err = h.db.Pool.QueryRow(r.Context(), `
		INSERT INTO control_test_results (
			organization_id, control_implementation_id, test_type,
			test_procedure, result, findings, recommendations,
			tested_by, tested_at, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),'{}')
		RETURNING id`,
		orgID, implID, req.TestType, req.TestProcedure,
		req.Result, req.Findings, req.Recommendations, userID,
	).Scan(&testID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record test result")
		return
	}

	// Update the control implementation's last test info
	h.db.Pool.Exec(r.Context(), `
		UPDATE control_implementations SET
			last_tested_at = NOW(),
			last_tested_by = $1,
			last_test_result = $2
		WHERE id = $3 AND organization_id = $4`,
		userID, req.Result, implID, orgID)

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"test_result_id": testID,
			"result":         req.Result,
			"message":        "Test result recorded",
		},
	})
}

func itoa(n int) string {
	return string(rune('0' + n))
}
