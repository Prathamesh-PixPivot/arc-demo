package repository

import (
	"fmt"
	"pixpivot/arc/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurposeRepository struct {
	db *gorm.DB
}

func NewPurposeRepository(db *gorm.DB) *PurposeRepository {
	return &PurposeRepository{db: db}
}

// GetPurposeByID retrieves a purpose by its ID
func (r *PurposeRepository) GetPurposeByID(purposeID, tenantID uuid.UUID) (*models.Purpose, error) {
	var purpose models.Purpose
	err := r.db.Where("id = ? AND tenant_id = ?", purposeID, tenantID).First(&purpose).Error
	if err != nil {
		return nil, err
	}
	return &purpose, nil
}

// GetPurposesByIDs retrieves multiple purposes by their IDs
func (r *PurposeRepository) GetPurposesByIDs(purposeIDs []uuid.UUID, tenantID uuid.UUID) ([]models.Purpose, error) {
	var purposes []models.Purpose
	err := r.db.Where("id IN ? AND tenant_id = ?", purposeIDs, tenantID).Find(&purposes).Error
	return purposes, err
}

// GetPurposesByTenant retrieves all purposes for a tenant
func (r *PurposeRepository) GetPurposesByTenant(tenantID uuid.UUID) ([]models.Purpose, error) {
	var purposes []models.Purpose
	err := r.db.Where("tenant_id = ?", tenantID).Find(&purposes).Error
	return purposes, err
}

// CreatePurpose creates a new purpose
func (r *PurposeRepository) CreatePurpose(purpose *models.Purpose) error {
	return r.db.Create(purpose).Error
}

// UpdatePurpose updates an existing purpose
func (r *PurposeRepository) UpdatePurpose(purpose *models.Purpose) error {
	return r.db.Where("id = ? AND tenant_id = ?", purpose.ID, purpose.TenantID).Save(purpose).Error
}

// DeletePurpose deletes a purpose
func (r *PurposeRepository) DeletePurpose(purposeID, tenantID uuid.UUID) error {
	return r.db.Where("id = ? AND tenant_id = ?", purposeID, tenantID).Delete(&models.Purpose{}).Error
}

// GetPurposeHierarchy retrieves the complete purpose hierarchy for a tenant
func (r *PurposeRepository) GetPurposeHierarchy(tenantID uuid.UUID) ([]models.PurposeTree, error) {
	var purposes []models.Purpose
	err := r.db.Where("tenant_id = ? AND active = ?", tenantID, true).
		Order("parent_purpose_id NULLS FIRST, name").
		Find(&purposes).Error
	if err != nil {
		return nil, err
	}

	// Build hierarchy tree
	purposeMap := make(map[uuid.UUID]*models.PurposeTree)
	var rootPurposes []*models.PurposeTree

	// First pass: create all nodes
	for _, purpose := range purposes {
		tree := &models.PurposeTree{
			Purpose:  purpose,
			Children: []*models.PurposeTree{},
		}
		purposeMap[purpose.ID] = tree

		if purpose.ParentPurposeID == nil {
			rootPurposes = append(rootPurposes, tree)
		}
	}

	// Second pass: build parent-child relationships
	for _, purpose := range purposes {
		if purpose.ParentPurposeID != nil {
			if parent, exists := purposeMap[*purpose.ParentPurposeID]; exists {
				if child, exists := purposeMap[purpose.ID]; exists {
					parent.Children = append(parent.Children, child)
				}
			}
		}
	}

	var result []models.PurposeTree
	for _, root := range rootPurposes {
		result = append(result, *root)
	}
	return result, nil
}

// ValidateHierarchy checks for circular references in purpose hierarchy
func (r *PurposeRepository) ValidateHierarchy(purposeID uuid.UUID, parentID *uuid.UUID) error {
	if parentID == nil {
		return nil
	}

	if purposeID == *parentID {
		return fmt.Errorf("purpose cannot be its own parent")
	}

	// Check for circular reference by traversing up the hierarchy
	visited := make(map[uuid.UUID]bool)
	currentID := parentID

	for currentID != nil {
		if visited[*currentID] {
			return fmt.Errorf("circular reference detected in purpose hierarchy")
		}
		visited[*currentID] = true

		if *currentID == purposeID {
			return fmt.Errorf("circular reference: purpose %s would become ancestor of itself", purposeID)
		}

		var parent models.Purpose
		// Note: We don't strictly need tenantID here if we assume DB integrity,
		// but ideally we should pass it down. For now, ID lookup is unique enough for validation logic.
		err := r.db.Select("parent_purpose_id").Where("id = ?", *currentID).First(&parent).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return err
		}
		currentID = parent.ParentPurposeID
	}

	return nil
}

