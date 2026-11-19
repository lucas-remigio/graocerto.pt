# Auth & Sharing Plan: trusted devices, TOTP, shared accounts

Each feature goes through: detailed plan → build → tests → manual check → prod. Features 1 and
2 both touch `handleLogin`, so 1 lands first and 2 builds on it.

Status legend: `todo` / `in progress` / `done` / `parked`.

**Scope (2026-11-19):** only features 1 and 2 for now. Feature 3 is parked and kept below so
the analysis isn't lost.

| # | Feature                                | Effort | Status |
| - | -------------------------------------- | ------ | ------ |
| 1 | Remember this device (90d) + device list | S–M  | done on `main` (`5e65106`), not deployed |
| 2 | TOTP 2FA (Ente Auth, Aegis, Google…)   | M      | in progress (built + tested, uncommitted) |
| 3 | Shared accounts (invite + co-edit)     | L      | parked |

Pre-req: the `telegram_bot` branch has uncommitted work. Land or commit it before starting
here, then do each feature on its own branch off `main`.

Reference anchors (verified 2026-11-19):

- Login: `backend/service/user/routes.go:71` `handleLogin`. MFA branch at `:96-117`. TOTP
  currently returns `501` at `:106`.
- OTP verify: `backend/service/user/auth_flow.go:90` `handleVerifyLoginOTP` → `issueAuthToken`.
  Challenge creation `auth_flow.go:208`.
- Session cookie: `routes.go:226` `issueAuthToken` (JWT, 7d default, `JWT_EXPIRATION_IN_SECONDS`).
- Token store: `types.AuthTokenStore` in `backend/types/auth_types.go`, impl
  `backend/service/user/auth_store.go`, table `auth_tokens` (migration `20260411120000`).
- Secret helpers: `backend/service/auth/one_time.go` (`GenerateOneTimeToken`, `HashSecret`,
  `CompareSecret`).
- MFA enum: `types.MfaMethod` = `email_otp | totp | off`, column `users.mfa_method`.
- Google login (`routes.go:120`) skips MFA entirely. Google is the second factor there, and
  that stays as is.
- Frontend login pages: `src/routes/login`, `src/routes/login-otp`. No settings page yet.

---

## 1. Remember this device: skip 2FA for 90 days, with a device list  `in progress`

**Progress (2026-11-19):** built and tested on the working tree off `main` (not committed).
- Done: migration `20261119120000_add-trusted-devices`, `service/user/trusted_device_store.go`,
  `service/user/trusted_devices.go` (cookie, label, list/revoke handlers), login skip, remember
  on OTP verify, revoke on password reset, `Secure` cookie fix via `X-Forwarded-Proto`.
  Frontend: checkbox on `login-otp`, `lib/TrustedDevices.svelte` inside `ProfileModal`.
- Verified: unit tests (`trusted_devices_test.go`); store SQL + migration up/down/up against a
  scratch Postgres 16; full curl E2E (remember → skip → wrong password still 404 → list
  `is_current` → revoke → challenge again → `Secure` behind proxy → CORS DELETE preflight).
- Left: Lucas eyeballs the UI locally → commit on a branch → **resolve the migration ordering
  with `telegram_bot` (see below)** → deploy.
- Migration ordering: `telegram_bot` has `20260806120000_add-telegram-integration`, which is
  older than `20261119120000`. golang-migrate only applies versions above the current one,
  so if this ships first, the telegram migration would be **silently skipped** in prod.
  **Decided (2026-11-19): Telegram stays a separate branch, not shipping now.** So before
  `telegram_bot` is ever merged, rename its migration to a timestamp newer than the latest on
  `main` (e.g. `2026112xxxxxxx_add-telegram-integration`). Its SQL is all `IF NOT EXISTS`, so
  it re-runs safely on DBs that already have it.
- Local dev DB (2026-11-19): it had run the telegram migration, so `main` failed to start
  ("no migration found for version 20260806120000"). Fixed by setting
  `schema_migrations.version` back to `20260729130000`. The telegram column/table remain
  locally, unused. Same fix applies whenever switching from `telegram_bot` back to `main`.

