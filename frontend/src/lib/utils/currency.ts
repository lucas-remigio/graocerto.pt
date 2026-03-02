import { browser } from '$app/environment';

const LOCALE_MAP: Record<string, string> = {
	pt: 'pt-PT',
	en: 'en-GB'
};

function getIntlLocale(appLocale?: string): string {
	const locale =
		appLocale ?? (browser ? (localStorage.getItem('preferred-language') ?? 'pt') : 'pt');
	return LOCALE_MAP[locale] ?? 'pt-PT';
}

export function formatCurrency(amount: number, maximumFractionDigits = 2, locale?: string): string {
	return new Intl.NumberFormat(getIntlLocale(locale), {
		style: 'currency',
		currency: 'EUR',
		maximumFractionDigits,
		useGrouping: true
	}).format(amount);
}
