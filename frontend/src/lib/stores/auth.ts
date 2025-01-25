import { writable } from 'svelte/store';

// Writable store to store the token
export const token = writable<string | null>(null);
export const user = writable<string>('');

// Writable store to manage authentication state
export const isAuthenticated = writable<boolean>(false);

// Update the `isAuthenticated` state whenever the token changes
token.subscribe((value) => {
	isAuthenticated.set(!!value);
});
