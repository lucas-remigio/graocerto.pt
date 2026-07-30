# Feature Plans — Grão Certo

Five features ordered by **value / effort ratio** (best ratio first). Each is grounded in
existing infrastructure — the recurring theme is *closing the loop on features that are already
80% built* (budgets are computed but never alerted; forecasts are computed but never surfaced as
a decision).

Status legend: `todo` / `in progress` / `done`.

Reference anchors (verified against current code):

- Scheduler: `backend/cmd/api/api.go:199` `runRecurringRuleScheduler` — hourly `time.Ticker`,
  calls `GeneratePendingTransactionsForDueRules` + `GenerateRecurringDueTomorrowNotifications`,
  then `pushRecurringNotifications` (`api.go:224`).
- Notifications: `backend/service/notification/store.go`, types in
  `backend/types/notification_types.go`, push in `backend/service/notification/push.go`.
- Stats engine: `backend/service/transaction/store.go:1488` `GetTransactionStatistics(accountToken, month, year)`
  — already returns per-category `budget_percentage`, `daily_totals`, `largest_*`, `totals`.
- Recurring forecast: `backend/service/recurring_rule/store.go:76`
  `GetRecurringForecast(userID, accountToken, days)` → `RecurringForecastResponse`.
- DI root: `backend/cmd/api/api.go:47-100`. Frontend API clients: `frontend/src/lib/services/`,
  stores: `frontend/src/lib/stores/`, types: `frontend/src/lib/types.ts`.

---

## 1. Budget alerts via the existing push pipeline  `done`

**Value: high · Effort: low · Best ratio.**

Categories already carry `budget`; `GetTransactionStatistics` already computes
`budget_percentage` per category. Today it is passive — only visible if the user opens stats.
Turn it into a proactive push at 80% and 100% of a category's monthly budget.

### Approach
Add a new generator to the notification service, called from the existing hourly scheduler
alongside `GenerateRecurringDueTomorrowNotifications`. For each user with `enabled` prefs and at
least one category with a non-null budget, compute month-to-date spend per budgeted category
(reuse the stats query path) and emit a notification when it crosses a threshold it hasn't
crossed before this month.

### Key design decision — schema
The `notifications` table is specialized for recurring alerts (`debit_count`, `total_debit`,
`target_date`, unique on `(user_id, type, account_token, target_date)`) and has **no category or
free-text field**. Two options:

- **(A, recommended) Reuse columns semantically.** New `type = 'budget_threshold'`. Store
  `category_id` in a new nullable `category_id INT` column (small migration), the threshold
  (80/100) reused via `notify_days_ahead` is a hack — instead add `threshold_pct SMALLINT` NULL.
  Dedup via a partial unique index `(user_id, type, category_id, threshold_pct, <month>)`.
  Keeps one notifications table and the existing push/read/unread UI works unchanged.
- **(B) Generic notification payload.** Add `title`/`body` TEXT columns and stop overloading the
  numeric columns. Cleaner long-term, but a bigger change to the existing recurring path and the
  frontend rendering. Defer unless a third notification type appears.

Go with **A**: migration adds `category_id INT NULL` + `threshold_pct SMALLINT NULL` and a
partial unique index scoped to the current month (e.g. include a `period_month DATE` set to the
first of the month) so each (category, threshold) fires once per month.

### Files
- `backend/cmd/migrate/migrations/<ts>_add-budget-notification-columns.{up,down}.sql`
- `backend/service/notification/store.go` — `GenerateBudgetThresholdNotifications() error`
  (single set-based INSERT ... SELECT mirroring the recurring generator's style, joining
  categories with budgets against month-to-date transaction sums).
- `backend/types/notification_types.go` — add `CategoryID *int`, `ThresholdPct *int` to
  `Notification`; add method to `NotificationStore` interface.
- `backend/cmd/api/api.go:204` — call the new generator inside the ticker loop.
- `backend/service/notification/push.go` / `pushRecurringNotifications` — extend payload builder
  to render a budget message for the new type (e.g. "80% of Groceries budget used — €340/€400").
- Frontend: `frontend/src/routes/(protected)/notifications/+page.svelte` +
  `frontend/src/lib/types.ts` (`NotificationItem` gets optional `category_id`, `threshold_pct`)
  — render the new type. Reuse existing notification list UI.

### Trade-off
Budget periods are monthly and per-category; multi-account budgeting is out of scope for v1
(budgets are per category, categories are per user, so aggregate across accounts). State this in
the UI copy.

### Tests
- `store_test.go`: no budget → no notification; spend at 79% → none; crosses 80% → one row;
  re-run same hour → no duplicate (dedup index); crosses 100% → second row; new month resets.

---

## 2. "Safe to spend" / end-of-month projection on the dashboard  `done`

