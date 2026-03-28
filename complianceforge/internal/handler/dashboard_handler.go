package handler

import (
	"net/http"
	"runtime"

	"github.com/complianceforge/platform/internal/database"
)

// HealthHandler handles system health and dashboard endpoints.
type HealthHandler struct {
	db      *database.DB
	version string
}

func NewHealthHandler(db *database.DB, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

// Health returns the system health status.
// GET /api/v1/health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "healthy"
	if err := h.db.Health(r.Context()); err != nil {
		dbStatus = "unhealthy"
	}

	stats := h.db.Stats()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": h.version,
		"services": map[string]interface{}{
			"database": map[string]interface{}{
				"status":           dbStatus,
				"total_conns":      stats.TotalConns(),
				"acquired_conns":   stats.AcquiredConns(),
				"idle_conns":       stats.IdleConns(),
			},
		},
		"runtime": map[string]interface{}{
			"go_version":    runtime.Version(),
			"goroutines":    runtime.NumGoroutine(),
			"alloc_mb":      mem.Alloc / 1024 / 1024,
			"sys_mb":        mem.Sys / 1024 / 1024,
		},
	})
}

// Ready is a lightweight readiness probe for Kubernetes.
// GET /api/v1/ready
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Health(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
