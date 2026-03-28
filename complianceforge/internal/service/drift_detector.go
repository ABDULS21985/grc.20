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
// DRIFT DETECTOR
// ============================================================

// DriftDetector analyses compliance monitor results and evidence
// collection outcomes to detect compliance drift. When drift is
// detected it creates drift events and emits notification jobs.
type DriftDetector struct {
	pool  *pgxpool.Pool
	queue *queue.Queue
}

// NewDriftDetector creates a new DriftDetector.
func NewDriftDetector(pool *pgxpool.Pool, q *queue.Queue) *DriftDetector {
	return &DriftDetector{pool: pool, queue: q}
}

// DriftEvent represents a compliance drift event.
type DriftEvent struct {
	ID               uuid.UUID  `json:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id"`
	DriftType        string     `json:"drift_type"`
	Severity         string     `json:"severity"`
	EntityType       string     `json:"entity_type"`
	EntityID         *uuid.UUID `json:"entity_id"`
	EntityRef        string     `json:"entity_ref"`
	Description      string     `json:"description"`
	PreviousState    string     `json:"previous_state"`
	CurrentState     string     `json:"current_state"`
	DetectedAt       time.Time  `json:"detected_at"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at"`
	AcknowledgedBy   *uuid.UUID `json:"acknowledged_by"`
	ResolvedAt       *time.Time `json:"resolved_at"`
	ResolvedBy       *uuid.UUID `json:"resolved_by"`
	ResolutionNotes  string     `json:"resolution_notes"`
	NotificationSent bool       `json:"notification_sent"`
	CreatedAt        time.Time  `json:"created_at"`
}

// DriftSummary provides an overview of drift events for the dashboard.
type DriftSummary struct {
	TotalActive     int64            `json:"total_active"`
	BySeverity      map[string]int64 `json:"by_severity"`
	ByType          map[string]int64 `json:"by_type"`
	Acknowledged    int64            `json:"acknowledged"`
	Unacknowledged  int64            `json:"unacknowledged"`
	ResolvedLast30d int64            `json:"resolved_last_30d"`
	RecentEvents    []DriftEvent     `json:"recent_events"`
}

// ============================================================
// ANALYZE
// ============================================================

// Analyze inspects failing monitors and stale evidence to create drift events.
// It should be called periodically (e.g., every 15 minutes).
func (dd *DriftDetector) Analyze(ctx context.Context, orgID uuid.UUID) error {
	if err := dd.detectControlDegradation(ctx, orgID); err != nil {
		log.Error().Err(err).Msg("Failed to detect control degradation")
	}
	if err := dd.detectExpiredEvidence(ctx, orgID); err != nil {
		log.Error().Err(err).Msg("Failed to detect expired evidence")
	}
	if err := dd.detectKRIBreaches(ctx, orgID); err != nil {
		log.Error().Err(err).Msg("Failed to detect KRI breaches")
	}
	if err := dd.detectUnattestedPolicies(ctx, orgID); err != nil {
		log.Error().Err(err).Msg("Failed to detect unattested policies")
	}
	if err := dd.detectOverdueVendors(ctx, orgID); err != nil {
		log.Error().Err(err).Msg("Failed to detect overdue vendors")
	}
	if err := dd.detectScoreDrops(ctx, orgID); err != nil {
		log.Error().Err(err).Msg("Failed to detect score drops")
	}
	return nil
}

// detectControlDegradation finds controls that have moved to a worse status.
func (dd *DriftDetector) detectControlDegradation(ctx context.Context, orgID uuid.UUID) error {
	// Find controls whose collection configs have consecutive failures above threshold
	rows, err := dd.pool.Query(ctx, `
		SELECT ecc.control_implementation_id, ci.status::TEXT, ecc.consecutive_failures, ecc.failure_threshold
		FROM evidence_collection_configs ecc
		JOIN control_implementations ci ON ci.id = ecc.control_implementation_id
		WHERE ecc.organization_id = $1
			AND ecc.is_active = true
			AND ecc.consecutive_failures >= ecc.failure_threshold
			AND ecc.control_implementation_id IS NOT NULL
			AND ci.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1 FROM compliance_drift_events cde
				WHERE cde.organization_id = $1
					AND cde.drift_type = 'control_degraded'
					AND cde.entity_id = ecc.control_implementation_id
					AND cde.resolved_at IS NULL
			)`, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID uuid.UUID
		var controlStatus string
		var failures, threshold int
		if err := rows.Scan(&entityID, &controlStatus, &failures, &threshold); err != nil {
			continue
		}

		dd.createDriftEvent(ctx, orgID, DriftEvent{
			DriftType:     "control_degraded",
			Severity:      "high",
			EntityType:    "control_implementation",
			EntityID:      &entityID,
			Description:   fmt.Sprintf("Evidence collection has failed %d consecutive times (threshold: %d)", failures, threshold),
			PreviousState: controlStatus,
			CurrentState:  "evidence_collection_failing",
		})
	}

	return nil
}