> **Shipped notes.** Endpoint is `GET /recurring-rules/projection?account_token=` (query param,
> mirroring the sibling `forecast` route — not the path param originally sketched). Base balance is
> the account **pending** balance, not the confirmed balance: recurring rules generate *pending*
> transactions and advance `next_run_date`, so those occurrences leave the forecast window and would
> otherwise be dropped. The `safe_to_spend` + user buffer was **not** built for v1 — there is no
> buffer storage, so it would have equalled `projected_balance`; the headline is the projected
> end-of-month balance. `GetRecurringForecast` was refactored to share a pure `expandRuleOccurrences`
> helper with the projection (unit-tested, no DB), which also avoids the `days <= 0 → 30` default
> misfiring on the last day of the month.
>
> **Frontend placement:** not a standalone dashboard card (rejected as too heavy). Final form is a
> row below the statistics-page summary grid (`TransactionStatistics.svelte`), current-month only.
> It **always shows `Saldo atual`** (the account's real/pending balance — the number the user cares
> about most) so it never disappears; when recurring movements are still expected it extends into a
> muted `→ −€X até ao fim do mês` (net, with an income/expenses tooltip) and a bold `Previsto`
> end-of-month figure. NB: the existing "Net Balance" stat is already labelled **Saldo Líquido** (the
> month's credit−debit), so the account balance is deliberately labelled **Saldo atual** to avoid the
> name clash.

**Value: high · Effort: low.**

`GetRecurringForecast` already returns upcoming credits/debits and a `summary`
(credit/debit/difference) over 30/60/90 days. The home dashboard doesn't surface the one number
users open the app for: *will I make it to payday?*

### Approach
Backend already has everything for a 30-day window. Add a thin endpoint/aggregation that returns:
`projected_end_of_month_balance = current_balance + upcoming_recurring_credits −
upcoming_recurring_debits` (clip forecast to end-of-month), plus a `safe_to_spend` figure
(projected balance minus a user-set buffer, default 0).

Prefer computing on the backend to keep the money math in one place. A new method
`GetCashflowProjection(userID, accountToken)` on the recurring or transaction store that reuses
`GetRecurringForecast(..., 30)` and the account balance.

### Files
- `backend/service/recurring_rule/store.go` — `GetCashflowProjection` (reuse `GetRecurringForecast`).
- `backend/service/recurring_rule/routes.go` — `GET /recurring-rules/projection/{accountToken}`.
- `backend/types/recurring_rule_types.go` — `CashflowProjectionResponse`.
- Frontend: new dashboard card in `frontend/src/routes/(protected)/home/+page.svelte`; API call in
  `frontend/src/lib/services/dataService.ts`; type in `types.ts`.

### Trade-off
Projection quality depends on recurring rules being complete — a user with no rules gets
`projected == current`. That's fine (honest), and it's also the hook that motivates feature #4
(auto-detect subscriptions) to make projections meaningful.

### Tests
- `store_test.go`: no rules → projection equals current balance; one €50 debit due before month
  end → projection = balance − 50; credit after month end → excluded.

---

## 3. Month-over-month spending comparison → shipped as a macro "Trends" screen  `done`

**Value: high · Effort: low–medium.**

> **Pivoted during build.** A two-month diff was the original plan, but the better product is a
> multi-month **Trends** screen — "vs last month" is just the last two points of a trend line, and a
> chart directly delivers the "Spot trends by category" promise. The pivot was cheap because the app
> already ships **Chart.js 4** (`InvestmentChart.svelte` pattern) and the backend is one grouped
> query, not N stat calls. The half-built comparison endpoint/UI was reverted (no dead code) — any
> "vs last month" delta is derivable client-side from the trends series.
>
> **Shipped:**
> - Backend: `GetCategoryMonthlyTrends(accountToken, months, transactionTypeID)` — one grouped SQL
>   query (`SUM(ABS(amount)) … GROUP BY year, month, root category`), subcategories rolled up into
>   their parent to match the stats view. Route `GET /transactions/trends/{accountToken}?months=&type=`
>   (`months` 1–24 default 12; `type` debit|credit default debit). Pure `buildMonthAxis` unit-tested
>   (window length, year rollover).
> - Frontend: new top-level **Trends** page (`src/routes/(protected)/trends/`, nav entry added) with
>   account picker, Spending/Income toggle, 6M/12M/24M range. Chart = **total line + category picker**
>   (chosen over multi-line/stacked to avoid spaghetti): always shows the grand total; users add
>   category overlays via a dropdown, remove via chips. `TrendsChart.svelte` mirrors the existing
>   Chart.js line-chart pattern (theme-aware, locale-aware month labels).

Stats are single-period. The i18n string already promises "Spot trends by category". Deliver it
by diffing two periods.

### Approach
Call `GetTransactionStatistics` for the current and previous month, compute per-category deltas
(absolute + %) and a headline total delta. Do the diff server-side in a new method to keep the
two calls and the delta logic together and testable; return a `StatisticsComparison`.

### Files
- `backend/service/transaction/store.go` — `GetStatisticsComparison(accountToken, month, year)`
  (calls existing `GetTransactionStatistics` twice, computes deltas; handle month rollover for
  the previous period).
- `backend/service/transaction/routes.go` — `GET /transactions/statistics/compare/...` (mirror
  the existing statistics route param parsing at `routes.go:275`).
- `backend/types/transaction_types.go` — `StatisticsComparison` / per-category delta struct.
- Frontend: extend the stats view to show up/down arrows + deltas per category; type in `types.ts`.

### Trade-off
New categories (no prior data) and deleted categories need explicit handling — show as
"new"/"—" rather than a misleading +100%/−100%. Handle the January→previous-December year
rollover in the previous-period calculation.

### Tests
- `store_test.go`: category up 20% flagged correctly; category with no prior month → "new";
  category present last month, absent now → shown as removed; Jan compares against prior Dec.

---

## 4. Auto-detect recurring subscriptions from history  `todo`

**Value: high (delight) · Effort: medium.**

Recurring rules are 100% manual. Every user has untracked subscriptions already sitting in their
transaction history. Detect repeated amount+description at regular intervals and suggest creating
a recurring rule with one tap. Doubles as a "subscription audit" — on-mission for financial
literacy.

### Approach
Read-only analysis pass over a user's transactions: group by normalized description + similar
amount, look for ≥3 occurrences at a roughly constant interval (monthly/weekly). Return
suggestions; the user confirms, which creates a rule via the existing
`CreateRecurringRule` path — no new write path, no LLM needed.

### Files
- `backend/service/recurring_rule/store.go` — `DetectRecurringCandidates(userID) ([]Candidate, error)`
  (pure grouping/interval logic over transactions; unit-testable without HTTP).
- `backend/service/recurring_rule/routes.go` — `GET /recurring-rules/suggestions`.
- `backend/types/recurring_rule_types.go` — `RecurringCandidate` (account, category guess,
  amount, description, detected frequency, sample dates).
- Frontend: a "Suggested subscriptions" section on
  `frontend/src/routes/(protected)/recurring-payments/`; "Add" reuses the existing create flow.

### Trade-off
Detection is heuristic — tune thresholds (min occurrences, amount tolerance ±5–10%, interval
tolerance ±3 days) to favor precision over recall (a missed suggestion is invisible; a bad one is
annoying). Keep it deterministic Go; explicitly **not** an LLM job.

### Tests
- `store_test.go`: 3 monthly €12.99 "Netflix" → one monthly candidate; irregular one-offs → none;
  amounts within tolerance grouped, outside tolerance split; already-covered-by-a-rule excluded.

---

## 5. Bulk import from bank statement (CSV) with AI-assisted categorization  `todo`

**Value: highest absolute · Effort: high → lands at #5 on ratio.**

Manual entry is the #1 reason expense trackers get abandoned. Portuguese users live in
Revolut/bank CSV exports. This removes the biggest adoption barrier.

### Approach — "LLM extracts, code decides, user confirms"
Reuse the exact philosophy already written in `TELEGRAM_BOT_PLAN.md`: the LLM is a pure
extractor, all validation and writes are deterministic Go, everything is user-confirmed.

1. Upload CSV → parse rows (date, description, amount, optional balance).
2. Column mapping UI (banks differ) — remember mapping per user later, hardcode common presets
   (Revolut, Millennium, CGD) first.
3. Batch category suggestion: send rows + the user's category list to the existing
   `OpenAIStore.GenerateGPT4Response` (`backend/service/openai/store.go:26`), model returns
   `category_id | null` per row **only** from the allowed list. Server validates ownership;
   null/foreign → user picks.
4. **Dedup** against existing transactions (same account, date, amount, description) to avoid
   double-imports on re-upload.
5. Preview table (editable category/account per row) → confirm → bulk `CreateTransaction`.

### Reuse note / caution
`cleanAiMessage` (`openai/store.go:97`) slices first `{` … last `}` and **corrupts JSON arrays** —
the Telegram plan already flags this. Use the object-wrapper approach
(`{ "rows": [...] }`) so the existing cleaner keeps working; do not hand-roll array parsing.

### Files
- `backend/service/transaction/` (or new `service/import/`) — parse, dedup, batch-create;
  `backend/service/openai/` — a categorization prompt in `backend/prompts/`.
- `backend/types/` — import payload/preview/response structs.
- Frontend: new import page + column-mapping + preview components; service in
  `frontend/src/lib/services/`.

### Trade-off
File parsing + mapping UI + dedup + preview is genuinely the heaviest lift here. Ship in slices:
(a) CSV parse + manual categorize + dedup + bulk create first (fully useful without AI), then
(b) add AI suggestions as an enhancement. This de-risks the AI dependency and cost.

### Tests
- Parser: known bank CSV fixtures → expected rows; malformed rows rejected with clear errors.
- Dedup: re-importing the same file adds zero transactions.
- Categorization: mocked `OpenAIStore` returns foreign/null category → validation forces user pick
  (mirror the mocked-LLM pattern the Telegram plan specifies).

---

## Honorable mention — Telegram bot

Already fully spec'd in `TELEGRAM_BOT_PLAN.md`. A real delight feature (log expenses via chat),
but a multi-phase new service with its own auth/security surface → lower value/effort ratio than
the quick wins above. Build after #1–#3 land.

## Recommended sequencing
Ship **#1 + #2 first** (≈1–2 days combined) — they convert two half-finished features (budgets,
forecasts) into daily-felt value with almost no new infrastructure. Then #3, #4, and finally the
#5 CSV import (sliced: non-AI first).
