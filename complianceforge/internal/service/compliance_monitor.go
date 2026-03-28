package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/pkg/queue"
)

// ============================================================
// COMPLIANCE MONITOR
// ============================================================

// ComplianceMonitor runs periodic checks against various compliance
// dimensions and records results. It detects when controls, evidence,
// KRIs, policies, vendors, or scores breach defined thresholds.
type ComplianceMonitor struct {
	pool  *pgxpool.Pool
	queue *queue.Queue
}

// NewComplianceMonitor creates a new ComplianceMonitor.
func NewComplianceMonitor(pool *pgxpool.Pool, q *queue.Queue) *ComplianceMonitor {
	return &ComplianceMonitor{pool: pool, queue: q}
}

// MonitorConfig represents a compliance monitor configuration.
type MonitorConfig struct {
	ID                 uuid.UUID       `json:"id"`
	OrganizationID     uuid.UUID       `json:"organization_id"`
	Name               string          `json:"name"`
	MonitorType        string          `json:"monitor_type"`
	TargetEntityType   string          `json:"target_entity_type"`
	TargetEntityID     *uuid.UUID      `json:"target_entity_id"`
	CheckFrequencyCron string          `json:"check_frequency_cron"`
	Conditions         json.RawMessage `json:"conditions"`
	AlertOnFailure     bool            `json:"alert_on_failure"`
	AlertSeverity      string          `json:"alert_severity"`
	IsActive           bool            `json:"is_active"`
	LastCheckAt        *time.Time      `json:"last_check_at"`
	LastCheckStatus    string          `json:"last_check_status"`
	ConsecutiveFailures int            `json:"consecutive_failures"`
	FailureSince       *time.Time      `json:"failure_since"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

// MonitorResult represents the outcome of a single monitor check.
type MonitorResult struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	MonitorID      uuid.UUID       `json:"monitor_id"`
	Status         string          `json:"status"`
	CheckTime      time.Time       `json:"check_time"`
	ResultData     json.RawMessage `json:"result_data"`
	Message        string          `json:"message"`
	CreatedAt      time.Time       `json:"created_at"`
}

// MonitorConditions holds the configurable conditions for a monitor check.
type MonitorConditions struct {
	MaxAgeDays        int     `json:"max_age_days"`
	MinScore          float64 `json:"min_score"`
	MinCompletionRate float64 `json:"min_completion_rate"`
	ScoreDropPercent  float64 `json:"score_drop_percent"`
}

// ============================================================
// MONITOR SCHEDULER
// ============================================================

// RunScheduledChecks runs all active monitors that are due for a check.
// Should be called periodically (e.g., every 5 minutes).
func (cm *ComplianceMonitor) RunScheduledChecks(ctx context.Context) error {
	rows, err := cm.pool.Query(ctx, `
		SELECT id, organization_id, name, monitor_type,
			target_entity_type, target_entity_id,
			conditions, alert_on_failure, alert_severity
		FROM compliance_monitors
		WHERE is_active = true
		ORDER BY last_check_at ASC NULLS FIRST
		LIMIT 50`)
	if err != nil {
		return fmt.Errorf("failed to query monitors: %w", err)
	}
	defer rows.Close()

	var monitors []MonitorConfig
	for rows.Next() {
		var m MonitorConfig
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.Name, &m.MonitorType,
			&m.TargetEntityType, &m.TargetEntityID,
			&m.Conditions, &m.AlertOnFailure, &m.AlertSeverity,
		); err != nil {
			log.Error().Err(err).Msg("Failed to scan monitor config")
			continue
		}
		monitors = append(monitors, m)
	}

	for _, m := range monitors {
		if err := cm.ExecuteCheck(ctx, m); err != nil {
			log.Error().Err(err).Str("monitor_id", m.ID.String()).Msg("Monitor check failed")
		}
	}

	return nil
}

// ExecuteCheck runs a single monitor check and records the result.
func (cm *ComplianceMonitor) ExecuteCheck(ctx context.Context, m MonitorConfig) error {
	var status string
	var message string
	var resultData json.RawMessage

	switch m.MonitorType {
	case "control_effectiveness":
		status, message, resultData = cm.CheckControlEffectiveness(ctx, m)
	case "evidence_freshness":
		status, message, resultData = cm.CheckEvidenceFreshness(ctx, m)
	case "kri_threshold":
		status, message, resultData = cm.CheckKRIThreshold(ctx, m)
	case "policy_attestation":
		status, message, resultData = cm.CheckPolicyAttestation(ctx, m)
	case "vendor_assessment":
		status, message, resultData = cm.CheckVendorAssessment(ctx, m)
	case "training_completion":
		status, message, resultData = cm.CheckComplianceScore(ctx, m)
	default:
		return fmt.Errorf("unsupported monitor type: %s", m.MonitorType)
	}

	checkTime := time.Now()

	// Insert result
	_, err := cm.pool.Exec(ctx, `
		INSERT INTO compliance_monitor_results (organization_id, monitor_id, status, check_time, result_data, message)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		m.OrganizationID, m.ID, status, checkTime, resultData, message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert monitor result")
	}

	// Update monitor status
	if status == "passing" {
		_, err = cm.pool.Exec(ctx, `
			UPDATE compliance_monitors
			SET last_check_at = $1, last_check_status = 'passing',
				consecutive_failures = 0, failure_since = NULL
			WHERE id = $2`, checkTime, m.ID)
	} else {
		_, err = cm.pool.Exec(ctx, `
			UPDATE compliance_monitors
			SET last_check_at = $1, last_check_status = 'failing',
				consecutive_failures = consecutive_failures + 1,
				failure_since = COALESCE(failure_since, $1)
			WHERE id = $2`, checkTime, m.ID)

		// Send alert if configured
		if m.AlertOnFailure {
			cm.sendAlert(ctx, m, message)
		}
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to update monitor status")
	}

	return nil
}

