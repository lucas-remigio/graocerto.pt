import { derived, writable } from 'svelte/store';
import { themeService } from '$lib/services/themeService';

export type ThemeOption = 'light' | 'dark' | 'system';
export type HeatmapDisplayMode = 'difference' | 'credit' | 'debit';

interface UIPreferences {
	hideBalances: boolean;
	selectedView: 'transactions' | 'statistics';
	theme: ThemeOption;
	showNonFavorites: boolean;
	heatmapDisplayMode: HeatmapDisplayMode;
}

const LOCAL_STORAGE_KEY = 'ui.preferences';

// Default preferences
const defaultPreferences: UIPreferences = {
	hideBalances: false,
	selectedView: 'transactions',
	theme: 'system',
	showNonFavorites: false,
	heatmapDisplayMode: 'difference'
};

// Load preferences from localStorage
function loadPreferences(): UIPreferences {
	if (typeof localStorage === 'undefined') {
		return { ...defaultPreferences };
	}

	try {
		const stored = localStorage.getItem(LOCAL_STORAGE_KEY);
		if (stored) {
			const parsed = JSON.parse(stored);
			// Merge with defaults to handle new preferences
			return { ...defaultPreferences, ...parsed };
		}
	} catch (error) {
		console.warn('Failed to load UI preferences:', error);
	}

	return { ...defaultPreferences };
}

// Save preferences to localStorage
function savePreferences(preferences: UIPreferences) {
	if (typeof localStorage === 'undefined') return;

	try {
		localStorage.setItem(LOCAL_STORAGE_KEY, JSON.stringify(preferences));
	} catch (error) {
		console.warn('Failed to save UI preferences:', error);
	}
}

// Create the main preferences store
export const uiPreferences = writable<UIPreferences>(loadPreferences());

// Individual stores derived from the main store
export const hideBalances = derived(uiPreferences, ($prefs) => $prefs.hideBalances);

export const selectedView = derived(uiPreferences, ($prefs) => $prefs.selectedView);

export const theme = derived(uiPreferences, ($prefs) => $prefs.theme);

export const showNonFavorites = derived(uiPreferences, ($prefs) => $prefs.showNonFavorites);

export const heatmapDisplayMode = derived(uiPreferences, ($prefs) => $prefs.heatmapDisplayMode);

// Helper functions to update individual preferences
export const updateHideBalances = (value: boolean) => {
	uiPreferences.update((prefs) => ({ ...prefs, hideBalances: value }));
};

export const updateSelectedView = (value: 'transactions' | 'statistics') => {
	uiPreferences.update((prefs) => ({ ...prefs, selectedView: value }));
};

export const updateTheme = (value: ThemeOption) => {
	uiPreferences.update((prefs) => ({ ...prefs, theme: value }));
};

export const updateShowNonFavorites = (value: boolean) => {
	uiPreferences.update((prefs) => ({ ...prefs, showNonFavorites: value }));
};

export const updateHeatmapDisplayMode = (value: HeatmapDisplayMode) => {
	uiPreferences.update((prefs) => ({ ...prefs, heatmapDisplayMode: value }));
};

// Applied theme store (derived from theme)
export const appliedTheme = derived(theme, ($theme, set) => {
	function resolveTheme(themeValue: ThemeOption): 'light' | 'dark' {
		if (themeValue === 'system') {
			return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
		}
		return themeValue;
	}

	// Set initial value
	set(resolveTheme($theme));

	// Listen for system changes if theme is 'system'
	let mql: MediaQueryList | null = null;
	function handleChange() {
		set(resolveTheme($theme));
	}
	if ($theme === 'system' && typeof window !== 'undefined') {
		mql = window.matchMedia('(prefers-color-scheme: dark)');
		mql.addEventListener('change', handleChange);
	}
	return () => {
		if (mql) mql.removeEventListener('change', handleChange);
	};
});

// Helper to apply theme
function applyTheme(themeValue: ThemeOption) {
	let applied: 'light' | 'dark';
	if (themeValue === 'system') {
		applied = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
	} else {
		applied = themeValue;
	}
	document.documentElement.setAttribute('data-theme', applied);
	document.documentElement.classList.toggle('dark', applied === 'dark');
	themeService.updateThemeColor(applied);
}

// Subscribe to changes and persist
if (typeof window !== 'undefined') {
	// Save preferences whenever they change
	uiPreferences.subscribe((prefs) => {
		savePreferences(prefs);
	});

	// Apply theme changes
	theme.subscribe((value) => {
		applyTheme(value);
	});

	// Listen for system theme changes
	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
		uiPreferences.subscribe((prefs) => {
			if (prefs.theme === 'system') {
				applyTheme('system');
			}
		})();
	});
}
