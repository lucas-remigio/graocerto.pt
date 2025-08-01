package types

type CategoryStore interface {
	GetCategoriesByUserId(userId int) ([]*Category, error)
	CreateCategory(category *Category) error
	GetCategoryDtoByUserId(userId int) ([]*CategoryDTO, error)
	GetCategoryById(id int, userId int) (*Category, error)
	UpdateCategory(category *Category, userId int) error
	DeleteCategory(id int, userId int) error
	SoftDeleteCategory(id int, userId int) error
}

type CreateCategoryPayload struct {
	TransactionTypeId int    `json:"transaction_type_id" validate:"required,numeric,min=1,max=3"`
	CategoryName      string `json:"category_name" validate:"required,max=255,min=3"`
	Color             string `json:"color" validate:"required,hexcolor"`
}

type UpdateCategoryPayload struct {
	CategoryName string `json:"category_name" validate:"required,max=255,min=3"`
	Color        string `json:"color" validate:"required,hexcolor"`
}

type Category struct {
	ID                int     `json:"id"`
	UserID            int     `json:"user_id"`
	TransactionTypeID int     `json:"transaction_type_id"`
	CategoryName      string  `json:"category_name"`
	Color             string  `json:"color"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	DeletedAt         *string `json:"deleted_at,omitempty"` // Nullable field for soft delete
}

type CategoryDTO struct {
	ID              int              `json:"id"`
	TransactionType *TransactionType `json:"transaction_type"`
	CategoryName    string           `json:"category_name"`
	Color           string           `json:"color"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
	DeletedAt       *string          `json:"deleted_at,omitempty"` // Nullable field for soft delete
}
