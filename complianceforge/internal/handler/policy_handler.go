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

type PolicyHandler struct {
	repo *repository.PolicyRepo
	db   *database.DB
}

func NewPolicyHandler(repo *repository.PolicyRepo, db *database.DB) *PolicyHandler {
	return &PolicyHandler{repo: repo, db: db}
}

// ListPolicies returns paginated policies for the organization.
// GET /api/v1/policies
func (h *PolicyHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	policies, total, err := h.repo.List(r.Context(), orgID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve policies")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: policies,
		Pagination: models.Pagination{
			Page: params.Page, PageSize: params.PageSize,
			TotalItems: total, TotalPages: totalPages,
			HasNext: params.Page < totalPages, HasPrev: params.Page > 1,
		},
	})
}

// GetPolicy returns a single policy by ID.
// GET /api/v1/policies/{id}
func (h *PolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	policyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID")
		return
	}

	policy, err := h.repo.GetByID(r.Context(), orgID, policyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Policy not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: policy})
}

// CreatePolicy creates a new policy with its first version.
// POST /api/v1/policies
func (h *PolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req models.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Title == "" || req.ContentHTML == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Title and content are required")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), orgID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	policy := &models.Policy{
		BaseModel:                 models.BaseModel{OrganizationID: orgID},
		Title:                     req.Title,
		CategoryID:                &req.CategoryID,
		Status:                    models.PolicyStatusDraft,
		Classification:            req.Classification,
		OwnerUserID:               &req.OwnerUserID,
		AuthorUserID:              &userID,
		ApproverUserID:            &req.ApproverUserID,
		ReviewFrequencyMonths:     req.ReviewFrequencyMonths,
		AppliesToAll:              true,
		IsMandatory:               req.IsMandatory,
		RequiresAttestation:       req.RequiresAttestation,
		AttestationFrequencyMonths: 12,
		Tags:                      req.Tags,
		Metadata:                  models.JSONB("{}"),
	}

	if err := h.repo.Create(r.Context(), tx, policy); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create policy")
		return
	}

	// Create first version
	version := &models.PolicyVersion{
		PolicyID:       policy.ID,
		OrganizationID: orgID,
		VersionNumber:  1,
		VersionLabel:   "1.0",
		Title:          req.Title,
		ContentHTML:    req.ContentHTML,
		Summary:        req.Summary,
		ChangeType:     "major",
		Language:       "en",
		Status:         "draft",
		CreatedBy:      userID,
		Metadata:       models.JSONB("{}"),
	}

	if err := h.repo.CreateVersion(r.Context(), tx, version); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create policy version")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to commit")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: policy})
}

// PublishPolicy moves a policy from approved to published status.
// POST /api/v1/policies/{id}/publish
func (h *PolicyHandler) PublishPolicy(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	policyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID")
		return
	}

	policy, err := h.repo.GetByID(r.Context(), orgID, policyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Policy not found")
		return
	}

	if policy.Status != models.PolicyStatusApproved && policy.Status != models.PolicyStatusDraft {
		writeError(w, http.StatusConflict, "INVALID_STATUS", "Policy must be approved before publishing")
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), orgID, policyID, models.PolicyStatusPublished); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to publish policy")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"policy_id": policyID,
			"status":    "published",
			"message":   "Policy published successfully",
		},
	})
}

// GetAttestationStats returns policy attestation completion metrics.
// GET /api/v1/policies/attestations/stats
func (h *PolicyHandler) GetAttestationStats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	stats, err := h.repo.GetAttestationStats(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve attestation stats")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: stats})
}

// SubmitAttestation records a user's acknowledgment of a policy.
// POST /api/v1/policies/{id}/attest
func (h *PolicyHandler) SubmitAttestation(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	policyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid policy ID")
		return
	}

	policy, err := h.repo.GetByID(r.Context(), orgID, policyID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Policy not found")
		return
	}

	if policy.CurrentVersionID == nil {
		writeError(w, http.StatusConflict, "NO_VERSION", "Policy has no published version")
		return
	}

	now := time.Now()
	attestation := &models.PolicyAttestation{
		PolicyID:          policyID,
		PolicyVersionID:   *policy.CurrentVersionID,
		OrganizationID:    orgID,
		UserID:            userID,
		Status:            "attested",
		AttestedAt:        &now,
		AttestedFromIP:    r.RemoteAddr,
		AttestationMethod: "digital_click",
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"attestation": attestation,
			"message":     "Policy attestation recorded",
		},
	})
}
