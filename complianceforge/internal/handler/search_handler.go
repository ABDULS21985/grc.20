package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/service"
)

// ============================================================
// SEARCH & KNOWLEDGE BASE HANDLER
// HTTP layer for unified search, knowledge articles, bookmarks,
// and search analytics endpoints.
// ============================================================

// SearchHandler handles search and knowledge base HTTP requests.
type SearchHandler struct {
	engine      *service.SearchEngine
	knowledge   *service.KnowledgeService
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(engine *service.SearchEngine, knowledge *service.KnowledgeService) *SearchHandler {
	return &SearchHandler{engine: engine, knowledge: knowledge}
}

// RegisterRoutes mounts all search and knowledge base routes.
func (h *SearchHandler) RegisterRoutes(r chi.Router) {
	// ── Unified Search ──────────────────────────────────
	r.Route("/search", func(r chi.Router) {
		r.Get("/", h.Search)
		r.Get("/autocomplete", h.Autocomplete)
		r.Get("/related/{entityType}/{entityId}", h.RelatedEntities)
		r.Get("/recent", h.RecentSearches)
		r.Get("/analytics", h.SearchAnalytics)
		r.Get("/index-stats", h.IndexStats)
		r.Post("/reindex", h.TriggerReindex)
		r.Post("/click", h.RecordClick)
	})

	// ── Knowledge Base ──────────────────────────────────
	r.Route("/knowledge", func(r chi.Router) {
		r.Get("/", h.BrowseArticles)
		r.Get("/recommended", h.RecommendedArticles)
		r.Get("/bookmarks", h.GetBookmarks)
		r.Post("/bookmarks/{articleId}", h.ToggleBookmark)
		r.Get("/for-control/{frameworkCode}/{controlCode}", h.ArticlesForControl)
		r.Post("/articles", h.CreateArticle)
		r.Put("/articles/{id}", h.UpdateArticle)
		r.Post("/articles/{id}/feedback", h.ArticleFeedback)
		r.Get("/{slug}", h.GetArticleBySlug)
	})
}

// ============================================================
// SEARCH ENDPOINTS
// ============================================================

// Search handles GET /search?q=&types=&frameworks=&statuses=&severities=&sort=&page=
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	req := service.SearchRequest{
		Query:    r.URL.Query().Get("q"),
		SortBy:   r.URL.Query().Get("sort"),
		DateFrom: r.URL.Query().Get("date_from"),
		DateTo:   r.URL.Query().Get("date_to"),
	}

	if v := r.URL.Query().Get("types"); v != "" {
		req.EntityTypes = strings.Split(v, ",")
	}
	if v := r.URL.Query().Get("frameworks"); v != "" {
		req.Frameworks = strings.Split(v, ",")
	}
	if v := r.URL.Query().Get("statuses"); v != "" {
		req.Statuses = strings.Split(v, ",")
	}
	if v := r.URL.Query().Get("severities"); v != "" {
		req.Severities = strings.Split(v, ",")
	}
	if v := r.URL.Query().Get("categories"); v != "" {
		req.Categories = strings.Split(v, ",")
	}
	if v := r.URL.Query().Get("tags"); v != "" {
		req.Tags = strings.Split(v, ",")
	}

	req.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	req.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))

	result, err := h.engine.Search(r.Context(), orgID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	// Record search for analytics
	go h.engine.RecordSearch(r.Context(), orgID, userID, req.Query, result.TotalResults)

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// Autocomplete handles GET /search/autocomplete?q=&limit=
func (h *SearchHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	prefix := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	results, err := h.engine.Autocomplete(r.Context(), orgID, prefix, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: results})
}

// RelatedEntities handles GET /search/related/{entityType}/{entityId}
func (h *SearchHandler) RelatedEntities(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	entityType := chi.URLParam(r, "entityType")
	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid entity_id")
		return
	}

	groups, err := h.engine.GetRelatedEntities(r.Context(), orgID, entityType, entityID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: groups})
}

// RecentSearches handles GET /search/recent?limit=
func (h *SearchHandler) RecentSearches(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	results, err := h.engine.GetRecentSearches(r.Context(), orgID, userID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: results})
}

// SearchAnalytics handles GET /search/analytics?days=30
func (h *SearchHandler) SearchAnalytics(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))

	data, err := h.engine.GetSearchAnalytics(r.Context(), orgID, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: data})
}

