// Package worker provides background job processing for ComplianceForge.
// RegulatoryScheduler checks for regulatory deadlines and compliance obligations
// on a periodic basis and emits events to the notification engine.
package worker

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// RegulatoryScheduler runs periodic checks for regulatory deadlines and
// emits events to the notification engine's event bus.
type RegulatoryScheduler struct {
	db       *pgxpool.Pool
	eventBus *service.EventBus
	interval time.Duration
	baseURL  string
}

// NewRegulatoryScheduler creates a new scheduler that checks every 15 minutes.
func NewRegulatoryScheduler(db *pgxpool.Pool, eventBus *service.EventBus, baseURL string) *RegulatoryScheduler {
	return &RegulatoryScheduler{
		db:       db,
		eventBus: eventBus,
		interval: 15 * time.Minute,
		baseURL:  baseURL,
	}
}

// Start begins the scheduler loop. It blocks until the context is cancelled.
func (s *RegulatoryScheduler) Start(ctx context.Context) {
	log.Info().Dur("interval", s.interval).Msg("Regulatory scheduler started")

	// Run immediately on start
	s.runChecks(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Regulatory scheduler shutting down")
			return
		case <-ticker.C:
			s.runChecks(ctx)
		}
	}
}

// runChecks executes all regulatory deadline checks.
func (s *RegulatoryScheduler) runChecks(ctx context.Context) {
	log.Debug().Msg("Running regulatory deadline checks")

	s.checkGDPRBreachDeadlines(ctx)
	s.checkNIS2EarlyWarnings(ctx)
	s.checkOverduePolicies(ctx)
	s.checkOverdueFindings(ctx)
	s.checkVendorAssessments(ctx)
	s.checkVendorMissingDPAs(ctx)
	s.checkRiskReviews(ctx)

	log.Debug().Msg("Regulatory deadline checks complete")
}

// ============================================================
// GDPR BREACH 72h DEADLINE
// ============================================================