// GetChildPurposes retrieves all child purposes for a given parent
func (r *PurposeRepository) GetChildPurposes(parentID, tenantID uuid.UUID) ([]models.Purpose, error) {
	var purposes []models.Purpose
	err := r.db.Where("parent_purpose_id = ? AND tenant_id = ? AND active = ?", parentID, tenantID, true).Find(&purposes).Error
	return purposes, err
}

// GetPurposeDepth calculates the depth of a purpose in the hierarchy
func (r *PurposeRepository) GetPurposeDepth(purposeID uuid.UUID) (int, error) {
	depth := 0
	currentID := &purposeID

	for currentID != nil {
		var purpose models.Purpose
		err := r.db.Select("parent_purpose_id").Where("id = ?", *currentID).First(&purpose).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return 0, err
		}

		if purpose.ParentPurposeID == nil {
			break
		}

		depth++
		currentID = purpose.ParentPurposeID

		// Prevent infinite loops
		if depth > 10 {
			return 0, fmt.Errorf("purpose hierarchy too deep, possible circular reference")
		}
	}

	return depth, nil
}

// InheritDataObjects inherits data objects from parent purposes
func (r *PurposeRepository) InheritDataObjects(purposeID uuid.UUID) ([]string, error) {
	var dataObjects []string
	currentID := &purposeID

	for currentID != nil {
		var purpose models.Purpose
		err := r.db.Where("id = ?", *currentID).First(&purpose).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return nil, err
		}

		// Add current purpose's data objects (this would need to be implemented based on your data model)
		// For now, we'll just return empty slice as the Purpose model doesn't have DataObjects field

		if purpose.ParentPurposeID == nil {
			break
		}
		currentID = purpose.ParentPurposeID
	}

	return dataObjects, nil
}

// UpdateComplianceStatus updates the compliance status of a purpose
func (r *PurposeRepository) UpdateComplianceStatus(purposeID uuid.UUID, status string) error {
	return r.db.Model(&models.Purpose{}).
		Where("id = ?", purposeID).
		Updates(map[string]interface{}{
			"compliance_status":     status,
			"last_compliance_check": gorm.Expr("NOW()"),
		}).Error
}

// GetPurposesForCompliance retrieves purposes that need compliance checking
func (r *PurposeRepository) GetPurposesForCompliance(tenantID uuid.UUID) ([]models.Purpose, error) {
	var purposes []models.Purpose
	err := r.db.Where("tenant_id = ? AND (last_compliance_check IS NULL OR last_compliance_check < NOW() - INTERVAL '30 days')", tenantID).
		Find(&purposes).Error
	return purposes, err
}

// GetPurposeUsageStats retrieves usage statistics for purposes
func (r *PurposeRepository) GetPurposeUsageStats(tenantID uuid.UUID) ([]models.UsageStats, error) {
	var stats []models.UsageStats

	query := `
		SELECT 
			p.id as purpose_id,
			COALESCE(active_consents.count, 0) as active_consents,
			COALESCE(consent_forms.count, 0) as consent_forms,
			COALESCE(expiring_consents.count, 0) as expiring_consents,
			p.total_granted,
			p.total_revoked,
			p.last_used_at as last_used
		FROM purposes p
		LEFT JOIN (
			SELECT purpose_id, COUNT(*) as count
			FROM user_consents 
			WHERE status = true AND (expires_at IS NULL OR expires_at > NOW())
			GROUP BY purpose_id
		) active_consents ON p.id = active_consents.purpose_id
		LEFT JOIN (
			SELECT cfp.purpose_id, COUNT(DISTINCT cfp.consent_form_id) as count
			FROM consent_form_purposes cfp
			INNER JOIN consent_forms cf ON cfp.consent_form_id = cf.id
			WHERE cf.published = true
			GROUP BY cfp.purpose_id
		) consent_forms ON p.id = consent_forms.purpose_id
		LEFT JOIN (
			SELECT purpose_id, COUNT(*) as count
			FROM user_consents 
			WHERE status = true AND expires_at BETWEEN NOW() AND NOW() + INTERVAL '30 days'
			GROUP BY purpose_id
		) expiring_consents ON p.id = expiring_consents.purpose_id
		WHERE p.tenant_id = ?
		ORDER BY p.name
	`

	err := r.db.Raw(query, tenantID).Scan(&stats).Error
	return stats, err
}

