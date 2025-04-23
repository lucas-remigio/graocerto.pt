import { writable } from 'svelte/store';

// Writable store to store the token
// Try to get an initial token from localStorage if it exists
export const storedToken =
	typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
const storedEmail = typeof localStorage !== 'undefined' ? localStorage.getItem('userEmail') : null;

export const token = writable<string | null>(storedToken);
export const userEmail = writable<string | null>(storedEmail);

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

	console.log('Token set:', newToken);
	console.log('Email set:', email);
}

// Helper function to clear both token and email when user logs out
export function logout() {
	token.set(null);
	userEmail.set(null);
}
