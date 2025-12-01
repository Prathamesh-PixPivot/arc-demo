package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"pixpivot/arc/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TranslationRepository handles database operations for translations
type TranslationRepository struct {
	db *gorm.DB
}

// NewTranslationRepository creates a new translation repository
func NewTranslationRepository(db *gorm.DB) *TranslationRepository {
	return &TranslationRepository{db: db}
}

// GetTranslation retrieves a translation by language code and key
func (r *TranslationRepository) GetTranslation(ctx context.Context, languageCode, key string) (*models.Translation, error) {
	var translation models.Translation
	err := r.db.WithContext(ctx).
		Where("language_code = ? AND key = ? AND is_active = ?", languageCode, key, true).
		First(&translation).Error
	
	if err != nil {
		return nil, fmt.Errorf("translation not found for %s:%s: %w", languageCode, key, err)
	}
	
	return &translation, nil
}

// GetTranslationByID retrieves a translation by ID
func (r *TranslationRepository) GetTranslationByID(ctx context.Context, id uuid.UUID) (*models.Translation, error) {
	var translation models.Translation
	err := r.db.WithContext(ctx).First(&translation, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("translation not found: %w", err)
	}
	return &translation, nil
}

// GetBulkTranslations retrieves multiple translations at once
func (r *TranslationRepository) GetBulkTranslations(ctx context.Context, languageCode string, keys []string) ([]models.Translation, error) {
	var translations []models.Translation
	err := r.db.WithContext(ctx).
		Where("language_code = ? AND key IN ? AND is_active = ?", languageCode, keys, true).
		Find(&translations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk translations: %w", err)
	}
	
	return translations, nil
}

// GetAllTranslations retrieves all translations for a language
func (r *TranslationRepository) GetAllTranslations(ctx context.Context, languageCode string) ([]models.Translation, error) {
	var translations []models.Translation
	err := r.db.WithContext(ctx).
		Where("language_code = ? AND is_active = ?", languageCode, true).
		Order("key").
		Find(&translations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get all translations: %w", err)
	}
	
	return translations, nil
}

// CreateTranslation creates a new translation
func (r *TranslationRepository) CreateTranslation(ctx context.Context, translation *models.Translation) error {
	if translation.ID == uuid.Nil {
		translation.ID = uuid.New()
	}
	
	err := r.db.WithContext(ctx).Create(translation).Error
	if err != nil {
		return fmt.Errorf("failed to create translation: %w", err)
	}
	
	return nil
}

// UpdateTranslation updates an existing translation
func (r *TranslationRepository) UpdateTranslation(ctx context.Context, translation *models.Translation) error {
	err := r.db.WithContext(ctx).Save(translation).Error
	if err != nil {
		return fmt.Errorf("failed to update translation: %w", err)
	}
	return nil
}

// UpsertTranslation creates or updates a translation
func (r *TranslationRepository) UpsertTranslation(ctx context.Context, translation *models.Translation) error {
	if translation.ID == uuid.Nil {
		translation.ID = uuid.New()
	}
	
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "language_code"}, {Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "context", "metadata", "updated_at"}),
		}).
		Create(translation).Error
	
	if err != nil {
		return fmt.Errorf("failed to upsert translation: %w", err)
	}
	
	return nil
}

// DeleteTranslation deletes a translation
func (r *TranslationRepository) DeleteTranslation(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).Delete(&models.Translation{}, "id = ?", id).Error
	if err != nil {
		return fmt.Errorf("failed to delete translation: %w", err)
	}
	return nil
}

// GetActiveLanguages retrieves all active languages
func (r *TranslationRepository) GetActiveLanguages(ctx context.Context) ([]models.Language, error) {
	var languages []models.Language
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order, name").
		Find(&languages).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get active languages: %w", err)
	}
	
	return languages, nil
}

// GetLanguageByCode retrieves a language by code
func (r *TranslationRepository) GetLanguageByCode(ctx context.Context, code string) (*models.Language, error) {
	var language models.Language
	err := r.db.WithContext(ctx).First(&language, "code = ?", code).Error
	if err != nil {
		return nil, fmt.Errorf("language not found: %w", err)
	}
	return &language, nil
}

// CreateLanguage creates a new language
func (r *TranslationRepository) CreateLanguage(ctx context.Context, language *models.Language) error {
	err := r.db.WithContext(ctx).Create(language).Error
	if err != nil {
		return fmt.Errorf("failed to create language: %w", err)
	}
	return nil
}

// UpdateLanguage updates a language
func (r *TranslationRepository) UpdateLanguage(ctx context.Context, language *models.Language) error {
	err := r.db.WithContext(ctx).Save(language).Error
	if err != nil {
		return fmt.Errorf("failed to update language: %w", err)
	}
	return nil
}

// GetTenantTranslation retrieves a tenant-specific translation
func (r *TranslationRepository) GetTenantTranslation(ctx context.Context, tenantID uuid.UUID, languageCode, key string) (*models.TenantTranslation, error) {
	var translation models.TenantTranslation
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND language_code = ? AND key = ?", tenantID, languageCode, key).
		First(&translation).Error
	
	if err != nil {
		return nil, fmt.Errorf("tenant translation not found: %w", err)
	}
	
	return &translation, nil
}

