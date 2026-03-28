// ComplianceForge — Database Migration Runner
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	direction := flag.String("direction", "up", "Migration direction: up or down")
	steps := flag.Int("steps", 0, "Number of migrations to run (0 = all)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	dsn := cfg.Database.DSN()

	log.Info().
		Str("direction", *direction).
		Int("steps", *steps).
		Str("database", cfg.Database.Name).
		Msg("Running database migrations")

	// In production, use golang-migrate CLI directly:
	// migrate -path sql/migrations -database "postgres://..." up
	fmt.Printf("Migration DSN: %s\n", dsn)
	fmt.Printf("Direction: %s, Steps: %d\n", *direction, *steps)
	fmt.Println("Use 'make migrate-up' or 'make migrate-down' for migrations.")
}