// checkGDPRBreachDeadlines checks for data breaches approaching the 72-hour
// GDPR notification deadline. Emits events at 48h, 12h, 6h, 1h remaining, and when expired.
func (s *RegulatoryScheduler) checkGDPRBreachDeadlines(ctx context.Context) {
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.organization_id, i.incident_ref, i.title, i.severity,
			   i.data_subjects_affected, i.notification_deadline,
			   EXTRACT(EPOCH FROM (i.notification_deadline - NOW())) / 3600 AS hours_remaining
		FROM incidents i
		WHERE i.is_data_breach = true
		  AND i.notification_required = true
		  AND i.dpa_notified_at IS NULL
		  AND i.notification_deadline IS NOT NULL
		  AND i.status NOT IN ('resolved', 'closed')
		  AND i.deleted_at IS NULL
		ORDER BY i.notification_deadline ASC`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check GDPR breach deadlines")
		return
	}
	defer rows.Close()

	// Thresholds at which we emit events (hours remaining)
	thresholds := []float64{48, 12, 6, 1, 0}

	for rows.Next() {
		var (
			incidentID     uuid.UUID
			orgID          uuid.UUID
			incidentRef    string
			title          string
			severity       string
			dataSubjects   int
			deadline       time.Time
			hoursRemaining float64
		)

		if err := rows.Scan(&incidentID, &orgID, &incidentRef, &title, &severity,
			&dataSubjects, &deadline, &hoursRemaining); err != nil {
			log.Error().Err(err).Msg("Failed to scan breach row")
			continue
		}

		dashURL := fmt.Sprintf("%s/incidents/%s", s.baseURL, incidentID)

		// Determine which threshold we are at
		for _, threshold := range thresholds {
			if hoursRemaining <= threshold && hoursRemaining > threshold-0.5 {
				eventType := service.EventBreachDeadlineApproaching
				if threshold == 0 || hoursRemaining <= 0 {
					eventType = service.EventBreachDeadlineExpired
				}

				s.eventBus.Publish(service.Event{
					Type:     eventType,
					OrgID:    orgID,
					Severity: "critical",
					Payload: map[string]interface{}{
						"IncidentRef":          incidentRef,
						"Title":                title,
						"Severity":             severity,
						"DataSubjectsAffected": fmt.Sprintf("%d", dataSubjects),
						"Deadline":             deadline.Format(time.RFC3339),
						"HoursRemaining":       fmt.Sprintf("%.1f", math.Max(0, hoursRemaining)),
						"DashboardURL":         dashURL,
						"incident_id":          incidentID.String(),
					},
					Timestamp: time.Now(),
				})

				log.Warn().
					Str("incident_ref", incidentRef).
					Float64("hours_remaining", hoursRemaining).
					Str("org_id", orgID.String()).
					Msg("GDPR breach deadline event emitted")

				break // Only emit one threshold event per breach per check
			}
		}
	}
}

// ============================================================
// NIS2 24h EARLY WARNING
// ============================================================

// checkNIS2EarlyWarnings checks for NIS2-reportable incidents that haven't
// had their 24-hour early warning submitted.
func (s *RegulatoryScheduler) checkNIS2EarlyWarnings(ctx context.Context) {
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.organization_id, i.incident_ref, i.title, i.severity,
			   i.reported_at,
			   EXTRACT(EPOCH FROM ((i.reported_at + interval '24 hours') - NOW())) / 3600 AS hours_remaining
		FROM incidents i
		WHERE i.is_nis2_reportable = true
		  AND i.nis2_early_warning_at IS NULL
		  AND i.status NOT IN ('resolved', 'closed')
		  AND i.deleted_at IS NULL
		  AND (i.reported_at + interval '24 hours') > NOW() - interval '1 hour'
		ORDER BY i.reported_at ASC`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check NIS2 early warnings")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			incidentID     uuid.UUID
			orgID          uuid.UUID
			incidentRef    string
			title          string
			severity       string
			reportedAt     time.Time
			hoursRemaining float64
		)

		if err := rows.Scan(&incidentID, &orgID, &incidentRef, &title, &severity,
			&reportedAt, &hoursRemaining); err != nil {
			log.Error().Err(err).Msg("Failed to scan NIS2 row")
			continue
		}

		deadline := reportedAt.Add(24 * time.Hour)
		dashURL := fmt.Sprintf("%s/incidents/%s", s.baseURL, incidentID)

		s.eventBus.Publish(service.Event{
			Type:     service.EventNIS2EarlyWarningDue,
			OrgID:    orgID,
			Severity: "high",
			Payload: map[string]interface{}{
				"IncidentRef":    incidentRef,
				"Title":          title,
				"Severity":       severity,
				"Deadline":       deadline.Format(time.RFC3339),
				"HoursRemaining": fmt.Sprintf("%.1f", math.Max(0, hoursRemaining)),
				"DashboardURL":   dashURL,
				"incident_id":    incidentID.String(),
			},
			Timestamp: time.Now(),
		})
	}
}

// ============================================================
// OVERDUE POLICIES
// ============================================================

