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

// ============================================================
// ANALYTICS HANDLER — REST endpoints for Advanced Analytics,
//                      Predictive Risk Scoring & BI Dashboard
// ============================================================

// AnalyticsHandler handles HTTP requests for the analytics module.
type AnalyticsHandler struct {
	engine  *service.AnalyticsEngine
	query   *service.AnalyticsQuery
	dashSvc *service.DashboardService
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(engine *service.AnalyticsEngine, query *service.AnalyticsQuery, dashSvc *service.DashboardService) *AnalyticsHandler {
	return &AnalyticsHandler{
		engine:  engine,
		query:   query,
		dashSvc: dashSvc,
	}
}

// RegisterRoutes mounts all analytics routes on the given router.
func (h *AnalyticsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/analytics", func(r chi.Router) {
		// Snapshots
		r.Get("/snapshots", h.ListSnapshots)

		// Compliance trends
		r.Get("/trends/compliance", h.ComplianceTrends)

		// Risk trends (from snapshots)
		r.Get("/trends/risks", h.RiskTrends)

		// Predictions
		r.Get("/predictions/risks/{riskId}", h.RiskPrediction)
		r.Get("/predictions/breach-probability", h.BreachProbability)

		// Benchmarks
		r.Get("/benchmarks", h.PeerBenchmarks)

		// Metrics
		r.Get("/metrics/{metric}", h.MetricTimeSeries)
		r.Get("/metrics/{metric}/compare", h.MetricComparison)

		// Top movers
		r.Get("/top-movers", h.TopMovers)

		// Distribution
		r.Get("/distribution/{entity}", h.Distribution)

		// Export
		r.Post("/export", h.ExportData)

		// Dashboards
		r.Get("/dashboards", h.ListDashboards)
		r.Post("/dashboards", h.CreateDashboard)
		r.Put("/dashboards/{id}", h.UpdateDashboard)
		r.Delete("/dashboards/{id}", h.DeleteDashboard)

		// Widget types
		r.Get("/widget-types", h.ListWidgetTypes)
	})
}

// ============================================================
// SNAPSHOT ENDPOINTS
// ============================================================

// ListSnapshots returns paginated analytics snapshots for the organization.
// GET /analytics/snapshots?type=daily&page=1&page_size=20
func (h *AnalyticsHandler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)
	snapshotType := r.URL.Query().Get("type")
	if snapshotType == "" {
		snapshotType = "daily"
	}

	rows, err := h.engine.Pool().Query(r.Context(), `
		SELECT id, organization_id, snapshot_type::TEXT, snapshot_date, metrics, created_at
		FROM analytics_snapshots
		WHERE organization_id = $1 AND snapshot_type = $2
		ORDER BY snapshot_date DESC
		LIMIT $3 OFFSET $4`,
		orgID, snapshotType, params.PageSize, params.Offset(),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve snapshots")
		return
	}
	defer rows.Close()

	type snapshotRow struct {
		ID             uuid.UUID       `json:"id"`
		OrganizationID uuid.UUID       `json:"organization_id"`
		SnapshotType   string          `json:"snapshot_type"`
		SnapshotDate   string          `json:"snapshot_date"`
		Metrics        json.RawMessage `json:"metrics"`
		CreatedAt      string          `json:"created_at"`
	}

	var snapshots []snapshotRow
	for rows.Next() {
		var s snapshotRow
		var snapshotDate, createdAt interface{}
		if err := rows.Scan(&s.ID, &s.OrganizationID, &s.SnapshotType, &snapshotDate, &s.Metrics, &createdAt); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to scan snapshot")
			return
		}
		if t, ok := snapshotDate.(interface{ Format(string) string }); ok {
			s.SnapshotDate = t.Format("2006-01-02")
		}
		if t, ok := createdAt.(interface{ Format(string) string }); ok {
			s.CreatedAt = t.Format("2006-01-02T15:04:05Z")
		}
		snapshots = append(snapshots, s)
	}

	// Get total count
	var total int64
	h.engine.Pool().QueryRow(r.Context(), `
		SELECT COUNT(*) FROM analytics_snapshots
		WHERE organization_id = $1 AND snapshot_type = $2`,
		orgID, snapshotType,
	).Scan(&total)

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: snapshots,
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

// ============================================================
// TREND ENDPOINTS
// ============================================================

// ComplianceTrends returns compliance score trends for the organization.
// GET /analytics/trends/compliance?framework=ISO27001&months=12
func (h *AnalyticsHandler) ComplianceTrends(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	frameworkCode := r.URL.Query().Get("framework")
	months := 12
	if m := r.URL.Query().Get("months"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v > 0 && v <= 60 {
			months = v
		}
	}

	trends, err := h.engine.GetComplianceTrends(r.Context(), orgID, frameworkCode, months)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve compliance trends")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: trends})
}

// RiskTrends returns risk metric trends extracted from snapshots.
// GET /analytics/trends/risks?period=90d&granularity=daily
func (h *AnalyticsHandler) RiskTrends(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "90d"
	}
	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "daily"
	}

	points, err := h.query.GetMetricTimeSeries(r.Context(), orgID, "avg_risk_score", period, granularity)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve risk trends")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: points})
}

// ============================================================
// PREDICTION ENDPOINTS
// ============================================================

