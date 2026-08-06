package telegram

import "github.com/lucas-remigio/wallet-tracker/types"

// The telegram package depends on narrow, locally declared interfaces rather
// than the full store interfaces in types/. The concrete stores satisfy these
// implicitly, so nothing changes at the wiring site, but this package stays
// unaffected by unrelated store changes and its tests need only tiny fakes.
// The pending-parse interface itself lives in types/ (types.TelegramStore),
// since this package owns and implements it.

// UserLinkStore is the slice of the user store the integration needs.
type UserLinkStore interface {
	GetUserById(id int) (*types.User, error)
	GetUserByTelegramChatID(chatID string) (*types.User, error)
	LinkTelegramChatID(userID int, chatID string) error
	UnlinkTelegram(userID int) error
}

// LinkTokenStore is the slice of the auth-token store used for link codes.
type LinkTokenStore interface {
	CreateAuthToken(token *types.AuthToken) error
	GetAuthTokenByPurposeAndSecret(purpose types.AuthTokenPurpose, secret string) (*types.AuthToken, error)
	ConsumeAuthToken(id string) error
	DeleteAuthTokensByUserAndPurpose(userID int, purpose types.AuthTokenPurpose) error
}

// CategoryReader supplies the categories a parsed transaction may use.
type CategoryReader interface {
	GetCategoriesDtoByUserId(userId int) ([]*types.CategoryDTO, error)
}

// AccountReader supplies the accounts a parsed transaction may target.
type AccountReader interface {
	GetAccountsByUserId(userId int) ([]*types.Account, error)
}

// TransactionWriter is the only mutation the integration performs.
type TransactionWriter interface {
	CreateTransactionAndReturn(transaction *types.Transaction, userId int) (*types.TransactionChangeResponse, error)
}

// ChatLimiter throttles how often one chat can make the bot work. The resource
// being protected is LLM spend, not server load, so it is applied per chat
// rather than per IP (every update arrives from Telegram's own address).
type ChatLimiter interface {
	Allow(chatID string) bool
}

// ChatLimiterFunc adapts a plain function, so cmd/api can hand over the
// existing ClientRateLimiter without this package importing it.
type ChatLimiterFunc func(chatID string) bool

func (f ChatLimiterFunc) Allow(chatID string) bool { return f(chatID) }