// GetPurposeTree retrieves the hierarchy for a specific purpose
func (r *PurposeRepository) GetPurposeTree(purposeID, tenantID uuid.UUID) (*models.PurposeTree, error) {
	// 1. Get the purpose
	purpose, err := r.GetPurposeByID(purposeID, tenantID)
	if err != nil {
		return nil, err
	}

	// 2. Get all purposes for the tenant to build the tree
	allPurposes, err := r.GetPurposesByTenant(purpose.TenantID)
	if err != nil {
		return nil, err
	}

	// Build map
	purposeMap := make(map[uuid.UUID]*models.PurposeTree)
	for _, p := range allPurposes {
		purposeMap[p.ID] = &models.PurposeTree{
			Purpose:  p,
			Children: []*models.PurposeTree{},
		}
	}

	// Build links
	for _, p := range allPurposes {
		if p.ParentPurposeID != nil {
			if parent, ok := purposeMap[*p.ParentPurposeID]; ok {
				if child, ok := purposeMap[p.ID]; ok {
					parent.Children = append(parent.Children, child)
				}
			}
		}
	}

	// Return the specific node
	if node, ok := purposeMap[purposeID]; ok {
		return node, nil
	}
	return nil, fmt.Errorf("purpose not found in tree")
}

// GetAllRootPurposes retrieves all root purposes for a tenant
func (r *PurposeRepository) GetAllRootPurposes(tenantID uuid.UUID) ([]models.Purpose, error) {
	var purposes []models.Purpose
	err := r.db.Where("tenant_id = ? AND parent_purpose_id IS NULL", tenantID).Find(&purposes).Error
	return purposes, err
}

// SetParentPurpose sets the parent of a purpose
func (r *PurposeRepository) SetParentPurpose(childID, parentID, tenantID uuid.UUID) error {
	// Verify parent exists and belongs to tenant
	var parent models.Purpose
	if err := r.db.First(&parent, "id = ? AND tenant_id = ?", parentID, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("parent purpose not found")
		}
		return err
	}

	return r.db.Model(&models.Purpose{}).Where("id = ? AND tenant_id = ?", childID, tenantID).Update("parent_purpose_id", parentID).Error
}

// GetPurposeAncestors retrieves all ancestors of a purpose
func (r *PurposeRepository) GetPurposeAncestors(purposeID, tenantID uuid.UUID) ([]models.Purpose, error) {
	var ancestors []models.Purpose
	currentID := &purposeID

	// First get the purpose to find its parent
	var purpose models.Purpose
	if err := r.db.First(&purpose, "id = ? AND tenant_id = ?", purposeID, tenantID).Error; err != nil {
		return nil, err
	}
	currentID = purpose.ParentPurposeID

	for currentID != nil {
		var parent models.Purpose
		// We can assume ancestors are in same tenant, but good to check or just trust ID if we trust the chain
		// For strictness, we check tenantID
		err := r.db.First(&parent, "id = ? AND tenant_id = ?", *currentID, tenantID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				break
			}
			return nil, err
		}
		ancestors = append(ancestors, parent)
		currentID = parent.ParentPurposeID

		// Safety break
		if len(ancestors) > 20 {
			break
		}
	}
	return ancestors, nil
}

// GetPurposeDescendants retrieves all descendants of a purpose (recursive)
func (r *PurposeRepository) GetPurposeDescendants(purposeID, tenantID uuid.UUID) ([]models.Purpose, error) {
	var descendants []models.Purpose

	// Get immediate children
	children, err := r.GetChildPurposes(purposeID, tenantID)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		descendants = append(descendants, child)
		// Recursively get grandchildren
		grandChildren, err := r.GetPurposeDescendants(child.ID, tenantID)
		if err != nil {
			return nil, err
		}
		descendants = append(descendants, grandChildren...)
	}

	return descendants, nil
}
