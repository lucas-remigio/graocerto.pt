package telegram

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lucas-remigio/wallet-tracker/prompts"
	"github.com/lucas-remigio/wallet-tracker/types"
)

const (
	// maxAmount mirrors CreateTransactionPayload's validation, so nothing the
	// model returns can reach the store and fail there.
	maxAmount = 999999999
	// maxDescriptionLength mirrors the same payload's max=255.
	maxDescriptionLength = 255
	// maxItemsPerMessage bounds how much a single message can book. The model
	// is not trusted to be sensible about a pathological input.
	maxItemsPerMessage = 20
)

// llmResponse is the object wrapper the model is asked for. A wrapper rather
// than a bare array so the existing OpenAI client's `{`...`}` extraction keeps
// working untouched.
type llmResponse struct {
	Transactions []llmTransaction `json:"transactions"`
	AccountToken *string          `json:"account_token"`
}

type llmTransaction struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	CategoryID  *int    `json:"category_id"`
}

// Parser turns free text into a transaction proposal. It only ever proposes:
// every field it returns has been checked against the catalog, and nothing is
// written until the user confirms.
type Parser struct {
	llm types.OpenAIStore
}

func NewParser(llm types.OpenAIStore) *Parser {
	return &Parser{llm: llm}
}

// Parse returns the items it could extract plus the account the user named, if
// any. Unresolvable slots come back nil for the conversation to ask about; an
// empty slice means the message contained nothing bookable.
func (p *Parser) Parse(text string, catalog *Catalog) ([]types.PendingItem, *string, error) {
	raw, err := p.llm.GenerateGPT4Response(buildPrompt(text, catalog))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse message with llm: %w", err)
	}

	var response llmResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return nil, nil, fmt.Errorf("failed to decode llm response: %w", err)
	}

	return validateItems(response.Transactions, catalog), validateAccountToken(response.AccountToken, catalog), nil
}

// buildPrompt appends the user's allowed values and the message to the static
// template. The catalog is the model's entire universe of valid ids.
func buildPrompt(text string, catalog *Catalog) string {
	var prompt strings.Builder

	prompt.WriteString(prompts.TelegramTransactions)

	prompt.WriteString("\nALLOWED CATEGORIES:\n")
	if len(catalog.Categories) == 0 {
		prompt.WriteString("(none)\n")
	}
	for _, category := range catalog.Categories {
		prompt.WriteString(fmt.Sprintf("- id=%d | name=%s | type=%s\n",
			category.ID,
			catalog.CategoryLabel(category),
			category.TransactionType.TypeSlug,
		))
	}

	prompt.WriteString("\nALLOWED ACCOUNTS:\n")
	if len(catalog.Accounts) == 0 {
		prompt.WriteString("(none)\n")
	}
	for _, account := range catalog.Accounts {
		prompt.WriteString(fmt.Sprintf("- token=%s | name=%s\n", account.Token, account.AccountName))
	}

	prompt.WriteString("\nMESSAGE:\n")
	prompt.WriteString(text)

	return prompt.String()
}

// validateItems re-derives every field from the catalog. The model's output is
// treated as a suggestion: an id the user does not own, or one that is not
// bookable as a single transaction, is downgraded to "unknown" so the
// conversation asks instead of guessing.
func validateItems(transactions []llmTransaction, catalog *Catalog) []types.PendingItem {
	items := []types.PendingItem{}

	for _, transaction := range transactions {
		if len(items) == maxItemsPerMessage {
			slog.Warn("telegram parse truncated", "limit", maxItemsPerMessage)
			break
		}

		if transaction.Amount <= 0 || transaction.Amount > maxAmount {
			slog.Warn("telegram parse dropped item with invalid amount", "amount", transaction.Amount)
			continue
		}

		item := types.PendingItem{
			Amount:      transaction.Amount,
			Description: truncate(strings.TrimSpace(transaction.Description), maxDescriptionLength),
		}

		if category := resolveCategory(transaction.CategoryID, catalog); category != nil {
			categoryID := category.ID
			transactionTypeID := category.TransactionType.ID
			item.CategoryID = &categoryID
			// The type always comes from the resolved category, never from the model.
			item.TransactionTypeID = &transactionTypeID
		}

		items = append(items, item)
	}

	return items
}

func resolveCategory(categoryID *int, catalog *Catalog) *types.CategoryDTO {
	if categoryID == nil {
		return nil
	}

	category := catalog.Category(*categoryID)
	if category == nil {
		slog.Warn("telegram parse rejected unusable category", "category_id", *categoryID)
	}

	return category
}

func validateAccountToken(token *string, catalog *Catalog) *string {
	if token == nil || *token == "" {
		return nil
	}

	account := catalog.Account(*token)
	if account == nil {
		slog.Warn("telegram parse rejected unknown account token")
		return nil
	}

	return &account.Token
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit])
}
