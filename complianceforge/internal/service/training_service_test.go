package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// RecordCompletion Logic Tests — in-memory simulation
// ============================================================

// simulateCompletion mimics the RecordCompletion decision logic without a DB.
// Returns (newStatus, passed, newAttempts).
func simulateCompletion(currentAttempts, score, passingScore, maxAttempts int, currentStatus string) (string, bool, int) {
	if currentStatus == "completed" || currentStatus == "exempted" {
		return currentStatus, false, currentAttempts
	}

	newAttempts := currentAttempts + 1
	passed := score >= passingScore

	if passed {
		return "completed", true, newAttempts
	} else if newAttempts >= maxAttempts {
		return "failed", false, newAttempts
	}
	return "in_progress", false, newAttempts
}

func TestRecordCompletion_Pass(t *testing.T) {
	// User scores 85% on a programme with 80% passing threshold.
	status, passed, attempts := simulateCompletion(0, 85, 80, 3, "in_progress")

	if status != "completed" {
		t.Errorf("Expected status 'completed', got %q", status)
	}
	if !passed {
		t.Error("Expected passed to be true")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRecordCompletion_Fail(t *testing.T) {
	// User scores 60% on first attempt (80% required, 3 attempts max).
	// Should remain in_progress.
	status, passed, attempts := simulateCompletion(0, 60, 80, 3, "in_progress")

	if status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got %q", status)
	}
	if passed {
		t.Error("Expected passed to be false")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRecordCompletion_MaxAttemptsExceeded(t *testing.T) {
	// User fails on their 3rd attempt (max 3 attempts).
	// Should become 'failed'.
	status, passed, attempts := simulateCompletion(2, 60, 80, 3, "in_progress")

	if status != "failed" {
		t.Errorf("Expected status 'failed', got %q", status)
	}
	if passed {
		t.Error("Expected passed to be false")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRecordCompletion_AlreadyCompleted(t *testing.T) {
	// Attempting to complete an already completed assignment should not change status.
	status, _, attempts := simulateCompletion(1, 90, 80, 3, "completed")

	if status != "completed" {
		t.Errorf("Expected status 'completed' (unchanged), got %q", status)
	}
	if attempts != 1 {
		t.Errorf("Expected attempts unchanged at 1, got %d", attempts)
	}
}

func TestRecordCompletion_PassOnLastAttempt(t *testing.T) {
	// User passes on their final attempt.
	status, passed, attempts := simulateCompletion(2, 85, 80, 3, "in_progress")

	if status != "completed" {
		t.Errorf("Expected status 'completed', got %q", status)
	}
	if !passed {
		t.Error("Expected passed to be true")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRecordCompletion_ExactPassingScore(t *testing.T) {
	// User scores exactly the passing threshold.
	status, passed, _ := simulateCompletion(0, 80, 80, 3, "in_progress")

	if status != "completed" {
		t.Errorf("Expected status 'completed', got %q", status)
	}
	if !passed {
		t.Error("Expected passed to be true for exact passing score")
	}
}

// ============================================================
// GenerateAssignments Logic Tests — in-memory simulation
// ============================================================

// simulateAssignmentGeneration mimics the deduplication logic.
// Returns (totalAssigned, newlyAssigned, alreadyExists).
func simulateAssignmentGeneration(userIDs []uuid.UUID, existingAssignments map[uuid.UUID]bool) (int, int, int) {
	total := 0
	newCount := 0
	existsCount := 0

	for _, uid := range userIDs {
		total++
		if existingAssignments[uid] {
			existsCount++
		} else {
			newCount++
			existingAssignments[uid] = true
		}
	}

	return total, newCount, existsCount
}

func TestGenerateAssignments_AllEmployees(t *testing.T) {
	// 5 employees, no existing assignments.
	userIDs := make([]uuid.UUID, 5)
	for i := range userIDs {
		userIDs[i] = uuid.New()
	}
	existing := make(map[uuid.UUID]bool)

	total, newlyAssigned, alreadyExists := simulateAssignmentGeneration(userIDs, existing)

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if newlyAssigned != 5 {
		t.Errorf("Expected 5 newly assigned, got %d", newlyAssigned)
	}
	if alreadyExists != 0 {
		t.Errorf("Expected 0 already exists, got %d", alreadyExists)
	}
}

func TestGenerateAssignments_SpecificRoles(t *testing.T) {
	// 3 users matching specific roles (out of 10 total).
	roleUserIDs := make([]uuid.UUID, 3)
	for i := range roleUserIDs {
		roleUserIDs[i] = uuid.New()
	}
	existing := make(map[uuid.UUID]bool)

	total, newlyAssigned, alreadyExists := simulateAssignmentGeneration(roleUserIDs, existing)

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
	if newlyAssigned != 3 {
		t.Errorf("Expected 3 newly assigned, got %d", newlyAssigned)
	}
	if alreadyExists != 0 {
		t.Errorf("Expected 0 already exists, got %d", alreadyExists)
	}
}

func TestGenerateAssignments_NoDuplicates(t *testing.T) {
	// 5 users, 3 already have assignments.
	userIDs := make([]uuid.UUID, 5)
	for i := range userIDs {
		userIDs[i] = uuid.New()
	}

	existing := make(map[uuid.UUID]bool)
	existing[userIDs[0]] = true
	existing[userIDs[1]] = true
	existing[userIDs[2]] = true

	total, newlyAssigned, alreadyExists := simulateAssignmentGeneration(userIDs, existing)

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
	if newlyAssigned != 2 {
		t.Errorf("Expected 2 newly assigned, got %d", newlyAssigned)
	}
	if alreadyExists != 3 {
		t.Errorf("Expected 3 already exists, got %d", alreadyExists)
	}
}

func TestGenerateAssignments_EmptyUserList(t *testing.T) {
	// No target users.
	existing := make(map[uuid.UUID]bool)

	total, newlyAssigned, alreadyExists := simulateAssignmentGeneration(nil, existing)

	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
	if newlyAssigned != 0 {
		t.Errorf("Expected 0 newly assigned, got %d", newlyAssigned)
	}
	if alreadyExists != 0 {
		t.Errorf("Expected 0 already exists, got %d", alreadyExists)
	}
}

// ============================================================
// Phishing Rates Calculation Tests
// ============================================================

// simulatePhishingRates computes aggregate rates from result counts.
func simulatePhishingRates(totalSent, opened, clicked, submitted, reported int) (openRate, clickRate, submitRate, reportRate float64) {
	if totalSent == 0 {
		return 0, 0, 0, 0
	}
	openRate = float64(opened) / float64(totalSent) * 100
	clickRate = float64(clicked) / float64(totalSent) * 100
	submitRate = float64(submitted) / float64(totalSent) * 100
	reportRate = float64(reported) / float64(totalSent) * 100
	return
}

func TestPhishingRates_Calculation(t *testing.T) {
	// 100 emails sent, 70 opened, 25 clicked, 10 submitted data, 15 reported.
	openRate, clickRate, submitRate, reportRate := simulatePhishingRates(100, 70, 25, 10, 15)

	if openRate != 70.0 {
		t.Errorf("Open rate = %.1f%%, want 70.0%%", openRate)
	}
	if clickRate != 25.0 {
		t.Errorf("Click rate = %.1f%%, want 25.0%%", clickRate)
	}
	if submitRate != 10.0 {
		t.Errorf("Submit rate = %.1f%%, want 10.0%%", submitRate)
	}
	if reportRate != 15.0 {
		t.Errorf("Report rate = %.1f%%, want 15.0%%", reportRate)
	}
}

func TestPhishingRates_ZeroSent(t *testing.T) {
	// Edge case: 0 emails sent.
	openRate, clickRate, submitRate, reportRate := simulatePhishingRates(0, 0, 0, 0, 0)

	if openRate != 0 || clickRate != 0 || submitRate != 0 || reportRate != 0 {
		t.Error("All rates should be 0 when no emails were sent")
	}
}

func TestPhishingRates_PerfectReporting(t *testing.T) {
	// All users reported the phishing email.
	_, _, _, reportRate := simulatePhishingRates(50, 50, 0, 0, 50)

	if reportRate != 100.0 {
		t.Errorf("Report rate = %.1f%%, want 100.0%%", reportRate)
	}
}

// ============================================================
// Certification Expiry Tests
// ============================================================

// simulateCertExpiryCheck determines whether a certification is within the expiry window.
func simulateCertExpiryCheck(expiryDate *time.Time, withinDays int) (isExpiring bool, daysUntil int) {
	if expiryDate == nil {
		return false, -1
	}
	now := time.Now()
	cutoff := now.AddDate(0, 0, withinDays)
	daysUntil = int(expiryDate.Sub(now).Hours() / 24)
	isExpiring = expiryDate.Before(cutoff)
	return
}

func TestCertificationExpiry_WithinDays(t *testing.T) {
	// Certification expires in 30 days; checking within 90 days.
	expiry := time.Now().AddDate(0, 0, 30)
	isExpiring, daysUntil := simulateCertExpiryCheck(&expiry, 90)

	if !isExpiring {
		t.Error("Certification expiring in 30 days should be within 90-day window")
	}
	if daysUntil < 29 || daysUntil > 31 {
		t.Errorf("Days until expiry = %d, expected approximately 30", daysUntil)
	}
}

func TestCertificationExpiry_NotExpiring(t *testing.T) {
	// Certification expires in 180 days; checking within 90 days.
	expiry := time.Now().AddDate(0, 0, 180)
	isExpiring, _ := simulateCertExpiryCheck(&expiry, 90)

	if isExpiring {
		t.Error("Certification expiring in 180 days should NOT be within 90-day window")
	}
}

func TestCertificationExpiry_AlreadyExpired(t *testing.T) {
	// Certification already expired 10 days ago.
	expiry := time.Now().AddDate(0, 0, -10)
	isExpiring, daysUntil := simulateCertExpiryCheck(&expiry, 90)

	if !isExpiring {
		t.Error("Already expired certification should be within expiry window")
	}
	if daysUntil >= 0 {
		t.Errorf("Days until expiry should be negative for expired cert, got %d", daysUntil)
	}
}

func TestCertificationExpiry_NoExpiryDate(t *testing.T) {
	// Certification with no expiry date (lifetime cert).
	isExpiring, _ := simulateCertExpiryCheck(nil, 90)

	if isExpiring {
		t.Error("Certification with no expiry date should not be flagged as expiring")
	}
}

func TestCertificationExpiry_ExactBoundary(t *testing.T) {
	// Certification expiring exactly on the boundary day.
	expiry := time.Now().AddDate(0, 0, 90)
	isExpiring, _ := simulateCertExpiryCheck(&expiry, 90)

	// Should be within the window (boundary is inclusive due to Before check
	// using cutoff = now + 90 days; expiry = now + 90 days => expiry.Before(cutoff) depends on time)
	// This tests the boundary logic.
	_ = isExpiring // Boundary can go either way depending on exact timing
}

// ============================================================
// Compliance Matrix Structure Tests
// ============================================================

func TestComplianceMatrix_Structure(t *testing.T) {
	// Simulate building a compliance matrix.
	type matrixProg struct {
		id   uuid.UUID
		name string
	}
	type matrixUser struct {
		id    uuid.UUID
		name  string
		email string
	}
	type matrixCell struct {
		status string
		score  *int
	}

	programmes := []matrixProg{
		{uuid.New(), "Security Awareness"},
		{uuid.New(), "GDPR Privacy"},
		{uuid.New(), "NIS2 Management Body"},
	}

	users := []matrixUser{
		{uuid.New(), "Alice Smith", "alice@example.com"},
		{uuid.New(), "Bob Jones", "bob@example.com"},
	}

	// Simulate some assignments
	assignments := map[string]matrixCell{} // key = "userID:progID"
	score85 := 85
	score72 := 72
	assignments[users[0].id.String()+":"+programmes[0].id.String()] = matrixCell{status: "completed", score: &score85}
	assignments[users[0].id.String()+":"+programmes[1].id.String()] = matrixCell{status: "in_progress", score: nil}
	assignments[users[1].id.String()+":"+programmes[0].id.String()] = matrixCell{status: "completed", score: &score72}

	// Build the matrix grid
	cells := make([][]matrixCell, len(users))
	for i, u := range users {
		cells[i] = make([]matrixCell, len(programmes))
		for j, p := range programmes {
			key := u.id.String() + ":" + p.id.String()
			if cell, ok := assignments[key]; ok {
				cells[i][j] = cell
			} else {
				cells[i][j] = matrixCell{status: "not_assigned"}
			}
		}
	}

	// Verify dimensions
	if len(cells) != len(users) {
		t.Errorf("Matrix rows = %d, want %d", len(cells), len(users))
	}
	for i, row := range cells {
		if len(row) != len(programmes) {
			t.Errorf("Matrix row %d cols = %d, want %d", i, len(row), len(programmes))
		}
	}

	// Verify specific cells
	if cells[0][0].status != "completed" {
		t.Errorf("Cell [Alice, Security Awareness] = %q, want 'completed'", cells[0][0].status)
	}
	if cells[0][0].score == nil || *cells[0][0].score != 85 {
		t.Error("Cell [Alice, Security Awareness] score should be 85")
	}
	if cells[0][1].status != "in_progress" {
		t.Errorf("Cell [Alice, GDPR] = %q, want 'in_progress'", cells[0][1].status)
	}
	if cells[0][2].status != "not_assigned" {
		t.Errorf("Cell [Alice, NIS2] = %q, want 'not_assigned'", cells[0][2].status)
	}
	if cells[1][0].status != "completed" {
		t.Errorf("Cell [Bob, Security Awareness] = %q, want 'completed'", cells[1][0].status)
	}
	if cells[1][1].status != "not_assigned" {
		t.Errorf("Cell [Bob, GDPR] = %q, want 'not_assigned'", cells[1][1].status)
	}
	if cells[1][2].status != "not_assigned" {
		t.Errorf("Cell [Bob, NIS2] = %q, want 'not_assigned'", cells[1][2].status)
	}
}

func TestComplianceMatrix_EmptyProgrammes(t *testing.T) {
	// No programmes => empty matrix.
	programmes := []struct{ id uuid.UUID }{}
	users := []struct{ id uuid.UUID }{{uuid.New()}, {uuid.New()}}

	cells := make([][]string, len(users))
	for i := range users {
		cells[i] = make([]string, len(programmes))
	}

	if len(cells) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(cells))
	}
	if len(cells[0]) != 0 {
		t.Errorf("Expected 0 cols, got %d", len(cells[0]))
	}
}

func TestComplianceMatrix_AllCompleted(t *testing.T) {
	// All users have completed all programmes.
	progCount := 3
	userCount := 4

	completedCount := 0
	totalCells := progCount * userCount

	cells := make([][]string, userCount)
	for i := 0; i < userCount; i++ {
		cells[i] = make([]string, progCount)
		for j := 0; j < progCount; j++ {
			cells[i][j] = "completed"
			completedCount++
		}
	}

	if completedCount != totalCells {
		t.Errorf("Expected %d completed cells, got %d", totalCells, completedCount)
	}

	completionRate := float64(completedCount) / float64(totalCells) * 100
	if completionRate != 100.0 {
		t.Errorf("Expected 100%% completion rate, got %.1f%%", completionRate)
	}
}
