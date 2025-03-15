package types

type TransactionTypesStore interface {
	GetTransactionTypes() ([]*TransactionType, error)
}

type TransactionTypeID int

const (
	CreditTransactionType   TransactionTypeID = 1
	DebitTransactionType    TransactionTypeID = 2
	TransferTransactionType TransactionTypeID = 3
)

type TransactionType struct {
	ID       int    `json:"id"`
	TypeName string `json:"type_name"`
	TypeSlug string `json:"type_slug"`
}
