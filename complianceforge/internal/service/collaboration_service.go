package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// CollaborationService
// Manages comments, activity feed, entity follows, and
// read markers for the entire platform.
// ============================================================

// CollaborationService implements business logic for collaboration features.
type CollaborationService struct {
	pool *pgxpool.Pool
}

// NewCollaborationService creates a new CollaborationService.
func NewCollaborationService(pool *pgxpool.Pool) *CollaborationService {
	return &CollaborationService{pool: pool}
}

// ============================================================
// CONSTANTS
// ============================================================

// MaxThreadDepth is the maximum nesting depth for comment threads.
const MaxThreadDepth = 3

// MaxCommentEditWindowHours is the time window (in hours) after creation
// during which the author may edit their own comment.
const MaxCommentEditWindowHours = 24

// Allowed reaction types.
var allowedReactions = map[string]bool{
	"thumbs_up":   true,
	"thumbs_down": true,
	"check":       true,
	"eyes":        true,
	"rocket":      true,
	"warning":     true,
}

// mentionRegex matches @mentions in comment text.
// Supports @user:UUID and @role:slug patterns.
var mentionUserRegex = regexp.MustCompile(`@user:([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})`)
var mentionRoleRegex = regexp.MustCompile(`@role:([a-z_]+)`)

// ============================================================
// DATA TYPES — Comments
// ============================================================

// Comment represents a comment on an entity.
type Comment struct {
	ID                 uuid.UUID           `json:"id"`
	OrganizationID     uuid.UUID           `json:"organization_id"`
	EntityType         string              `json:"entity_type"`
	EntityID           uuid.UUID           `json:"entity_id"`
	ParentCommentID    *uuid.UUID          `json:"parent_comment_id,omitempty"`
	ThreadDepth        int                 `json:"thread_depth"`
	AuthorUserID       uuid.UUID           `json:"author_user_id"`
	AuthorEmail        string              `json:"author_email,omitempty"`
	AuthorName         string              `json:"author_name,omitempty"`
	Content            string              `json:"content"`
	ContentHTML        string              `json:"content_html"`
	IsInternal         bool                `json:"is_internal"`
	IsResolutionNote   bool                `json:"is_resolution_note"`
	IsPinned           bool                `json:"is_pinned"`
	MentionedUserIDs   []uuid.UUID         `json:"mentioned_user_ids"`
	MentionedRoleSlugs []string            `json:"mentioned_role_slugs"`
	AttachmentPaths    []string            `json:"attachment_paths"`
	AttachmentNames    []string            `json:"attachment_names"`
	AttachmentSizes    []int64             `json:"attachment_sizes"`
	Reactions          map[string][]string `json:"reactions"`
	IsEdited           bool                `json:"is_edited"`
	EditedAt           *time.Time          `json:"edited_at,omitempty"`
	IsDeleted          bool                `json:"is_deleted"`
	DeletedAt          *time.Time          `json:"deleted_at,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	Replies            []Comment           `json:"replies,omitempty"`
}

// CreateCommentRequest is the payload for creating a comment.
type CreateCommentRequest struct {
	Content          string     `json:"content"`
	ParentCommentID  *uuid.UUID `json:"parent_comment_id,omitempty"`
	IsInternal       bool       `json:"is_internal"`
	IsResolutionNote bool       `json:"is_resolution_note"`
	AttachmentPaths  []string   `json:"attachment_paths,omitempty"`
	AttachmentNames  []string   `json:"attachment_names,omitempty"`
	AttachmentSizes  []int64    `json:"attachment_sizes,omitempty"`
}

// EditCommentRequest is the payload for editing a comment.
type EditCommentRequest struct {
	Content string `json:"content"`
}

// ReactRequest is the payload for toggling a reaction.
type ReactRequest struct {
	ReactionType string `json:"reaction_type"`
}

// ============================================================
// DATA TYPES — Activity Feed
// ============================================================

// ActivityEntry represents a single entry in the activity feed.
type ActivityEntry struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	ActorUserID    *uuid.UUID      `json:"actor_user_id,omitempty"`
	ActorEmail     string          `json:"actor_email,omitempty"`
	ActorName      string          `json:"actor_name,omitempty"`
	Action         string          `json:"action"`
	EntityType     string          `json:"entity_type"`
	EntityID       uuid.UUID       `json:"entity_id"`
	EntityRef      string          `json:"entity_ref"`
	EntityTitle    string          `json:"entity_title"`
	Description    string          `json:"description"`
	Changes        json.RawMessage `json:"changes"`
	IsSystem       bool            `json:"is_system"`
	Visibility     string          `json:"visibility"`
	CreatedAt      time.Time       `json:"created_at"`
}

// RecordActivityRequest is the payload for recording an activity entry.
type RecordActivityRequest struct {
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    uuid.UUID       `json:"entity_id"`
	EntityRef   string          `json:"entity_ref"`
	EntityTitle string          `json:"entity_title"`
	Description string          `json:"description"`
	Changes     json.RawMessage `json:"changes,omitempty"`
	IsSystem    bool            `json:"is_system"`
	Visibility  string          `json:"visibility"`
}

// ActivityFeedFilters controls filtering for activity feed queries.
type ActivityFeedFilters struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Action     string `json:"action"`
	ActorID    string `json:"actor_id"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
}