**Goal:** after a successful 2FA, the browser can be remembered for 90 days. Later logins with
email + password on that browser skip the 2FA step. The password is **still required**. Only
the second factor is skipped. A Settings page lists remembered devices and can revoke each one
or all of them.

### Approach (decided): dedicated `trusted_devices` table
- Migration: `trusted_devices (id, user_id FK ON DELETE CASCADE, token_hash UNIQUE, label TEXT,
  created_at, last_used_at, expires_at)`. `label` comes from the User-Agent, e.g.
  "Chrome on macOS". Parse it with a few `strings.Contains` checks, no UA library.
- On successful 2FA verify (email OTP now, TOTP later) **and** a "Remember this device for 90
  days" checkbox ticked, generate `GenerateOneTimeToken()` and store `HashSecret(token)` with
  `expires_at = now + 90d`. Set an `HttpOnly`, `Secure`, `SameSite=Strict` cookie
  `trustedDevice` with the raw token, `MaxAge=90d`.
- In `handleLogin`, after the password check and before creating a challenge: read the cookie
  and look up the hash. If it exists, is not expired and `device.UserID == user.ID`, bump
  `last_used_at` and call `issueAuthToken` directly. Otherwise run the normal 2FA flow.
- Endpoints (authenticated): `GET /auth/devices` (list, flags the current one),
  `DELETE /auth/devices/{id}` (revoke one, scoped to the user), `DELETE /auth/devices`
  (revoke all).
- Revoke all on password reset and on MFA method change.
- Fixed expiry, no sliding renewal (KISS).

### Alternatives considered
- **Reuse `auth_tokens` with purpose `trusted_device`**: no migration, but no label and no
  last-used time, so the device list can't be built. Rejected after Lucas asked for the list.
- **Just make the JWT last 90 days.** Rejected: a stolen session cookie would then work for
  90 days. The trusted-device token only skips 2FA and still needs the password.

### Watch-outs
- `issueAuthToken` sets `Secure: r.TLS != nil`. Behind the prod reverse proxy, `r.TLS` is nil,
  so **the cookie is likely not `Secure` in prod today**. Fix this for both cookies (check
  `X-Forwarded-Proto` or use a config flag). Verify the proxy setup first.
- Cookie path: the frontend calls the backend through a proxy. Confirm the cookie reaches
  `/login` (same origin in prod, ports differ in dev).

### Files
- `backend/cmd/migrate/migrations/<ts>_add-trusted-devices.{up,down}.sql`
- `backend/types/auth_types.go`: `TrustedDevice` struct + `TrustedDeviceStore` interface. Add
  `RememberDevice bool` to `VerifyLoginOTPPayload`.
- `backend/service/user/trusted_device_store.go`, plus device list/revoke handlers.
- `backend/service/user/routes.go`: trusted-device check in `handleLogin`. Fix `Secure`.
- `backend/service/user/auth_flow.go`: issue the cookie on verify. Revoke on password reset.
- `frontend/src/routes/login-otp/+page.svelte`: the "Remember this device" checkbox.
- `frontend/src/lib/TrustedDevices.svelte`, rendered in `ProfileModal.svelte` (see decisions).
  Feature 2 adds the TOTP section next to it.

### Tests
- Verify with `remember_device=true` sets the cookie and stores a hashed token.
- Login with a valid cookie skips the challenge (200 + JWT, not 202).
- Expired cookie, another user's cookie and an unknown cookie all fall back to the challenge.
- Wrong password with a valid cookie still gets 404 (the cookie never bypasses the password).
- Password reset revokes trusted devices.
- The list shows only the caller's devices. Revoking another user's device id returns 404.

---

## 2. TOTP 2FA (authenticator apps)  `in progress`

**Opt-in (confirmed by Lucas 2026-11-19):** email codes stay the default; each user chooses the
app in the Profile modal and can switch back. With the app active there is no email fallback,
recovery codes cover a lost phone.

**Progress (2026-11-19):** built and tested on the working tree on top of `5e65106`.
- Backend: `service/auth/totp.go` (RFC 6238, stdlib), `service/auth/encrypt.go` (AES-256-GCM),
  migration `20261119130000_add-users-totp`, `service/user/totp_store.go`, `service/user/totp.go`
  (status/setup/enable/disable + login challenge + recovery codes in `auth_tokens`), login and
  `/auth/verify-login-otp` now dispatch on the challenge purpose; challenge responses carry
  `method`. `GenerateReadableCode` copied verbatim from `telegram_bot` so that merge is clean.
