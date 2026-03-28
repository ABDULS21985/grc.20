// ComplianceForge GRC Platform — API Server
//
// Enterprise Compliance Management Solution supporting:
// ISO/IEC 27001:2022, UK GDPR, NCSC CAF, Cyber Essentials,
// NIST SP 800-53, NIST CSF 2.0, PCI DSS v4.0, ITIL 4, COBIT 2019
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/pkg/cache"
	"github.com/complianceforge/platform/internal/pkg/crypto"
	"github.com/complianceforge/platform/internal/pkg/storage"
	"github.com/complianceforge/platform/internal/router"
)

func main() {
	// ============================================================
	// CONFIGURATION
	// ============================================================
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// ============================================================
	// LOGGING
	// ============================================================
	level, err := zerolog.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if cfg.IsDevelopment() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	log.Info().
		Str("app", cfg.App.Name).
		Str("version", cfg.App.Version).
		Str("env", cfg.App.Env).
		Int("port", cfg.App.Port).
		Msg("Starting ComplianceForge GRC Platform")

	// ============================================================
	// DATABASE
	// ============================================================
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// ============================================================
	// REDIS CACHE
	// ============================================================
	cacheClient, err := cache.New(cfg.Redis)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to Redis — caching disabled")
		// Non-fatal: platform works without cache, just slower
	} else {
		defer cacheClient.Close()
	}

	// ============================================================
	// ENCRYPTION
	// ============================================================
	encryptor, err := crypto.NewEncryptor(cfg.Security.EncryptionKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialise encryption")
	}
	_ = encryptor // Available for services that need to encrypt sensitive data

	// ============================================================
	// FILE STORAGE
	// ============================================================
	fileStorage, err := storage.NewStorage(cfg.Storage)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialise file storage")
	}
	_ = fileStorage // Available for evidence upload, policy documents, etc.

	// ============================================================
	// ROUTER
	// ============================================================
	r := router.New(cfg, db)

	// ============================================================
	// HTTP SERVER
	// ============================================================
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Info().Str("addr", addr).Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Log startup summary
	log.Info().
		Str("database", cfg.Database.Name).
		Str("redis", cfg.Redis.Addr()).
		Str("storage", cfg.Storage.Driver).
		Str("data_residency", cfg.Security.DataResidency).
		Msg("ComplianceForge is ready")

	// ============================================================
	// GRACEFUL SHUTDOWN
	// ============================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info().Str("signal", sig.String()).Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("ComplianceForge server stopped gracefully")
}
