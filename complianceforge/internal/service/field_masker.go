package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// FIELD MASKER — field-level permission enforcement
// ============================================================

// FieldMasker applies field-level permissions (visible, masked, hidden) based on
// the ABAC policies assigned to a user. It queries the field_level_permissions
// table and transforms response data accordingly.
type FieldMasker struct {
	pool   *pgxpool.Pool
	engine *ABACEngine
}

// NewFieldMasker creates a new FieldMasker.
func NewFieldMasker(pool *pgxpool.Pool, engine *ABACEngine) *FieldMasker {
	return &FieldMasker{pool: pool, engine: engine}
}

// FieldPermission represents a single field permission entry.
type FieldPermission struct {
	ID             uuid.UUID `json:"id"`
	OrgID          uuid.UUID `json:"org_id"`
	AccessPolicyID uuid.UUID `json:"access_policy_id"`
	ResourceType   string    `json:"resource_type"`
	FieldName      string    `json:"field_name"`
	Permission     string    `json:"permission"`
	MaskPattern    string    `json:"mask_pattern,omitempty"`
	CreatedAt      string    `json:"created_at"`
}

// GetFieldPermissions returns a map of field_name -> permission_type for the given
// user and resource type. It only considers policies assigned to the user that have
// associated field_level_permissions entries.
func (fm *FieldMasker) GetFieldPermissions(ctx context.Context, orgID, userID uuid.UUID, roles []string, resourceType string) (map[string]string, error) {
	// Get all applicable policies for the user
	policies, err := fm.engine.GetUserPolicies(ctx, orgID, userID, roles)
	if err != nil {
		return nil, fmt.Errorf("get user policies for field masking: %w", err)
	}

	if len(policies) == 0 {
		return make(map[string]string), nil
	}

	// Collect policy IDs
	policyIDs := make([]uuid.UUID, 0, len(policies))
	for _, p := range policies {
		policyIDs = append(policyIDs, p.ID)
	}

	// Fetch field-level permissions for these policies and resource type
	rows, err := fm.pool.Query(ctx, `
		SELECT field_name, permission::text, COALESCE(mask_pattern, '')
		FROM field_level_permissions
		WHERE org_id = $1
		  AND resource_type = $2
		  AND access_policy_id = ANY($3)
		ORDER BY field_name`, orgID, resourceType, policyIDs)
	if err != nil {
		return nil, fmt.Errorf("query field permissions: %w", err)
	}
	defer rows.Close()

	permissions := make(map[string]string)
	patterns := make(map[string]string)
	for rows.Next() {
		var fieldName, permission, pattern string
		if err := rows.Scan(&fieldName, &permission, &pattern); err != nil {
			return nil, fmt.Errorf("scan field permission: %w", err)
		}
		// Most restrictive wins: hidden > masked > visible
		existing, exists := permissions[fieldName]
		if !exists || moreRestrictive(permission, existing) {
			permissions[fieldName] = permission
			patterns[fieldName] = pattern
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Store patterns in a context-safe way by encoding into the permission string
	// The caller can use GetFieldPermissionsDetailed for pattern info
	return permissions, nil
}

// GetFieldPermissionsDetailed returns field permissions with their mask patterns.
func (fm *FieldMasker) GetFieldPermissionsDetailed(ctx context.Context, orgID, userID uuid.UUID, roles []string, resourceType string) ([]FieldPermission, error) {
	policies, err := fm.engine.GetUserPolicies(ctx, orgID, userID, roles)
	if err != nil {
		return nil, err
	}

	if len(policies) == 0 {
		return nil, nil
	}

	policyIDs := make([]uuid.UUID, 0, len(policies))
	for _, p := range policies {
		policyIDs = append(policyIDs, p.ID)
	}

	rows, err := fm.pool.Query(ctx, `
		SELECT id, org_id, access_policy_id, resource_type, field_name,
			permission::text, COALESCE(mask_pattern,''), created_at::text
		FROM field_level_permissions
		WHERE org_id = $1
		  AND resource_type = $2
		  AND access_policy_id = ANY($3)
		ORDER BY field_name`, orgID, resourceType, policyIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []FieldPermission
	for rows.Next() {
		var fp FieldPermission
		if err := rows.Scan(&fp.ID, &fp.OrgID, &fp.AccessPolicyID, &fp.ResourceType,
			&fp.FieldName, &fp.Permission, &fp.MaskPattern, &fp.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, fp)
	}
	return perms, rows.Err()
}

// MaskFields applies field-level permissions to a data map.
// For "hidden" fields: the key is removed entirely.
// For "masked" fields: the value is replaced with a masked version.
// For "visible" fields (or fields not in the map): no change.
func (fm *FieldMasker) MaskFields(data map[string]interface{}, permissions map[string]string) map[string]interface{} {
	if len(permissions) == 0 {
		return data
	}

	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		perm, exists := permissions[k]
		if !exists || perm == "visible" {
			result[k] = v
			continue
		}

		switch perm {
		case "hidden":
			// Field is completely removed from the result
			continue
		case "masked":
			result[k] = MaskValue(v, "")
		default:
			result[k] = v
		}
	}
	return result
}

// MaskFieldsJSON applies field permissions to a JSON-serialisable structure.
// It marshals to a map, applies masking, then returns the masked map.
func (fm *FieldMasker) MaskFieldsJSON(data interface{}, permissions map[string]string) (map[string]interface{}, error) {
	if len(permissions) == 0 {
		return nil, nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal for masking: %w", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, fmt.Errorf("unmarshal for masking: %w", err)
	}

	return fm.MaskFields(m, permissions), nil
}

// MaskValue applies a mask pattern to a value. If pattern is empty, a default
// name-style mask is applied (e.g., "John Smith" -> "J*** S****").
// For numeric values, it returns "***,***".
func MaskValue(value interface{}, pattern string) string {
	if value == nil {
		return "****"
	}

	if pattern != "" {
		return pattern
	}

	str := fmt.Sprintf("%v", value)
	if str == "" {
		return "****"
	}

	// Check if numeric
	isNumeric := true
	for _, ch := range str {
		if ch != '.' && ch != ',' && ch != '-' && (ch < '0' || ch > '9') {
			isNumeric = false
			break
		}
	}
	if isNumeric {
		return "***,***"
	}

	// Apply name-style masking: preserve first letter of each word, mask the rest
	words := strings.Fields(str)
	masked := make([]string, 0, len(words))
	for _, word := range words {
		if utf8.RuneCountInString(word) == 0 {
			continue
		}
		runes := []rune(word)
		first := string(runes[0])
		stars := strings.Repeat("*", len(runes)-1)
		masked = append(masked, first+stars)
	}

	if len(masked) == 0 {
		return "****"
	}
	return strings.Join(masked, " ")
}

// moreRestrictive returns true if perm is more restrictive than existing.
// Order: hidden > masked > visible
func moreRestrictive(perm, existing string) bool {
	order := map[string]int{
		"visible": 0,
		"masked":  1,
		"hidden":  2,
	}
	p, ok1 := order[perm]
	e, ok2 := order[existing]
	if !ok1 || !ok2 {
		log.Warn().Str("perm", perm).Str("existing", existing).Msg("unknown field permission level")
		return false
	}
	return p > e
}
