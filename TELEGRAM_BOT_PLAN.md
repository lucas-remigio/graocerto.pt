# Telegram Bot Integration Plan for Grão Certo (wallet-tracker)

## Overview
Build a Telegram bot that lets a linked user send natural-language text (e.g. `3.19 in cookies, 4.50 gasoil`), have an LLM parse it into structured transactions, then **fill any missing details (category, account) through a short numbered-reply conversation** and confirm before persisting. If the user's message already contains everything — including naming the account — the bot skips straight to a single confirmation.

This plan is aligned with the existing Go backend:
- LLM access already exists via `service/openai` (`OpenAIStore.GenerateGPT4Response`, gpt-4.1-mini, temp 0) — we **reuse it**, we do not add a new provider.
- Transactions are created through the existing `TransactionStore.CreateTransactionAndReturn` and require `account_token`, `category_id`, `amount`, `description`, `date`.
- Categories belong to a `transaction_type` (debit/credit) and are per-user (`GetCategoriesDtoByUserId`).
- All existing routes are JWT-protected. The Telegram endpoints are machine-to-machine and use a **separate shared-secret auth** (see Phase 4), never JWT.

## Conversation model (slot-filling — the core UX)
The bot is **stateful per chat** and runs a **slot-filling** flow. Each transaction needs four slots: `amount`, `description`, `category`, and (per message) the target `account`. The backend resolves everything it can from the first message, then **asks only for what's missing**, one slot at a time, and finishes with a single confirmation. **If the user already supplied everything — including naming the account in the text — it skips straight to confirmation.**

Required slots per transaction: `amount`, `description`, `category_id`. Shared per message: `account`.

**State machine (backend owns it; the bot just renders replies):**
1. **Parse** free text → items `[{amount, description, category_id|null}]` + a detected account (if the user named one).
2. **Next-missing-slot resolution**, in order:
   - **Any item with `category_id = null`** → ask the user to pick a category for that item (numbered list of their categories). The chosen category also determines the transaction type, so type is never asked separately. Repeat until every item has a category.
   - **Account still unknown** → ask with a numbered account list, with a **default** marked (see default-account rule below). The account question is always shown when the user didn't name an account in the text, so the choice is always consciously confirmed.
   - **Everything resolved** → show the full summary and ask for confirmation.

   **Default-account rule:** if the user named an account in the text, use it (skip the question). Otherwise the default is: exactly one favorite → that favorite; **multiple favorites → the first by `order_index`, still shown for confirmation**; no favorites → no pre-selected default (user must pick a number). Empty/`yes` on the account step selects the marked default. The final confirmation always names the chosen account, so an assumed default is never applied silently.
3. Each user reply fills the awaited slot, then re-runs step 2.
4. **Confirm** → create the transactions, clear pending state, report result.

**Example A — everything supplied (`3.19 in cookies from savings`, "cookies" matches a category):** skips straight to confirmation.

```
I understood 1 transaction on "Savings":
1. €3.19 — cookies (Groceries · debit)

Confirm? Reply "yes", or "cancel".
```

**Example B — missing category, then account (`3.19 in cookies, 4.50 gasoil`):**

```
Step 1/… I couldn't match a category for "gasoil". Pick one:
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

`users` table additions:
- `telegram_chat_id VARCHAR(64)` — nullable. **Partial unique index** where not null: `CREATE UNIQUE INDEX ... ON users (telegram_chat_id) WHERE telegram_chat_id IS NOT NULL;`
- `telegram_link_token VARCHAR(16)` — nullable, temporary.
- `telegram_link_token_expires_at TIMESTAMPTZ` — nullable.

New `telegram_pending_transactions` table (holds the in-flight slot-filling state; survives restarts, works with both long-polling and webhooks):
- `id BIGSERIAL PRIMARY KEY`
- `user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `chat_id VARCHAR(64) NOT NULL UNIQUE` — one conversation per chat; a new parse upserts/replaces the old.
- `transactions JSONB NOT NULL` — the partially-resolved items `[{amount, description, category_id|null, transaction_type_id|null}]` (nulls get filled as the user answers).
- `account_token VARCHAR(255)` — nullable until the account slot is filled.
- `awaiting VARCHAR(32) NOT NULL` — current state: `category` / `account` / `confirmation`.
- `awaiting_item_index INT` — which item's category is being asked (when `awaiting = 'category'`).
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- `expires_at TIMESTAMPTZ NOT NULL`