// ============================================================
// CONTROL EFFECTIVENESS MONITOR
// ============================================================

// CheckControlEffectiveness checks that control evidence is current and
// the control has a passing status (implemented or effective).
func (cm *ComplianceMonitor) CheckControlEffectiveness(ctx context.Context, m MonitorConfig) (string, string, json.RawMessage) {
	var conditions MonitorConditions
	if err := json.Unmarshal(m.Conditions, &conditions); err != nil {
		conditions.MaxAgeDays = 90
	}
	if conditions.MaxAgeDays <= 0 {
		conditions.MaxAgeDays = 90
	}

	// Check if the target control has recent evidence and a passing status
	var controlStatus string
	var latestEvidenceAge *int
	var evidenceCount int

	query := `
		SELECT
			ci.status::TEXT,
			EXTRACT(DAY FROM NOW() - MAX(ce.collected_at))::INT,
			COUNT(ce.id)
		FROM control_implementations ci
		LEFT JOIN control_evidence ce ON ce.control_implementation_id = ci.id
			AND ce.is_current = true AND ce.deleted_at IS NULL
		WHERE ci.organization_id = $1 AND ci.deleted_at IS NULL`

	args := []interface{}{m.OrganizationID}
	if m.TargetEntityID != nil {
		query += " AND ci.id = $2"
		args = append(args, *m.TargetEntityID)
	}
	query += " GROUP BY ci.status"

	err := cm.pool.QueryRow(ctx, query, args...).Scan(&controlStatus, &latestEvidenceAge, &evidenceCount)
	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return "failing", "Failed to query control status", data
	}

	passing := true
	var messages []string

	if controlStatus != "implemented" && controlStatus != "effective" {
		passing = false
		messages = append(messages, fmt.Sprintf("Control status is '%s', expected 'implemented' or 'effective'", controlStatus))
	}

	if evidenceCount == 0 {
		passing = false
		messages = append(messages, "No current evidence found")
	} else if latestEvidenceAge != nil && *latestEvidenceAge > conditions.MaxAgeDays {
		passing = false
		messages = append(messages, fmt.Sprintf("Latest evidence is %d days old (threshold: %d)", *latestEvidenceAge, conditions.MaxAgeDays))
	}

	status := "passing"
	msg := "Control effectiveness check passed"
	if !passing {
		status = "failing"
		msg = fmt.Sprintf("Control effectiveness check failed: %s", joinMessages(messages))
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"control_status":     controlStatus,
		"evidence_count":     evidenceCount,
		"latest_evidence_age": latestEvidenceAge,
		"max_age_days":       conditions.MaxAgeDays,
	})

	return status, msg, resultData
}

