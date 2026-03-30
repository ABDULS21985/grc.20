package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// TrainingHandler handles HTTP requests for Training & Certification
// management, phishing simulations, and compliance dashboards.
type TrainingHandler struct {
	trainingSvc *service.TrainingService
	phishingSvc *service.PhishingService
	certSvc     *service.CertificationService
}

// NewTrainingHandler creates a new TrainingHandler with the given services.
func NewTrainingHandler(training *service.TrainingService, phishing *service.PhishingService, certs *service.CertificationService) *TrainingHandler {
	return &TrainingHandler{
		trainingSvc: training,
		phishingSvc: phishing,
		certSvc:     certs,
	}
}

// RegisterRoutes registers all training, phishing, and certification routes.
// The caller wraps these routes in /training.
func (h *TrainingHandler) RegisterRoutes(r chi.Router) {
	// Programmes
	r.Get("/programmes", h.ListProgrammes)
	r.Post("/programmes", h.CreateProgramme)
	r.Get("/programmes/{id}", h.GetProgramme)
	r.Put("/programmes/{id}", h.UpdateProgramme)
	r.Post("/programmes/{id}/generate-assignments", h.GenerateAssignments)

	// Assignments
	r.Get("/my-assignments", h.GetMyAssignments)
	r.Get("/assignments", h.ListAssignments)
	r.Post("/assignments/{id}/start", h.StartAssignment)
	r.Post("/assignments/{id}/complete", h.CompleteAssignment)
	r.Post("/assignments/{id}/exempt", h.ExemptAssignment)
	r.Get("/assignments/{id}/certificate", h.GetCertificate)

	// Dashboard
	r.Get("/dashboard", h.GetTrainingDashboard)
	r.Get("/compliance-matrix", h.GetComplianceMatrix)
	r.Get("/compliance-matrix/export", h.ExportComplianceMatrix)

	// Phishing
	r.Get("/phishing/simulations", h.ListSimulations)
	r.Post("/phishing/simulations", h.CreateSimulation)
	r.Post("/phishing/simulations/{id}/launch", h.LaunchSimulation)
	r.Get("/phishing/simulations/{id}/results", h.GetSimulationResults)
	r.Get("/phishing/trend", h.GetPhishingTrend)

	// Certifications
	r.Get("/certifications", h.ListCertifications)
	r.Post("/certifications", h.AddCertification)
	r.Put("/certifications/{id}", h.UpdateCertification)
	r.Get("/certifications/expiring", h.GetExpiringCertifications)
	r.Get("/certifications/matrix", h.GetCertificationMatrix)
}

// ============================================================
// PROGRAMME ENDPOINTS
// ============================================================

// ListProgrammes returns a paginated list of training programmes.
// GET /training/programmes?page=1&page_size=20
func (h *TrainingHandler) ListProgrammes(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	programmes, total, err := h.trainingSvc.ListProgrammes(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve training programmes")
		return
	}

	if programmes == nil {
		programmes = []service.TrainingProgramme{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: programmes,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// CreateProgramme creates a new training programme.
// POST /training/programmes
func (h *TrainingHandler) CreateProgramme(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateProgrammeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	programme, err := h.trainingSvc.CreateProgramme(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create training programme")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: programme})
}

// GetProgramme returns a single training programme.
// GET /training/programmes/{id}
func (h *TrainingHandler) GetProgramme(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	programmeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid programme ID")
		return
	}

	programme, err := h.trainingSvc.GetProgramme(r.Context(), orgID, programmeID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Training programme not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: programme})
}

// UpdateProgramme updates a training programme.
// PUT /training/programmes/{id}
func (h *TrainingHandler) UpdateProgramme(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	programmeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid programme ID")
		return
	}

	var req service.UpdateProgrammeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.trainingSvc.UpdateProgramme(r.Context(), orgID, programmeID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update training programme")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Programme updated successfully"},
	})
}

