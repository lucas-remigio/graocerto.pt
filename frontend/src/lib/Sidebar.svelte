<!-- src/lib/Sidebar.svelte — visible only on lg+ screens; expands on hover (Instagram-style) -->
<script lang="ts">
	import { page } from '$app/stores';
	import { LogOut, LogIn, User } from 'lucide-svelte';
	import { t, locale } from '$lib/i18n';
	import { isAuthenticated, logout, userEmail } from '$stores/auth';
	import NavActions from './NavActions.svelte';
	import ProfileModal from './ProfileModal.svelte';
	import { theme } from '$stores/uiPreferences';
	import { navLinks, toggleTheme, toggleLanguage } from '$lib/nav';
	import { unreadNotificationCount } from '$stores/notifications';

	let showProfileModal = false;

	// The rail is collapsed at rest and expands on hover / keyboard focus,
	// overlaying the page content instead of pushing it.
	let hovered = false;
	$: collapsed = !hovered;

	function handleFocusOut(event: FocusEvent) {
		const next = event.relatedTarget as Node | null;
		if (!event.currentTarget || !(event.currentTarget as HTMLElement).contains(next)) {
			hovered = false;
		}
	}
</script>

<aside
	class="fixed inset-y-0 left-0 z-40 hidden flex-col overflow-x-clip border-r border-base-300 bg-base-100 transition-[width] duration-200 ease-in-out lg:flex
		{collapsed ? 'w-16' : 'w-64 shadow-xl'}"
	on:mouseenter={() => (hovered = true)}
	on:mouseleave={() => (hovered = false)}
	on:focusin={() => (hovered = true)}
	on:focusout={handleFocusOut}
>
	<!-- Logo -->
	<div class="flex h-16 shrink-0 items-center border-b border-base-300 {collapsed ? 'justify-center px-0' : 'px-4'}">
		<a
			href={$isAuthenticated ? '/home' : '/'}
			class="flex items-center gap-3 overflow-hidden text-xl font-semibold"
		>
			<img src="/logo.png" alt="Logo" class="h-8 w-8 shrink-0" />
			{#if !collapsed}
				<span class="whitespace-nowrap">Grão Certo</span>
			{/if}
		</a>
	</div>

	<!-- Navigation links (authenticated only) -->
	{#if $isAuthenticated}
		<nav class="flex-1 px-2 py-4 {collapsed ? 'overflow-y-visible' : 'overflow-y-auto'}">
			<ul class="space-y-1">
				{#each navLinks as link}
					{@const active = $page.url.pathname === link.href}
					<li>
						{#if collapsed}
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
			class="mb-2 flex {collapsed
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
				tooltipDir={collapsed ? 'right' : 'bottom'}
			/>
		</div>

		<!-- User section -->
		{#if $isAuthenticated}
			<div class="space-y-1 border-t border-base-300 pt-2">
				{#if collapsed}
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
		{:else if collapsed}
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
