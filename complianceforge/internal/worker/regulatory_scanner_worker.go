package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/service"
)

// RegulatoryWorker is a background process that periodically scans
// regulatory RSS feeds based on each source's configured scan_frequency.
//
// Scheduling:
//   - Hourly:  sources with frequency='hourly' are scanned every hour.
//   - Daily:   sources with frequency='daily' are scanned at 06:00 UTC.
//   - Weekly:  sources with frequency='weekly' are scanned Monday at 06:00 UTC.
type RegulatoryWorker struct {
	scanner *service.RegulatoryScanner
	pool    *pgxpool.Pool
}

// NewRegulatoryWorker creates a new RegulatoryWorker.
func NewRegulatoryWorker(scanner *service.RegulatoryScanner, pool *pgxpool.Pool) *RegulatoryWorker {
	return &RegulatoryWorker{
		scanner: scanner,
		pool:    pool,
	}
}

// Start begins the background scan loop. It blocks until ctx is cancelled.
//
// The loop wakes every minute to check whether any scan is due:
//   - Hourly scans fire on every whole hour.
//   - Daily scans fire once per day at 06:00 UTC.
//   - Weekly scans fire on Monday at 06:00 UTC.
func (w *RegulatoryWorker) Start(ctx context.Context) {
	log.Info().Msg("Regulatory scanner worker started")

	// Run an initial hourly scan on startup
	go func() {
		if err := w.RunScanCycle(ctx, "hourly"); err != nil {
			log.Error().Err(err).Msg("initial hourly scan failed")
		}
	}()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	var lastHourlyScan time.Time
	var lastDailyScan time.Time
	var lastWeeklyScan time.Time

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Regulatory scanner worker shutting down")
			return

		case now := <-ticker.C:
			utcNow := now.UTC()

			// Hourly: fire if we haven't scanned in the current hour
			currentHour := utcNow.Truncate(time.Hour)
			if lastHourlyScan.Before(currentHour) {
				lastHourlyScan = utcNow
				go func() {
					if err := w.RunScanCycle(ctx, "hourly"); err != nil {
						log.Error().Err(err).Msg("hourly scan cycle failed")
					}
				}()
			}

			// Daily: fire at 06:xx UTC if we haven't scanned today
			currentDay := time.Date(utcNow.Year(), utcNow.Month(), utcNow.Day(), 0, 0, 0, 0, time.UTC)
			if utcNow.Hour() == 6 && lastDailyScan.Before(currentDay) {
				lastDailyScan = utcNow
				log.Info().Msg("triggering daily regulatory scan")
				go func() {
					if err := w.RunScanCycle(ctx, "daily"); err != nil {
						log.Error().Err(err).Msg("daily scan cycle failed")
					}
				}()
			}

			// Weekly: fire Monday at 06:xx UTC if we haven't scanned this week
			currentWeekMonday := currentDay.AddDate(0, 0, -int(utcNow.Weekday()-time.Monday))
			if utcNow.Weekday() == time.Monday && utcNow.Hour() == 6 && lastWeeklyScan.Before(currentWeekMonday) {
				lastWeeklyScan = utcNow
				log.Info().Msg("triggering weekly regulatory scan")
				go func() {
					if err := w.RunScanCycle(ctx, "weekly"); err != nil {
						log.Error().Err(err).Msg("weekly scan cycle failed")
					}
				}()
			}
		}
	}
}

// RunScanCycle scans all active regulatory sources that match the given
// frequency ('hourly', 'daily', or 'weekly'). Each source is scanned
// individually so that a failure in one does not block the others.
func (w *RegulatoryWorker) RunScanCycle(ctx context.Context, frequency string) error {
	log.Info().Str("frequency", frequency).Msg("starting regulatory scan cycle")
	startTime := time.Now()

	rows, err := w.pool.Query(ctx, `
		SELECT id, name
		FROM regulatory_sources
		WHERE is_active = true
		  AND scan_frequency = $1
		  AND rss_feed_url IS NOT NULL
		  AND rss_feed_url != ''
		ORDER BY name`, frequency)
	if err != nil {
		return fmt.Errorf("query %s sources: %w", frequency, err)
	}
	defer rows.Close()

	type sourceInfo struct {
		id   uuid.UUID
		name string
	}
	var sources []sourceInfo
	for rows.Next() {
		var si sourceInfo
		if err := rows.Scan(&si.id, &si.name); err != nil {
			log.Error().Err(err).Msg("scan source info row")
			continue
		}
		sources = append(sources, si)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate %s sources: %w", frequency, err)
	}

	if len(sources) == 0 {
		log.Debug().Str("frequency", frequency).Msg("no sources to scan")
		return nil
	}

	var successCount, errorCount int
	var scanErrors []string
	for _, src := range sources {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result, err := w.scanner.ScanSource(ctx, src.id)
		if err != nil {
			errorCount++
			scanErrors = append(scanErrors, fmt.Sprintf("%s: %v", src.name, err))
			log.Error().Err(err).Str("source", src.name).Msg("source scan failed")
			continue
		}

		successCount++
		log.Info().
			Str("source", src.name).
			Int("entries", result.EntriesFound).
			Int("new", result.NewChanges).
			Int("dupes", result.Duplicates).
			Msg("source scan completed")
	}

	duration := time.Since(startTime)
	log.Info().
		Str("frequency", frequency).
		Int("total_sources", len(sources)).
		Int("success", successCount).
		Int("errors", errorCount).
		Dur("duration", duration).
		Msg("regulatory scan cycle complete")

	if len(scanErrors) > 0 {
		return fmt.Errorf("%d/%d source scans failed: %s",
			errorCount, len(sources), joinMax(scanErrors, 5))
	}
	return nil
}

// joinMax joins at most n strings with "; ".
func joinMax(ss []string, n int) string {
	if len(ss) <= n {
		result := ""
		for i, s := range ss {
			if i > 0 {
				result += "; "
			}
			result += s
		}
		return result
	}
	result := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			result += "; "
		}
		result += ss[i]
	}
	return fmt.Sprintf("%s (and %d more)", result, len(ss)-n)
}