// ============================================================
// DATA TYPES — Following
// ============================================================

// UserFollow represents a user's follow on an entity.
type UserFollow struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	EntityType     string    `json:"entity_type"`
	EntityID       uuid.UUID `json:"entity_id"`
	FollowType     string    `json:"follow_type"`
	CreatedAt      time.Time `json:"created_at"`
}

// ============================================================
// DATA TYPES — Read Markers
// ============================================================

// UnreadCounts holds the summary of unread items for a user.
type UnreadCounts struct {
	TotalUnread     int                `json:"total_unread"`
	ByEntityType    map[string]int     `json:"by_entity_type"`
	UnreadEntityIDs []UnreadEntityInfo `json:"unread_entities"`
}

// UnreadEntityInfo provides unread info for a specific entity.
type UnreadEntityInfo struct {
	EntityType  string    `json:"entity_type"`
	EntityID    uuid.UUID `json:"entity_id"`
	UnreadCount int       `json:"unread_count"`
	LastReadAt  time.Time `json:"last_read_at"`
}

// ============================================================
// RLS HELPER
// ============================================================

func (s *CollaborationService) setRLS(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) error {
	_, err := tx.Exec(ctx, "SET LOCAL app.current_org = '"+orgID.String()+"'")
	return err
}

func (s *CollaborationService) setRLSConn(ctx context.Context, conn *pgxpool.Conn, orgID uuid.UUID) error {
	_, err := conn.Exec(ctx, "SET LOCAL app.current_org = '"+orgID.String()+"'")
	return err
}

// ============================================================
// MENTION PARSING
// ============================================================

// ParseMentions extracts @user:UUID and @role:slug mentions from text.
func ParseMentions(content string) (userIDs []uuid.UUID, roleSlugs []string) {
	userMatches := mentionUserRegex.FindAllStringSubmatch(content, -1)
	seen := make(map[uuid.UUID]bool)
	for _, m := range userMatches {
		if len(m) < 2 {
			continue
		}
		uid, err := uuid.Parse(m[1])
		if err != nil {
			continue
		}
		if !seen[uid] {
			userIDs = append(userIDs, uid)
			seen[uid] = true
		}
	}

	roleMatches := mentionRoleRegex.FindAllStringSubmatch(content, -1)
	seenRoles := make(map[string]bool)
	for _, m := range roleMatches {
		if len(m) < 2 {
			continue
		}
		slug := m[1]
		if !seenRoles[slug] {
			roleSlugs = append(roleSlugs, slug)
			seenRoles[slug] = true
		}
	}

	return userIDs, roleSlugs
}

// ============================================================
// HTML SANITIZATION
// ============================================================

// SanitizeContentHTML produces a safe HTML version of the content.
// We escape HTML entities to prevent XSS. The frontend renders
// markdown on its own, so we store the escaped version for
// contexts that need pre-rendered HTML.
func SanitizeContentHTML(content string) string {
	escaped := html.EscapeString(content)

	// Convert line breaks to <br> for simple display contexts
	escaped = strings.ReplaceAll(escaped, "\n", "<br>")

	return escaped
}

// ============================================================
// COMMENTS — CREATE
// ============================================================

