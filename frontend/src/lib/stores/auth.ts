import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { goto } from '$app/navigation';

// Initialize stores with null - will be hydrated on client
export const token = writable<string | null>(null);
export const userEmail = writable<string | null>(null);
export const userCreatedAt = writable<string | null>(null);
export const isAuthenticated = writable<boolean>(false);
export const authHydrated = writable<boolean>(false);

// Hydrate from localStorage on client side
if (browser) {
	const storedToken = localStorage.getItem('token');
	const storedEmail = localStorage.getItem('userEmail');
	const storedCreated = localStorage.getItem('userCreated');

	if (storedToken) {
		token.set(storedToken);
	}
	if (storedEmail) {
		userEmail.set(storedEmail);
	}
	if (storedCreated) {
		userCreatedAt.set(storedCreated);
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

// Update localStorage whenever userCreatedAt changes
userCreatedAt.subscribe((value) => {
	if (typeof localStorage !== 'undefined') {
		if (value) {
			localStorage.setItem('userCreated', value);
		} else {
			localStorage.removeItem('userCreated');
		}
	}
});

// Helper function to set both token and email when user logs in
export function login(newToken: string, email: string, createdAt?: string | null) {
	token.set(newToken);
	userEmail.set(email);
	if (createdAt) userCreatedAt.set(createdAt);
}

// Helper function to clear both token and email when user logs out
export function logout() {
	token.set(null);
	userEmail.set(null);
	userCreatedAt.set(null);

	// Remove auth cookie
	if (typeof document !== 'undefined') {
		document.cookie = 'authToken=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
	}

	// Redirect to login page
	goto('/login');
}
