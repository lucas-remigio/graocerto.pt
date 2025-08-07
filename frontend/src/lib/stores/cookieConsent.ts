import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface CookieConsent {
	necessary: boolean;
	analytics: boolean;
	marketing: boolean;
	hasChosenPreferences: boolean;
}

const defaultConsent: CookieConsent = {
	necessary: true, // Always true, required for the app to work
	analytics: false,
	marketing: false,
	hasChosenPreferences: false
};

function createCookieConsentStore() {
	const { subscribe, set, update } = writable<CookieConsent>(defaultConsent);

	return {
		subscribe,
		init: () => {
			if (browser) {
				const stored = localStorage.getItem('cookie-consent');
				if (stored) {
					try {
						const consent = JSON.parse(stored);
						set({ ...defaultConsent, ...consent, hasChosenPreferences: true });
					} catch (e) {
						console.error('Failed to parse stored cookie consent:', e);
					}
				}
			}
		},
		acceptAll: () => {
			const consent: CookieConsent = {
				necessary: true,
				analytics: true,
				marketing: true,
				hasChosenPreferences: true
			};
			set(consent);
			if (browser) {
				localStorage.setItem('cookie-consent', JSON.stringify(consent));
			}
		},
		rejectAll: () => {
			const consent: CookieConsent = {
				necessary: true,
				analytics: false,
				marketing: false,
				hasChosenPreferences: true
			};
			set(consent);
			if (browser) {
				localStorage.setItem('cookie-consent', JSON.stringify(consent));
			}
		},
		setPreferences: (preferences: Partial<CookieConsent>) => {
			update((current) => {
				const newConsent = {
					...current,
					...preferences,
					necessary: true, // Always required
					hasChosenPreferences: true
				};
				if (browser) {
					localStorage.setItem('cookie-consent', JSON.stringify(newConsent));
				}
				return newConsent;
			});
		}
	};
}

export const cookieConsent = createCookieConsentStore();
