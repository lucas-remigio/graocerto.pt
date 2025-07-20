import type { TransactionType } from './types';

export enum TransactionTypeSlug {
	Credit = 'credit',
	Debit = 'debit',
	Transfer = 'transfer'
}

export const TransactionTypes: TransactionType[] = [
	{ id: 1, type_name: 'Credit', type_slug: TransactionTypeSlug.Credit },
	{ id: 2, type_name: 'Debit', type_slug: TransactionTypeSlug.Debit },
	{ id: 3, type_name: 'Transfer', type_slug: TransactionTypeSlug.Transfer }
];

export enum TransactionTypeId {
	Credit = 1,
	Debit = 2,
	Transfer = 3
}
