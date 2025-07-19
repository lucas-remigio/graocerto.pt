import { browser } from '$app/environment';
import {
	init,
	register,
	locale,
	getLocaleFromNavigator,
	_,
	isLoading as svelteI18nLoading
} from 'svelte-i18n';
import { derived, writable } from 'svelte/store';

// Use svelte-i18n's built-in loading state
export const isLoading = derived(svelteI18nLoading, ($loading) => $loading);

// Track if i18n is fully initialized
export const i18nReady = writable(false);

register('en', () => import('./locales/en.json'));
register('pt', () => import('./locales/pt.json'));

let initialized = false;

export function setupI18n() {
	if (initialized) return;

	const initialLocale = browser
		? localStorage.getItem('preferred-language') || getLocaleFromNavigator() || 'pt'
		: 'pt';

	init({
		fallbackLocale: 'pt',
		initialLocale,
		loadingDelay: 100 // Reduced for faster feedback in dev
	});

	// Mark as initialized immediately to prevent multiple calls
	initialized = true;

	// Wait for i18n to be fully ready
	let unsub: (() => void) | null = null;
	unsub = svelteI18nLoading.subscribe((loading) => {
		if (!loading) {
			i18nReady.set(true);
			if (unsub) unsub();
		}
	});
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
