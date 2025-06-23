import { browser } from '$app/environment';
import {
	init,
	register,
	locale,
	getLocaleFromNavigator,
	_,
	isLoading as svelteI18nLoading
} from 'svelte-i18n';
import { derived } from 'svelte/store';

// Use svelte-i18n's built-in loading state
export const isLoading = derived(svelteI18nLoading, ($loading) => $loading);

register('en', () => import('./locales/en.json'));
register('pt', () => import('./locales/pt.json'));

let initialized = false;

export function setupI18n() {
	if (initialized) return;

	const initialLocale = browser
		? localStorage.getItem('preferred-language') || getLocaleFromNavigator() || 'en'
		: 'en';

	init({
		fallbackLocale: 'en',
		initialLocale,
		loadingDelay: 100 // Reduced for faster feedback in dev
	});

	initialized = true;
}

export function setLocale(newLocale: string) {
	locale.set(newLocale);
	if (browser) {
		localStorage.setItem('preferred-language', newLocale);
	}
}

export const t = _;
export { locale };

// Initialize immediately for both client and server
setupI18n();
