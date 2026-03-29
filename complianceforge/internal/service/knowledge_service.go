package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================
// KNOWLEDGE BASE SERVICE
// Manages compliance guidance articles, bookmarks, feedback,
// and personalised recommendations. System articles are shared
// across all tenants; org-specific articles are tenant-scoped.
// ============================================================

// KnowledgeService manages the compliance knowledge base.
type KnowledgeService struct {
	pool *pgxpool.Pool
}

// NewKnowledgeService creates a new KnowledgeService.
func NewKnowledgeService(pool *pgxpool.Pool) *KnowledgeService {
	return &KnowledgeService{pool: pool}
}

// ============================================================
// DATA TYPES
// ============================================================

// KnowledgeArticle represents a knowledge base article.
type KnowledgeArticle struct {
	ID                    uuid.UUID  `json:"id"`
	OrganizationID        *uuid.UUID `json:"organization_id,omitempty"`
	ArticleType           string     `json:"article_type"`
	Title                 string     `json:"title"`
	Slug                  string     `json:"slug"`
	ContentMarkdown       string     `json:"content_markdown"`
	Summary               string     `json:"summary"`
	ApplicableFrameworks  []string   `json:"applicable_frameworks"`
	ApplicableControlCodes []string  `json:"applicable_control_codes"`
	Tags                  []string   `json:"tags"`
	Difficulty            string     `json:"difficulty"`
	ReadingTimeMinutes    int        `json:"reading_time_minutes"`
	IsSystem              bool       `json:"is_system"`
	IsPublished           bool       `json:"is_published"`
	ViewCount             int        `json:"view_count"`
	HelpfulCount          int        `json:"helpful_count"`
	NotHelpfulCount       int        `json:"not_helpful_count"`
	AuthorUserID          *uuid.UUID `json:"author_user_id,omitempty"`
	IsBookmarked          bool       `json:"is_bookmarked,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CreateArticleRequest is the request to create a knowledge article.
type CreateArticleRequest struct {
	ArticleType            string   `json:"article_type"`
	Title                  string   `json:"title"`
	Slug                   string   `json:"slug"`
	ContentMarkdown        string   `json:"content_markdown"`
	Summary                string   `json:"summary"`
	ApplicableFrameworks   []string `json:"applicable_frameworks"`
	ApplicableControlCodes []string `json:"applicable_control_codes"`
	Tags                   []string `json:"tags"`
	Difficulty             string   `json:"difficulty"`
	ReadingTimeMinutes     int      `json:"reading_time_minutes"`
	IsPublished            bool     `json:"is_published"`
}

// UpdateArticleRequest is the request to update a knowledge article.
type UpdateArticleRequest struct {
	Title                  *string  `json:"title,omitempty"`
	ContentMarkdown        *string  `json:"content_markdown,omitempty"`
	Summary                *string  `json:"summary,omitempty"`
	ApplicableFrameworks   []string `json:"applicable_frameworks,omitempty"`
	ApplicableControlCodes []string `json:"applicable_control_codes,omitempty"`
	Tags                   []string `json:"tags,omitempty"`
	Difficulty             *string  `json:"difficulty,omitempty"`
	ReadingTimeMinutes     *int     `json:"reading_time_minutes,omitempty"`
	IsPublished            *bool    `json:"is_published,omitempty"`
}

// ArticleFeedback represents a user's feedback on an article.
type ArticleFeedback struct {
	ArticleID uuid.UUID `json:"article_id"`
	Action    string    `json:"action"` // "helpful", "not_helpful", "view"
	Comment   string    `json:"comment,omitempty"`
}

// KnowledgeBrowseRequest holds parameters for browsing articles.
type KnowledgeBrowseRequest struct {
	Query      string   `json:"query,omitempty"`
	Types      []string `json:"types,omitempty"`
	Frameworks []string `json:"frameworks,omitempty"`
	Difficulty string   `json:"difficulty,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	SortBy     string   `json:"sort_by,omitempty"` // relevance, date, popular, helpful
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
}

