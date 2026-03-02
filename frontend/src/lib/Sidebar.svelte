<!-- src/lib/Sidebar.svelte — visible only on lg+ screens -->
<script lang="ts">
	import { page } from '$app/stores';
	import { Calculator, Home, List, LogOut, User } from 'lucide-svelte';
	import { t, locale } from '$lib/i18n';
	import { isAuthenticated, logout, userEmail } from '$stores/auth';
	import NavActions from './NavActions.svelte';
	import ProfileModal from './ProfileModal.svelte';
	import { theme } from '$stores/uiPreferences';
	import { navLinks, toggleTheme, toggleLanguage } from '$lib/nav';

	let showProfileModal = false;
</script>

<aside
	class="fixed inset-y-0 left-0 z-40 hidden w-64 flex-col border-r border-base-300 bg-base-100 lg:flex"
>
	<!-- Logo -->
	<a
		href={$isAuthenticated ? '/home' : '/'}
		class="flex h-16 shrink-0 items-center gap-3 border-b border-base-300 px-6 text-xl font-semibold"
	>
		<img src="/logo.png" alt="Logo" class="h-8 w-8" />
		Grão Certo
	</a>

	<!-- Navigation links (authenticated only) -->
	{#if $isAuthenticated}
		<nav class="flex-1 overflow-y-auto px-3 py-4">
			<ul class="space-y-1">
				{#each navLinks as link}
					{@const active = $page.url.pathname === link.href}
					<li>
						<a
							href={link.href}
							class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors
								{active
								? 'bg-primary/10 text-primary'
								: 'text-base-content/70 hover:bg-base-200 hover:text-base-content'}"
						>
							<svelte:component this={link.icon} size={18} />
							{$t(link.labelKey)}
						</a>
					</li>
				{/each}
			</ul>
		</nav>
	{:else}
		<div class="flex-1"></div>
	{/if}

	<!-- Bottom section: actions + user -->
	<div class="shrink-0 border-t border-base-300 px-3 py-4">
		<!-- Theme / Language / Balance as a row of icon buttons -->
		<div class="mb-3 flex items-center justify-around">
			<NavActions
				theme={$theme}
				{toggleTheme}
				locale={$locale || 'pt'}
				{toggleLanguage}
				t={$t}
				isAuthenticated={$isAuthenticated}
			/>
		</div>

		<!-- User section -->
		{#if $isAuthenticated}
			<div class="space-y-1 border-t border-base-300 pt-3">
				<button
					class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors hover:bg-base-200"
					on:click={() => (showProfileModal = true)}
				>
					<User size={18} class="shrink-0 text-base-content/60" />
					<span class="truncate text-left text-base-content/70">
						{$userEmail ?? 'Profile'}
					</span>
				</button>
				<button
					class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-error transition-colors hover:bg-error/10"
					on:click={logout}
				>
					<LogOut size={18} class="shrink-0" />
					{$t('auth.logout', { default: 'Logout' })}
				</button>
			</div>
		{:else}
			<a href="/login" class="btn btn-primary btn-sm w-full">
				{$t('auth.login', { default: 'Login' })}
			</a>
		{/if}
	</div>
</aside>

<ProfileModal bind:showModal={showProfileModal} {logout} />
