package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// DashboardV2 handles executive dashboard, gap analysis, and cross-framework mapping.
type DashboardV2 struct {
	engine *service.ComplianceEngine
}

func NewDashboardV2(engine *service.ComplianceEngine) *DashboardV2 {
	return &DashboardV2{engine: engine}
}

// Summary returns the executive dashboard with aggregated metrics across all modules.
// GET /api/v1/dashboard/summary
func (h *DashboardV2) Summary(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	summary, err := h.engine.DashboardSummary(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: summary})
}

// GapAnalysis returns all controls not fully implemented across adopted frameworks.
// GET /api/v1/compliance/gaps?framework_id=...
func (h *DashboardV2) GapAnalysis(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var frameworkID *uuid.UUID
	if fid := r.URL.Query().Get("framework_id"); fid != "" {
		if parsed, err := uuid.Parse(fid); err == nil {
			frameworkID = &parsed
		}
	}

	gaps, err := h.engine.GapAnalysis(r.Context(), orgID, frameworkID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to run gap analysis")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"total_gaps": len(gaps),
			"gaps":       gaps,
		},
	})
}

// CrossFrameworkCoverage shows how implementing one framework covers another.
// GET /api/v1/compliance/cross-mapping
func (h *DashboardV2) CrossFrameworkCoverage(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	coverage, err := h.engine.CrossFrameworkCoverage(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate coverage map")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"total_mappings": len(coverage),
			"mappings":       coverage,
		},
	})
}
