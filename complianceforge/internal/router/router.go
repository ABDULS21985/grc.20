// Package router configures all HTTP routes for the ComplianceForge API.
package router

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/handler"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/pkg/crypto"
	"github.com/complianceforge/platform/internal/pkg/queue"
	"github.com/complianceforge/platform/internal/pkg/storage"
	"github.com/complianceforge/platform/internal/repository"
	"github.com/complianceforge/platform/internal/service"
	"github.com/complianceforge/platform/internal/worker"
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
	onboardingSvc := service.NewOnboardingService(db.Pool)

	// Batch 3: Queue (Redis-backed)
	var jobQueue *queue.Queue
	q, err := queue.NewFromAddr(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create job queue — background jobs disabled")
	} else {
		jobQueue = q
	}

	// Batch 3: File storage
	fileStore, err := storage.NewStorage(cfg.Storage)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create file storage — report downloads disabled")
	}

	// Batch 3: Notification Engine
	eventBus := service.NewEventBus(256)
	notificationEngine := service.NewNotificationEngine(db.Pool, eventBus)

	// Batch 3: Advanced Report Engine
	reportEngine := service.NewReportEngine(db.Pool, fileStore)

	// Batch 3: DSR Service (with PII encryption)
	encryptor, err := crypto.NewEncryptor(cfg.Security.EncryptionKey)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create encryptor — DSR PII encryption disabled")
	}
	dsrSvc := service.NewDSRService(db.Pool, encryptor)

	// Batch 3: NIS2 Service
	nis2Svc := service.NewNIS2Service(db.Pool)

	// Batch 3: Continuous Monitoring
	evidenceCollector := service.NewEvidenceCollector(db.Pool, jobQueue)
	complianceMonitor := service.NewComplianceMonitor(db.Pool, jobQueue)
	driftDetector := service.NewDriftDetector(db.Pool, jobQueue)

	// Batch 4: Workflow Engine
	workflowEngine := service.NewWorkflowEngine(db.Pool)

	// Batch 4: Integration Hub
	integrationSvc := service.NewIntegrationService(db.Pool, cfg.Security.EncryptionKey)

	// Batch 4: Onboarding Wizard & Subscriptions
	onboardingWizard := service.NewOnboardingWizard(db.Pool)
	subscriptionSvc := service.NewSubscriptionService(db.Pool)

	// Batch 4: ABAC Engine
	abacEngine := service.NewABACEngine(db.Pool)
	fieldMasker := service.NewFieldMasker(db.Pool, abacEngine)

	// Batch 5: AI Service & Remediation Planner
	aiService := service.NewAIService(db.Pool, cfg.Security.AIAPIKey, cfg.Security.AIModel)
	remediationPlanner := service.NewRemediationPlanner(db.Pool, aiService)

	// Batch 5: Control Library Marketplace
	marketplaceSvc := service.NewMarketplaceService(db.Pool)
	packageBuilder := service.NewPackageBuilder(db.Pool)

	// Batch 5: Regulatory Change Management
	regulatoryScanner := service.NewRegulatoryScanner(db.Pool)
	regulatorySvc := service.NewRegulatoryService(db.Pool, regulatoryScanner)

	// Batch 5: BIA & Business Continuity
	biaSvc := service.NewBIAService(db.Pool)
	continuitySvc := service.NewContinuityService(db.Pool)

	// Batch 5: Advanced Analytics
	analyticsEngine := service.NewAnalyticsEngine(db.Pool)
	analyticsQuery := service.NewAnalyticsQuery(db.Pool)
	dashboardSvc := service.NewDashboardService(db.Pool)

	// Batch 7: Compliance Calendar
	calendarSvc := service.NewCalendarService(db.Pool)

	// Batch 7: Search & Knowledge Base
	searchEngine := service.NewSearchEngine(db.Pool)
	knowledgeSvc := service.NewKnowledgeService(db.Pool)

	// Batch 7: Collaboration & Activity Feed
	collaborationSvc := service.NewCollaborationService(db.Pool)

	// Batch 7: Push Notifications & Mobile API
	pushSvc := service.NewPushService(db.Pool, &service.StubPushProvider{})

	// Batch 7: Push Integration (bridges EventBus → PushService)
	pushIntegration := service.NewPushIntegration(db.Pool, pushSvc, eventBus)
	_ = pushIntegration // started via eventBus subscriptions internally

	// Batch 7: Branding & White-Labelling
	brandingSvc := service.NewBrandingService(db.Pool)

	// Batch 6: Exception Management
	exceptionSvc := service.NewExceptionService(db.Pool)

	// Batch 6: Board Reporting & Governance
	boardSvc := service.NewBoardService(db.Pool)

	// Batch 6: Data Classification, Data Mapping & ROPA
	dataClassSvc := service.NewDataClassificationService(db.Pool)
	ropaSvc := service.NewROPAService(db.Pool)

	// Batch 6: TPRM — Questionnaires & Vendor Assessments
	questionnaireSvc := service.NewQuestionnaireService(db.Pool)
	vendorAssessmentSvc := service.NewVendorAssessmentService(db.Pool)

	// Batch 6: Evidence Template Library & Automated Evidence Testing
	evidenceTemplateSvc := service.NewEvidenceTemplateService(db.Pool)
	evidenceTestRunner := service.NewEvidenceTestRunner(db.Pool)

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
	onboardH := handler.NewOnboardingHandler(onboardingSvc)
	controlH := handler.NewControlHandler(db)

	// Batch 3 Handlers
	notifH := handler.NewNotificationHandler(notificationEngine, db.Pool)
	reportV2H := handler.NewReportHandlerV2(reportEngine)
	dsrH := handler.NewDSRHandler(dsrSvc)
	nis2H := handler.NewNIS2Handler(nis2Svc)
	monitorH := handler.NewMonitoringHandler(evidenceCollector, complianceMonitor, driftDetector)

	// Batch 4 Handlers
	workflowH := handler.NewWorkflowHandler(workflowEngine)
	integrationH := handler.NewIntegrationHandler(integrationSvc, cfg.App.BaseURL)
	onboardH = handler.NewOnboardingHandlerWithWizard(onboardingSvc, onboardingWizard)
	subscriptionH := handler.NewSubscriptionHandler(subscriptionSvc)
	accessH := handler.NewAccessHandler(abacEngine, fieldMasker)

	// Batch 5 Handlers
	remediationH := handler.NewRemediationHandler(remediationPlanner, aiService)
	marketplaceH := handler.NewMarketplaceHandler(marketplaceSvc, packageBuilder)
	regulatoryH := handler.NewRegulatoryHandler(regulatorySvc)
	biaH := handler.NewBIAHandler(biaSvc, continuitySvc)
	analyticsH := handler.NewAnalyticsHandler(analyticsEngine, analyticsQuery, dashboardSvc)

	// Batch 7 Handlers
	calendarH := handler.NewCalendarHandler(calendarSvc)
	searchH := handler.NewSearchHandler(searchEngine, knowledgeSvc)
	collaborationH := handler.NewCollaborationHandler(collaborationSvc)
	mobileH := handler.NewMobileHandler(db.Pool, pushSvc)
	brandingH := handler.NewBrandingHandler(brandingSvc)

	// Batch 6 Handlers
	exceptionH := handler.NewExceptionHandler(exceptionSvc)
	boardH := handler.NewBoardHandler(boardSvc)
	boardPortalH := handler.NewBoardPortalHandler(boardSvc)

	// Batch 6 Handlers (continued)
	ropaH := handler.NewROPAHandler(dataClassSvc, ropaSvc)
	questionnaireH := handler.NewQuestionnaireHandler(questionnaireSvc, vendorAssessmentSvc)
	vendorPortalH := handler.NewVendorPortalHandler(vendorAssessmentSvc, questionnaireSvc)
	evidenceTemplateH := handler.NewEvidenceTemplateHandler(evidenceTemplateSvc, evidenceTestRunner)

	// ── Background Workers ─────────────────────────────────
	calendarWorker := worker.NewCalendarWorker(calendarSvc, db.Pool)
	_ = calendarWorker // started by main goroutine

	searchIndexer := worker.NewSearchIndexer(db.Pool, searchEngine)
	_ = searchIndexer // started by main goroutine

	analyticsScheduler := worker.NewAnalyticsScheduler(analyticsEngine, db.Pool)
	_ = analyticsScheduler // started by main goroutine

	// ── Routes ───────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {

		// Public
		r.Get("/health", healthH.Health)
		r.Get("/ready", healthH.Ready)
		r.Post("/auth/login", authH.Login)
		r.Post("/onboard", onboardH.Onboard)

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
				r.Get("/{id}/implementations", controlH.ListImplementations)
			})

			// ── Control Implementations ─────────────────
			r.Route("/controls", func(r chi.Router) {
				r.Get("/{id}", controlH.GetImplementation)
				r.Put("/{id}", controlH.UpdateImplementation)
				r.Post("/{id}/test", controlH.RecordTestResult)
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

			// ── Reports (Legacy) ────────────────────────
			r.Get("/reports/compliance", reportH.ComplianceReport)
			r.Get("/reports/risk", reportH.RiskReport)

			// ── Reports (Advanced — Batch 3) ────────────
			r.Route("/reports", func(r chi.Router) {
				r.Post("/generate", reportV2H.GenerateReport)
				r.Get("/history", reportV2H.ListHistory)
				r.Get("/runs/{id}", reportV2H.GetReportStatus)
				r.Get("/download/{id}", reportV2H.DownloadReport)
				r.Route("/definitions", func(r chi.Router) {
					r.Get("/", reportV2H.ListDefinitions)
					r.Post("/", reportV2H.CreateDefinition)
					r.Put("/{id}", reportV2H.UpdateDefinition)
					r.Delete("/{id}", reportV2H.DeleteDefinition)
					r.Post("/{id}/generate", reportV2H.GenerateFromDefinition)
				})
				r.Route("/schedules", func(r chi.Router) {
					r.Get("/", reportV2H.ListSchedules)
					r.Post("/", reportV2H.CreateSchedule)
					r.Put("/{id}", reportV2H.UpdateSchedule)
					r.Delete("/{id}", reportV2H.DeleteSchedule)
				})
			})

			// ── Notifications (Batch 3) ─────────────────
			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", notifH.ListNotifications)
				r.Get("/unread-count", notifH.GetUnreadCount)
				r.Put("/read-all", notifH.MarkAllRead)
				r.Put("/{id}/read", notifH.MarkRead)
				r.Get("/preferences", notifH.GetPreferences)
				r.Put("/preferences", notifH.UpdatePreferences)
				r.Route("/rules", func(r chi.Router) {
					r.Get("/", notifH.ListRules)
					r.Post("/", notifH.CreateRule)
					r.Put("/{id}", notifH.UpdateRule)
					r.Delete("/{id}", notifH.DeleteRule)
				})
				r.Route("/channels", func(r chi.Router) {
					r.Get("/", notifH.ListChannels)
					r.Post("/", notifH.CreateChannel)
					r.Post("/{id}/test", notifH.TestChannel)
				})
			})

			// ── GDPR DSR Management (Batch 3) ───────────
			r.Route("/dsr", func(r chi.Router) {
				r.Get("/", dsrH.ListRequests)
				r.Post("/", dsrH.CreateRequest)
				r.Get("/dashboard", dsrH.GetDashboard)
				r.Get("/overdue", dsrH.GetOverdue)
				r.Put("/tasks/{id}", dsrH.UpdateTask)
				r.Get("/{id}", dsrH.GetRequest)
				r.Put("/{id}", dsrH.UpdateRequest)
				r.Post("/{id}/verify", dsrH.VerifyIdentity)
				r.Post("/{id}/assign", dsrH.AssignRequest)
				r.Post("/{id}/extend", dsrH.ExtendDeadline)
				r.Post("/{id}/complete", dsrH.CompleteRequest)
				r.Post("/{id}/reject", dsrH.RejectRequest)
				r.Get("/{id}/tasks", dsrH.GetTasks)
				r.Get("/{id}/audit-trail", dsrH.GetAuditTrail)
			})

			// ── NIS2 Compliance (Batch 3) ────────────────
			r.Route("/nis2", func(r chi.Router) {
				r.Get("/dashboard", nis2H.GetDashboard)
				r.Get("/assessment", nis2H.GetAssessment)
				r.Post("/assessment", nis2H.SubmitAssessment)
				r.Route("/incidents", func(r chi.Router) {
					r.Get("/", nis2H.ListIncidentReports)
					r.Get("/{id}", nis2H.GetIncidentReport)
					r.Post("/{id}/early-warning", nis2H.SubmitEarlyWarning)
					r.Post("/{id}/notification", nis2H.SubmitNotification)
					r.Post("/{id}/final-report", nis2H.SubmitFinalReport)
				})
				r.Route("/measures", func(r chi.Router) {
					r.Get("/", nis2H.ListMeasures)
					r.Put("/{id}", nis2H.UpdateMeasure)
				})
				r.Route("/management", func(r chi.Router) {
					r.Get("/", nis2H.ListManagement)
					r.Post("/", nis2H.RecordManagement)
				})
			})

			// ── Continuous Monitoring (Batch 3) ──────────
			r.Route("/monitoring", func(r chi.Router) {
				r.Get("/dashboard", monitorH.GetDashboard)
				r.Route("/evidence", func(r chi.Router) {
					r.Get("/", monitorH.ListCollectionConfigs)
					r.Post("/", monitorH.CreateCollectionConfig)
					r.Put("/{id}", monitorH.UpdateCollectionConfig)
					r.Post("/{id}/run", monitorH.RunCollectionNow)
					r.Get("/{id}/runs", monitorH.ListCollectionRuns)
				})
				r.Route("/monitors", func(r chi.Router) {
					r.Get("/", monitorH.ListMonitors)
					r.Post("/", monitorH.CreateMonitor)
					r.Put("/{id}", monitorH.UpdateMonitor)
					r.Get("/{id}/results", monitorH.GetMonitorResults)
				})
				r.Route("/drift", func(r chi.Router) {
					r.Get("/", monitorH.ListDriftEvents)
					r.Post("/{id}/acknowledge", monitorH.AcknowledgeDrift)
					r.Post("/{id}/resolve", monitorH.ResolveDrift)
				})
			})

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

			// ── Workflow Engine (Batch 4) ────────────────
			workflowH.RegisterRoutes(r)

			// ── Integration Hub (Batch 4) ────────────────
			integrationH.RegisterRoutes(r)

			// ── Onboarding Wizard (Batch 4) ──────────────
			r.Get("/onboard/progress", onboardH.GetProgress)
			r.Put("/onboard/step/{n}", onboardH.SaveStep)
			r.Post("/onboard/step/{n}/skip", onboardH.SkipStep)
			r.Post("/onboard/complete", onboardH.CompleteOnboarding)
			r.Get("/onboard/recommendations", onboardH.GetRecommendations)

			// ── Subscription Management (Batch 4) ────────
			r.Route("/subscription", func(r chi.Router) {
				r.Get("/", subscriptionH.GetSubscription)
				r.Put("/plan", subscriptionH.ChangePlan)
				r.Post("/cancel", subscriptionH.CancelSubscription)
				r.Post("/pause", subscriptionH.PauseSubscription)
				r.Post("/resume", subscriptionH.ResumeSubscription)
				r.Get("/plans", subscriptionH.ListPlans)
				r.Get("/usage", subscriptionH.GetUsage)
				r.Post("/create", subscriptionH.CreateSubscription)
			})

			// ── Access Control / ABAC (Batch 4) ──────────
			r.Route("/access", func(r chi.Router) {
				r.Get("/policies", accessH.ListPolicies)
				r.Post("/policies", accessH.CreatePolicy)
				r.Put("/policies/{id}", accessH.UpdatePolicy)
				r.Delete("/policies/{id}", accessH.DeletePolicy)
				r.Post("/policies/{id}/assignments", accessH.CreateAssignment)
				r.Delete("/policies/{id}/assignments/{assignmentId}", accessH.DeleteAssignment)
				r.Post("/evaluate", accessH.EvaluateAccess)
				r.Get("/audit-log", accessH.ListAuditLog)
				r.Get("/my-permissions", accessH.GetMyPermissions)
				r.Get("/field-permissions", accessH.GetFieldPermissions)
			})

			// ── AI Remediation Planner (Batch 5) ────────
			r.Route("/remediation", func(r chi.Router) {
				remediationH.RegisterRoutes(r)
			})

			// ── Control Library Marketplace (Batch 5) ───
			r.Route("/marketplace", func(r chi.Router) {
				marketplaceH.RegisterRoutes(r)
			})

			// ── Regulatory Change Management (Batch 5) ──
			r.Route("/regulatory", func(r chi.Router) {
				regulatoryH.RegisterRoutes(r)
			})

			// ── BIA & Business Continuity (Batch 5) ─────
			biaH.RegisterRoutes(r)

			// ── Advanced Analytics (Batch 5) ────────────
			r.Route("/analytics", func(r chi.Router) {
				analyticsH.RegisterRoutes(r)
			})

			// ── Exception Management (Batch 6) ──────────
			r.Route("/exceptions", func(r chi.Router) {
				exceptionH.RegisterRoutes(r)
			})

			// ── Board Reporting & Governance (Batch 6) ──
			r.Route("/board", func(r chi.Router) {
				boardH.RegisterRoutes(r)
			})

			// ── Data Classification, Mapping & ROPA (Batch 6) ──
			r.Route("/data", func(r chi.Router) {
				ropaH.RegisterRoutes(r)
			})

			// ── Evidence Template Library & Testing (Batch 6) ──
			r.Route("/evidence", func(r chi.Router) {
				evidenceTemplateH.RegisterRoutes(r)
			})

			// ── TPRM — Questionnaires & Vendor Assessments (Batch 6) ──
			questionnaireH.RegisterRoutes(r)

			// ── Compliance Calendar (Batch 7) ───────────
			r.Route("/calendar", func(r chi.Router) {
				calendarH.RegisterRoutes(r)
			})

			// ── Search & Knowledge Base (Batch 7) ───────
			searchH.RegisterRoutes(r)

			// ── Collaboration & Activity Feed (Batch 7) ─
			collaborationH.RegisterRoutes(r)

			// ── Mobile API & Push Notifications (Batch 7)
			r.Route("/mobile", func(r chi.Router) {
				mobileH.RegisterRoutes(r)
			})

			// ── Branding & White-Labelling (Batch 7) ────
			brandingH.RegisterRoutes(r)
		})

		// Public — Board Portal (token-based, no JWT)
		r.Route("/board-portal", func(r chi.Router) {
			boardPortalH.RegisterRoutes(r)
		})

		// Public — Vendor Assessment Portal (token-based, no JWT)
		r.Route("/vendor-portal", func(r chi.Router) {
			vendorPortalH.RegisterRoutes(r)
		})

		// Public — Calendar iCal Feed (token-based, no JWT)
		r.Route("/calendar", func(r chi.Router) {
			calendarH.RegisterPublicRoutes(r)
		})
	})

	return r
}
