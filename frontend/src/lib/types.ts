export interface Account {
	id: number;
	token: string;
	user_id: number;
	account_name: string;
	balance: number;
	pending_balance: number;
	created_at: string;
	order_index: number;
	is_favorite: boolean;
}

export interface AccountsResponse {
	accounts: Account[];
}

// Types for grouped transactions
export interface TransactionGroup {
	month: number; // Month number (1-12)
	year: number;
	transactions: TransactionDto[];
}

export interface TransactionDto {
	id: number;
	account_token: string;
	amount: number;
	description: string;
	date: string;
	balance: number;
	is_pending: boolean;
	created_at: string;
	category: CategoryDto;
	transaction_type: TransactionType;
	transfer_group_id?: string; // Optional, present if part of a transfer
}

export interface Transaction {
	account_token: string;
	amount: number;
	description: string;
	date: string;
	transaction_type_id: number; // Foreign key to TransactionType
	category_id: number; // Foreign key to Category
}

export interface TransactionsResponse {
	transactions: TransactionDto[];
}
// this applies to both create, edit and delete transaction responses
// ...existing code...

export interface TransactionChangeResponse {
	transaction: TransactionDto;
	account_balance: number;
	account_pending_balance?: number;
	months: MonthYear[];
	is_transfer: boolean;
	paired_account_token?: string;
	paired_account_balance?: number;
	paired_account_pending_balance?: number;
	paired_account_months?: MonthYear[];
}

// ...existing code...

export interface CreateTransferPayload {
	source_account_token: string;
	destination_account_token: string;
	debit_category_id: number;
	credit_category_id: number;
	amount: number;
	description: string;
	date: string;
}

export interface TransferResponse {
	transfer_group_id: string;
	debit_transaction: TransactionDto;
	credit_transaction: TransactionDto;
	source_account_balance: number;
	destination_account_balance: number;
	source_account_months: MonthYear[];
	destination_account_months: MonthYear[];
}

export interface AccountChangeResponse {
	account: Account;
}

export interface CategoryChangeResponse {
	category: CategoryDto;
}

export interface TransactionsTotals {
	debit: number;
	credit: number;
	difference: number;
}

export interface CategoryDto {
	id: number;
	transaction_type: TransactionType;
	parent_category_id?: number | null;
	parent_category?: CategoryDto;
	subcategories?: CategoryDto[];
	category_name: string;
	color: string;
	budget?: number | null;
	order_index: number;
	created_at: string;
	updated_at: string;
	deleted_at?: string | null;
}

export interface CreateCategoryPayload {
	transaction_type_id: number;
	parent_category_id?: number | null;
	category_name: string;
	color: string;
	budget?: number | null;
}

export interface UpdateCategoryPayload {
	parent_category_id?: number | null;
	category_name: string;
	color: string;
	budget?: number | null;
}

export interface Category {
	id: number;
	transaction_type_id: number;
	category_name: string;
	color: string;
	budget: number | null;
	created_at: string;
	updated_at: string;
}

export interface CategoriesResponse {
	categories: Category[];
}

export interface CategoriesDtoResponse {
	categories: CategoryDto[];
}

export interface TransactionType {
	id: number;
	type_name: string;
	type_slug: string;
}

export interface TransactionTypesResponse {
	transaction_types: TransactionType[];
}

export interface AiFeedbackResponse {
	feedback_message: string;
	in_depth_analysis: string;
}

export interface MonthYear {
	month: number; // 1-12 (1 = January)
	year: number;
	count: number; // Number of transactions in that month/year
}

export interface InvestmentCalculatorResponse {
	total_investment: number;
	total_return: number;
	total_value: number;
	yearly_breakdown: YearlyBreakdown[];
}

export interface YearlyBreakdown {
	year: number;
	total_investment: number;
	total_return: number;
	total_value: number;
}

export interface InvestmentCalculatorInput {
	initial_investment: number;
	monthly_contribution: number;
	annual_return_rate: number; // As a percentage (e.g., 5 for 5%)
	investment_duration_years: number; // Number of years to calculate
}

export interface CategoryStatistic {
	name: string;
	count: number;
	total: number;
	percentage: number;
	color: string;
	budget: number | null;
	budget_percentage: number;
	subcategories?: CategoryStatistic[];
}

export interface DailyTotals {
	date: string; // Format: YYYY-MM-DD
	credit: number;
	debit: number;
	difference: number;
}

