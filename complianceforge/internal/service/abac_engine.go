// Package service contains the Advanced ABAC (Attribute-Based Access Control) Engine
// for the ComplianceForge GRC platform. It evaluates access requests against JSONB-based
// policies using subject, resource, and environment attribute conditions with
// deny-overrides combining.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// TYPES
// ============================================================

// AccessRequest represents an incoming authorization request.
type AccessRequest struct {
	SubjectID    uuid.UUID  `json:"subject_id"`
	OrgID        uuid.UUID  `json:"org_id"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
}

// AccessDecision is the result of an ABAC policy evaluation.
type AccessDecision struct {
	Effect          string     `json:"effect"`
	PolicyID        *uuid.UUID `json:"policy_id,omitempty"`
	PolicyName      string     `json:"policy_name,omitempty"`
	Reason          string     `json:"reason"`
	EvaluationTimeUS int       `json:"evaluation_time_us"`
}

// SubjectAttributes describes the requesting user.
type SubjectAttributes struct {
	UserID         uuid.UUID `json:"user_id"`
	OrgID          uuid.UUID `json:"org_id"`
	Roles          []string  `json:"roles"`
	Department     string    `json:"department,omitempty"`
	Region         string    `json:"region,omitempty"`
	ClearanceLevel string    `json:"clearance_level,omitempty"`
	MFAVerified    bool      `json:"mfa_verified"`
}

// ResourceAttributes holds dynamic resource metadata fetched at evaluation time.
type ResourceAttributes map[string]interface{}

// EnvironmentAttributes captures request-time context.
type EnvironmentAttributes struct {
	IP        string    `json:"ip,omitempty"`
	Time      time.Time `json:"time"`
	MFAStatus bool      `json:"mfa_status"`
	TimeHour  int       `json:"time_hour"`
}

// Condition is a single attribute match rule stored in policy JSONB arrays.
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// PolicyRecord represents a row from the access_policies table.
type PolicyRecord struct {
	ID                    uuid.UUID   `json:"id"`
	OrgID                 uuid.UUID   `json:"org_id"`
	Name                  string      `json:"name"`
	Description           string      `json:"description"`
	Priority              int         `json:"priority"`
	Effect                string      `json:"effect"`
	IsActive              bool        `json:"is_active"`
	SubjectConditions     []Condition `json:"subject_conditions"`
	ResourceType          string      `json:"resource_type"`
	ResourceConditions    []Condition `json:"resource_conditions"`
	Actions               []string    `json:"actions"`
	EnvironmentConditions []Condition `json:"environment_conditions"`
	ValidFrom             *time.Time  `json:"valid_from,omitempty"`
	ValidUntil            *time.Time  `json:"valid_until,omitempty"`
	CreatedBy             *uuid.UUID  `json:"created_by,omitempty"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
}