// CreateComment creates a new comment on an entity.
func (s *CollaborationService) CreateComment(
	ctx context.Context,
	orgID, userID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	req CreateCommentRequest,
) (*Comment, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("comment content is required")
	}

	// Parse mentions
	mentionedUsers, mentionedRoles := ParseMentions(req.Content)

	// Sanitize HTML
	contentHTML := SanitizeContentHTML(req.Content)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	// Determine thread depth
	threadDepth := 0
	if req.ParentCommentID != nil {
		var parentDepth int
		err := tx.QueryRow(ctx,
			`SELECT thread_depth FROM comments
			 WHERE id = $1 AND organization_id = $2 AND is_deleted = false`,
			req.ParentCommentID, orgID,
		).Scan(&parentDepth)
		if err != nil {
			return nil, fmt.Errorf("parent comment not found: %w", err)
		}
		threadDepth = parentDepth + 1
		if threadDepth > MaxThreadDepth {
			return nil, fmt.Errorf("maximum thread depth of %d exceeded", MaxThreadDepth)
		}
	}

	comment := &Comment{}
	err = tx.QueryRow(ctx,
		`INSERT INTO comments (
			organization_id, entity_type, entity_id, parent_comment_id,
			thread_depth, author_user_id, content, content_html,
			is_internal, is_resolution_note,
			mentioned_user_ids, mentioned_role_slugs,
			attachment_paths, attachment_names, attachment_sizes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id, organization_id, entity_type, entity_id,
			parent_comment_id, thread_depth, author_user_id,
			content, content_html, is_internal, is_resolution_note,
			is_pinned, mentioned_user_ids, mentioned_role_slugs,
			attachment_paths, attachment_names, attachment_sizes,
			reactions, is_edited, is_deleted, created_at, updated_at`,
		orgID, entityType, entityID, req.ParentCommentID,
		threadDepth, userID, req.Content, contentHTML,
		req.IsInternal, req.IsResolutionNote,
		mentionedUsers, mentionedRoles,
		req.AttachmentPaths, req.AttachmentNames, req.AttachmentSizes,
	).Scan(
		&comment.ID, &comment.OrganizationID, &comment.EntityType,
		&comment.EntityID, &comment.ParentCommentID, &comment.ThreadDepth,
		&comment.AuthorUserID, &comment.Content, &comment.ContentHTML,
		&comment.IsInternal, &comment.IsResolutionNote, &comment.IsPinned,
		&comment.MentionedUserIDs, &comment.MentionedRoleSlugs,
		&comment.AttachmentPaths, &comment.AttachmentNames,
		&comment.AttachmentSizes, &comment.Reactions,
		&comment.IsEdited, &comment.IsDeleted,
		&comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert comment: %w", err)
	}

	// Auto-follow the commenter as 'participating'
	_, err = tx.Exec(ctx,
		`INSERT INTO user_follows (organization_id, user_id, entity_type, entity_id, follow_type)
		 VALUES ($1, $2, $3, $4, 'participating')
		 ON CONFLICT (organization_id, user_id, entity_type, entity_id) DO NOTHING`,
		orgID, userID, entityType, entityID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("failed to auto-follow commenter")
	}

	// Auto-follow mentioned users as 'mentioned'
	for _, mentionedUID := range mentionedUsers {
		_, err = tx.Exec(ctx,
			`INSERT INTO user_follows (organization_id, user_id, entity_type, entity_id, follow_type)
			 VALUES ($1, $2, $3, $4, 'mentioned')
			 ON CONFLICT (organization_id, user_id, entity_type, entity_id) DO NOTHING`,
			orgID, mentionedUID, entityType, entityID,
		)
		if err != nil {
			log.Warn().Err(err).Str("mentioned_user", mentionedUID.String()).Msg("failed to auto-follow mentioned user")
		}
	}

	// Increment unread counts for followers (excluding the author)
	_, err = tx.Exec(ctx,
		`UPDATE user_read_markers
		 SET unread_count = unread_count + 1
		 WHERE organization_id = $1
		   AND entity_type = $2
		   AND entity_id = $3
		   AND user_id != $4`,
		orgID, entityType, entityID, userID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("failed to increment unread counts")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("comment_id", comment.ID.String()).
		Str("entity_type", entityType).
		Str("entity_id", entityID.String()).
		Int("mentions", len(mentionedUsers)).
		Msg("comment created")

	return comment, nil
}

// ============================================================
// COMMENTS — EDIT
// ============================================================