// IndexStats handles GET /search/index-stats
func (h *SearchHandler) IndexStats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	stats, err := h.engine.GetIndexStats(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: stats})
}

// TriggerReindex handles POST /search/reindex
func (h *SearchHandler) TriggerReindex(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())

	report, err := h.engine.IndexAllEntities(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: report})
}

// RecordClick handles POST /search/click
func (h *SearchHandler) RecordClick(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Query      string `json:"query"`
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid entity_id")
		return
	}

	h.engine.RecordSearchClick(r.Context(), orgID, userID, req.Query, req.EntityType, entityID)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true})
}

// ============================================================
// KNOWLEDGE BASE ENDPOINTS
// ============================================================

// BrowseArticles handles GET /knowledge?query=&types=&frameworks=&difficulty=&sort=&page=
func (h *SearchHandler) BrowseArticles(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	req := service.KnowledgeBrowseRequest{
		Query:      r.URL.Query().Get("query"),
		Difficulty: r.URL.Query().Get("difficulty"),
		SortBy:     r.URL.Query().Get("sort"),
	}
	if v := r.URL.Query().Get("types"); v != "" {
		req.Types = strings.Split(v, ",")
	}
	if v := r.URL.Query().Get("frameworks"); v != "" {
		req.Frameworks = strings.Split(v, ",")
	}
	if v := r.URL.Query().Get("tags"); v != "" {
		req.Tags = strings.Split(v, ",")
	}
	req.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	req.PageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))

	result, err := h.knowledge.BrowseArticles(r.Context(), orgID, userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: result})
}

// GetArticleBySlug handles GET /knowledge/{slug}
func (h *SearchHandler) GetArticleBySlug(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	slug := chi.URLParam(r, "slug")

	article, err := h.knowledge.GetArticleBySlug(r.Context(), orgID, userID, slug)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "article not found")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: article})
}

// ArticlesForControl handles GET /knowledge/for-control/{frameworkCode}/{controlCode}
func (h *SearchHandler) ArticlesForControl(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	frameworkCode := chi.URLParam(r, "frameworkCode")
	controlCode := chi.URLParam(r, "controlCode")

	articles, err := h.knowledge.GetArticlesForControl(r.Context(), orgID, frameworkCode, controlCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: articles})
}

// RecommendedArticles handles GET /knowledge/recommended?limit=
func (h *SearchHandler) RecommendedArticles(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	articles, err := h.knowledge.GetRecommendedArticles(r.Context(), orgID, userID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: articles})
}

// CreateArticle handles POST /knowledge/articles
func (h *SearchHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())

	var req service.CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Title == "" || req.ContentMarkdown == "" || req.ArticleType == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "title, content_markdown, and article_type are required")
		return
	}

	article, err := h.knowledge.CreateArticle(r.Context(), orgID, userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: article})
}

// UpdateArticle handles PUT /knowledge/articles/{id}
func (h *SearchHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	articleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid article id")
		return
	}

	var req service.UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	article, err := h.knowledge.UpdateArticle(r.Context(), orgID, articleID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: article})
}

// ArticleFeedback handles POST /knowledge/articles/{id}/feedback
func (h *SearchHandler) ArticleFeedback(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	articleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid article id")
		return
	}

	var req service.ArticleFeedback
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if req.Action != "helpful" && req.Action != "not_helpful" && req.Action != "view" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "action must be 'helpful', 'not_helpful', or 'view'")
		return
	}

	req.ArticleID = articleID
	if err := h.knowledge.TrackArticleEngagement(r.Context(), articleID, userID, req); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true})
}

// GetBookmarks handles GET /knowledge/bookmarks?page=&page_size=
func (h *SearchHandler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	articles, total, err := h.knowledge.GetBookmarks(r.Context(), orgID, userID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: articles,
		Pagination: models.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: int64(total),
		},
	})
}

// ToggleBookmark handles POST /knowledge/bookmarks/{articleId}
func (h *SearchHandler) ToggleBookmark(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	articleID, err := uuid.Parse(chi.URLParam(r, "articleId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid article id")
		return
	}

	added, err := h.knowledge.BookmarkArticle(r.Context(), userID, articleID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]bool{"bookmarked": added},
	})
}
