export interface Account {
	id: number;
	token: string;
	user_id: number;
	account_name: string;
	balance: number;
	created_at: string;
}

export interface AccountsResponse {
	accounts: Account[];
}

export interface TransactionsResponseDto {
	transactions: TransactionDto[];
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