// EditComment edits a comment's content. Only the author may edit,
// and only within 24 hours of creation.
func (s *CollaborationService) EditComment(
	ctx context.Context,
	orgID, userID, commentID uuid.UUID,
	req EditCommentRequest,
) (*Comment, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("comment content is required")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	// Fetch the existing comment to verify ownership and edit window
	var authorID uuid.UUID
	var createdAt time.Time
	var isDeleted bool
	err = tx.QueryRow(ctx,
		`SELECT author_user_id, created_at, is_deleted FROM comments
		 WHERE id = $1 AND organization_id = $2`,
		commentID, orgID,
	).Scan(&authorID, &createdAt, &isDeleted)
	if err != nil {
		return nil, fmt.Errorf("comment not found: %w", err)
	}
	if isDeleted {
		return nil, fmt.Errorf("cannot edit a deleted comment")
	}
	if authorID != userID {
		return nil, fmt.Errorf("only the author may edit this comment")
	}
	if time.Since(createdAt).Hours() > MaxCommentEditWindowHours {
		return nil, fmt.Errorf("comments may only be edited within %d hours of creation", MaxCommentEditWindowHours)
	}

	// Re-parse mentions
	mentionedUsers, mentionedRoles := ParseMentions(req.Content)
	contentHTML := SanitizeContentHTML(req.Content)
	now := time.Now().UTC()

	comment := &Comment{}
	err = tx.QueryRow(ctx,
		`UPDATE comments SET
			content = $1,
			content_html = $2,
			mentioned_user_ids = $3,
			mentioned_role_slugs = $4,
			is_edited = true,
			edited_at = $5
		WHERE id = $6 AND organization_id = $7
		RETURNING id, organization_id, entity_type, entity_id,
			parent_comment_id, thread_depth, author_user_id,
			content, content_html, is_internal, is_resolution_note,
			is_pinned, mentioned_user_ids, mentioned_role_slugs,
			attachment_paths, attachment_names, attachment_sizes,
			reactions, is_edited, edited_at, is_deleted, created_at, updated_at`,
		req.Content, contentHTML, mentionedUsers, mentionedRoles,
		now, commentID, orgID,
	).Scan(
		&comment.ID, &comment.OrganizationID, &comment.EntityType,
		&comment.EntityID, &comment.ParentCommentID, &comment.ThreadDepth,
		&comment.AuthorUserID, &comment.Content, &comment.ContentHTML,
		&comment.IsInternal, &comment.IsResolutionNote, &comment.IsPinned,
		&comment.MentionedUserIDs, &comment.MentionedRoleSlugs,
		&comment.AttachmentPaths, &comment.AttachmentNames,
		&comment.AttachmentSizes, &comment.Reactions,
		&comment.IsEdited, &comment.EditedAt,
		&comment.IsDeleted, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update comment: %w", err)
	}

	// Auto-follow newly mentioned users
	for _, mentionedUID := range mentionedUsers {
		_, _ = tx.Exec(ctx,
			`INSERT INTO user_follows (organization_id, user_id, entity_type, entity_id, follow_type)
			 VALUES ($1, $2, $3, $4, 'mentioned')
			 ON CONFLICT (organization_id, user_id, entity_type, entity_id) DO NOTHING`,
			orgID, mentionedUID, comment.EntityType, comment.EntityID,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("comment_id", commentID.String()).
		Msg("comment edited")

	return comment, nil
}

// ============================================================
// COMMENTS — DELETE (soft)
// ============================================================

// DeleteComment soft-deletes a comment. Only the author or an admin
// (identified by isAdmin flag) may delete.
func (s *CollaborationService) DeleteComment(
	ctx context.Context,
	orgID, userID, commentID uuid.UUID,
	isAdmin bool,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set rls: %w", err)
	}

	// Verify ownership or admin
	var authorID uuid.UUID
	var isDeleted bool
	err = tx.QueryRow(ctx,
		`SELECT author_user_id, is_deleted FROM comments
		 WHERE id = $1 AND organization_id = $2`,
		commentID, orgID,
	).Scan(&authorID, &isDeleted)
	if err != nil {
		return fmt.Errorf("comment not found: %w", err)
	}
	if isDeleted {
		return fmt.Errorf("comment already deleted")
	}
	if authorID != userID && !isAdmin {
		return fmt.Errorf("only the author or an admin may delete this comment")
	}

	now := time.Now().UTC()
	_, err = tx.Exec(ctx,
		`UPDATE comments SET
			is_deleted = true,
			deleted_at = $1,
			deleted_by_user_id = $2
		WHERE id = $3 AND organization_id = $4`,
		now, userID, commentID, orgID,
	)
	if err != nil {
		return fmt.Errorf("soft delete comment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("comment_id", commentID.String()).
		Bool("is_admin", isAdmin).
		Msg("comment deleted")

	return nil
}

// ============================================================
// COMMENTS — PIN
// ============================================================

// PinComment toggles the pinned state of a comment.
func (s *CollaborationService) PinComment(
	ctx context.Context,
	orgID, commentID uuid.UUID,
) (*Comment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	comment := &Comment{}
	err = tx.QueryRow(ctx,
		`UPDATE comments SET is_pinned = NOT is_pinned
		 WHERE id = $1 AND organization_id = $2 AND is_deleted = false
		 RETURNING id, organization_id, entity_type, entity_id,
			parent_comment_id, thread_depth, author_user_id,
			content, content_html, is_internal, is_resolution_note,
			is_pinned, mentioned_user_ids, mentioned_role_slugs,
			attachment_paths, attachment_names, attachment_sizes,
			reactions, is_edited, is_deleted, created_at, updated_at`,
		commentID, orgID,
	).Scan(
		&comment.ID, &comment.OrganizationID, &comment.EntityType,
		&comment.EntityID, &comment.ParentCommentID, &comment.ThreadDepth,
		&comment.AuthorUserID, &comment.Content, &comment.ContentHTML,
		&comment.IsInternal, &comment.IsResolutionNote, &comment.IsPinned,
		&comment.MentionedUserIDs, &comment.MentionedRoleSlugs,
		&comment.AttachmentPaths, &comment.AttachmentNames,
		&comment.AttachmentSizes, &comment.Reactions,
		&comment.IsEdited, &comment.IsDeleted,
		&comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("toggle pin: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return comment, nil
}

// ============================================================
// COMMENTS — REACT
// ============================================================

// ReactToComment toggles a reaction on a comment for a user.
// If the user already has the reaction, it is removed.
func (s *CollaborationService) ReactToComment(
	ctx context.Context,
	orgID, userID, commentID uuid.UUID,
	reactionType string,
) (map[string][]string, error) {
	if !allowedReactions[reactionType] {
		return nil, fmt.Errorf("invalid reaction type: %s", reactionType)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	// Read current reactions
	var reactionsJSON []byte
	err = tx.QueryRow(ctx,
		`SELECT reactions FROM comments
		 WHERE id = $1 AND organization_id = $2 AND is_deleted = false`,
		commentID, orgID,
	).Scan(&reactionsJSON)
	if err != nil {
		return nil, fmt.Errorf("comment not found: %w", err)
	}

	reactions := make(map[string][]string)
	if len(reactionsJSON) > 0 {
		if err := json.Unmarshal(reactionsJSON, &reactions); err != nil {
			reactions = make(map[string][]string)
		}
	}

	// Toggle the user's reaction
	userStr := userID.String()
	existing := reactions[reactionType]
	found := false
	var updated []string
	for _, uid := range existing {
		if uid == userStr {
			found = true
			continue
		}
		updated = append(updated, uid)
	}
	if !found {
		updated = append(updated, userStr)
	}

	if len(updated) == 0 {
		delete(reactions, reactionType)
	} else {
		reactions[reactionType] = updated
	}

	reactionsBytes, err := json.Marshal(reactions)
	if err != nil {
		return nil, fmt.Errorf("marshal reactions: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE comments SET reactions = $1
		 WHERE id = $2 AND organization_id = $3`,
		reactionsBytes, commentID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("update reactions: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return reactions, nil
}

// ============================================================
// COMMENTS — GET (threaded)
// ============================================================

// GetComments retrieves threaded comments for an entity.
// Pinned comments come first, then sorted by the specified order.
func (s *CollaborationService) GetComments(
	ctx context.Context,
	orgID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	sortBy string,
) ([]Comment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	orderClause := "c.created_at ASC"
	switch sortBy {
	case "newest":
		orderClause = "c.created_at DESC"
	case "oldest":
		orderClause = "c.created_at ASC"
	default:
		orderClause = "c.created_at ASC"
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.organization_id, c.entity_type, c.entity_id,
			c.parent_comment_id, c.thread_depth, c.author_user_id,
			COALESCE(u.email, ''), COALESCE(u.first_name || ' ' || u.last_name, ''),
			c.content, c.content_html, c.is_internal, c.is_resolution_note,
			c.is_pinned, c.mentioned_user_ids, c.mentioned_role_slugs,
			c.attachment_paths, c.attachment_names, c.attachment_sizes,
			c.reactions, c.is_edited, c.edited_at, c.is_deleted, c.created_at, c.updated_at
		FROM comments c
		LEFT JOIN users u ON u.id = c.author_user_id
		WHERE c.organization_id = $1
		  AND c.entity_type = $2
		  AND c.entity_id = $3
		  AND c.is_deleted = false
		ORDER BY c.is_pinned DESC, %s
	`, orderClause)

	rows, err := tx.Query(ctx, query, orgID, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	var allComments []Comment
	for rows.Next() {
		var c Comment
		err := rows.Scan(
			&c.ID, &c.OrganizationID, &c.EntityType, &c.EntityID,
			&c.ParentCommentID, &c.ThreadDepth, &c.AuthorUserID,
			&c.AuthorEmail, &c.AuthorName,
			&c.Content, &c.ContentHTML, &c.IsInternal, &c.IsResolutionNote,
			&c.IsPinned, &c.MentionedUserIDs, &c.MentionedRoleSlugs,
			&c.AttachmentPaths, &c.AttachmentNames, &c.AttachmentSizes,
			&c.Reactions, &c.IsEdited, &c.EditedAt,
			&c.IsDeleted, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		c.Replies = []Comment{}
		allComments = append(allComments, c)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Build threaded structure
	return buildCommentTree(allComments), nil
}

// buildCommentTree assembles flat comments into a threaded tree.
func buildCommentTree(comments []Comment) []Comment {
	byID := make(map[uuid.UUID]*Comment, len(comments))
	var roots []Comment

	// Index all comments
	for i := range comments {
		comments[i].Replies = []Comment{}
		byID[comments[i].ID] = &comments[i]
	}

	// Build tree
	for i := range comments {
		c := &comments[i]
		if c.ParentCommentID == nil {
			roots = append(roots, *c)
		} else {
			parent, ok := byID[*c.ParentCommentID]
			if ok {
				parent.Replies = append(parent.Replies, *c)
			} else {
				// Orphaned reply — treat as root
				roots = append(roots, *c)
			}
		}
	}

	// Copy replies from byID back into roots
	var result []Comment
	for _, root := range roots {
		if ref, ok := byID[root.ID]; ok {
			root.Replies = ref.Replies
		}
		populateReplies(&root, byID)
		result = append(result, root)
	}

	if result == nil {
		result = []Comment{}
	}

	return result
}

// populateReplies recursively populates nested replies from the index.
func populateReplies(c *Comment, byID map[uuid.UUID]*Comment) {
	for i := range c.Replies {
		child := &c.Replies[i]
		if ref, ok := byID[child.ID]; ok {
			child.Replies = ref.Replies
		}
		populateReplies(child, byID)
	}
}

// ============================================================
// ACTIVITY FEED — RECORD
// ============================================================

// RecordActivity creates an immutable activity feed entry and
// increments unread markers for followers.
func (s *CollaborationService) RecordActivity(
	ctx context.Context,
	orgID uuid.UUID,
	userID *uuid.UUID,
	req RecordActivityRequest,
) (*ActivityEntry, error) {
	if req.Action == "" || req.EntityType == "" || req.EntityID == uuid.Nil {
		return nil, fmt.Errorf("action, entity_type, and entity_id are required")
	}

	if req.Visibility == "" {
		req.Visibility = "all"
	}

	if req.Changes == nil {
		req.Changes = json.RawMessage("{}")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	entry := &ActivityEntry{}
	err = tx.QueryRow(ctx,
		`INSERT INTO activity_feed (
			organization_id, actor_user_id, action, entity_type, entity_id,
			entity_ref, entity_title, description, changes, is_system, visibility
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, organization_id, actor_user_id, action, entity_type,
			entity_id, entity_ref, entity_title, description, changes,
			is_system, visibility, created_at`,
		orgID, userID, req.Action, req.EntityType, req.EntityID,
		req.EntityRef, req.EntityTitle, req.Description,
		req.Changes, req.IsSystem, req.Visibility,
	).Scan(
		&entry.ID, &entry.OrganizationID, &entry.ActorUserID,
		&entry.Action, &entry.EntityType, &entry.EntityID,
		&entry.EntityRef, &entry.EntityTitle, &entry.Description,
		&entry.Changes, &entry.IsSystem, &entry.Visibility,
		&entry.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert activity: %w", err)
	}

	// Increment unread markers for followers of this entity
	if userID != nil {
		_, err = tx.Exec(ctx,
			`UPDATE user_read_markers
			 SET unread_count = unread_count + 1
			 WHERE organization_id = $1
			   AND entity_type = $2
			   AND entity_id = $3
			   AND user_id != $4`,
			orgID, req.EntityType, req.EntityID, *userID,
		)
		if err != nil {
			log.Warn().Err(err).Msg("failed to increment unread for activity")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Debug().
		Str("activity_id", entry.ID.String()).
		Str("action", req.Action).
		Str("entity", req.EntityType+"/"+req.EntityID.String()).
		Msg("activity recorded")

	return entry, nil
}

// ============================================================
// ACTIVITY FEED — GET PERSONAL FEED
// ============================================================

// GetActivityFeed returns a paginated activity feed for the user.
// If filters.EntityType and filters.EntityID are set, returns the
// entity-specific feed. Otherwise returns the personal feed
// (entities the user follows or acted on).
func (s *CollaborationService) GetActivityFeed(
	ctx context.Context,
	orgID, userID uuid.UUID,
	filters ActivityFeedFilters,
	page, pageSize int,
) ([]ActivityEntry, int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, 0, fmt.Errorf("set rls: %w", err)
	}

	offset := (page - 1) * pageSize

	// Build query based on filters
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("a.organization_id = $%d", argIdx))
	args = append(args, orgID)
	argIdx++

	// Visibility filter: non-admins can't see admin_only
	conditions = append(conditions, fmt.Sprintf("a.visibility IN ('all', 'internal')"))

	if filters.EntityType != "" && filters.EntityID != "" {
		eid, err := uuid.Parse(filters.EntityID)
		if err == nil {
			conditions = append(conditions, fmt.Sprintf("a.entity_type = $%d", argIdx))
			args = append(args, filters.EntityType)
			argIdx++
			conditions = append(conditions, fmt.Sprintf("a.entity_id = $%d", argIdx))
			args = append(args, eid)
			argIdx++
		}
	} else {
		// Personal feed: show activity for entities the user follows
		conditions = append(conditions, fmt.Sprintf(
			`(a.actor_user_id = $%d OR EXISTS (
				SELECT 1 FROM user_follows uf
				WHERE uf.organization_id = a.organization_id
				  AND uf.user_id = $%d
				  AND uf.entity_type = a.entity_type
				  AND uf.entity_id = a.entity_id
			))`, argIdx, argIdx))
		args = append(args, userID)
		argIdx++
	}

	if filters.Action != "" {
		conditions = append(conditions, fmt.Sprintf("a.action = $%d", argIdx))
		args = append(args, filters.Action)
		argIdx++
	}

	if filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("a.created_at >= $%d::timestamptz", argIdx))
		args = append(args, filters.DateFrom)
		argIdx++
	}

	if filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("a.created_at <= $%d::timestamptz", argIdx))
		args = append(args, filters.DateTo)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count
	var total int64
	countQuery := "SELECT COUNT(*) FROM activity_feed a WHERE " + whereClause
	err = tx.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count activity: %w", err)
	}

	// Fetch
	dataQuery := fmt.Sprintf(`
		SELECT a.id, a.organization_id, a.actor_user_id,
			COALESCE(u.email, ''), COALESCE(u.first_name || ' ' || u.last_name, ''),
			a.action, a.entity_type, a.entity_id, a.entity_ref, a.entity_title,
			a.description, a.changes, a.is_system, a.visibility, a.created_at
		FROM activity_feed a
		LEFT JOIN users u ON u.id = a.actor_user_id
		WHERE %s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := tx.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query activity: %w", err)
	}
	defer rows.Close()

	var entries []ActivityEntry
	for rows.Next() {
		var e ActivityEntry
		err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.ActorUserID,
			&e.ActorEmail, &e.ActorName,
			&e.Action, &e.EntityType, &e.EntityID,
			&e.EntityRef, &e.EntityTitle, &e.Description,
			&e.Changes, &e.IsSystem, &e.Visibility, &e.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan activity: %w", err)
		}
		entries = append(entries, e)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("commit: %w", err)
	}

	if entries == nil {
		entries = []ActivityEntry{}
	}

	return entries, total, nil
}

// ============================================================
// ACTIVITY FEED — ORG-WIDE (admin)
// ============================================================

// GetOrgActivityFeed returns the full org-wide activity feed.
func (s *CollaborationService) GetOrgActivityFeed(
	ctx context.Context,
	orgID uuid.UUID,
	filters ActivityFeedFilters,
	page, pageSize int,
) ([]ActivityEntry, int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, 0, fmt.Errorf("set rls: %w", err)
	}

	offset := (page - 1) * pageSize

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("a.organization_id = $%d", argIdx))
	args = append(args, orgID)
	argIdx++

	if filters.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("a.entity_type = $%d", argIdx))
		args = append(args, filters.EntityType)
		argIdx++
	}

	if filters.Action != "" {
		conditions = append(conditions, fmt.Sprintf("a.action = $%d", argIdx))
		args = append(args, filters.Action)
		argIdx++
	}

	if filters.ActorID != "" {
		actorUID, err := uuid.Parse(filters.ActorID)
		if err == nil {
			conditions = append(conditions, fmt.Sprintf("a.actor_user_id = $%d", argIdx))
			args = append(args, actorUID)
			argIdx++
		}
	}

	if filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("a.created_at >= $%d::timestamptz", argIdx))
		args = append(args, filters.DateFrom)
		argIdx++
	}

	if filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("a.created_at <= $%d::timestamptz", argIdx))
		args = append(args, filters.DateTo)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	var total int64
	err = tx.QueryRow(ctx,
		"SELECT COUNT(*) FROM activity_feed a WHERE "+whereClause,
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count org activity: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT a.id, a.organization_id, a.actor_user_id,
			COALESCE(u.email, ''), COALESCE(u.first_name || ' ' || u.last_name, ''),
			a.action, a.entity_type, a.entity_id, a.entity_ref, a.entity_title,
			a.description, a.changes, a.is_system, a.visibility, a.created_at
		FROM activity_feed a
		LEFT JOIN users u ON u.id = a.actor_user_id
		WHERE %s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := tx.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query org activity: %w", err)
	}
	defer rows.Close()

	var entries []ActivityEntry
	for rows.Next() {
		var e ActivityEntry
		err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.ActorUserID,
			&e.ActorEmail, &e.ActorName,
			&e.Action, &e.EntityType, &e.EntityID,
			&e.EntityRef, &e.EntityTitle, &e.Description,
			&e.Changes, &e.IsSystem, &e.Visibility, &e.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan org activity: %w", err)
		}
		entries = append(entries, e)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("commit: %w", err)
	}

	if entries == nil {
		entries = []ActivityEntry{}
	}

	return entries, total, nil
}

// ============================================================
// ACTIVITY FEED — ENTITY-SPECIFIC
// ============================================================

// GetEntityActivity returns the activity feed for a specific entity.
func (s *CollaborationService) GetEntityActivity(
	ctx context.Context,
	orgID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	page, pageSize int,
) ([]ActivityEntry, int64, error) {
	return s.GetActivityFeed(ctx, orgID, uuid.Nil, ActivityFeedFilters{
		EntityType: entityType,
		EntityID:   entityID.String(),
	}, page, pageSize)
}

// ============================================================
// UNREAD COUNTS
// ============================================================

// GetUnreadCounts returns the unread activity counts for a user.
func (s *CollaborationService) GetUnreadCounts(
	ctx context.Context,
	orgID, userID uuid.UUID,
) (*UnreadCounts, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	rows, err := tx.Query(ctx,
		`SELECT entity_type, entity_id, unread_count, last_read_at
		 FROM user_read_markers
		 WHERE organization_id = $1 AND user_id = $2 AND unread_count > 0
		 ORDER BY unread_count DESC
		 LIMIT 100`,
		orgID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query unread: %w", err)
	}
	defer rows.Close()

	result := &UnreadCounts{
		ByEntityType:    make(map[string]int),
		UnreadEntityIDs: []UnreadEntityInfo{},
	}

	for rows.Next() {
		var info UnreadEntityInfo
		err := rows.Scan(&info.EntityType, &info.EntityID, &info.UnreadCount, &info.LastReadAt)
		if err != nil {
			return nil, fmt.Errorf("scan unread: %w", err)
		}
		result.TotalUnread += info.UnreadCount
		result.ByEntityType[info.EntityType] += info.UnreadCount
		result.UnreadEntityIDs = append(result.UnreadEntityIDs, info)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// ============================================================
// MARK READ
// ============================================================

// MarkEntityRead marks an entity as read for a user, resetting
// the unread count to zero.
func (s *CollaborationService) MarkEntityRead(
	ctx context.Context,
	orgID, userID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set rls: %w", err)
	}

	now := time.Now().UTC()
	_, err = tx.Exec(ctx,
		`INSERT INTO user_read_markers (organization_id, user_id, entity_type, entity_id, last_read_at, unread_count)
		 VALUES ($1, $2, $3, $4, $5, 0)
		 ON CONFLICT (organization_id, user_id, entity_type, entity_id)
		 DO UPDATE SET last_read_at = $5, unread_count = 0`,
		orgID, userID, entityType, entityID, now,
	)
	if err != nil {
		return fmt.Errorf("upsert read marker: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// ============================================================
// FOLLOWING — FOLLOW ENTITY
// ============================================================

// FollowEntity creates or updates a follow for the user on an entity.
func (s *CollaborationService) FollowEntity(
	ctx context.Context,
	orgID, userID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	followType string,
) (*UserFollow, error) {
	if followType == "" {
		followType = "watching"
	}
	if followType != "watching" && followType != "participating" && followType != "mentioned" {
		return nil, fmt.Errorf("invalid follow_type: %s", followType)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	follow := &UserFollow{}
	err = tx.QueryRow(ctx,
		`INSERT INTO user_follows (organization_id, user_id, entity_type, entity_id, follow_type)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (organization_id, user_id, entity_type, entity_id)
		 DO UPDATE SET follow_type = $5
		 RETURNING id, organization_id, user_id, entity_type, entity_id, follow_type, created_at`,
		orgID, userID, entityType, entityID, followType,
	).Scan(
		&follow.ID, &follow.OrganizationID, &follow.UserID,
		&follow.EntityType, &follow.EntityID, &follow.FollowType,
		&follow.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert follow: %w", err)
	}

	// Also ensure a read marker exists
	_, err = tx.Exec(ctx,
		`INSERT INTO user_read_markers (organization_id, user_id, entity_type, entity_id, last_read_at, unread_count)
		 VALUES ($1, $2, $3, $4, NOW(), 0)
		 ON CONFLICT (organization_id, user_id, entity_type, entity_id) DO NOTHING`,
		orgID, userID, entityType, entityID,
	)
	if err != nil {
		log.Warn().Err(err).Msg("failed to create read marker on follow")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("entity", entityType+"/"+entityID.String()).
		Str("follow_type", followType).
		Msg("entity followed")

	return follow, nil
}

// ============================================================
// FOLLOWING — UNFOLLOW ENTITY
// ============================================================

// UnfollowEntity removes a user's follow on an entity.
func (s *CollaborationService) UnfollowEntity(
	ctx context.Context,
	orgID, userID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return fmt.Errorf("set rls: %w", err)
	}

	result, err := tx.Exec(ctx,
		`DELETE FROM user_follows
		 WHERE organization_id = $1 AND user_id = $2
		   AND entity_type = $3 AND entity_id = $4`,
		orgID, userID, entityType, entityID,
	)
	if err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("follow not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("user_id", userID.String()).
		Str("entity", entityType+"/"+entityID.String()).
		Msg("entity unfollowed")

	return nil
}

// ============================================================
// FOLLOWING — GET FOLLOWED ENTITIES
// ============================================================

// GetFollowedEntities returns all entities a user is following.
func (s *CollaborationService) GetFollowedEntities(
	ctx context.Context,
	orgID, userID uuid.UUID,
	page, pageSize int,
) ([]UserFollow, int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.setRLS(ctx, tx, orgID); err != nil {
		return nil, 0, fmt.Errorf("set rls: %w", err)
	}

	offset := (page - 1) * pageSize

	var total int64
	err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_follows
		 WHERE organization_id = $1 AND user_id = $2`,
		orgID, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count follows: %w", err)
	}

	rows, err := tx.Query(ctx,
		`SELECT id, organization_id, user_id, entity_type, entity_id, follow_type, created_at
		 FROM user_follows
		 WHERE organization_id = $1 AND user_id = $2
		 ORDER BY created_at DESC
		 LIMIT $3 OFFSET $4`,
		orgID, userID, pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("query follows: %w", err)
	}
	defer rows.Close()

	var follows []UserFollow
	for rows.Next() {
		var f UserFollow
		err := rows.Scan(&f.ID, &f.OrganizationID, &f.UserID,
			&f.EntityType, &f.EntityID, &f.FollowType, &f.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan follow: %w", err)
		}
		follows = append(follows, f)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("commit: %w", err)
	}

	if follows == nil {
		follows = []UserFollow{}
	}

	return follows, total, nil
}

// ============================================================
// AUTO-FOLLOW RULES
// ============================================================

// AutoFollowOwner adds a 'participating' follow for the owner/assignee
// of an entity. Should be called when an entity is created or assigned.
func (s *CollaborationService) AutoFollowOwner(
	ctx context.Context,
	orgID, ownerUserID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
) {
	_, err := s.FollowEntity(ctx, orgID, ownerUserID, entityType, entityID, "participating")
	if err != nil {
		log.Warn().Err(err).
			Str("user_id", ownerUserID.String()).
			Str("entity", entityType+"/"+entityID.String()).
			Msg("auto-follow owner failed")
	}
}