- Frontend: `lib/TwoFactorSettings.svelte` in `ProfileModal` (QR via `qrcode@1.5.4`), `login-otp`
  shows app/email wording + "use a recovery code" toggle, shared `LoginOtpState` in `types.ts`.
- Config: `TOTP_ENC_KEY` in `config/env.go`, both compose files; dev keys added to local `.env`s.
- Verified: RFC vectors + skew/replay unit tests, 11 handler tests, store + migration up/down on
  scratch Postgres (down moves `totp` users back to `email_otp`), full curl E2E incl. replay and
  recovery codes.
- Prod key: Lucas added the `TOTP_ENC_KEY` GitHub secret; `deploy.yaml` now exports it (every
  `${VAR}` in `docker-compose.prod.yaml` is now exported by the workflow).
- Dev gotcha: `@types/qrcode` pulls `@types/node`, which made pnpm keep two vite copies and broke
  `pnpm dev` ("Cannot find module '$app/environment'"). Fixed with `pnpm prune` + clearing
  `node_modules/.vite` and `.svelte-kit`.
- Left: Lucas tests with a real authenticator app → commit → deploy. Key notes:
  (`openssl rand -base64 32`, separate from dev; never rotate without a migration plan, or
  every TOTP user loses access) → deploy.

**Goal:** a user can switch 2FA from email codes to an authenticator app (Ente Auth, Aegis,
Google Authenticator, 1Password…). Login then asks for the 6-digit app code instead of
emailing one.

### Approach (recommended): RFC 6238 in stdlib, no new Go dependency
TOTP is about 40 lines: HMAC-SHA1 over the 30s time step, dynamic truncation, 6 digits. It needs
only `crypto/hmac`, `crypto/sha1`, `encoding/base32` and `encoding/binary`. Put it in
`backend/service/auth/totp.go` with the RFC 6238 test vectors.

- **Enrol** (Settings page):
  1. `POST /auth/totp/setup` generates a 20-byte secret, stores it as *pending*, and returns
     the `otpauth://totp/Grão Certo:<email>?secret=…&issuer=Grão Certo` URI + base32 secret.
  2. The user scans the QR or types the secret, then `POST /auth/totp/enable {code}` checks the
     code against the pending secret, sets `mfa_method='totp'`, returns **recovery codes**
     (shown once) and revokes trusted devices.
- **Login:** replace the `501` at `routes.go:106` with a challenge (reuse `auth_tokens`
  purpose `login_totp`, so attempts and expiry come for free). The frontend `login-otp` page
  shows "Enter the code from your authenticator app". Verify accepts ±1 step of clock skew and
  rejects reuse of the last accepted step (replay).
- **Disable:** `POST /auth/totp/disable {code}` sets `mfa_method` back to `email_otp`.
- **Recovery:** 8 single-use recovery codes, stored hashed. They are accepted in place of a
  TOTP code. **Email OTP is not the fallback**: if the inbox were enough, TOTP would add
  nothing.
- Storage: migration adds `users.totp_secret TEXT NULL`, `users.totp_pending_secret TEXT NULL`
  and `users.totp_last_step BIGINT NULL`. Encrypt the secret at rest with AES-GCM (stdlib)
  using a new `TOTP_ENC_KEY` env var. Unlike passwords, it can't be hashed because the server
  needs the raw value to verify codes.
- **QR (decided):** `qrcode` npm package (soldair/node-qrcode, MIT, the de facto standard).
  Render client-side. The secret never goes to a third-party QR service.

### Alternatives considered
- **`github.com/pquerna/otp`**: well known, but adds a dependency for ~40 lines of fully
  specified code. Not worth it.
- **WebAuthn / passkeys**: stronger and phishing-resistant, but much more work. A good
  follow-up, out of scope here.

