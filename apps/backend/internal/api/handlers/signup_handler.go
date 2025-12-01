package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	logger "log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"pixpivot/arc/config"
	"pixpivot/arc/internal/auth"
	"pixpivot/arc/internal/core/services"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/pkg/log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Organization struct {
	Name        string `json:"name"`
	Industry    string `json:"industry"`
	CompanySize string `json:"companySize"`
	TaxID       string `json:"taxId,omitempty"`
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	Country     string `json:"country,omitempty"`
}

// FiduciarySignupRequest defines the shape of a signup request for a DF user.
// This user will manage a tenant and have specific permissions.
type FiduciarySignupRequest struct {
	Email        string       `json:"email"`
	FirstName    string       `json:"firstName"`
	LastName     string       `json:"lastName"`
	Phone        string       `json:"phone"`
	Password     string       `json:"password"`
	ConfirmPass  string       `json:"confirmPassword"`
	Role         string       `json:"role"`
	Organization Organization `json:"organization"`
}

// DataPrincipalSignupRequest defines the shape of a signup request for a DP user.
// This user is the data subject and is created by a Fiduciary.
type DataPrincipalSignupRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Age           int    `json:"age"`
	GuardianEmail string `json:"guardianEmail,omitempty"`
	Location      string `json:"location,omitempty"`
	Phone         string `json:"phone"`
}

type SignupHandler struct {
	MasterDB            *gorm.DB
	Cfg                 config.Config
	OrganizationService *services.OrganizationService
	EmailService        *services.EmailService
	AuditService        *services.AuditService
}

func NewSignupHandler(
	masterDB *gorm.DB,
	cfg config.Config,
	organizationService *services.OrganizationService, // Add this parameter
	emailService *services.EmailService,
	auditService *services.AuditService,
) *SignupHandler {
	return &SignupHandler{
		MasterDB:            masterDB,
		Cfg:                 cfg,
		OrganizationService: organizationService, // Assign the new parameter
		EmailService:        emailService,
		AuditService:        auditService,
	}
}