// RiskPrediction returns the predicted risk score trajectory for a specific risk.
// GET /analytics/predictions/risks/{riskId}
func (h *AnalyticsHandler) RiskPrediction(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	riskID, err := uuid.Parse(chi.URLParam(r, "riskId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid risk ID")
		return
	}

	prediction, err := h.engine.PredictRiskScoreTrajectory(r.Context(), orgID, riskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PREDICTION_ERROR", "Failed to generate risk prediction")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: prediction})
}

// BreachProbability returns the estimated breach probability for the organization.
// GET /analytics/predictions/breach-probability
func (h *AnalyticsHandler) BreachProbability(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	prediction, err := h.engine.PredictBreachProbability(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PREDICTION_ERROR", "Failed to generate breach probability")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: prediction})
}

// ============================================================
// BENCHMARK ENDPOINTS
// ============================================================

// PeerBenchmarks returns the organization's position relative to anonymized peer benchmarks.
// GET /analytics/benchmarks
func (h *AnalyticsHandler) PeerBenchmarks(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	comparison, err := h.engine.CompareToPeers(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve peer benchmarks")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: comparison})
}

// ============================================================
// METRIC ENDPOINTS
// ============================================================

// MetricTimeSeries returns a time series for the requested metric.
// GET /analytics/metrics/{metric}?period=6m&granularity=weekly
func (h *AnalyticsHandler) MetricTimeSeries(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	metric := chi.URLParam(r, "metric")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "6m"
	}
	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "daily"
	}

	points, err := h.query.GetMetricTimeSeries(r.Context(), orgID, metric, period, granularity)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_METRIC", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: points})
}

// MetricComparison compares a metric across two time periods.
// GET /analytics/metrics/{metric}/compare?current=2024-01&previous=2023-12
func (h *AnalyticsHandler) MetricComparison(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	metric := chi.URLParam(r, "metric")
	currentPeriod := r.URL.Query().Get("current")
	previousPeriod := r.URL.Query().Get("previous")

	if currentPeriod == "" || previousPeriod == "" {
		writeError(w, http.StatusBadRequest, "MISSING_PARAMS", "current and previous period parameters are required")
		return
	}

	comparison, err := h.query.GetMetricComparison(r.Context(), orgID, metric, currentPeriod, previousPeriod)
	if err != nil {
		writeError(w, http.StatusBadRequest, "COMPARISON_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: comparison})
}

// ============================================================
// TOP MOVERS & DISTRIBUTION
// ============================================================

// TopMovers returns entities with the largest metric changes.
// GET /analytics/top-movers?metric=compliance_score&period=30d&direction=improving&limit=10
func (h *AnalyticsHandler) TopMovers(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "compliance_score"
	}
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "improving"
	}
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	movers, err := h.query.GetTopMovers(r.Context(), orgID, metric, period, direction, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve top movers")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: movers})
}

// Distribution returns the distribution of an entity grouped by a category.
// GET /analytics/distribution/{entity}?group_by=severity
func (h *AnalyticsHandler) Distribution(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	entity := chi.URLParam(r, "entity")
	groupBy := r.URL.Query().Get("group_by")
	if groupBy == "" {
		groupBy = "severity"
	}

	entries, err := h.query.GetDistribution(r.Context(), orgID, entity, groupBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, "DISTRIBUTION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: entries})
}

// ============================================================
// EXPORT
// ============================================================

// ExportData exports analytics data in CSV or JSON format.
// POST /analytics/export
func (h *AnalyticsHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	var config service.AnalyticsExportConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if config.Format == "" {
		config.Format = "json"
	}

	data, contentType, err := h.query.ExportAnalyticsData(r.Context(), orgID, config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "EXPORT_ERROR", "Failed to export analytics data")
		return
	}

	w.Header().Set("Content-Type", contentType)
	if config.Format == "csv" {
		w.Header().Set("Content-Disposition", "attachment; filename=analytics_export.csv")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ============================================================
// DASHBOARD ENDPOINTS
// ============================================================

// ListDashboards returns all dashboards for the organization.
// GET /analytics/dashboards
func (h *AnalyticsHandler) ListDashboards(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	dashboards, err := h.dashSvc.ListDashboards(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve dashboards")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboards})
}

// CreateDashboard creates a new custom dashboard.
// POST /analytics/dashboards
func (h *AnalyticsHandler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req service.CreateDashboardReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Dashboard name is required")
		return
	}

	dashboard, err := h.dashSvc.CreateDashboard(r.Context(), orgID, userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create dashboard")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: dashboard})
}

// UpdateDashboard updates an existing dashboard.
// PUT /analytics/dashboards/{id}
func (h *AnalyticsHandler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	dashboardID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid dashboard ID")
		return
	}

	var req service.UpdateDashboardReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.dashSvc.UpdateDashboard(r.Context(), orgID, dashboardID, req); err != nil {
		if err.Error() == "dashboard not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{"message": "Dashboard updated"}})
}

// DeleteDashboard removes a dashboard.
// DELETE /analytics/dashboards/{id}
func (h *AnalyticsHandler) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	dashboardID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid dashboard ID")
		return
	}

	if err := h.dashSvc.DeleteDashboard(r.Context(), orgID, dashboardID); err != nil {
		if err.Error() == "dashboard not found" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete dashboard")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]string{"message": "Dashboard deleted"}})
}

// ============================================================
// WIDGET TYPE ENDPOINTS
// ============================================================

// ListWidgetTypes returns all available widget types for the dashboard builder.
// GET /analytics/widget-types
func (h *AnalyticsHandler) ListWidgetTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.dashSvc.ListWidgetTypes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve widget types")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: types})
}
