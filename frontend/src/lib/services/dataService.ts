// src/lib/services/dataService.ts
import api_axios from '$lib/axios';
import type {
	Account,
	AccountsResponse,
	TransactionDto,
	CategoryDto,
	CategoriesDtoResponse,
	MonthYear,
	TransactionStatistics,
	TransactionType,
	TransactionTypesResponse,
	TransactionsResponse,
	TransactionChangeResponse,
	CategoryChangeResponse,
	RecurringRule,
	RecurringForecastResponse,
	RecurringForecastRangeDays,
	RecurringRulesResponse,
	NotificationItem,
	NotificationsResponse,
	UnreadNotificationCountResponse,
	NotificationPreferences,
	NotificationPreferencesResponse,
	UpdateNotificationPreferencesPayload,
	CreateRecurringTransferPayload,
	CreateRecurringRulePayload,
	UpdateRecurringTransferPayload,
	UpdateRecurringRulePayload
} from '$lib/types';

// Cache types
type TransactionsCacheValue = {
	transactions: TransactionDto[];
};

// Generic timed cache wrapper
interface TimedValue<T> {
	data: T;
	expiresAt: number;
}

class DataService {
	private readonly TTL_MS = 5 * 60 * 1000; // 5 minutes

	private statisticsCache = new Map<string, TimedValue<TransactionStatistics>>();
	private transactionsCache = new Map<string, TimedValue<TransactionsCacheValue>>();
	private availableMonthsCache = new Map<string, TimedValue<MonthYear[]>>();
	private categoriesCache: TimedValue<CategoryDto[]> | null = null;
	private transactionTypesCache: TimedValue<TransactionType[]> | null = null;
	private recurringRulesCache: TimedValue<RecurringRule[]> | null = null;
	private notificationsCache: TimedValue<NotificationItem[]> | null = null;
	private notificationPreferencesCache: TimedValue<NotificationPreferences> | null = null;

	private now() {
		return Date.now();
	}
	private isValid<T>(entry: TimedValue<T> | undefined | null): entry is TimedValue<T> {
		return !!entry && entry.expiresAt > this.now();
	}
	private wrap<T>(data: T): TimedValue<T> {
		return { data, expiresAt: this.now() + this.TTL_MS };
	}

	// Generate cache key for account/month/year combination
	private getCacheKey(accountToken: string, month: number | null, year: number | null): string {
		return `${accountToken}-${month ?? 'null'}-${year ?? 'null'}`;
	}

	private mutateCategoriesCache(mutator: (list: CategoryDto[]) => CategoryDto[]): void {
		// If cache valid, mutate & refresh TTL. If invalid, do nothing (will repopulate on next fetch).
		if (this.isValid(this.categoriesCache)) {
			const next = mutator([...this.categoriesCache!.data]);
			this.categoriesCache = this.wrap(next);
		}
	}

	// Clear all caches
	clearAllCaches(): void {
		this.statisticsCache.clear();
		this.transactionsCache.clear();
		this.availableMonthsCache.clear();
		this.categoriesCache = null;
		this.transactionTypesCache = null;
		this.recurringRulesCache = null;
		this.notificationsCache = null;
		this.notificationPreferencesCache = null;
	}