// SignupFiduciary handles the creation of a new tenant and its first admin user (a FiduciaryUser).
func (h *SignupHandler) SignupFiduciary(w http.ResponseWriter, r *http.Request) {
	var req FiduciarySignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Logger.Error().Err(err).Msg("Invalid request body for fiduciary signup")
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !validateFiduciaryRequest(&req, w) {
		log.Logger.Error().Msg("Fiduciary signup validation failed")
		return
	}

	// 1. Check for duplicates (Pre-check)
	var existingUser models.FiduciaryUser
	if err := h.MasterDB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		log.Logger.Error().Str("email", req.Email).Msg("Fiduciary user already exists")
		writeError(w, http.StatusBadRequest, "A user with this email already exists")
		return
	}
	var existingOrg models.OrganizationEntity
	if err := h.MasterDB.Where("email = ?", req.Organization.Email).First(&existingOrg).Error; err == nil {
		log.Logger.Error().Str("email", req.Organization.Email).Msg("Organization with this email already exists")
		writeError(w, http.StatusBadRequest, "An organization with this email already exists")
		return
	}

	tenantID := uuid.New()
	dbName := "tenant_" + strings.ReplaceAll(tenantID.String(), "-", "")

	// 2. Create the physical database (Must be done outside transaction)
	// We use the MasterDB connection but need to ensure we are not in a transaction block here.
	// Also, CREATE DATABASE cannot run inside a transaction block.
	if err := h.MasterDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error; err != nil {
		log.Logger.Error().Err(err).Str("dbname", dbName).Msg("Failed to create tenant database")
		writeError(w, http.StatusInternalServerError, "Failed to provision tenant resources")
		return
	}

	// 3. Initialize the new DB (Run migrations)
	// We can use db.GetTenantDB which now connects to the new DB and runs migrations
	// But we need to make sure the record exists in MasterDB first for GetTenantDB to work?
	// registry.go's GetTenantDB checks MasterDB. So we should create the Tenant record first?
	// But if we create Tenant record first in a transaction, we can't commit it before creating DB if we want atomic.
	// Actually, `GetTenantDB` check is just a safeguard. `loadTenantDB` does the work.
	// Let's manually trigger `loadTenantDB`-like logic or just create the Tenant record now.

	// Strategy:
	// A. Create Database (Done)
	// B. Start Transaction on MasterDB
	// C. Create Tenant, User, Org records
	// D. Commit
	// E. If Commit fails, Drop Database (Cleanup)

	// We need to run migrations on the new DB. We can do this AFTER committing the tenant record,
	// or we can do it now using a direct connection.
	// Let's do it now to ensure DB is valid before creating user records.

	// We can't use db.GetTenantDB yet because the tenant record doesn't exist.
	// Let's use a temporary connection or modify db.GetTenantDB to allow skipping check?
	// Or just replicate the connection logic here for the initial setup.

	// Replicating connection logic for setup:
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		h.Cfg.DBHost, h.Cfg.DBUser, h.Cfg.DBPassword, dbName, h.Cfg.DBPort)

	newTenantDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to connect to new tenant DB")
		// Cleanup
		h.MasterDB.Exec(fmt.Sprintf("DROP DATABASE %s", dbName))
		writeError(w, http.StatusInternalServerError, "Failed to initialize tenant database")
		return
	}

	// Run Migrations
	if err := newTenantDB.AutoMigrate(
		&models.Consent{},
		&models.ConsentHistory{},
		&models.APIKey{},
		&models.Purpose{},
		&models.DataPrincipal{},
		&models.Grievance{},
		&models.Notification{},
		&models.AuditLog{},
		&models.DSRRequest{},
		&models.TPRMAssessment{},
		&models.TPRMEvidence{},
		&models.TPRMFinding{},
	); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to migrate new tenant DB")
		h.MasterDB.Exec(fmt.Sprintf("DROP DATABASE %s", dbName))
		writeError(w, http.StatusInternalServerError, "Failed to initialize tenant schema")
		return
	}

	// 4. Create Records in MasterDB
	err = h.MasterDB.Transaction(func(tx *gorm.DB) error {
		tenant := models.Tenant{
			TenantID:    tenantID,
			Name:        req.Organization.Name,
			Industry:    req.Organization.Industry,
			CompanySize: req.Organization.CompanySize,
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}

		// Create OrganizationEntity
		org := models.OrganizationEntity{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        req.Organization.Name,
			TaxID:       req.Organization.TaxID,
			Website:     req.Organization.Website,
			Email:       req.Organization.Email,
			Phone:       req.Organization.Phone,
			CompanySize: req.Organization.CompanySize,
			Industry:    req.Organization.Industry,
			Address:     req.Organization.Address,
			Country:     req.Organization.Country,
			CreatedAt:   time.Now(),
		}
		// Use service or tx directly? Service uses repo which uses MasterDB (global).
		// If we use service, it might not be in this transaction unless we pass tx.
		// For simplicity/correctness in this refactor, let's just use tx.Create.
		if err := tx.Create(&org).Error; err != nil {
			return err
		}

		// Create FiduciaryUser
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// RBAC: Create Super Admin Role
		// Note: Roles are stored in Global DB (MasterDB) linked to TenantID?
		// Yes, models.Role has TenantID.

		var allPermissions []*models.Permission
		if err := tx.Find(&allPermissions).Error; err != nil {
			return err
		}

		superAdminRole := models.Role{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "Super Admin",
			Description: "Full access to all features and settings.",
			Permissions: allPermissions,
		}
		if err := tx.Create(&superAdminRole).Error; err != nil {
			return err
		}

		verificationToken := auth.GenerateSecureToken()
		fiduciary := models.FiduciaryUser{
			ID:                 uuid.New(),
			TenantID:           tenantID,
			Email:              req.Email,
			Name:               req.FirstName + " " + req.LastName,
			Phone:              req.Phone,
			PasswordHash:       string(hashedPassword),
			IsVerified:         false,
			VerificationToken:  verificationToken,
			VerificationExpiry: time.Now().Add(48 * time.Hour),
			Roles:              []*models.Role{&superAdminRole},
			AuthProvider:       "email",
		}
		if err := tx.Create(&fiduciary).Error; err != nil {
			return err
		}

		// CRITICAL: Send verification email BEFORE committing transaction
		// If email fails, entire signup (including DB and user record) will rollback
		verificationLink := h.Cfg.BaseURL + "/auth/verify-fiduciary?token=" + verificationToken
		emailBody := fmt.Sprintf(`
			<html>
			<body>
				<h2>Welcome to Arc Privacy Platform!</h2>
				<p>Thank you for registering your organization with us.</p>
				<p>Please verify your account by clicking the link below:</p>
				<p><a href="%s">Verify Account</a></p>
				<p>This link will expire in 48 hours.</p>
				<p>If you did not request this, please ignore this email.</p>
			</body>
			</html>
		`, verificationLink)

		if err := h.EmailService.Send(req.Email, "Verify Your Arc Privacy Account", emailBody); err != nil {
			log.Logger.Error().Err(err).Str("email", req.Email).Msg("Failed to send fiduciary verification email")
			// Return error to rollback transaction
			if errors.Is(err, services.ErrSMTPAuth) {
				return fmt.Errorf("email service authentication failed: please contact support")
			} else if errors.Is(err, services.ErrSMTPConnection) {
				return fmt.Errorf("email service unreachable: please try again later")
			}
			return fmt.Errorf("failed to send verification email: %w", err)
		}

		log.Logger.Info().Str("email", req.Email).Str("tenantId", tenantID.String()).Msg("Fiduciary signup successful, verification email sent")

		return nil
	})

	if err != nil {
		log.Logger.Error().Err(err).Msg("Fiduciary signup transaction failed")
		// Cleanup: Drop the created database since the tenant record failed
		h.MasterDB.Exec(fmt.Sprintf("DROP DATABASE %s", dbName))

		// Provide specific error messages to frontend
		errMsg := err.Error()
		if strings.Contains(errMsg, "email service authentication failed") {
			writeError(w, http.StatusServiceUnavailable, "Email service configuration error. Please contact support.")
		} else if strings.Contains(errMsg, "email service unreachable") {
			writeError(w, http.StatusServiceUnavailable, "Email service is temporarily unavailable. Please try again in a few minutes.")
		} else if strings.Contains(errMsg, "failed to send verification email") {
			writeError(w, http.StatusServiceUnavailable, "Unable to send verification email. Please try again later.")
		} else if strings.Contains(err.Error(), "idx_organization_entities_email") {
			writeError(w, http.StatusBadRequest, "An organization with this email already exists.")
		} else if strings.Contains(errMsg, "already exists") {
			writeError(w, http.StatusBadRequest, "A user with this email already exists.")
		} else {
			writeError(w, http.StatusInternalServerError, "Signup failed. Please try again.")
		}
		return
	}

	// Success response - transaction committed, email sent
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Account created successfully. Please check your email for a verification link.",
	})
}

