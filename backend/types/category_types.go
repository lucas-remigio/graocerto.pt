package types

type CategoryStore interface {
	GetCategoriesByUserId(userId int) ([]*Category, error)
	CreateCategory(category *Category) error
	GetCategoryDtoByUserId(userId int) ([]*CategoryDTO, error)
}

type CreateCategoryPayload struct {
	TransactionTypeId int    `json:"transaction_type_id" validate:"required,numeric,min=1"`
	CategoryName      string `json:"category_name" validate:"required,max=255,min=3"`
	Color             string `json:"color" validate:"required,hexcolor"`
}

type Category struct {
	ID                int    `json:"id"`
	UserID            int    `json:"user_id"`
	TransactionTypeID int    `json:"transaction_type_id"`
	CategoryName      string `json:"category_name"`
	Color             string `json:"color"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type CategoryDTO struct {
	ID              int              `json:"id"`
	TransactionType *TransactionType `json:"transaction_type"`
	CategoryName    string           `json:"category_name"`
	Color           string           `json:"color"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
}