export interface TransactionStatistics {
	total_transactions: number;
	largest_debit: number;
	largest_credit: number;
	credit_category_breakdown: CategoryStatistic[];
	debit_category_breakdown: CategoryStatistic[];
	totals: TransactionsTotals;
	daily_totals: DailyTotals[];
	start_date: string; // Format: YYYY-MM-DD
	end_date: string; // Format: YYYY-MM-DD
}

export interface TrendMonth {
	month: number;
	year: number;
}

export interface CategoryTrend {
	id: number;
	name: string;
	color: string;
	totals: number[]; // aligned to CategoryTrendsResponse.months
	total: number; // sum over the window
	subcategories?: CategoryTrend[]; // only on root categories; parent totals include these
}

export interface CategoryTrendsResponse {
	account_token: string;
	transaction_type: number;
	months: TrendMonth[];
	totals: number[]; // grand total per month, aligned to months
	categories: CategoryTrend[];
}

export type TrendsRangeMonths = 6 | 12 | 24;
export type TrendsType = 'debit' | 'credit';

export type RecurringFrequency = 'daily' | 'weekly' | 'monthly' | 'every_x_days';

export interface RecurringRule {
	id: number;
	user_id: number;
	account_token: string;
	category_id: number;
	transaction_type_id: number;
	recurring_transfer_group_id?: string;
	amount: number;
	description: string;
	frequency: RecurringFrequency;
	interval_value: number;
	next_run_date: string;
	active: boolean;
	created_at: string;
	updated_at: string;
}

export interface RecurringRulesResponse {
	recurring_rules: RecurringRule[];
}

export type RecurringForecastRangeDays = 30 | 60 | 90;

export interface RecurringForecastItem {
	recurring_rule_id: number;
	account_token: string;
	category_id: number;
	transaction_type_id: number;
	recurring_transfer_group_id?: string;
	amount: number;
	description: string;
	date: string;
}

export interface RecurringForecastSummary {
	credit: number;
	debit: number;
	difference: number;
}

export interface RecurringForecastResponse {
	account_token: string;
	days: RecurringForecastRangeDays;
	items: RecurringForecastItem[];
	summary: RecurringForecastSummary;
}

export interface CashflowProjectionResponse {
	account_token: string;
	current_balance: number;
	upcoming_credits: number;
	upcoming_debits: number;
	projected_balance: number;
	days_remaining: number;
	period_end: string;
}

export interface NotificationItem {
	id: number;
	user_id: number;
	type: string;
	account_token?: string;
	target_date?: string;
	notify_days_ahead: number;
	debit_count: number;
	total_debit: number;
	credit_count: number;
	total_credit: number;
	is_read: boolean;
	created_at: string;
	// Budget-threshold notifications only. For this type total_debit is the amount
	// spent and total_credit is the category budget.
	category_id?: number;
	threshold_pct?: number;
	category_name?: string;
}

export interface NotificationsResponse {
	notifications: NotificationItem[];
}

export interface NotificationPreferences {
	user_id: number;
	enabled: boolean;
	notify_days_ahead: number;
	min_total_debit?: number | null;
	budget_alert_threshold: number;
	push_endpoints: string[];
	updated_at: string;
}

export interface NotificationPreferencesResponse {
	preferences: NotificationPreferences;
}

export interface UnreadNotificationCountResponse {
	count: number;
}

export interface PushSubscriptionPayload {
	endpoint: string;
	p256dh: string;
	auth: string;
}

export interface UpdateNotificationPreferencesPayload {
	enabled: boolean;
	notify_days_ahead: number;
	min_total_debit?: number | null;
	budget_alert_threshold: number;
}

export interface CreateRecurringRulePayload {
	account_token: string;
	category_id: number;
	transaction_type_id: number;
	amount: number;
	description: string;
	frequency: RecurringFrequency;
	interval_value: number;
	execution_day?: number;
	execution_weekday?: number;
	active?: boolean;
}

export interface UpdateRecurringRulePayload extends CreateRecurringRulePayload {
	next_run_date?: string;
	active: boolean;
}

export interface CreateRecurringTransferPayload {
	source_account_token: string;
	destination_account_token: string;
	debit_category_id: number;
	credit_category_id: number;
	amount: number;
	description: string;
	frequency: RecurringFrequency;
	interval_value: number;
	execution_day?: number;
	execution_weekday?: number;
	active?: boolean;
}

export interface UpdateRecurringTransferPayload extends CreateRecurringTransferPayload {
	next_run_date?: string;
	active: boolean;
}