New `telegram_processed_updates` table (idempotency for webhook retries):
- `update_id BIGINT PRIMARY KEY`
- `processed_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- (periodic cleanup of rows older than N days)

Remember matching `down` migrations dropping columns/tables/indexes.

### 1.2 Type & interface changes
- `types/user_types.go`: add `TelegramChatID *string` to `User` (JSON tag `telegram_chat_id,omitempty`). Extend `UserStore`:
  - `SetTelegramLinkToken(userID int, token string, expiresAt time.Time) error`
  - `GetUserByLinkToken(token string) (*User, error)` — only returns if token unexpired.
  - `LinkTelegramChatID(userID int, chatID string) error` — sets chat_id, clears token + expiry.
  - `GetUserByTelegramChatID(chatID string) (*User, error)`
  - `UnlinkTelegram(userID int) error`
- New `types/telegram_types.go` + `service/telegram/` store for pending conversations & processed updates. `PendingParse` carries the items, `account_token`, `awaiting` state and `awaiting_item_index`:
  - `UpsertPendingParse(p *PendingParse) error` — replaces any existing pending row for the chat.
  - `GetPendingParse(chatID string) (*PendingParse, error)` — nil if none/expired.
  - `DeletePendingParse(chatID string) error`
  - `MarkUpdateProcessed(updateID int64) (bool, error)` — returns false if already processed.
- **GDPR:** add `TelegramChatID` to the export in `types/user_types.go` `ExportData` / wherever user profile is serialized, so linked status is included/erasable.

### 1.3 Authenticated user endpoints (JWT, existing middleware)
Registered on the user handler alongside existing routes:
- `POST /api/v1/users/telegram/link-token` — generates a cryptographically random token, sets it + expiry (e.g. 15 min), returns `{ token, expires_at }`. **Single-use** (cleared on link) and **short-lived**.
- `GET /api/v1/users/telegram` — returns `{ linked: bool }` for the settings UI.
- `DELETE /api/v1/users/telegram` — unlink (clear chat_id).

Token generation: use `crypto/rand`, 8 chars from an unambiguous alphabet (no `0/O/1/l`). Not 6 — a bit more entropy, still typeable.

### 1.4 Frontend (SvelteKit)
- "Connect Telegram" section in settings/profile.
- Calls `link-token`, shows: *"Send `/link <TOKEN>` to @GraoCertoBot within 15 minutes."*
- Shows linked/unlinked state via `GET`, and an "Unlink" button.

---

## Phase 2: LLM Parsing (reuse OpenAIStore)

**Goal:** Turn free text + the user's categories into structured, validated transactions.

### 2.1 Reuse existing LLM plumbing
- Inject the existing `OpenAIStore` (already constructed as `openAiStore` in `cmd/api/api.go`) into the Telegram service.
- **Do not reuse `cleanAiMessage` as-is** — it slices first `{` … last `}`, which corrupts a JSON **array**. Add an array-aware extractor (first `[` … last `]`) or, better, prompt the model to return an **object wrapper** `{ "transactions": [...] }` so the existing object-cleaner keeps working. Prefer the wrapper approach to minimize new parsing code.

### 2.2 Prompt design
- System/user prompt includes:
  - The raw text.
  - The user's categories as an allowed list: `id`, `name`, `transaction_type` (debit/credit). The model picks `category_id` **only** from this list; **if nothing confidently fits, it returns `category_id: null`** (we do NOT force a fallback category — the bot will ask the user, per decision #1).
  - The user's accounts as an allowed list: `token`, `name`. If the user named an account in the text, return its `token` as `account_token`; otherwise `account_token: null`.
  - Instruction to output `{ "transactions": [{ "amount": number, "description": string, "category_id": number|null }], "account_token": string|null }`.
  - Currency is EUR; amounts are positive; one object per detected item.
- Temperature stays 0 (already the default in the client).

### 2.3 Validation & derivation (server-side, do not trust the LLM)
For each parsed item:
- If `category_id` is non-null it **must** belong to this user (validate against fetched categories); if it doesn't, treat it as null (→ ask the user).
- `transaction_type_id` is derived from the resolved category, never from the LLM.
- Enforce `amount` bounds matching `CreateTransactionPayload` validation (`gte=0, lte=999999999`).
- `date` defaults to **now** (server time) — the LLM does not set dates in v1.
- `account_token`, if non-null, must belong to this user; otherwise treat as unresolved (→ ask).
- The resolved/partial result seeds the pending-conversation row; the state machine (Conversation model) then decides the next question.

### 2.4 Machine endpoints (shared-secret auth, see Phase 4 — NOT JWT)
Two endpoints. The **backend owns the whole slot-filling state machine**, so the bot stays dumb — it forwards every message and renders whatever `reply` comes back.

- `POST /api/v1/telegram/link` — body `{ chat_id, token }`. Looks up user by unexpired token, links chat_id, clears token. Returns `{ first_name }` or an error if token invalid/expired.
- `POST /api/v1/telegram/message` — body `{ chat_id, text }`. This is the single entry point for all non-`/link` messages:
  - If **no** pending conversation exists → treat `text` as new transaction input: run the LLM parse (2.2/2.3), seed pending state, compute the next question.
  - If a pending conversation **exists** → interpret `text` as the answer to the current `awaiting` slot:
    - `awaiting = category` → parse a number, set `category_id` (+ derived type) on `awaiting_item_index`.
    - `awaiting = account` → parse a number (or empty/`yes` → favorite default), set `account_token`.
    - `awaiting = confirmation` → `yes` creates the transactions via `CreateTransactionAndReturn(..., userId)` and clears pending; anything else re-shows the summary.
  - `cancel` at any point clears pending.
  - **Always returns `{ reply: string, done: bool }`** — `reply` is the exact text to send; `done` marks the conversation complete (created or cancelled). Out-of-range selections re-ask with the list.

Both endpoints are testable with a **mocked `OpenAIStore`** (interface already exists) and a test DB, mirroring `cmd/api/api_test.go`. The state-machine transitions are pure functions over `(PendingParse, text)` → `(nextState, reply)` and unit-testable without HTTP.

---

## Phase 3: The Telegram Bot Service

**Goal:** The process users talk to. **Go-native**, reusing repo config — not a separate Node script (keeps one runtime/deploy unit; the backend is Go).

### 3.1 Runtime & mode
- **Dev: long-polling** (no public URL needed — testable purely locally). Poll `getUpdates`; the offset naturally dedups.
- **Prod: webhooks** (you already terminate TLS in `docker-compose.prod.yaml`). Register via `setWebhook` with a `secret_token`; dedup with `telegram_processed_updates`.
- One env flag selects the mode. Bot talks to the backend Telegram endpoints over the shared secret.

### 3.2 Commands — the bot is deliberately dumb
- `/start` — greeting + how to link.
- `/link <TOKEN>` — calls `POST /telegram/link`. Replies success or "invalid/expired token".
- **Every other message** (free text, a number answering a question, `yes`, `cancel`, …):
  - Send chat action `typing`.
  - Forward verbatim to `POST /telegram/message`; send back the returned `reply`.
  - The bot does **not** track conversation state — the backend does. This keeps the bot trivial and the logic fully unit-testable server-side.

### 3.3 Failure/UX cases (all decided by the backend, surfaced via `reply`)
- Unlinked chat → prompt to `/link`.
- Empty/garbled parse (0 transactions) → ask to rephrase.
- Invalid number / out-of-range selection → re-ask with the list.
- Reply after the 15-min expiry → "That request expired, please resend."

---

## Phase 4: Security Hardening (Telegram ↔ backend)

The Telegram endpoints accept a `chat_id` in the body — **identity must come from a trusted channel, never from the body alone**. Layers:

1. **Shared secret bot ↔ backend (always on).** New env `TELEGRAM_BOT_SECRET` (in `config`). The bot sends it on every call to `/api/v1/telegram/*` via header (e.g. `X-Telegram-Bot-Secret`). A dedicated middleware validates it with **constant-time compare** (`crypto/subtle`) and rejects otherwise. These routes are registered **without** `AuthMiddleware` and **with** this secret middleware instead.
2. **Telegram webhook secret (prod).** On `setWebhook`, set `secret_token`; verify the `X-Telegram-Bot-Api-Secret-Token` header on every inbound webhook (constant-time) before processing. Reject mismatches.
3. **Optional IP allowlist (prod).** Restrict the webhook path to Telegram's published CIDR ranges at the proxy layer.
4. **Rate limiting.** Reuse/extend the existing `ClientRateLimiter`; also throttle **per chat_id** to blunt spam/abuse from a single linked account.
5. **Link-token hygiene.** `crypto/rand` generation, short expiry, single-use (cleared on link), and a cap on link attempts to prevent brute force. Partial-unique index prevents one chat linking to two users.
6. **Idempotency.** Webhook retries deduped via `telegram_processed_updates` so a retried delivery never double-books.
7. **Least trust in LLM output.** Category ownership + type + amount bounds validated server-side (Phase 2.3); the model can never write to a category/account the user doesn't own.
8. **Secrets management.** `TELEGRAM_BOT_TOKEN`, `TELEGRAM_BOT_SECRET`, `TELEGRAM_WEBHOOK_SECRET` live in env/`.env` (already git-ignored), surfaced through `config.Envs`, never logged.

---

## Design boundary: LLM extracts, code decides (why v1 is not agentic)
The LLM is a **pure extractor** (text → a structured *proposal*). All state, slot-filling, validation, and every mutation are **deterministic Go**, server-validated and user-confirmed. We deliberately do **not** give the model write/delete tools or an autonomous agent loop, because:
- It keeps money-writing predictable, cheap (one completion/message), and exhaustively unit-testable.
- The model can never write to a category/account the user doesn't own, or act without confirmation.

**Extraction quality is the real risk, not architecture.** Splitting multi-item text, amounts, and ambiguous descriptions live or die on the prompt. Mitigate with a small **eval fixture set** (10–20 representative Portuguese/English inputs with expected structured output) that runs against the parser with a mocked-then-real LLM, so prompt changes are regression-checked.

## Phase 5 (optional, future): read-only query agent
When richer intelligence is wanted, add tool-calling **for reads only** — e.g. "how much did I spend on groceries this month?", "what's my Savings balance?". Tools map to existing read stores (`GetTransactionStatistics`, `GetAccountsByUserId`, `GetTransactionsDTOByAccountToken`). **All writes stay on the deterministic confirmed path from Phases 2–3.** This is where an agentic layer earns its keep without putting the model near mutations.

---

## Testing Strategy (per AGENTS.md: every feature ships with tests)
- **Unit — user store:** token set/expire/single-use, link/unlink, `GetUserByTelegramChatID`, partial-unique constraint.
- **Unit — telegram store:** pending upsert/get/expire/delete, `MarkUpdateProcessed` dedup.
- **Unit — parse:** mocked `OpenAIStore` returning canned JSON → assert validation rejects foreign category/account (→ null), derives transaction_type, defaults date, wrapper extraction.
- **Unit — state machine** (the core of decision #1): pure `(PendingParse, text) → (nextState, reply)` transitions:
  - everything supplied (category + account in text) → jumps straight to `confirmation`.
  - one null category → `awaiting = category`; answering advances to `account`; answering advances to `confirmation`.
  - empty/`yes` on the account step → favorite default selected.
  - out-of-range / non-numeric answers → re-ask, state unchanged.
  - `confirmation` + `yes` → creates N transactions on the right account; `cancel` → discards.
- **Unit — auth middleware:** wrong/missing `X-Telegram-Bot-Secret` → 401; correct → passes (constant-time).
- **Integration:** mirror `cmd/api/api_test.go` for `link → message(parse) → message(category) → message(account) → message(yes)` against a test DB with a mocked LLM.
- **Manual dev loop:** long-polling bot against local backend with a real BotFather token — no tunnel required.

---

## Build Order (executable increments)
1. Migration (users columns + pending + processed_updates) + `User`/`UserStore` changes + telegram store + tests. *(No external deps — fully dev-testable.)*
2. JWT user endpoints (`link-token`, status, unlink) + tests.
3. Shared-secret middleware + `/telegram/link` + tests.
4. LLM parse (reuse `OpenAIStore`, wrapper prompt, null category/account support, validation) + tests with mocked LLM.
5. Slot-filling state machine + `/telegram/message` (category → account → confirmation → create) + state-transition unit tests.
6. Go long-polling bot (dumb forwarder) wired to the endpoints; manual end-to-end in dev.
7. Frontend "Connect Telegram" settings UI.
8. Prod: webhook mode + `secret_token` + idempotency + deploy.

## Decisions (locked)
- **Unmatched category → ask, don't fallback.** If the LLM can't confidently match, it returns `category_id: null` and the bot asks the user to pick from a numbered list (decision #1). Same pattern as accounts. If the user supplied everything (incl. account named in text), skip straight to confirmation.
- **Default account = favorite account** (decision #2). Marked `(default)` in the numbered list; empty/`yes` picks it. **If several favorites exist, assume the first by `order_index` but still show the account question for confirmation**; if none, no default is pre-selected. The final confirmation always names the account, so a default is never applied silently.
- **Pending TTL = 15 minutes** (decision #3).