// PolicyAssignment represents a row from the access_policy_assignments table.
type PolicyAssignment struct {
	ID             uuid.UUID  `json:"id"`
	OrgID          uuid.UUID  `json:"org_id"`
	AccessPolicyID uuid.UUID  `json:"access_policy_id"`
	AssigneeType   string     `json:"assignee_type"`
	AssigneeID     *uuid.UUID `json:"assignee_id,omitempty"`
	ValidFrom      *time.Time `json:"valid_from,omitempty"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// AuditLogEntry represents a row from the access_audit_log table.
type AuditLogEntry struct {
	ID                    uuid.UUID              `json:"id"`
	OrgID                 uuid.UUID              `json:"org_id"`
	UserID                uuid.UUID              `json:"user_id"`
	Action                string                 `json:"action"`
	ResourceType          string                 `json:"resource_type"`
	ResourceID            *uuid.UUID             `json:"resource_id,omitempty"`
	Decision              string                 `json:"decision"`
	MatchedPolicyID       *uuid.UUID             `json:"matched_policy_id,omitempty"`
	EvaluationTimeUS      int                    `json:"evaluation_time_us"`
	SubjectAttributes     map[string]interface{} `json:"subject_attributes"`
	ResourceAttributes    map[string]interface{} `json:"resource_attributes"`
	EnvironmentAttributes map[string]interface{} `json:"environment_attributes"`
	CreatedAt             time.Time              `json:"created_at"`
}

// ============================================================
// ABAC ENGINE
// ============================================================

// ABACEngine is the core evaluation engine for attribute-based access control.
type ABACEngine struct {
	pool *pgxpool.Pool
}

// NewABACEngine creates a new ABACEngine backed by the provided connection pool.
func NewABACEngine(pool *pgxpool.Pool) *ABACEngine {
	return &ABACEngine{pool: pool}
}

// Evaluate is the primary entry point. It evaluates an AccessRequest against all
// applicable policies using deny-overrides combining:
//   1. Collect policies assigned to the subject (direct user, role, all_users).
//   2. Filter by resource_type, action, and temporal constraints.
//   3. Evaluate subject_conditions, resource_conditions, environment_conditions.
//   4. Apply deny-overrides: any matching deny => deny; else any allow => allow; else deny (default).
//   5. Log the decision in access_audit_log.
func (e *ABACEngine) Evaluate(ctx context.Context, req AccessRequest, subject SubjectAttributes, env EnvironmentAttributes) AccessDecision {
	start := time.Now()

	// Fetch all applicable policies for this subject
	policies, err := e.GetUserPolicies(ctx, req.OrgID, req.SubjectID, subject.Roles)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.SubjectID.String()).Msg("ABAC: failed to fetch user policies")
		elapsed := int(time.Since(start).Microseconds())
		decision := AccessDecision{Effect: "deny", Reason: "policy retrieval error", EvaluationTimeUS: elapsed}
		e.logDecisionAsync(ctx, req, decision, subject, nil, env)
		return decision
	}

	// Fetch resource attributes if a specific resource is targeted
	var resAttrs ResourceAttributes
	if req.ResourceID != nil {
		resAttrs, err = e.fetchResourceAttributes(ctx, req.OrgID, req.ResourceType, *req.ResourceID)
		if err != nil {
			log.Warn().Err(err).Str("resource_type", req.ResourceType).Msg("ABAC: failed to fetch resource attributes")
			resAttrs = make(ResourceAttributes)
		}
	} else {
		resAttrs = make(ResourceAttributes)
	}

	now := env.Time
	if now.IsZero() {
		now = time.Now().UTC()
	}
	env.TimeHour = now.Hour()

	// Build attribute maps for condition evaluation
	subjectMap := subjectToMap(subject)
	resourceMap := map[string]interface{}(resAttrs)
	envMap := envToMap(env)

	var matchedDeny *PolicyRecord
	var matchedAllow *PolicyRecord

	for i := range policies {
		p := &policies[i]

		// Skip inactive policies
		if !p.IsActive {
			continue
		}

		// Check temporal validity on the policy itself
		if p.ValidFrom != nil && now.Before(*p.ValidFrom) {
			continue
		}
		if p.ValidUntil != nil && now.After(*p.ValidUntil) {
			continue
		}

		// Check resource_type match (wildcard or exact)
		if p.ResourceType != "*" && !strings.EqualFold(p.ResourceType, req.ResourceType) {
			continue
		}

		// Check action match
		if !actionMatches(p.Actions, req.Action) {
			continue
		}

		// Evaluate subject conditions
		if len(p.SubjectConditions) > 0 {
			if !EvaluateConditions(p.SubjectConditions, subjectMap, subjectMap) {
				continue
			}
		}

		// Evaluate resource conditions
		if len(p.ResourceConditions) > 0 {
			if !EvaluateConditions(p.ResourceConditions, resourceMap, subjectMap) {
				continue
			}
		}

		// Evaluate environment conditions
		if len(p.EnvironmentConditions) > 0 {
			if !EvaluateConditions(p.EnvironmentConditions, envMap, subjectMap) {
				continue
			}
		}

		// Policy matched — record by effect
		if p.Effect == "deny" {
			if matchedDeny == nil || p.Priority < matchedDeny.Priority {
				matchedDeny = p
			}
		} else if p.Effect == "allow" {
			if matchedAllow == nil || p.Priority < matchedAllow.Priority {
				matchedAllow = p
			}
		}
	}

	elapsed := int(time.Since(start).Microseconds())

	// Deny-overrides combining algorithm
	var decision AccessDecision
	if matchedDeny != nil {
		decision = AccessDecision{
			Effect:           "deny",
			PolicyID:         &matchedDeny.ID,
			PolicyName:       matchedDeny.Name,
			Reason:           fmt.Sprintf("Denied by policy: %s", matchedDeny.Name),
			EvaluationTimeUS: elapsed,
		}
	} else if matchedAllow != nil {
		decision = AccessDecision{
			Effect:           "allow",
			PolicyID:         &matchedAllow.ID,
			PolicyName:       matchedAllow.Name,
			Reason:           fmt.Sprintf("Allowed by policy: %s", matchedAllow.Name),
			EvaluationTimeUS: elapsed,
		}
	} else {
		decision = AccessDecision{
			Effect:           "deny",
			Reason:           "No matching policy found — default deny",
			EvaluationTimeUS: elapsed,
		}
	}

	// Log decision asynchronously
	e.logDecisionAsync(ctx, req, decision, subject, resAttrs, env)

	return decision
}

// GetUserPolicies fetches all active access policies that are assigned to the given user,
// either directly (by user ID), by any of their roles, or via all_users assignments.
// Results are ordered by priority ascending (highest priority = lowest number).
func (e *ABACEngine) GetUserPolicies(ctx context.Context, orgID, userID uuid.UUID, roles []string) ([]PolicyRecord, error) {
	query := `
		SELECT DISTINCT ON (ap.id)
			ap.id, ap.org_id, ap.name, COALESCE(ap.description,''), ap.priority,
			ap.effect::text, ap.is_active,
			ap.subject_conditions, ap.resource_type, ap.resource_conditions,
			ap.actions, ap.environment_conditions,
			ap.valid_from, ap.valid_until, ap.created_by,
			ap.created_at, ap.updated_at
		FROM access_policies ap
		INNER JOIN access_policy_assignments apa ON apa.access_policy_id = ap.id AND apa.org_id = ap.org_id
		WHERE ap.org_id = $1
		  AND ap.is_active = true
		  AND (
		      (apa.assignee_type = 'user' AND apa.assignee_id = $2)
		      OR (apa.assignee_type = 'role' AND apa.assignee_id IS NULL)
		      OR (apa.assignee_type = 'all_users')
		  )
		  AND (apa.valid_from IS NULL OR apa.valid_from <= now())
		  AND (apa.valid_until IS NULL OR apa.valid_until >= now())
		ORDER BY ap.id, ap.priority ASC`

	rows, err := e.pool.Query(ctx, query, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("query user policies: %w", err)
	}
	defer rows.Close()

	var policies []PolicyRecord
	for rows.Next() {
		var p PolicyRecord
		var subjJSON, resJSON, envJSON []byte
		err := rows.Scan(
			&p.ID, &p.OrgID, &p.Name, &p.Description, &p.Priority,
			&p.Effect, &p.IsActive,
			&subjJSON, &p.ResourceType, &resJSON,
			&p.Actions, &envJSON,
			&p.ValidFrom, &p.ValidUntil, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan policy row: %w", err)
		}
		if err := json.Unmarshal(subjJSON, &p.SubjectConditions); err != nil {
			p.SubjectConditions = nil
		}
		if err := json.Unmarshal(resJSON, &p.ResourceConditions); err != nil {
			p.ResourceConditions = nil
		}
		if err := json.Unmarshal(envJSON, &p.EnvironmentConditions); err != nil {
			p.EnvironmentConditions = nil
		}
		policies = append(policies, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate policy rows: %w", err)
	}
	return policies, nil
}

// EvaluateConditions checks all conditions against the given attribute map.
// All conditions must match (AND semantics). The subjectMap is passed separately
// to support the equals_subject operator.
func EvaluateConditions(conditions []Condition, attributes map[string]interface{}, subjectMap map[string]interface{}) bool {
	for _, c := range conditions {
		if !evaluateSingleCondition(c, attributes, subjectMap) {
			return false
		}
	}
	return true
}

// evaluateSingleCondition dispatches to the appropriate operator implementation.
func evaluateSingleCondition(c Condition, attrs map[string]interface{}, subjectMap map[string]interface{}) bool {
	attrVal, exists := attrs[c.Field]

	switch c.Operator {
	case "equals":
		if !exists {
			return false
		}
		return compareEquals(attrVal, c.Value)

	case "not_equals":
		if !exists {
			return true // field absent means it does not equal the value
		}
		return !compareEquals(attrVal, c.Value)

	case "in":
		if !exists {
			return false
		}
		return valueInList(attrVal, c.Value)

	case "not_in":
		if !exists {
			return true
		}
		return !valueInList(attrVal, c.Value)

	case "contains_any":
		if !exists {
			return false
		}
		return containsAny(attrVal, c.Value)

	case "greater_than":
		if !exists {
			return false
		}
		return compareGreaterThan(attrVal, c.Value)

	case "less_than":
		if !exists {
			return false
		}
		return compareLessThan(attrVal, c.Value)

	case "in_cidr":
		if !exists {
			return false
		}
		return checkCIDR(attrVal, c.Value)

	case "between":
		if !exists {
			return false
		}
		return checkBetween(attrVal, c.Value)

	case "equals_subject":
		// value specifies which subject attribute field to compare against
		fieldName, ok := c.Value.(string)
		if !ok {
			return false
		}
		subjectVal, subExists := subjectMap[fieldName]
		if !subExists || !exists {
			return false
		}
		return compareEquals(attrVal, subjectVal)

	default:
		log.Warn().Str("operator", c.Operator).Msg("ABAC: unknown operator")
		return false
	}
}

// ============================================================
// OPERATOR IMPLEMENTATIONS
// ============================================================

// compareEquals performs a type-flexible equality comparison.
func compareEquals(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// valueInList checks if a scalar value exists in a list value.
func valueInList(attrVal, listVal interface{}) bool {
	list := toStringSlice(listVal)
	target := fmt.Sprintf("%v", attrVal)
	for _, item := range list {
		if item == target {
			return true
		}
	}
	return false
}

// containsAny checks if an attribute (which should be a slice) contains any of the given values.
func containsAny(attrVal, condVal interface{}) bool {
	attrList := toStringSlice(attrVal)
	condList := toStringSlice(condVal)
	for _, a := range attrList {
		for _, c := range condList {
			if a == c {
				return true
			}
		}
	}
	return false
}

// compareGreaterThan compares numeric values.
func compareGreaterThan(a, b interface{}) bool {
	af := toFloat64(a)
	bf := toFloat64(b)
	return af > bf
}

// compareLessThan compares numeric values.
func compareLessThan(a, b interface{}) bool {
	af := toFloat64(a)
	bf := toFloat64(b)
	return af < bf
}

// checkCIDR checks if an IP address string is within a CIDR range.
func checkCIDR(attrVal, cidrVal interface{}) bool {
	ipStr, ok := attrVal.(string)
	if !ok {
		return false
	}
	cidrStr, ok := cidrVal.(string)
	if !ok {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	_, network, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false
	}
	return network.Contains(ip)
}

// checkBetween checks if a value falls between two bounds [low, high].
// Supports both numeric and time-based comparisons.
func checkBetween(attrVal, boundsVal interface{}) bool {
	bounds, ok := boundsVal.([]interface{})
	if !ok || len(bounds) != 2 {
		return false
	}

	// Try time-based comparison first
	if timeStr, ok := attrVal.(string); ok {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			lowStr, ok1 := bounds[0].(string)
			highStr, ok2 := bounds[1].(string)
			if ok1 && ok2 {
				low, err1 := time.Parse(time.RFC3339, lowStr)
				high, err2 := time.Parse(time.RFC3339, highStr)
				if err1 == nil && err2 == nil {
					return !t.Before(low) && !t.After(high)
				}
			}
		}
	}

	// Numeric comparison
	val := toFloat64(attrVal)
	low := toFloat64(bounds[0])
	high := toFloat64(bounds[1])
	return val >= low && val <= high
}

// ============================================================
// HELPERS
// ============================================================

// actionMatches checks if the requested action is in the policy's action list.
// A wildcard "*" in the policy matches any action.
func actionMatches(policyActions []string, requestedAction string) bool {
	for _, a := range policyActions {
		if a == "*" || strings.EqualFold(a, requestedAction) {
			return true
		}
	}
	return false
}

// subjectToMap converts SubjectAttributes to a generic map for condition evaluation.
func subjectToMap(s SubjectAttributes) map[string]interface{} {
	m := map[string]interface{}{
		"user_id":         s.UserID.String(),
		"org_id":          s.OrgID.String(),
		"roles":           s.Roles,
		"mfa_verified":    s.MFAVerified,
		"department":      s.Department,
		"region":          s.Region,
		"clearance_level": s.ClearanceLevel,
	}
	return m
}

// envToMap converts EnvironmentAttributes to a generic map.
func envToMap(e EnvironmentAttributes) map[string]interface{} {
	m := map[string]interface{}{
		"ip":           e.IP,
		"time":         e.Time.Format(time.RFC3339),
		"mfa_verified": e.MFAStatus,
		"mfa_status":   e.MFAStatus,
		"time_hour":    e.TimeHour,
	}
	return m
}

// toStringSlice converts various types to a string slice.
func toStringSlice(v interface{}) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result
	case string:
		return []string{val}
	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}

// toFloat64 converts a value to float64 for numeric comparisons.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

// ============================================================
// RESOURCE ATTRIBUTE FETCHING
// ============================================================

// fetchResourceAttributes dynamically fetches attributes for a specific resource.
// It queries the appropriate table based on resource_type and returns a map of attributes.
func (e *ABACEngine) fetchResourceAttributes(ctx context.Context, orgID uuid.UUID, resourceType string, resourceID uuid.UUID) (ResourceAttributes, error) {
	attrs := make(ResourceAttributes)

	var query string
	switch resourceType {
	case "control":
		query = `SELECT id, COALESCE(owner_user_id::text,''), COALESCE(status::text,''), COALESCE(classification,'')
			FROM organization_controls WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	case "risk":
		query = `SELECT id, COALESCE(owner_user_id::text,''), COALESCE(status::text,''), COALESCE(financial_impact_eur,0)
			FROM risks WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	case "incident":
		query = `SELECT id, COALESCE(is_data_breach, false), COALESCE(severity::text,''), COALESCE(status::text,''), COALESCE(classification,'')
			FROM incidents WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	case "policy":
		query = `SELECT id, COALESCE(owner_user_id::text,''), COALESCE(status::text,''), COALESCE(classification,'')
			FROM policies WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	case "dsr_request":
		query = `SELECT id, COALESCE(request_type::text,''), COALESCE(status::text,''), COALESCE(assigned_to::text,'')
			FROM dsr_requests WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	case "vendor":
		query = `SELECT id, COALESCE(status::text,''), COALESCE(risk_tier::text,'')
			FROM vendors WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`
	default:
		return attrs, nil
	}

	rows, err := e.pool.Query(ctx, query, resourceID, orgID)
	if err != nil {
		return attrs, fmt.Errorf("fetch resource attrs: %w", err)
	}
	defer rows.Close()

	cols := rows.FieldDescriptions()
	if rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return attrs, fmt.Errorf("read resource values: %w", err)
		}
		for i, col := range cols {
			if i < len(values) {
				attrs[string(col.Name)] = values[i]
			}
		}
	}
	return attrs, nil
}