// ============================================================
// EVIDENCE FRESHNESS MONITOR
// ============================================================

// CheckEvidenceFreshness alerts when evidence is stale.
// Default staleness thresholds: API-collected = 24 hours, manual = 90 days.
func (cm *ComplianceMonitor) CheckEvidenceFreshness(ctx context.Context, m MonitorConfig) (string, string, json.RawMessage) {
	var conditions MonitorConditions
	if err := json.Unmarshal(m.Conditions, &conditions); err != nil {
		conditions.MaxAgeDays = 90
	}

	// Count stale evidence items
	apiMaxAge := 1   // 24 hours for API evidence
	manualMaxAge := conditions.MaxAgeDays
	if manualMaxAge <= 0 {
		manualMaxAge = 90
	}

	var staleAPICount, staleManualCount, totalCount int64
	err := cm.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE collection_method = 'automated_collection'
				AND collected_at < NOW() - INTERVAL '1 day' * $2),
			COUNT(*) FILTER (WHERE collection_method != 'automated_collection'
				AND collected_at < NOW() - INTERVAL '1 day' * $3),
			COUNT(*)
		FROM control_evidence
		WHERE organization_id = $1 AND is_current = true AND deleted_at IS NULL`,
		m.OrganizationID, apiMaxAge, manualMaxAge,
	).Scan(&staleAPICount, &staleManualCount, &totalCount)

	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return "failing", "Failed to query evidence freshness", data
	}

	totalStale := staleAPICount + staleManualCount
	status := "passing"
	msg := fmt.Sprintf("All %d evidence items are fresh", totalCount)
	if totalStale > 0 {
		status = "failing"
		msg = fmt.Sprintf("%d of %d evidence items are stale (API: %d, manual: %d)",
			totalStale, totalCount, staleAPICount, staleManualCount)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"total_evidence":      totalCount,
		"stale_api_evidence":  staleAPICount,
		"stale_manual_evidence": staleManualCount,
		"api_max_age_days":    apiMaxAge,
		"manual_max_age_days": manualMaxAge,
	})

	return status, msg, resultData
}

// ============================================================
// KRI THRESHOLD MONITOR
// ============================================================

// CheckKRIThreshold checks KRI values against their red thresholds.
func (cm *ComplianceMonitor) CheckKRIThreshold(ctx context.Context, m MonitorConfig) (string, string, json.RawMessage) {
	query := `
		SELECT id, name, current_value, threshold_red
		FROM risk_indicators
		WHERE organization_id = $1`

	args := []interface{}{m.OrganizationID}
	if m.TargetEntityID != nil {
		query += " AND id = $2"
		args = append(args, *m.TargetEntityID)
	}

	rows, err := cm.pool.Query(ctx, query, args...)
	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return "failing", "Failed to query KRI thresholds", data
	}
	defer rows.Close()

	type kriResult struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		CurrentValue float64   `json:"current_value"`
		ThresholdRed float64   `json:"threshold_red"`
		Breached     bool      `json:"breached"`
	}

	var results []kriResult
	breachedCount := 0

	for rows.Next() {
		var kr kriResult
		if err := rows.Scan(&kr.ID, &kr.Name, &kr.CurrentValue, &kr.ThresholdRed); err != nil {
			continue
		}
		// Value >= red threshold means breached
		kr.Breached = kr.CurrentValue >= kr.ThresholdRed
		if kr.Breached {
			breachedCount++
		}
		results = append(results, kr)
	}

	status := "passing"
	msg := fmt.Sprintf("All %d KRIs within thresholds", len(results))
	if breachedCount > 0 {
		status = "failing"
		msg = fmt.Sprintf("%d of %d KRIs breached red threshold", breachedCount, len(results))
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"total_kris":    len(results),
		"breached_kris": breachedCount,
		"details":       results,
	})

	return status, msg, resultData
}

// ============================================================
// POLICY ATTESTATION MONITOR
// ============================================================

// CheckPolicyAttestation checks attestation completion rates.
func (cm *ComplianceMonitor) CheckPolicyAttestation(ctx context.Context, m MonitorConfig) (string, string, json.RawMessage) {
	var conditions MonitorConditions
	if err := json.Unmarshal(m.Conditions, &conditions); err != nil {
		conditions.MinCompletionRate = 90.0
	}
	if conditions.MinCompletionRate <= 0 {
		conditions.MinCompletionRate = 90.0
	}

	var totalAttestation, completedAttestation int64
	err := cm.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'attested')
		FROM policy_attestations
		WHERE organization_id = $1
			AND (expires_at IS NULL OR expires_at > NOW())`,
		m.OrganizationID,
	).Scan(&totalAttestation, &completedAttestation)

	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return "failing", "Failed to query attestation data", data
	}

	var completionRate float64
	if totalAttestation > 0 {
		completionRate = float64(completedAttestation) / float64(totalAttestation) * 100
	}

	status := "passing"
	msg := fmt.Sprintf("Attestation rate: %.1f%% (%d/%d)", completionRate, completedAttestation, totalAttestation)
	if completionRate < conditions.MinCompletionRate {
		status = "failing"
		msg = fmt.Sprintf("Attestation rate %.1f%% below threshold %.1f%% (%d/%d)",
			completionRate, conditions.MinCompletionRate, completedAttestation, totalAttestation)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"total_attestations":     totalAttestation,
		"completed_attestations": completedAttestation,
		"completion_rate":        completionRate,
		"threshold":              conditions.MinCompletionRate,
	})

	return status, msg, resultData
}

