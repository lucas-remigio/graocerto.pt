<!-- src/lib/Sidebar.svelte — visible only on lg+ screens -->
<script lang="ts">
	import { page } from '$app/stores';
	import { ChevronLeft, ChevronRight, LogOut, LogIn, User } from 'lucide-svelte';
	import { t, locale } from '$lib/i18n';
	import { isAuthenticated, logout, userEmail } from '$stores/auth';
	import NavActions from './NavActions.svelte';
	import ProfileModal from './ProfileModal.svelte';
	import { theme, sidebarCollapsed, updateSidebarCollapsed } from '$stores/uiPreferences';
	import { navLinks, toggleTheme, toggleLanguage } from '$lib/nav';
	import { unreadNotificationCount } from '$stores/notifications';

	let showProfileModal = false;

	function toggle() {
		updateSidebarCollapsed(!$sidebarCollapsed);
	}
</script>

<aside
	class="fixed inset-y-0 left-0 z-40 hidden flex-col overflow-x-clip border-r border-base-300 bg-base-100 transition-[width] duration-200 ease-in-out lg:flex
		{$sidebarCollapsed ? 'w-16' : 'w-64'}"
>
	<!-- Logo + collapse toggle -->
	<div
		class="flex shrink-0 items-center border-b border-base-300
		{$sidebarCollapsed ? 'h-auto flex-col gap-1 px-0 py-3' : 'h-16 justify-between px-4'}"
	>
		{#if !$sidebarCollapsed}
			<a
				href={$isAuthenticated ? '/home' : '/'}
				class="flex items-center gap-3 text-xl font-semibold"
			>
				<img src="/logo.png" alt="Logo" class="h-8 w-8 shrink-0" />
				Grão Certo
			</a>
			<button
				class="btn btn-square btn-ghost btn-sm shrink-0"
				on:click={toggle}
				aria-label="Collapse sidebar"
			>
				<ChevronLeft size={18} />
			</button>
		{:else}
			<a
				href={$isAuthenticated ? '/home' : '/'}
				class="tooltip tooltip-right flex items-center justify-center"
				data-tip="Grão Certo"
			>
				<img src="/logo.png" alt="Logo" class="h-8 w-8 shrink-0" />
			</a>
			<button class="btn btn-square btn-ghost btn-sm" on:click={toggle} aria-label="Expand sidebar">
				<ChevronRight size={18} />
			</button>
		{/if}
	</div>

	<!-- Navigation links (authenticated only) -->
	{#if $isAuthenticated}
		<nav class="flex-1 px-2 py-4 {$sidebarCollapsed ? 'overflow-y-visible' : 'overflow-y-auto'}">
			<ul class="space-y-1">
				{#each navLinks as link}
					{@const active = $page.url.pathname === link.href}
					<li>
						{#if $sidebarCollapsed}
							<a
								href={link.href}
								class="tooltip tooltip-right relative flex items-center justify-center rounded-lg p-2.5 transition-colors
									{active
									? 'bg-primary/10 text-primary'
									: 'text-base-content/70 hover:bg-base-200 hover:text-base-content'}"
								data-tip={$t(link.labelKey)}
							>
								<svelte:component this={link.icon} size={20} />
								{#if link.href === '/notifications' && $unreadNotificationCount > 0}
									<span class="absolute right-[2px] top-[2px] h-2.5 w-2.5 rounded-full bg-error"
									></span>
								{/if}
							</a>
						{:else}
							<a
								href={link.href}
								class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors
									{active
									? 'bg-primary/10 text-primary'
									: 'text-base-content/70 hover:bg-base-200 hover:text-base-content'}"
							>
								<div class="relative">
									<svelte:component this={link.icon} size={18} />
									{#if link.href === '/notifications' && $unreadNotificationCount > 0}
										<span class="absolute -right-1 -top-1 h-2.5 w-2.5 rounded-full bg-error"></span>
									{/if}
								</div>
								{$t(link.labelKey)}
								{#if link.href === '/notifications' && $unreadNotificationCount > 0}
									<span class="badge badge-error badge-sm ml-auto text-white"
										>{$unreadNotificationCount}</span
									>
								{/if}
							</a>
						{/if}
					</li>
				{/each}
			</ul>
		</nav>
	{:else}
		<div class="flex-1"></div>
	{/if}

	<!-- Bottom section: actions + user -->
	<div class="shrink-0 border-t border-base-300 px-2 py-3">
		<!-- Theme / Language / Balance icon buttons -->
		<div
			class="mb-2 flex {$sidebarCollapsed
				? 'flex-col items-center gap-1'
				: 'items-center justify-around'}"
		>
			<NavActions
				theme={$theme}
				{toggleTheme}
				locale={$locale || 'pt'}
				{toggleLanguage}
				t={$t}
				isAuthenticated={$isAuthenticated}
				tooltipDir={$sidebarCollapsed ? 'right' : 'bottom'}
			/>
		</div>

		<!-- User section -->
		{#if $isAuthenticated}
			<div class="space-y-1 border-t border-base-300 pt-2">
				{#if $sidebarCollapsed}
					<button
						class="tooltip tooltip-right flex w-full items-center justify-center rounded-lg p-2 transition-colors hover:bg-base-200"
						on:click={() => (showProfileModal = true)}
						data-tip={$userEmail ?? 'Profile'}
					>
						<User size={18} class="shrink-0 text-base-content/60" />
					</button>
					<button
						class="tooltip tooltip-right flex w-full items-center justify-center rounded-lg p-2 text-error transition-colors hover:bg-error/10"
						on:click={logout}
						data-tip={$t('auth.logout', { default: 'Logout' })}
					>
						<LogOut size={18} class="shrink-0" />
					</button>
				{:else}
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
				{/if}
			</div>
		{:else if $sidebarCollapsed}
			<a
				href="/login"
				class="btn btn-primary text-base-100 btn-sm tooltip tooltip-right flex w-full items-center justify-center"
				data-tip={$t('auth.login', { default: 'Login' })}
			>
				<LogIn size={16} />
			</a>
		{:else}
			<a href="/login" class="btn btn-primary btn-sm w-full text-base-100">
				<LogIn size={16} />
				{$t('auth.login', { default: 'Login' })}
			</a>
		{/if}
	</div>
</aside>

<ProfileModal bind:showModal={showProfileModal} {logout} />