// SignupDataPrincipal handles the creation of a new end-user (a DataPrincipal).
// This action is typically performed by an authenticated FiduciaryUser.
func (h *SignupHandler) SignupDataPrincipal(w http.ResponseWriter, r *http.Request) {
	var req DataPrincipalSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check for duplicate DataPrincipal
	var existingUser models.DataPrincipal
	if err := h.MasterDB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		writeError(w, http.StatusBadRequest, "A user with this email already exists")
		return
	}
	tenantID := uuid.New() // Placeholder

	// Handle guardian verification for minors
	isGuardianRequired := req.Age > 0 && req.Age < 18
	var guardianToken string
	var guardianTokenExpiry time.Time
	if isGuardianRequired {
		if !isValidEmail(req.GuardianEmail) {
			writeError(w, http.StatusBadRequest, "A valid guardian email is required for minors")
			return
		}
		guardianToken = auth.GenerateSecureToken()
		guardianTokenExpiry = time.Now().Add(48 * time.Hour)
	}

	// Validate password
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	// Hash the password with detailed logging
	logger.Printf("Hashing password for user: %s", req.Email)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Printf("Failed to hash password for user %s: %v", req.Email, err)
		writeError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}
	logger.Printf("Successfully hashed password for user: %s", req.Email)

	// Create DataPrincipal with detailed logging
	dataPrincipal := models.DataPrincipal{
		ID:                         uuid.New(),
		TenantID:                   tenantID,
		Email:                      req.Email,
		FirstName:                  req.FirstName,
		LastName:                   req.LastName,
		Age:                        req.Age,
		Location:                   req.Location,
		Phone:                      req.Phone,
		PasswordHash:               string(hashedPassword),
		IsVerified:                 !isGuardianRequired, // Verified unless a guardian is needed
		IsGuardianVerified:         false,
		GuardianEmail:              req.GuardianEmail,
		GuardianVerificationToken:  guardianToken,
		GuardianVerificationExpiry: guardianTokenExpiry,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	// Log the data principal creation (without sensitive data)
	logger.Printf("Creating data principal: email=%s, first_name=%s, last_name=%s, age=%d, is_verified=%v",
		dataPrincipal.Email,
		dataPrincipal.FirstName,
		dataPrincipal.LastName,
		dataPrincipal.Age,
		dataPrincipal.IsVerified,
	)

	// Create the user in a transaction with atomic email sending
	err = h.MasterDB.Transaction(func(tx *gorm.DB) error {
		// Create the user record
		if err := tx.Create(&dataPrincipal).Error; err != nil {
			logger.Printf("Failed to create data principal in database: %v", err)
			return err
		}

		logger.Printf("Successfully created user %s with ID %s",
			dataPrincipal.Email, dataPrincipal.ID)

		// CRITICAL: Send verification email BEFORE committing transaction
		// For minors, send guardian verification email
		// For adults, send regular verification email (if not auto-verified)
		if isGuardianRequired {
			verificationLink := h.Cfg.BaseURL + "/auth/verify-guardian?token=" + guardianToken
			emailBody := fmt.Sprintf(`
				<html>
				<body>
					<h2>Guardian Consent Required</h2>
					<p>A minor has registered for an Arc Privacy account and listed you as their guardian.</p>
					<p><strong>Child's Name:</strong> %s %s</p>
					<p><strong>Child's Email:</strong> %s</p>
					<p>Please verify and approve this account creation by clicking the link below:</p>
					<p><a href="%s">Approve and Verify</a></p>
					<p>This link will expire in 48 hours.</p>
					<p>If you did not expect this, please ignore this email.</p>
				</body>
				</html>
			`, dataPrincipal.FirstName, dataPrincipal.LastName, dataPrincipal.Email, verificationLink)

			if err := h.EmailService.Send(req.GuardianEmail, "Guardian Consent Required - Arc Privacy", emailBody); err != nil {
				log.Logger.Error().Err(err).Str("guardianEmail", req.GuardianEmail).Msg("Failed to send guardian verification email")
				// Return error to rollback transaction
				if errors.Is(err, services.ErrSMTPAuth) {
					return fmt.Errorf("email service authentication failed")
				} else if errors.Is(err, services.ErrSMTPConnection) {
					return fmt.Errorf("email service unreachable")
				}
				return fmt.Errorf("failed to send guardian verification email: %w", err)
			}
			log.Logger.Info().Str("guardianEmail", req.GuardianEmail).Str("userId", dataPrincipal.ID.String()).Msg("Guardian verification email sent successfully")
		}

		return nil
	})

	if err != nil {
		logger.Printf("Transaction failed during user creation: %v", err)

		// Provide specific error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "email service authentication failed") {
			writeError(w, http.StatusServiceUnavailable, "Email service configuration error. Please contact support.")
		} else if strings.Contains(errMsg, "email service unreachable") {
			writeError(w, http.StatusServiceUnavailable, "Email service is temporarily unavailable. Please try again in a few minutes.")
		} else if strings.Contains(errMsg, "failed to send") && strings.Contains(errMsg, "email") {
			writeError(w, http.StatusServiceUnavailable, "Unable to send verification email. Please try again later.")
		} else if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
			writeError(w, http.StatusBadRequest, "A user with this email already exists.")
		} else {
			writeError(w, http.StatusInternalServerError, "Could not create user. Please try again.")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"dataPrincipalId": dataPrincipal.ID,
		"message":         "Data principal created. If a minor, a verification email has been sent to the guardian.",
	})
}

