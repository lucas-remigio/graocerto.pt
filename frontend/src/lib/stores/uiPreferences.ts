import { derived, writable } from 'svelte/store';
import { themeService } from '$lib/services/themeService';

export type ThemeOption = 'light' | 'dark' | 'system';
export type HeatmapDisplayMode = 'difference' | 'credit' | 'debit';

const LOCAL_STORAGE_HIDE_BALANCES = 'ui.hideBalances';
const LOCAL_STORAGE_SELECTED_VIEW = 'ui.selectedView';
const LOCAL_STORAGE_THEME = 'ui.theme';
const LOCAL_STORAGE_SHOW_NON_FAVORITES = 'ui.showNonFavorites';
const LOCAL_STORAGE_HEATMAP_MODE = 'ui.heatmapDisplayMode';

function getInitialHideBalances() {
	if (typeof localStorage !== 'undefined') {
		return localStorage.getItem(LOCAL_STORAGE_HIDE_BALANCES) === 'true';
	}
	return false;
}

function getInitialSelectedView() {
	if (typeof localStorage !== 'undefined') {
		const stored = localStorage.getItem(LOCAL_STORAGE_SELECTED_VIEW);
		if (stored === 'statistics' || stored === 'transactions') {
			return stored;
		}
	}
	return 'transactions';
}

function getInitialTheme(): ThemeOption {
	if (typeof localStorage !== 'undefined') {
		const stored = localStorage.getItem(LOCAL_STORAGE_THEME);
		if (stored === 'light' || stored === 'dark' || stored === 'system') {
			return stored;
		}
	}
	return 'system';
}

function getInitialShowNonFavorites() {
	if (typeof localStorage !== 'undefined') {
		return localStorage.getItem(LOCAL_STORAGE_SHOW_NON_FAVORITES) === 'true';
	}
	return false;
}

function getInitialHeatmapMode(): HeatmapDisplayMode {
	if (typeof localStorage !== 'undefined') {
		const stored = localStorage.getItem(LOCAL_STORAGE_HEATMAP_MODE);
		if (stored === 'difference' || stored === 'credit' || stored === 'debit') {
			return stored;
		}
	}
	return 'difference'; // Default to 'difference'
}

export const hideBalances = writable(getInitialHideBalances());
export const selectedView = writable<'transactions' | 'statistics'>(getInitialSelectedView());
export const theme = writable<ThemeOption>(getInitialTheme());
export const showNonFavorites = writable(getInitialShowNonFavorites());
export const heatmapDisplayMode = writable<HeatmapDisplayMode>(getInitialHeatmapMode());

// This derived store always returns 'light' or 'dark'
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

// Persist and apply theme changes
if (typeof window !== 'undefined') {
	hideBalances.subscribe((value) => {
		localStorage.setItem(LOCAL_STORAGE_HIDE_BALANCES, value.toString());
	});
	selectedView.subscribe((value) => {
		localStorage.setItem(LOCAL_STORAGE_SELECTED_VIEW, value);
	});
	theme.subscribe((value) => {
		localStorage.setItem(LOCAL_STORAGE_THEME, value);
		applyTheme(value);
	});
	// Listen for system theme changes if "system" is selected
	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
		if (localStorage.getItem(LOCAL_STORAGE_THEME) === 'system') {
			applyTheme('system');
		}
	});
	showNonFavorites.subscribe((value) => {
		localStorage.setItem(LOCAL_STORAGE_SHOW_NON_FAVORITES, value.toString());
	});
	heatmapDisplayMode.subscribe((value) => {
		localStorage.setItem(LOCAL_STORAGE_HEATMAP_MODE, value);
	});
}
