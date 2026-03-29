package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// MODEL TYPES
// ============================================================

// EvidenceTestSuite represents a collection of evidence test cases that
// can be executed together.
type EvidenceTestSuite struct {
	ID                   uuid.UUID  `json:"id"`
	OrganizationID       uuid.UUID  `json:"organization_id"`
	Name                 string     `json:"name"`
	Description          string     `json:"description"`
	TestType             string     `json:"test_type"`
	ScheduleCron         string     `json:"schedule_cron"`
	IsActive             bool       `json:"is_active"`
	LastRunAt            *time.Time `json:"last_run_at"`
	LastRunStatus        *string    `json:"last_run_status"`
	PassThresholdPercent float64    `json:"pass_threshold_percent"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	// Computed fields
	TestCaseCount        int        `json:"test_case_count,omitempty"`
}

// EvidenceTestCase represents a single test case within a test suite.
type EvidenceTestCase struct {
	ID                       uuid.UUID       `json:"id"`
	OrganizationID           uuid.UUID       `json:"organization_id"`
	TestSuiteID              uuid.UUID       `json:"test_suite_id"`
	Name                     string          `json:"name"`
	Description              string          `json:"description"`
	TestType                 string          `json:"test_type"`
	TargetControlCode        string          `json:"target_control_code"`
	TargetEvidenceTemplateID *uuid.UUID      `json:"target_evidence_template_id"`
	TestConfig               json.RawMessage `json:"test_config"`
	ExpectedResult           string          `json:"expected_result"`
	SortOrder                int             `json:"sort_order"`
	IsCritical               bool            `json:"is_critical"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
}

// EvidenceTestRun represents a single execution of a test suite.
type EvidenceTestRun struct {
	ID              uuid.UUID       `json:"id"`
	OrganizationID  uuid.UUID       `json:"organization_id"`
	TestSuiteID     uuid.UUID       `json:"test_suite_id"`
	Status          string          `json:"status"`
	StartedAt       *time.Time      `json:"started_at"`
	CompletedAt     *time.Time      `json:"completed_at"`
	TotalTests      int             `json:"total_tests"`
	Passed          int             `json:"passed"`
	Failed          int             `json:"failed"`
	Skipped         int             `json:"skipped"`
	Errors          int             `json:"errors"`
	PassRate        float64         `json:"pass_rate"`
	ThresholdMet    bool            `json:"threshold_met"`
	Results         json.RawMessage `json:"results"`
	TriggeredBy     string          `json:"triggered_by"`
	TriggeredByUser *uuid.UUID      `json:"triggered_by_user"`
	CreatedAt       time.Time       `json:"created_at"`
	// Joined fields
	SuiteName       string          `json:"suite_name,omitempty"`
}

