<script lang="ts">
	import '../app.css';
	import Navbar from '$lib/Navbar.svelte';
	import { afterNavigate, goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { isAuthenticated, token, authHydrated } from '$lib/stores/auth';
	import axios from '$lib/axios';
	import { get } from 'svelte/store';
	import { onMount, onDestroy } from 'svelte';
	import { isLoading, setupI18n, t, i18nReady } from '$lib/i18n';
	import Footer from '$lib/Footer.svelte';
	import { cookieConsent } from '$lib/stores/cookieConsent';
	import CookieBanner from '$components/CookieBanner.svelte';

	let { children } = $props();

	const publicRoutes = ['/login', '/register', '/'];

	// add app-wide loading state - using $state() for Svelte 5 reactivity
	let appReady = $state(false);

	// Prevent multiple simultaneous auth checks
	let authCheckInProgress = false;
	let authCheckTimeoutId: ReturnType<typeof setTimeout> | null = null;

	async function checkAuth(currentPath: string) {
		if (!browser) return; // Ensure this logic runs only in the browser

		// Wait for auth to be hydrated from localStorage
		if (!get(authHydrated)) {
			return;
		}

		// Prevent duplicate auth checks
		if (authCheckInProgress) {
			return;
		}

		authCheckInProgress = true;

		const isPublicRoute = publicRoutes.includes(currentPath);
		const authToken = get(token);

		if (isPublicRoute) {
			// If user is authenticated but on a public route (login/register), redirect to dashboard
			authCheckInProgress = false;
			return;
		}

		try {
			if (!authToken) {
				throw new Error('No token found');
			}

			// Verify the token with the backend
			await axios.get('/verify-token');

			isAuthenticated.set(true); // Token is valid
		} catch (error: any) {
			// Only redirect for actual authentication failures (401, invalid token, etc.)
			if (error.response?.status === 401) {
				goto('/login');
			}
		} finally {
			authCheckInProgress = false;
		}
	}

	function debouncedCheckAuth(currentPath: string) {
		// Clear any existing timeout
		if (authCheckTimeoutId) {
			clearTimeout(authCheckTimeoutId);
		}

		// Set a new timeout to debounce rapid calls
		authCheckTimeoutId = setTimeout(() => {
			checkAuth(currentPath);
			authCheckTimeoutId = null;
		}, 50); // 50ms debounce
	}

	if (browser) {
		// Wait for auth hydration before checking auth on initial load
		authHydrated.subscribe((hydrated) => {
			if (hydrated) {
				debouncedCheckAuth(window.location.pathname);
			}
		});

		// Run on every navigation (auth will already be hydrated by then)
		afterNavigate((navigation) => {
			debouncedCheckAuth(navigation.to?.url.pathname || '/');
		});
	}

	onDestroy(() => {
		// Cleanup timeout if component is destroyed
		if (authCheckTimeoutId) {
			clearTimeout(authCheckTimeoutId);
		}
	});

	onMount(() => {
		setupI18n();

		// Wait for i18n to be fully ready (this covers both loading states)
		let unsubscribe: (() => void) | undefined;
		unsubscribe = i18nReady.subscribe((ready) => {
			if (ready) {
				appReady = true;
				unsubscribe?.();
			}
		});

		// Initialize cookie consent store globally
		cookieConsent.init();
	});
</script>

{#if !appReady}
	<div class="bg-neutral flex h-screen items-center justify-center">
		<div class="text-center">
			<span class="loading loading-spinner loading-lg text-primary"></span>
		</div>
	</div>
{:else}
	<Navbar />
	<main class="min-h-screen">
		{@render children()}
	</main>
	<Footer />

	<CookieBanner />
{/if}
