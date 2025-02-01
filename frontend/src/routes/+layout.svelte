<script lang="ts">
	import '../app.css';
	import Navbar from '$lib/Navbar.svelte';
	import { afterNavigate, goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import { page } from '$app/state';

	let { children } = $props();

	const publicRoutes = ['/login', '/register'];

	function checkAuth(currentPath: string) {
		if (!browser) return; // Ensure this logic runs only in the browser

		const authToken = document.cookie.split('; ').find((row) => row.startsWith('authToken='));
		console.log('Layout: ' + authToken);
		const isAuthenticated = !!authToken;

		const isPublicRoute = publicRoutes.includes(currentPath);

		if (!isAuthenticated && !isPublicRoute) {
			goto('/login');
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