// ============================================================
// DECISION LOGGING
// ============================================================

// logDecisionAsync logs the access decision to the audit log asynchronously.
func (e *ABACEngine) logDecisionAsync(ctx context.Context, req AccessRequest, decision AccessDecision, subject SubjectAttributes, resAttrs ResourceAttributes, env EnvironmentAttributes) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := e.LogDecision(bgCtx, req, decision, subject, resAttrs, env); err != nil {
			log.Error().Err(err).Str("user_id", req.SubjectID.String()).Str("action", req.Action).Msg("ABAC: failed to log decision")
		}
	}()
}

// LogDecision writes an access decision to the access_audit_log table.
func (e *ABACEngine) LogDecision(ctx context.Context, req AccessRequest, decision AccessDecision, subject SubjectAttributes, resAttrs ResourceAttributes, env EnvironmentAttributes) error {
	subjJSON, _ := json.Marshal(subjectToMap(subject))
	resJSON, _ := json.Marshal(resAttrs)
	envJSON, _ := json.Marshal(envToMap(env))

	_, err := e.pool.Exec(ctx, `
		INSERT INTO access_audit_log (
			id, org_id, user_id, action, resource_type, resource_id,
			decision, matched_policy_id, evaluation_time_us,
			subject_attributes, resource_attributes, environment_attributes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::access_decision_type, $8, $9, $10, $11, $12, now())`,
		uuid.New(), req.OrgID, req.SubjectID, req.Action, req.ResourceType, req.ResourceID,
		decision.Effect, decision.PolicyID, decision.EvaluationTimeUS,
		subjJSON, resJSON, envJSON,
	)
	return err
}

