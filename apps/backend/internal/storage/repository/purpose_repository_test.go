package repository

import (
	"pixpivot/arc/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRepositoryTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// Auto-migrate the schema
	db.AutoMigrate(&models.Purpose{}, &models.ConsentFormPurpose{})

	return db
}

func TestGetChildPurposes_Success(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	parentID := uuid.New()

	// Create parent purpose
	parent := &models.Purpose{
		ID:       parentID,
		Name:     "Parent Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	// Create child purposes
	child1 := &models.Purpose{
		ID:              uuid.New(),
		Name:            "Child Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &parentID,
		Active:          true,
	}

	child2 := &models.Purpose{
		ID:              uuid.New(),
		Name:            "Child Purpose 2",
		TenantID:        tenantID,
		ParentPurposeID: &parentID,
		Active:          true,
	}

	// Create inactive child (should not be returned)
	inactiveChild := &models.Purpose{
		ID:              uuid.New(),
		Name:            "Inactive Child",
		TenantID:        tenantID,
		ParentPurposeID: &parentID,
		Active:          false,
	}

	err := db.Create(parent).Error
	assert.NoError(t, err)
	err = db.Create(child1).Error
	assert.NoError(t, err)
	err = db.Create(child2).Error
	assert.NoError(t, err)
	err = db.Create(inactiveChild).Error
	assert.NoError(t, err)

	// Test getting child purposes
	children, err := repo.GetChildPurposes(parentID, tenantID)
	assert.NoError(t, err)
	assert.Len(t, children, 2) // Should only return active children

	childNames := []string{children[0].Name, children[1].Name}
	assert.Contains(t, childNames, "Child Purpose 1")
	assert.Contains(t, childNames, "Child Purpose 2")
}

func TestGetPurposeTree_Success(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	rootID := uuid.New()
	child1ID := uuid.New()
	grandchild1ID := uuid.New()

	// Create hierarchy: Root -> Child1 -> Grandchild1
	root := &models.Purpose{
		ID:       rootID,
		Name:     "Root Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	child1 := &models.Purpose{
		ID:              child1ID,
		Name:            "Child Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &rootID,
		Active:          true,
	}

	grandchild1 := &models.Purpose{
		ID:              grandchild1ID,
		Name:            "Grandchild Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &child1ID,
		Active:          true,
	}

	err := db.Create(root).Error
	assert.NoError(t, err)
	err = db.Create(child1).Error
	assert.NoError(t, err)
	err = db.Create(grandchild1).Error
	assert.NoError(t, err)

	// Test getting purpose tree
	tree, err := repo.GetPurposeTree(rootID, tenantID)
	assert.NoError(t, err)
	assert.NotNil(t, tree)
	assert.Equal(t, "Root Purpose", tree.Purpose.Name)
	assert.Len(t, tree.Children, 1)
	assert.Equal(t, "Child Purpose 1", tree.Children[0].Purpose.Name)
	assert.Len(t, tree.Children[0].Children, 1)
	assert.Equal(t, "Grandchild Purpose 1", tree.Children[0].Children[0].Purpose.Name)
}

func TestGetPurposeTree_NotFound(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	// Test getting tree for non-existent purpose
	_, err := repo.GetPurposeTree(uuid.New(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestGetAllRootPurposes_Success(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	otherTenantID := uuid.New()

	// Create root purposes for target tenant
	root1 := &models.Purpose{
		ID:              uuid.New(),
		Name:            "Root Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: nil, // Root purpose
		Active:          true,
	}

	root2 := &models.Purpose{
		ID:              uuid.New(),
		Name:            "Root Purpose 2",
		TenantID:        tenantID,
		ParentPurposeID: nil, // Root purpose
		Active:          true,
	}

	// Create child purpose (should not be returned)
	child := &models.Purpose{
		ID:              uuid.New(),
		Name:            "Child Purpose",
		TenantID:        tenantID,
		ParentPurposeID: &root1.ID,
		Active:          true,
	}

	// Create root purpose for different tenant (should not be returned)
	otherRoot := &models.Purpose{
		ID:              uuid.New(),
		Name:            "Other Tenant Root",
		TenantID:        otherTenantID,
		ParentPurposeID: nil,
		Active:          true,
	}

	err := db.Create(root1).Error
	assert.NoError(t, err)
	err = db.Create(root2).Error
	assert.NoError(t, err)
	err = db.Create(child).Error
	assert.NoError(t, err)
	err = db.Create(otherRoot).Error
	assert.NoError(t, err)

	// Test getting root purposes
	roots, err := repo.GetAllRootPurposes(tenantID)
	assert.NoError(t, err)
	assert.Len(t, roots, 2)

	rootNames := []string{roots[0].Name, roots[1].Name}
	assert.Contains(t, rootNames, "Root Purpose 1")
	assert.Contains(t, rootNames, "Root Purpose 2")
}

func TestValidateHierarchy_Success(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	parentID := uuid.New()
	childID := uuid.New()

	// Create valid hierarchy
	parent := &models.Purpose{
		ID:       parentID,
		Name:     "Parent Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	child := &models.Purpose{
		ID:              childID,
		Name:            "Child Purpose",
		TenantID:        tenantID,
		ParentPurposeID: &parentID,
		Active:          true,
	}

	err := db.Create(parent).Error
	assert.NoError(t, err)
	err = db.Create(child).Error
	assert.NoError(t, err)

	// Test validating hierarchy
	err = repo.ValidateHierarchy(childID, &parentID)
	assert.NoError(t, err)
}

func TestValidateHierarchy_CircularReference(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	purpose1ID := uuid.New()
	purpose2ID := uuid.New()

	// Create circular reference: Purpose1 -> Purpose2 -> Purpose1
	purpose1 := &models.Purpose{
		ID:              purpose1ID,
		Name:            "Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &purpose2ID, // Points to purpose2
		Active:          true,
	}

	purpose2 := &models.Purpose{
		ID:              purpose2ID,
		Name:            "Purpose 2",
		TenantID:        tenantID,
		ParentPurposeID: &purpose1ID, // Points back to purpose1
		Active:          true,
	}

	err := db.Create(purpose1).Error
	assert.NoError(t, err)
	err = db.Create(purpose2).Error
	assert.NoError(t, err)

	// Test validating hierarchy with circular reference
	// We are trying to set purpose1's parent to purpose2 (which is already set, but we validate it)
	err = repo.ValidateHierarchy(purpose1ID, &purpose2ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular reference detected")
}

func TestSetParentPurpose_Success(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	parentID := uuid.New()
	childID := uuid.New()

	// Create purposes
	parent := &models.Purpose{
		ID:       parentID,
		Name:     "Parent Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	child := &models.Purpose{
		ID:       childID,
		Name:     "Child Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	err := db.Create(parent).Error
	assert.NoError(t, err)
	err = db.Create(child).Error
	assert.NoError(t, err)

	// Test setting parent purpose
	err = repo.SetParentPurpose(childID, parentID, tenantID)
	assert.NoError(t, err)

	// Verify the relationship was set
	var updatedChild models.Purpose
	err = db.First(&updatedChild, "id = ?", childID).Error
	assert.NoError(t, err)
	assert.NotNil(t, updatedChild.ParentPurposeID)
	assert.Equal(t, parentID, *updatedChild.ParentPurposeID)
}

func TestSetParentPurpose_ParentNotFound(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	childID := uuid.New()
	nonExistentParentID := uuid.New()

	// Create child purpose
	child := &models.Purpose{
		ID:       childID,
		Name:     "Child Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	err := db.Create(child).Error
	assert.NoError(t, err)

	// Test setting non-existent parent
	err = repo.SetParentPurpose(childID, nonExistentParentID, tenantID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent purpose not found")
}

func TestGetPurposeDepth_Success(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	rootID := uuid.New()
	child1ID := uuid.New()
	grandchild1ID := uuid.New()

	// Create hierarchy: Root (depth 0) -> Child1 (depth 1) -> Grandchild1 (depth 2)
	root := &models.Purpose{
		ID:       rootID,
		Name:     "Root Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	child1 := &models.Purpose{
		ID:              child1ID,
		Name:            "Child Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &rootID,
		Active:          true,
	}

	grandchild1 := &models.Purpose{
		ID:              grandchild1ID,
		Name:            "Grandchild Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &child1ID,
		Active:          true,
	}

	err := db.Create(root).Error
	assert.NoError(t, err)
	err = db.Create(child1).Error
	assert.NoError(t, err)
	err = db.Create(grandchild1).Error
	assert.NoError(t, err)

	// Test getting depths
	depth, err := repo.GetPurposeDepth(rootID)
	assert.NoError(t, err)
	assert.Equal(t, 0, depth)

	depth, err = repo.GetPurposeDepth(child1ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, depth)

	depth, err = repo.GetPurposeDepth(grandchild1ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, depth)
}

func TestGetPurposeAncestors_Success(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	rootID := uuid.New()
	child1ID := uuid.New()
	grandchild1ID := uuid.New()

	// Create hierarchy
	root := &models.Purpose{
		ID:       rootID,
		Name:     "Root Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	child1 := &models.Purpose{
		ID:              child1ID,
		Name:            "Child Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &rootID,
		Active:          true,
	}

	grandchild1 := &models.Purpose{
		ID:              grandchild1ID,
		Name:            "Grandchild Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &child1ID,
		Active:          true,
	}

	err := db.Create(root).Error
	assert.NoError(t, err)
	err = db.Create(child1).Error
	assert.NoError(t, err)
	err = db.Create(grandchild1).Error
	assert.NoError(t, err)

	// Test getting ancestors of grandchild
	ancestors, err := repo.GetPurposeAncestors(grandchild1ID, tenantID)
	assert.NoError(t, err)
	assert.Len(t, ancestors, 2) // Should have child1 and root as ancestors

	// Ancestors should be in order: immediate parent first, then grandparent, etc.
	assert.Equal(t, "Child Purpose 1", ancestors[0].Name)
	assert.Equal(t, "Root Purpose", ancestors[1].Name)

	// Test getting ancestors of root (should be empty)
	ancestors, err = repo.GetPurposeAncestors(rootID, tenantID)
	assert.NoError(t, err)
	assert.Len(t, ancestors, 0)
}

func TestGetPurposeDescendants_Success(t *testing.T) {
	db := setupRepositoryTestDB()
	repo := NewPurposeRepository(db)

	tenantID := uuid.New()
	rootID := uuid.New()
	child1ID := uuid.New()
	child2ID := uuid.New()
	grandchild1ID := uuid.New()

	// Create hierarchy
	root := &models.Purpose{
		ID:       rootID,
		Name:     "Root Purpose",
		TenantID: tenantID,
		Active:   true,
	}

	child1 := &models.Purpose{
		ID:              child1ID,
		Name:            "Child Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &rootID,
		Active:          true,
	}

	child2 := &models.Purpose{
		ID:              child2ID,
		Name:            "Child Purpose 2",
		TenantID:        tenantID,
		ParentPurposeID: &rootID,
		Active:          true,
	}

	grandchild1 := &models.Purpose{
		ID:              grandchild1ID,
		Name:            "Grandchild Purpose 1",
		TenantID:        tenantID,
		ParentPurposeID: &child1ID,
		Active:          true,
	}

	err := db.Create(root).Error
	assert.NoError(t, err)
	err = db.Create(child1).Error
	assert.NoError(t, err)
	err = db.Create(child2).Error
	assert.NoError(t, err)
	err = db.Create(grandchild1).Error
	assert.NoError(t, err)

	// Test getting descendants of root
	descendants, err := repo.GetPurposeDescendants(rootID, tenantID)
	assert.NoError(t, err)
	assert.Len(t, descendants, 3) // Should have child1, child2, and grandchild1

	descendantNames := make([]string, len(descendants))
	for i, desc := range descendants {
		descendantNames[i] = desc.Name
	}
	assert.Contains(t, descendantNames, "Child Purpose 1")
	assert.Contains(t, descendantNames, "Child Purpose 2")
	assert.Contains(t, descendantNames, "Grandchild Purpose 1")

	// Test getting descendants of leaf node (should be empty)
	descendants, err = repo.GetPurposeDescendants(grandchild1ID, tenantID)
	assert.NoError(t, err)
	assert.Len(t, descendants, 0)
}