// CreateTenantTranslation creates a tenant-specific translation
func (r *TranslationRepository) CreateTenantTranslation(ctx context.Context, translation *models.TenantTranslation) error {
	if translation.ID == uuid.Nil {
		translation.ID = uuid.New()
	}
	
	err := r.db.WithContext(ctx).Create(translation).Error
	if err != nil {
		return fmt.Errorf("failed to create tenant translation: %w", err)
	}
	
	return nil
}

// UpdateTenantTranslation updates a tenant-specific translation
func (r *TranslationRepository) UpdateTenantTranslation(ctx context.Context, translation *models.TenantTranslation) error {
	err := r.db.WithContext(ctx).Save(translation).Error
	if err != nil {
		return fmt.Errorf("failed to update tenant translation: %w", err)
	}
	return nil
}

// GetUserLanguagePreference retrieves a user's language preference
func (r *TranslationRepository) GetUserLanguagePreference(ctx context.Context, userID uuid.UUID) (*models.UserLanguagePreference, error) {
	var pref models.UserLanguagePreference
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pref).Error
	if err != nil {
		return nil, fmt.Errorf("user language preference not found: %w", err)
	}
	return &pref, nil
}

// UpsertUserLanguagePreference creates or updates a user's language preference
func (r *TranslationRepository) UpsertUserLanguagePreference(ctx context.Context, pref *models.UserLanguagePreference) error {
	if pref.ID == uuid.Nil {
		pref.ID = uuid.New()
	}
	
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"language_code", "updated_at"}),
		}).
		Create(pref).Error
	
	if err != nil {
		return fmt.Errorf("failed to upsert user language preference: %w", err)
	}
	
	return nil
}

// GetTranslationTemplates retrieves all translation templates
func (r *TranslationRepository) GetTranslationTemplates(ctx context.Context) ([]models.TranslationTemplate, error) {
	var templates []models.TranslationTemplate
	err := r.db.WithContext(ctx).Order("category, key").Find(&templates).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get translation templates: %w", err)
	}
	return templates, nil
}

// GetTranslationTemplate retrieves a translation template by key
func (r *TranslationRepository) GetTranslationTemplate(ctx context.Context, key string) (*models.TranslationTemplate, error) {
	var template models.TranslationTemplate
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&template).Error
	if err != nil {
		return nil, fmt.Errorf("translation template not found: %w", err)
	}
	return &template, nil
}

// CreateTranslationTemplate creates a new translation template
func (r *TranslationRepository) CreateTranslationTemplate(ctx context.Context, template *models.TranslationTemplate) error {
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	
	err := r.db.WithContext(ctx).Create(template).Error
	if err != nil {
		return fmt.Errorf("failed to create translation template: %w", err)
	}
	
	return nil
}

// GetTotalTranslationKeys gets the total number of unique translation keys
func (r *TranslationRepository) GetTotalTranslationKeys(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Translation{}).
		Select("COUNT(DISTINCT key)").
		Where("language_code = ?", "en"). // Count from default language
		Scan(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to get total translation keys: %w", err)
	}
	
	return count, nil
}

// GetTranslatedKeysCount gets the number of translated keys for a language
func (r *TranslationRepository) GetTranslatedKeysCount(ctx context.Context, languageCode string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Translation{}).
		Where("language_code = ? AND is_active = ?", languageCode, true).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to get translated keys count: %w", err)
	}
	
	return count, nil
}

// SearchTranslations searches translations by key or value
func (r *TranslationRepository) SearchTranslations(ctx context.Context, languageCode, query string, limit int) ([]models.Translation, error) {
	var translations []models.Translation
	
	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("language_code = ? AND (key LIKE ? OR value LIKE ?) AND is_active = ?", 
			languageCode, searchQuery, searchQuery, true).
		Limit(limit).
		Find(&translations).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to search translations: %w", err)
	}
	
	return translations, nil
}

// GetMissingTranslations gets keys that are missing translations for a language
func (r *TranslationRepository) GetMissingTranslations(ctx context.Context, languageCode string) ([]string, error) {
	var missingKeys []string
	
	// Get all keys from default language (English)
	var defaultKeys []string
	err := r.db.WithContext(ctx).
		Model(&models.Translation{}).
		Where("language_code = ? AND is_active = ?", "en", true).
		Pluck("key", &defaultKeys).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get default keys: %w", err)
	}
	
	// Get translated keys for the target language
	var translatedKeys []string
	err = r.db.WithContext(ctx).
		Model(&models.Translation{}).
		Where("language_code = ? AND is_active = ?", languageCode, true).
		Pluck("key", &translatedKeys).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get translated keys: %w", err)
	}
	
	// Find missing keys
	translatedMap := make(map[string]bool)
	for _, key := range translatedKeys {
		translatedMap[key] = true
	}
	
	for _, key := range defaultKeys {
		if !translatedMap[key] {
			missingKeys = append(missingKeys, key)
		}
	}
	
	return missingKeys, nil
}

