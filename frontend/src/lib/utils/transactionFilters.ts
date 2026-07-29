/**
 * Shared transaction-style filtering used by the transactions table and the
 * recurring rules list, so the filter predicate lives in a single place.
 */

export interface TransactionFilterValues {
	searchTerm: string;
	categoryId: number | null;
	typeSlug: string | null;
	minAmount: number | null;
	maxAmount: number | null;
	startDate: string;
	endDate: string;
}

/** A normalised view of any item that can be filtered (transaction, recurring rule, ...). */
export interface FilterableItem {
	description?: string | null;
	categoryId: number;
	typeSlug: string;
	/** Any date string; the time part (after "T") is ignored. */
	date: string;
	amount: number;
}

export function emptyTransactionFilters(): TransactionFilterValues {
	return {
		searchTerm: '',
		categoryId: null,
		typeSlug: null,
		minAmount: null,
		maxAmount: null,
		startDate: '',
		endDate: ''
	};
}

export function countActiveFilters(filters: TransactionFilterValues): number {
	return Object.values(filters).filter(Boolean).length;
}

export function matchesTransactionFilters(
	filters: TransactionFilterValues,
	item: FilterableItem
): boolean {
	// Search term (description)
	if (
		filters.searchTerm &&
		!item.description?.toLowerCase().includes(filters.searchTerm.toLowerCase())
	) {
		return false;
	}

	// Category filter
	if (filters.categoryId && item.categoryId !== filters.categoryId) {
		return false;
	}

	// Type filter
	if (filters.typeSlug && item.typeSlug !== filters.typeSlug) {
		return false;
	}

	// Amount range
	if (filters.minAmount !== null && item.amount < filters.minAmount) {
		return false;
	}
	if (filters.maxAmount !== null && item.amount > filters.maxAmount) {
		return false;
	}

	// Date range - compare only the date part (YYYY-MM-DD)
	const itemDate = item.date?.split('T')[0] ?? '';
	if (filters.startDate && itemDate < filters.startDate) {
		return false;
	}
	if (filters.endDate && itemDate > filters.endDate) {
		return false;
	}

	return true;
}
