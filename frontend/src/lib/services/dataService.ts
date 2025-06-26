// src/lib/services/dataService.ts
import api_axios from '$lib/axios';
import type {
	Account,
	AccountsResponse,
	TransactionDto,
	GroupedTransactionsResponse,
	TransactionGroup,
	CategoryDto,
	MonthYear,
	TransactionsTotals,
	TransactionStatistics
} from '$lib/types';

// Cache types
type TransactionsCacheValue = {
	transactionGroups: TransactionGroup[];
	totals: TransactionsTotals;
};

class DataService {
	private statisticsCache = new Map<string, TransactionStatistics>();
	private transactionsCache = new Map<string, TransactionsCacheValue>();
	private availableMonthsCache = new Map<string, MonthYear[]>();

	// Generate cache key for account/month/year combination
	private getCacheKey(accountToken: string, month: number | null, year: number | null): string {
		return `${accountToken}-${month ?? 'null'}-${year ?? 'null'}`;
	}

	// Clear all caches
	clearCaches(): void {
		this.statisticsCache.clear();
		this.transactionsCache.clear();
		this.availableMonthsCache.clear();
	}

	// Clear caches for a specific account
	clearAccountCaches(accountToken: string): void {
		// Remove all entries that start with the account token
		for (const key of this.statisticsCache.keys()) {
			if (key.startsWith(accountToken)) {
				this.statisticsCache.delete(key);
			}
		}
		for (const key of this.transactionsCache.keys()) {
			if (key.startsWith(accountToken)) {
				this.transactionsCache.delete(key);
			}
		}
		this.availableMonthsCache.delete(accountToken);
	}

	// Fetch accounts
	async fetchAccounts(): Promise<Account[]> {
		const res = await api_axios('accounts');

		if (res.status !== 200) {
			throw new Error(`Failed to fetch accounts: ${res.status}`);
		}

		const data: AccountsResponse = res.data;
		return data.accounts;
	}

	// Fetch transactions with caching
	async fetchTransactions(
		accountToken: string,
		month: number | null,
		year: number | null
	): Promise<TransactionsCacheValue> {
		const cacheKey = this.getCacheKey(accountToken, month, year);

		// Check cache first
		if (this.transactionsCache.has(cacheKey)) {
			return this.transactionsCache.get(cacheKey)!;
		}

		const res = await api_axios('transactions/dto/' + accountToken, {
			params: {
				month,
				year
			}
		});

		if (res.status !== 200) {
			throw new Error(`Failed to fetch transactions: ${res.status}`);
		}

		const data: GroupedTransactionsResponse = res.data;
		const result: TransactionsCacheValue = {
			transactionGroups: data.groups,
			totals: data.totals
		};

		// Cache the result
		this.transactionsCache.set(cacheKey, result);
		return result;
	}

	// Fetch statistics with caching
	async fetchStatistics(
		accountToken: string,
		month: number | null,
		year: number | null
	): Promise<TransactionStatistics | null> {
		const cacheKey = this.getCacheKey(accountToken, month, year);

		// Check cache first
		if (this.statisticsCache.has(cacheKey)) {
			return this.statisticsCache.get(cacheKey)!;
		}

		const params: { month?: number; year?: number } = {};
		if (month !== null) params.month = month;
		if (year !== null) params.year = year;

		const response = await api_axios.get(`transactions/statistics/${accountToken}`, { params });

		if (response.status !== 200) {
			throw new Error(`Failed to fetch statistics: ${response.status}`);
		}

		const statistics: TransactionStatistics = response.data;

		// Cache the result
		if (statistics) {
			this.statisticsCache.set(cacheKey, statistics);
		}

		return statistics;
	}

	// Fetch available months with caching
	async fetchAvailableMonths(accountToken: string): Promise<MonthYear[]> {
		// Check cache first
		if (this.availableMonthsCache.has(accountToken)) {
			return this.availableMonthsCache.get(accountToken)!;
		}

		const res = await api_axios('transactions/months/' + accountToken);

		if (res.status !== 200) {
			throw new Error(`Failed to fetch available months: ${res.status}`);
		}

		const months = res.data.months as MonthYear[];

		// Cache the result
		this.availableMonthsCache.set(accountToken, months);
		return months;
	}

	// Fetch categories
	async fetchCategories(): Promise<CategoryDto[]> {
		const res = await api_axios('categories/dto');

		if (res.status !== 200) {
			throw new Error(`Failed to fetch categories: ${res.status}`);
		}

		return res.data.categories;
	}

	// Delete account
	async deleteAccount(accountToken: string): Promise<void> {
		const response = await api_axios.delete(`accounts/${accountToken}`);

		if (response.status !== 200) {
			throw new Error(`Failed to delete account: ${response.status}`);
		}

		// Clear caches for this account
		this.clearAccountCaches(accountToken);
	}

	// Delete transaction
	async deleteTransaction(transaction: TransactionDto): Promise<void> {
		const response = await api_axios.delete(`transactions/${transaction.id}`);

		if (response.status !== 200) {
			throw new Error(`Failed to delete transaction: ${response.status}`);
		}

		// Clear caches only for the account this transaction belongs to
		this.clearAccountCaches(transaction.account_token);
	}
}

// Export a singleton instance
export const dataService = new DataService();
