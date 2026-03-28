// Package router configures all HTTP routes for the ComplianceForge API.
package router

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/complianceforge/platform/internal/config"
	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/handler"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/repository"
	"github.com/complianceforge/platform/internal/service"
)

func New(cfg *config.Config, db *database.DB) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.App.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Total-Count", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// ── Repositories ─────────────────────────────────────────
	frameworkRepo := repository.NewFrameworkRepo(db.Pool)
	riskRepo := repository.NewRiskRepo(db.Pool)
	policyRepo := repository.NewPolicyRepo(db.Pool)
	auditRepo := repository.NewAuditRepo(db.Pool)
	incidentRepo := repository.NewIncidentRepo(db.Pool)
	vendorRepo := repository.NewVendorRepo(db.Pool)
	assetRepo := repository.NewAssetRepo(db.Pool)
	userRepo := repository.NewUserRepo(db.Pool)
	orgRepo := repository.NewOrganizationRepo(db.Pool)

	// ── Services ─────────────────────────────────────────────
	authSvc := service.NewAuthService(db.Pool, cfg.JWT)
	complianceSvc := service.NewComplianceEngine(db.Pool)
	reportingSvc := service.NewReportingService(db.Pool)

	// ── Handlers ─────────────────────────────────────────────
	healthH := handler.NewHealthHandler(db, cfg.App.Version)
	authH := handler.NewAuthHandler(authSvc)
	frameworkH := handler.NewFrameworkHandler(frameworkRepo)
	riskH := handler.NewRiskHandler(riskRepo, db)
	policyH := handler.NewPolicyHandler(policyRepo, db)
	auditH := handler.NewAuditHandler(auditRepo, db)
	incidentH := handler.NewIncidentHandler(incidentRepo, db)
	vendorH := handler.NewVendorHandler(vendorRepo, db)
	assetH := handler.NewAssetHandler(assetRepo)
	dashH := handler.NewDashboardV2(complianceSvc)
	reportH := handler.NewReportHandler(reportingSvc)
	settingsH := handler.NewSettingsHandler(userRepo, orgRepo, authSvc)

	// ── Routes ───────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {

		// Public
		r.Get("/health", healthH.Health)
		r.Get("/ready", healthH.Ready)
		r.Post("/auth/login", authH.Login)

		// Authenticated
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWT))
			r.Use(middleware.Tenant(db))

			// ── Auth ─────────────────────────────────────
			r.Get("/auth/me", authH.Me)

			// ── Compliance Frameworks ────────────────────
			r.Route("/frameworks", func(r chi.Router) {
				r.Get("/", frameworkH.ListFrameworks)
				r.Get("/controls/search", frameworkH.SearchControls)
				r.Get("/{id}", frameworkH.GetFramework)
				r.Get("/{id}/controls", frameworkH.GetFrameworkControls)
			})

			// ── Compliance Engine ────────────────────────
			r.Get("/compliance/scores", frameworkH.GetComplianceScores)
			r.Get("/compliance/gaps", dashH.GapAnalysis)
			r.Get("/compliance/cross-mapping", dashH.CrossFrameworkCoverage)

			// ── Risk Management ──────────────────────────
			r.Route("/risks", func(r chi.Router) {
				r.Get("/", riskH.ListRisks)
				r.Post("/", riskH.CreateRisk)
				r.Get("/heatmap", riskH.GetHeatmap)
				r.Get("/{id}", riskH.GetRisk)
			})

			// ── Policy Management ────────────────────────
			r.Route("/policies", func(r chi.Router) {
				r.Get("/", policyH.ListPolicies)
				r.Post("/", policyH.CreatePolicy)
				r.Get("/attestations/stats", policyH.GetAttestationStats)
				r.Get("/{id}", policyH.GetPolicy)
				r.Post("/{id}/publish", policyH.PublishPolicy)
				r.Post("/{id}/attest", policyH.SubmitAttestation)
			})

			// ── Audit Management ─────────────────────────
			r.Route("/audits", func(r chi.Router) {
				r.Get("/", auditH.ListAudits)
				r.Post("/", auditH.CreateAudit)
				r.Get("/findings/stats", auditH.GetFindingsStats)
				r.Get("/{id}", auditH.GetAudit)
				r.Post("/{id}/findings", auditH.CreateFinding)
				r.Get("/{id}/findings", auditH.ListFindings)
			})

			// ── Incident Management ──────────────────────
			r.Route("/incidents", func(r chi.Router) {
				r.Get("/", incidentH.ListIncidents)
				r.Post("/", incidentH.ReportIncident)
				r.Get("/stats", incidentH.GetIncidentStats)
				r.Get("/breaches/urgent", incidentH.GetBreachesNearDeadline)
				r.Get("/{id}", incidentH.GetIncident)
				r.Post("/{id}/notify-dpa", incidentH.NotifyDPA)
				r.Post("/{id}/nis2-early-warning", incidentH.SubmitNIS2EarlyWarning)
			})

			// ── Vendor Management ────────────────────────
			r.Route("/vendors", func(r chi.Router) {
				r.Get("/", vendorH.ListVendors)
				r.Post("/", vendorH.OnboardVendor)
				r.Get("/stats", vendorH.GetVendorStats)
				r.Get("/{id}", vendorH.GetVendor)
			})

			// ── Asset Management ─────────────────────────
			r.Route("/assets", func(r chi.Router) {
				r.Get("/", assetH.ListAssets)
				r.Post("/", assetH.RegisterAsset)
				r.Get("/stats", assetH.GetAssetStats)
				r.Get("/{id}", assetH.GetAsset)
			})

			// ── Dashboard ────────────────────────────────
			r.Get("/dashboard/summary", dashH.Summary)

			// ── Reports ──────────────────────────────────
			r.Get("/reports/compliance", reportH.ComplianceReport)
			r.Get("/reports/risk", reportH.RiskReport)

			// ── Settings (Admin Only) ────────────────────
			r.Route("/settings", func(r chi.Router) {
				r.Use(middleware.RequireRole("org_admin"))
				r.Get("/", settingsH.GetOrganization)
				r.Put("/", settingsH.UpdateOrganization)
				r.Get("/users", settingsH.ListUsers)
				r.Post("/users", settingsH.CreateUser)
				r.Get("/users/{id}", settingsH.GetUser)
				r.Put("/users/{id}", settingsH.UpdateUser)
				r.Delete("/users/{id}", settingsH.DeactivateUser)
				r.Post("/users/{id}/roles", settingsH.AssignRole)
				r.Get("/roles", settingsH.ListRoles)
				r.Get("/audit-log", settingsH.GetAuditLog)
			})
		})
	})

	return r
}