// detectExpiredEvidence finds controls with no recent evidence.
func (dd *DriftDetector) detectExpiredEvidence(ctx context.Context, orgID uuid.UUID) error {
	rows, err := dd.pool.Query(ctx, `
		SELECT ci.id, ci.status::TEXT, MAX(ce.collected_at) AS latest
		FROM control_implementations ci
		LEFT JOIN control_evidence ce ON ce.control_implementation_id = ci.id
			AND ce.is_current = true AND ce.deleted_at IS NULL
		WHERE ci.organization_id = $1
			AND ci.status IN ('implemented', 'effective')
			AND ci.deleted_at IS NULL
		GROUP BY ci.id, ci.status
		HAVING MAX(ce.collected_at) < NOW() - INTERVAL '90 days'
			OR MAX(ce.collected_at) IS NULL
		AND NOT EXISTS (
			SELECT 1 FROM compliance_drift_events cde
			WHERE cde.organization_id = $1
				AND cde.drift_type = 'evidence_expired'
				AND cde.entity_id = ci.id
				AND cde.resolved_at IS NULL
		)`, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID uuid.UUID
		var controlStatus string
		var latestEvidence *time.Time
		if err := rows.Scan(&entityID, &controlStatus, &latestEvidence); err != nil {
			continue
		}

		desc := "No evidence collected for this control"
		if latestEvidence != nil {
			daysAgo := int(time.Since(*latestEvidence).Hours() / 24)
			desc = fmt.Sprintf("Latest evidence is %d days old", daysAgo)
		}

		dd.createDriftEvent(ctx, orgID, DriftEvent{
			DriftType:     "evidence_expired",
			Severity:      "medium",
			EntityType:    "control_implementation",
			EntityID:      &entityID,
			Description:   desc,
			PreviousState: "evidence_current",
			CurrentState:  "evidence_expired",
		})
	}

	return nil
}

// detectKRIBreaches finds KRIs above their red threshold.
func (dd *DriftDetector) detectKRIBreaches(ctx context.Context, orgID uuid.UUID) error {
	rows, err := dd.pool.Query(ctx, `
		SELECT ri.id, ri.name, ri.current_value, ri.threshold_red
		FROM risk_indicators ri
		WHERE ri.organization_id = $1
			AND ri.current_value >= ri.threshold_red
			AND ri.threshold_red IS NOT NULL
			AND NOT EXISTS (
				SELECT 1 FROM compliance_drift_events cde
				WHERE cde.organization_id = $1
					AND cde.drift_type = 'kri_breached'
					AND cde.entity_id = ri.id
					AND cde.resolved_at IS NULL
			)`, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID uuid.UUID
		var name string
		var currentValue, threshold float64
		if err := rows.Scan(&entityID, &name, &currentValue, &threshold); err != nil {
			continue
		}

		dd.createDriftEvent(ctx, orgID, DriftEvent{
			DriftType:     "kri_breached",
			Severity:      "critical",
			EntityType:    "risk_indicator",
			EntityID:      &entityID,
			EntityRef:     name,
			Description:   fmt.Sprintf("KRI '%s' value %.2f exceeds red threshold %.2f", name, currentValue, threshold),
			PreviousState: "within_threshold",
			CurrentState:  fmt.Sprintf("value_%.2f", currentValue),
		})
	}

	return nil
}

