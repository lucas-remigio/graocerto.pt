export function validateAccountForm(
	balance: string | number,
	account_name: string,
	t: (key: string) => string
): { error: string | null; balance: number; account_name: string } {
	const parsedBalance = parseFloat(balance.toString().replace(',', '.'));
	let error: string | null = null;

	if (isNaN(parsedBalance)) {
		error = t('accounts.balance-must-be-number');
		return { error, balance: 0, account_name };
	}

	const roundedBalance = Math.round(parsedBalance * 100) / 100;

	if (roundedBalance < 0) {
		error = t('accounts.balance-negative');
	}
	if (roundedBalance > 999999999) {
		error = t('accounts.balance-too-large');
	}
	if (account_name.length < 3) {
		error = t('accounts.account-name-too-short');
	}
	if (account_name.length > 50) {
		error = t('accounts.account-name-too-long');
	}

	return { error, balance: roundedBalance, account_name };
}
