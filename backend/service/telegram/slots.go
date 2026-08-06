package telegram

import (
	"strconv"
	"strings"

	"github.com/lucas-remigio/wallet-tracker/types"
)

// slot is the question a conversation is currently waiting on. It is never
// stored: it is derived from the pending parse every time, so the data and the
// question can never drift apart.
type slot int

const (
	slotCategory slot = iota
	slotAccount
	slotConfirmation
)

// nextSlot reports what still needs answering, and for a category question
// which item it belongs to (-1 otherwise).
func nextSlot(p *types.PendingParse) (slot, int) {
	for i := range p.Items {
		if p.Items[i].CategoryID == nil {
			return slotCategory, i
		}
	}

	if p.AccountToken == nil || *p.AccountToken == "" {
		return slotAccount, -1
	}

	return slotConfirmation, -1
}

// applyAnswer interprets text as the answer to the outstanding question and
// fills the slot in place. It is pure: no IO, no clock, no persistence.
//
// It returns confirmed=true when the user just approved the summary, and a
// non-empty problem when the answer could not be used — in which case the parse
// is left untouched and the caller re-asks.
func applyAnswer(p *types.PendingParse, text string, catalog *Catalog) (confirmed bool, problem string) {
	current, index := nextSlot(p)

	switch current {
	case slotCategory:
		choice, ok := parseChoice(text, len(catalog.Categories))
		if !ok {
			return false, msgPickCategoryNumber
		}

		category := catalog.Categories[choice-1]
		categoryID := category.ID
		transactionTypeID := category.TransactionType.ID
		p.Items[index].CategoryID = &categoryID
		p.Items[index].TransactionTypeID = &transactionTypeID
		return false, ""

	case slotAccount:
		account, problem := resolveAccountAnswer(text, catalog)
		if problem != "" {
			return false, problem
		}

		p.AccountToken = &account.Token
		return false, ""

	default:
		if isAffirmative(text) {
			return true, ""
		}
		return false, msgConfirmOrCancel
	}
}

// resolveAccountAnswer accepts either a number from the list or a bare "yes",
// which selects the account marked as the default.
func resolveAccountAnswer(text string, catalog *Catalog) (*types.Account, string) {
	if strings.TrimSpace(text) == "" || isAffirmative(text) {
		if def := catalog.DefaultAccount(); def != nil {
			return def, ""
		}
		return nil, msgPickAccountNumber
	}

	choice, ok := parseChoice(text, len(catalog.Accounts))
	if !ok {
		return nil, msgPickAccountNumber
	}

	return catalog.Accounts[choice-1], ""
}

// parseChoice reads a 1-based selection, tolerating the punctuation people add
// when replying to a numbered list ("2.", "#2").
func parseChoice(text string, optionCount int) (int, bool) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.TrimPrefix(cleaned, "#")
	cleaned = strings.TrimRight(cleaned, ".)")

	choice, err := strconv.Atoi(strings.TrimSpace(cleaned))
	if err != nil || choice < 1 || choice > optionCount {
		return 0, false
	}

	return choice, true
}

func isAffirmative(text string) bool {
	switch normalizeWord(text) {
	case "yes", "y", "yeah", "ok", "sim", "s", "confirm", "confirmar":
		return true
	default:
		return false
	}
}

func isCancel(text string) bool {
	switch normalizeWord(text) {
	case "cancel", "/cancel", "cancelar", "no", "n", "nao", "não":
		return true
	default:
		return false
	}
}

func normalizeWord(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}
