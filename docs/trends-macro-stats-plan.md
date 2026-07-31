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

Table columns: `% income · this mo · MoM · avg/mo`, shown as a wrapping row per category (no
horizontal scroll, so DaisyUI tooltips on each stat label aren't clipped). Every stat label —
here and on the KPI tiles — has an on-hover tooltip explaining the exact math. Still no new
backend work — all three are thin derivations of `income[]` + `totals[]` already on the client.

## Iteration 3 — "What changed this month" (movers rework)  `done`

The old movers card mixed two scopes (total best/worst-month chips + per-category
*recent-half-vs-earlier-half* momentum) under one caption, and flooded with "New" on short
histories. Reworked to **one consistent metric — month-over-month**:

- `computeCategoryMovers` (`store.go`) now compares the **latest month vs the previous month** per
  root, ranked by **absolute euro change** (a big € swing beats a big % on a tiny category). A
  category that fell to exactly zero is skipped (within a possibly-incomplete current month an
  absence reads as "not yet", not a decrease). "new" = spend this month, none last month.
- The card (`TrendsMovers`) shows a context line (total MoM + peak month) then **Biggest rises /
  Biggest drops**, each row = name · sparkline · euro change + % (or "New"). Consistent with the
  per-category MoM in the detail table.

## Iteration 4 — customisable comparison months  `done`

The MoM story was hardwired to `latest vs previous` = axis indices `n-1`/`n-2`. But `buildMonthAxis`
(`store.go:1826`) always ends on the **current (incomplete) calendar month**, so on the 1st of a
month the whole "what changed" card compared an empty new month to a full previous one. Fixed by
letting the user **compare any two months** — with the real logic kept **on the backend** (single,
unit-tested implementation), driven by a refetch (prior data stays on screen, so no jarring spinner):

- **Backend owns the policy + ranking.** `GetCategoryMonthlyTrends` gained `compareBase,
  compareCurrent` params (axis indices; `-1`/out-of-range → default). `defaultCompareIndices(axis,
  now)` (`store.go`) picks the default — when the last bucket is the current month and **< 1/3 of it
  has elapsed**, it falls back to the two most recent *complete* months (fixes the 1st-of-month
  view); otherwise keeps current-vs-previous. `computeCategoryMovers` is now parametrised by the two
  indices. The response echoes `compare_base` / `compare_current` so the pickers stay in sync. Route
  (`routes.go`) validates the pair is within `[0, months)`. Tests: `TestDefaultCompareIndices` +
  the updated `TestComputeCategoryMovers`.
- **Frontend is pure visualization.** `TrendsCompareControls.svelte` = two `[base] → [current]`
  `<select>`s in the movers card header. On change it refetches via
  `dataService.fetchCategoryTrends(..., compareBase, compareCurrent)`; the page (`+page.svelte`)
  mirrors the echoed indices into `baseIdx`/`curIdx` and passes them to `TrendsMovers` (context +
  delta) and `TrendsChartCard` (detail-table `MoM` + `% of month` follow the focus month). The
  structural `$effect` (account/range/type) does **not** read the indices, so echoing them back
  can't loop. Movers come straight from `trends.movers`.
- **What deliberately stayed on the client** (thin ratios over aggregates the backend already
  returns, i.e. the visualization layer): `windowShare`, `monthShare`, `perMonth`, `concentration`,
  the context peak/MoM, and the 3-mo moving-average chart guide.
- i18n: `changed-title` dropped "this month", added `compare-label`, `col-mom` → generic "change",
  `tip-mom` reworded, removed `vs-last-month`.
- Scoped out of v1: multi-month *period-vs-period* averaging (the "3 and 4 months" idea).

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

---

## Iteration 5 — visual identity: "ledger instrument"  `done` (Phase 1)

A design pass giving `/trends` a committed identity instead of a stack of equal grey cards.
Direction chosen from two mockups (v1 calm-analytic vs v2 "ledger instrument"); v2 won.

- **Signature — money speaks in monospace.** Every figure (KPI values, legend amounts, detail
  stats, mover deltas, month pickers, axis ticks) is `font-mono tabular-nums` — a ledger/terminal
  vernacular that also aligns numerals in columns. System mono stack, no webfont.
- **Chart as a live readout** (`TrendsChart.svelte`, rewritten): primary-blue line + gradient
  wash, x-grid removed / y-grid receded to a hairline, monospace ticks, a custom `trendsReadout`
  plugin that direct-labels the **endpoint value** and marks the **peak**, a `trendsCrosshair`
  plugin on hover, and a reduced-motion-aware draw-in. The hardcoded indigo `#6366f1` total colour
  is gone — the line now uses DaisyUI `primary` via `themeService.getThemeColors().seriesTotal`,
  and the Total legend pill uses the `primary` token so the two always match.
- **Colour system** unchanged semantically but now consistent: spend=`error`, income=`success`,
  blue=total. Grid colours softened in `themeService` for recessiveness.
- **KPI rail** (`TrendsKpis.svelte`): dropped the non-insight *categories count* tile for a
  **top-3 concentration** tile (thin client-side derivation, computed in `+page.svelte`); the two
  time-series tiles carry a subtle `CategorySparkline` of `totals[]`. New i18n keys
  `kpi-concentration` / `kpi-concentration-tip`.

- **Legible category lines regardless of picked colour.** User-chosen colours can sit almost on the
  surface (a pale line on white, a dark one on black) and vanish. Rather than alter the colour, the
  **real colour is kept** and a contrasting **casing** is drawn behind each category line — a soft
  canvas-shadow halo (`lineCasingPlugin` in `TrendsChart.svelte`), dark on the light theme, light on
  the dark theme, so it only "shows" where the line is close to the surface. The filled total line
  and the dashed smoothed guides are excluded. (Small legend/detail swatch dots still use the raw
  colour; add a hairline ring there if they read faint.)

This closes Future-work **#3 (hierarchy)** and **#10 (consistency pass)** below.

**Phase 2 — deferred, needs backend:** per-KPI **"vs previous period" deltas** (needs a prior-window
aggregate or a second fetch) and a **savings-rate hero tile** (needs a designated savings category
or a net-of-transfers figure). The mockups show these; they were intentionally not faked in v1.

---

# Future work — macro UI/UX  `todo`

A macro read of the whole `/trends` page (after iterations 1–3). The page is now a vertical stack
of **equal-weight** cards (KPIs → chart → what-changed) scoped to **one account**. The items below
are ordered by impact/effort. None are started.

## Not a net KPI — savings rate already reads off the category chart  `rejected`

The earlier "Why % of income, not net" section already settled this: this user ends every month
at ~zero balance because **savings is a transfer out** (a Debit carrying `transfer_group_id`), so
`net = window_income − window_total ≈ 0` and is worse than useless. The savings rate is already
visible as the **savings category's % of income** on the existing chart. A "Saved this period"
tile would be redundant (Option A) or would require redefining spend as net-of-transfers
everywhere (Option B) — not worth it. **Decision: do not add a net/savings KPI.** Leave the
category chart as the savings read.

## Tier 1 — structure & the savings story

1. **~~"Saved this period" KPI~~** `rejected` — see above. Net ≈ 0 for a transfer-based savings
   flow; savings rate is already the savings category's % of income on the chart.
2. **Cash-flow view (rethink the Spending/Income toggle).** `todo` · **impact: high · effort:
   high.** Replace either/or with income-vs-spending per month (paired bars) + a **net line**, so
   the whole story reads at once. The transformative version of #1; treat as its own effort after
   #1 lands. Touches `TrendsChart`, `TrendsControls`, and likely the store (return both series in
   one response instead of one-type-per-fetch).
