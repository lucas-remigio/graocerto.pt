package telegram

import (
	"fmt"

	"github.com/lucas-remigio/wallet-tracker/types"
)

// Catalog is everything a single incoming message may refer to: the categories
// a transaction can be filed under and the accounts it can land on. It is
// loaded once per message and shared by the parser (which offers it to the
// model) and the conversation (which renders it into numbered questions), so
// the two can never disagree about what the user is allowed to pick.
type Catalog struct {
	Categories []*types.CategoryDTO
	Accounts   []*types.Account

	categoriesByID    map[int]*types.CategoryDTO
	accountsByToken   map[string]*types.Account
	categoryNamesByID map[int]string
}

// LoadCatalog fetches the user's spendable categories and accounts.
func LoadCatalog(categories CategoryReader, accounts AccountReader, userID int) (*Catalog, error) {
	allCategories, err := categories.GetCategoriesDtoByUserId(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load categories: %w", err)
	}

	userAccounts, err := accounts.GetAccountsByUserId(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load accounts: %w", err)
	}

	return NewCatalog(allCategories, userAccounts), nil
}

// NewCatalog keeps only the categories a bot-created transaction may use.
// Transfer categories are excluded: a transfer needs a second account and
// TransactionStore.CreateTransfer, so booking one as a plain transaction would
// silently leave the other side of the movement missing.
func NewCatalog(categories []*types.CategoryDTO, accounts []*types.Account) *Catalog {
	catalog := &Catalog{
		Accounts:          accounts,
		categoriesByID:    map[int]*types.CategoryDTO{},
		accountsByToken:   map[string]*types.Account{},
		categoryNamesByID: map[int]string{},
	}

	for _, category := range categories {
		if !isSpendableCategory(category) {
			continue
		}
		catalog.Categories = append(catalog.Categories, category)
		catalog.categoriesByID[category.ID] = category
	}

	for _, account := range accounts {
		catalog.accountsByToken[account.Token] = account
	}

	// Names are resolved after filtering so a subcategory can still be labelled
	// with its parent even when the parent itself is not selectable.
	for _, category := range categories {
		catalog.categoryNamesByID[category.ID] = category.CategoryName
	}

	return catalog
}

// isSpendableCategory reports whether a category can back a single transaction.
func isSpendableCategory(category *types.CategoryDTO) bool {
	if category == nil || category.TransactionType == nil {
		return false
	}

	if category.DeletedAt != nil {
		return false
	}

	switch types.TransactionTypeID(category.TransactionType.ID) {
	case types.CreditTransactionType, types.DebitTransactionType:
		return true
	default:
		return false
	}
}

// Category returns the category with this id, or nil when the user does not own
// a selectable category with that id.
func (c *Catalog) Category(id int) *types.CategoryDTO {
	return c.categoriesByID[id]
}

// Account returns the account with this token, or nil when the user does not
// own it.
func (c *Catalog) Account(token string) *types.Account {
	return c.accountsByToken[token]
}

// CategoryLabel renders a category the way both the model and the user see it,
// qualifying subcategories with their parent ("Fuel (Car)").
func (c *Catalog) CategoryLabel(category *types.CategoryDTO) string {
	if category == nil {
		return ""
	}

	if category.ParentCategoryId != nil {
		if parent, ok := c.categoryNamesByID[*category.ParentCategoryId]; ok {
			return fmt.Sprintf("%s (%s)", category.CategoryName, parent)
		}
	}

	return category.CategoryName
}

// AutoAccount is the account to use without asking, because there is nothing to
// choose between: a single account, or a single favourite. Returns nil when the
// user genuinely has a choice to make.
func (c *Catalog) AutoAccount() *types.Account {
	if len(c.Accounts) == 1 {
		return c.Accounts[0]
	}

	favourites := c.favourites()
	if len(favourites) == 1 {
		return favourites[0]
	}

	return nil
}

// DefaultAccount is the one marked "(default)" when the account question is
// asked, i.e. the first favourite by order. Nil when no favourite exists, in
// which case the user must pick a number explicitly.
func (c *Catalog) DefaultAccount() *types.Account {
	favourites := c.favourites()
	if len(favourites) == 0 {
		return nil
	}

	return favourites[0]
}

// favourites preserves the store's order_index ordering.
func (c *Catalog) favourites() []*types.Account {
	var favourites []*types.Account
	for _, account := range c.Accounts {
		if account.IsFavorite {
			favourites = append(favourites, account)
		}
	}
	return favourites
}
