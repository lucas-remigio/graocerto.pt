# Wallet Tracker Code Rules

## Core Standards

- Follow SOLID principles and keep each module focused on one responsibility.
- Prefer the simplest readable design over clever abstractions.
- Keep functions small, explicit, and easy to test.
- Use descriptive names for files, variables, functions, and types.

## TypeScript Rules

- Use strict TypeScript typing everywhere.
- Do not use `any`.
- Prefer exact types, narrow unions, and reusable interfaces over inferred ambiguity.
- Add types at module boundaries, API responses, stores, and shared utilities.

## Architecture Rules

- Keep frontend, backend, sockets, and shared utilities modular.
- Do not mix unrelated concerns in the same file or component.
- Reuse existing helpers before introducing new ones.
- Avoid hidden side effects and implicit coupling.

### Frontend is the visualization layer

- **Business logic, calculations, and aggregations live on the backend.** If a value can be
  computed server-side, compute it there and send the ready-to-render result. Never derive money
  figures (totals, rates, percentages, projections) with client-side math that the backend could
  own — the backend is the single source of truth and the only place that math is unit-tested.
- **The frontend renders; it does not decide.** A component's job is presentation (formatting,
  i18n, layout) and **interaction state only** (which series/tab is selected, drill-down open,
  optimistic UI). Keep components thin — the page is an orchestrator, not a calculator.
- This is separation of concerns, not "less code at all costs": interaction state genuinely
  belongs on the client (e.g. legend toggles in `TrendsChartCard.svelte`). The line is
  *responsibility* — computation on the backend, presentation on the frontend.
- Never trust the client for correctness or security: even when the UI validates for UX, the
  backend re-validates as the authority.

## Testing Rules

- Every new feature must include unit tests.
- Tests should cover success paths, validation failures, and edge cases.
- Keep test setup minimal and close to the behavior being verified.
- Update or add tests before considering the feature complete.

