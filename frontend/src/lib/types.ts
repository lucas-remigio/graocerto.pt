export interface Account {
	id: number;
	token: string;
	user_id: number;
	account_name: string;
	balance: number;
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

export interface GroupedTransactionsResponse {
	groups: TransactionGroup[];
	totals: TransactionsTotals;
}

export interface TransactionDto {
	id: number;
	account_token: string;
	amount: number;
	description: string;
	date: string;
	balance: number;
	created_at: string;
	category: CategoryDto;
}

export interface Transaction {
	account_token: string;
	amount: number;
	description: string;
	date: string;
	transaction_type_id: number; // Foreign key to TransactionType
	category_id: number; // Foreign key to Category
}

export interface TransactionsTotals {
	debit: number;
	credit: number;
	difference: number;
}

export interface CategoryDto {
	id: number;
	transaction_type: TransactionType;
	category_name: string;
	color: string;
	created_at: string;
	updated_at: string;
}

export interface Category {
	id: number;
	transaction_type_id: number;
	category_name: string;
	color: string;
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
}

export interface DailyTotals {
	date: string; // Format: YYYY-MM-DD
	total: number;
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
