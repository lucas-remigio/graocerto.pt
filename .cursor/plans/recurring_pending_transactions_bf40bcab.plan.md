---
name: Recurring Pending Transactions
overview: Implement recurring payments with pending transaction approvals in a single transactions table, ensuring pending rows are financially inert until approved. Add backend scheduling, approval/rejection APIs, and frontend management screens with minimal disruption to existing flows.
todos:
  - id: db-schema
    content: Add migrations for transactions pending support and recurring_rules table with indexes and constraints
    status: completed
  - id: recurring-backend
    content: Implement recurring rules CRUD and materialization logic with idempotency and ownership checks
    status: completed
  - id: pending-actions
    content: Add approve/reject pending transaction endpoints with atomic balance-safe behavior
    status: completed
  - id: query-updates
    content: Update list/statistics/month queries so pending is visible where needed but excluded from financial aggregates
    status: completed
  - id: scheduler
    content: Wire periodic backend job to generate pending transactions from due recurring rules
    status: completed
  - id: frontend-recurring
    content: Build recurring payments screen for rule management in protected routes
    status: completed
  - id: frontend-pending
    content: Add pending badges and approve/reject actions in TransactionsTable and wire dataService APIs
    status: completed
  - id: tests
    content: Add unit and integration tests for recurrence, idempotency, approve/reject, and aggregate correctness
    status: completed
isProject: false
---

# Recurring Payments With Pending Approval

## Architecture Decision

- Keep a single `transactions` table and add `is_pending` boolean.
- Rejected recurring transactions are hard-deleted.
- Pending transactions are visible in transaction UI, but **must not** affect account balances/statistics until approved.
- Recurrence definitions live in a dedicated table (`recurring_rules`) and daily worker materializes pending transactions.

## Reuse First (DRY)

- Reuse existing transaction creation and transfer logic from [`/Users/lucasremigio/Developer/vps-projects/wallet-tracker/backend/service/transaction/store.go`](/Users/lucasremigio/Developer/vps-projects/wallet-tracker/backend/service/transaction/store.go) for approved/finalized writes.
- Reuse existing transaction DTO/list/statistics endpoints from [`/Users/lucasremigio/Developer/vps-projects/wallet-tracker/backend/service/transaction/routes.go`](/Users/lucasremigio/Developer/vps-projects/wallet-tracker/backend/service/transaction/routes.go), extending behavior with `is_pending` filters/flags.
- Reuse frontend transaction table patterns from [`/Users/lucasremigio/Developer/vps-projects/wallet-tracker/frontend/src/components/TransactionsTable.svelte`](/Users/lucasremigio/Developer/vps-projects/wallet-tracker/frontend/src/components/TransactionsTable.svelte).

## Backend Changes

- Add migration(s):
  - `transactions.is_pending BOOLEAN NOT NULL DEFAULT FALSE`
  - `transactions.recurring_rule_id INT NULL` (FK to recurring rules)
  - optional helper index for pending lookup by account/date.
- Create new recurring domain:
  - `recurring_rules` table with: user/account/category/type, amount, description, frequency (`daily|weekly|monthly|every_x_days`), interval, next_run_date, active.
  - idempotency field(s) to prevent duplicate daily generation (e.g., unique rule/date materialization key).
- Implement recurring service/store/routes:
  - CRUD for recurring rules.
  - materialization method that inserts pending transactions only (no balance/account mutation).
- Add approval endpoints in transaction routes:
  - approve pending transaction: set `is_pending=false` and apply balance math atomically.
  - reject pending transaction: delete transaction.
- Ensure transaction list endpoint can include pending rows for UI (default include both, with optional filter).
- Ensure statistics/month aggregation endpoints ignore pending rows (`WHERE is_pending = false`).
- Add ownership/security checks for rule CRUD and approval/rejection paths.

## Scheduler / Job

- Add worker entrypoint in backend runtime that periodically scans due recurring rules.
- For each due rule:
  - create pending transaction(s) (single or transfer pair if transfer rule),
  - advance `next_run_date`,
  - do all in transaction with idempotency guard.
- Keep batch processing small and indexed for efficiency.

## Frontend Changes

- Add new protected route for recurring payments management:
  - list rules, create/edit/delete rule, active toggle.
- Extend transaction table UI to visually mark pending rows and offer actions:
  - `Approve` and `Reject` buttons for `is_pending=true`.
- Extend `dataService` in [`/Users/lucasremigio/Developer/vps-projects/wallet-tracker/frontend/src/lib/services/dataService.ts`](/Users/lucasremigio/Developer/vps-projects/wallet-tracker/frontend/src/lib/services/dataService.ts):
  - recurring rule APIs,
  - approve/reject pending APIs,
  - cache invalidation for impacted account data.
- Keep existing transaction flow unchanged for normal manual create/edit/delete.

## Correctness Guarantees (Fail-Safe)

- Pending creation path never updates account balances.
- Approval path is atomic: validate ownership + ensure pending + update account balance + flip `is_pending=false` in one DB transaction.
- Job idempotency prevents duplicate pending rows for the same rule/date.
- Delete/reject remains consistent with existing delete behavior.

## Test Plan

- Unit tests:
  - recurrence next-date calculation for all frequencies,
  - approve/reject logic,
  - idempotent materialization.
- Integration tests:
  - pending rows excluded from statistics/totals,
  - pending rows visible in transaction list,
  - approval updates account balance exactly once,
  - reject deletes without balance changes.
- API contract tests for new endpoints and validation errors.

## Rollout Strategy

- Deploy migration first (backward-compatible default values).
- Deploy backend routes/services.
- Deploy frontend recurring screen + pending actions.
- Monitor pending counts and approval success/failure rates after release.

## Residual Risks (and Mitigation)

- Timezone/date drift on recurrence: store rule timezone and compute dates consistently server-side.
- Concurrent approval/job races: protect with row locking or conditional updates (`is_pending=true` precondition).
- Transfer recurrence complexity: start V1 with non-transfer recurring, then enable transfer recurring in V1.1 if needed.
