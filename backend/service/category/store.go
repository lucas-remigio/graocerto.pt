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

func (s *Store) CreateCategory(category *types.Category) error {
	return db.ExecWithValidation(s.db,
		"INSERT INTO categories (user_id, transaction_type_id, category_name, color) VALUES (?, ?, ?, ?)",
		category.UserID, category.TransactionTypeID, category.CategoryName, category.Color)
}

func (s *Store) GetCategoriesByUserId(userId int) ([]*types.Category, error) {
	return db.QueryList(s.db,
		"SELECT id, user_id, transaction_type_id, category_name, color, created_at, updated_at FROM categories WHERE user_id = ?",
		scanRowsIntoCategory, userId)
}

func (s *Store) GetCategoryById(id int, userId int) (*types.Category, error) {
	return db.QuerySingle(s.db,
		"SELECT id, user_id, transaction_type_id, category_name, color, created_at, updated_at FROM categories WHERE id = ? AND user_id = ?",
		scanRowIntoCategory, id, userId)
}

func (s *Store) GetCategoryDtoByUserId(userId int) ([]*types.CategoryDTO, error) {
	return db.QueryList(s.db,
		"SELECT "+
			"c.id, c.category_name, c.color, c.created_at, c.updated_at, "+
			"tt.id, tt.type_name, tt.type_slug "+
			"FROM categories c "+
			"JOIN transaction_types tt ON c.transaction_type_id = tt.id "+
			"WHERE c.user_id = ? "+
			"ORDER BY c.created_at DESC",
		scanRowIntoCategoryDto, userId)
}

func (s *Store) UpdateCategory(category *types.Category, userId int) error {
	// get current category to check if incoming user is the same
	currentCategory, err := s.GetCategoryById(category.ID, userId)
	if err != nil {
		return err
	}

	if err := db.ValidateOwnership(currentCategory.UserID, userId, "category"); err != nil {
		return err
	}

	return db.ExecWithValidation(s.db,
		"UPDATE categories SET category_name = ?, color = ? WHERE id = ?",
		category.CategoryName, category.Color, category.ID)
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
	if err := db.CheckResourceExists(s.db, "SELECT id FROM transactions WHERE category_id = ?", "category", id); err != nil {
		return err
	}

	return db.ExecWithValidation(s.db, "DELETE FROM categories WHERE id = ?", id)
}

func scanRowsIntoCategory(rows *sql.Rows) (*types.Category, error) {
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

func scanRowIntoCategory(row *sql.Row) (*types.Category, error) {
	c := new(types.Category)

	err := row.Scan(&c.ID, &c.UserID, &c.TransactionTypeID, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}
