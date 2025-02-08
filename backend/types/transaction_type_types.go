package types

type TransactionTypesStore interface {
	GetTransactionTypes() ([]*TransactionType, error)
}

type TransactionType struct {
	ID       int    `json:"id"`
	TypeName string `json:"type_name"`
	TypeSlug string `json:"type_slug"`
}