3. **Visual hierarchy / hero.** `todo` · **impact: medium · effort: low-med.** Everything uses the
   same `rounded-xl border shadow-sm`; nothing leads. Promote the savings/cash-flow number to a
   hero tile; demote the rest to supporting. Makes it read as a dashboard, not a list.
4. **"All accounts" aggregate.** `todo` · **impact: medium · effort: med.** `AccountsSplitLayout`
   scopes trends to one account; a true macro view wants the whole picture. Add an "All accounts"
   option (sum across the user's accounts). On a stats page, also consider a slim account
   **dropdown** instead of the full split panel to give charts more width.

## Tier 2 — turn insight into action

5. **Drill-through to transactions.** `todo` · **impact: high · effort: med.** Nothing on the page
   is clickable — it's a dead end. Clicking a category (legend pill / mover row) or a chart month
   should open the underlying transactions ("dining +32%" → *which ones*). The missing verb of the
   whole page.
6. **AI "insight of the period".** `todo` · **impact: high · effort: med.** Reuse the existing
   OpenAI integration (`backend/service/openai/`, prompts in `backend/prompts/`) to surface one
   scoped sentence ("Dining is your biggest change, +150€ vs last month"). Feed it the aggregates
   we already compute (movers, concentration, savings). Turns a chart into guidance; differentiator.
7. **Period-over-period on the KPIs.** `todo` · **impact: medium · effort: low-med.** KPIs show
   absolutes with no baseline. Add "▲8% vs previous 12M" to each headline (same MoM idea as
   `computeCategoryMovers`, applied to the window — needs the store to also compute the prior
   window, or a second fetch).

## Tier 3 — polish

8. **Budgets/targets overlay.** `todo` · Categories already carry `budget` and
   `GetTransactionStatistics` computes `budget_percentage` (see `docs/feature-plans.md`). Overlay a
   target line / "30% over dining" on the chart once budgets are surfaced.
9. **Chart affordances.** `todo` · Click a point to focus that month; annotate the peak; skeleton
   loaders instead of the bare spinner in `+page.svelte`.
10. **Consistency pass.** `todo` · `tabular-nums` on every figure; one red=spend / green=income rule
    everywhere; unify the "% / MoM / concentration" family so the metrics read as one system.

## Suggested order

**#5 (drill-through) → #6 (AI insight)** — best impact/effort, and each proves appetite before the
larger **#2 (cash-flow redesign)**. (#1 was dropped — the savings rate already reads off the
category chart, so there's no net-KPI seed to build on.)
