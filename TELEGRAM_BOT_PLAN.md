# Telegram Bot Integration Plan for Grão Certo (wallet-tracker)

## Overview
Build a Telegram bot that lets a linked user send natural-language text (e.g. `3.19 in cookies, 4.50 gasoil`), have an LLM parse it into structured transactions, then **fill any missing details (category, account) through a short numbered-reply conversation** and confirm before persisting. If the user's message already contains everything — or there's nothing to choose from — the bot skips straight to a single confirmation.

This plan is aligned with the existing Go backend:
- LLM access already exists via `service/openai` (`OpenAIStore.GenerateGPT4Response`, gpt-4.1-mini, temp 0) — we **reuse it**, we do not add a new provider.
- One-time tokens already exist via the `auth_tokens` table + `types.AuthTokenStore` (`types/auth_types.go`) — the Telegram link code **reuses it** with a new purpose, no new columns.
- Transactions are created through the existing `TransactionStore.CreateTransactionAndReturn` and require `account_token`, `category_id`, `amount`, `description`, `date`.
- Categories belong to a `transaction_type` (credit/debit/**transfer**) and are per-user (`GetCategoriesDtoByUserId`).
- **The bot runs in-process** as a goroutine in the API binary (same pattern as `runRecurringRuleScheduler` in `cmd/api/api.go`). There is no bot↔backend HTTP hop, therefore no machine-to-machine auth to build.

## Conversation model (slot-filling — the core UX)
The bot is **stateful per chat** and runs a **slot-filling** flow. Each transaction needs three slots: `amount`, `description`, `category_id`. Shared per message: `account`. The backend resolves everything it can from the first message, then **asks only for what's genuinely ambiguous**, one slot at a time, and finishes with a single confirmation.

**State is derived, never stored.** A single pure function inspects the pending row and decides what to ask next:

```go
// nextQuestion returns what still needs answering for this pending parse.
// state ∈ {stateCategory, stateAccount, stateConfirmation}
func nextQuestion(p *PendingParse) (state slotState, itemIndex int)
```

Resolution order:
1. **First item with `category_id = null`** → ask the user to pick a category for that item (numbered list). The chosen category also determines the transaction type, so type is never asked separately. Repeat until every item has a category.
2. **`account_token` still null** → ask with a numbered account list (see account rule below).
3. **Everything filled** → show the summary and ask for confirmation.

Because the state is computed from the data, there are no `awaiting` / `awaiting_item_index` columns to drift out of sync with the items.

**Account rule (in order):**
- The user named an account in the text → use it.
- The user has exactly **one account** → use it (nothing to choose).
- Exactly **one favorite** → use it.
- **Multiple favorites** → ask, marking the first by `order_index` as `(default)`; empty/`yes` picks it.
- **No favorites, multiple accounts** → ask, no pre-selected default.

The final confirmation **always names the chosen account**, so an assumed account is never applied silently — the user sees it and can `cancel`.

**Example A — everything resolvable (`3.19 in cookies from savings`, "cookies" matches a category):**

```
I understood 1 transaction on "Savings":
1. €3.19 — cookies (Groceries · debit)

Confirm? Reply "yes", or "cancel".
```

**Example B — missing category, multiple favorites (`3.19 in cookies, 4.50 gasoil`):**

```
I couldn't match a category for "gasoil". Pick one:
1. Groceries
2. Car
3. Leisure
…
Reply with the number.
```
→ user replies `2` →
```
Which account?
1. Main (default)
2. Savings
3. Cash
Reply with the number (or just "yes" for the default).
```
→ user replies `1` (or `yes`) →
```
I understood 2 transactions on "Main":
1. €3.19 — cookies (Groceries · debit)
2. €4.50 — gasoil (Car · debit)
Confirm? Reply "yes", or "cancel".
```

**Result on confirm:**

```
✅ Added 2 transactions to "Main":
- €3.19 cookies (Groceries)
- €4.50 gasoil (Car)
New balance: €412.31
```

`cancel` / `/cancel` discards the pending parse at any step. Pending state **expires after 15 minutes**; replying after expiry asks the user to resend.

---

## Phase 1: Database & User Linking

**Goal:** Securely link a Telegram chat to a Grão Certo user, and hold pending parses.

### 1.1 Migration (golang-migrate, follow existing convention)
New timestamped `up`/`down` pair under `cmd/migrate/migrations/` (e.g. `<ts>_add-telegram-integration.up.sql`).

`users` table — **one** new column:
- `telegram_chat_id VARCHAR(64)` — nullable, with a **partial unique index**: `CREATE UNIQUE INDEX ... ON users (telegram_chat_id) WHERE telegram_chat_id IS NOT NULL;`

No link-token columns: the link code lives in the existing `auth_tokens` table (see 1.2).

New `telegram_pending_transactions` table (in-flight slot-filling state; survives restarts):
- `id BIGSERIAL PRIMARY KEY`
- `user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `chat_id VARCHAR(64) NOT NULL UNIQUE` — one conversation per chat; a new parse upserts/replaces the old.
- `transactions JSONB NOT NULL` — partially-resolved items `[{amount, description, category_id|null, transaction_type_id|null}]` (nulls get filled as the user answers).
- `account_token VARCHAR(255)` — nullable until the account slot is filled.
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- `expires_at TIMESTAMPTZ NOT NULL`

No `awaiting` columns (derived — see Conversation model) and **no processed-updates table** (see 3.1 on why dedup isn't needed).

Remember the matching `down` migration dropping the table, column and index.

### 1.2 Link token — reuse `auth_tokens`, don't add columns
Add to `types/auth_types.go`:
- `AuthTokenPurposeTelegramLink AuthTokenPurpose = "telegram_link"`

The flow mirrors email verification in `service/user/auth_flow.go`:
- Generate a short, user-typeable code: 8 chars from an unambiguous alphabet (no `0/O/1/l`) via `crypto/rand`. ~40 bits — enough given the per-chat rate limit in 4.3.
- Store it as `AuthToken{ID: auth.GenerateOneTimeToken(), Purpose: telegram_link, SecretHash: auth.HashSecret(code), ExpiresAt: now+15min, MaxAttempts: 5}`. Note the ID is a separate random value here (unlike email verification, which reuses the raw token as the ID) because the user types the short code, not the ID.
- `DeleteAuthTokensByUserAndPurpose` before issuing a new one, so only the latest code is live.
- Redeem with `GetAuthTokenByPurposeAndSecret(telegram_link, code)` → `ConsumeAuthToken(id)`.

This gives hashed-at-rest storage, expiry, single-use and attempt caps for free. Caveat to be aware of: because lookup is by hash, `attempts` does not throttle *guessing* (a wrong guess matches no row) — brute-force protection comes from the code's entropy plus the per-chat rate limit.

### 1.3 Type & store changes
- `types/user_types.go`: add `TelegramChatID *string` to `User` (JSON tag `telegram_chat_id,omitempty`). **The SELECT column lists in `service/user/store.go` are explicit**, so update every user query and `scanRowIntoUser` accordingly. Extend `UserStore`:
  - `LinkTelegramChatID(userID int, chatID string) error`
  - `GetUserByTelegramChatID(chatID string) (*User, error)`
  - `UnlinkTelegram(userID int) error` — clears `chat_id` **and** deletes any pending parse for that chat.
- New `types/telegram_types.go` + `service/telegram/store.go`:
  - `UpsertPendingParse(p *PendingParse) error` — replaces any existing pending row for the chat.
  - `GetPendingParse(chatID string) (*PendingParse, error)` — nil if none/expired.
  - `DeletePendingParse(chatID string) error`
- **GDPR:** `ExportData` embeds `*User`, so adding the field to the struct covers export automatically. `ON DELETE CASCADE` covers erasure of pending rows on account deletion.

### 1.4 Authenticated user endpoints (JWT, existing middleware)
Registered on the user handler alongside existing routes:
- `POST /api/v1/users/telegram/link-token` — issues the code per 1.2, returns `{ token, expires_at }`.
- `GET /api/v1/users/telegram` — returns `{ linked: bool }` for the settings UI.
- `DELETE /api/v1/users/telegram` — unlink.

These are the **only** new HTTP endpoints in v1.

### 1.5 Frontend (SvelteKit)
- "Connect Telegram" section in settings/profile.
- Calls `link-token`, shows: *"Send `/link <TOKEN>` to @GraoCertoBot within 15 minutes."*
- Shows linked/unlinked state via `GET`, and an "Unlink" button.

---

## Phase 2: LLM Parsing (reuse OpenAIStore)

**Goal:** Turn free text + the user's categories into structured, validated transactions.

### 2.1 Reuse existing LLM plumbing
- Inject the existing `OpenAIStore` (already constructed as `openAiStore` in `cmd/api/api.go`) into the Telegram service.
- **Do not reuse `cleanAiMessage` as-is** — it slices first `{` … last `}`, which corrupts a JSON **array**. Prompt the model to return an **object wrapper** `{ "transactions": [...], "account_token": ... }` so the existing object-cleaner keeps working, rather than writing a second extractor.

### 2.2 Prompt design
- Prompt includes:
  - The raw text.
  - The user's categories as an allowed list: `id`, `name`, `transaction_type`. **Only credit and debit categories** — transfer categories are excluded (see 2.3). Soft-deleted categories (`deleted_at IS NOT NULL`) are excluded too.
  - The user's accounts as an allowed list: `token`, `name`. If the user named an account in the text, return its `token`; otherwise `account_token: null`.
  - If nothing confidently fits a description, return `category_id: null` — **never guess, never fall back to a default category**.
  - Output shape: `{ "transactions": [{ "amount": number, "description": string, "category_id": number|null }], "account_token": string|null }`.
  - Currency is EUR; amounts are positive; one object per detected item.
- Temperature stays 0 (already the client default).

### 2.3 Validation & derivation (server-side, do not trust the LLM)
For each parsed item:
- `category_id`, if non-null, **must** belong to this user, must not be soft-deleted, and its `transaction_type` must be **credit or debit**. A transfer-type category would create a half-transfer via `CreateTransactionAndReturn` and silently corrupt balances — transfers require `CreateTransfer` and two accounts, which is out of scope for v1. Anything failing these checks is treated as `null` → ask the user.
- `transaction_type_id` is derived from the resolved category, never from the LLM.
- Enforce `amount` bounds matching `CreateTransactionPayload` (`gte=0, lte=999999999`), and `description` `max=255`.
- `date` defaults to **now** (server time) — the LLM does not set dates in v1.
- `account_token`, if non-null, must belong to this user; otherwise treat as unresolved.

The resolved/partial result seeds the pending row; `nextQuestion` then decides what to ask.

### 2.4 Service entry point (in-process, no HTTP)
The telegram service exposes two methods, called directly by the poller goroutine:

- `HandleLink(chatID, code string) (reply string, err error)` — redeems the auth token, links `chat_id`, consumes the token. Handles the "this chat is already linked to another account" case (partial unique index violation) with a clear message.
- `HandleMessage(chatID, text string) (reply string, done bool, err error)` — the single entry point for all non-`/link` messages:
  - **No pending conversation** → treat `text` as new transaction input: LLM parse (2.2/2.3), seed pending state, return the next question.
  - **Pending exists** → interpret `text` as the answer to whatever `nextQuestion` says is outstanding:
    - category → parse a number, set `category_id` (+ derived type) on that item.
    - account → parse a number (or empty/`yes` → marked default), set `account_token`.
    - confirmation → `yes` creates the transactions via `CreateTransactionAndReturn(..., userId)` and clears pending; anything else re-shows the summary.
  - `cancel` at any point clears pending.
  - Out-of-range or non-numeric selections re-ask with the list, state unchanged.

The slot-filling logic is pure over `(PendingParse, text) → (PendingParse, reply)` and unit-testable with no HTTP and no DB.

---

## Phase 3: The Bot Runtime

**Goal:** Talk to Telegram. **In-process Go goroutine**, not a separate service and not a Node script — one runtime, one deploy unit.

### 3.1 Long-polling, in-process
- Started from `Run()` in `cmd/api/api.go` alongside `go s.runRecurringRuleScheduler(...)`, guarded by `if config.Envs.TelegramBotToken != ""` so it's simply off when unconfigured (same pattern as the VAPID/SMTP guards).
- Poll `getUpdates` with the running offset; Telegram drops confirmed updates server-side, so the offset **is** the dedup mechanism — no `processed_updates` table.
- Works identically in dev and prod with **no public URL, no tunnel, no webhook secret**.

Why no idempotency table even later: the state machine is already near-idempotent. A duplicate `yes` arrives after the pending row is gone, so it's treated as fresh text, the LLM finds no transaction, and the user gets "couldn't understand" — no double-booking. That covers a user double-tapping `yes` too, which a dedup table wouldn't.

### 3.2 Commands — the runtime is deliberately dumb
- `/start` — greeting + how to link.
- `/link <TOKEN>` — calls `HandleLink`. Replies success or "invalid/expired token".
- **Every other message** (free text, a number, `yes`, `cancel`, …):
  - Send chat action `typing`.
  - Forward verbatim to `HandleMessage`; send back the returned `reply`.
  - The runtime holds **no** conversation state — the DB + `nextQuestion` do.

### 3.3 Failure/UX cases (all decided by the service, surfaced via `reply`)
- Unlinked chat → prompt to `/link`.
- Empty/garbled parse (0 transactions) → ask to rephrase.
- Invalid number / out-of-range selection → re-ask with the list.
- Reply after the 15-min expiry → "That request expired, please resend."
- LLM/API error → generic "couldn't process that, try again"; never leak internals.

---

## Phase 4: Security

Because the bot runs in-process, there is **no machine-to-machine channel to secure** — no `TELEGRAM_BOT_SECRET`, no shared-secret middleware, no unauthenticated `/telegram/*` routes. What remains:

1. **Identity comes from Telegram, not from a request body.** `chat_id` is read from the update returned by `getUpdates` over TLS with the bot token. The only mapping from chat → user is the linked `telegram_chat_id`; an unlinked chat can do nothing but `/link`.
2. **Link-token hygiene.** `crypto/rand`, 8-char code, 15-min expiry, hashed at rest, single-use via `ConsumeAuthToken`, superseded on reissue. The partial unique index prevents one chat linking to two users.
3. **Per-chat rate limiting.** The existing `ClientRateLimiter` is per-IP on `/api/v1` and never sees polled messages. Add a small per-chat limiter (`golang.org/x/time/rate`, already a dependency, ~15 lines mirroring `ClientRateLimiter`). The resource being protected is **OpenAI spend**, not server load, so limit the messages that trigger a parse.
4. **Least trust in LLM output.** Category ownership, non-deleted, credit/debit-only, and amount bounds are all validated server-side (2.3). The model can never write to a category or account the user doesn't own.
5. **Secrets.** `TELEGRAM_BOT_TOKEN` in env/`.env` (git-ignored), surfaced via `config.Envs`, never logged.
6. **Nothing is written without an explicit `yes`**, and the confirmation always names the account and every category.

### Later, only if polling becomes insufficient: webhooks
Add **one** endpoint that Telegram posts to directly, verifying the `X-Telegram-Bot-Api-Secret-Token` header (constant-time, `crypto/subtle`) set at `setWebhook` time, plus optionally an IP allowlist for Telegram's published CIDRs at the proxy. Identity still arrives over a trusted channel — a bot↔backend shared secret is never needed in either mode.

---

## Design boundary: LLM extracts, code decides (why v1 is not agentic)
The LLM is a **pure extractor** (text → a structured *proposal*). All state, slot-filling, validation, and every mutation are **deterministic Go**, server-validated and user-confirmed. We deliberately do **not** give the model write/delete tools or an autonomous agent loop, because:
- It keeps money-writing predictable, cheap (one completion per parse), and exhaustively unit-testable.
- The model can never write to a category/account the user doesn't own, or act without confirmation.

**Extraction quality is the real risk, not architecture.** Splitting multi-item text, amounts, and ambiguous descriptions live or die on the prompt. Mitigate with a small **eval fixture set** (start at ~10 representative Portuguese/English inputs with expected structured output) that runs against the parser with a mocked LLM, so prompt changes are regression-checked.

## Phase 5 (optional, future): read-only query agent
When richer intelligence is wanted, add tool-calling **for reads only** — e.g. "how much did I spend on groceries this month?", "what's my Savings balance?". Tools map to existing read stores (`GetTransactionStatistics`, `GetAccountsByUserId`, `GetTransactionsDTOByAccountToken`). **All writes stay on the deterministic confirmed path from Phases 2–3.**

---

## Testing Strategy (per AGENTS.md: every feature ships with tests)
- **Unit — user store:** link/unlink, `GetUserByTelegramChatID`, partial-unique constraint rejects a second user for the same chat, unlink deletes the pending row.
- **Unit — link flow:** code issued → redeemed → consumed; expired code rejected; reissue supersedes the old code. Reuses the `memoryAuthTokenStore` fake already in `service/user/auth_flow_test.go`.
- **Unit — telegram store:** pending upsert replaces, get returns nil when expired, delete.
- **Unit — parse:** mocked `OpenAIStore` returning canned JSON → asserts foreign category → null, soft-deleted category → null, **transfer-type category → null**, foreign account token → null, type derived from category, date defaults to now, wrapper extraction works.
- **Unit — `nextQuestion` + transitions** (the core): pure `(PendingParse, text) → (PendingParse, reply)`:
  - everything supplied → straight to confirmation.
  - single account / single favorite → account question skipped, account still named in the summary.
  - multiple favorites → asked, `(default)` marked, empty/`yes` selects it.
  - one null category → category asked; answering advances; two null categories ask twice.
  - out-of-range / non-numeric → re-ask, state unchanged.
  - `yes` at confirmation → creates N transactions on the right account; `cancel` → discards.
- **Integration:** mirror `cmd/api/api_test.go` for `link → message(parse) → message(category) → message(account) → message(yes)` against a test DB with a mocked LLM, driving `HandleMessage` directly.
- **Manual dev loop:** run the API with a real BotFather token; polling means no tunnel and no extra process.

---

## Build Order (executable increments)
1. ✅ Migration (`users.telegram_chat_id` + `telegram_pending_transactions`) + `User`/`UserStore` changes + telegram store.
2. ✅ `telegram_link` auth-token purpose + the JWT endpoints (`service/telegram/routes.go`) + tests.
3. ✅ LLM parse (`parser.go` + `catalog.go`, embedded prompt, credit/debit-only categories, validation) + tests with a stub LLM.
4. ✅ `nextSlot` + slot-filling transitions (`slots.go`) + `HandleMessage`/`HandleLink` (`conversation.go`) + unit tests.
5. ✅ Long-polling runtime (`bot.go`) wired into `cmd/api/api.go` behind `TELEGRAM_BOT_TOKEN`.
6. ✅ Frontend "Connect Telegram" (`src/lib/TelegramLink.svelte`, inside the profile modal).
7. *Only if polling proves insufficient:* webhook endpoint + `secret_token` verification.

**Remaining before first use:** apply the migration (`make migrate-up`), create the bot with BotFather and set
`TELEGRAM_BOT_TOKEN`, and set `TELEGRAM_BOT_USERNAME` if the handle is not `GraoCertoBot`. Bot replies are
English-only for now; every string lives in `service/telegram/messages.go`.

**Package layout** (`backend/service/telegram/`): `deps.go` (narrow interfaces onto other stores) · `store.go`
(pending parses) · `link.go` (link codes) · `routes.go` (JWT endpoints) · `catalog.go` (selectable
categories/accounts + the account rule) · `parser.go` (LLM extraction + validation) · `slots.go` (derived state,
pure) · `messages.go` (all user-facing text) · `conversation.go` (orchestration) · `bot.go` (Telegram polling).

Steps 1–6 need exactly one new env var (`TELEGRAM_BOT_TOKEN`), one new DB table, one new column, and three new HTTP endpoints (all JWT).

## Decisions (locked)
- **Bot runs in-process** as a goroutine, not as a separate service calling the API. Removes the loopback HTTP hop, the shared-secret middleware, and two env secrets, with no loss of safety — identity still comes from Telegram over TLS.
- **Link codes reuse `auth_tokens`** with a `telegram_link` purpose instead of new `users` columns. Hashed at rest, expiring, single-use, attempt-capped for free.
- **Conversation state is derived** from the pending row, not stored in `awaiting` columns — one source of truth, no drift.
- **No idempotency table.** Polling offsets dedup; the state machine absorbs the rest.
- **Unmatched category → ask, don't fallback.** `category_id: null` → numbered list.
- **Transfer-type categories are excluded** from v1 in both the prompt and validation.
- **Account is only asked when genuinely ambiguous** (multiple favorites, or no favorites with several accounts). The confirmation always names it, so nothing is silent.
- **Pending TTL = 15 minutes.**