// KnowledgeBrowseResponse holds paginated article results.
type KnowledgeBrowseResponse struct {
	Articles   []KnowledgeArticle `json:"articles"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// ============================================================
// BROWSE & SEARCH
// ============================================================

// BrowseArticles returns paginated, filtered knowledge articles.
// Includes both system articles and the org's own articles.
func (s *KnowledgeService) BrowseArticles(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, req KnowledgeBrowseRequest) (*KnowledgeBrowseResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 50 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "date"
	}

	args := []interface{}{orgID}
	argN := 2

	conditions := []string{
		"(ka.organization_id = $1 OR ka.is_system = true)",
		"ka.is_published = true",
	}

	if req.Query != "" {
		conditions = append(conditions, fmt.Sprintf(
			"ka.search_vector @@ to_tsquery('english', $%d)", argN))
		tsq := BuildTSQuery(req.Query)
		args = append(args, tsq)
		argN++
	}
	if len(req.Types) > 0 {
		conditions = append(conditions, fmt.Sprintf("ka.article_type::text = ANY($%d)", argN))
		args = append(args, req.Types)
		argN++
	}
	if len(req.Frameworks) > 0 {
		conditions = append(conditions, fmt.Sprintf("ka.applicable_frameworks && $%d", argN))
		args = append(args, req.Frameworks)
		argN++
	}
	if req.Difficulty != "" {
		conditions = append(conditions, fmt.Sprintf("ka.difficulty::text = $%d", argN))
		args = append(args, req.Difficulty)
		argN++
	}
	if len(req.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("ka.tags && $%d", argN))
		args = append(args, req.Tags)
		argN++
	}

	where := strings.Join(conditions, " AND ")

	orderBy := "ka.updated_at DESC"
	switch req.SortBy {
	case "popular":
		orderBy = "ka.view_count DESC"
	case "helpful":
		orderBy = "ka.helpful_count DESC"
	case "relevance":
		if req.Query != "" {
			orderBy = "ts_rank_cd(ka.search_vector, to_tsquery('english', '" +
				strings.ReplaceAll(BuildTSQuery(req.Query), "'", "''") + "')) DESC"
		}
	}

	offset := (req.Page - 1) * req.PageSize

	query := fmt.Sprintf(`
		SELECT ka.id, ka.organization_id, ka.article_type::text, ka.title, ka.slug,
			ka.content_markdown, COALESCE(ka.summary, ''),
			COALESCE(ka.applicable_frameworks, '{}'),
			COALESCE(ka.applicable_control_codes, '{}'),
			COALESCE(ka.tags, '{}'),
			ka.difficulty::text, ka.reading_time_minutes,
			ka.is_system, ka.is_published,
			ka.view_count, ka.helpful_count, ka.not_helpful_count,
			ka.author_user_id,
			EXISTS(SELECT 1 FROM knowledge_bookmarks kb WHERE kb.article_id = ka.id AND kb.user_id = $%d) AS is_bookmarked,
			ka.created_at, ka.updated_at
		FROM knowledge_articles ka
		WHERE %s
		ORDER BY %s
		LIMIT %d OFFSET %d
	`, argN, where, orderBy, req.PageSize, offset)
	args = append(args, userID)
	argN++

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_articles ka WHERE %s", where)
	countArgs := args[:len(args)-1] // exclude userID

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("browse begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// No need for SET LOCAL here since system articles have NULL org
	// and the RLS policy allows them

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("browse query: %w", err)
	}
	defer rows.Close()

	var articles []KnowledgeArticle
	for rows.Next() {
		var a KnowledgeArticle
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.ArticleType, &a.Title, &a.Slug,
			&a.ContentMarkdown, &a.Summary,
			&a.ApplicableFrameworks, &a.ApplicableControlCodes,
			&a.Tags, &a.Difficulty, &a.ReadingTimeMinutes,
			&a.IsSystem, &a.IsPublished,
			&a.ViewCount, &a.HelpfulCount, &a.NotHelpfulCount,
			&a.AuthorUserID, &a.IsBookmarked,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("browse scan: %w", err)
		}
		articles = append(articles, a)
	}

	var total int
	if err := tx.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("browse count: %w", err)
	}

	tx.Commit(ctx)

	return &KnowledgeBrowseResponse{
		Articles:   articles,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(req.PageSize))),
	}, nil
}

// GetArticleBySlug returns a single article by its URL slug.
func (s *KnowledgeService) GetArticleBySlug(ctx context.Context, orgID, userID uuid.UUID, slug string) (*KnowledgeArticle, error) {
	var a KnowledgeArticle
	err := s.pool.QueryRow(ctx, `
		SELECT ka.id, ka.organization_id, ka.article_type::text, ka.title, ka.slug,
			ka.content_markdown, COALESCE(ka.summary, ''),
			COALESCE(ka.applicable_frameworks, '{}'),
			COALESCE(ka.applicable_control_codes, '{}'),
			COALESCE(ka.tags, '{}'),
			ka.difficulty::text, ka.reading_time_minutes,
			ka.is_system, ka.is_published,
			ka.view_count, ka.helpful_count, ka.not_helpful_count,
			ka.author_user_id,
			EXISTS(SELECT 1 FROM knowledge_bookmarks kb WHERE kb.article_id = ka.id AND kb.user_id = $3) AS is_bookmarked,
			ka.created_at, ka.updated_at
		FROM knowledge_articles ka
		WHERE ka.slug = $2
		  AND (ka.organization_id = $1 OR ka.is_system = true)
		  AND ka.is_published = true
		LIMIT 1
	`, orgID, slug, userID).Scan(
		&a.ID, &a.OrganizationID, &a.ArticleType, &a.Title, &a.Slug,
		&a.ContentMarkdown, &a.Summary,
		&a.ApplicableFrameworks, &a.ApplicableControlCodes,
		&a.Tags, &a.Difficulty, &a.ReadingTimeMinutes,
		&a.IsSystem, &a.IsPublished,
		&a.ViewCount, &a.HelpfulCount, &a.NotHelpfulCount,
		&a.AuthorUserID, &a.IsBookmarked,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("article by slug: %w", err)
	}

	// Increment view count asynchronously
	go func() {
		s.pool.Exec(context.Background(), `
			UPDATE knowledge_articles SET view_count = view_count + 1 WHERE id = $1
		`, a.ID)
	}()

	return &a, nil
}

// GetArticlesForControl returns guidance articles applicable to a specific control.
func (s *KnowledgeService) GetArticlesForControl(ctx context.Context, orgID uuid.UUID, frameworkCode, controlCode string) ([]KnowledgeArticle, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, article_type::text, title, slug,
			'' AS content_markdown, COALESCE(summary, ''),
			COALESCE(applicable_frameworks, '{}'),
			COALESCE(applicable_control_codes, '{}'),
			COALESCE(tags, '{}'),
			difficulty::text, reading_time_minutes,
			is_system, is_published,
			view_count, helpful_count, not_helpful_count,
			author_user_id, created_at, updated_at
		FROM knowledge_articles
		WHERE (organization_id = $1 OR is_system = true)
		  AND is_published = true
		  AND (
			applicable_control_codes && ARRAY[$3]
			OR applicable_frameworks && ARRAY[$2]
		  )
		ORDER BY
			CASE WHEN $3 = ANY(applicable_control_codes) THEN 0 ELSE 1 END,
			helpful_count DESC
		LIMIT 10
	`, orgID, frameworkCode, controlCode)
	if err != nil {
		return nil, fmt.Errorf("articles for control: %w", err)
	}
	defer rows.Close()

	return scanArticleList(rows)
}

