import type { CategoryDto } from '$lib/types';

export function validateTransactionForm(
	amount: string | number,
	category_id: number | string,
	transaction_type_id: number | string,
	categories: CategoryDto[],
	date: string,
	t: (key: string) => string
): {
	error: string | null;
	amount: number;
	date: string;
	category?: CategoryDto;
} {
	const parsedAmount = parseFloat(amount.toString().replace(',', '.'));
	let error: string | null = null;

	if (isNaN(parsedAmount)) {
		error = t('transactions.amount-must-be-number');
		return { error, amount: 0, date };
	}

	const roundedAmount = Math.round(parsedAmount * 100) / 100;

	if (roundedAmount <= 0) {
		error = t('transactions.amount-greater-zero');
		return { error, amount: roundedAmount, date };
	}

	if (roundedAmount > 999999999) {
		error = t('transactions.amount-too-large');
		return { error, amount: roundedAmount, date };
	}

	const category: CategoryDto | undefined = categories.find(
		(cat) => cat.id === Number(category_id)
	);
	if (!category) {
		error = t('transactions.category-required');
		return { error, amount: roundedAmount, date };
	}

	if (category.transaction_type.id !== Number(transaction_type_id)) {
		error = t('transactions.category-must-match');
		return { error, amount: roundedAmount, date, category };
	}

	if (!date) {
		date = new Date().toISOString().split('T')[0];
	}

	return { error: null, amount: roundedAmount, date, category };
}
