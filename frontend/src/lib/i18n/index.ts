import { browser } from '$app/environment';
import { init, register, locale, getLocaleFromNavigator, _ } from 'svelte-i18n';
import { writable } from 'svelte/store';

export const isLoading = writable(true);

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
		loadingDelay: 200
	});

	initialized = true;

	// Set loading to false after initialization
	setTimeout(
		() => {
			isLoading.set(false);
		},
		browser ? 50 : 0
	);
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