// GetRecommendedArticles returns personalized article recommendations
// based on the org's adopted frameworks and popular articles.
func (s *KnowledgeService) GetRecommendedArticles(ctx context.Context, orgID, userID uuid.UUID, limit int) ([]KnowledgeArticle, error) {
	if limit < 1 || limit > 20 {
		limit = 5
	}

	rows, err := s.pool.Query(ctx, `
		SELECT ka.id, ka.organization_id, ka.article_type::text, ka.title, ka.slug,
			'' AS content_markdown, COALESCE(ka.summary, ''),
			COALESCE(ka.applicable_frameworks, '{}'),
			COALESCE(ka.applicable_control_codes, '{}'),
			COALESCE(ka.tags, '{}'),
			ka.difficulty::text, ka.reading_time_minutes,
			ka.is_system, ka.is_published,
			ka.view_count, ka.helpful_count, ka.not_helpful_count,
			ka.author_user_id, ka.created_at, ka.updated_at
		FROM knowledge_articles ka
		WHERE (ka.organization_id = $1 OR ka.is_system = true)
		  AND ka.is_published = true
		  AND NOT EXISTS (
			SELECT 1 FROM knowledge_article_feedback kaf
			WHERE kaf.article_id = ka.id AND kaf.user_id = $2 AND kaf.action = 'view'
		  )
		ORDER BY ka.helpful_count DESC, ka.view_count DESC
		LIMIT $3
	`, orgID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("recommended articles: %w", err)
	}
	defer rows.Close()

	return scanArticleList(rows)
}

