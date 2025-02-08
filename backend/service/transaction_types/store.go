package transaction_types

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

func (s *Store) GetTransactionTypes() ([]*types.TransactionType, error) {
	rows, err := s.db.Query("SELECT id, type_name, type_slug FROM transaction_types")
	if err != nil {
		return nil, err
	}

	transactionTypes := make([]*types.TransactionType, 0)
	for rows.Next() {
		transactionType, err := scanRowIntoTransactionType(rows)
		if err != nil {
			return nil, err
		}

		transactionTypes = append(transactionTypes, transactionType)
	}

	return transactionTypes, nil
}

func scanRowIntoTransactionType(rows *sql.Rows) (*types.TransactionType, error) {
	tt := new(types.TransactionType)

	err := rows.Scan(&tt.ID, &tt.TypeName, &tt.TypeSlug)

	if err != nil {
		return nil, err
	}

	return tt, nil
}
