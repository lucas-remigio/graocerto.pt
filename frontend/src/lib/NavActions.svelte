<script lang="ts">
	import { Eye, EyeOff, Moon, Sun } from 'lucide-svelte';
	import { hideBalances } from '$stores/uiPreferences';
	export let theme: 'light' | 'dark';
	export let toggleTheme: () => void;
	export let locale: string;
	export let toggleLanguage: () => void;
	export let t: (key: string) => string;
	export let isMenu = false; // If true, render as menu items (li), else as inline buttons
	export let isAuthenticated: boolean = false;

	// Helper for icon and text
	$: actions = [
		{
			key: 'balance',
			onClick: () => hideBalances.update((v) => !v),
			icon: () => ($hideBalances ? EyeOff : Eye),
			text: () => t($hideBalances ? 'navbar.show-balance' : 'navbar.hide-balance'),
			show: isAuthenticated
		},
		{
			key: 'theme',
			onClick: toggleTheme,
			icon: () => (theme === 'dark' ? Sun : Moon),
			text: () => t(theme === 'dark' ? 'navbar.light-mode' : 'navbar.dark-mode'),
			show: true
		},
		{
			key: 'language',
			onClick: toggleLanguage,
			icon: null,
			text: () => t('navbar.language'),
			emoji: () => (locale === 'en' ? '🇬🇧' : '🇵🇹'),
			show: true
		}
	];
</script>

{#if isMenu}
	{#each actions.filter((a) => a.show) as action (action.key)}
		<li>
			<button class="btn btn-ghost w-full justify-start font-normal" on:click={action.onClick}>
				{#if action.icon}
					<svelte:component this={action.icon()} size={18} class="mr-2" />
				{:else if action.emoji}
					<span class="mr-2">{action.emoji()}</span>
				{/if}
				{action.text()}
			</button>
		</li>
	{/each}
{:else}
	{#each actions.filter((a) => a.show) as action (action.key)}
		<button
			class="btn btn-ghost btn-circle tooltip tooltip-bottom flex items-center justify-center"
			on:click={action.onClick}
			aria-label={action.text()}
			data-tip={action.text()}
		>
			{#if action.icon}
				<svelte:component this={action.icon()} size={20} />
			{:else if action.emoji}
				<span class="flex h-full w-full items-center justify-center text-xl">{action.emoji()}</span>
			{/if}
		</button>
	{/each}
{/if}
