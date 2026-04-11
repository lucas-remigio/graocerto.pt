---
name: Email Auth Hardening
overview: Add SMTP-backed email verification, email OTP login, and password reset flows to wallet-tracker while keeping Google sign-in and the existing JWT session model intact.
todos:
  - id: config-mailer
    content: Add SMTP/config wiring and a reusable mailer interface for backend auth flows
    status: pending
  - id: schema-auth
    content: Add auth-related user fields and token tables for verification, otp, and reset flows
    status: pending
  - id: backend-auth
    content: Implement register, verify-email, login-otp, resend, forgot-password, and reset-password handlers
    status: pending
  - id: frontend-auth
    content: Add login/register/reset UI states and optional MFA selection for future TOTP support
    status: pending
  - id: tests
    content: Add unit tests for token generation, expiry, reuse prevention, and handler behavior
    status: pending
isProject: false
---

# Email Auth Hardening

## Summary

- Reuse the current auth stack in [`backend/service/user/routes.go`](/Users/lucasremigio/Developer/vps-projects/wallet-tracker/backend/service/user/routes.go), [`backend/service/auth/jwt.go`](/Users/lucasremigio/Developer/vps-projects/wallet-tracker/backend/service/auth/jwt.go), and [`backend/service/auth/password.go`](/Users/lucasremigio/Developer/vps-projects/wallet-tracker/backend/service/auth/password.go) to avoid rewriting core login/session logic.
- Add a dedicated mailer + one-time-token layer for verification, login OTP, and password reset.
- Use email OTP as the default first-phase second step, keep TOTP as a later optional factor, grandfather existing users as verified, and auto-verify Google users.

## Key Changes

- Add SMTP settings to backend config and compose files so the app can send mail through the existing server.
- Introduce a mailer interface with a concrete SMTP implementation and small auth-specific templates.
- Extend `users` with verification/MFA state, and add a token table for:
  - email verification
  - login OTP
  - password reset
- Split auth into clear flows:
  - register -> create user -> send verification email
  - login -> validate password -> send OTP challenge -> issue JWT after OTP success
  - forgot password -> send reset link -> set new password after token validation
- Keep error messages generic and add cooldown/rate-limit behavior to prevent abuse and account enumeration.
- Frontend changes:
  - add OTP challenge screen after password login
  - add email verification notice after register
  - add forgot/reset password pages
  - keep a place in the UI for future TOTP selection without enabling it yet

## Test Plan

- Unit tests for token creation, hashing, expiry, and single-use behavior.
- Handler tests for register, verify-email, login-otp, resend, forgot-password, and reset-password.
- Negative tests for invalid email, expired token, reused token, wrong OTP, and wrong reset link.
- Verify existing accounts remain able to log in after migration.

## Assumptions

- Existing users are marked verified during migration.
- Email OTP is the first release's default second step.
- Google sign-in is treated as already verified.
- TOTP is deferred to a later session and should be added as an optional extension, not mixed into v1.

## Risks

- Email OTP is weaker than app-based TOTP, so it should be framed as an interim step.
- Password reset and OTP endpoints need rate limiting and short TTLs.
- Schema migration must preserve current login behavior for grandfathered users.