// detectUnattestedPolicies finds policies with low attestation completion.
func (dd *DriftDetector) detectUnattestedPolicies(ctx context.Context, orgID uuid.UUID) error {
	rows, err := dd.pool.Query(ctx, `
		SELECT p.id, p.policy_ref, p.title,
			COUNT(pa.id) AS total,
			COUNT(pa.id) FILTER (WHERE pa.status = 'attested') AS attested
		FROM policies p
		JOIN policy_attestations pa ON pa.policy_id = p.id
		WHERE p.organization_id = $1
			AND p.requires_attestation = true
			AND p.deleted_at IS NULL
			AND (pa.expires_at IS NULL OR pa.expires_at > NOW())
		GROUP BY p.id, p.policy_ref, p.title
		HAVING COUNT(pa.id) FILTER (WHERE pa.status = 'attested')::FLOAT / NULLIF(COUNT(pa.id), 0) < 0.5
			AND NOT EXISTS (
				SELECT 1 FROM compliance_drift_events cde
				WHERE cde.organization_id = $1
					AND cde.drift_type = 'policy_unattested'
					AND cde.entity_id = p.id
					AND cde.resolved_at IS NULL
			)`, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID uuid.UUID
		var policyRef, title string
		var total, attested int64
		if err := rows.Scan(&entityID, &policyRef, &title, &total, &attested); err != nil {
			continue
		}

		rate := float64(attested) / float64(total) * 100

		dd.createDriftEvent(ctx, orgID, DriftEvent{
			DriftType:     "policy_unattested",
			Severity:      "medium",
			EntityType:    "policy",
			EntityID:      &entityID,
			EntityRef:     policyRef,
			Description:   fmt.Sprintf("Policy '%s' (%s) attestation rate is %.1f%% (%d/%d)", title, policyRef, rate, attested, total),
			PreviousState: "attested",
			CurrentState:  fmt.Sprintf("attestation_rate_%.0f", rate),
		})
	}

	return nil
}

// detectOverdueVendors finds vendors with overdue assessments.
func (dd *DriftDetector) detectOverdueVendors(ctx context.Context, orgID uuid.UUID) error {
	rows, err := dd.pool.Query(ctx, `
		SELECT v.id, v.vendor_ref, v.name, v.next_assessment_date, v.risk_tier::TEXT
		FROM vendors v
		WHERE v.organization_id = $1
			AND v.status = 'active'
			AND v.next_assessment_date < CURRENT_DATE
			AND v.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1 FROM compliance_drift_events cde
				WHERE cde.organization_id = $1
					AND cde.drift_type = 'vendor_overdue'
					AND cde.entity_id = v.id
					AND cde.resolved_at IS NULL
			)`, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID uuid.UUID
		var vendorRef, name, riskTier string
		var nextAssessment time.Time
		if err := rows.Scan(&entityID, &vendorRef, &name, &nextAssessment, &riskTier); err != nil {
			continue
		}

		daysOverdue := int(time.Since(nextAssessment).Hours() / 24)
		severity := "medium"
		if riskTier == "critical" || riskTier == "high" {
			severity = "high"
		}

		dd.createDriftEvent(ctx, orgID, DriftEvent{
			DriftType:     "vendor_overdue",
			Severity:      severity,
			EntityType:    "vendor",
			EntityID:      &entityID,
			EntityRef:     vendorRef,
			Description:   fmt.Sprintf("Vendor '%s' (%s) assessment overdue by %d days (risk tier: %s)", name, vendorRef, daysOverdue, riskTier),
			PreviousState: "assessment_current",
			CurrentState:  fmt.Sprintf("overdue_%d_days", daysOverdue),
		})
	}

	return nil
}

// detectScoreDrops finds frameworks whose compliance score dropped significantly.
func (dd *DriftDetector) detectScoreDrops(ctx context.Context, orgID uuid.UUID) error {
	rows, err := dd.pool.Query(ctx, `
		SELECT of2.id, cf.code, cf.name, of2.compliance_score
		FROM organization_frameworks of2
		JOIN compliance_frameworks cf ON of2.framework_id = cf.id
		WHERE of2.organization_id = $1
			AND of2.compliance_score < 70
			AND NOT EXISTS (
				SELECT 1 FROM compliance_drift_events cde
				WHERE cde.organization_id = $1
					AND cde.drift_type = 'score_dropped'
					AND cde.entity_id = of2.id
					AND cde.resolved_at IS NULL
			)`, orgID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var entityID uuid.UUID
		var code, name string
		var score float64
		if err := rows.Scan(&entityID, &code, &name, &score); err != nil {
			continue
		}

		severity := "medium"
		if score < 50 {
			severity = "critical"
		} else if score < 60 {
			severity = "high"
		}

		dd.createDriftEvent(ctx, orgID, DriftEvent{
			DriftType:     "score_dropped",
			Severity:      severity,
			EntityType:    "organization_framework",
			EntityID:      &entityID,
			EntityRef:     code,
			Description:   fmt.Sprintf("Framework '%s' (%s) compliance score dropped to %.1f%%", name, code, score),
			PreviousState: "above_threshold",
			CurrentState:  fmt.Sprintf("score_%.1f", score),
		})
	}

	return nil
}

