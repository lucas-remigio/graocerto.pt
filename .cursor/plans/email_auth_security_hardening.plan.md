---
name: Email Auth Security Hardening
overview: Harden email confirmation, OTP login, and password reset flows against enumeration, abuse, and token leakage while keeping the current JWT/session model intact.
todos:
  - id: generic-responses
    content: Make public auth entrypoints return generic outcomes to prevent email enumeration
    status: planned
  - id: cooldowns
    content: Add per-email cooldowns for verification and password-reset mail sends
    status: planned
  - id: otp-url-state
    content: Remove OTP challenge state from the URL and use transient client state instead
    status: planned
  - id: token-hygiene
    content: Strip token-bearing URLs from browser history and avoid client-side token logging
    status: planned
  - id: resend-ui
    content: Add a visible resend-verification action in the frontend with cooldown enforcement
    status: planned
  - id: tests
    content: Add handler and regression tests for generic responses, cooldowns, and token reuse/expiry
    status: planned
isProject: false
---

# Email Auth Security Hardening

## Summary

Harden the email confirmation, OTP login, and password reset flows against enumeration, mailbox spam, and token leakage without changing the JWT/session model.

## Key Changes

- Make all public auth entrypoints return generic outcomes.
  - `register`, `resend-verification`, and `forgot-password` must not reveal whether an email exists or is verified.
  - `register` should always return the same accepted-style message after attempting the appropriate action.
- Add per-email cooldowns on outbound auth mail.
  - Enforce separate cooldown windows for verification and password-reset emails.
  - Keep the existing IP rate limiter as a baseline, but do not rely on it as the only abuse control.
- Remove OTP challenge state from the URL.
  - Login should store the `challenge_id` and email in transient browser state and redirect to `/login-otp` without query params.
  - The OTP page should read that state and fail closed if it is missing.
- Reduce token exposure on verification/reset pages.
  - Keep the one-time token in the email link, but immediately strip it from browser history after page load.
  - Never log token-bearing URLs client-side.
- Add a visible resend-verification action in the frontend, backed by the same cooldown rules.
- Keep token TTLs, hashing, and single-use semantics unchanged unless a test proves they need adjustment.

## Test Plan

- Backend handler tests for:
  - generic `register` responses for verified, unverified, and new accounts
  - cooldown behavior for verification and password-reset sends
  - expired, reused, and invalid verify/reset tokens
  - wrong OTP, missing challenge state, and max-attempt exhaustion
- Regression tests for email body and routing:
  - verification links must use the intended frontend path
  - OTP redirect must not include `challenge_id` in the URL
- Run `go test ./service/user ./service/auth` and `pnpm -C frontend check`.

## Assumptions

- Existing users remain grandfathered as verified.
- Email OTP stays the interim second factor; TOTP remains out of scope.
- No CAPTCHA or external anti-abuse service is added in this pass; the baseline is generic responses plus per-email cooldowns.
