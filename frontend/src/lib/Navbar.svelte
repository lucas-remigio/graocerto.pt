<script lang="ts">
	import { goto } from '$app/navigation';
	import { LogIn, Menu } from 'lucide-svelte';
	import { t, locale } from '$lib/i18n';
	import { isAuthenticated, logout } from '$stores/auth';
	import UserMenu from './UserMenu.svelte';
	import NavActions from './NavActions.svelte';
	import { theme } from '$stores/uiPreferences';
	import { isLargeScreen } from '$stores/screen';
	import { unreadNotificationCount } from '$stores/notifications';
	import { navLinks, toggleTheme, toggleLanguage } from '$lib/nav';

	let isDropdownOpen = false;

	let loginUrl = '/login';
	let homeUrl = '/home';

	// Flag to indicate that a touch event already handled the toggle
	let touchHandled = false;

	function handleNavigation(url: string) {
		isDropdownOpen = false;
		goto(url);
	}
</script>

<div class="navbar h-16 min-h-16 border-b border-base-300 bg-base-100 lg:hidden">
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
					class="btn btn-circle btn-ghost"
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
					<div class="indicator relative flex items-center justify-center">
						<Menu size={20} class="h-5 w-5" />
						{#if $unreadNotificationCount > 0}
							<span class="badge indicator-item badge-error badge-xs"></span>
						{/if}
					</div>
				</button>
				{#if isDropdownOpen}
					<ul
						class="menu dropdown-content menu-sm z-[50] mt-3 w-52 rounded-box bg-base-100 p-2 shadow"
					>
						{#each navLinks as link}
							<li>
								<button
									type="button"
									on:click={() => handleNavigation(link.href)}
									class="text-lg"
									aria-label={$t(link.labelKey)}
								>
									<div class="relative flex items-center justify-center">
										<svelte:component this={link.icon} size={18} class="mr-2" />
										{#if link.href === '/notifications' && $unreadNotificationCount > 0}
											<span class="absolute -top-1 right-1 h-2 w-2 rounded-full bg-error"></span>
										{/if}
									</div>
									{$t(link.labelKey)}
									{#if link.href === '/notifications' && $unreadNotificationCount > 0}
										<span class="badge badge-error ml-auto text-xs text-white"
											>{$unreadNotificationCount}</span
										>
									{/if}
								</button>
							</li>
						{/each}
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
			<a href={loginUrl} class="btn btn-circle btn-ghost" aria-label="Login">
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