### Files
- `backend/service/auth/totp.go` + `totp_test.go` (RFC vectors, skew, replay).
- `backend/cmd/migrate/migrations/<ts>_add-users-totp.{up,down}.sql`
- `backend/config/env.go`: `TOTP_ENC_KEY`. Add it to the prod env and `docker-compose.prod.yaml`.
- `backend/types/auth_types.go`: payloads, purposes, store methods.
- `backend/service/user/`: setup/enable/disable handlers, login branch, recovery codes.
- `frontend/src/lib/ProfileModal.svelte`: TOTP section (enrol with QR, disable). Move to a
  settings page if the modal gets crowded.
- `frontend/src/routes/login-otp/`: TOTP + recovery-code mode.

### Tests
- TOTP generation matches the RFC 6238 vectors. ±1 step accepted, ±2 rejected, same step
  replay rejected.
- Enable with a wrong code doesn't switch the method. Enable with the right code switches it
  and returns recovery codes.
- Login as a TOTP user creates a challenge and no email is sent. Too many wrong codes lock the
  challenge (existing `max_attempts`).
- A recovery code works once.
- A trusted device (feature 1) still skips TOTP.

---

## 3. Shared accounts  `parked`

**Goal:** the owner of an in-app account (e.g. "Vacation") invites another user by email. Both
then see the account and can add, edit and delete its transactions. Other accounts stay private.

Notes from the 2026-11-19 analysis:
- Ownership is hard-coded as `a.user_id = $N` in about 50 places (`service/account/store.go`
  14, `service/recurring_rule/store.go` 28, `service/transaction/store.go` 4, plus the telegram
  and notification stores). Plan: `account_members (account_id, user_id, role,
  order_index, is_favorite)`, backfill owners, and replace every check with one shared
  membership SQL fragment. Every changed query needs a test proving a non-member gets nothing.
  `order_index`/`is_favorite` move to the membership row because they are per user.
- Invites: `account_invites` table + email via the existing `mailer`, accepted at
  `/invites/<token>`.
- **Categories are per user.** Proposed: shared accounts use the owner's categories.
  Alternatives: per-account categories (big refactor) or each user's own (breaks stats).
- Deleting the owner's user cascades and deletes the shared account for the member as well.
  Either warn, or transfer ownership.
- Also: `transactions.created_by`, transfers across shared/personal accounts, recurring rules
  and notifications for both members, Telegram bot, export scope, realtime socket updates.
- Slicing: (1) membership access, no UI change, (2) invites, (3) shared UX, (4) realtime.

---

## Decisions log
- 2026-11-19: Order 1 → 2. 1 and 2 share login code, so 1 lands first.
- 2026-11-19: Trusted devices get a dedicated `trusted_devices` table (not `auth_tokens`)
  because Lucas wants a device list. 90-day fixed expiry.
- 2026-11-19: TOTP implemented with the stdlib, no `pquerna/otp`.
- 2026-11-19: `qrcode` npm package approved for the TOTP enrolment QR.
- 2026-11-19: Shared accounts (feature 3) parked. Only 2FA work for now.
- 2026-11-19: Trusted-device UI lives in the existing `ProfileModal` instead of a new Settings
  route + sidebar entry. Less code and it's where account settings already are. Revisit if
  the modal gets crowded once TOTP is added.
- 2026-11-19: `TrustedDeviceStore` is its own interface (implemented by the user `Store`)
  instead of growing `UserStore`, so the existing `UserStore` test fakes stay untouched.
- 2026-11-19: TOTP secret key is a dedicated `TOTP_ENC_KEY` (not derived from `JWT_SECRET`), so
  rotating the JWT secret never locks TOTP users out.
- 2026-11-19: TOTP login reuses `/auth/verify-login-otp` and `auth_tokens` (purpose `login_totp`)
  for expiry + attempt limits; recovery codes are `auth_tokens` rows (purpose `totp_recovery`,
  max 1 use) rather than a new table. A `TOTPStore` interface keeps `UserStore` fakes untouched.
- 2026-11-19: "Forget device" only stops skipping 2FA. It doesn't end the session (the JWT
  stays valid up to 7d). Real session revocation would need server-side sessions, out of scope.

## Open questions
- (Feature 3, when unparked) Shared-account categories: the owner's? Owner deletes their user:
  delete the shared account or transfer it?
