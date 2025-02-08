<script lang="ts">
	import '../app.css';
	import Navbar from '$lib/Navbar.svelte';
	import { afterNavigate, goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { isAuthenticated, storedToken, token } from '$lib/stores/auth';
	import axios from '$lib/axios';
	import { get } from 'svelte/store';

	let { children } = $props();

	const publicRoutes = ['/login', '/register'];

	async function checkAuth(currentPath: string) {
		if (!browser) return; // Ensure this logic runs only in the browser

		const isPublicRoute = publicRoutes.includes(currentPath);
		const authToken = get(token);

		if (isPublicRoute) {
			return;
		}

		try {
			if (!authToken) {
				throw new Error('No token found');
			}

			// Verify the token with the backend
			await axios.get('/verify-token');

			isAuthenticated.set(true); // Token is valid
		} catch (error) {
			console.error('Token verification failed:', error);
			isAuthenticated.set(false); // Token is invalid
			goto('/login'); // Redirect to login if verification fails
		}
	}

	if (browser) {
		// Run on initial load
		checkAuth(window.location.pathname);

		// Run on every navigation
		afterNavigate((navigation) => {
			checkAuth(navigation.to?.url.pathname || '/');
		});
	}
</script>

<Navbar />

<main>
	{@render children()}
</main>
