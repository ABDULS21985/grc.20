package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// CollaborationHandler
// Handles HTTP requests for comments, activity feed, following,
// and read markers.
// ============================================================

// CollaborationHandler handles collaboration-related HTTP requests.
type CollaborationHandler struct {
	svc *service.CollaborationService
}

// NewCollaborationHandler creates a new CollaborationHandler.
func NewCollaborationHandler(svc *service.CollaborationService) *CollaborationHandler {
	return &CollaborationHandler{svc: svc}
}

// RegisterRoutes mounts all collaboration endpoints on the given router.
func (h *CollaborationHandler) RegisterRoutes(r chi.Router) {
	// Comment endpoints
	r.Get("/comments/{entityType}/{entityId}", h.GetComments)
	r.Post("/comments/{entityType}/{entityId}", h.CreateComment)
	r.Put("/comments/{id}", h.EditComment)
	r.Delete("/comments/{id}", h.DeleteComment)
	r.Post("/comments/{id}/pin", h.PinComment)
	r.Post("/comments/{id}/react", h.ReactToComment)

	// Activity feed endpoints
	r.Get("/activity/feed", h.GetPersonalFeed)
	r.Get("/activity/org", h.GetOrgFeed)
	r.Get("/activity/{entityType}/{entityId}", h.GetEntityActivity)
	r.Get("/activity/unread", h.GetUnreadCounts)
	r.Post("/activity/{entityType}/{entityId}/mark-read", h.MarkRead)

	// Following endpoints
	r.Get("/following", h.GetFollowing)
	r.Post("/following/{entityType}/{entityId}", h.FollowEntity)
	r.Delete("/following/{entityType}/{entityId}", h.UnfollowEntity)
}

// ============================================================
// COMMENT ENDPOINTS
// ============================================================

// GetComments returns threaded comments for an entity.
// GET /comments/{entityType}/{entityId}
func (h *CollaborationHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	entityType := chi.URLParam(r, "entityType")
	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid entity ID")
		return
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "oldest"
	}

	comments, err := h.svc.GetComments(r.Context(), orgID, entityType, entityID, sortBy)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve comments")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: comments})
}

// CreateComment creates a new comment on an entity.
// POST /comments/{entityType}/{entityId}
func (h *CollaborationHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	entityType := chi.URLParam(r, "entityType")
	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid entity ID")
		return
	}

	var req service.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Comment content is required")
		return
	}

	comment, err := h.svc.CreateComment(r.Context(), orgID, userID, entityType, entityID, req)
	if err != nil {
		if isUserError(err) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create comment")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: comment})
}

// EditComment edits an existing comment.
// PUT /comments/{id}
func (h *CollaborationHandler) EditComment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid comment ID")
		return
	}

	var req service.EditCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Comment content is required")
		return
	}

	comment, err := h.svc.EditComment(r.Context(), orgID, userID, commentID, req)
	if err != nil {
		if isUserError(err) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to edit comment")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: comment})
}

// DeleteComment soft-deletes a comment.
// DELETE /comments/{id}
func (h *CollaborationHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid comment ID")
		return
	}

	// Check if the user has admin role
	roles := middleware.GetRoles(r.Context())
	isAdmin := false
	for _, role := range roles {
		if role == "admin" || role == "org_admin" || role == "super_admin" {
			isAdmin = true
			break
		}
	}

	err = h.svc.DeleteComment(r.Context(), orgID, userID, commentID, isAdmin)
	if err != nil {
		if isUserError(err) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete comment")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Comment deleted successfully"},
	})
}

// PinComment toggles pin state on a comment.
// POST /comments/{id}/pin
func (h *CollaborationHandler) PinComment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid comment ID")
		return
	}

	comment, err := h.svc.PinComment(r.Context(), orgID, commentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to toggle pin")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: comment})
}

// ReactToComment toggles a reaction on a comment.
// POST /comments/{id}/react
func (h *CollaborationHandler) ReactToComment(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid comment ID")
		return
	}

	var req service.ReactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.ReactionType == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reaction_type is required")
		return
	}

	reactions, err := h.svc.ReactToComment(r.Context(), orgID, userID, commentID, req.ReactionType)
	if err != nil {
		if isUserError(err) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to toggle reaction")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: reactions})
}

