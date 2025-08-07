import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface CookieConsent {
	necessary: boolean;
	analytics: boolean;
	marketing: boolean;
	hasChosenPreferences: boolean;
}

const DEFAULT_CONSENT: CookieConsent = {
	necessary: true, // Always required for app functionality
	analytics: false,
	marketing: false,
	hasChosenPreferences: false
};

const STORAGE_KEY = 'cookie-consent';

function createCookieConsentStore() {
	const { subscribe, set } = writable<CookieConsent>(DEFAULT_CONSENT);

	const saveToStorage = (consent: CookieConsent) => {
		if (browser) {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(consent));
		}
	};

	const loadFromStorage = (): CookieConsent | null => {
		if (!browser) return null;

		try {
			const stored = localStorage.getItem(STORAGE_KEY);
			return stored ? JSON.parse(stored) : null;
		} catch (error) {
			console.error('Failed to parse cookie consent from storage:', error);
			return null;
		}
	};

	const updateConsent = (newConsent: Partial<CookieConsent>) => {
		const consent: CookieConsent = {
			...DEFAULT_CONSENT,
			...newConsent,
			necessary: true, // Always enforce necessary cookies
			hasChosenPreferences: true
		};

		set(consent);
		saveToStorage(consent);
		return consent;
	};

	return {
		subscribe,

		init() {
			const storedConsent = loadFromStorage();
			if (storedConsent) {
				set({ ...DEFAULT_CONSENT, ...storedConsent, hasChosenPreferences: true });
			}
		},

		acceptAll() {
			return updateConsent({
				analytics: true,
				marketing: true
			});
		},

		rejectAll() {
			return updateConsent({
				analytics: false,
				marketing: false
			});
		},

		setPreferences(preferences: Partial<CookieConsent>) {
			return updateConsent(preferences);
		},

		reopenBanner() {
			if (browser) {
				localStorage.removeItem(STORAGE_KEY);
			}
			set({ ...DEFAULT_CONSENT, hasChosenPreferences: false });
		}
	};
}

export const cookieConsent = createCookieConsentStore();
