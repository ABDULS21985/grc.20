package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// BoardPortalHandler handles PUBLIC HTTP requests for the executive board
// portal. Access is gated by opaque tokens rather than JWT authentication.
// Board members receive a unique link containing a token; the handler hashes
// the token and resolves it to the correct organisation via the service layer.
type BoardPortalHandler struct {
	svc *service.BoardService
}

// NewBoardPortalHandler creates a new BoardPortalHandler.
func NewBoardPortalHandler(svc *service.BoardService) *BoardPortalHandler {
	return &BoardPortalHandler{svc: svc}
}

// RegisterRoutes mounts all portal routes on the given router.
// These are PUBLIC endpoints — no JWT authentication middleware.
func (h *BoardPortalHandler) RegisterRoutes(r chi.Router) {
	r.Get("/{token}", h.GetDashboard)
	r.Get("/{token}/meetings", h.GetMeetings)
	r.Get("/{token}/meetings/{id}/pack", h.DownloadPack)
	r.Get("/{token}/decisions", h.GetDecisions)
}

// hashToken derives the SHA-256 hash of the raw portal token. The database
// stores only the hash, so we never persist plaintext tokens.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// ============================================================
// PORTAL ENDPOINTS
// ============================================================

// GetDashboard returns the board dashboard for a portal session.
// GET /{token}
func (h *BoardPortalHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TOKEN", "Portal access token is required")
		return
	}

	tokenHash := hashToken(token)
	dashboard, err := h.svc.GetDashboardByToken(r.Context(), tokenHash)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired portal token")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: dashboard})
}

// GetMeetings returns upcoming meetings for a portal session.
// GET /{token}/meetings
func (h *BoardPortalHandler) GetMeetings(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TOKEN", "Portal access token is required")
		return
	}

	tokenHash := hashToken(token)
	meetings, err := h.svc.GetMeetingsByToken(r.Context(), tokenHash)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired portal token")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: meetings})
}

// DownloadPack returns the board pack download path for a portal session.
// GET /{token}/meetings/{id}/pack
func (h *BoardPortalHandler) DownloadPack(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TOKEN", "Portal access token is required")
		return
	}

	meetingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid meeting ID")
		return
	}

	tokenHash := hashToken(token)
	path, err := h.svc.GetBoardPackPathByToken(r.Context(), tokenHash, meetingID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Board pack not available: "+err.Error())
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

// GetDecisions returns recent decisions for a portal session.
// GET /{token}/decisions
func (h *BoardPortalHandler) GetDecisions(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TOKEN", "Portal access token is required")
		return
	}

	tokenHash := hashToken(token)
	decisions, err := h.svc.GetDecisionsByToken(r.Context(), tokenHash)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired portal token")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: decisions})
}