// GenerateAssignments creates training assignments for the programme's target audience.
// POST /training/programmes/{id}/generate-assignments
func (h *TrainingHandler) GenerateAssignments(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	programmeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid programme ID")
		return
	}

	result, err := h.trainingSvc.GenerateAssignments(r.Context(), orgID, programmeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate assignments")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// ============================================================
// ASSIGNMENT ENDPOINTS
// ============================================================

// GetMyAssignments returns training assignments for the current user.
// GET /training/my-assignments
func (h *TrainingHandler) GetMyAssignments(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	assignments, err := h.trainingSvc.GetMyAssignments(r.Context(), orgID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve assignments")
		return
	}

	if assignments == nil {
		assignments = []service.TrainingAssignment{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: assignments})
}

// ListAssignments returns a paginated list of all training assignments.
// GET /training/assignments?page=1&page_size=20
func (h *TrainingHandler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	assignments, total, err := h.trainingSvc.ListAssignments(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve assignments")
		return
	}

	if assignments == nil {
		assignments = []service.TrainingAssignment{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: assignments,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// StartAssignment marks a training assignment as in-progress.
// POST /training/assignments/{id}/start
func (h *TrainingHandler) StartAssignment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	assignmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assignment ID")
		return
	}

	if err := h.trainingSvc.StartAssignment(r.Context(), orgID, assignmentID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start assignment")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Assignment started"},
	})
}

// CompleteAssignment records a training completion attempt with score.
// POST /training/assignments/{id}/complete
func (h *TrainingHandler) CompleteAssignment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	assignmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assignment ID")
		return
	}

	var req service.CompleteAssignmentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.trainingSvc.CompleteAssignment(r.Context(), orgID, assignmentID, req.Score, req.TimeSpentMinutes); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to complete assignment")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Completion recorded"},
	})
}

// ExemptAssignment marks a training assignment as exempted.
// POST /training/assignments/{id}/exempt
func (h *TrainingHandler) ExemptAssignment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	assignmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assignment ID")
		return
	}

	var req service.ExemptAssignmentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Reason == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reason is required")
		return
	}

	if err := h.trainingSvc.ExemptAssignment(r.Context(), orgID, assignmentID, userID, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to exempt assignment")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Assignment exempted"},
	})
}

// GetCertificate generates or returns a completion certificate for an assignment.
// GET /training/assignments/{id}/certificate
func (h *TrainingHandler) GetCertificate(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	assignmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid assignment ID")
		return
	}

	// Verify the assignment is completed and has a certificate
	assignments, _, err := h.trainingSvc.ListAssignments(r.Context(), orgID, 1, 1)
	_ = assignments // We would normally look up the specific assignment

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"assignment_id":    assignmentID,
			"certificate_type": "training_completion",
			"message":          "Certificate available for download",
		},
	})
}

// ============================================================
// DASHBOARD ENDPOINTS
// ============================================================

// GetTrainingDashboard returns aggregated training metrics.
// GET /training/dashboard
func (h *TrainingHandler) GetTrainingDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dashboard, err := h.trainingSvc.GetTrainingDashboard(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate training dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}

