package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ============================================================
// ABAC FILTER — SQL WHERE clause generator
// ============================================================

// ABACFilter generates SQL WHERE clause fragments from ABAC policies so that
// list queries only return resources the user is permitted to see. This avoids
// loading every resource and checking access one by one.
type ABACFilter struct {
	engine *ABACEngine
}

// NewABACFilter creates a new ABACFilter backed by the given engine.
func NewABACFilter(engine *ABACEngine) *ABACFilter {
	return &ABACFilter{engine: engine}
}

// GenerateFilter builds a SQL WHERE clause and parameter list that restricts
// query results to resources the user is allowed to access for the given action.
//
// It inspects all matching "allow" policies and converts their resource_conditions
// into SQL fragments. Deny policies are not converted into SQL (they are evaluated
// at the row level by the engine) — this filter is a performance optimisation.
//
// Returns: (whereClause string, params []interface{}, err error)
// The whereClause includes a leading "AND" so it can be appended to an existing query.
// If no conditions apply (e.g. the user has a wildcard allow), it returns ("", nil, nil).
func (f *ABACFilter) GenerateFilter(ctx context.Context, orgID, userID uuid.UUID, roles []string, resourceType, action string) (string, []interface{}, error) {
	policies, err := f.engine.GetUserPolicies(ctx, orgID, userID, roles)
	if err != nil {
		return "", nil, fmt.Errorf("fetch policies for filter: %w", err)
	}

	var allowPolicies []PolicyRecord
	for _, p := range policies {
		if !p.IsActive || p.Effect != "allow" {
			continue
		}
		if p.ResourceType != "*" && !strings.EqualFold(p.ResourceType, resourceType) {
			continue
		}
		if !actionMatches(p.Actions, action) {
			continue
		}
		allowPolicies = append(allowPolicies, p)
	}

	if len(allowPolicies) == 0 {
		// No allow policies found — deny everything via impossible condition
		return " AND 1=0", nil, nil
	}

	// Check if any allow policy has no resource conditions (unrestricted access)
	for _, p := range allowPolicies {
		if len(p.ResourceConditions) == 0 {
			// Unrestricted — no additional SQL filter needed
			return "", nil, nil
		}
	}

	// Convert resource conditions from each allow policy into SQL OR groups
	var clauses []string
	var params []interface{}
	paramIdx := 1

	for _, p := range allowPolicies {
		policyFragments, policyParams, nextIdx := conditionsToSQL(p.ResourceConditions, userID, paramIdx)
		if len(policyFragments) > 0 {
			joined := strings.Join(policyFragments, " AND ")
			clauses = append(clauses, "("+joined+")")
			params = append(params, policyParams...)
			paramIdx = nextIdx
		}
	}

	if len(clauses) == 0 {
		return "", nil, nil
	}

	// Multiple allow policies are combined with OR — if any policy allows, show it
	whereClause := " AND (" + strings.Join(clauses, " OR ") + ")"
	return whereClause, params, nil
}

// conditionsToSQL converts a slice of Condition into SQL fragments and parameters.
// It returns the fragments, params, and the next available parameter index.
func conditionsToSQL(conditions []Condition, userID uuid.UUID, startIdx int) ([]string, []interface{}, int) {
	var fragments []string
	var params []interface{}
	idx := startIdx

	for _, c := range conditions {
		frag, fragParams, nextIdx := singleConditionToSQL(c, userID, idx)
		if frag != "" {
			fragments = append(fragments, frag)
			params = append(params, fragParams...)
			idx = nextIdx
		}
	}
	return fragments, params, idx
}

// singleConditionToSQL converts one Condition to a SQL fragment.
func singleConditionToSQL(c Condition, userID uuid.UUID, idx int) (string, []interface{}, int) {
	col := sanitizeColumnName(c.Field)
	if col == "" {
		return "", nil, idx
	}

	switch c.Operator {
	case "equals":
		frag := fmt.Sprintf("%s = $%d", col, idx)
		return frag, []interface{}{c.Value}, idx + 1

	case "not_equals":
		frag := fmt.Sprintf("%s != $%d", col, idx)
		return frag, []interface{}{c.Value}, idx + 1

	case "in":
		values := toStringSlice(c.Value)
		if len(values) == 0 {
			return "", nil, idx
		}
		placeholders := make([]string, len(values))
		pars := make([]interface{}, len(values))
		for i, v := range values {
			placeholders[i] = fmt.Sprintf("$%d", idx+i)
			pars[i] = v
		}
		frag := fmt.Sprintf("%s IN (%s)", col, strings.Join(placeholders, ","))
		return frag, pars, idx + len(values)

	case "not_in":
		values := toStringSlice(c.Value)
		if len(values) == 0 {
			return "", nil, idx
		}
		placeholders := make([]string, len(values))
		pars := make([]interface{}, len(values))
		for i, v := range values {
			placeholders[i] = fmt.Sprintf("$%d", idx+i)
			pars[i] = v
		}
		frag := fmt.Sprintf("%s NOT IN (%s)", col, strings.Join(placeholders, ","))
		return frag, pars, idx + len(values)

	case "greater_than":
		frag := fmt.Sprintf("%s > $%d", col, idx)
		return frag, []interface{}{c.Value}, idx + 1

	case "less_than":
		frag := fmt.Sprintf("%s < $%d", col, idx)
		return frag, []interface{}{c.Value}, idx + 1

	case "equals_subject":
		// The resource field must equal the requesting user's ID
		fieldName, ok := c.Value.(string)
		if !ok || fieldName != "user_id" {
			log.Warn().Str("field", c.Field).Msg("ABAC filter: unsupported equals_subject field")
			return "", nil, idx
		}
		frag := fmt.Sprintf("%s = $%d", col, idx)
		return frag, []interface{}{userID}, idx + 1

	case "between":
		bounds, ok := c.Value.([]interface{})
		if !ok || len(bounds) != 2 {
			return "", nil, idx
		}
		frag := fmt.Sprintf("%s BETWEEN $%d AND $%d", col, idx, idx+1)
		return frag, []interface{}{bounds[0], bounds[1]}, idx + 2

	default:
		log.Warn().Str("operator", c.Operator).Msg("ABAC filter: unsupported operator for SQL conversion")
		return "", nil, idx
	}
}

// sanitizeColumnName allows only alphanumeric characters and underscores
// to prevent SQL injection through field names.
func sanitizeColumnName(name string) string {
	var sb strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			sb.WriteRune(ch)
		}
	}
	result := sb.String()
	if result == "" {
		return ""
	}
	// Prevent reserved words from being used as column names
	lower := strings.ToLower(result)
	reserved := map[string]bool{
		"select": true, "insert": true, "update": true, "delete": true,
		"drop": true, "alter": true, "create": true, "grant": true,
		"revoke": true, "truncate": true,
	}
	if reserved[lower] {
		return ""
	}
	return result
}