// TestCaseResult holds the result of an individual test case execution.
type TestCaseResult struct {
	TestCaseID   uuid.UUID `json:"test_case_id"`
	TestCaseName string    `json:"test_case_name"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	IsCritical   bool      `json:"is_critical"`
	Duration     string    `json:"duration"`
}

// PreAuditReport provides a comprehensive pre-audit readiness assessment.
type PreAuditReport struct {
	ID                    uuid.UUID         `json:"id"`
	FrameworkCode         string            `json:"framework_code"`
	GeneratedAt           time.Time         `json:"generated_at"`
	OverallReadiness      float64           `json:"overall_readiness"`
	ReadinessLevel        string            `json:"readiness_level"`
	TotalControls         int               `json:"total_controls"`
	ControlsWithEvidence  int               `json:"controls_with_evidence"`
	ControlsMissingEvidence int             `json:"controls_missing_evidence"`
	EvidenceCompletion    float64           `json:"evidence_completion"`
	ValidationPassRate    float64           `json:"validation_pass_rate"`
	CriticalGaps          []PreAuditGap     `json:"critical_gaps"`
	ControlReadiness      []ControlReady    `json:"control_readiness"`
	Recommendations       []string          `json:"recommendations"`
	EstimatedRemediationHours float64       `json:"estimated_remediation_hours"`
}

// PreAuditGap represents a critical gap found during pre-audit checks.
type PreAuditGap struct {
	ControlCode   string `json:"control_code"`
	ControlName   string `json:"control_name"`
	GapType       string `json:"gap_type"`
	Severity      string `json:"severity"`
	Description   string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// ControlReady represents the readiness status of a single control.
type ControlReady struct {
	ControlCode      string  `json:"control_code"`
	ControlTitle     string  `json:"control_title"`
	EvidenceCount    int     `json:"evidence_count"`
	RequiredCount    int     `json:"required_count"`
	ValidationPassed int     `json:"validation_passed"`
	ReadinessPercent float64 `json:"readiness_percent"`
	Status           string  `json:"status"`
}

// ============================================================
// REQUEST TYPES
// ============================================================

// CreateTestSuiteRequest defines input for creating a new test suite.
type CreateTestSuiteRequest struct {
	Name                 string              `json:"name"`
	Description          string              `json:"description"`
	TestType             string              `json:"test_type"`
	ScheduleCron         string              `json:"schedule_cron"`
	PassThresholdPercent float64             `json:"pass_threshold_percent"`
	TestCases            []CreateTestCaseReq `json:"test_cases"`
}

// CreateTestCaseReq defines input for creating a single test case.
type CreateTestCaseReq struct {
	Name                     string          `json:"name"`
	Description              string          `json:"description"`
	TestType                 string          `json:"test_type"`
	TargetControlCode        string          `json:"target_control_code"`
	TargetEvidenceTemplateID *uuid.UUID      `json:"target_evidence_template_id"`
	TestConfig               json.RawMessage `json:"test_config"`
	ExpectedResult           string          `json:"expected_result"`
	SortOrder                int             `json:"sort_order"`
	IsCritical               bool            `json:"is_critical"`
}

// ============================================================
// SERVICE
// ============================================================

// EvidenceTestRunner provides functionality to manage and execute evidence
// test suites, run pre-audit checks, and generate readiness reports.
type EvidenceTestRunner struct {
	pool *pgxpool.Pool
}

// NewEvidenceTestRunner creates a new EvidenceTestRunner.
func NewEvidenceTestRunner(pool *pgxpool.Pool) *EvidenceTestRunner {
	return &EvidenceTestRunner{pool: pool}
}

// setRLS sets the RLS context for the current transaction.
func (r *EvidenceTestRunner) setRLS(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) error {
	_, err := tx.Exec(ctx, "SET LOCAL app.current_org = '"+orgID.String()+"'")
	return err
}

// ============================================================
// TEST SUITE MANAGEMENT
// ============================================================

// ListTestSuites returns all test suites for an organization.
func (r *EvidenceTestRunner) ListTestSuites(ctx context.Context, orgID uuid.UUID) ([]EvidenceTestSuite, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT ts.id, ts.organization_id, ts.name, ts.description, ts.test_type,
		       ts.schedule_cron, ts.is_active, ts.last_run_at, ts.last_run_status,
		       ts.pass_threshold_percent, ts.created_at, ts.updated_at,
		       (SELECT COUNT(*) FROM evidence_test_cases tc WHERE tc.test_suite_id = ts.id) AS test_case_count
		FROM evidence_test_suites ts
		WHERE ts.organization_id = $1
		ORDER BY ts.name ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("query test suites: %w", err)
	}
	defer rows.Close()

	var suites []EvidenceTestSuite
	for rows.Next() {
		var s EvidenceTestSuite
		if err := rows.Scan(
			&s.ID, &s.OrganizationID, &s.Name, &s.Description, &s.TestType,
			&s.ScheduleCron, &s.IsActive, &s.LastRunAt, &s.LastRunStatus,
			&s.PassThresholdPercent, &s.CreatedAt, &s.UpdatedAt,
			&s.TestCaseCount,
		); err != nil {
			return nil, fmt.Errorf("scan test suite: %w", err)
		}
		suites = append(suites, s)
	}
	if suites == nil {
		suites = []EvidenceTestSuite{}
	}
	return suites, nil
}

// CreateTestSuite creates a new test suite with optional test cases.
func (r *EvidenceTestRunner) CreateTestSuite(ctx context.Context, orgID uuid.UUID, req CreateTestSuiteRequest) (*EvidenceTestSuite, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("test suite name is required")
	}
	if req.TestType == "" {
		req.TestType = "on_demand"
	}
	if req.PassThresholdPercent <= 0 {
		req.PassThresholdPercent = 80.00
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	var suite EvidenceTestSuite
	err = tx.QueryRow(ctx, `
		INSERT INTO evidence_test_suites (
			organization_id, name, description, test_type,
			schedule_cron, pass_threshold_percent
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, organization_id, name, description, test_type,
		          schedule_cron, is_active, last_run_at, last_run_status,
		          pass_threshold_percent, created_at, updated_at`,
		orgID, req.Name, req.Description, req.TestType,
		req.ScheduleCron, req.PassThresholdPercent,
	).Scan(
		&suite.ID, &suite.OrganizationID, &suite.Name, &suite.Description,
		&suite.TestType, &suite.ScheduleCron, &suite.IsActive,
		&suite.LastRunAt, &suite.LastRunStatus,
		&suite.PassThresholdPercent, &suite.CreatedAt, &suite.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert test suite: %w", err)
	}

	// Create test cases if provided
	for i, tc := range req.TestCases {
		if tc.Name == "" {
			tc.Name = fmt.Sprintf("Test Case %d", i+1)
		}
		if tc.TestType == "" {
			tc.TestType = "exists"
		}
		if tc.ExpectedResult == "" {
			tc.ExpectedResult = "pass"
		}
		if tc.TestConfig == nil {
			tc.TestConfig = json.RawMessage(`{}`)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO evidence_test_cases (
				organization_id, test_suite_id, name, description,
				test_type, target_control_code, target_evidence_template_id,
				test_config, expected_result, sort_order, is_critical
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			orgID, suite.ID, tc.Name, tc.Description,
			tc.TestType, tc.TargetControlCode, tc.TargetEvidenceTemplateID,
			tc.TestConfig, tc.ExpectedResult, tc.SortOrder, tc.IsCritical,
		)
		if err != nil {
			return nil, fmt.Errorf("insert test case %d: %w", i, err)
		}
		suite.TestCaseCount++
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().Str("suite_id", suite.ID.String()).Str("name", suite.Name).Msg("test suite created")
	return &suite, nil
}

// ============================================================
// TEST EXECUTION
// ============================================================

// RunTestSuite executes all test cases in a test suite and records results.
func (r *EvidenceTestRunner) RunTestSuite(ctx context.Context, orgID, suiteID uuid.UUID, triggeredBy string) (*EvidenceTestRun, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	// Verify suite exists
	var suiteName string
	var passThreshold float64
	err = tx.QueryRow(ctx, `
		SELECT name, pass_threshold_percent FROM evidence_test_suites
		WHERE id = $1 AND organization_id = $2`,
		suiteID, orgID).Scan(&suiteName, &passThreshold)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("test suite not found")
		}
		return nil, fmt.Errorf("get suite: %w", err)
	}

	// Get test cases
	caseRows, err := tx.Query(ctx, `
		SELECT id, name, test_type, target_control_code,
		       target_evidence_template_id, test_config,
		       expected_result, is_critical
		FROM evidence_test_cases
		WHERE test_suite_id = $1 AND organization_id = $2
		ORDER BY sort_order ASC, created_at ASC`,
		suiteID, orgID)
	if err != nil {
		return nil, fmt.Errorf("query test cases: %w", err)
	}
	defer caseRows.Close()

	var cases []EvidenceTestCase
	for caseRows.Next() {
		var tc EvidenceTestCase
		if err := caseRows.Scan(
			&tc.ID, &tc.Name, &tc.TestType, &tc.TargetControlCode,
			&tc.TargetEvidenceTemplateID, &tc.TestConfig,
			&tc.ExpectedResult, &tc.IsCritical,
		); err != nil {
			return nil, fmt.Errorf("scan test case: %w", err)
		}
		cases = append(cases, tc)
	}

	// Create the test run record
	now := time.Now()
	run := &EvidenceTestRun{
		ID:             uuid.New(),
		OrganizationID: orgID,
		TestSuiteID:    suiteID,
		Status:         "running",
		StartedAt:      &now,
		TotalTests:     len(cases),
		TriggeredBy:    triggeredBy,
		SuiteName:      suiteName,
	}

	// Execute each test case
	var caseResults []TestCaseResult
	for _, tc := range cases {
		start := time.Now()
		result := r.executeTestCase(ctx, tx, orgID, tc)
		result.Duration = time.Since(start).String()
		caseResults = append(caseResults, result)

		switch result.Status {
		case "pass":
			run.Passed++
		case "fail":
			run.Failed++
		case "skip":
			run.Skipped++
		case "error":
			run.Errors++
		}
	}

	// Calculate pass rate
	if run.TotalTests > 0 {
		run.PassRate = float64(run.Passed) / float64(run.TotalTests) * 100
	}
	run.ThresholdMet = run.PassRate >= passThreshold
	run.Status = "completed"
	completedAt := time.Now()
	run.CompletedAt = &completedAt

	resultsJSON, _ := json.Marshal(caseResults)
	run.Results = resultsJSON

	// Insert the test run record
	_, err = tx.Exec(ctx, `
		INSERT INTO evidence_test_runs (
			id, organization_id, test_suite_id, status,
			started_at, completed_at, total_tests, passed, failed,
			skipped, errors, pass_rate, threshold_met, results,
			triggered_by, triggered_by_user
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		run.ID, orgID, suiteID, run.Status,
		run.StartedAt, run.CompletedAt, run.TotalTests, run.Passed, run.Failed,
		run.Skipped, run.Errors, run.PassRate, run.ThresholdMet, run.Results,
		run.TriggeredBy, run.TriggeredByUser,
	)
	if err != nil {
		return nil, fmt.Errorf("insert test run: %w", err)
	}

	// Update suite last run info
	statusStr := "completed"
	_, _ = tx.Exec(ctx, `
		UPDATE evidence_test_suites SET
			last_run_at = $1, last_run_status = $2, updated_at = NOW()
		WHERE id = $3 AND organization_id = $4`,
		run.CompletedAt, statusStr, suiteID, orgID)

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	log.Info().
		Str("run_id", run.ID.String()).
		Str("suite", suiteName).
		Int("passed", run.Passed).
		Int("failed", run.Failed).
		Float64("pass_rate", run.PassRate).
		Bool("threshold_met", run.ThresholdMet).
		Msg("test suite run completed")

	return run, nil
}

// executeTestCase runs a single test case and returns the result.
func (r *EvidenceTestRunner) executeTestCase(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, tc EvidenceTestCase) TestCaseResult {
	result := TestCaseResult{
		TestCaseID:   tc.ID,
		TestCaseName: tc.Name,
		IsCritical:   tc.IsCritical,
	}

	switch tc.TestType {
	case "exists":
		// Check if evidence exists for the target control or template
		var count int
		var err error
		if tc.TargetEvidenceTemplateID != nil {
			err = tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM evidence_requirements
				WHERE organization_id = $1 AND evidence_template_id = $2
				  AND status IN ('collected','validated')`,
				orgID, *tc.TargetEvidenceTemplateID).Scan(&count)
		} else if tc.TargetControlCode != "" {
			err = tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM evidence_requirements er
				JOIN evidence_templates et ON er.evidence_template_id = et.id
				WHERE er.organization_id = $1 AND et.framework_control_code = $2
				  AND er.status IN ('collected','validated')`,
				orgID, tc.TargetControlCode).Scan(&count)
		} else {
			result.Status = "skip"
			result.Message = "No target specified"
			return result
		}

		if err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("Query error: %v", err)
			return result
		}

		if count > 0 {
			result.Status = "pass"
			result.Message = fmt.Sprintf("Evidence found (%d items)", count)
		} else {
			result.Status = "fail"
			result.Message = "No collected evidence found"
		}

	case "not_empty":
		// Check that collected evidence has non-zero file size
		var count int
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM control_evidence ce
			JOIN evidence_requirements er ON ce.id = er.last_evidence_id
			WHERE er.organization_id = $1
			  AND ce.file_size_bytes > 0
			  AND er.evidence_template_id = $2`,
			orgID, tc.TargetEvidenceTemplateID).Scan(&count)
		if err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("Query error: %v", err)
			return result
		}
		if count > 0 {
			result.Status = "pass"
			result.Message = "Evidence files are non-empty"
		} else {
			result.Status = "fail"
			result.Message = "No non-empty evidence files found"
		}

	case "freshness_check":
		// Check that evidence was collected within required timeframe
		maxDays := 90
		var config map[string]interface{}
		if json.Unmarshal(tc.TestConfig, &config) == nil {
			if d, ok := config["max_days"]; ok {
				if v, ok := d.(float64); ok {
					maxDays = int(v)
				}
			}
		}

		var count int
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM evidence_requirements
			WHERE organization_id = $1
			  AND last_collected_at > NOW() - ($2 || ' days')::interval
			  AND evidence_template_id = $3`,
			orgID, fmt.Sprintf("%d", maxDays), tc.TargetEvidenceTemplateID).Scan(&count)
		if err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("Query error: %v", err)
			return result
		}
		if count > 0 {
			result.Status = "pass"
			result.Message = fmt.Sprintf("Evidence collected within %d days", maxDays)
		} else {
			result.Status = "fail"
			result.Message = fmt.Sprintf("No evidence collected within %d days", maxDays)
		}

	case "date_within":
		maxDays := 365
		var config map[string]interface{}
		if json.Unmarshal(tc.TestConfig, &config) == nil {
			if d, ok := config["max_days"]; ok {
				if v, ok := d.(float64); ok {
					maxDays = int(v)
				}
			}
		}

		var count int
		query := `
			SELECT COUNT(*) FROM evidence_requirements er
			JOIN evidence_templates et ON er.evidence_template_id = et.id
			WHERE er.organization_id = $1
			  AND er.last_collected_at > NOW() - ($2 || ' days')::interval`
		if tc.TargetControlCode != "" {
			query += fmt.Sprintf(" AND et.framework_control_code = '%s'", tc.TargetControlCode)
		}
		err := tx.QueryRow(ctx, query, orgID, fmt.Sprintf("%d", maxDays)).Scan(&count)
		if err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("Query error: %v", err)
			return result
		}
		if count > 0 {
			result.Status = "pass"
			result.Message = fmt.Sprintf("Evidence collected within %d days (%d items)", maxDays, count)
		} else {
			result.Status = "fail"
			result.Message = fmt.Sprintf("No evidence collected within %d days", maxDays)
		}

	case "approval_check":
		// Check that evidence has been reviewed/approved
		var count int
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM control_evidence ce
			WHERE ce.organization_id = $1
			  AND ce.review_status = 'accepted'
			  AND ce.control_implementation_id IS NOT NULL`,
			orgID).Scan(&count)
		if err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("Query error: %v", err)
			return result
		}
		if count > 0 {
			result.Status = "pass"
			result.Message = fmt.Sprintf("%d approved evidence items found", count)
		} else {
			result.Status = "fail"
			result.Message = "No approved evidence found"
		}

	case "coverage_check":
		// Check evidence coverage percentage for a framework
		var total, collected int
		var config map[string]interface{}
		frameworkCode := ""
		minCoverage := 80.0
		if json.Unmarshal(tc.TestConfig, &config) == nil {
			if fc, ok := config["framework_code"]; ok {
				if s, ok := fc.(string); ok {
					frameworkCode = s
				}
			}
			if mc, ok := config["min_coverage"]; ok {
				if v, ok := mc.(float64); ok {
					minCoverage = v
				}
			}
		}

		if frameworkCode == "" {
			result.Status = "skip"
			result.Message = "No framework_code in test config"
			return result
		}

		err := tx.QueryRow(ctx, `
			SELECT COUNT(*),
			       COUNT(*) FILTER (WHERE er.status IN ('collected','validated'))
			FROM evidence_requirements er
			JOIN evidence_templates et ON er.evidence_template_id = et.id
			WHERE er.organization_id = $1 AND et.framework_code = $2`,
			orgID, frameworkCode).Scan(&total, &collected)
		if err != nil {
			result.Status = "error"
			result.Message = fmt.Sprintf("Query error: %v", err)
			return result
		}

		coverage := 0.0
		if total > 0 {
			coverage = float64(collected) / float64(total) * 100
		}
		if coverage >= minCoverage {
			result.Status = "pass"
			result.Message = fmt.Sprintf("Coverage %.1f%% meets threshold %.1f%%", coverage, minCoverage)
		} else {
			result.Status = "fail"
			result.Message = fmt.Sprintf("Coverage %.1f%% below threshold %.1f%%", coverage, minCoverage)
		}

	default:
		result.Status = "skip"
		result.Message = fmt.Sprintf("Unknown test type: %s", tc.TestType)
	}

	return result
}

// ============================================================
// TEST RESULTS
// ============================================================

// GetTestRunResults returns the test run history for a test suite.
func (r *EvidenceTestRunner) GetTestRunResults(ctx context.Context, orgID, suiteID uuid.UUID) ([]EvidenceTestRun, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	rows, err := tx.Query(ctx, `
		SELECT tr.id, tr.organization_id, tr.test_suite_id, tr.status,
		       tr.started_at, tr.completed_at, tr.total_tests, tr.passed,
		       tr.failed, tr.skipped, tr.errors, tr.pass_rate,
		       tr.threshold_met, tr.results, tr.triggered_by,
		       tr.triggered_by_user, tr.created_at,
		       COALESCE(ts.name, '') AS suite_name
		FROM evidence_test_runs tr
		LEFT JOIN evidence_test_suites ts ON tr.test_suite_id = ts.id
		WHERE tr.test_suite_id = $1 AND tr.organization_id = $2
		ORDER BY tr.created_at DESC
		LIMIT 50`, suiteID, orgID)
	if err != nil {
		return nil, fmt.Errorf("query test runs: %w", err)
	}
	defer rows.Close()

	var runs []EvidenceTestRun
	for rows.Next() {
		var run EvidenceTestRun
		if err := rows.Scan(
			&run.ID, &run.OrganizationID, &run.TestSuiteID, &run.Status,
			&run.StartedAt, &run.CompletedAt, &run.TotalTests, &run.Passed,
			&run.Failed, &run.Skipped, &run.Errors, &run.PassRate,
			&run.ThresholdMet, &run.Results, &run.TriggeredBy,
			&run.TriggeredByUser, &run.CreatedAt,
			&run.SuiteName,
		); err != nil {
			return nil, fmt.Errorf("scan test run: %w", err)
		}
		runs = append(runs, run)
	}
	if runs == nil {
		runs = []EvidenceTestRun{}
	}
	return runs, nil
}

// ============================================================
// PRE-AUDIT CHECK
// ============================================================

// RunPreAuditChecks performs a comprehensive pre-audit readiness assessment
// for a specific framework.
func (r *EvidenceTestRunner) RunPreAuditChecks(ctx context.Context, orgID, frameworkID uuid.UUID) (*PreAuditReport, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.setRLS(ctx, tx, orgID); err != nil {
		return nil, fmt.Errorf("set rls: %w", err)
	}

	// Get framework code
	var frameworkCode string
	err = tx.QueryRow(ctx, `SELECT code FROM compliance_frameworks WHERE id = $1`, frameworkID).Scan(&frameworkCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("framework not found")
		}
		return nil, fmt.Errorf("get framework: %w", err)
	}

	report := &PreAuditReport{
		ID:            uuid.New(),
		FrameworkCode: frameworkCode,
		GeneratedAt:   time.Now(),
	}

	// Get unique control codes for this framework from templates
	controlRows, err := tx.Query(ctx, `
		SELECT DISTINCT et.framework_control_code
		FROM evidence_templates et
		WHERE et.framework_code = $1 AND et.is_system = true
		ORDER BY et.framework_control_code`, frameworkCode)
	if err != nil {
		return nil, fmt.Errorf("query controls: %w", err)
	}
	defer controlRows.Close()

	var controlCodes []string
	for controlRows.Next() {
		var code string
		if controlRows.Scan(&code) == nil {
			controlCodes = append(controlCodes, code)
		}
	}
	report.TotalControls = len(controlCodes)

	// Check evidence status for each control
	var totalRequired, totalCollected, totalValidated int
	for _, code := range controlCodes {
		var required, collected, validated int
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*) AS required,
			       COUNT(*) FILTER (WHERE er.status IN ('collected','validated')) AS collected,
			       COUNT(*) FILTER (WHERE er.validation_status = 'pass') AS validated
			FROM evidence_requirements er
			JOIN evidence_templates et ON er.evidence_template_id = et.id
			WHERE er.organization_id = $1 AND et.framework_control_code = $2 AND et.framework_code = $3`,
			orgID, code, frameworkCode).Scan(&required, &collected, &validated)
		if err != nil {
			continue
		}

		totalRequired += required
		totalCollected += collected
		totalValidated += validated

		readiness := 0.0
		if required > 0 {
			readiness = float64(collected) / float64(required) * 100
		}

		status := "complete"
		if readiness == 0 {
			status = "missing"
		} else if readiness < 100 {
			status = "partial"
		}

		// Get control title
		var controlTitle string
		_ = tx.QueryRow(ctx, `
			SELECT COALESCE(title, name) FROM compliance_controls
			WHERE code = $1 LIMIT 1`, code).Scan(&controlTitle)
		if controlTitle == "" {
			controlTitle = code
		}

		report.ControlReadiness = append(report.ControlReadiness, ControlReady{
			ControlCode:      code,
			ControlTitle:     controlTitle,
			EvidenceCount:    collected,
			RequiredCount:    required,
			ValidationPassed: validated,
			ReadinessPercent: readiness,
			Status:           status,
		})

		// Track controls with/without evidence
		if collected > 0 {
			report.ControlsWithEvidence++
		} else if required > 0 {
			report.ControlsMissingEvidence++
			report.CriticalGaps = append(report.CriticalGaps, PreAuditGap{
				ControlCode:    code,
				ControlName:    controlTitle,
				GapType:        "missing_evidence",
				Severity:       "high",
				Description:    fmt.Sprintf("No evidence collected for %d required items", required),
				Recommendation: fmt.Sprintf("Collect evidence for control %s before audit", code),
			})
		}
	}

	// Calculate overall metrics
	if totalRequired > 0 {
		report.EvidenceCompletion = float64(totalCollected) / float64(totalRequired) * 100
	}
	if totalCollected > 0 {
		report.ValidationPassRate = float64(totalValidated) / float64(totalCollected) * 100
	}
	report.OverallReadiness = report.EvidenceCompletion * 0.7 + report.ValidationPassRate * 0.3

	// Determine readiness level
	switch {
	case report.OverallReadiness >= 90:
		report.ReadinessLevel = "audit_ready"
	case report.OverallReadiness >= 70:
		report.ReadinessLevel = "mostly_ready"
	case report.OverallReadiness >= 50:
		report.ReadinessLevel = "significant_gaps"
	default:
		report.ReadinessLevel = "not_ready"
	}

	// Estimate remediation hours
	report.EstimatedRemediationHours = float64(report.ControlsMissingEvidence) * 4.0

	// Generate recommendations
	if report.OverallReadiness < 90 {
		report.Recommendations = append(report.Recommendations, "Focus on collecting evidence for controls with 'critical' and 'high' auditor priority first.")
	}
	if report.ValidationPassRate < 80 {
		report.Recommendations = append(report.Recommendations, "Review and fix evidence that failed validation. Common issues include outdated evidence and incomplete documentation.")
	}
	if report.ControlsMissingEvidence > 0 {
		report.Recommendations = append(report.Recommendations, fmt.Sprintf("Address %d controls with no evidence collected. Assign collection tasks to responsible team members.", report.ControlsMissingEvidence))
	}
	if len(report.CriticalGaps) > 10 {
		report.Recommendations = append(report.Recommendations, "Consider scheduling a focused evidence collection sprint to address the large number of gaps.")
	}
	if len(report.Recommendations) == 0 {
		report.Recommendations = append(report.Recommendations, "Evidence collection is in good shape. Continue monitoring for expiring evidence.")
	}

	if report.ControlReadiness == nil {
		report.ControlReadiness = []ControlReady{}
	}
	if report.CriticalGaps == nil {
		report.CriticalGaps = []PreAuditGap{}
	}

	log.Info().
		Str("framework", frameworkCode).
		Float64("readiness", report.OverallReadiness).
		Str("level", report.ReadinessLevel).
		Msg("pre-audit check completed")

	return report, nil
}
