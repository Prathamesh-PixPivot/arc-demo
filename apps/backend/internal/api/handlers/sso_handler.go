package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pixpivot/arc/config"
	"pixpivot/arc/internal/auth"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/pkg/log"

	"crypto/rsa"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
	"gorm.io/gorm"
)

type SSOHandler struct {
	MasterDB   *gorm.DB
	Cfg        config.Config
	PrivateKey *rsa.PrivateKey
}

func NewSSOHandler(masterDB *gorm.DB, cfg config.Config, privateKey *rsa.PrivateKey) *SSOHandler {
	if masterDB == nil {
		log.Logger.Error().Msg("NewSSOHandler received nil masterDB")
	} else {
		log.Logger.Info().Msg("NewSSOHandler received valid masterDB")
	}
	return &SSOHandler{
		MasterDB:   masterDB,
		Cfg:        cfg,
		PrivateKey: privateKey,
	}
}

// SSOState encodes the authentication flow parameters
type SSOState struct {
	Mode     string `json:"mode"`     // "login" or "signup"
	UserType string `json:"userType"` // "user" or "fiduciary"
}

// ==================== Google SSO ====================

func (h *SSOHandler) googleConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.Cfg.GoogleClientID,
		ClientSecret: h.Cfg.GoogleClientSecret,
		RedirectURL:  h.Cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (h *SSOHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	userType := r.URL.Query().Get("userType")

	// Default to login if not specified
	if mode == "" {
		mode = "login"
	}

	state := SSOState{
		Mode:     mode,
		UserType: userType,
	}

	stateBytes, err := json.Marshal(state)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to marshal SSO state")
		writeError(w, http.StatusInternalServerError, "Failed to create state")
		return
	}

	url := h.googleConfig().AuthCodeURL(string(stateBytes), oauth2.AccessTypeOffline)
	log.Logger.Info().Str("mode", mode).Str("userType", userType).Msg("Redirecting to Google SSO")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *SSOHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	stateStr := r.URL.Query().Get("state")

	if code == "" {
		log.Logger.Error().Msg("Google callback: code not found")
		writeError(w, http.StatusBadRequest, "Code not found")
		return
	}

	var state SSOState
	if err := json.Unmarshal([]byte(stateStr), &state); err != nil {
		log.Logger.Warn().Err(err).Msg("Failed to parse SSO state, defaulting to login")
		state = SSOState{Mode: "login"}
	}

	token, err := h.googleConfig().Exchange(context.Background(), code)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Google token exchange failed")
		writeError(w, http.StatusInternalServerError, "Token exchange failed")
		return
	}

	client := h.googleConfig().Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get user info from Google")
		writeError(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to decode Google user info")
		writeError(w, http.StatusInternalServerError, "Failed to decode user info")
		return
	}

	log.Logger.Info().Str("email", userInfo.Email).Str("provider", "google").Msg("Google SSO callback successful")
	h.handleSSOFlow(w, r, userInfo.Email, userInfo.Name, "google", userInfo.ID, state)
}

// ==================== Microsoft SSO ====================

func (h *SSOHandler) microsoftConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.Cfg.MicrosoftClientID,
		ClientSecret: h.Cfg.MicrosoftClientSecret,
		RedirectURL:  h.Cfg.MicrosoftRedirectURL,
		Scopes:       []string{"User.Read"},
		Endpoint:     microsoft.AzureADEndpoint(h.Cfg.MicrosoftTenantID),
	}
}