// ============================================================
// DRIFT EVENT MANAGEMENT
// ============================================================

func (dd *DriftDetector) createDriftEvent(ctx context.Context, orgID uuid.UUID, event DriftEvent) {
	var eventID uuid.UUID
	err := dd.pool.QueryRow(ctx, `
		INSERT INTO compliance_drift_events (
			organization_id, drift_type, severity,
			entity_type, entity_id, entity_ref,
			description, previous_state, current_state, detected_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())
		RETURNING id`,
		orgID, event.DriftType, event.Severity,
		event.EntityType, event.EntityID, event.EntityRef,
		event.Description, event.PreviousState, event.CurrentState,
	).Scan(&eventID)

	if err != nil {
		log.Error().Err(err).Str("drift_type", event.DriftType).Msg("Failed to create drift event")
		return
	}

	log.Info().
		Str("event_id", eventID.String()).
		Str("drift_type", event.DriftType).
		Str("severity", event.Severity).
		Msg("Compliance drift detected")

	// Emit notification
	dd.emitDriftNotification(ctx, orgID, event)
}

func (dd *DriftDetector) emitDriftNotification(ctx context.Context, orgID uuid.UUID, event DriftEvent) {
	severityLabel := "Medium"
	switch event.Severity {
	case "critical":
		severityLabel = "CRITICAL"
	case "high":
		severityLabel = "High"
	case "low":
		severityLabel = "Low"
	}

	subject := fmt.Sprintf("[ComplianceForge] %s Drift Detected: %s", severityLabel, event.DriftType)
	body := fmt.Sprintf(
		"<h3>Compliance Drift Detected</h3>"+
			"<p><strong>Type:</strong> %s</p>"+
			"<p><strong>Severity:</strong> %s</p>"+
			"<p><strong>Entity:</strong> %s %s</p>"+
			"<p>%s</p>"+
			"<p><strong>Previous State:</strong> %s</p>"+
			"<p><strong>Current State:</strong> %s</p>",
		event.DriftType, event.Severity,
		event.EntityType, event.EntityRef,
		event.Description,
		event.PreviousState, event.CurrentState,
	)

	queueName := queue.QueueDefault
	if event.Severity == "critical" {
		queueName = queue.QueueCritical
	} else if event.Severity == "high" {
		queueName = queue.QueueHigh
	}

	payload := queue.EmailPayload{
		Subject: subject,
		Body:    body,
	}

	if _, err := dd.queue.Enqueue(ctx, "send_email", queueName, orgID.String(), payload); err != nil {
		log.Error().Err(err).Msg("Failed to enqueue drift notification")
	}
}

// ============================================================
// QUERY METHODS
// ============================================================