// ============================================================
// VENDOR ASSESSMENT MONITOR
// ============================================================

// CheckVendorAssessment checks that vendor assessments are current.
func (cm *ComplianceMonitor) CheckVendorAssessment(ctx context.Context, m MonitorConfig) (string, string, json.RawMessage) {
	query := `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE next_assessment_date < CURRENT_DATE)
		FROM vendors
		WHERE organization_id = $1 AND status = 'active' AND deleted_at IS NULL`

	args := []interface{}{m.OrganizationID}
	if m.TargetEntityID != nil {
		query = `
			SELECT
				1,
				CASE WHEN next_assessment_date < CURRENT_DATE THEN 1 ELSE 0 END
			FROM vendors
			WHERE organization_id = $1 AND id = $2 AND deleted_at IS NULL`
		args = append(args, *m.TargetEntityID)
	}

	var totalVendors, overdueVendors int64
	err := cm.pool.QueryRow(ctx, query, args...).Scan(&totalVendors, &overdueVendors)
	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return "failing", "Failed to query vendor assessments", data
	}

	status := "passing"
	msg := fmt.Sprintf("All %d active vendors have current assessments", totalVendors)
	if overdueVendors > 0 {
		status = "failing"
		msg = fmt.Sprintf("%d of %d active vendors have overdue assessments", overdueVendors, totalVendors)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"total_vendors":   totalVendors,
		"overdue_vendors": overdueVendors,
	})

	return status, msg, resultData
}