// ============================================================
// POLICY CRUD OPERATIONS (used by handler)
// ============================================================

// ListPolicies returns all access policies for an organisation with pagination.
func (e *ABACEngine) ListPolicies(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]PolicyRecord, int64, error) {
	var total int64
	err := e.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM access_policies WHERE org_id = $1`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count policies: %w", err)
	}

	rows, err := e.pool.Query(ctx, `
		SELECT id, org_id, name, COALESCE(description,''), priority,
			effect::text, is_active,
			subject_conditions, resource_type, resource_conditions,
			actions, environment_conditions,
			valid_from, valid_until, created_by,
			created_at, updated_at
		FROM access_policies
		WHERE org_id = $1
		ORDER BY priority ASC, name ASC
		LIMIT $2 OFFSET $3`,
		orgID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()

	policies, err := scanPolicies(rows)
	if err != nil {
		return nil, 0, err
	}
	return policies, total, nil
}

// GetPolicyByID retrieves a single access policy by ID.
func (e *ABACEngine) GetPolicyByID(ctx context.Context, orgID, policyID uuid.UUID) (*PolicyRecord, error) {
	row := e.pool.QueryRow(ctx, `
		SELECT id, org_id, name, COALESCE(description,''), priority,
			effect::text, is_active,
			subject_conditions, resource_type, resource_conditions,
			actions, environment_conditions,
			valid_from, valid_until, created_by,
			created_at, updated_at
		FROM access_policies
		WHERE id = $1 AND org_id = $2`, policyID, orgID)

	p, err := scanSinglePolicy(row)
	if err != nil {
		return nil, fmt.Errorf("get policy: %w", err)
	}
	return p, nil
}

// CreatePolicy inserts a new access policy.
func (e *ABACEngine) CreatePolicy(ctx context.Context, p *PolicyRecord) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	subjJSON, _ := json.Marshal(p.SubjectConditions)
	resJSON, _ := json.Marshal(p.ResourceConditions)
	envJSON, _ := json.Marshal(p.EnvironmentConditions)

	_, err := e.pool.Exec(ctx, `
		INSERT INTO access_policies (
			id, org_id, name, description, priority, effect, is_active,
			subject_conditions, resource_type, resource_conditions,
			actions, environment_conditions,
			valid_from, valid_until, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6::access_policy_effect,$7,$8,$9,$10,$11,$12,$13,$14,$15,now(),now())`,
		p.ID, p.OrgID, p.Name, p.Description, p.Priority,
		p.Effect, p.IsActive,
		subjJSON, p.ResourceType, resJSON,
		p.Actions, envJSON,
		p.ValidFrom, p.ValidUntil, p.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("create policy: %w", err)
	}
	return nil
}

// UpdatePolicy modifies an existing access policy.
func (e *ABACEngine) UpdatePolicy(ctx context.Context, p *PolicyRecord) error {
	subjJSON, _ := json.Marshal(p.SubjectConditions)
	resJSON, _ := json.Marshal(p.ResourceConditions)
	envJSON, _ := json.Marshal(p.EnvironmentConditions)

	tag, err := e.pool.Exec(ctx, `
		UPDATE access_policies SET
			name=$1, description=$2, priority=$3, effect=$4::access_policy_effect, is_active=$5,
			subject_conditions=$6, resource_type=$7, resource_conditions=$8,
			actions=$9, environment_conditions=$10,
			valid_from=$11, valid_until=$12
		WHERE id=$13 AND org_id=$14`,
		p.Name, p.Description, p.Priority, p.Effect, p.IsActive,
		subjJSON, p.ResourceType, resJSON,
		p.Actions, envJSON,
		p.ValidFrom, p.ValidUntil,
		p.ID, p.OrgID,
	)
	if err != nil {
		return fmt.Errorf("update policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("policy not found")
	}
	return nil
}

// DeletePolicy removes an access policy and its assignments.
func (e *ABACEngine) DeletePolicy(ctx context.Context, orgID, policyID uuid.UUID) error {
	tag, err := e.pool.Exec(ctx,
		`DELETE FROM access_policies WHERE id=$1 AND org_id=$2`, policyID, orgID)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("policy not found")
	}
	return nil
}

// ============================================================
// ASSIGNMENT CRUD
// ============================================================

// CreateAssignment adds a policy assignment.
func (e *ABACEngine) CreateAssignment(ctx context.Context, a *PolicyAssignment) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	_, err := e.pool.Exec(ctx, `
		INSERT INTO access_policy_assignments (
			id, org_id, access_policy_id, assignee_type, assignee_id,
			valid_from, valid_until, created_by, created_at
		) VALUES ($1,$2,$3,$4::assignee_type,$5,$6,$7,$8,now())`,
		a.ID, a.OrgID, a.AccessPolicyID, a.AssigneeType, a.AssigneeID,
		a.ValidFrom, a.ValidUntil, a.CreatedBy,
	)
	return err
}

// DeleteAssignment removes a policy assignment.
func (e *ABACEngine) DeleteAssignment(ctx context.Context, orgID, assignmentID uuid.UUID) error {
	tag, err := e.pool.Exec(ctx,
		`DELETE FROM access_policy_assignments WHERE id=$1 AND org_id=$2`, assignmentID, orgID)
	if err != nil {
		return fmt.Errorf("delete assignment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("assignment not found")
	}
	return nil
}

// ListAssignments returns all assignments for a given policy.
func (e *ABACEngine) ListAssignments(ctx context.Context, orgID, policyID uuid.UUID) ([]PolicyAssignment, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, org_id, access_policy_id, assignee_type::text, assignee_id,
			valid_from, valid_until, created_by, created_at
		FROM access_policy_assignments
		WHERE org_id=$1 AND access_policy_id=$2
		ORDER BY created_at DESC`, orgID, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []PolicyAssignment
	for rows.Next() {
		var a PolicyAssignment
		if err := rows.Scan(&a.ID, &a.OrgID, &a.AccessPolicyID, &a.AssigneeType, &a.AssigneeID,
			&a.ValidFrom, &a.ValidUntil, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

// ============================================================
// AUDIT LOG QUERIES
// ============================================================

// ListAuditLog returns paginated access decisions for an organisation.
func (e *ABACEngine) ListAuditLog(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]AuditLogEntry, int64, error) {
	var total int64
	if err := e.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM access_audit_log WHERE org_id=$1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := e.pool.Query(ctx, `
		SELECT id, org_id, user_id, action, resource_type, resource_id,
			decision::text, matched_policy_id, evaluation_time_us,
			subject_attributes, resource_attributes, environment_attributes, created_at
		FROM access_audit_log
		WHERE org_id=$1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, orgID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []AuditLogEntry
	for rows.Next() {
		var entry AuditLogEntry
		var subjJSON, resJSON, envJSON []byte
		if err := rows.Scan(
			&entry.ID, &entry.OrgID, &entry.UserID, &entry.Action,
			&entry.ResourceType, &entry.ResourceID,
			&entry.Decision, &entry.MatchedPolicyID, &entry.EvaluationTimeUS,
			&subjJSON, &resJSON, &envJSON, &entry.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(subjJSON, &entry.SubjectAttributes)
		json.Unmarshal(resJSON, &entry.ResourceAttributes)
		json.Unmarshal(envJSON, &entry.EnvironmentAttributes)
		entries = append(entries, entry)
	}
	return entries, total, rows.Err()
}

// GetMyPermissions returns a summary of what a user can do based on their active policies.
func (e *ABACEngine) GetMyPermissions(ctx context.Context, orgID, userID uuid.UUID, roles []string) ([]map[string]interface{}, error) {
	policies, err := e.GetUserPolicies(ctx, orgID, userID, roles)
	if err != nil {
		return nil, err
	}

	var perms []map[string]interface{}
	for _, p := range policies {
		if p.Effect == "allow" {
			perms = append(perms, map[string]interface{}{
				"policy_id":     p.ID,
				"policy_name":   p.Name,
				"resource_type": p.ResourceType,
				"actions":       p.Actions,
				"priority":      p.Priority,
				"conditions":    len(p.ResourceConditions) > 0,
			})
		}
	}
	return perms, nil
}

// ============================================================
// INTERNAL SCAN HELPERS
// ============================================================

func scanPolicies(rows pgx.Rows) ([]PolicyRecord, error) {
	var policies []PolicyRecord
	for rows.Next() {
		p, err := scanPolicyFromRows(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, *p)
	}
	return policies, rows.Err()
}

func scanPolicyFromRows(rows pgx.Rows) (*PolicyRecord, error) {
	var p PolicyRecord
	var subjJSON, resJSON, envJSON []byte
	err := rows.Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Description, &p.Priority,
		&p.Effect, &p.IsActive,
		&subjJSON, &p.ResourceType, &resJSON,
		&p.Actions, &envJSON,
		&p.ValidFrom, &p.ValidUntil, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan policy: %w", err)
	}
	json.Unmarshal(subjJSON, &p.SubjectConditions)
	json.Unmarshal(resJSON, &p.ResourceConditions)
	json.Unmarshal(envJSON, &p.EnvironmentConditions)
	return &p, nil
}

func scanSinglePolicy(row pgx.Row) (*PolicyRecord, error) {
	var p PolicyRecord
	var subjJSON, resJSON, envJSON []byte
	err := row.Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Description, &p.Priority,
		&p.Effect, &p.IsActive,
		&subjJSON, &p.ResourceType, &resJSON,
		&p.Actions, &envJSON,
		&p.ValidFrom, &p.ValidUntil, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(subjJSON, &p.SubjectConditions)
	json.Unmarshal(resJSON, &p.ResourceConditions)
	json.Unmarshal(envJSON, &p.EnvironmentConditions)
	return &p, nil
}

// ============================================================
// MIDDLEWARE ADAPTER
// ============================================================

// NewEvalFunc returns a function matching middleware.ABACEvalFunc that bridges
// the HTTP middleware layer to the ABAC engine without creating an import cycle.
// Wire it at application startup:
//
//	engine := service.NewABACEngine(pool)
//	evalFn := engine.NewEvalFunc()
//	r.With(middleware.ABAC(evalFn, "read", "risk")).Get("/", handler.ListRisks)
func (e *ABACEngine) NewEvalFunc() func(
	r *http.Request,
	userID, orgID uuid.UUID,
	roles []string,
	action, resourceType string,
	ip string,
	mfaVerified bool,
) (bool, string) {
	return func(
		r *http.Request,
		userID, orgID uuid.UUID,
		roles []string,
		action, resourceType string,
		ip string,
		mfaVerified bool,
	) (bool, string) {
		ctx := r.Context()

		req := AccessRequest{
			SubjectID:    userID,
			OrgID:        orgID,
			Action:       action,
			ResourceType: resourceType,
		}

		subject := SubjectAttributes{
			UserID:      userID,
			OrgID:       orgID,
			Roles:       roles,
			MFAVerified: mfaVerified,
		}

		now := time.Now().UTC()
		env := EnvironmentAttributes{
			IP:        ip,
			Time:      now,
			MFAStatus: mfaVerified,
			TimeHour:  now.Hour(),
		}

		decision := e.Evaluate(ctx, req, subject, env)
		return decision.Effect == "allow", decision.Reason
	}
}
