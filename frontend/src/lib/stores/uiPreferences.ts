import { writable } from 'svelte/store';

const LOCAL_STORAGE_HIDE_BALANCES = 'ui.hideBalances';

function getInitialHideBalances() {
	if (typeof localStorage !== 'undefined') {
		return localStorage.getItem(LOCAL_STORAGE_HIDE_BALANCES) === 'true';
	}
	return false;
}

export const hideBalances = writable(getInitialHideBalances());

// Only subscribe to localStorage updates in the browser
if (typeof window !== 'undefined') {
	hideBalances.subscribe((value) => {
		localStorage.setItem(LOCAL_STORAGE_HIDE_BALANCES, value.toString());
	});
}
