<script lang="ts">
	import '../app.css';
	import Navbar from '$lib/Navbar.svelte';
	import { afterNavigate, goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { isAuthenticated, token } from '$lib/stores/auth';
	import axios from '$lib/axios';
	import { get } from 'svelte/store';
	import { onMount } from 'svelte';
	import { isLoading, setupI18n } from '$lib/i18n';

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
		setupI18n();
		// Run on initial load
		checkAuth(window.location.pathname);

		// Run on every navigation
		afterNavigate((navigation) => {
			checkAuth(navigation.to?.url.pathname || '/');
		});
	}

	onMount(() => {
		if (!browser) {
			setupI18n();
		}
	});
</script>

<Navbar />

<main>
	{@render children()}
</main>
