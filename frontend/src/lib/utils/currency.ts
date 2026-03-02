const formatter = (maximumFractionDigits: number) =>
	new Intl.NumberFormat('pt-PT', { style: 'currency', currency: 'EUR', maximumFractionDigits });

export function formatCurrency(amount: number, maximumFractionDigits = 2): string {
	return formatter(maximumFractionDigits).format(amount);
}
