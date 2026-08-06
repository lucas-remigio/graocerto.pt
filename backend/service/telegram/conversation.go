package telegram

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/lucas-remigio/wallet-tracker/types"
)

// transactionDateLayout matches how the rest of the backend writes transaction
// dates (see recurring_rule.Store).
const transactionDateLayout = "2006-01-02"

// ServiceDeps are the collaborators the conversation needs. A struct rather
// than seven positional arguments, so the wiring in cmd/api reads clearly.
type ServiceDeps struct {
	Users        UserLinkStore
	Pending      types.TelegramStore
	Categories   CategoryReader
	Accounts     AccountReader
	Transactions TransactionWriter
	LLM          types.OpenAIStore
	Links        *LinkService
}

// Service is the whole bot brain: it owns the conversation, and it is the only
// thing that writes. The Telegram runtime does nothing but hand it text and
// send back the reply it returns.
type Service struct {
	users        UserLinkStore
	pending      types.TelegramStore
	categories   CategoryReader
	accounts     AccountReader
	transactions TransactionWriter
	parser       *Parser
	links        *LinkService
	// now is a field so tests can pin the transaction date.
	now func() time.Time
}

func NewService(deps ServiceDeps) *Service {
	return &Service{
		users:        deps.Users,
		pending:      deps.Pending,
		categories:   deps.Categories,
		accounts:     deps.Accounts,
		transactions: deps.Transactions,
		parser:       NewParser(deps.LLM),
		links:        deps.Links,
		now:          time.Now,
	}
}

// HandleLink redeems a code typed into the chat. Like HandleMessage it always
// returns a sendable reply; a non-nil error is for the log, not the user.
func (s *Service) HandleLink(chatID, code string) (string, error) {
	if strings.TrimSpace(code) == "" {
		return msgLinkUsage, nil
	}

	user, err := s.links.Redeem(chatID, code)
	if err != nil {
		if errors.Is(err, ErrInvalidLinkCode) {
			return msgLinkInvalid, nil
		}
		return msgLinkFailed, err
	}

	return msgLinked(user.FirstName), nil
}

// HandleMessage is the single entry point for every non-/link message. The
// reply is always safe to send as-is; the error is only for logging.
func (s *Service) HandleMessage(chatID, text string) (string, error) {
	text = strings.TrimSpace(text)

	user, err := s.users.GetUserByTelegramChatID(chatID)
	if err != nil || user == nil {
		// An unlinked chat is an ordinary situation, not a failure.
		return msgNotLinked, nil
	}

	pending, err := s.pending.GetPendingParse(chatID)
	if err != nil {
		return msgGenericError, err
	}

	if isCancel(text) {
		return s.cancel(chatID, pending)
	}

	catalog, err := LoadCatalog(s.categories, s.accounts, user.ID)
	if err != nil {
		return msgGenericError, err
	}

	if pending == nil {
		return s.start(chatID, user, text, catalog)
	}

	return s.advance(pending, user, text, catalog)
}

func (s *Service) cancel(chatID string, pending *types.PendingParse) (string, error) {
	if pending == nil {
		return msgNothingToDo, nil
	}

	if err := s.pending.DeletePendingParse(chatID); err != nil {
		return msgGenericError, err
	}

	return msgCancelled, nil
}

// start treats the message as new transaction input.
func (s *Service) start(chatID string, user *types.User, text string, catalog *Catalog) (string, error) {
	if len(catalog.Accounts) == 0 {
		return msgNoAccounts, nil
	}

	if len(catalog.Categories) == 0 {
		return msgNoCategories, nil
	}

	items, accountToken, err := s.parser.Parse(text, catalog)
	if err != nil {
		return msgGenericError, err
	}

	if len(items) == 0 {
		return msgNotUnderstood, nil
	}

	// Only fall back to an assumed account when there is nothing to choose
	// between; the summary names it either way before anything is written.
	if accountToken == nil {
		if auto := catalog.AutoAccount(); auto != nil {
			accountToken = &auto.Token
		}
	}

	pending := &types.PendingParse{
		UserID:       user.ID,
		ChatID:       chatID,
		Items:        items,
		AccountToken: accountToken,
	}

	if err := s.pending.UpsertPendingParse(pending); err != nil {
		return msgGenericError, err
	}

	return renderQuestion(pending, catalog), nil
}

// advance applies the answer to the outstanding question.
func (s *Service) advance(pending *types.PendingParse, user *types.User, text string, catalog *Catalog) (string, error) {
	confirmed, problem := applyAnswer(pending, text, catalog)
	if problem != "" {
		// The parse is untouched, so re-asking is enough.
		return problem + "\n\n" + renderQuestion(pending, catalog), nil
	}

	if !confirmed {
		if err := s.pending.UpsertPendingParse(pending); err != nil {
			return msgGenericError, err
		}
		return renderQuestion(pending, catalog), nil
	}

	return s.commit(pending, user, catalog)
}

// commit writes the transactions. The pending row is dropped first: whatever
// happens next, a second "yes" must never book the same items twice.
func (s *Service) commit(pending *types.PendingParse, user *types.User, catalog *Catalog) (string, error) {
	if pending.AccountToken == nil {
		return msgGenericError, errors.New("cannot commit a parse without an account")
	}

	if err := s.pending.DeletePendingParse(pending.ChatID); err != nil {
		return msgGenericError, err
	}

	date := s.now().Format(transactionDateLayout)
	created := make([]*types.TransactionChangeResponse, 0, len(pending.Items))

	for _, item := range pending.Items {
		if item.CategoryID == nil {
			return msgGenericError, errors.New("cannot commit an item without a category")
		}

		response, err := s.transactions.CreateTransactionAndReturn(&types.Transaction{
			AccountToken: *pending.AccountToken,
			CategoryId:   *item.CategoryID,
			Amount:       item.Amount,
			Description:  item.Description,
			Date:         date,
		}, user.ID)
		if err != nil {
			// Stop at the first failure and report honestly: the user is told
			// exactly how many landed, so nothing is silently lost.
			slog.Error("telegram failed to create transaction", "userID", user.ID, "error", err)
			if len(created) == 0 {
				return msgNothingSaved, err
			}
			return renderCreated(pending, catalog, created), err
		}

		created = append(created, response)
	}

	return renderCreated(pending, catalog, created), nil
}
