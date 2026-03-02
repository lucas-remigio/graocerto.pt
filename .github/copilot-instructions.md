# Copilot Instructions – Grão Certo (Wallet Tracker)

## Architecture Overview

Three independent services orchestrated via Docker Compose:

| Service      | Stack                           | Port            |
| ------------ | ------------------------------- | --------------- |
| `app`        | Go 1.23 + PostgreSQL            | `$BACKEND_PORT` |
| `frontend`   | SvelteKit + TailwindCSS/DaisyUI | 3000            |
| `websockets` | Node.js + `ws`                  | `$SOCKETS_PORT` |

The backend is the Go module `github.com/lucas-remigio/wallet-tracker`. All shared types and store **interfaces** live in `backend/types/` — not inside service packages.

## Backend Patterns

### Service structure

Every service follows the same 2-file pattern:

- `service/<name>/routes.go` — `Handler` struct, `NewHandler(store)`, `RegisterRoutes(*http.ServeMux)`
- `service/<name>/store.go` — `Store` struct, implements the interface declared in `types/`

When adding a new service, register its store and handler in `cmd/api/api.go`.

### Database helpers (use these, don't write raw loops)

```go
// returns []*T, always a slice even when empty
db.QueryList(s.db, query, scanRowsIntoFoo, args...)
// returns *T via QueryRow
db.QuerySingle(s.db, query, scanRowIntoFoo, args...)
```

PostgreSQL positional params: `$1`, `$2`, … (not `?`).

### Auth in handlers

```go
// Wrap the handler at registration time
router.HandleFunc("/foo", middleware.AuthMiddleware(h.MyHandler))

// Inside the handler, extract user ID
userId, ok := middleware.RequireAuth(w, r)
if !ok { return }
```

### HTTP response helpers

```go
middleware.WriteDataResponse(w, payload)   // 200 + JSON body
middleware.WriteSuccessResponse(w)         // 200 + {"status":"success"}
utils.WriteError(w, http.StatusBadRequest, err)
```

### Validation

Add `validate:"..."` tags (go-playground/validator) to payload structs in `types/`, then call:

```go
if !middleware.ValidatePayloadAndRespond(w, r, &payload) { return }
```

### Migrations

```bash
make migration <name>   # creates up/down SQL files in cmd/migrate/migrations/
make migrate-up         # apply
make migrate-down       # rollback one
```

## Developer Workflow

```bash
# Backend
make run          # build + run binary
make test         # go test -v ./...

# Full stack (dev)
docker compose up

# Frontend (standalone)
cd frontend && pnpm dev
```

Configuration is loaded from `.env` via `godotenv` — see `config/env.go` for all keys (`JWT_SECRET`, `DATABASE_URL`, `OPENAI_API_KEY`, `FRONTEND_URL`, etc.).

## Frontend Patterns

- All API calls go through `$lib/axios.ts` — a pre-configured Axios instance that attaches the JWT `Bearer` token from the `auth` Svelte store and redirects to `/login` on 401.
- In **production**, the frontend calls relative `/api/v1/...` (Nginx proxies to the backend). In dev, it calls `http://$BACKEND_URL:$BACKEND_PORT/api/v1`.
- Server-side auth guard lives in `src/hooks.server.ts`; public routes are explicitly whitelisted (`/login`, `/register`).
- Shared types: `$lib/types.ts`. Svelte stores: `$lib/stores/`. Per-entity API service files: `$lib/services/`.

## WebSocket Service

Room-based pub/sub keyed by user **email** (`sockets/src/index.js`). Clients join a room by sending `{ type: "join_room", email }`. The backend does **not** directly communicate with the socket server; the frontend manages its own WebSocket connection to the `websockets` service.

## Key Files for Reference

- `backend/cmd/api/api.go` — service wiring / dependency injection root
- `backend/db/utils.go` — generic query helpers
- `backend/middleware/auth.go` — auth middleware + response helpers
- `backend/types/` — all interfaces and payload/response structs
- `frontend/src/lib/axios.ts` — API client configuration
- `frontend/src/hooks.server.ts` — SSR route guard

## Code Quality Standards

### Before writing any code

1. **Search first**: Check the workspace for existing functions/components that can be reused or slightly extended. State explicitly what you found (or didn't find) before proceeding.
2. **Plan briefly**: Describe your approach and any meaningful trade-offs in 2–3 sentences.
3. **Clarify ambiguity**: Ask 1–3 targeted questions if requirements are unclear before generating code.

### Implementation principles

- **KISS**: Choose the most boring, standard, readable implementation. Avoid clever abstractions.
- **DRY**: Never duplicate logic. Extend an existing helper rather than create a near-identical one.
- **SRP**: Functions do one thing and stay under ~20 lines. Each file/module has a single reason to change.
- **Guard clauses**: Use early returns to reduce nesting — see `RequireAuth` in `middleware/auth.go` as the project model.
- **Self-documenting names**: Names should make comments about _what_ redundant. Comments explain _why_.
- **No speculative code**: Only implement what is explicitly required.

### Testing

Always include unit test cases for core logic. Backend tests use the standard `go test` harness (see `make test`). Place test files alongside the code they test (`_test.go` suffix).
