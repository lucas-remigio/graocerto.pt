package category

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lucas-remigio/wallet-tracker/db"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetCategoriesByUserId(userId int) ([]*types.Category, error) {
	query := `SELECT id, user_id, transaction_type_id, parent_category_id, category_name, color, created_at, updated_at, deleted_at, budget, order_index
          FROM categories
          WHERE user_id = $1
          ORDER BY order_index ASC, id ASC`
	return db.QueryList(s.db, query, scanRowsIntoCategory, userId)
}

func (s *Store) GetCategoryById(id int, userId int) (*types.Category, error) {
	query := `SELECT id, user_id, transaction_type_id, parent_category_id, category_name, color, created_at, updated_at, deleted_at, budget, order_index
          FROM categories
          WHERE id = $1 AND user_id = $2`
	return db.QuerySingle(s.db, query, scanRowIntoCategory, id, userId)
}

func (s *Store) GetCategoriesDtoByUserId(userId int) ([]*types.CategoryDTO, error) {
	query := `SELECT c.id, c.parent_category_id, c.category_name, c.color, c.created_at, c.updated_at, c.deleted_at, c.budget, c.order_index,
                 tt.id, tt.type_name, tt.type_slug
          FROM categories c
          JOIN transaction_types tt ON c.transaction_type_id = tt.id
          WHERE c.user_id = $1 AND c.deleted_at IS NULL
          ORDER BY c.order_index ASC, c.id ASC`

	return db.QueryList(s.db, query, scanRowsIntoCategoryDto, userId)
}

func (s *Store) GetCategoryDtoById(id int, userId int) (*types.CategoryDTO, error) {
	query := `SELECT c.id, c.parent_category_id, c.category_name, c.color, c.created_at, c.updated_at, c.deleted_at, c.budget, c.order_index,
                 tt.id, tt.type_name, tt.type_slug
          FROM categories c
          JOIN transaction_types tt ON c.transaction_type_id = tt.id
          WHERE c.id = $1 AND c.user_id = $2 AND c.deleted_at IS NULL`
	return db.QuerySingle(s.db, query, scanRowIntoCategoryDto, id, userId)
}

func (s *Store) CreateCategory(category *types.Category) (*types.Category, error) {
	// Validate parent category if provided
	if category.ParentCategoryId != nil {
		if err := s.validateParentCategory(*category.ParentCategoryId, category.UserID, category.TransactionTypeId); err != nil {
			return nil, err
		}
	}

	var id int
	err := s.db.QueryRow(
		"INSERT INTO categories (user_id, transaction_type_id, parent_category_id, category_name, color, budget) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		category.UserID, category.TransactionTypeId, category.ParentCategoryId, category.CategoryName, category.Color, category.Budget).Scan(&id)
	if err != nil {
		return nil, err
	}

	// Return the created category with its ID
	category.ID = id

	return category, nil
}

func (s *Store) CreateCategoryAndReturn(category *types.Category) (*types.CategoryDTO, error) {
	updatedCategory, err := s.CreateCategory(category)
	if err != nil {
		return nil, err
	}

	// Fetch the full DTO for the created category
	dto, err := s.GetCategoryDtoById(updatedCategory.ID, updatedCategory.UserID)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (s *Store) UpdateCategory(editCategory *types.Category, userId int) (*types.Category, error) {
	// get current category to check if incoming user is the same
	currentCategory, err := s.GetCategoryById(editCategory.ID, userId)
	if err != nil {
		return nil, err
	}

	if err := db.ValidateOwnership(currentCategory.UserID, userId, "category"); err != nil {
		return nil, err
	}

	// Validate parent category if provided
	if editCategory.ParentCategoryId != nil {
		if err := s.validateParentCategory(*editCategory.ParentCategoryId, userId, currentCategory.TransactionTypeId); err != nil {
			return nil, err
		}

		// Prevent creating circular references
		if *editCategory.ParentCategoryId == editCategory.ID {
			return nil, fmt.Errorf("category cannot be its own parent")
		}

		// Prevent setting a subcategory as parent
		if err := s.preventCircularReference(editCategory.ID, *editCategory.ParentCategoryId); err != nil {
			return nil, err
		}
	}

	_, err = db.ExecWithValidation(s.db,
		"UPDATE categories SET parent_category_id = $1, category_name = $2, color = $3, budget = $4 WHERE id = $5",
		editCategory.ParentCategoryId, editCategory.CategoryName, editCategory.Color, editCategory.Budget, editCategory.ID)

	if err != nil {
		return nil, err
	}

	currentCategory.ParentCategoryId = editCategory.ParentCategoryId
	currentCategory.CategoryName = editCategory.CategoryName
	currentCategory.Color = editCategory.Color
	currentCategory.Budget = editCategory.Budget

	return currentCategory, nil
}

// Helper function to validate parent category
func (s *Store) validateParentCategory(parentId int, userId int, transactionTypeId int) error {
	parent, err := s.GetCategoryById(parentId, userId)
	if err != nil {
		return fmt.Errorf("parent category not found")
	}

	// Parent must belong to the same user
	if err := db.ValidateOwnership(parent.UserID, userId, "parent category"); err != nil {
		return err
	}

	// Parent must have the same transaction type
	if parent.TransactionTypeId != transactionTypeId {
		return fmt.Errorf("parent category must have the same transaction type")
	}

	// Parent cannot itself be a subcategory (only 1 level deep)
	if parent.ParentCategoryId != nil {
		return fmt.Errorf("cannot create subcategory of a subcategory (max 1 level)")
	}

	return nil
}

// Helper function to prevent circular references
func (s *Store) preventCircularReference(categoryId int, parentId int) error {
	// Check if the parent is actually a child of this category
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM categories WHERE id = $1 AND parent_category_id = $2`,
		parentId, categoryId,
	).Scan(&count)

	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("circular reference detected: parent is a subcategory of this category")
	}

	return nil
}

func (s *Store) UpdateCategoryAndReturn(editCategory *types.Category, userId int) (*types.CategoryDTO, error) {
	updatedCategory, err := s.UpdateCategory(editCategory, userId)
	if err != nil {
		return nil, err
	}

	// Fetch the full DTO for the updated category
	dto, err := s.GetCategoryDtoById(updatedCategory.ID, updatedCategory.UserID)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

func (s *Store) DeleteCategory(id int, userId int) error {
	// get current category to check if incoming user is the same
	currentCategory, err := s.GetCategoryById(id, userId)
	if err != nil {
		return err
	}

	if err := db.ValidateOwnership(userId, currentCategory.UserID, "category"); err != nil {
		return err
	}

	// Check if category has subcategories
	var hasSubcategories bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE parent_category_id = $1 AND deleted_at IS NULL)", id).Scan(&hasSubcategories)
	if err != nil {
		return err
	}

	if hasSubcategories {
		return fmt.Errorf("cannot delete category with subcategories")
	}

	// check if the category is used in any transactions
	var exists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM transactions WHERE category_id = $1)", id).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Soft delete if used in transactions
		_, err := s.db.Exec(
			`UPDATE categories SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
			id, userId,
		)
		return err
	}

	// Hard delete if not used
	_, err = db.ExecWithValidation(s.db, "DELETE FROM categories WHERE id = $1", id)

	return err
}

