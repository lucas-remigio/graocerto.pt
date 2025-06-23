import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Initialize stores with null - will be hydrated on client
export const token = writable<string | null>(null);
export const userEmail = writable<string | null>(null);
export const isAuthenticated = writable<boolean>(false);
export const authHydrated = writable<boolean>(false);

// Hydrate from localStorage on client side
if (browser) {
	const storedToken = localStorage.getItem('token');
	const storedEmail = localStorage.getItem('userEmail');

	if (storedToken) {
		token.set(storedToken);
	}
	if (storedEmail) {
		userEmail.set(storedEmail);
	}

	// Mark auth as hydrated after setting initial values
	authHydrated.set(true);
}

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

// Update localStorage whenever userEmail changes
userEmail.subscribe((value) => {
	if (typeof localStorage !== 'undefined') {
		if (value) {
			localStorage.setItem('userEmail', value);
		} else {
			localStorage.removeItem('userEmail');
		}
	}
});

// Helper function to set both token and email when user logs in
export function login(newToken: string, email: string) {
	token.set(newToken);
	userEmail.set(email);
}

// Helper function to clear both token and email when user logs out
export function logout() {
	token.set(null);
	userEmail.set(null);
}
