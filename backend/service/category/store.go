package category

import (
	"database/sql"

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

func scanRowIntoCategory(rows *sql.Rows) (*types.Category, error) {
	c := new(types.Category)

	err := rows.Scan(&c.ID, &c.UserID, &c.TransactionTypeID, &c.CategoryName, &c.Color, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}
