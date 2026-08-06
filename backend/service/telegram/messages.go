package telegram

import (
	"fmt"
	"strings"

	"github.com/lucas-remigio/wallet-tracker/types"
)

// Every user-facing string the bot can send lives here, so wording (and, later,
// language) is changed in one place instead of being scattered through the
// conversation logic.
const (
	msgNotLinked = "This chat is not linked to a Grão Certo account yet.\n" +
		"Open the app, go to Settings → Telegram, generate a code and send it here as /link YOURCODE."
	msgLinkUsage     = "Send the code from the app like this: /link YOURCODE"
	msgLinkInvalid   = "That code is invalid or has expired. Generate a new one in Settings → Telegram."
	msgLinkFailed    = "I couldn't link this chat. If it is already linked to another account, unlink it there first."
	msgCancelled     = "Discarded. Nothing was saved."
	msgNothingToDo   = "There is nothing pending to cancel."
	msgNoAccounts    = "You don't have any account yet. Create one in the app first."
	msgNoCategories  = "You don't have any category yet. Create one in the app first."
	msgNotUnderstood = "I couldn't find a transaction in that message.\n" +
		"Try something like: 3.19 cookies, 4.50 fuel"
	msgGenericError       = "Something went wrong on my side. Please try again in a moment."
	msgNothingSaved       = "I couldn't save that. Nothing was written — please try again, or add it in the app."
	msgPickCategoryNumber = "I need the number of one of the categories in the list."
	msgPickAccountNumber  = "I need the number of one of the accounts in the list."
	msgConfirmOrCancel    = `Reply "yes" to save, or "cancel" to discard.`
	msgStart              = "Hi! I turn messages like \"3.19 cookies, 4.50 fuel\" into transactions in Grão Certo.\n" +
		"To get started, generate a code in the app under Settings → Telegram and send it here as /link YOURCODE."
)

func msgLinked(firstName string) string {
	return fmt.Sprintf("Linked to %s's account. Send me a transaction whenever you like, for example: 3.19 cookies", firstName)
}

// renderQuestion turns the outstanding slot into the message the user sees.
// The confirmation "question" is the full summary, so the user always sees the
// account and every category before anything is written.
func renderQuestion(p *types.PendingParse, catalog *Catalog) string {
	current, index := nextSlot(p)

	switch current {
	case slotCategory:
		return renderCategoryQuestion(p.Items[index], catalog)
	case slotAccount:
		return renderAccountQuestion(catalog)
	default:
		return renderSummary(p, catalog)
	}
}

func renderCategoryQuestion(item types.PendingItem, catalog *Catalog) string {
	var message strings.Builder

	description := item.Description
	if description == "" {
		description = formatMoney(item.Amount)
	}

	fmt.Fprintf(&message, "I couldn't match a category for %q. Pick one:\n", description)
	for i, category := range catalog.Categories {
		fmt.Fprintf(&message, "%d. %s\n", i+1, catalog.CategoryLabel(category))
	}
	message.WriteString("\nReply with the number.")

	return message.String()
}

func renderAccountQuestion(catalog *Catalog) string {
	var message strings.Builder
	def := catalog.DefaultAccount()

	message.WriteString("Which account?\n")
	for i, account := range catalog.Accounts {
		fmt.Fprintf(&message, "%d. %s", i+1, account.AccountName)
		if def != nil && account.Token == def.Token {
			message.WriteString(" (default)")
		}
		message.WriteString("\n")
	}

	if def != nil {
		message.WriteString(`
Reply with the number (or just "yes" for the default).`)
	} else {
		message.WriteString("\nReply with the number.")
	}

	return message.String()
}

func renderSummary(p *types.PendingParse, catalog *Catalog) string {
	var message strings.Builder

	fmt.Fprintf(&message, "I understood %s on %q:\n", pluralTransactions(len(p.Items)), accountName(p, catalog))
	for i, item := range p.Items {
		fmt.Fprintf(&message, "%d. %s — %s (%s)\n", i+1, formatMoney(item.Amount), item.Description, categoryAndType(item, catalog))
	}
	message.WriteString("\n" + msgConfirmOrCancel)

	return message.String()
}

// renderCreated reports what was actually written. created may be shorter than
// the items when a later insert failed, and the message says so rather than
// implying everything landed.
func renderCreated(p *types.PendingParse, catalog *Catalog, created []*types.TransactionChangeResponse) string {
	var message strings.Builder

	fmt.Fprintf(&message, "✅ Added %s to %q:\n", pluralTransactions(len(created)), accountName(p, catalog))
	for i := range created {
		item := p.Items[i]
		fmt.Fprintf(&message, "- %s %s (%s)\n", formatMoney(item.Amount), item.Description, categoryName(item, catalog))
	}

	if balance := latestBalance(created); balance != nil {
		fmt.Fprintf(&message, "New balance: %s", formatMoney(*balance))
	}

	if len(created) < len(p.Items) {
		fmt.Fprintf(&message, "\n\n⚠️ %d of %d could not be saved. Please add %s in the app.",
			len(p.Items)-len(created), len(p.Items), pluralThem(len(p.Items)-len(created)))
	}

	return message.String()
}

func latestBalance(created []*types.TransactionChangeResponse) *float64 {
	for i := len(created) - 1; i >= 0; i-- {
		if created[i] != nil && created[i].AccountBalance != nil {
			return created[i].AccountBalance
		}
	}
	return nil
}

func accountName(p *types.PendingParse, catalog *Catalog) string {
	if p.AccountToken == nil {
		return ""
	}

	if account := catalog.Account(*p.AccountToken); account != nil {
		return account.AccountName
	}

	return ""
}

func categoryName(item types.PendingItem, catalog *Catalog) string {
	if item.CategoryID == nil {
		return ""
	}

	return catalog.CategoryLabel(catalog.Category(*item.CategoryID))
}

// categoryAndType shows the derived transaction type next to the category, so
// the user can catch a debit that should have been a credit before confirming.
func categoryAndType(item types.PendingItem, catalog *Catalog) string {
	if item.CategoryID == nil {
		return ""
	}

	category := catalog.Category(*item.CategoryID)
	if category == nil {
		return ""
	}

	return fmt.Sprintf("%s · %s", catalog.CategoryLabel(category), category.TransactionType.TypeSlug)
}

func formatMoney(amount float64) string {
	return fmt.Sprintf("€%.2f", amount)
}

func pluralTransactions(count int) string {
	if count == 1 {
		return "1 transaction"
	}
	return fmt.Sprintf("%d transactions", count)
}

func pluralThem(count int) string {
	if count == 1 {
		return "it"
	}
	return "them"
}
