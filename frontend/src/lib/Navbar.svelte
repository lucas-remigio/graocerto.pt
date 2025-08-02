<script lang="ts">
	import { goto } from '$app/navigation';
	import { Calculator, Home, List, LogIn, Menu } from 'lucide-svelte';
	import { t, locale, setLocale } from '$lib/i18n';
	import { isAuthenticated, logout } from '$stores/auth';
	import UserMenu from './UserMenu.svelte';
	import NavActions from './NavActions.svelte';
	import { theme, type ThemeOption } from '$stores/uiPreferences';
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
	const themeCycle: ThemeOption[] = ['system', 'dark', 'light'];

	function toggleTheme() {
		theme.update((current) => {
			const idx = themeCycle.indexOf(current);
			return themeCycle[(idx + 1) % themeCycle.length];
		});
	}

	function handleNavigation(url: string) {
		isDropdownOpen = false;
		goto(url);
	}
</script>

<div class="navbar bg-base-100 border-base-300 h-16 min-h-16 border-b">
	<div class="navbar-start">
		<!--  If the user is not authenticated and on mobile, we show the logo on the left -->
		{#if !$isAuthenticated}
			<a href="/" class="btn btn-ghost flex items-center gap-2 text-xl lg:hidden">
				<img src="/logo.png" alt="Logo" class="h-8 w-8" />
				Grão Certo
			</a>
			<!-- Hamburger menu only visible on the left if user is logged on -->
		{:else}
			<div class="dropdown relative {isDropdownOpen ? 'dropdown-open' : ''}">
				<button
					type="button"
					class="btn btn-ghost btn-circle"
					on:touchend={(event) => {
						event.preventDefault();
						event.stopPropagation();
						touchHandled = true;
						isDropdownOpen = !isDropdownOpen;
					}}
					on:click={(event) => {
						event.preventDefault();
						event.stopPropagation();
						if (touchHandled) {
							touchHandled = false;
							return;
						}
						isDropdownOpen = !isDropdownOpen;
					}}
					aria-label="Open menu"
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
								on:click={() => handleNavigation(homeUrl)}
								class="text-lg"
								aria-label="Home"
							>
								<Home size={18} class="mr-2" />
								{$t('navbar.home')}
							</button>
						</li>
						<li>
							<button
								type="button"
								on:click={() => handleNavigation(categoriesUrl)}
								class="text-lg"
								aria-label="Categories"
							>
								<List size={18} class="mr-2" />
								{$t('navbar.categories')}
							</button>
						</li>
						<li>
							<button
								type="button"
								on:click={() => handleNavigation(investmentCalculatorUrl)}
								class="text-lg"
								aria-label="Investment Calculator"
							>
								<Calculator size={18} class="mr-2" />
								{$t('navbar.calculator')}
							</button>
						</li>
					</ul>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Navbar Center. If user authenticated will always be centered. If not authenticated, -->
	<!-- will be either hidden or shown based on screen size -->
	<div class="navbar-center flex flex-1 justify-center">
		<a
			href={$isAuthenticated ? homeUrl : '/'}
			class="btn btn-ghost flex items-center gap-2 text-xl {!$isAuthenticated
				? 'hidden lg:flex'
				: ''}"
		>
			<img src="/logo.png" alt="Logo" class="h-8 w-8" />
			Grão Certo
		</a>
	</div>

	<!-- End of the navbar -->
	<div class="navbar-end">
		<!-- Desktop: show theme/language always -->
		<div class="hidden items-center gap-1 lg:flex">
			<NavActions
				theme={$theme}
				{toggleTheme}
				locale={$locale || 'pt'}
				{toggleLanguage}
				t={$t}
				isAuthenticated={$isAuthenticated}
			/>
		</div>

		<!-- Mobile: add to profile dropdown if logged in, else show inline -->
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
						isAuthenticated={$isAuthenticated}
					/>
				{/if}
			</UserMenu>
		{:else}
			<!-- Show theme/language inline on mobile if not logged in -->
			<div class="flex items-center gap-1 lg:hidden">
				<NavActions theme={$theme} {toggleTheme} locale={$locale || 'pt'} {toggleLanguage} t={$t} />
			</div>
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