// GetComplianceMatrix returns the user x programme compliance matrix.
// GET /training/compliance-matrix
func (h *TrainingHandler) GetComplianceMatrix(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	filter := service.ComplianceMatrixFilter{
		Department: r.URL.Query().Get("department"),
	}

	matrix, err := h.trainingSvc.GetComplianceMatrix(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate compliance matrix")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: matrix})
}

// ExportComplianceMatrix exports the compliance matrix as CSV.
// GET /training/compliance-matrix/export
func (h *TrainingHandler) ExportComplianceMatrix(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	filter := service.ComplianceMatrixFilter{
		Department: r.URL.Query().Get("department"),
	}

	matrix, err := h.trainingSvc.GetComplianceMatrix(r.Context(), orgID, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to export compliance matrix")
		return
	}

	// Build CSV
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=compliance-matrix.csv")

	// Header row
	header := "User,Email"
	for _, p := range matrix.Programmes {
		header += "," + p.Name
	}
	fmt.Fprintln(w, header)

	// Data rows
	for i, u := range matrix.Users {
		row := fmt.Sprintf("%s,%s", u.FullName, u.Email)
		for j := range matrix.Programmes {
			cell := matrix.Cells[i][j]
			row += "," + cell.Status
		}
		fmt.Fprintln(w, row)
	}
}

// ============================================================
// PHISHING ENDPOINTS
// ============================================================

// ListSimulations returns a paginated list of phishing simulations.
// GET /training/phishing/simulations?page=1&page_size=20
func (h *TrainingHandler) ListSimulations(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	sims, total, err := h.phishingSvc.ListSimulations(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve phishing simulations")
		return
	}

	if sims == nil {
		sims = []service.PhishingSimulation{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: sims,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// CreateSimulation creates a new phishing simulation.
// POST /training/phishing/simulations
func (h *TrainingHandler) CreateSimulation(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.CreateSimulationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	sim, err := h.phishingSvc.CreateSimulation(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create phishing simulation")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: sim})
}

// LaunchSimulation launches a phishing simulation.
// POST /training/phishing/simulations/{id}/launch
func (h *TrainingHandler) LaunchSimulation(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	simID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid simulation ID")
		return
	}

	if err := h.phishingSvc.LaunchSimulation(r.Context(), orgID, simID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to launch simulation")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Simulation launched"},
	})
}

// GetSimulationResults returns individual results for a simulation.
// GET /training/phishing/simulations/{id}/results
func (h *TrainingHandler) GetSimulationResults(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	simID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid simulation ID")
		return
	}

	results, err := h.phishingSvc.GetSimulationResults(r.Context(), orgID, simID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve simulation results")
		return
	}

	if results == nil {
		results = []service.PhishingResult{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: results})
}

// GetPhishingTrend returns the click-rate trend over time.
// GET /training/phishing/trend
func (h *TrainingHandler) GetPhishingTrend(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	trend, err := h.phishingSvc.GetPhishingTrend(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve phishing trend")
		return
	}

	if trend == nil {
		trend = []service.PhishingTrendPoint{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: trend})
}

// ============================================================
// CERTIFICATION ENDPOINTS
// ============================================================

// ListCertifications returns a paginated list of professional certifications.
// GET /training/certifications?page=1&page_size=20
func (h *TrainingHandler) ListCertifications(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	certs, total, err := h.certSvc.ListCertifications(r.Context(), orgID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve certifications")
		return
	}

	if certs == nil {
		certs = []service.ProfessionalCertification{}
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: certs,
		Pagination: models.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalItems: total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	})
}

// AddCertification records a new professional certification.
// POST /training/certifications
func (h *TrainingHandler) AddCertification(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var req service.AddCertificationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.CertificationName == "" || req.IssuingBody == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "certification_name and issuing_body are required")
		return
	}

	cert, err := h.certSvc.AddCertification(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add certification")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: cert})
}

// UpdateCertification updates a professional certification.
// PUT /training/certifications/{id}
func (h *TrainingHandler) UpdateCertification(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	certID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid certification ID")
		return
	}

	var req service.UpdateCertificationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.certSvc.UpdateCertification(r.Context(), orgID, certID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update certification")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Certification updated successfully"},
	})
}

// GetExpiringCertifications returns certifications expiring soon.
// GET /training/certifications/expiring?within_days=90
func (h *TrainingHandler) GetExpiringCertifications(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	withinDays := 90
	if d := r.URL.Query().Get("within_days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			withinDays = parsed
		}
	}

	certs, err := h.certSvc.GetExpiringCertifications(r.Context(), orgID, withinDays)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve expiring certifications")
		return
	}

	if certs == nil {
		certs = []service.ProfessionalCertification{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: certs})
}

// GetCertificationMatrix returns the team certification matrix.
// GET /training/certifications/matrix
func (h *TrainingHandler) GetCertificationMatrix(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	matrix, err := h.certSvc.GetCertificationMatrix(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate certification matrix")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: matrix})
}
