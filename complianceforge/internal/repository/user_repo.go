package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/complianceforge/platform/internal/models"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) List(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) ([]models.User, int64, error) {
	var total int64
	countQuery := "SELECT COUNT(*) FROM users WHERE organization_id = $1 AND deleted_at IS NULL"
	if params.Search != "" {
		countQuery += " AND (first_name ILIKE $2 OR last_name ILIKE $2 OR email ILIKE $2)"
		r.pool.QueryRow(ctx, countQuery, orgID, "%"+params.Search+"%").Scan(&total)
	} else {
		r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total)
	}

	query := `
		SELECT id, organization_id, email, first_name, last_name, job_title,
			department, phone, avatar_url, status, is_super_admin,
			timezone, language, last_login_at, created_at, updated_at
		FROM users
		WHERE organization_id = $1 AND deleted_at IS NULL`

	args := []interface{}{orgID}
	argIdx := 2

	if params.Search != "" {
		query += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	sortField := "created_at"
	allowed := map[string]bool{"created_at": true, "first_name": true, "last_name": true, "email": true, "last_login_at": true, "status": true}
	if allowed[params.SortBy] {
		sortField = params.SortBy
	}
	sortDir := "DESC"
	if params.SortDir == "asc" {
		sortDir = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortField, sortDir, argIdx, argIdx+1)
	args = append(args, params.PageSize, params.Offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.OrganizationID, &u.Email, &u.FirstName, &u.LastName,
			&u.JobTitle, &u.Department, &u.Phone, &u.AvatarURL, &u.Status,
			&u.IsSuperAdmin, &u.Timezone, &u.Language, &u.LastLoginAt,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepo) GetByID(ctx context.Context, orgID, userID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, organization_id, email, first_name, last_name, job_title,
			department, phone, avatar_url, status, is_super_admin,
			timezone, language, last_login_at, password_changed_at,
			failed_login_attempts, notification_preferences, metadata,
			created_at, updated_at
		FROM users
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`

	var u models.User
	err := r.pool.QueryRow(ctx, query, userID, orgID).Scan(
		&u.ID, &u.OrganizationID, &u.Email, &u.FirstName, &u.LastName,
		&u.JobTitle, &u.Department, &u.Phone, &u.AvatarURL, &u.Status,
		&u.IsSuperAdmin, &u.Timezone, &u.Language, &u.LastLoginAt,
		&u.PasswordChangedAt, &u.FailedLoginAttempts, &u.NotificationPreferences,
		&u.Metadata, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Load roles
	roleRows, err := r.pool.Query(ctx, `
		SELECT r.id, r.name, r.slug, r.description, r.is_system_role
		FROM roles r JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND ur.organization_id = $2 AND r.deleted_at IS NULL`,
		userID, orgID)
	if err == nil {
		defer roleRows.Close()
		for roleRows.Next() {
			var role models.Role
			roleRows.Scan(&role.ID, &role.Name, &role.Slug, &role.Description, &role.IsSystemRole)
			u.Roles = append(u.Roles, role)
		}
	}

	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, orgID, userID uuid.UUID, req models.UpdateUserRequest) error {
	sets := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.FirstName != nil {
		sets = append(sets, fmt.Sprintf("first_name = $%d", argIdx))
		args = append(args, *req.FirstName)
		argIdx++
	}
	if req.LastName != nil {
		sets = append(sets, fmt.Sprintf("last_name = $%d", argIdx))
		args = append(args, *req.LastName)
		argIdx++
	}
	if req.JobTitle != nil {
		sets = append(sets, fmt.Sprintf("job_title = $%d", argIdx))
		args = append(args, *req.JobTitle)
		argIdx++
	}
	if req.Department != nil {
		sets = append(sets, fmt.Sprintf("department = $%d", argIdx))
		args = append(args, *req.Department)
		argIdx++
	}
	if req.Phone != nil {
		sets = append(sets, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *req.Phone)
		argIdx++
	}
	if req.Language != nil {
		sets = append(sets, fmt.Sprintf("language = $%d", argIdx))
		args = append(args, *req.Language)
		argIdx++
	}
	if req.Timezone != nil {
		sets = append(sets, fmt.Sprintf("timezone = $%d", argIdx))
		args = append(args, *req.Timezone)
		argIdx++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d AND organization_id = $%d AND deleted_at IS NULL",
		strings.Join(sets, ", "), argIdx, argIdx+1)
	args = append(args, userID, orgID)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *UserRepo) Deactivate(ctx context.Context, orgID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET status = 'inactive', deleted_at = NOW() WHERE id = $1 AND organization_id = $2",
		userID, orgID)
	return err
}

func (r *UserRepo) AssignRole(ctx context.Context, orgID, userID, roleID, assignedBy uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"INSERT INTO user_roles (user_id, role_id, organization_id, assigned_by) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING",
		userID, roleID, orgID, assignedBy)
	return err
}

func (r *UserRepo) RemoveRole(ctx context.Context, orgID, userID, roleID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2 AND organization_id = $3",
		userID, roleID, orgID)
	return err
}

func (r *UserRepo) ListRoles(ctx context.Context, orgID uuid.UUID) ([]models.Role, error) {
	query := `
		SELECT r.id, r.organization_id, r.name, r.slug, r.description, r.is_system_role, r.is_custom,
			r.created_at, r.updated_at
		FROM roles r
		WHERE (r.organization_id = $1 OR r.organization_id IS NULL) AND r.deleted_at IS NULL
		ORDER BY r.is_system_role DESC, r.name ASC`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Slug,
			&role.Description, &role.IsSystemRole, &role.IsCustom,
			&role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *UserRepo) GetAuditLog(ctx context.Context, orgID uuid.UUID, params models.PaginationParams) ([]models.AuditLog, int64, error) {
	var total int64
	r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1", orgID).Scan(&total)

	query := `
		SELECT al.id, al.organization_id, al.user_id, al.action, al.entity_type,
			al.entity_id, al.changes, al.ip_address, al.user_agent, al.metadata, al.created_at
		FROM audit_logs al
		WHERE al.organization_id = $1
		ORDER BY al.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, orgID, params.PageSize, params.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.OrganizationID, &l.UserID, &l.Action,
			&l.EntityType, &l.EntityID, &l.Changes, &l.IPAddress,
			&l.UserAgent, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}