// ============================================================
// COMPLIANCE SCORE MONITOR
// ============================================================

// CheckComplianceScore detects drops in compliance score below a threshold.
func (cm *ComplianceMonitor) CheckComplianceScore(ctx context.Context, m MonitorConfig) (string, string, json.RawMessage) {
	var conditions MonitorConditions
	if err := json.Unmarshal(m.Conditions, &conditions); err != nil {
		conditions.MinScore = 70.0
	}
	if conditions.MinScore <= 0 {
		conditions.MinScore = 70.0
	}

	rows, err := cm.pool.Query(ctx, `
		SELECT framework_code, framework_name, compliance_score
		FROM v_compliance_score_by_framework
		WHERE organization_id = $1`,
		m.OrganizationID)
	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		return "failing", "Failed to query compliance scores", data
	}
	defer rows.Close()

	type scoreEntry struct {
		FrameworkCode string  `json:"framework_code"`
		FrameworkName string  `json:"framework_name"`
		Score         float64 `json:"score"`
		BelowMin      bool    `json:"below_minimum"`
	}

	var entries []scoreEntry
	failingCount := 0

	for rows.Next() {
		var e scoreEntry
		if err := rows.Scan(&e.FrameworkCode, &e.FrameworkName, &e.Score); err != nil {
			continue
		}
		e.BelowMin = e.Score < conditions.MinScore
		if e.BelowMin {
			failingCount++
		}
		entries = append(entries, e)
	}

	status := "passing"
	msg := fmt.Sprintf("All %d framework scores above %.1f%%", len(entries), conditions.MinScore)
	if failingCount > 0 {
		status = "failing"
		msg = fmt.Sprintf("%d of %d frameworks below minimum score of %.1f%%", failingCount, len(entries), conditions.MinScore)
	}

	resultData, _ := json.Marshal(map[string]interface{}{
		"min_score":      conditions.MinScore,
		"failing_count":  failingCount,
		"framework_scores": entries,
	})

	return status, msg, resultData
}

// ============================================================
// HELPERS
// ============================================================

func (cm *ComplianceMonitor) sendAlert(ctx context.Context, m MonitorConfig, message string) {
	subject := fmt.Sprintf("[ComplianceForge] Monitor Alert: %s", m.Name)
	payload := queue.EmailPayload{
		Subject: subject,
		Body:    fmt.Sprintf("<h3>Compliance Monitor Alert</h3><p>Monitor: <strong>%s</strong></p><p>Type: %s</p><p>Severity: %s</p><p>%s</p>", m.Name, m.MonitorType, m.AlertSeverity, message),
	}

	_, err := cm.queue.Enqueue(ctx, "send_email", queue.QueueHigh, m.OrganizationID.String(), payload)
	if err != nil {
		log.Error().Err(err).Str("monitor_id", m.ID.String()).Msg("Failed to enqueue alert")
	}
}

func joinMessages(msgs []string) string {
	if len(msgs) == 0 {
		return ""
	}
	result := msgs[0]
	for i := 1; i < len(msgs); i++ {
		result += "; " + msgs[i]
	}
	return result
}

// ============================================================
// CRUD OPERATIONS
// ============================================================

