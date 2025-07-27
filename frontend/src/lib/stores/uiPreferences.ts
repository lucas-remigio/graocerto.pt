import { writable } from 'svelte/store';

const LOCAL_STORAGE_HIDE_BALANCES = 'ui.hideBalances';
const LOCAL_STORAGE_SELECTED_VIEW = 'ui.selectedView';

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

export const hideBalances = writable(getInitialHideBalances());
export const selectedView = writable<'transactions' | 'statistics'>(getInitialSelectedView());

// Only subscribe to localStorage updates in the browser
if (typeof window !== 'undefined') {
	hideBalances.subscribe((value) => {
		localStorage.setItem(LOCAL_STORAGE_HIDE_BALANCES, value.toString());
	});
	selectedView.subscribe((value) => {
		localStorage.setItem(LOCAL_STORAGE_SELECTED_VIEW, value);
	});
}
