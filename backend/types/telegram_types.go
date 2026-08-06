package types

import "time"

// TelegramStore persists the in-flight slot-filling conversation of a chat.
// There is at most one pending parse per chat; a new parse replaces the old one.
type TelegramStore interface {
	UpsertPendingParse(p *PendingParse) error
	// GetPendingParse returns nil (no error) when there is no live parse for the chat.
	GetPendingParse(chatID string) (*PendingParse, error)
	DeletePendingParse(chatID string) error
	DeletePendingParsesByUserID(userID int) error
}

// TelegramLinkTokenResponse is what the settings UI shows the user to type
// into the bot.
type TelegramLinkTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TelegramStatusResponse struct {
	Linked bool `json:"linked"`
}

// PendingItem is one transaction being assembled. Nil pointers are unfilled
// slots that the conversation still has to ask about.
type PendingItem struct {
	Amount            float64 `json:"amount"`
	Description       string  `json:"description"`
	CategoryID        *int    `json:"category_id"`
	TransactionTypeID *int    `json:"transaction_type_id"`
}

// PendingParse is internal conversation state; it is never served over the API.
type PendingParse struct {
	ID           int
	UserID       int
	ChatID       string
	Items        []PendingItem
	AccountToken *string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}