func scanArticleList(rows interface{ Next() bool; Scan(...interface{}) error; Close() }) ([]KnowledgeArticle, error) {
	var articles []KnowledgeArticle
	for rows.Next() {
		var a KnowledgeArticle
		if err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.ArticleType, &a.Title, &a.Slug,
			&a.ContentMarkdown, &a.Summary,
			&a.ApplicableFrameworks, &a.ApplicableControlCodes,
			&a.Tags, &a.Difficulty, &a.ReadingTimeMinutes,
			&a.IsSystem, &a.IsPublished,
			&a.ViewCount, &a.HelpfulCount, &a.NotHelpfulCount,
			&a.AuthorUserID, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		articles = append(articles, a)
	}
	return articles, nil
}

// ============================================================
// ARTICLE CRUD (Org-Specific)
// ============================================================

// CreateArticle creates a new org-specific knowledge article.
func (s *KnowledgeService) CreateArticle(ctx context.Context, orgID, userID uuid.UUID, req CreateArticleRequest) (*KnowledgeArticle, error) {
	if req.Slug == "" {
		req.Slug = slugify(req.Title)
	}
	if req.ApplicableFrameworks == nil {
		req.ApplicableFrameworks = []string{}
	}
	if req.ApplicableControlCodes == nil {
		req.ApplicableControlCodes = []string{}
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if req.ReadingTimeMinutes < 1 {
		req.ReadingTimeMinutes = estimateReadingTime(req.ContentMarkdown)
	}

	var a KnowledgeArticle
	err := s.pool.QueryRow(ctx, `
		INSERT INTO knowledge_articles (
			organization_id, article_type, title, slug,
			content_markdown, summary,
			applicable_frameworks, applicable_control_codes,
			tags, difficulty, reading_time_minutes,
			is_system, is_published, author_user_id
		) VALUES ($1, $2::article_type, $3, $4, $5, $6, $7, $8, $9, $10::article_difficulty, $11, false, $12, $13)
		RETURNING id, organization_id, article_type::text, title, slug,
			content_markdown, COALESCE(summary, ''),
			applicable_frameworks, applicable_control_codes,
			tags, difficulty::text, reading_time_minutes,
			is_system, is_published,
			view_count, helpful_count, not_helpful_count,
			author_user_id, created_at, updated_at
	`, orgID, req.ArticleType, req.Title, req.Slug,
		req.ContentMarkdown, req.Summary,
		req.ApplicableFrameworks, req.ApplicableControlCodes,
		req.Tags, req.Difficulty, req.ReadingTimeMinutes,
		req.IsPublished, userID,
	).Scan(
		&a.ID, &a.OrganizationID, &a.ArticleType, &a.Title, &a.Slug,
		&a.ContentMarkdown, &a.Summary,
		&a.ApplicableFrameworks, &a.ApplicableControlCodes,
		&a.Tags, &a.Difficulty, &a.ReadingTimeMinutes,
		&a.IsSystem, &a.IsPublished,
		&a.ViewCount, &a.HelpfulCount, &a.NotHelpfulCount,
		&a.AuthorUserID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}
	return &a, nil
}

// UpdateArticle updates an org-specific knowledge article.
func (s *KnowledgeService) UpdateArticle(ctx context.Context, orgID uuid.UUID, articleID uuid.UUID, req UpdateArticleRequest) (*KnowledgeArticle, error) {
	// Build dynamic SET clauses
	sets := []string{}
	args := []interface{}{orgID, articleID}
	argN := 3

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argN))
		args = append(args, *req.Title)
		argN++
	}
	if req.ContentMarkdown != nil {
		sets = append(sets, fmt.Sprintf("content_markdown = $%d", argN))
		args = append(args, *req.ContentMarkdown)
		argN++
	}
	if req.Summary != nil {
		sets = append(sets, fmt.Sprintf("summary = $%d", argN))
		args = append(args, *req.Summary)
		argN++
	}
	if req.ApplicableFrameworks != nil {
		sets = append(sets, fmt.Sprintf("applicable_frameworks = $%d", argN))
		args = append(args, req.ApplicableFrameworks)
		argN++
	}
	if req.ApplicableControlCodes != nil {
		sets = append(sets, fmt.Sprintf("applicable_control_codes = $%d", argN))
		args = append(args, req.ApplicableControlCodes)
		argN++
	}
	if req.Tags != nil {
		sets = append(sets, fmt.Sprintf("tags = $%d", argN))
		args = append(args, req.Tags)
		argN++
	}
	if req.Difficulty != nil {
		sets = append(sets, fmt.Sprintf("difficulty = $%d::article_difficulty", argN))
		args = append(args, *req.Difficulty)
		argN++
	}
	if req.ReadingTimeMinutes != nil {
		sets = append(sets, fmt.Sprintf("reading_time_minutes = $%d", argN))
		args = append(args, *req.ReadingTimeMinutes)
		argN++
	}
	if req.IsPublished != nil {
		sets = append(sets, fmt.Sprintf("is_published = $%d", argN))
		args = append(args, *req.IsPublished)
		argN++
	}

	if len(sets) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	sets = append(sets, "updated_at = NOW()")

	query := fmt.Sprintf(`
		UPDATE knowledge_articles
		SET %s
		WHERE id = $2 AND organization_id = $1 AND is_system = false
		RETURNING id, organization_id, article_type::text, title, slug,
			content_markdown, COALESCE(summary, ''),
			applicable_frameworks, applicable_control_codes,
			tags, difficulty::text, reading_time_minutes,
			is_system, is_published,
			view_count, helpful_count, not_helpful_count,
			author_user_id, created_at, updated_at
	`, strings.Join(sets, ", "))

	var a KnowledgeArticle
	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&a.ID, &a.OrganizationID, &a.ArticleType, &a.Title, &a.Slug,
		&a.ContentMarkdown, &a.Summary,
		&a.ApplicableFrameworks, &a.ApplicableControlCodes,
		&a.Tags, &a.Difficulty, &a.ReadingTimeMinutes,
		&a.IsSystem, &a.IsPublished,
		&a.ViewCount, &a.HelpfulCount, &a.NotHelpfulCount,
		&a.AuthorUserID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update article: %w", err)
	}
	return &a, nil
}

