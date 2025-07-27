<script lang="ts">
	import { goto } from '$app/navigation';
	import { LogIn, Menu } from 'lucide-svelte';
	import { t, locale, setLocale } from '$lib/i18n';
	import { isAuthenticated, logout } from '$stores/auth';
	import UserMenu from './UserMenu.svelte';
	import NavActions from './NavActions.svelte';
	import { theme } from '$stores/uiPreferences';
	import { isLargeScreen } from '$stores/screen';

	let isDropdownOpen = false;

	let categoriesUrl = '/categories';
	let investmentCalculatorUrl = '/investment-calculator';
	let loginUrl = '/login';
	let homeUrl = '/home';

	// Flag to indicate that a touch event already handled the toggle
	let touchHandled = false;

	function toggleLanguage() {
		const newLang = $locale === 'en' ? 'pt' : 'en';
		$locale = newLang;
		setLocale(newLang);
	}

	// Toggle theme function
	function toggleTheme() {
		theme.update((current) => (current === 'light' ? 'dark' : 'light'));
	}

	function handleNavigation(url: string) {
		isDropdownOpen = false;
		goto(url);
	}
</script>

<div class="navbar bg-base-100 border-base-300 h-16 min-h-16 border-b">
	<div class="navbar-start">
		{#if $isAuthenticated}
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
		{/if}
	</div>

	{#if $isAuthenticated}
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
	{/if}
	<div class="navbar-end">
		<!-- Desktop: show buttons inline -->
		<div class="hidden items-center gap-1 lg:flex">
			{#if $isAuthenticated}
				<NavActions theme={$theme} {toggleTheme} locale={$locale || 'pt'} {toggleLanguage} t={$t} />
			{/if}
		</div>

		<!-- Mobile: add to profile dropdown -->
		{#if $isAuthenticated}
			<UserMenu {logout}>
				{#if !$isLargeScreen}
					<NavActions
						theme={$theme}
						{toggleTheme}
						locale={$locale || 'pt'}
						{toggleLanguage}
						t={$t}
						isMenu={true}
					/>
				{/if}
			</UserMenu>
		{:else}
			<a href={loginUrl} class="btn btn-ghost btn-circle" aria-label="Login">
				<LogIn size={20} class="h-5 w-5" />
			</a>
		{/if}
	</div>
</div>

<style>
	.dropdown:not(.dropdown-open) .dropdown-content {
		display: none;
	}
</style>
