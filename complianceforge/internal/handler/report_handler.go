package handler

import (
	"net/http"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

type ReportHandler struct {
	reportingSvc *service.ReportingService
}

func NewReportHandler(reportingSvc *service.ReportingService) *ReportHandler {
	return &ReportHandler{reportingSvc: reportingSvc}
}

// ComplianceReport generates a comprehensive compliance status report.
// GET /api/v1/reports/compliance
func (h *ReportHandler) ComplianceReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	report, err := h.reportingSvc.GenerateComplianceReport(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate compliance report")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: report})
}

// RiskReport generates a comprehensive risk register report.
// GET /api/v1/reports/risk
func (h *ReportHandler) RiskReport(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	report, err := h.reportingSvc.GenerateRiskReport(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate risk report")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: report})
}
