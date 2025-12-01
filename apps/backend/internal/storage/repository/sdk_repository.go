package repository

import (
	"pixpivot/arc/internal/models"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SDKRepository struct {
	db    *gorm.DB
	cache map[string]cacheItem
	mutex sync.RWMutex
}

type cacheItem struct {
	value     string
	expiresAt time.Time
}

func NewSDKRepository(db *gorm.DB) *SDKRepository {
	return &SDKRepository{
		db:    db,
		cache: make(map[string]cacheItem),
		mutex: sync.RWMutex{},
	}
}

func (r *SDKRepository) Create(config *models.SDKConfig) error {
	return r.db.Create(config).Error
}

func (r *SDKRepository) Update(config *models.SDKConfig) error {
	return r.db.Save(config).Error
}

func (r *SDKRepository) GetByID(id uuid.UUID) (*models.SDKConfig, error) {
	var config models.SDKConfig
	err := r.db.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *SDKRepository) GetByFormID(tenantID, formID uuid.UUID) (*models.SDKConfig, error) {
	var config models.SDKConfig
	err := r.db.Where("tenant_id = ? AND consent_form_id = ?", tenantID, formID).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *SDKRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.SDKConfig{}, id).Error
}

func (r *SDKRepository) ListByTenant(tenantID uuid.UUID) ([]*models.SDKConfig, error) {
	var configs []*models.SDKConfig
	err := r.db.Where("tenant_id = ?", tenantID).Find(&configs).Error
	return configs, err
}

// Cache operations (simplified in-memory cache)
// In production, you'd want to use Redis or similar
func (r *SDKRepository) CacheSDK(key, value string, ttl time.Duration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.cache[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	
	return nil
}

func (r *SDKRepository) GetCachedSDK(key string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	item, exists := r.cache[key]
	if !exists {
		return "", fmt.Errorf("cache miss")
	}
	
	if time.Now().After(item.expiresAt) {
		// Clean up expired item
		delete(r.cache, key)
		return "", fmt.Errorf("cache expired")
	}
	
	return item.value, nil
}

func (r *SDKRepository) InvalidateCache(key string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	delete(r.cache, key)
	return nil
}

// Clean up expired cache entries periodically
func (r *SDKRepository) CleanupExpiredCache() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	now := time.Now()
	for key, item := range r.cache {
		if now.After(item.expiresAt) {
			delete(r.cache, key)
		}
	}
}