// ============================================================
// FEEDBACK & BOOKMARKS
// ============================================================

// TrackArticleEngagement records a user action on an article.
func (s *KnowledgeService) TrackArticleEngagement(ctx context.Context, articleID, userID uuid.UUID, feedback ArticleFeedback) error {
	// Upsert feedback
	_, err := s.pool.Exec(ctx, `
		INSERT INTO knowledge_article_feedback (article_id, user_id, action, comment)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (article_id, user_id, action)
		DO UPDATE SET comment = EXCLUDED.comment, created_at = NOW()
	`, articleID, userID, feedback.Action, feedback.Comment)
	if err != nil {
		return fmt.Errorf("track engagement: %w", err)
	}

	// Update article counters
	switch feedback.Action {
	case "helpful":
		_, err = s.pool.Exec(ctx, `
			UPDATE knowledge_articles SET helpful_count = helpful_count + 1 WHERE id = $1
		`, articleID)
	case "not_helpful":
		_, err = s.pool.Exec(ctx, `
			UPDATE knowledge_articles SET not_helpful_count = not_helpful_count + 1 WHERE id = $1
		`, articleID)
	case "view":
		_, err = s.pool.Exec(ctx, `
			UPDATE knowledge_articles SET view_count = view_count + 1 WHERE id = $1
		`, articleID)
	}
	return err
}

