package types

type CategoryStore interface {
	CreateCategory(category *Category) (*Category, error)
	CreateCategoryAndReturn(category *Category) (*CategoryDTO, error)
	UpdateCategory(category *Category, userId int) (*Category, error)
	UpdateCategoryAndReturn(category *Category, userId int) (*CategoryDTO, error)
	GetCategoryById(id int, userId int) (*Category, error)
	GetCategoriesByUserId(userId int) ([]*Category, error)
	GetCategoryDtoById(id int, userId int) (*CategoryDTO, error)
	GetCategoriesDtoByUserId(userId int) ([]*CategoryDTO, error)
	DeleteCategory(id int, userId int) error
	SoftDeleteCategory(id int, userId int) error
}

type CreateCategoryPayload struct {
	TransactionTypeId int    `json:"transaction_type_id" validate:"required,numeric,min=1,max=3"`
	CategoryName      string `json:"category_name" validate:"required,max=255,min=3"`
	Color             string `json:"color" validate:"required,hexcolor"`
	Budget            *int   `json:"budget" validate:"omitempty,numeric,min=1,max=99999"`
}

type UpdateCategoryPayload struct {
	CategoryName string `json:"category_name" validate:"required,max=255,min=3"`
	Color        string `json:"color" validate:"required,hexcolor"`
	Budget       *int   `json:"budget" validate:"omitempty,numeric,min=1,max=99999"`
}

type Category struct {
	ID                int     `json:"id"`
	UserID            int     `json:"user_id"`
	TransactionTypeID int     `json:"transaction_type_id"`
	ParentCategoryID  *int    `json:"parent_category_id,omitempty"`
	CategoryName      string  `json:"category_name"`
	Color             string  `json:"color"`
	Budget            *int    `json:"budget"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	DeletedAt         *string `json:"deleted_at,omitempty"` // Nullable field for soft delete
}

type CategoryDTO struct {
	ID               int              `json:"id"`
	TransactionType  *TransactionType `json:"transaction_type"`
	ParentCategoryID *int             `json:"parent_category_id,omitempty"` // NEW - nullable
	ParentCategory   *CategoryDTO     `json:"parent_category,omitempty"`    // NEW - for display
	Subcategories    []*CategoryDTO   `json:"subcategories,omitempty"`      // NEW - for hierarchy
	CategoryName     string           `json:"category_name"`
	Color            string           `json:"color"`
	Budget           *int             `json:"budget"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
	DeletedAt        *string          `json:"deleted_at,omitempty"` // Nullable field for soft delete
}
