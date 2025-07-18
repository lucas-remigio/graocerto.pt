import type { Transaction } from '$lib/types';

let draftTransaction: Transaction | null = null;
let lastAccountToken: string | null = null;

export function setDraftTransactionAccountToken(accountToken: string | null) {
	if (lastAccountToken !== accountToken) {
		// Reset draft transaction if the account token changes
		draftTransaction = null;
		lastAccountToken = accountToken;
	}
}

export function setDraftTransaction(transaction: Transaction | null) {
	draftTransaction = transaction;
}

export function getDraftTransaction(): Transaction | null {
	return draftTransaction;
}
