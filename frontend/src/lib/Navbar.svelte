<script lang="ts">
	import { goto } from '$app/navigation';
	import { LogIn, LogOut, Menu, Moon, Sun, User } from 'lucide-svelte';
	import { t, locale, setLocale } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { isAuthenticated } from './stores/auth';

	let isDropdownOpen = false;
	let categoriesUrl = '/categories';
	let investmentCalculatorUrl = '/investment-calculator';
	let loginUrl = '/login';
	let homeUrl = '/home';

	// Track theme state
	let theme: 'light' | 'dark' = 'light';

	// Flag to indicate that a touch event already handled the toggle
	let touchHandled = false;

	// Profile dropdown state
	let isProfileDropdownOpen = false;

	function toggleLanguage() {
		const newLang = $locale === 'en' ? 'pt' : 'en';
		$locale = newLang;
		setLocale(newLang);
	}

	const logout = async () => {
		localStorage.removeItem('token');
		document.cookie = 'authToken=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';

		window.location.href = '/login';
	};

	// Toggle theme function
	function toggleTheme() {
		theme = theme === 'light' ? 'dark' : 'light';
		applyTheme(theme);
		localStorage.setItem('theme', theme);
	}

	// Apply theme to HTML element
	function applyTheme(newTheme: string) {
		document.documentElement.setAttribute('data-theme', newTheme);
		document.documentElement.classList.toggle('dark', newTheme === 'dark');
	}

	function handleNavigation(url: string) {
		isDropdownOpen = false;
		goto(url);
	}

	// Initialize theme on mount
	onMount(() => {
		// Check localStorage first (for user preference)
		const savedTheme = localStorage.getItem('theme');

		if (savedTheme) {
			// User has a preference
			theme = savedTheme as 'light' | 'dark';
			applyTheme(theme);
		} else {
			// No saved preference, check system preference
			const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
			theme = prefersDark ? 'dark' : 'light';
			applyTheme(theme);
		}
	});
</script>

<div class="navbar bg-base-100 border-base-300 h-16 min-h-16 border-b">
	<div class="navbar-start">
		<div class="dropdown relative {isDropdownOpen ? 'dropdown-open' : ''}">
			<button
				type="button"
				class="btn btn-ghost lg:hidden"
				on:touchend={(event) => {
					// Prevent the synthetic click from firing after touchend
					event.preventDefault();
					event.stopPropagation();
					touchHandled = true;
					isDropdownOpen = !isDropdownOpen;
				}}
				on:click={(event) => {
					event.preventDefault();
					event.stopPropagation();
					// If the touch event already toggled the state, ignore this click
					if (touchHandled) {
						touchHandled = false;
						return;
					}
					isDropdownOpen = !isDropdownOpen;
				}}
			>
				<Menu size={20} class="h-5 w-5" />
			</button>
			{#if isDropdownOpen}
				<ul
					class="menu menu-sm dropdown-content bg-base-100 rounded-box z-[50] mt-3 w-52 p-2 shadow"
				>
					<li>
						<button
							type="button"
							on:click={() => handleNavigation(categoriesUrl)}
							class="text-lg"
							aria-label="Categories">{$t('navbar.categories')}</button
						>
					</li>
					<li>
						<button
							type="button"
							on:click={() => handleNavigation(investmentCalculatorUrl)}
							class="text-lg"
							aria-label="Investment Calculator">{$t('navbar.calculator')}</button
						>
					</li>
				</ul>
			{/if}
		</div>
		<a href={homeUrl} class="btn btn-ghost text-xl">
			<img src="/logo.png" alt="Logo" class="mr-2 h-8 w-8" />
			Grão Certo
		</a>
	</div>
	<div class="navbar-center hidden lg:flex">
		<ul class="menu menu-horizontal px-1">
			<li>
				<button
					type="button"
					on:click={() => handleNavigation(categoriesUrl)}
					class="text-lg"
					aria-label="Categories">{$t('navbar.categories')}</button
				>
			</li>
			<li>
				<button
					type="button"
					on:click={() => handleNavigation(investmentCalculatorUrl)}
					class="text-lg"
					aria-label="Investment Calculator">{$t('navbar.calculator')}</button
				>
			</li>
		</ul>
	</div>
	<div class="navbar-end">
		<!-- Language Selector -->
		<div class="dropdown dropdown-end">
			<!-- Language Selector Toggle -->
			<button class="btn btn-ghost" on:click={toggleLanguage}>
				<span class="font-bold">{$locale === 'en' ? '🇬🇧' : '🇵🇹'}</span>
			</button>
		</div>

		<!-- Theme Toggle Button -->
		<button class="btn btn-ghost btn-circle" on:click={toggleTheme} aria-label="Toggle theme">
			{#if theme === 'dark'}
				<Sun size={20} class="h-5 w-5" />
			{:else}
				<Moon size={20} class="h-5 w-5" />
			{/if}
		</button>

		<!-- Profile Dropdown (if authenticated) -->
		{#if $isAuthenticated}
			<div class="dropdown dropdown-end {isProfileDropdownOpen ? 'dropdown-open' : ''}">
				<button
					class="btn btn-ghost btn-circle"
					on:click={() => (isProfileDropdownOpen = !isProfileDropdownOpen)}
					aria-haspopup="true"
					aria-expanded={isProfileDropdownOpen}
					aria-label="User menu"
				>
					<User size={20} class="h-5 w-5" />
				</button>

				{#if isProfileDropdownOpen}
					<ul class="dropdown-content menu bg-base-100 rounded-box z-[100] mt-4 w-64 p-4 shadow">
						<li class="mb-2 flex items-center justify-center">
							<span
								class="text-base-content select-text rounded px-2 py-1 text-center text-sm font-medium"
							>
								{localStorage.getItem('userEmail') || 'unknown@anonymous.pt'}
							</span>
						</li>
						<li class="border-base-200 mt-2 border-t pt-2">
							<button
								class="btn btn-error btn-sm w-full"
								on:click={() => {
									logout();
									isProfileDropdownOpen = false;
								}}
							>
								<LogOut size={18} class="mr-2" />
							</button>
						</li>
					</ul>
				{/if}
			</div>
		{:else}
			<a href={loginUrl} class="btn btn-ghost">
				<div class="flex w-full items-center">
					<LogIn size={20} class="text-base-100 h-5 w-5 flex-shrink-0" />
				</div>
			</a>
		{/if}
	</div>
</div>

<style>
	.dropdown:not(.dropdown-open) .dropdown-content {
		display: none;
	}
</style>
