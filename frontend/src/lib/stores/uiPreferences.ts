import { writable } from 'svelte/store';
import { themeService } from '$lib/services/themeService';

const LOCAL_STORAGE_HIDE_BALANCES = 'ui.hideBalances';
const LOCAL_STORAGE_SELECTED_VIEW = 'ui.selectedView';
const LOCAL_STORAGE_THEME = 'ui.theme';

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

function getInitialTheme(): 'light' | 'dark' {
	if (typeof localStorage !== 'undefined') {
		const stored = localStorage.getItem(LOCAL_STORAGE_THEME);
		if (stored === 'light' || stored === 'dark') {
			return stored;
		}
	}
	// Fallback to system preference
	if (typeof window !== 'undefined') {
		return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
	}
	return 'light';
}

export const hideBalances = writable(getInitialHideBalances());
export const selectedView = writable<'transactions' | 'statistics'>(getInitialSelectedView());
export const theme = writable<'light' | 'dark'>(getInitialTheme());

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
		document.documentElement.setAttribute('data-theme', value);
		document.documentElement.classList.toggle('dark', value === 'dark');
		themeService.updateThemeColor(value);
	});
}