	clearTransactionCaches(): void {
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

	// Clear category and transaction type caches (when categories are modified)
	clearCategoryCaches(): void {
		this.categoriesCache = null;
		this.transactionTypesCache = null;
	}

	clearRecurringRulesCache(): void {
		this.recurringRulesCache = null;
	}

	clearNotificationsCache(): void {
		this.notificationsCache = null;
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
		const cached = this.transactionsCache.get(cacheKey);
		if (this.isValid(cached)) return cached.data;
		if (cached) this.transactionsCache.delete(cacheKey);

		const res = await api_axios('transactions/dto/' + accountToken, {
			params: {
				month,
				year
			}
		});

		if (res.status !== 200) {
			throw new Error(`Failed to fetch transactions: ${res.status}`);
		}

		const data: TransactionsResponse = res.data;
		const result: TransactionsCacheValue = {
			transactions: data.transactions
		};

		// Cache the result
		this.transactionsCache.set(cacheKey, this.wrap(result));
		return result;
	}

	// Fetch statistics with caching
	async fetchStatistics(
		accountToken: string,
		month: number | null,
		year: number | null
	): Promise<TransactionStatistics | null> {
		const cacheKey = this.getCacheKey(accountToken, month, year);
		const cached = this.statisticsCache.get(cacheKey);
		if (this.isValid(cached)) return cached.data;
		if (cached) this.statisticsCache.delete(cacheKey);

		const params: { month?: number; year?: number } = {};
		if (month !== null) params.month = month;
		if (year !== null) params.year = year;

		const response = await api_axios.get(`transactions/statistics/${accountToken}`, { params });

		if (response.status !== 200) {
			throw new Error(`Failed to fetch statistics: ${response.status}`);
		}

		const statistics: TransactionStatistics = response.data;

		// Cache the result
		if (statistics) this.statisticsCache.set(cacheKey, this.wrap(statistics));

		return statistics;
	}

	// Fetch available months with caching
	async fetchAvailableMonths(accountToken: string): Promise<MonthYear[]> {
		// Check cache first
		const cached = this.availableMonthsCache.get(accountToken);
		if (this.isValid(cached)) return cached.data;
		if (cached) this.availableMonthsCache.delete(accountToken);

		const res = await api_axios('transactions/months/' + accountToken);

		if (res.status !== 200) {
			throw new Error(`Failed to fetch available months: ${res.status}`);
		}

		const months = res.data.months as MonthYear[];

		// Cache the result
		this.availableMonthsCache.set(accountToken, this.wrap(months));
		return months;
	}

	// Fetch categories with caching (returns flat array for backward compatibility)
	async fetchCategories(): Promise<CategoryDto[]> {
		if (this.isValid(this.categoriesCache)) return this.categoriesCache!.data;
		this.categoriesCache = null;

		const res = await api_axios('categories/dto');

		if (res.status !== 200) {
			throw new Error(`Failed to fetch categories: ${res.status}`);
		}

		const data: CategoriesDtoResponse = res.data;
		// Cache the result
		this.categoriesCache = this.wrap(data.categories);
		return this.categoriesCache.data;
	}

	// Fetch transaction types with caching
	async fetchTransactionTypes(): Promise<TransactionType[]> {
		// Check cache first
		if (this.isValid(this.transactionTypesCache)) return this.transactionTypesCache!.data;
		this.transactionTypesCache = null;

		const res = await api_axios('transaction-types');

		if (res.status !== 200) {
			throw new Error(`Failed to fetch transaction types: ${res.status}`);
		}

		const data: TransactionTypesResponse = res.data;
		// Cache the result
		this.transactionTypesCache = this.wrap(data.transaction_types);
		return this.transactionTypesCache.data;
	}

	// Create category
	async createCategory(categoryData: {
		transaction_type_id: number;
		category_name: string;
		color: string;
		budget: number | null;
	}): Promise<CategoryChangeResponse> {
		const response = await api_axios.post('categories', categoryData);
		if (response.status !== 200) {
			throw new Error(`Failed to create category: ${response.status}`);
		}

		const change: CategoryChangeResponse = response.data;
		const newCat: CategoryDto | undefined = change.category;

		if (newCat?.id) {
			this.mutateCategoriesCache((list) => {
				// Avoid duplicates
				if (!list.find((c) => c.id === newCat.id)) list.push(newCat);
				return list;
			});
		}
		return change;
	}

	// Edit category
	async editCategory(
		categoryId: number,
		categoryData: { category_name: string; color: string; budget: number | null }
	): Promise<CategoryChangeResponse> {
		const response = await api_axios.put(`categories/${categoryId}`, categoryData);

		if (response.status !== 200) {
			throw new Error(`Failed to edit category: ${response.status}`);
		}

		const change: CategoryChangeResponse = response.data;
		const updatedCat: CategoryDto | undefined = change.category;

		this.mutateCategoriesCache((list) =>
			list.map((c) =>
				c.id === categoryId
					? {
							...c,
							category_name: updatedCat?.category_name ?? categoryData.category_name,
							color: updatedCat?.color ?? categoryData.color
						}
					: c
			)
		);

		return change;
	}

	// Reorder categories
	async reorderCategories(categories: { id: number; order_index: number }[]): Promise<void> {
		const response = await api_axios.post('categories/reorder', { categories });
		if (response.status !== 200) {
			throw new Error(`Failed to reorder categories: ${response.status}`);
		}
		this.mutateCategoriesCache((list) => {
			const indexMap = new Map(categories.map((c) => [c.id, c.order_index]));
			return list.map((c) => (indexMap.has(c.id) ? { ...c, order_index: indexMap.get(c.id)! } : c));
		});
	}

	// Delete category
	async deleteCategory(categoryId: number): Promise<void> {
		const response = await api_axios.delete(`categories/${categoryId}`);

		if (response.status !== 200) {
			throw new Error(`Failed to delete category: ${response.status}`);
		}

		// Clear category and transaction type caches since category data changed
		this.clearCategoryCaches();
		this.mutateCategoriesCache((list) => list.filter((c) => c.id !== categoryId));
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
	async deleteTransaction(transaction: TransactionDto): Promise<TransactionChangeResponse> {
		const response = await api_axios.delete(`transactions/${transaction.id}`);

		if (response.status !== 200) {
			throw new Error(`Failed to delete transaction: ${response.status}`);
		}

		// Clear caches only for the account this transaction belongs to
		this.clearAccountCaches(transaction.account_token);
		return response.data;
	}

	async approvePendingTransaction(transaction: TransactionDto): Promise<TransactionChangeResponse> {
		const response = await api_axios.post(`transactions/approve/${transaction.id}`);
		if (response.status !== 200) {
			throw new Error(`Failed to approve transaction: ${response.status}`);
		}
		this.clearAccountCaches(transaction.account_token);
		return response.data;
	}

	async fetchRecurringRules(): Promise<RecurringRule[]> {
		if (this.isValid(this.recurringRulesCache)) return this.recurringRulesCache.data;
		const res = await api_axios.get('recurring-rules');
		if (res.status !== 200) {
			throw new Error(`Failed to fetch recurring rules: ${res.status}`);
		}
		const data: RecurringRulesResponse = res.data;
		this.recurringRulesCache = this.wrap(data.recurring_rules);
		return data.recurring_rules;
	}

	async fetchRecurringForecast(
		accountToken: string,
		days: RecurringForecastRangeDays
	): Promise<RecurringForecastResponse> {
		const res = await api_axios.get('recurring-rules/forecast', {
			params: {
				account_token: accountToken,
				days
			}
		});
		if (res.status !== 200) {
			throw new Error(`Failed to fetch recurring forecast: ${res.status}`);
		}
		return res.data as RecurringForecastResponse;
	}

	async createRecurringRule(payload: CreateRecurringRulePayload): Promise<RecurringRule> {
		const res = await api_axios.post('recurring-rules', payload);
		if (res.status !== 200) {
			throw new Error(`Failed to create recurring rule: ${res.status}`);
		}
		this.clearRecurringRulesCache();
		return res.data.recurring_rule as RecurringRule;
	}

	async createRecurringTransfer(payload: CreateRecurringTransferPayload): Promise<RecurringRule[]> {
		const res = await api_axios.post('recurring-transfers', payload);
		if (res.status !== 200) {
			throw new Error(`Failed to create recurring transfer: ${res.status}`);
		}
		this.clearRecurringRulesCache();
		return res.data.recurring_rules as RecurringRule[];
	}

	async updateRecurringTransfer(
		groupId: string,
		payload: UpdateRecurringTransferPayload
	): Promise<RecurringRule[]> {
		const res = await api_axios.put(`recurring-transfers/${groupId}`, payload);
		if (res.status !== 200) {
			throw new Error(`Failed to update recurring transfer: ${res.status}`);
		}
		this.clearRecurringRulesCache();
		return res.data.recurring_rules as RecurringRule[];
	}

	async updateRecurringRule(
		id: number,
		payload: UpdateRecurringRulePayload
	): Promise<RecurringRule> {
		const res = await api_axios.put(`recurring-rules/${id}`, payload);
		if (res.status !== 200) {
			throw new Error(`Failed to update recurring rule: ${res.status}`);
		}
		this.clearRecurringRulesCache();
		return res.data.recurring_rule as RecurringRule;
	}

	async deleteRecurringRule(id: number): Promise<void> {
		const res = await api_axios.delete(`recurring-rules/${id}`);
		if (res.status !== 200) {
			throw new Error(`Failed to delete recurring rule: ${res.status}`);
		}
		this.clearRecurringRulesCache();
	}

	async fetchNotifications(): Promise<NotificationItem[]> {
		if (this.isValid(this.notificationsCache)) return this.notificationsCache.data;
		const res = await api_axios.get('notifications');
		if (res.status !== 200) {
			throw new Error(`Failed to fetch notifications: ${res.status}`);
		}
		const data: NotificationsResponse = res.data;
		this.notificationsCache = this.wrap(data.notifications);
		return data.notifications;
	}

	async markNotificationAsRead(notificationId: number): Promise<void> {
		const res = await api_axios.patch(`notifications/${notificationId}/read`);
		if (res.status !== 200) {
			throw new Error(`Failed to mark notification as read: ${res.status}`);
		}
		this.clearNotificationsCache();
	}

	async fetchUnreadNotificationCount(): Promise<number> {
		if (this.isValid(this.notificationsCache)) {
			return this.notificationsCache.data.filter((n) => !n.is_read).length;
		}
		try {
			const res = await api_axios.get('notifications/unread-count');
			if (res.status === 200) {
				const data: UnreadNotificationCountResponse = res.data;
				return data.count;
			}
		} catch (err) {
			console.error('Failed to fetch unread notification count:', err);
		}
		return 0;
	}

	async fetchNotificationPreferences(): Promise<NotificationPreferences> {
		if (this.isValid(this.notificationPreferencesCache))
			return this.notificationPreferencesCache.data;
		const res = await api_axios.get('notifications/preferences');
		if (res.status !== 200) {
			throw new Error(`Failed to fetch notification preferences: ${res.status}`);
		}
		const data: NotificationPreferencesResponse = res.data;
		this.notificationPreferencesCache = this.wrap(data.preferences);
		return data.preferences;
	}

	async updateNotificationPreferences(
		payload: UpdateNotificationPreferencesPayload
	): Promise<NotificationPreferences> {
		const res = await api_axios.put('notifications/preferences', payload);
		if (res.status !== 200) {
			throw new Error(`Failed to update notification preferences: ${res.status}`);
		}
		const data: NotificationPreferencesResponse = res.data;
		this.notificationPreferencesCache = this.wrap(data.preferences);
		return data.preferences;
	}

	async registerPushSubscription(sub: PushSubscription): Promise<void> {
		const res = await api_axios.post('notifications/push-subscriptions', sub);
		if (res.status !== 200) {
			throw new Error(`Failed to register push subscription: ${res.status}`);
		}
	}

	async unregisterPushSubscription(endpoint: string): Promise<void> {
		const res = await api_axios.delete('notifications/push-subscriptions', { data: { endpoint } });
		if (res.status !== 200) {
			throw new Error(`Failed to unregister push subscription: ${res.status}`);
		}
	}
}

// Export a singleton instance
export const dataService = new DataService();
