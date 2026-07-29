# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Read this first

`.github/copilot-instructions.md` is the **canonical reference** for backend/frontend patterns
(service structure, DB helpers, auth, response helpers, validation, WebSocket, key files).
Read it before writing code — this file does not repeat it. `AGENTS.md` holds the terse code
rules. Don't duplicate what those already say.

## How to work here (non-negotiable)

- **Search before you write.** Explicitly state which existing function/component/helper you
  found and are reusing or extending. Prefer extending an existing helper over adding a
  near-identical one. Shared store **interfaces** and payload/response structs live in
  `backend/types/` — check there before defining new ones.
- **Plan first.** Describe the approach and one meaningful trade-off in a few sentences before
  coding. Ask 1–3 clarifying questions when the request is ambiguous.
- **KISS / SRP / guard clauses.** Boring readable code; functions do one thing; early returns
  over nesting (model: `RequireAuth` in `backend/middleware/auth.go`). Comments explain *why*.
- **Strict types.** Go: avoid `interface{}`, wrap errors with `%w`, pass `context.Context`,
  log with `log/slog`. TS: no `any`, type module boundaries / API responses / stores.
- **Tests are part of the feature.** Add `_test.go` next to the code (success, validation
  failure, edge cases). For a bug fix, write the reproducing test first.
- **No speculative code / no new deps.** Only what's asked; justify any new dependency before
  adding it.

## Architecture (big picture)

Three independent services, orchestrated by Docker Compose. Go module:
`github.com/lucas-remigio/wallet-tracker`.

| Service      | Path         | Stack                            | Dev port |
| ------------ | ------------ | -------------------------------- | -------- |
| `app`        | `backend/`   | Go 1.26 + PostgreSQL             | 3001     |
| `frontend`   | `frontend/`  | SvelteKit + Tailwind/DaisyUI     | 3000     |
| `websockets` | `sockets/`   | Node.js + `ws` (optional)        | 3002     |

- **Backend** is layered API → Service → Store. Each service in `backend/service/<name>/` is a
  `routes.go` (`Handler`, `NewHandler`, `RegisterRoutes`) + `store.go` (implements the
  interface from `backend/types/`). Wire new services in `backend/cmd/api/api.go` (DI root).
  Current services: account, auth, category, investment_calculator, mailer, notification,
  openai, recurring_rule, transaction, transaction_types, user.
- **Frontend** talks to the backend only through `frontend/src/lib/axios.ts` (attaches JWT,
  redirects to `/login` on 401). SSR route guard: `frontend/src/hooks.server.ts` (public routes
  whitelisted). Protected pages live under `src/routes/(protected)/`. Per-entity API calls in
  `src/lib/services/`, reactive state in `src/lib/stores/`, shared types in `src/lib/types.ts`.
- **WebSockets** is room-based pub/sub keyed by user **email**; the frontend owns the socket
  connection, the backend does not talk to it directly.
- **OpenAI** integration lives in `backend/service/openai/` with prompts in `backend/prompts/`.

## Commands

Backend (`backend/`, via `make`):

```bash
make run            # build + run the binary (bin/backend)
make test           # go test -v ./...
go test -v -run TestName ./cmd/api/...   # single test
make migration <name>   # scaffold up/down SQL in cmd/migrate/migrations/
make migrate-up         # apply migrations
make migrate-down       # roll back one
```

PostgreSQL uses positional params `$1, $2, …` (not `?`). Use `db.QueryList` / `db.QuerySingle`
(`backend/db/utils.go`) instead of hand-rolled row loops.

Frontend (`frontend/`, pnpm):

```bash
pnpm dev            # vite dev server
pnpm build          # production build
pnpm check          # svelte-check (type check)
pnpm lint           # eslint + prettier --check
pnpm format         # prettier --write
```

Full stack:

```bash
docker compose up --build                          # app + frontend
docker compose --profile websockets up --build     # + websockets
```

## Config

Backend config is loaded from `.env` via `godotenv`; all keys are in `backend/config/env.go`
(`JWT_SECRET`, `DATABASE_URL`, `OPENAI_API_KEY`, `FRONTEND_URL`, ports, …). Never hardcode
secrets.