// GetMonitor retrieves a single compliance monitor.
func (cm *ComplianceMonitor) GetMonitor(ctx context.Context, orgID, monitorID uuid.UUID) (*MonitorConfig, error) {
	var m MonitorConfig
	err := cm.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, monitor_type,
			target_entity_type, target_entity_id, check_frequency_cron,
			conditions, alert_on_failure, alert_severity,
			is_active, last_check_at, last_check_status,
			consecutive_failures, failure_since, created_at, updated_at
		FROM compliance_monitors
		WHERE id = $1 AND organization_id = $2`,
		monitorID, orgID,
	).Scan(
		&m.ID, &m.OrganizationID, &m.Name, &m.MonitorType,
		&m.TargetEntityType, &m.TargetEntityID, &m.CheckFrequencyCron,
		&m.Conditions, &m.AlertOnFailure, &m.AlertSeverity,
		&m.IsActive, &m.LastCheckAt, &m.LastCheckStatus,
		&m.ConsecutiveFailures, &m.FailureSince, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("monitor not found: %w", err)
	}
	return &m, nil
}

// ListMonitors returns paginated compliance monitors.
func (cm *ComplianceMonitor) ListMonitors(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]MonitorConfig, int64, error) {
	var total int64
	err := cm.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_monitors WHERE organization_id = $1`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := cm.pool.Query(ctx, `
		SELECT id, organization_id, name, monitor_type,
			target_entity_type, target_entity_id, check_frequency_cron,
			conditions, alert_on_failure, alert_severity,
			is_active, last_check_at, last_check_status,
			consecutive_failures, failure_since, created_at, updated_at
		FROM compliance_monitors
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var monitors []MonitorConfig
	for rows.Next() {
		var m MonitorConfig
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.Name, &m.MonitorType,
			&m.TargetEntityType, &m.TargetEntityID, &m.CheckFrequencyCron,
			&m.Conditions, &m.AlertOnFailure, &m.AlertSeverity,
			&m.IsActive, &m.LastCheckAt, &m.LastCheckStatus,
			&m.ConsecutiveFailures, &m.FailureSince, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		monitors = append(monitors, m)
	}

	return monitors, total, nil
}

// ListMonitorResults returns paginated results for a monitor.
func (cm *ComplianceMonitor) ListMonitorResults(ctx context.Context, orgID, monitorID uuid.UUID, limit, offset int) ([]MonitorResult, int64, error) {
	var total int64
	err := cm.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_monitor_results
		WHERE organization_id = $1 AND monitor_id = $2`, orgID, monitorID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := cm.pool.Query(ctx, `
		SELECT id, organization_id, monitor_id, status, check_time, result_data, message, created_at
		FROM compliance_monitor_results
		WHERE organization_id = $1 AND monitor_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		orgID, monitorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []MonitorResult
	for rows.Next() {
		var r MonitorResult
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.MonitorID,
			&r.Status, &r.CheckTime, &r.ResultData, &r.Message, &r.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		results = append(results, r)
	}

	return results, total, nil
}

// CreateMonitor creates a new compliance monitor.
func (cm *ComplianceMonitor) CreateMonitor(ctx context.Context, m *MonitorConfig) error {
	return cm.pool.QueryRow(ctx, `
		INSERT INTO compliance_monitors (
			organization_id, name, monitor_type,
			target_entity_type, target_entity_id, check_frequency_cron,
			conditions, alert_on_failure, alert_severity, is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at`,
		m.OrganizationID, m.Name, m.MonitorType,
		m.TargetEntityType, m.TargetEntityID, m.CheckFrequencyCron,
		m.Conditions, m.AlertOnFailure, m.AlertSeverity, m.IsActive,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

// Pool returns the database pool for direct queries by the handler layer.
func (cm *ComplianceMonitor) Pool() *pgxpool.Pool {
	return cm.pool
}

// UpdateMonitor updates an existing compliance monitor.
func (cm *ComplianceMonitor) UpdateMonitor(ctx context.Context, m *MonitorConfig) error {
	_, err := cm.pool.Exec(ctx, `
		UPDATE compliance_monitors SET
			name = $1, monitor_type = $2,
			target_entity_type = $3, target_entity_id = $4,
			check_frequency_cron = $5, conditions = $6,
			alert_on_failure = $7, alert_severity = $8, is_active = $9
		WHERE id = $10 AND organization_id = $11`,
		m.Name, m.MonitorType,
		m.TargetEntityType, m.TargetEntityID,
		m.CheckFrequencyCron, m.Conditions,
		m.AlertOnFailure, m.AlertSeverity, m.IsActive,
		m.ID, m.OrganizationID,
	)
	return err
}
