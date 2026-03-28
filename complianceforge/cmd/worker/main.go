// ComplianceForge — Background Worker
//
// Runs scheduled tasks:
// - GDPR breach notification deadline monitoring
// - Policy review reminders
// - Audit finding escalation
// - Compliance score recalculation
// - KRI threshold monitoring
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
	"github.com/complianceforge/platform/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.IsDevelopment() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	log.Info().Msg("Starting ComplianceForge Background Worker")

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduled tasks
	go runScheduler(ctx, db)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Worker shutting down...")
	cancel()
	time.Sleep(2 * time.Second)
	log.Info().Msg("Worker stopped")
}

func runScheduler(ctx context.Context, db *database.DB) {
	// Run every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	runAllTasks(ctx, db)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runAllTasks(ctx, db)
		}
	}
}

func runAllTasks(ctx context.Context, db *database.DB) {
	log.Info().Msg("Running scheduled compliance tasks")

	checkBreachDeadlines(ctx, db)
	checkPolicyReviews(ctx, db)
	checkOverdueFindings(ctx, db)
	updatePolicyReviewStatus(ctx, db)

	log.Info().Msg("Scheduled tasks completed")
}

// checkBreachDeadlines finds GDPR breaches approaching the 72-hour notification deadline.
func checkBreachDeadlines(ctx context.Context, db *database.DB) {
	rows, err := db.Pool.Query(ctx, `
		SELECT incident_ref, notification_deadline, data_subjects_affected
		FROM incidents
		WHERE is_data_breach = true
			AND notification_required = true
			AND dpa_notified_at IS NULL
			AND notification_deadline <= NOW() + INTERVAL '6 hours'
			AND deleted_at IS NULL`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check breach deadlines")
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var ref string
		var deadline time.Time
		var affected int
		if err := rows.Scan(&ref, &deadline, &affected); err != nil {
			continue
		}
		log.Warn().
			Str("incident", ref).
			Time("deadline", deadline).
			Int("affected", affected).
			Msg("GDPR breach approaching notification deadline")
		count++
	}

	if count > 0 {
		log.Warn().Int("count", count).Msg("Breaches approaching 72-hour deadline")
	}
}

// checkPolicyReviews finds policies that are due or overdue for review.
func checkPolicyReviews(ctx context.Context, db *database.DB) {
	var due, overdue int64
	err := db.Pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE next_review_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'),
			COUNT(*) FILTER (WHERE next_review_date < CURRENT_DATE)
		FROM policies
		WHERE status = 'published' AND deleted_at IS NULL`,
	).Scan(&due, &overdue)

	if err != nil {
		log.Error().Err(err).Msg("Failed to check policy reviews")
		return
	}

	if due > 0 || overdue > 0 {
		log.Info().Int64("due_in_30_days", due).Int64("overdue", overdue).Msg("Policy reviews status")
	}
}

// checkOverdueFindings logs audit findings that have passed their due date.
func checkOverdueFindings(ctx context.Context, db *database.DB) {
	var count int64
	err := db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM audit_findings
		WHERE status IN ('open', 'in_progress')
			AND due_date < CURRENT_DATE
			AND deleted_at IS NULL`,
	).Scan(&count)

	if err != nil {
		log.Error().Err(err).Msg("Failed to check overdue findings")
		return
	}

	if count > 0 {
		log.Warn().Int64("count", count).Msg("Overdue audit findings requiring escalation")
	}
}

// updatePolicyReviewStatus updates policies with overdue reviews.
func updatePolicyReviewStatus(ctx context.Context, db *database.DB) {
	result, err := db.Pool.Exec(ctx, `
		UPDATE policies SET review_status = 'overdue'
		WHERE review_status = 'current'
			AND next_review_date < CURRENT_DATE
			AND status = 'published'
			AND deleted_at IS NULL`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update policy review status")
		return
	}

	if result.RowsAffected() > 0 {
		log.Info().Int64("updated", result.RowsAffected()).Msg("Policies marked as review overdue")
	}
}
