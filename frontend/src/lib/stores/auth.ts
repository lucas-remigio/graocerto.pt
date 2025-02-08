import { writable } from 'svelte/store';

// Writable store to store the token
// Try to get an initial token from localStorage if it exists
export const storedToken =
	typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;

export const token = writable<string | null>(storedToken);

// Writable store to manage authentication state
export const isAuthenticated = writable<boolean>(false);

// Update the `isAuthenticated` state whenever the token changes
token.subscribe((value) => {
	isAuthenticated.set(!!value);

	if (typeof localStorage !== 'undefined') {
		if (value) {
			localStorage.setItem('token', value);
		} else {
			localStorage.removeItem('token');
		}
	}
});
