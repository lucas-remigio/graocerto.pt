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

## Testing Rules

- Every new feature must include unit tests.
- Tests should cover success paths, validation failures, and edge cases.
- Keep test setup minimal and close to the behavior being verified.
- Update or add tests before considering the feature complete.