// BookmarkArticle adds or removes a bookmark on an article.
func (s *KnowledgeService) BookmarkArticle(ctx context.Context, userID, articleID uuid.UUID) (bool, error) {
	// Toggle: if exists remove, otherwise add
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM knowledge_bookmarks WHERE user_id = $1 AND article_id = $2
	`, userID, articleID)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() > 0 {
		return false, nil // removed
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO knowledge_bookmarks (user_id, article_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, articleID)
	if err != nil {
		return false, err
	}
	return true, nil // added
}

// GetBookmarks returns articles bookmarked by the user.
func (s *KnowledgeService) GetBookmarks(ctx context.Context, orgID, userID uuid.UUID, page, pageSize int) ([]KnowledgeArticle, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	rows, err := s.pool.Query(ctx, `
		SELECT ka.id, ka.organization_id, ka.article_type::text, ka.title, ka.slug,
			'' AS content_markdown, COALESCE(ka.summary, ''),
			COALESCE(ka.applicable_frameworks, '{}'),
			COALESCE(ka.applicable_control_codes, '{}'),
			COALESCE(ka.tags, '{}'),
			ka.difficulty::text, ka.reading_time_minutes,
			ka.is_system, ka.is_published,
			ka.view_count, ka.helpful_count, ka.not_helpful_count,
			ka.author_user_id, ka.created_at, ka.updated_at
		FROM knowledge_articles ka
		JOIN knowledge_bookmarks kb ON kb.article_id = ka.id
		WHERE kb.user_id = $1
		  AND (ka.organization_id = $2 OR ka.is_system = true)
		  AND ka.is_published = true
		ORDER BY kb.created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, orgID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	articles, err := scanArticleList(rows)
	if err != nil {
		return nil, 0, err
	}

	// Mark all as bookmarked
	for i := range articles {
		articles[i].IsBookmarked = true
	}

	var total int
	s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM knowledge_bookmarks kb
		JOIN knowledge_articles ka ON ka.id = kb.article_id
		WHERE kb.user_id = $1 AND (ka.organization_id = $2 OR ka.is_system = true) AND ka.is_published = true
	`, userID, orgID).Scan(&total)

	return articles, total, nil
}

// ============================================================
// HELPERS
// ============================================================

// slugify converts a title to a URL-friendly slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, s)
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

// estimateReadingTime estimates reading time in minutes (200 WPM).
func estimateReadingTime(content string) int {
	words := len(strings.Fields(content))
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}
