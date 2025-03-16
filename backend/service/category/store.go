package category

import (
	"database/sql"
	"fmt"

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

func (s *Store) CreateCategory(category *types.Category) error {
	_, err := s.db.Exec("INSERT INTO categories (user_id, transaction_type_id, category_name, color) VALUES (?, ?, ?, ?)",
		category.UserID, category.TransactionTypeID, category.CategoryName, category.Color)

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetCategoriesByUserId(userId int) ([]*types.Category, error) {
	rows, err := s.db.Query("SELECT id, user_id, transaction_type_id, category_name, color, created_at, updated_at FROM categories WHERE user_id = ?", userId)
	if err != nil {
		return nil, err
	}

	categories := make([]*types.Category, 0)
	for rows.Next() {
		category, err := scanRowIntoCategory(rows)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (s *Store) GetCategoryById(id int) (*types.Category, error) {
	row := s.db.QueryRow("SELECT id, user_id, transaction_type_id, category_name, color, created_at, updated_at FROM categories WHERE id = ?", id)

	category := new(types.Category)
	if err := row.Scan(&category.ID, &category.UserID, &category.TransactionTypeID, &category.CategoryName, &category.Color, &category.CreatedAt, &category.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return category, nil
}

func (s *Store) GetCategoryDtoByUserId(userId int) ([]*types.CategoryDTO, error) {
	rows, err := s.db.Query("SELECT "+
		"c.id, c.category_name, c.color, c.created_at, c.updated_at, "+
		"tt.id, tt.type_name, tt.type_slug "+
		"FROM categories c "+
		"JOIN transaction_types tt ON c.transaction_type_id = tt.id "+
		"WHERE c.user_id = ? "+
		"ORDER BY c.created_at DESC", userId)

	if err != nil {
		return nil, err
	}

	categories := make([]*types.CategoryDTO, 0)
	for rows.Next() {
		category, err := scanRowIntoCategoryDto(rows)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (s *Store) UpdateCategory(category *types.Category) error {
	// get current category to check if incomding user is the same
	currentCategory, err := s.GetCategoryById(category.ID)
	if err != nil {
		return err
	}

	if currentCategory.UserID != category.UserID {
		return fmt.Errorf("user does not have permission to update this category")
	}

	_, err = s.db.Exec("UPDATE categories SET category_name = ?, color = ? WHERE id = ?", category.CategoryName, category.Color, category.ID)
	if err != nil {
		return err
	}

	return nil
}

func scanRowIntoCategory(rows *sql.Rows) (*types.Category, error) {
	c := new(types.Category)

	err := rows.Scan(&c.ID, &c.UserID, &c.TransactionTypeID, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func scanRowIntoCategoryDto(rows *sql.Rows) (*types.CategoryDTO, error) {
	c := new(types.CategoryDTO)

	// Initialize nested structs so they're not nil
	c.TransactionType = &types.TransactionType{}

	err := rows.Scan(
		&c.ID, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt,
		&c.TransactionType.ID, &c.TransactionType.TypeName, &c.TransactionType.TypeSlug)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Store) DeleteCategory(id int, userId int) error {
	// get current category to check if incomding user is the same
	currentCategory, err := s.GetCategoryById(id)
	if err != nil {
		return err
	}

	if currentCategory.UserID != userId {
		return fmt.Errorf("user does not have permission to delete this category")
	}

	// first we must check if the category is used in any transactions
	rows, err := s.db.Query("SELECT id FROM transactions WHERE category_id = ?", id)
	if err != nil {
		return err
	}

	if rows.Next() {
		// TODO: in the future, in this case, we soft delete it
		// for now, we just return an error
		return fmt.Errorf("category is used in at least one transaction")
	}

	_, err = s.db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
