import { writable } from 'svelte/store';

function getIsLargeScreen() {
	if (typeof window !== 'undefined') {
		return window.innerWidth >= 1024;
	}
	return false;
}

export const isLargeScreen = writable(getIsLargeScreen());

if (typeof window !== 'undefined') {
	const update = () => isLargeScreen.set(getIsLargeScreen());
	window.addEventListener('resize', update);
}