// ============================================================
// ACTIVITY FEED ENDPOINTS
// ============================================================

// GetPersonalFeed returns the authenticated user's personal activity feed.
// GET /activity/feed
func (h *CollaborationHandler) GetPersonalFeed(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	params := parsePagination(r)

	filters := service.ActivityFeedFilters{
		EntityType: r.URL.Query().Get("entity_type"),
		Action:     r.URL.Query().Get("action"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}

	entries, total, err := h.svc.GetActivityFeed(r.Context(), orgID, userID, filters, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve activity feed")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: entries,
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

// GetOrgFeed returns the org-wide activity feed (admin view).
// GET /activity/org
func (h *CollaborationHandler) GetOrgFeed(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	params := parsePagination(r)

	filters := service.ActivityFeedFilters{
		EntityType: r.URL.Query().Get("entity_type"),
		Action:     r.URL.Query().Get("action"),
		ActorID:    r.URL.Query().Get("actor_id"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}

	entries, total, err := h.svc.GetOrgActivityFeed(r.Context(), orgID, filters, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve org activity feed")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: entries,
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

// GetEntityActivity returns the activity feed for a specific entity.
// GET /activity/{entityType}/{entityId}
func (h *CollaborationHandler) GetEntityActivity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	entityType := chi.URLParam(r, "entityType")
	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid entity ID")
		return
	}

	params := parsePagination(r)

	entries, total, err := h.svc.GetEntityActivity(r.Context(), orgID, entityType, entityID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve entity activity")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: entries,
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

// GetUnreadCounts returns unread activity/comment counts for the user.
// GET /activity/unread
func (h *CollaborationHandler) GetUnreadCounts(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	counts, err := h.svc.GetUnreadCounts(r.Context(), orgID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve unread counts")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: counts})
}

// MarkRead marks an entity as read for the authenticated user.
// POST /activity/{entityType}/{entityId}/mark-read
func (h *CollaborationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	entityType := chi.URLParam(r, "entityType")
	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid entity ID")
		return
	}

	err = h.svc.MarkEntityRead(r.Context(), orgID, userID, entityType, entityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark entity as read")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Marked as read"},
	})
}

// ============================================================
// FOLLOWING ENDPOINTS
// ============================================================

// GetFollowing returns entities the authenticated user is following.
// GET /following
func (h *CollaborationHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	params := parsePagination(r)

	follows, total, err := h.svc.GetFollowedEntities(r.Context(), orgID, userID, params.Page, params.PageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve followed entities")
		return
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize != 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: follows,
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

// FollowEntity creates a follow for the authenticated user.
// POST /following/{entityType}/{entityId}
func (h *CollaborationHandler) FollowEntity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	entityType := chi.URLParam(r, "entityType")
	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid entity ID")
		return
	}

	followType := "watching"
	var body struct {
		FollowType string `json:"follow_type"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.FollowType != "" {
			followType = body.FollowType
		}
	}

	follow, err := h.svc.FollowEntity(r.Context(), orgID, userID, entityType, entityID, followType)
	if err != nil {
		if isUserError(err) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to follow entity")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: follow})
}

// UnfollowEntity removes a follow for the authenticated user.
// DELETE /following/{entityType}/{entityId}
func (h *CollaborationHandler) UnfollowEntity(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	entityType := chi.URLParam(r, "entityType")
	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid entity ID")
		return
	}

	err = h.svc.UnfollowEntity(r.Context(), orgID, userID, entityType, entityID)
	if err != nil {
		if isUserError(err) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unfollow entity")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Unfollowed successfully"},
	})
}

// ============================================================
// HELPER
// ============================================================

// isUserError returns true if the error is caused by user input
// (validation, authorization, not found, etc.) rather than a
// system failure. Used to decide HTTP status codes.
func isUserError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	userPrefixes := []string{
		"comment content is required",
		"parent comment not found",
		"maximum thread depth",
		"cannot edit a deleted",
		"only the author",
		"comments may only be edited",
		"comment already deleted",
		"comment not found",
		"invalid reaction type",
		"invalid follow_type",
		"follow not found",
		"action, entity_type",
	}
	for _, prefix := range userPrefixes {
		if len(msg) >= len(prefix) && msg[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
