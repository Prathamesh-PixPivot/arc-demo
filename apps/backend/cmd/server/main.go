package main

import (
	"os"
	"pixpivot/arc/config"
	"pixpivot/arc/internal/api/handlers"
	"pixpivot/arc/internal/api/middleware"
	"pixpivot/arc/internal/auth"
	"pixpivot/arc/internal/db"
	"pixpivot/arc/internal/licensing"
	"pixpivot/arc/internal/realtime"
	"pixpivot/arc/internal/storage/repository"

	"fmt"
	"net/http"
	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/pkg/encryption"
	"pixpivot/arc/pkg/jwtlink"
	"pixpivot/arc/pkg/log"

	"github.com/go-redis/redis/v8"
	muxHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mvrilo/go-redoc"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pixpivot/arc/internal/models"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

func seedPermissions(db *gorm.DB) {
	permissions := []models.Permission{
		{Name: "users:create", Description: "Can create new fiduciary users"},
		{Name: "users:read", Description: "Can view fiduciary users"},
		{Name: "users:update", Description: "Can update fiduciary users"},
		{Name: "users:delete", Description: "Can delete fiduciary users"},
		{Name: "users:impersonate", Description: "Can impersonate another user within the tenant"},
		{Name: "roles:manage", Description: "Can create, update, and delete roles"},
		{Name: "organizations:manage", Description: "Can manage organization entities"},
		{Name: "consents:read", Description: "Can view consent records"},
		{Name: "consents:update", Description: "Can update consent records"},
		{Name: "purposes:manage", Description: "Can manage consent purposes"},
		{Name: "consent-forms:manage", Description: "Can manage consent forms"},
		{Name: "grievances:read", Description: "Can view grievances"},
		{Name: "grievances:respond", Description: "Can respond to grievances"},
		{Name: "audit-logs:read", Description: "Can view audit logs"},
		{Name: "api-keys:manage", Description: "Can manage API keys"},
		{Name: "breaches:manage", Description: "Can manage breach notifications"},
		{Name: "dpas:manage", Description: "Can manage Data Processing Agreements"},
	}

	for _, p := range permissions {
		db.FirstOrCreate(&p, models.Permission{Name: p.Name})
	}
	log.Logger.Info().Msg("Permissions seeded successfully.")
}

func seedDatabase(gormDB *gorm.DB) {
	// Check if a default tenant exists, and if not, create one.
	var tenantCount int64
	gormDB.Model(&models.Tenant{}).Count(&tenantCount)
	if tenantCount == 0 {
		testTenantID := uuid.New()
		defaultTenant := models.Tenant{
			TenantID: testTenantID,
			Name:     "Default Test Tenant",
		}
		if err := gormDB.Create(&defaultTenant).Error; err != nil {
			log.Logger.Fatal().Err(err).Msg("Failed to seed database with default tenant")
		}
		log.Logger.Info().Str("tenant_id", testTenantID.String()).Msg("Created default test tenant")
	} else {
		log.Logger.Info().Msg("Database already seeded.")
	}
}

func main() {
	// Load config and init systems
	cfg := config.LoadConfig()
	log.InitLogger()
	jwtlink.Init(cfg.JWTSecret)
	if err := encryption.InitEncryption(); err != nil {
		log.Logger.Fatal().Err(err).Msg("encryption init failed")
	}
	log.Logger.Info().Msg("encryption ready")

	// API Docs
	doc := redoc.Redoc{
		Title:       "Consent Manager API",
		Description: "Manage user consents, grievances & notifications",
		SpecFile:    "./docs/swagger.json",
		SpecPath:    "/swagger.json",
		DocsPath:    "/docs",
	}

	// DB init
	db.InitDB(cfg)
	seedPermissions(db.MasterDB)
	seedDatabase(db.MasterDB)

	// JWT keys
	privateKey, err := auth.LoadPrivateKey("private.pem")
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to load private key")
	}
	publicKey, err := auth.LoadPublicKey("public.pem")
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to load public key")
	}

	// ==== LICENSING SYSTEM ====
	// 1. Redis for Usage Tracking
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	usageTracker := licensing.NewRedisUsageTracker(rdb)

	// 2. License Manager
	// Load public key for license verification (reusing JWT public key or separate one)
	// For now using the same public key path as JWT
	licenseVerifier, err := licensing.NewVerifierFromFile(cfg.PublicKeyPath)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to load license public key")
	}
	licenseManager := licensing.NewLicenseManager(licenseVerifier)

	// 3. Load License File
	licensePath := os.Getenv("LICENSE_FILE")
	if licensePath == "" {
		licensePath = "license.lic"
	}
	if _, err := os.Stat(licensePath); err == nil {
		licenseBytes, err := os.ReadFile(licensePath)
		if err == nil {
			if err := licenseManager.LoadLicense(string(licenseBytes)); err != nil {
				log.Logger.Error().Err(err).Msg("failed to load license")
			} else {
				log.Logger.Info().Msg("License loaded successfully")
			}
		}
	} else {
		log.Logger.Warn().Msg("No license file found. System running in restricted mode.")
	}

	// 4. Middleware
	licenseMiddleware := middleware.NewLicenseMiddleware(licenseManager, usageTracker)

	// Router & CORS
	r := mux.NewRouter()

	// Apply License Middleware globally (except health/docs which are usually public)
	// But for simplicity, applying to all. Ideally, we'd wrap specific routes.
	// Apply License Middleware globally (except health/docs)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip license check for health, docs, and authentication endpoints
			if r.URL.Path == "/health" ||
				r.URL.Path == "/swagger.json" ||
				len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/docs" ||
				len(r.URL.Path) >= 13 && r.URL.Path[:13] == "/api/v1/auth/" ||
				len(r.URL.Path) >= 10 && r.URL.Path[:10] == "/auth/sso/" {
				next.ServeHTTP(w, r)
				return
			}
			licenseMiddleware.EnforceLicense(next).ServeHTTP(w, r)
		})
	})

	cors := muxHandlers.CORS(
		muxHandlers.AllowedOrigins([]string{cfg.FrontendBaseURL}),
		muxHandlers.AllowedMethods([]string{
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions,
		}),
		muxHandlers.AllowedHeaders([]string{
			"Authorization", "Content-Type", "X-Requested-With", "Accept", "Origin",
		}),
		muxHandlers.AllowCredentials(),
	)

	// Health & docs
	r.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")
	r.Handle(doc.DocsPath, doc.Handler())
	r.Handle(doc.SpecPath, http.FileServer(http.Dir("./cmd/server/docs/")))

	// Core repos/services/hub
	consentRepo := repository.NewConsentRepository(db.MasterDB)
	auditRepo := repository.NewAuditRepo(db.MasterDB)
	auditService := services.NewAuditService(auditRepo)
	consentSvc := services.NewConsentService(consentRepo, auditService)
	notifRepo := repository.NewNotificationRepo(db.MasterDB)
	hub := realtime.NewHub()
	consentFormRepo := repository.NewConsentFormRepository(db.MasterDB)
	consentFormSvc := services.NewConsentFormService(consentFormRepo)
	userConsentRepo := repository.NewUserConsentRepository(db.MasterDB)
	webhookSvc := services.NewWebhookService(db.MasterDB)

	// SDK Service
	sdkRepo := repository.NewSDKRepository(db.MasterDB)
	sdkService := services.NewSDKGeneratorService(sdkRepo, consentFormRepo, cfg.BaseURL)

	// Cookie Services
	cookieRepo := repository.NewCookieRepository(db.MasterDB)
	cookieService := services.NewCookieService(cookieRepo)
	cookieScannerService := services.NewCookieScannerService(cookieRepo)

	// Email Service (needed by enhanced breach notification)
	emailService := services.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)

	// Breach Notification Service
	breachNotificationRepo := repository.NewBreachNotificationRepository(db.MasterDB)
	breachImpactAssessmentRepo := repository.NewBreachImpactAssessmentRepository(db.MasterDB)
	breachStakeholderRepo := repository.NewBreachStakeholderRepository(db.MasterDB)
	breachWorkflowRepo := repository.NewBreachWorkflowStageRepository(db.MasterDB)
	breachCommunicationRepo := repository.NewBreachCommunicationRepository(db.MasterDB)
	breachEvidenceRepo := repository.NewBreachEvidenceRepository(db.MasterDB)
	breachTimelineRepo := repository.NewBreachTimelineRepository(db.MasterDB)
	breachTemplateRepo := repository.NewBreachNotificationTemplateRepository(db.MasterDB)

	// Legacy breach service (for backward compatibility)
	breachNotificationSvc := services.NewBreachNotificationService(breachNotificationRepo)

	// Enhanced DPDP-compliant breach service
	enhancedBreachNotificationSvc := services.NewEnhancedBreachNotificationService(
		breachNotificationRepo,
		breachImpactAssessmentRepo,
		breachStakeholderRepo,
		breachWorkflowRepo,
		breachCommunicationRepo,
		breachEvidenceRepo,
		breachTimelineRepo,
		breachTemplateRepo,
		emailService,
	)

	// DSR Service
	dsrRepo := repository.NewDSRRepository(db.MasterDB, nil) // TenantDB is fetched dynamically
	dsrService := services.NewDSRService(dsrRepo)

	notificationPreferencesRepo := repository.NewNotificationPreferencesRepo(db.MasterDB)

	// Fiduciary Service
	fiduciaryRepo := repository.NewFiduciaryRepository(db.MasterDB)
	fiduciaryService := services.NewFiduciaryService(fiduciaryRepo)

	// Superadmin Service
	issuedLicenseRepo := repository.NewIssuedLicenseRepository(db.MasterDB)
	tenantRepo := repository.NewTenantRepository(db.MasterDB)
	superAdminService := services.NewSuperAdminService(issuedLicenseRepo, tenantRepo, fiduciaryRepo, privateKey)
	superAdminHandler := handlers.NewSuperAdminHandler(superAdminService)

	// New Services
	notificationPreferencesService := services.NewNotificationPreferencesService(notificationPreferencesRepo)
	notificationService := services.NewNotificationService(notifRepo, notificationPreferencesRepo, emailService, hub, fiduciaryService)

	// Backup Service
	backupService := services.NewBackupService(db.MasterDB, cfg)
	backupService.Start()

	// Receipt Service (supports local or S3/MinIO based on config)
	receiptRepo := repository.NewReceiptRepository(db.MasterDB)
	purposeRepo := repository.NewPurposeRepository(db.MasterDB)
	receiptService := services.NewReceiptService(
		receiptRepo,
		userConsentRepo,
		purposeRepo,
		emailService,
		cfg.BaseURL,
		cfg.StorageType,
		cfg.StoragePath,
		cfg.S3Bucket,
		cfg.S3Endpoint,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		cfg.S3Region,
		cfg.S3UseSSL,
		cfg.S3ForcePathStyle,
	)

	// User Consent Service (needs receipt service)
	userConsentSvc := services.NewUserConsentService(userConsentRepo, consentFormRepo, receiptService)

	// Auth middleware
	dataPrincipalAuth := middleware.RequireDataPrincipalAuth(publicKey)
	fiduciaryAuth := middleware.RequireFiduciaryAuth(publicKey)
	apiKeyAuth := middleware.APIKeyAuthMiddleware(db.MasterDB)
	requirePerm := middleware.RequirePermission

	// Wrapper for sending password reset email
	sendResetEmail := func(to, token string) error {
		resetLink := fmt.Sprintf("%s/reset-password?token=%s", cfg.BaseURL, token)
		body := fmt.Sprintf("Please click on the following link to reset your password: <a href=\"%s\">Reset Password</a>", resetLink)
		return emailService.Send(to, "Password Reset Request", body)
	}

	// ==== AUTH: USER ====
	r.Handle("/api/v1/auth/user/login", handlers.UserLoginHandler(db.MasterDB, cfg, privateKey)).Methods("POST")
	r.Handle("/api/v1/auth/user/forgot-password", dataPrincipalAuth(handlers.UserForgotPasswordHandler(db.MasterDB, sendResetEmail))).Methods("POST")
	r.Handle("/api/v1/auth/user/reset-password", dataPrincipalAuth(handlers.UserResetPasswordHandler(db.MasterDB))).Methods("POST")
	r.Handle("/api/v1/auth/user/refresh", dataPrincipalAuth(handlers.UserRefreshHandler(db.MasterDB, cfg, privateKey, publicKey))).Methods("POST")
	r.Handle("/api/v1/auth/user/logout", dataPrincipalAuth(handlers.UserLogoutHandler())).Methods("POST")
	r.Handle("/api/v1/user/profile", dataPrincipalAuth(handlers.UpdateUserHandler(db.MasterDB, auditService))).Methods("PUT")
	r.Handle("/api/v1/auth/user/me", dataPrincipalAuth(handlers.UserMeHandler(db.MasterDB, auditService))).Methods("GET")

	// ==== FIDUCIARY AUTH ====
	authRouter := r.PathPrefix("/api/v1/auth/fiduciary").Subrouter()
	authRouter.HandleFunc("/login", handlers.FiduciaryLoginHandler(db.MasterDB, cfg, privateKey)).Methods("POST")
	authRouter.HandleFunc("/refresh", handlers.FiduciaryRefreshHandler(db.MasterDB, cfg, privateKey, publicKey)).Methods("POST")
	authRouter.Handle("/me", fiduciaryAuth(handlers.FiduciaryMeHandler(db.MasterDB))).Methods("GET")
	authRouter.HandleFunc("/forgot-password", handlers.FiduciaryForgotPasswordHandler(db.MasterDB, sendResetEmail)).Methods("POST")
	authRouter.HandleFunc("/reset-password", handlers.FiduciaryResetPasswordHandler(db.MasterDB)).Methods("POST")
	authRouter.HandleFunc("/logout", handlers.FiduciaryLogoutHandler()).Methods("POST")

	// ==== SSO ====
	if db.MasterDB == nil {
		log.Logger.Fatal().Msg("MasterDB is nil in main before creating SSO handler")
	} else {
		log.Logger.Info().Msg("MasterDB is NOT nil in main before creating SSO handler")
	}
	ssoHandler := handlers.NewSSOHandler(db.MasterDB, cfg, privateKey)
	r.HandleFunc("/auth/sso/google", ssoHandler.GoogleLogin).Methods("GET")
	r.HandleFunc("/auth/sso/google/callback", ssoHandler.GoogleCallback).Methods("GET")
	r.HandleFunc("/auth/sso/microsoft", ssoHandler.MicrosoftLogin).Methods("GET")
	r.HandleFunc("/auth/sso/microsoft/callback", ssoHandler.MicrosoftCallback).Methods("GET")
	log.Logger.Info().Msg("Registered SSO routes: /auth/sso/google, /auth/sso/microsoft")

	// Debug Middleware to log all requests
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("Incoming request (matched)")
			next.ServeHTTP(w, r)
		})
	})

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("Incoming request (NOT MATCHED)")
		http.NotFound(w, r)
	})

	// ==== ORGANIZATION MANAGEMENT ====
	orgRepo := repository.NewOrganizationRepository(db.MasterDB)
	orgService := services.NewOrganizationService(orgRepo)
	orgHandler := handlers.NewOrganizationHandler(orgService, orgRepo)

	// Public or all-authenticated users can list/get organization details
	r.Handle("/public/api/v1/organizations", http.HandlerFunc(orgHandler.ListOrganizations)).Methods("GET")
	r.Handle("/public/api/v1/organizations/{id}", http.HandlerFunc(orgHandler.GetOrganizationByID)).Methods("GET")
	r.Handle("/public/api/v1/organizations/name/{name}", http.HandlerFunc(orgHandler.GetOrganizationByName)).Methods("GET")
	r.Handle("/public/api/v1/organizations/industry/{industry}", http.HandlerFunc(orgHandler.GetOrganizationsByIndustry)).Methods("GET")

	// Fiduciary & Superadmin can create/update/delete
	r.Handle("/api/v1/fiduciary/organizations", fiduciaryAuth(middleware.RequirePermission("organizations:manage")(http.HandlerFunc(orgHandler.CreateOrganization)))).Methods("POST")
	r.Handle("/api/v1/fiduciary/organizations/{id}", fiduciaryAuth(middleware.RequirePermission("organizations:manage")(http.HandlerFunc(orgHandler.UpdateOrganization)))).Methods("PUT")
	r.Handle("/api/v1/fiduciary/organizations/{id}", fiduciaryAuth(middleware.RequirePermission("organizations:manage")(http.HandlerFunc(orgHandler.DeleteOrganization)))).Methods("DELETE")

	// ==== DATA PRINCIPAL MANAGEMENT BY FIDUCIARY ====
	adminUserRouter := r.PathPrefix("/api/v1/fiduciary/users").Subrouter()
	adminUserRouter.Use(fiduciaryAuth)
	adminUserRouter.HandleFunc("", handlers.FiduciaryCreateUserHandler(db.MasterDB, auditService)).Methods("POST")
	adminUserRouter.HandleFunc("/identify", handlers.IdentifyUserHandler(db.MasterDB)).Methods("POST")
	adminUserRouter.Handle("/{userId}/impersonate", requirePerm("users:impersonate")(handlers.ImpersonateUserHandler(db.MasterDB, auditService, privateKey, cfg.AdminTokenTTL))).Methods("POST")

	// ==== SIGNUP ====
	signupHandler := handlers.NewSignupHandler(db.MasterDB, cfg, orgService, emailService, auditService)
	r.HandleFunc("/api/v1/auth/user/signup", signupHandler.SignupDataPrincipal).Methods("POST")
	r.HandleFunc("/api/v1/auth/fiduciary/signup", signupHandler.SignupFiduciary).Methods("POST")
	r.HandleFunc("/api/v1/auth/verify-guardian", signupHandler.VerifyGuardian).Methods("GET")

	// ==== CONSENT ====
	consentHandler := handlers.NewConsentHandler(consentSvc, auditService) // TODO: This needs refactoring
	consentHandler.RegisterRoutes(r, db.MasterDB)                          // TODO: This needs refactoring

	// ==== PARTNER API ====
	partnerR := r.PathPrefix("/api/v1/partner").Subrouter()
	partnerR.HandleFunc("/consents", func(w http.ResponseWriter, r *http.Request) {
		apiKeyAuth(http.HandlerFunc(handlers.PartnerGetConsentsHandler)).ServeHTTP(w, r) // TODO: This needs refactoring
	}).Methods("GET")

	// ==== USER DSR ====
	DSRhandlers := handlers.NewDataRequestHandler(db.MasterDB, dsrService, auditService)
	r.Handle("/api/v1/user/requests", dataPrincipalAuth(http.HandlerFunc(DSRhandlers.ListUserRequests))).Methods("GET")
	r.Handle("/api/v1/user/requests", dataPrincipalAuth(http.HandlerFunc(DSRhandlers.CreateUserRequest))).Methods("POST")
	r.Handle("/api/v1/user/requests/{id}", dataPrincipalAuth(http.HandlerFunc(DSRhandlers.GetRequestDetails))).Methods("GET")

	// ==== FIDUCIARY DSR ====
	r.Handle("/api/v1/fiduciary/requests", fiduciaryAuth(middleware.RequirePermission("consents:read")(http.HandlerFunc(DSRhandlers.ListAdminRequests)))).Methods("GET")
	r.Handle("/api/v1/fiduciary/requests/{id}", fiduciaryAuth(middleware.RequirePermission("consents:read")(http.HandlerFunc(DSRhandlers.GetAdminRequestDetails)))).Methods("GET")
	r.Handle("/api/v1/fiduciary/requests/{id}/approve", fiduciaryAuth(middleware.RequirePermission("consents:update")(http.HandlerFunc(DSRhandlers.ApproveRequest)))).Methods("POST")
	r.Handle("/api/v1/fiduciary/requests/{id}/reject", fiduciaryAuth(middleware.RequirePermission("consents:update")(http.HandlerFunc(DSRhandlers.RejectRequest)))).Methods("POST")

	// ==== PURPOSES ====
	purposeHandler := handlers.NewPurposeHandler(db.MasterDB)
	purposeRouter := r.PathPrefix("/api/v1/fiduciary/purposes").Subrouter()
	purposeRouter.Use(fiduciaryAuth)
	purposeRouter.HandleFunc("", handlers.CreatePurposeHandler()).Methods("POST")
	purposeRouter.HandleFunc("", handlers.ListPurposesHandler()).Methods("GET")
	purposeRouter.HandleFunc("/{id}/toggle", purposeHandler.ToggleActive).Methods("POST")
	purposeRouter.HandleFunc("/{id}", handlers.UpdatePurposeHandler()).Methods("PUT")
	purposeRouter.HandleFunc("/{id}", handlers.DeletePurposeHandler()).Methods("DELETE")

	userPurposeRouter := r.PathPrefix("/api/v1/user/purposes").Subrouter()
	userPurposeRouter.Use(dataPrincipalAuth)
	userPurposeRouter.HandleFunc("/{id}", handlers.UserGetPurposeHandler()).Methods("GET")
	userPurposeRouter.HandleFunc("/tenant/{tenantID}", handlers.UserGetPurposeByTenant()).Methods("GET")

	// ==== API KEYS ====
	r.Handle("/api/v1/fiduciary/api-keys", fiduciaryAuth(middleware.RequirePermission("api-keys:manage")(http.HandlerFunc(handlers.CreateAPIKeyHandler(db.MasterDB, publicKey))))).Methods("POST")
	r.Handle("/api/v1/fiduciary/api-keys", fiduciaryAuth(middleware.RequirePermission("api-keys:manage")(http.HandlerFunc(handlers.ListAPIKeysHandler(db.MasterDB, publicKey))))).Methods("GET")
	r.Handle("/api/v1/fiduciary/api-keys/revoke", fiduciaryAuth(middleware.RequirePermission("api-keys:manage")(http.HandlerFunc(handlers.RevokeAPIKeyHandler(db.MasterDB, publicKey))))).Methods("PUT")

	// ==== TENANT SETTINGS ====
	r.Handle("/api/v1/fiduciary/tenant/settings", fiduciaryAuth(middleware.RequirePermission("roles:manage")(http.HandlerFunc(handlers.UpdateTenantSettingsHandler(db.MasterDB))))).Methods("PUT")

	// ==== AUDIT LOGS ====
	fiduciaryGR := r.PathPrefix("/api/v1/fiduciary").Subrouter()
	fiduciaryGR.Use(fiduciaryAuth)
	fiduciaryGR.Use(middleware.RequirePermission("audit-logs:read"))
	fiduciaryGR.Use(handlers.TenantContextMiddleware)
	fiduciaryGR.HandleFunc("/audit/logs", handlers.GetTenantAuditLogsHandler()).Methods("GET")

	// ==== GRIEVANCES ====
	grievHandler := handlers.NewGrievanceHandler(notificationService, hub, auditService)
	r.Handle("/api/v1/dashboard/grievances", dataPrincipalAuth(http.HandlerFunc(grievHandler.Create))).Methods("POST")
	r.Handle("/api/v1/dashboard/grievances", dataPrincipalAuth(http.HandlerFunc(grievHandler.ListForUser))).Methods("GET")

	fiduciaryGR.Use(fiduciaryAuth)
	fiduciaryGR.Handle("/grievances", middleware.RequirePermission("grievances:read")(http.HandlerFunc(grievHandler.List))).Methods("GET")
	fiduciaryGR.Handle("/grievances/{id}", middleware.RequirePermission("grievances:respond")(http.HandlerFunc(grievHandler.Update))).Methods("PUT")

	// ===== Grievance Comments =====
	r.Handle("/api/v1/dashboard/grievances/{id}/comments", dataPrincipalAuth(http.HandlerFunc(grievHandler.AddComment))).Methods("POST")
	r.Handle("/api/v1/dashboard/grievances/{id}/comments", dataPrincipalAuth(http.HandlerFunc(grievHandler.GetComments))).Methods("GET")
	r.Handle("/api/v1/dashboard/grievances/comments/{commentID}", dataPrincipalAuth(http.HandlerFunc(grievHandler.DeleteComment))).Methods("DELETE")

	fiduciaryGR.Handle("/grievances/{id}/comments", middleware.RequirePermission("grievances:respond")(http.HandlerFunc(grievHandler.AddComment))).Methods("POST")
	fiduciaryGR.Handle("/grievances/{id}/comments", middleware.RequirePermission("grievances:read")(http.HandlerFunc(grievHandler.GetComments))).Methods("GET")
	fiduciaryGR.Handle("/grievances/comments/{commentID}", middleware.RequirePermission("grievances:respond")(http.HandlerFunc(grievHandler.DeleteComment))).Methods("DELETE")

	// ==== VENDOR ====
	vendorRepo := repository.NewVendorRepository(db.MasterDB)
	// TPRM repo/service
	tprmRepo := repository.NewTPRMRepository(db.MasterDB)
	tprmService := services.NewTPRMService(
		tprmRepo,
		cfg.BaseURL,
		cfg.StorageType,
		cfg.StoragePath,
		cfg.S3Bucket,
		cfg.S3Endpoint,
		cfg.S3AccessKey,
		cfg.S3SecretKey,
		cfg.S3Region,
		cfg.S3UseSSL,
		cfg.S3ForcePathStyle,
	)
	vendorService := services.NewVendorService(vendorRepo)
	vendorHandler := handlers.NewVendorHandler(vendorService)

	// Public or all-authenticated users can list/get vendor details
	r.Handle("/api/v1/vendors", dataPrincipalAuth(http.HandlerFunc(vendorHandler.ListVendors))).Methods("GET")
	r.Handle("/api/v1/vendors/{id}", dataPrincipalAuth(http.HandlerFunc(vendorHandler.GetVendorByID))).Methods("GET")

	// Fiduciary & Superadmin can create/update/delete
	r.Handle("/api/v1/fiduciary/vendors", fiduciaryAuth(middleware.RequirePermission("dpas:manage")(http.HandlerFunc(vendorHandler.CreateVendor)))).Methods("POST")
	r.Handle("/api/v1/fiduciary/vendors/{id}", fiduciaryAuth(middleware.RequirePermission("dpas:manage")(http.HandlerFunc(vendorHandler.UpdateVendor)))).Methods("PUT")
	r.Handle("/api/v1/fiduciary/vendors/{id}", fiduciaryAuth(middleware.RequirePermission("dpas:manage")(http.HandlerFunc(vendorHandler.DeleteVendor)))).Methods("DELETE")

	// ==== CONSENT FORMS ==== (managed by fiduciary)
	consentFormHandler := handlers.NewConsentFormHandler(consentFormSvc, auditService)
	consentFormRouter := r.PathPrefix("/api/v1/fiduciary/consent-forms").Subrouter()
	consentFormRouter.Use(fiduciaryAuth)
	consentFormRouter.HandleFunc("", http.HandlerFunc(consentFormHandler.CreateConsentForm)).Methods("POST")
	consentFormRouter.HandleFunc("", http.HandlerFunc(consentFormHandler.ListConsentForms)).Methods("GET")
	consentFormRouter.HandleFunc("/{formId}", http.HandlerFunc(consentFormHandler.GetConsentForm)).Methods("GET")
	consentFormRouter.HandleFunc("/{formId}", http.HandlerFunc(consentFormHandler.UpdateConsentForm)).Methods("PUT")
	consentFormRouter.HandleFunc("/{formId}", http.HandlerFunc(consentFormHandler.DeleteConsentForm)).Methods("DELETE")
	consentFormRouter.HandleFunc("/{formId}/purposes", http.HandlerFunc(consentFormHandler.AddPurposeToConsentForm)).Methods("POST")
	consentFormRouter.HandleFunc("/{formId}/purposes/{purposeId}", http.HandlerFunc(consentFormHandler.UpdatePurposeInConsentForm)).Methods("PUT")
	consentFormRouter.HandleFunc("/{formId}/purposes/{purposeId}", http.HandlerFunc(consentFormHandler.RemovePurposeFromConsentForm)).Methods("DELETE")
	consentFormRouter.HandleFunc("/{formId}/script", http.HandlerFunc(consentFormHandler.GetIntegrationScript)).Methods("GET")
	consentFormRouter.HandleFunc("/{formId}/integration", http.HandlerFunc(consentFormHandler.GetIntegrationScript)).Methods("GET")
	// Validation and versioning endpoints
	consentFormRouter.HandleFunc("/{formId}/validate", http.HandlerFunc(consentFormHandler.ValidateConsentForm)).Methods("POST")
	consentFormRouter.HandleFunc("/{formId}/publish", http.HandlerFunc(consentFormHandler.PublishConsentForm)).Methods("POST")
	consentFormRouter.HandleFunc("/{formId}/submit-for-review", http.HandlerFunc(consentFormHandler.SubmitForReview)).Methods("POST")
	consentFormRouter.HandleFunc("/{formId}/versions", http.HandlerFunc(consentFormHandler.GetVersionHistory)).Methods("GET")
	consentFormRouter.HandleFunc("/{formId}/versions/{versionId}", http.HandlerFunc(consentFormHandler.GetVersion)).Methods("GET")
	consentFormRouter.HandleFunc("/{formId}/rollback/{versionId}", http.HandlerFunc(consentFormHandler.RollbackToVersion)).Methods("POST")

	// ==== SDK MANAGEMENT ====
	sdkHandler := handlers.NewSDKHandler(sdkService, auditService)
	// Public SDK endpoint
	r.HandleFunc("/api/v1/public/sdk/{tenantId}/{formId}", sdkHandler.GetSDK).Methods("GET")
	// Fiduciary SDK management endpoints
	sdkRouter := r.PathPrefix("/api/v1/fiduciary").Subrouter()
	sdkRouter.Use(fiduciaryAuth)
	sdkRouter.HandleFunc("/sdk-config/{formId}", http.HandlerFunc(sdkHandler.GetSDKConfig)).Methods("GET")
	sdkRouter.HandleFunc("/sdk-config", http.HandlerFunc(sdkHandler.CreateSDKConfig)).Methods("POST")
	sdkRouter.HandleFunc("/sdk-config/{configId}", http.HandlerFunc(sdkHandler.UpdateSDKConfig)).Methods("PUT")
	sdkRouter.HandleFunc("/sdk-config/{configId}", http.HandlerFunc(sdkHandler.DeleteSDKConfig)).Methods("DELETE")
	sdkRouter.HandleFunc("/integration-code/{formId}", http.HandlerFunc(sdkHandler.GetIntegrationCode)).Methods("GET")

	// ==== COOKIE MANAGEMENT ====
	cookieHandler := handlers.NewCookieHandler(cookieService, cookieScannerService, auditService)
	// Public cookie endpoint for SDK
	r.HandleFunc("/api/v1/public/cookies/{tenantId}", cookieHandler.GetAllowedCookies).Methods("GET")
	// Fiduciary cookie management endpoints
	cookieRouter := r.PathPrefix("/api/v1/fiduciary/cookies").Subrouter()
	cookieRouter.Use(fiduciaryAuth)
	cookieRouter.HandleFunc("", http.HandlerFunc(cookieHandler.CreateCookie)).Methods("POST")
	cookieRouter.HandleFunc("", http.HandlerFunc(cookieHandler.ListCookies)).Methods("GET")
	cookieRouter.HandleFunc("/{cookieId}", http.HandlerFunc(cookieHandler.GetCookie)).Methods("GET")
	cookieRouter.HandleFunc("/{cookieId}", http.HandlerFunc(cookieHandler.UpdateCookie)).Methods("PUT")
	cookieRouter.HandleFunc("/{cookieId}", http.HandlerFunc(cookieHandler.DeleteCookie)).Methods("DELETE")
	cookieRouter.HandleFunc("/bulk-categorize", http.HandlerFunc(cookieHandler.BulkCategorizeCookies)).Methods("PUT")
	cookieRouter.HandleFunc("/stats", http.HandlerFunc(cookieHandler.GetCookieStats)).Methods("GET")
	cookieRouter.HandleFunc("/scan", http.HandlerFunc(cookieHandler.ScanWebsite)).Methods("POST")
	cookieRouter.HandleFunc("/scans", http.HandlerFunc(cookieHandler.GetScanHistory)).Methods("GET")
	cookieRouter.HandleFunc("/scans/{scanId}", http.HandlerFunc(cookieHandler.GetScanResults)).Methods("GET")

	// ==== PUBLIC CONSENT FLOW ====
	publicConsentHandler := handlers.NewPublicConsentHandler(userConsentSvc, consentFormSvc, webhookSvc)
	publicConsentRouter := r.PathPrefix("/api/v1/public/consent-forms").Subrouter()
	publicConsentRouter.Use(apiKeyAuth)
	publicConsentRouter.HandleFunc("/{formId}", http.HandlerFunc(publicConsentHandler.GetConsentForm)).Methods("GET")

	userConsentRouter := r.PathPrefix("/api/v1/user/consents").Subrouter()
	userConsentRouter.Use(dataPrincipalAuth)
	userConsentRouter.Handle("/submit/{formId}", http.HandlerFunc(publicConsentHandler.SubmitConsent)).Methods("POST")
	userConsentRouter.Handle("", http.HandlerFunc(publicConsentHandler.GetUserConsents)).Methods("GET")
	userConsentRouter.Handle("/withdraw/{purposeId}", http.HandlerFunc(publicConsentHandler.WithdrawConsent)).Methods("POST")
	userConsentRouter.Handle("/{purposeId}", http.HandlerFunc(publicConsentHandler.GetUserConsentForPurpose)).Methods("GET")

	// ==== RECEIPT MANAGEMENT ====
	receiptHandler := handlers.NewReceiptHandler(receiptService)

	// User receipt endpoints
	userConsentRouter.Handle("/{consentId}/receipt", http.HandlerFunc(receiptHandler.GenerateReceipt)).Methods("POST")

	receiptRouter := r.PathPrefix("/api/v1/user/receipts").Subrouter()
	receiptRouter.Use(dataPrincipalAuth)
	receiptRouter.Handle("/{receiptId}", http.HandlerFunc(receiptHandler.GetReceipt)).Methods("GET")
	receiptRouter.Handle("/{receiptId}/download", http.HandlerFunc(receiptHandler.DownloadReceipt)).Methods("GET")
	receiptRouter.Handle("/{receiptId}/email", http.HandlerFunc(receiptHandler.EmailReceipt)).Methods("POST")

	// Public receipt verification endpoint
	r.Handle("/api/v1/public/receipts/verify/{receiptNumber}", http.HandlerFunc(receiptHandler.VerifyReceipt)).Methods("GET")

	// Fiduciary bulk receipt endpoints
	fiduciaryReceiptRouter := r.PathPrefix("/api/v1/fiduciary/receipts").Subrouter()
	fiduciaryReceiptRouter.Use(fiduciaryAuth)
	fiduciaryReceiptRouter.Handle("/bulk-generate", http.HandlerFunc(receiptHandler.BulkGenerateReceipts)).Methods("POST")
	fiduciaryReceiptRouter.Handle("/bulk-download", http.HandlerFunc(receiptHandler.BulkDownloadReceipts)).Methods("POST")

	// ==== NOTIFICATION PREFERENCES ====
	notificationPreferencesHandler := handlers.NewNotificationPreferencesHandler(notificationPreferencesService)
	notificationPreferencesRouter := r.PathPrefix("/api/v1/user/notification-preferences").Subrouter()
	notificationPreferencesRouter.Use(dataPrincipalAuth)
	notificationPreferencesRouter.Handle("", http.HandlerFunc(notificationPreferencesHandler.Get)).Methods("GET")
	notificationPreferencesRouter.Handle("", http.HandlerFunc(notificationPreferencesHandler.Update)).Methods("PUT")

	// ==== FIDUCIARY MANAGEMENT ====
	fiduciaryManagementRouter := r.PathPrefix("/api/v1/fiduciaries").Subrouter()
	fiduciaryManagementRouter.Use(fiduciaryAuth, middleware.RequirePermission("users:read", "users:create", "users:update", "users:delete"))
	fiduciaryManagementRouter.HandleFunc("", handlers.ListAllFiduciariesHandler(fiduciaryService)).Methods("GET")
	fiduciaryManagementRouter.HandleFunc("", handlers.CreateNewFiduciaryHandler(fiduciaryService, licenseManager)).Methods("POST")
	fiduciaryManagementRouter.HandleFunc("/stats", handlers.FiduciaryStatsHandler(fiduciaryService)).Methods("GET")
	fiduciaryManagementRouter.HandleFunc("/{fiduciaryId}", handlers.GetFiduciaryByIDHandler(fiduciaryService)).Methods("GET")
	fiduciaryManagementRouter.HandleFunc("/{fiduciaryId}", handlers.UpdateFiduciaryDataHandler(fiduciaryService)).Methods("PUT")
	fiduciaryManagementRouter.HandleFunc("/{fiduciaryId}", handlers.DeleteFiduciaryByIDHandler(fiduciaryService)).Methods("DELETE")

	// === CONSENT MANAGEMENT FOR FIDUCIARY ===
	consentManagementRouter := r.PathPrefix("/api/v1/fiduciary/consents").Subrouter()
	consentManagementRouter.Use(fiduciaryAuth)
	consentManagementRouter.HandleFunc("", handlers.ListConsentsHandler(consentSvc)).Methods("GET")
	consentManagementRouter.HandleFunc("/stats", consentHandler.GetConsentStats).Methods("GET")
	consentManagementRouter.HandleFunc("/{consentId}", handlers.GetConsentByIDHandler(consentSvc)).Methods("GET")

	// ==== BREACH NOTIFICATIONS (Legacy) ====
	breachNotificationHandler := handlers.NewBreachNotificationHandler(breachNotificationSvc, auditService)
	breachNotificationRouter := r.PathPrefix("/api/v1/fiduciary/breach-notifications").Subrouter()
	breachNotificationRouter.Use(fiduciaryAuth, middleware.RequirePermission("breaches:manage"))
	breachNotificationRouter.Handle("", http.HandlerFunc(breachNotificationHandler.CreateBreachNotification)).Methods("POST")
	breachNotificationRouter.Handle("", http.HandlerFunc(breachNotificationHandler.ListBreachNotifications)).Methods("GET")
	breachNotificationRouter.Handle("/stats", http.HandlerFunc(breachNotificationHandler.GetBreachStats)).Methods("GET")
	breachNotificationRouter.Handle("/{notificationId}", http.HandlerFunc(breachNotificationHandler.GetBreachNotification)).Methods("GET")
	breachNotificationRouter.Handle("/{notificationId}", http.HandlerFunc(breachNotificationHandler.UpdateBreachNotification)).Methods("PUT")
	breachNotificationRouter.Handle("/{notificationId}", http.HandlerFunc(breachNotificationHandler.DeleteBreachNotification)).Methods("DELETE")

	// ==== ENHANCED BREACH NOTIFICATIONS (DPDP-Compliant) ====
	enhancedBreachHandler := handlers.NewEnhancedBreachNotificationHandler(enhancedBreachNotificationSvc, auditService)
	dpdpBreachRouter := r.PathPrefix("/api/v1/fiduciary/dpdp-breaches").Subrouter()
	dpdpBreachRouter.Use(fiduciaryAuth, middleware.RequirePermission("breaches:manage"))

	// Breach CRUD
	dpdpBreachRouter.HandleFunc("", enhancedBreachHandler.CreateBreachNotification).Methods("POST")
	dpdpBreachRouter.HandleFunc("/register", enhancedBreachHandler.GetBreachRegister).Methods("GET")

	// Workflow endpoints
	dpdpBreachRouter.HandleFunc("/{id}/submit-verification", enhancedBreachHandler.SubmitForVerification).Methods("POST")
	dpdpBreachRouter.HandleFunc("/{id}/verify", enhancedBreachHandler.VerifyBreach).Methods("POST")
	dpdpBreachRouter.HandleFunc("/{id}/approve-data-principal-notification", enhancedBreachHandler.ApproveDataPrincipalNotification).Methods("POST")

	// Notifications
	dpdpBreachRouter.HandleFunc("/{id}/notify-dpb", enhancedBreachHandler.SendDPBNotification).Methods("POST")
	dpdpBreachRouter.HandleFunc("/{id}/notify-data-principals", enhancedBreachHandler.SendDataPrincipalNotifications).Methods("POST")

	// Compliance
	dpdpBreachRouter.HandleFunc("/{id}/sla-status", enhancedBreachHandler.CheckSLACompliance).Methods("GET")

	// ==== THIRD-PARTY RISK MANAGEMENT (TPRM) ====
	tprmHandler := handlers.NewTPRMHandler(tprmService)
	tprmRouter := r.PathPrefix("/api/v1/fiduciary/tprm").Subrouter()
	tprmRouter.Use(fiduciaryAuth, middleware.RequirePermission("vendors:manage"))
	tprmHandler.RegisterRoutes(tprmRouter)

	// ==== DATA PROCESSING AGREEMENTS ====
	dpaHandler := handlers.NewDPAHandler(db.MasterDB, auditService)
	dpaRouter := r.PathPrefix("/api/v1/fiduciary/dpas").Subrouter()
	dpaRouter.Use(fiduciaryAuth, middleware.RequirePermission("dpas:manage"))
	dpaHandler.RegisterRoutes(dpaRouter)

	// ==== RBAC MANAGEMENT (Roles, Permissions) ====
	rbacHandler := handlers.NewRBACHandler(db.MasterDB)
	rbacRouter := r.PathPrefix("/api/v1/fiduciary").Subrouter()
	rbacRouter.Use(fiduciaryAuth, middleware.RequirePermission("roles:manage"))
	rbacRouter.HandleFunc("/permissions", rbacHandler.ListPermissions).Methods("GET")
	rbacRouter.HandleFunc("/roles", rbacHandler.ListRoles).Methods("GET")
	rbacRouter.HandleFunc("/roles", rbacHandler.CreateRole).Methods("POST")
	rbacRouter.HandleFunc("/roles/{roleId}", rbacHandler.UpdateRole).Methods("PUT")
	rbacRouter.HandleFunc("/roles/{roleId}", rbacHandler.DeleteRole).Methods("DELETE")
	rbacRouter.HandleFunc("/users/{userId}/roles", rbacHandler.AssignRolesToUser).Methods("PUT")

	// ==== PUBLIC API ====
	publicApiRouter := r.PathPrefix("/api/v1/public").Subrouter()
	publicApiRouter.Use(apiKeyAuth)
	publicAPIHandler := handlers.NewPublicAPIHandler(db.MasterDB, dsrService, userConsentSvc, auditService, webhookSvc)
	publicApiRouter.HandleFunc("/users", publicAPIHandler.CreateDataPrincipal).Methods("POST")
	publicApiRouter.HandleFunc("/users/{userId}/consents", publicAPIHandler.GetDataPrincipalConsents).Methods("GET")
	publicApiRouter.HandleFunc("/consents/verify", publicAPIHandler.VerifyConsents).Methods("POST")
	publicApiRouter.HandleFunc("/consents/submit", publicAPIHandler.SubmitConsentViaAPI).Methods("POST")
	publicApiRouter.HandleFunc("/dsr", publicAPIHandler.CreateDSR).Methods("POST")

	// ==== WEBHOOK MANAGEMENT ====
	webhookHandler := handlers.NewWebhookHandler(db.MasterDB)
	webhookRouter := r.PathPrefix("/api/v1/fiduciary/webhooks").Subrouter()
	webhookRouter.Use(fiduciaryAuth, middleware.RequirePermission("roles:manage")) // Reuse a high-level permission
	webhookRouter.HandleFunc("", webhookHandler.CreateWebhook).Methods("POST")
	webhookRouter.HandleFunc("", webhookHandler.ListWebhooks).Methods("GET")
	webhookRouter.HandleFunc("/{webhookId}", webhookHandler.DeleteWebhook).Methods("DELETE")

	// ==== DATA DISCOVERY ====
	discoveryHandler := handlers.NewDataDiscoveryHandler(db.MasterDB)
	discoveryRouter := r.PathPrefix("/api/v1/fiduciary/discovery").Subrouter()
	discoveryRouter.Use(fiduciaryAuth)
	// TODO: Add specific permissions for discovery
	discoveryRouter.HandleFunc("/sources", discoveryHandler.CreateDataSource).Methods("POST")
	discoveryRouter.HandleFunc("/sources", discoveryHandler.ListDataSources).Methods("GET")
	discoveryRouter.HandleFunc("/sources/{id}/scan", discoveryHandler.StartScan).Methods("POST")
	discoveryRouter.HandleFunc("/jobs/{id}", discoveryHandler.GetJobResults).Methods("GET")
	discoveryRouter.HandleFunc("/dashboard", discoveryHandler.GetDashboardStats).Methods("GET")

	// ==== CHILD CONSENT MANAGEMENT ====
	childRepo := repository.NewChildConsentRepository(db.MasterDB)
	childService := services.NewChildConsentService(childRepo)
	childHandler := handlers.NewChildConsentHandler(childService)

	childRouter := r.PathPrefix("/api/v1/user/children").Subrouter()
	childRouter.Use(dataPrincipalAuth)
	childRouter.HandleFunc("", childHandler.AddChild).Methods("POST")
	childRouter.HandleFunc("", childHandler.ListChildren).Methods("GET")
	childRouter.HandleFunc("/{childId}/consent-request", childHandler.CreateConsentRequest).Methods("POST")

	parentRouter := r.PathPrefix("/api/v1/user/parental-consent").Subrouter()
	parentRouter.Use(dataPrincipalAuth)
	parentRouter.HandleFunc("/{requestId}/approve", childHandler.ApproveRequest).Methods("POST")
	parentRouter.HandleFunc("/{requestId}/reject", childHandler.RejectRequest).Methods("POST")

	// ==== REVIEW TOKEN & METRICS ====
	r.HandleFunc("/api/v1/review", handlers.ReviewTokenHandler(db.MasterDB, publicKey)).Methods("GET")
	r.Handle("/metrics", promhttp.Handler())

	// ==== SUPERADMIN DASHBOARD ====
	superAdminRouter := r.PathPrefix("/api/v1/superadmin").Subrouter()
	superAdminRouter.Use(middleware.RequireSuperAdmin)

	// Licenses
	superAdminRouter.HandleFunc("/licenses", superAdminHandler.GenerateLicense).Methods("POST")
	superAdminRouter.HandleFunc("/licenses", superAdminHandler.ListLicenses).Methods("GET")
	superAdminRouter.HandleFunc("/licenses/{id}/revoke", superAdminHandler.RevokeLicense).Methods("PUT")

	// Tenants
	superAdminRouter.HandleFunc("/tenants", superAdminHandler.ListTenants).Methods("GET")
	superAdminRouter.HandleFunc("/tenants", superAdminHandler.CreateTenant).Methods("POST")
	superAdminRouter.HandleFunc("/tenants/{id}", superAdminHandler.UpdateTenant).Methods("PUT")
	superAdminRouter.HandleFunc("/tenants/{id}", superAdminHandler.DeleteTenant).Methods("DELETE")

	// ==== START SERVER ====
	handler := cors(r)
	log.Logger.Info().Msgf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Logger.Fatal().Err(err).Msg("server failed")
	}
}