func (s *Store) SoftDeleteCategory(id int, userId int) error {
	_, err := s.db.Exec(
		`UPDATE categories SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userId,
	)
	return err
}

func (s *Store) ReorderCategories(userId int, categories []types.ReorderCategory) error {
	// Verify all IDs belong to the user
	ids := make([]any, len(categories))
	placeholders := make([]string, len(categories))
	for i, cat := range categories {
		ids[i] = cat.ID
		placeholders[i] = fmt.Sprintf("$%d", i+2) // $1 is userId
	}
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM categories WHERE user_id = $1 AND id IN (%s)",
		strings.Join(placeholders, ","),
	)
	args := append([]any{userId}, ids...)
	var count int
	if err := s.db.QueryRow(query, args...).Scan(&count); err != nil {
		return fmt.Errorf("failed to verify category ownership: %w", err)
	}
	if count != len(categories) {
		return fmt.Errorf("one or more categories do not belong to the user")
	}

	// Verify order indexes are unique within the submitted list
	orderIndexes := make(map[int]bool)
	for _, cat := range categories {
		if orderIndexes[cat.OrderIndex] {
			return fmt.Errorf("duplicate order_index found: %d", cat.OrderIndex)
		}
		orderIndexes[cat.OrderIndex] = true
	}

	// Bulk update
	for _, cat := range categories {
		_, err := db.ExecWithValidation(s.db,
			"UPDATE categories SET order_index = $1 WHERE id = $2 AND user_id = $3",
			cat.OrderIndex, cat.ID, userId)
		if err != nil {
			return err
		}
	}
	return nil
}

// Update scanner functions
func scanRowsIntoCategory(rows *sql.Rows) (*types.Category, error) {
	c := new(types.Category)
	err := rows.Scan(&c.ID, &c.UserID, &c.TransactionTypeId, &c.ParentCategoryId,
		&c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.Budget, &c.OrderIndex)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func scanRowsIntoCategoryDto(rows *sql.Rows) (*types.CategoryDTO, error) {
	c := new(types.CategoryDTO)
	c.TransactionType = &types.TransactionType{}

	err := rows.Scan(
		&c.ID, &c.ParentCategoryId, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.Budget, &c.OrderIndex,
		&c.TransactionType.ID, &c.TransactionType.TypeName, &c.TransactionType.TypeSlug)

	if err != nil {
		return nil, err
	}
	return c, nil
}

func scanRowIntoCategoryDto(row *sql.Row) (*types.CategoryDTO, error) {
	c := new(types.CategoryDTO)
	c.TransactionType = &types.TransactionType{}

	err := row.Scan(
		&c.ID, &c.ParentCategoryId, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.Budget, &c.OrderIndex,
		&c.TransactionType.ID, &c.TransactionType.TypeName, &c.TransactionType.TypeSlug)

	if err != nil {
		return nil, err
	}
	return c, nil
}

func scanRowIntoCategory(row *sql.Row) (*types.Category, error) {
	c := new(types.Category)
	err := row.Scan(&c.ID, &c.UserID, &c.TransactionTypeId, &c.ParentCategoryId,
		&c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.Budget, &c.OrderIndex)
	if err != nil {
		return nil, err
	}
	return c, nil
}