func (h *SSOHandler) MicrosoftLogin(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	userType := r.URL.Query().Get("userType")

	if mode == "" {
		mode = "login"
	}

	state := SSOState{
		Mode:     mode,
		UserType: userType,
	}

	stateBytes, err := json.Marshal(state)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to marshal SSO state")
		writeError(w, http.StatusInternalServerError, "Failed to create state")
		return
	}

	url := h.microsoftConfig().AuthCodeURL(string(stateBytes), oauth2.AccessTypeOffline)
	log.Logger.Info().Str("mode", mode).Str("userType", userType).Msg("Redirecting to Microsoft SSO")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *SSOHandler) MicrosoftCallback(w http.ResponseWriter, r *http.Request) {
	log.Logger.Info().Msg("MicrosoftCallback: Started")
	code := r.URL.Query().Get("code")
	stateStr := r.URL.Query().Get("state")

	if code == "" {
		log.Logger.Error().Msg("Microsoft callback: code not found")
		writeError(w, http.StatusBadRequest, "Code not found")
		return
	}

	var state SSOState
	if err := json.Unmarshal([]byte(stateStr), &state); err != nil {
		log.Logger.Warn().Err(err).Msg("Failed to parse SSO state, defaulting to login")
		state = SSOState{Mode: "login"}
	}

	token, err := h.microsoftConfig().Exchange(context.Background(), code)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Microsoft token exchange failed")
		writeError(w, http.StatusInternalServerError, "Token exchange failed")
		return
	}

	client := h.microsoftConfig().Client(context.Background(), token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get user info from Microsoft")
		writeError(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID                string `json:"id"`
		UserPrincipalName string `json:"userPrincipalName"`
		DisplayName       string `json:"displayName"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.Logger.Error().Err(err).Msg("Failed to decode Microsoft user info")
		writeError(w, http.StatusInternalServerError, "Failed to decode user info")
		return
	}

	log.Logger.Info().Str("email", userInfo.UserPrincipalName).Str("provider", "microsoft").Msg("Microsoft SSO callback successful")
	h.handleSSOFlow(w, r, userInfo.UserPrincipalName, userInfo.DisplayName, "microsoft", userInfo.ID, state)
}

// ==================== Common SSO Flow Logic ====================

func (h *SSOHandler) handleSSOFlow(w http.ResponseWriter, r *http.Request, email, name, provider, providerID string, state SSOState) {
	log.Logger.Info().
		Str("email", email).
		Str("provider", provider).
		Str("mode", state.Mode).
		Str("userType", state.UserType).
		Msg("Processing SSO flow")

	// ========== LOGIN MODE ==========
	if state.Mode == "login" {
		h.handleSSOLogin(w, r, email, provider, providerID, state)
		return
	}

	// ========== SIGNUP MODE ==========
	if state.Mode == "signup" {
		h.handleSSOSignup(w, r, email, name, provider, providerID, state)
		return
	}

	// Invalid mode
	log.Logger.Error().Str("mode", state.Mode).Msg("Invalid SSO mode")
	redirectURL := fmt.Sprintf("%s/login?error=invalid_mode", h.Cfg.FrontendBaseURL)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// handleSSOLogin - STRICT: Block if user doesn't exist
func (h *SSOHandler) handleSSOLogin(w http.ResponseWriter, r *http.Request, email, provider, providerID string, state SSOState) {
	// Try Fiduciary table first
	var fiduciary models.FiduciaryUser
	errFid := h.MasterDB.Where("email = ? AND auth_provider = ?", email, provider).First(&fiduciary).Error

	if errFid == nil {
		// Found in Fiduciary - generate token and redirect
		log.Logger.Info().Str("email", email).Msg("SSO Login: Fiduciary user found")
		token, err := auth.GenerateFiduciaryToken(fiduciary, h.PrivateKey, h.Cfg.UserTokenTTL)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to generate fiduciary token")
			writeError(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}
		redirectURL := fmt.Sprintf("%s/auth/callback?token=%s&userType=fiduciary", h.Cfg.FrontendBaseURL, token)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Try DataPrincipal table
	var dataPrincipal models.DataPrincipal
	errDP := h.MasterDB.Where("email = ? AND auth_provider = ?", email, provider).First(&dataPrincipal).Error

	if errDP == nil {
		// Found in DataPrincipal - generate token and redirect
		log.Logger.Info().Str("email", email).Msg("SSO Login: Data Principal found")
		token, err := auth.GenerateDataPrincipalToken(dataPrincipal, h.PrivateKey, h.Cfg.UserTokenTTL)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to generate data principal token")
			writeError(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}
		redirectURL := fmt.Sprintf("%s/auth/callback?token=%s&userType=user", h.Cfg.FrontendBaseURL, token)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// *** STRICT MODE: User not found, BLOCK login ***
	log.Logger.Warn().Str("email", email).Str("provider", provider).Msg("SSO Login blocked: Account not registered")
	redirectURL := fmt.Sprintf("%s/login?error=not_registered&hint=Please+sign+up+first", h.Cfg.FrontendBaseURL)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// handleSSOSignup - Create user and redirect to onboarding
func (h *SSOHandler) handleSSOSignup(w http.ResponseWriter, r *http.Request, email, name, provider, providerID string, state SSOState) {
	// Check if already exists (prevent duplicate)
	var existsFiduciary, existsDataPrincipal bool
	var countFid, countDP int64

	h.MasterDB.Model(&models.FiduciaryUser{}).Where("email = ?", email).Count(&countFid)
	existsFiduciary = countFid > 0

	h.MasterDB.Model(&models.DataPrincipal{}).Where("email = ?", email).Count(&countDP)
	existsDataPrincipal = countDP > 0

	if existsFiduciary || existsDataPrincipal {
		log.Logger.Info().Str("email", email).Msg("SSO Signup: User already exists, redirecting to login")
		redirectURL := fmt.Sprintf("%s/login?info=already_registered", h.Cfg.FrontendBaseURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Create new user based on type
	if state.UserType == "fiduciary" {
		newUser := models.FiduciaryUser{
			ID:           uuid.New(),
			Email:        email,
			Name:         name,
			AuthProvider: provider,
			ProviderID:   providerID,
			IsVerified:   true, // SSO users are auto-verified
			CreatedAt:    time.Now(),
			Role:         "admin", // Default role for SSO signup
		}

		if err := h.MasterDB.Create(&newUser).Error; err != nil {
			log.Logger.Error().Err(err).Str("email", email).Msg("Failed to create SSO fiduciary user")
			writeError(w, http.StatusInternalServerError, "Failed to create account")
			return
		}

		log.Logger.Info().Str("email", email).Str("userId", newUser.ID.String()).Msg("SSO Signup: Fiduciary user created")

		// Generate token and redirect to organization onboarding
		token, err := auth.GenerateFiduciaryToken(newUser, h.PrivateKey, h.Cfg.UserTokenTTL)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to generate token for new fiduciary")
			writeError(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		redirectURL := fmt.Sprintf("%s/auth/callback?token=%s&userType=fiduciary&next=/onboarding/organization", h.Cfg.FrontendBaseURL, token)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)

	} else if state.UserType == "user" {
		// Create Data Principal (NOTE: AuthProvider/ProviderID not in model yet)
		newUser := models.DataPrincipal{
			ID:         uuid.New(),
			TenantID:   uuid.Nil, // Will be assigned later during onboarding or invitation
			Email:      email,
			FirstName:  name, // SSO provides full name, we'll use it as FirstName for now
			IsVerified: true, // SSO users are auto-verified
			CreatedAt:  time.Now(),
		}

		if err := h.MasterDB.Create(&newUser).Error; err != nil {
			log.Logger.Error().Err(err).Str("email", email).Msg("Failed to create SSO data principal")
			writeError(w, http.StatusInternalServerError, "Failed to create account")
			return
		}

		log.Logger.Info().Str("email", email).Str("userId", newUser.ID.String()).Msg("SSO Signup: Data Principal created")

		// Generate token and redirect to profile onboarding
		token, err := auth.GenerateDataPrincipalToken(newUser, h.PrivateKey, h.Cfg.UserTokenTTL)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to generate token for new data principal")
			writeError(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		redirectURL := fmt.Sprintf("%s/auth/callback?token=%s&userType=user&next=/onboarding/profile", h.Cfg.FrontendBaseURL, token)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)

	} else {
		log.Logger.Error().Str("userType", state.UserType).Msg("Invalid user type for SSO signup")
		redirectURL := fmt.Sprintf("%s/signup?error=invalid_user_type", h.Cfg.FrontendBaseURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}
