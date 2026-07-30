import type { Account } from '$lib/types';

// localStorage key holding the token of the account the user last selected.
// Shared so reads (pickInitialAccount) and writes (rememberSelectedAccount) can't drift.
export const SELECTED_ACCOUNT_KEY = 'selectedAccount';

/**
 * Pick the account to show initially: the one persisted in localStorage, else
 * the first account. Returns `current` unchanged when a selection already
 * exists or there are no accounts to choose from.
 */
export function pickInitialAccount(accounts: Account[], current: Account | null): Account | null {
	if (current || accounts.length === 0) return current;
	const stored = localStorage.getItem(SELECTED_ACCOUNT_KEY);
	return accounts.find((account) => account.token === stored) ?? accounts[0];
}

/** Persist the selected account so it survives reloads and is shared across pages. */
export function rememberSelectedAccount(token: string): void {
	localStorage.setItem(SELECTED_ACCOUNT_KEY, token);
}

/**
 * Insert or replace an account by token, returning a new list sorted by
 * order_index. Does not mutate the input.
 */
export function upsertAccount(accounts: Account[], account: Account): Account[] {
	const next = accounts.filter((existing) => existing.token !== account.token);
	next.push(account);
	return next.sort((a, b) => a.order_index - b.order_index);
}