// GetDriftSummary returns a dashboard summary of drift events for an organization.
func (dd *DriftDetector) GetDriftSummary(ctx context.Context, orgID uuid.UUID) (*DriftSummary, error) {
	summary := &DriftSummary{
		BySeverity: make(map[string]int64),
		ByType:     make(map[string]int64),
	}

	// Total active (unresolved)
	err := dd.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_drift_events
		WHERE organization_id = $1 AND resolved_at IS NULL`, orgID).Scan(&summary.TotalActive)
	if err != nil {
		return nil, fmt.Errorf("failed to count active drift: %w", err)
	}

	// By severity
	rows, err := dd.pool.Query(ctx, `
		SELECT severity::TEXT, COUNT(*)
		FROM compliance_drift_events
		WHERE organization_id = $1 AND resolved_at IS NULL
		GROUP BY severity`, orgID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var sev string
		var cnt int64
		if err := rows.Scan(&sev, &cnt); err == nil {
			summary.BySeverity[sev] = cnt
		}
	}
	rows.Close()

	// By type
	rows, err = dd.pool.Query(ctx, `
		SELECT drift_type::TEXT, COUNT(*)
		FROM compliance_drift_events
		WHERE organization_id = $1 AND resolved_at IS NULL
		GROUP BY drift_type`, orgID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var dt string
		var cnt int64
		if err := rows.Scan(&dt, &cnt); err == nil {
			summary.ByType[dt] = cnt
		}
	}
	rows.Close()

	// Acknowledged vs unacknowledged
	err = dd.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE acknowledged_at IS NOT NULL),
			COUNT(*) FILTER (WHERE acknowledged_at IS NULL)
		FROM compliance_drift_events
		WHERE organization_id = $1 AND resolved_at IS NULL`, orgID,
	).Scan(&summary.Acknowledged, &summary.Unacknowledged)
	if err != nil {
		return nil, err
	}

	// Resolved in last 30 days
	err = dd.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_drift_events
		WHERE organization_id = $1 AND resolved_at > NOW() - INTERVAL '30 days'`, orgID,
	).Scan(&summary.ResolvedLast30d)
	if err != nil {
		return nil, err
	}

	// Recent events (limit 10)
	recentRows, err := dd.pool.Query(ctx, `
		SELECT id, organization_id, drift_type::TEXT, severity::TEXT,
			entity_type, entity_id, entity_ref, description,
			previous_state, current_state, detected_at,
			acknowledged_at, acknowledged_by, resolved_at, resolved_by,
			resolution_notes, notification_sent, created_at
		FROM compliance_drift_events
		WHERE organization_id = $1 AND resolved_at IS NULL
		ORDER BY detected_at DESC
		LIMIT 10`, orgID)
	if err != nil {
		return nil, err
	}
	defer recentRows.Close()

	for recentRows.Next() {
		var e DriftEvent
		if err := recentRows.Scan(
			&e.ID, &e.OrganizationID, &e.DriftType, &e.Severity,
			&e.EntityType, &e.EntityID, &e.EntityRef, &e.Description,
			&e.PreviousState, &e.CurrentState, &e.DetectedAt,
			&e.AcknowledgedAt, &e.AcknowledgedBy, &e.ResolvedAt, &e.ResolvedBy,
			&e.ResolutionNotes, &e.NotificationSent, &e.CreatedAt,
		); err != nil {
			continue
		}
		summary.RecentEvents = append(summary.RecentEvents, e)
	}

	return summary, nil
}

// ListDriftEvents returns paginated active drift events.
func (dd *DriftDetector) ListDriftEvents(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]DriftEvent, int64, error) {
	var total int64
	err := dd.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_drift_events
		WHERE organization_id = $1 AND resolved_at IS NULL`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := dd.pool.Query(ctx, `
		SELECT id, organization_id, drift_type::TEXT, severity::TEXT,
			entity_type, entity_id, entity_ref, description,
			previous_state, current_state, detected_at,
			acknowledged_at, acknowledged_by, resolved_at, resolved_by,
			resolution_notes, notification_sent, created_at
		FROM compliance_drift_events
		WHERE organization_id = $1 AND resolved_at IS NULL
		ORDER BY
			CASE severity
				WHEN 'critical' THEN 0
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
			END,
			detected_at DESC
		LIMIT $2 OFFSET $3`,
		orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []DriftEvent
	for rows.Next() {
		var e DriftEvent
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.DriftType, &e.Severity,
			&e.EntityType, &e.EntityID, &e.EntityRef, &e.Description,
			&e.PreviousState, &e.CurrentState, &e.DetectedAt,
			&e.AcknowledgedAt, &e.AcknowledgedBy, &e.ResolvedAt, &e.ResolvedBy,
			&e.ResolutionNotes, &e.NotificationSent, &e.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}

	return events, total, nil
}

// AcknowledgeDrift marks a drift event as acknowledged by a user.
func (dd *DriftDetector) AcknowledgeDrift(ctx context.Context, orgID, eventID, userID uuid.UUID) error {
	tag, err := dd.pool.Exec(ctx, `
		UPDATE compliance_drift_events
		SET acknowledged_at = NOW(), acknowledged_by = $1
		WHERE id = $2 AND organization_id = $3 AND resolved_at IS NULL`,
		userID, eventID, orgID)
	if err != nil {
		return fmt.Errorf("failed to acknowledge drift: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("drift event not found or already resolved")
	}
	return nil
}

// ResolveDrift marks a drift event as resolved with optional notes.
func (dd *DriftDetector) ResolveDrift(ctx context.Context, orgID, eventID, userID uuid.UUID, notes string) error {
	tag, err := dd.pool.Exec(ctx, `
		UPDATE compliance_drift_events
		SET resolved_at = NOW(), resolved_by = $1, resolution_notes = $2
		WHERE id = $3 AND organization_id = $4 AND resolved_at IS NULL`,
		userID, notes, eventID, orgID)
	if err != nil {
		return fmt.Errorf("failed to resolve drift: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("drift event not found or already resolved")
	}
	return nil
}