// checkOverduePolicies checks for policies whose review date has passed or is imminent.
func (s *RegulatoryScheduler) checkOverduePolicies(ctx context.Context) {
	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.organization_id, p.policy_ref, p.title,
			   p.next_review_date, p.last_review_date,
			   COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned') AS owner_name,
			   EXTRACT(DAY FROM (NOW() - p.next_review_date)) AS days_overdue
		FROM policies p
		LEFT JOIN users u ON p.owner_user_id = u.id
		WHERE p.next_review_date IS NOT NULL
		  AND p.next_review_date <= CURRENT_DATE + interval '7 days'
		  AND p.status NOT IN ('archived', 'retired')
		  AND p.deleted_at IS NULL
		ORDER BY p.next_review_date ASC`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check overdue policies")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			policyID       uuid.UUID
			orgID          uuid.UUID
			policyRef      string
			title          string
			nextReviewDate time.Time
			lastReviewDate *time.Time
			ownerName      string
			daysOverdue    float64
		)

		if err := rows.Scan(&policyID, &orgID, &policyRef, &title,
			&nextReviewDate, &lastReviewDate, &ownerName, &daysOverdue); err != nil {
			log.Error().Err(err).Msg("Failed to scan policy row")
			continue
		}

		dashURL := fmt.Sprintf("%s/policies/%s", s.baseURL, policyID)
		lastReviewStr := "Never"
		if lastReviewDate != nil {
			lastReviewStr = lastReviewDate.Format("2006-01-02")
		}

		eventType := service.EventPolicyReviewDue
		severity := "medium"
		if daysOverdue > 0 {
			eventType = service.EventPolicyReviewOverdue
			severity = "high"
		}

		s.eventBus.Publish(service.Event{
			Type:     eventType,
			OrgID:    orgID,
			Severity: severity,
			Payload: map[string]interface{}{
				"PolicyRef":      policyRef,
				"Title":          title,
				"DueDate":        nextReviewDate.Format("2006-01-02"),
				"OwnerName":      ownerName,
				"LastReviewDate": lastReviewStr,
				"DaysOverdue":    fmt.Sprintf("%.0f", math.Max(0, daysOverdue)),
				"DashboardURL":   dashURL,
				"policy_id":      policyID.String(),
			},
			Timestamp: time.Now(),
		})
	}
}

// ============================================================
// OVERDUE FINDINGS
// ============================================================

// checkOverdueFindings checks for audit findings that have exceeded their remediation deadline.
func (s *RegulatoryScheduler) checkOverdueFindings(ctx context.Context) {
	rows, err := s.db.Query(ctx, `
		SELECT f.id, f.organization_id, f.finding_ref, f.title, f.severity,
			   f.due_date, f.audit_id,
			   COALESCE(a.audit_ref, '') AS audit_ref,
			   COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned') AS responsible_name,
			   EXTRACT(DAY FROM (NOW() - f.due_date)) AS days_overdue
		FROM audit_findings f
		JOIN audits a ON f.audit_id = a.id
		LEFT JOIN users u ON f.responsible_user_id = u.id
		WHERE f.due_date IS NOT NULL
		  AND f.due_date < CURRENT_DATE
		  AND f.status IN ('open', 'in_progress')
		  AND f.deleted_at IS NULL
		ORDER BY f.severity, f.due_date ASC`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check overdue findings")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			findingID       uuid.UUID
			orgID           uuid.UUID
			findingRef      string
			title           string
			severity        string
			dueDate         time.Time
			auditID         uuid.UUID
			auditRef        string
			responsibleName string
			daysOverdue     float64
		)

		if err := rows.Scan(&findingID, &orgID, &findingRef, &title, &severity,
			&dueDate, &auditID, &auditRef, &responsibleName, &daysOverdue); err != nil {
			log.Error().Err(err).Msg("Failed to scan finding row")
			continue
		}

		dashURL := fmt.Sprintf("%s/audits/%s/findings/%s", s.baseURL, auditID, findingID)

		s.eventBus.Publish(service.Event{
			Type:     service.EventFindingRemediationOverdue,
			OrgID:    orgID,
			Severity: severity,
			Payload: map[string]interface{}{
				"FindingRef":      findingRef,
				"Title":           title,
				"Severity":        severity,
				"DueDate":         dueDate.Format("2006-01-02"),
				"DaysOverdue":     fmt.Sprintf("%.0f", daysOverdue),
				"AuditRef":        auditRef,
				"ResponsibleName": responsibleName,
				"DashboardURL":    dashURL,
				"finding_id":      findingID.String(),
			},
			Timestamp: time.Now(),
		})
	}
}

// ============================================================
// VENDOR ASSESSMENTS
// ============================================================

// checkVendorAssessments checks for vendor assessments that are due or overdue.
func (s *RegulatoryScheduler) checkVendorAssessments(ctx context.Context) {
	rows, err := s.db.Query(ctx, `
		SELECT v.id, v.organization_id, v.vendor_ref, v.name, v.risk_tier,
			   v.next_assessment_date,
			   COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned') AS owner_name
		FROM vendors v
		LEFT JOIN users u ON v.owner_user_id = u.id
		WHERE v.next_assessment_date IS NOT NULL
		  AND v.next_assessment_date <= CURRENT_DATE + interval '14 days'
		  AND v.status = 'active'
		  AND v.deleted_at IS NULL
		ORDER BY v.next_assessment_date ASC`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check vendor assessments")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			vendorID       uuid.UUID
			orgID          uuid.UUID
			vendorRef      string
			vendorName     string
			riskTier       string
			nextAssessment time.Time
			ownerName      string
		)

		if err := rows.Scan(&vendorID, &orgID, &vendorRef, &vendorName, &riskTier,
			&nextAssessment, &ownerName); err != nil {
			log.Error().Err(err).Msg("Failed to scan vendor row")
			continue
		}

		dashURL := fmt.Sprintf("%s/vendors/%s", s.baseURL, vendorID)

		s.eventBus.Publish(service.Event{
			Type:     service.EventVendorAssessmentDue,
			OrgID:    orgID,
			Severity: "medium",
			Payload: map[string]interface{}{
				"VendorRef":   vendorRef,
				"VendorName":  vendorName,
				"RiskTier":    riskTier,
				"DueDate":     nextAssessment.Format("2006-01-02"),
				"OwnerName":   ownerName,
				"DashboardURL": dashURL,
				"vendor_id":   vendorID.String(),
			},
			Timestamp: time.Now(),
		})
	}
}

// ============================================================
// VENDOR MISSING DPAs
// ============================================================

// checkVendorMissingDPAs checks for active vendors that process personal data
// without a Data Processing Agreement in place (GDPR Article 28 requirement).
func (s *RegulatoryScheduler) checkVendorMissingDPAs(ctx context.Context) {
	rows, err := s.db.Query(ctx, `
		SELECT v.id, v.organization_id, v.vendor_ref, v.name, v.risk_tier,
			   COALESCE(array_to_string(v.data_categories, ', '), '') AS data_categories
		FROM vendors v
		WHERE v.data_processing = true
		  AND v.dpa_in_place = false
		  AND v.status = 'active'
		  AND v.deleted_at IS NULL
		ORDER BY v.risk_tier, v.name`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check vendor DPAs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			vendorID       uuid.UUID
			orgID          uuid.UUID
			vendorRef      string
			vendorName     string
			riskTier       string
			dataCategories string
		)

		if err := rows.Scan(&vendorID, &orgID, &vendorRef, &vendorName,
			&riskTier, &dataCategories); err != nil {
			log.Error().Err(err).Msg("Failed to scan vendor DPA row")
			continue
		}

		dashURL := fmt.Sprintf("%s/vendors/%s", s.baseURL, vendorID)

		s.eventBus.Publish(service.Event{
			Type:     service.EventVendorMissingDPA,
			OrgID:    orgID,
			Severity: "high",
			Payload: map[string]interface{}{
				"VendorRef":      vendorRef,
				"VendorName":     vendorName,
				"RiskTier":       riskTier,
				"DataCategories": dataCategories,
				"DashboardURL":   dashURL,
				"vendor_id":      vendorID.String(),
			},
			Timestamp: time.Now(),
		})
	}
}

// ============================================================
// RISK REVIEWS
// ============================================================

// checkRiskReviews checks for risks whose next review date has passed or is imminent.
func (s *RegulatoryScheduler) checkRiskReviews(ctx context.Context) {
	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.organization_id, r.risk_ref, r.title,
			   r.residual_risk_level, r.next_review_date,
			   COALESCE(u.first_name || ' ' || u.last_name, 'Unassigned') AS owner_name
		FROM risks r
		LEFT JOIN users u ON r.owner_user_id = u.id
		WHERE r.next_review_date IS NOT NULL
		  AND r.next_review_date <= CURRENT_DATE + interval '7 days'
		  AND r.status NOT IN ('closed')
		  AND r.deleted_at IS NULL
		ORDER BY r.next_review_date ASC`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check risk reviews")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			riskID     uuid.UUID
			orgID      uuid.UUID
			riskRef    string
			title      string
			riskLevel  string
			reviewDate time.Time
			ownerName  string
		)

		if err := rows.Scan(&riskID, &orgID, &riskRef, &title,
			&riskLevel, &reviewDate, &ownerName); err != nil {
			log.Error().Err(err).Msg("Failed to scan risk review row")
			continue
		}

		dashURL := fmt.Sprintf("%s/risks/%s", s.baseURL, riskID)

		s.eventBus.Publish(service.Event{
			Type:     service.EventRiskReviewDue,
			OrgID:    orgID,
			Severity: "medium",
			Payload: map[string]interface{}{
				"RiskRef":      riskRef,
				"Title":        title,
				"RiskLevel":    riskLevel,
				"DueDate":      reviewDate.Format("2006-01-02"),
				"OwnerName":    ownerName,
				"DashboardURL": dashURL,
				"risk_id":      riskID.String(),
			},
			Timestamp: time.Now(),
		})
	}
}
