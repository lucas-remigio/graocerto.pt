package transaction_types

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

func (s *Store) GetTransactionTypes() ([]*types.TransactionType, error) {
	return db.QueryList(s.db,
		"SELECT id, type_name, type_slug FROM transaction_types",
		scanRowIntoTransactionType)
}

func scanRowIntoTransactionType(rows *sql.Rows) (*types.TransactionType, error) {
	tt := new(types.TransactionType)

	err := rows.Scan(&tt.ID, &tt.TypeName, &tt.TypeSlug)

	if err != nil {
		return nil, err
	}

	return tt, nil
}