// VerifyFiduciary handles the token-based verification for a new FiduciaryUser.
func (h *SignupHandler) VerifyFiduciary(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "Verification token is missing")
		return
	}

	var user models.FiduciaryUser
	if err := h.MasterDB.Where("verification_token = ?", token).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "Invalid or expired verification token")
		return
	}

	if time.Now().After(user.VerificationExpiry) {
		writeError(w, http.StatusBadRequest, "Verification token has expired")
		return
	}

	user.IsVerified = true
	user.VerificationToken = "" // Clear token after use
	if err := h.MasterDB.Save(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update user verification status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Account successfully verified."})
}

// VerifyGuardian handles token-based verification for a minor's guardian.
func (h *SignupHandler) VerifyGuardian(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "Verification token is missing")
		return
	}

	var user models.DataPrincipal
	if err := h.MasterDB.Where("guardian_verification_token = ?", token).First(&user).Error; err != nil {
		writeError(w, http.StatusNotFound, "Invalid or expired verification token")
		return
	}

	if time.Now().After(user.GuardianVerificationExpiry) {
		writeError(w, http.StatusBadRequest, "Verification token has expired")
		return
	}

	user.IsGuardianVerified = true
	user.IsVerified = true              // The user account is now fully active
	user.GuardianVerificationToken = "" // Clear token
	if err := h.MasterDB.Save(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update user verification status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Guardian verified. The user's account is now active."})
}

// ===== Helper validation functions =====
func validateFiduciaryRequest(req *FiduciarySignupRequest, w http.ResponseWriter) bool {
	if !isValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "Invalid email format")
		return false
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return false
	}
	if req.Organization.Name == "" || req.FirstName == "" || req.LastName == "" {
		writeError(w, http.StatusBadRequest, "Company and user name are required")
		return false
	}
	return true
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	return re.MatchString(email)
}
