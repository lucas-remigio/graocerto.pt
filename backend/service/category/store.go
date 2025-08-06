package category

import (
	"database/sql"

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
	query := `SELECT id, user_id, transaction_type_id, category_name, color, created_at, updated_at, deleted_at
		  FROM categories
		  WHERE user_id = $1 AND deleted_at IS NULL
		  ORDER BY created_at DESC`
	return db.QueryList(s.db, query, scanRowsIntoCategory, userId)
}

func (s *Store) GetCategoryById(id int, userId int) (*types.Category, error) {
	query := `SELECT id, user_id, transaction_type_id, category_name, color, created_at, updated_at, deleted_at
		  FROM categories
		  WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`
	return db.QuerySingle(s.db, query, scanRowIntoCategory, id, userId)
}

func (s *Store) GetCategoriesDtoByUserId(userId int) ([]*types.CategoryDTO, error) {
	query := `SELECT c.id, c.category_name, c.color, c.created_at, c.updated_at, c.deleted_at,
				 tt.id, tt.type_name, tt.type_slug
		  FROM categories c
		  JOIN transaction_types tt ON c.transaction_type_id = tt.id
		  WHERE c.user_id = $1 AND c.deleted_at IS NULL
		  ORDER BY c.created_at DESC`
	return db.QueryList(s.db,
		query,
		scanRowsIntoCategoryDto, userId)
}

func (s *Store) GetCategoryDtoById(id int, userId int) (*types.CategoryDTO, error) {
	query := `SELECT c.id, c.category_name, c.color, c.created_at, c.updated_at, c.deleted_at,
				 tt.id, tt.type_name, tt.type_slug
		  FROM categories c
		  JOIN transaction_types tt ON c.transaction_type_id = tt.id
		  WHERE c.id = $1 AND c.user_id = $2 AND c.deleted_at IS NULL`
	return db.QuerySingle(s.db, query, scanRowIntoCategoryDto, id, userId)
}

func (s *Store) CreateCategory(category *types.Category) (*types.Category, error) {
	var id int
	err := s.db.QueryRow(
		"INSERT INTO categories (user_id, transaction_type_id, category_name, color) VALUES ($1, $2, $3, $4) RETURNING id",
		category.UserID, category.TransactionTypeID, category.CategoryName, category.Color).Scan(&id)
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

	_, err = db.ExecWithValidation(s.db,
		"UPDATE categories SET category_name = $1, color = $2 WHERE id = $3",
		editCategory.CategoryName, editCategory.Color, editCategory.ID)

	if err != nil {
		return nil, err
	}

	currentCategory.CategoryName = editCategory.CategoryName
	currentCategory.Color = editCategory.Color

	return currentCategory, nil
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

func scanRowsIntoCategory(rows *sql.Rows) (*types.Category, error) {
	c := new(types.Category)

	err := rows.Scan(&c.ID, &c.UserID, &c.TransactionTypeID, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func scanRowsIntoCategoryDto(rows *sql.Rows) (*types.CategoryDTO, error) {
	c := new(types.CategoryDTO)

	// Initialize nested structs so they're not nil
	c.TransactionType = &types.TransactionType{}

	err := rows.Scan(
		&c.ID, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		&c.TransactionType.ID, &c.TransactionType.TypeName, &c.TransactionType.TypeSlug)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func scanRowIntoCategoryDto(row *sql.Row) (*types.CategoryDTO, error) {
	c := new(types.CategoryDTO)

	// Initialize nested structs so they're not nil
	c.TransactionType = &types.TransactionType{}

	err := row.Scan(
		&c.ID, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		&c.TransactionType.ID, &c.TransactionType.TypeName, &c.TransactionType.TypeSlug)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func scanRowIntoCategory(row *sql.Row) (*types.Category, error) {
	c := new(types.Category)

	err := row.Scan(&c.ID, &c.UserID, &c.TransactionTypeID, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}
