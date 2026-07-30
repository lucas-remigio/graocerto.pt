# Trends / Macro-Statistics Plan — Grão Certo

A set of small, high-value additions to the **macro statistics** page (`/trends`), optimised for
one goal: **intuitive, informative, and simple to read at a glance.** The whole thing hangs off
the chart that already exists — you select a category, and the page tells you more about it. No
new page, no new graph, no configuration.

Status legend: `todo` / `in progress` / `done`.

## Design principles

1. **One question per visual.** Glance → answer. No chart needs a paragraph to explain it.
2. **Reuse the chart that exists.** `TrendsChartCard` already lets you toggle any category's
   trendline. Everything here enriches *that* interaction rather than adding new surfaces.
3. **Backend computes, frontend draws** (`AGENTS.md` → "Frontend is the visualization layer").
   Money math is aggregated server-side; the client formats and lays out.
4. **No hidden state.** The category you care about is the one you clicked on the chart — nothing
   to pick twice, nothing to persist.

## The only backend change

Add **`income []float64`** (aligned to `Months`) plus its sum **`window_income`** to
`CategoryTrendsResponse` (`backend/types/transaction_types.go:61`): per-month `Σ Credit amount
WHERE transfer_group_id IS NULL` (exclude money transferred *in* so it isn't counted as income).

**Why it's a second query, not "the same pass":** the main trends query is hard-filtered to a
single `transaction_type_id = $2` (`store.go:1593`), so when the user views *spending* (debit)
the result set contains **zero credit rows** — there is no income in that pass to sum. Income is
always the credit side (`CreditTransactionType = 1`), so it needs its own small aggregate query
(`GROUP BY year, month`, `transfer_group_id IS NULL`). It reuses the same `indexByKey` axis map.
**No new endpoint, route, DI, migration, or persistence.**

To keep it testable without a DB (all existing store tests are pure-function tests), the
bucketing is extracted into a pure helper `bucketMonthlyAmounts(rows, indexByKey, n)` — the SQL
just feeds it, and the unit test targets the helper (same shape as `computeCategoryMovers`).

That single array unlocks the flagship feature below and works in both chart modes (income is
always the credit side, regardless of the debit/credit toggle).

## Why "% of income", not net (credit − debit)

A transfer out is stored on the source account as a **Debit** with a non-null `transfer_group_id`
(`store.go:293-307`), so money moved to savings already counts as an expense and `credit − debit
≈ 0`. The honest, simple read is **a picked category as a share of income**: for your savings
category that share **is** the savings rate; for rent it reads "30% of income". Same mechanism,
useful for every category.

---

## Section A — the category chart, enriched  `done`

**Question:** *"Where does my money go, how much of my income is it, and is it trending?"*
**Value: high · Effort: low.** This is where four of the five features live.

### On selection — inline stats for the picked category
When one or more categories are toggled on in the legend (already supported), show a small stat
line next to the selection / in the KPI strip, from data now all on the client:

- **% of income** = selected total ÷ window income → **savings rate** for the savings category;
  "X% of income" for any other. *(feature 1 — the flagship)*
- **% of spend**  = selected total ÷ `window_total`. *(feature 3, applied to the live selection)*
- **~€/mo · €/yr** = selected total ÷ `months.length`, × 12. *(feature 5)*

Optional, free: the chart tooltip can also show that month's `category[m] / income[m]` — a
per-month % — with `income[m] = 0` months shown as "—", never ∞. Ship the window figure first;
the tooltip is a trivial add on the same `income[]` data.

### Always-on chart polish
- **Concentration line** above the chart: "Top 3 categories = 71% of spending", from
  `categories[].total` vs `window_total`. Answers "where does it go" before any click. *(feature 3
  aggregate)*
- **Trailing 3-month smoothing toggle** — optional dashed line from `totals[]` so the noisy
  monthly line shows its real direction. Client-side. *(feature 4)*

### Done when
Selecting the savings category shows its **% of income (= savings rate)**, % of spend, and
~€/mo·€/yr; the concentration line and smoothing toggle render; everything updates with the range
and debit/credit toggles. `income[]` excludes incoming transfers (the SQL filters
`transfer_group_id IS NULL`). One `store_test.go` case asserts the pure `bucketMonthlyAmounts`
helper buckets amounts onto the axis and drops out-of-window months.

---

## Section B — what's changing  `done`

**Question:** *"What shifted recently, and how does this month compare?"*
**Value: medium · Effort: trivial · Frontend-only.**

Keep the existing movers card (`TrendsMovers`). Add a slim row of **context chips** from
`totals[]` + `months[]`: **highest-spend month · lowest-spend month · vs last month %**
(`Math.max/min`, ignoring leading-zero months; hide MoM under 2 months). *(feature 2)*

### Done when
Chips show correct labels and %; handle <2 months and all-zero windows cleanly.

---

## Iteration 2 — per-category detail  `done`

The window-aggregate stat strip was replaced with a **per-category detail table** (one row per
selected category) so the selection answers "how is *this* category doing", not just the whole
window:

- **3-month curve per category** — the smoothing toggle now adds a dashed 3-mo average to *every*
  drawn line (total + each selected series), computed inside `TrendsChart` (`movingAverage`).
  Dashed lines are filtered out of the legend so they read as guides, not series.
- **MoM per category** — each row shows latest month vs previous (`momOf`): signed %, `new` when
  spend appeared with no prior month, `—` under 2 months.
- **% per month per category** — the row's `this mo` column is the latest month's share; the chart
  **tooltip** appends `(Y%)` = that point's share of the month's grand total for every category,
  so hovering any month reveals the full per-month × per-category matrix without extra surfaces.

Table columns: `% income · this mo · MoM · avg/mo`, wrapped in `overflow-x-auto`. Still no new
backend work — all three are thin derivations of `income[]` + `totals[]` already on the client.

## Traceability — the 5 features (none dropped)

| # | Feature | Lives in |
|---|---------|----------|
| 1 | Savings rate = **% of income** for the picked category | Section A · on-selection |
| 2 | Best / worst month + month-over-month | Section B · context chips |
| 3 | Category concentration (top-3 share) | Section A · concentration line (+ % of spend on selection) |
| 4 | 3-month rolling average | Section A · smoothing toggle |
| 5 | Per-category €/mo + annualized | Section A · on-selection |

## Decisions

- **No dedicated savings graph, no savings picker, no localStorage.** Savings rate is just the
  picked category's **% of income** shown inline on the existing chart. (Reverted the Money Flow
  graph — it re-introduced complexity we'd already removed.)
- **One backend change:** add `income[]` + `window_income` to `CategoryTrendsResponse`, fed by
  one extra credit-only aggregate query (not the same pass — the main query is single-type). No
  new endpoint.
- **Window figure first**, per-month % in the tooltip as a trivial optional on the same data.

## Build order

Basically one PR: add `income[]` (backend) → inline selection stats + concentration + smoothing
(Section A) → context chips (Section B). All small; Section A's flagship (% of income) is the
piece that carries the value.

## Reference anchors (verified)

- Trends store: `backend/service/transaction/store.go:1561` `GetCategoryMonthlyTrends`; axis
  `buildMonthAxis` (`store.go:1769`); movers `computeCategoryMovers` (`store.go:1695`).
- Transfer storage: `store.go:293-307` (source debit carries `transfer_group_id`).
- Types: `backend/types/transaction_types.go:61` `CategoryTrendsResponse` (add `Income`).
- Frontend: page `frontend/src/routes/(protected)/trends/+page.svelte`; components
  `TrendsChartCard`, `TrendsChart`, `TrendsKpis`, `TrendsMovers`; API
  `dataService.fetchCategoryTrends`; types `frontend/src/lib/types.ts:244`.

## Conventions

Every label via `$t(...)` (existing `trends.*` keys). Follow **"Frontend is the visualization
layer"** (`AGENTS.md`): the one real aggregate (`income[]`) is computed in the store; the inline
stats, concentration, smoothing, and chips are thin presentational derivations of aggregates
already on the client.
