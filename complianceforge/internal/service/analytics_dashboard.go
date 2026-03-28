package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================
// DASHBOARD SERVICE — CRUD for Custom Dashboards & Widget Types
// ============================================================

// DashboardService manages custom analytics dashboards and widget types.
type DashboardService struct {
	pool *pgxpool.Pool
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(pool *pgxpool.Pool) *DashboardService {
	return &DashboardService{pool: pool}
}

// ============================================================
// STRUCT DEFINITIONS
// ============================================================

// Dashboard represents a custom analytics dashboard.
type Dashboard struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Layout         json.RawMessage `json:"layout"`
	IsDefault      bool            `json:"is_default"`
	IsShared       bool            `json:"is_shared"`
	OwnerUserID    *uuid.UUID      `json:"owner_user_id,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// WidgetType represents a dashboard widget type from the registry.
type WidgetType struct {
	ID               uuid.UUID       `json:"id"`
	WidgetType       string          `json:"widget_type"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	AvailableMetrics []string        `json:"available_metrics"`
	DefaultConfig    json.RawMessage `json:"default_config"`
	MinWidth         int             `json:"min_width"`
	MinHeight        int             `json:"min_height"`
	CreatedAt        time.Time       `json:"created_at"`
}

// CreateDashboardReq is the request payload for creating a dashboard.
type CreateDashboardReq struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Layout      json.RawMessage `json:"layout"`
	IsDefault   bool            `json:"is_default"`
	IsShared    bool            `json:"is_shared"`
}

// UpdateDashboardReq is the request payload for updating a dashboard.
type UpdateDashboardReq struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Layout      *json.RawMessage `json:"layout,omitempty"`
	IsDefault   *bool            `json:"is_default,omitempty"`
	IsShared    *bool            `json:"is_shared,omitempty"`
}

// ============================================================
// DASHBOARD CRUD
// ============================================================

// ListDashboards returns all dashboards visible to the organization.
// Includes dashboards owned by the user and shared dashboards.
func (s *DashboardService) ListDashboards(ctx context.Context, orgID uuid.UUID) ([]Dashboard, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, name, COALESCE(description, ''),
			layout, is_default, is_shared, owner_user_id,
			created_at, updated_at
		FROM analytics_custom_dashboards
		WHERE organization_id = $1
		ORDER BY is_default DESC, name ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query dashboards: %w", err)
	}
	defer rows.Close()

	var dashboards []Dashboard
	for rows.Next() {
		var d Dashboard
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.Name, &d.Description,
			&d.Layout, &d.IsDefault, &d.IsShared, &d.OwnerUserID,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dashboard: %w", err)
		}
		dashboards = append(dashboards, d)
	}
	return dashboards, nil
}

// CreateDashboard creates a new custom dashboard for the organization.
func (s *DashboardService) CreateDashboard(ctx context.Context, orgID, userID uuid.UUID, req CreateDashboardReq) (*Dashboard, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("dashboard name is required")
	}
	if req.Layout == nil {
		req.Layout = json.RawMessage("[]")
	}

	// If setting as default, clear other defaults for this org
	if req.IsDefault {
		_, err := s.pool.Exec(ctx, `
			UPDATE analytics_custom_dashboards SET is_default = false
			WHERE organization_id = $1 AND is_default = true`, orgID)
		if err != nil {
			return nil, fmt.Errorf("clear default dashboards: %w", err)
		}
	}

	d := &Dashboard{}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO analytics_custom_dashboards
			(organization_id, name, description, layout, is_default, is_shared, owner_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, organization_id, name, COALESCE(description, ''),
			layout, is_default, is_shared, owner_user_id, created_at, updated_at`,
		orgID, req.Name, req.Description, req.Layout,
		req.IsDefault, req.IsShared, userID,
	).Scan(
		&d.ID, &d.OrganizationID, &d.Name, &d.Description,
		&d.Layout, &d.IsDefault, &d.IsShared, &d.OwnerUserID,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert dashboard: %w", err)
	}

	return d, nil
}

// UpdateDashboard updates an existing dashboard's properties.
func (s *DashboardService) UpdateDashboard(ctx context.Context, orgID, dashboardID uuid.UUID, req UpdateDashboardReq) error {
	// If setting as default, clear other defaults
	if req.IsDefault != nil && *req.IsDefault {
		_, err := s.pool.Exec(ctx, `
			UPDATE analytics_custom_dashboards SET is_default = false
			WHERE organization_id = $1 AND is_default = true AND id != $2`,
			orgID, dashboardID)
		if err != nil {
			return fmt.Errorf("clear default dashboards: %w", err)
		}
	}

	// Build dynamic update query
	query := "UPDATE analytics_custom_dashboards SET updated_at = NOW()"
	args := []interface{}{orgID, dashboardID}
	argIdx := 3

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argIdx)
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Layout != nil {
		query += fmt.Sprintf(", layout = $%d", argIdx)
		args = append(args, *req.Layout)
		argIdx++
	}
	if req.IsDefault != nil {
		query += fmt.Sprintf(", is_default = $%d", argIdx)
		args = append(args, *req.IsDefault)
		argIdx++
	}
	if req.IsShared != nil {
		query += fmt.Sprintf(", is_shared = $%d", argIdx)
		args = append(args, *req.IsShared)
		argIdx++
	}

	query += " WHERE organization_id = $1 AND id = $2"

	result, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update dashboard: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("dashboard not found")
	}

	return nil
}

// DeleteDashboard removes a dashboard from the organization.
func (s *DashboardService) DeleteDashboard(ctx context.Context, orgID, dashboardID uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM analytics_custom_dashboards
		WHERE organization_id = $1 AND id = $2`, orgID, dashboardID)
	if err != nil {
		return fmt.Errorf("delete dashboard: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("dashboard not found")
	}
	return nil
}

// ============================================================
// WIDGET TYPES
// ============================================================

// ListWidgetTypes returns all available widget types from the registry.
func (s *DashboardService) ListWidgetTypes(ctx context.Context) ([]WidgetType, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, widget_type, name, COALESCE(description, ''),
			available_metrics, COALESCE(default_config, '{}'),
			min_width, min_height, created_at
		FROM analytics_widget_types
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("query widget types: %w", err)
	}
	defer rows.Close()

	var types []WidgetType
	for rows.Next() {
		var wt WidgetType
		if err := rows.Scan(
			&wt.ID, &wt.WidgetType, &wt.Name, &wt.Description,
			&wt.AvailableMetrics, &wt.DefaultConfig,
			&wt.MinWidth, &wt.MinHeight, &wt.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan widget type: %w", err)
		}
		types = append(types, wt)
	}
	return types, nil
}
