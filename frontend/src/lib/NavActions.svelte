<script lang="ts">
	import { Eye, EyeOff, MonitorSmartphone, Moon, Phone, Sun } from 'lucide-svelte';
	import { hideBalances, type ThemeOption } from '$stores/uiPreferences';
	export let theme: ThemeOption = 'system';
	export let toggleTheme: () => void;
	export let locale: string;
	export let toggleLanguage: () => void;
	export let t: (key: string) => string;
	export let isMenu = false; // If true, render as menu items (li), else as inline buttons
	export let isAuthenticated: boolean = false;

	const themeCycle: ThemeOption[] = ['system', 'dark', 'light'];

	function getNextTheme(current: ThemeOption): ThemeOption {
		const idx = themeCycle.indexOf(current);
		return themeCycle[(idx + 1) % themeCycle.length];
	}

	function getThemeIcon() {
		switch (theme) {
			case 'dark':
				return Moon;
			case 'light':
				return Sun;
			case 'system':
				return MonitorSmartphone;
			default:
				return MonitorSmartphone;
		}
	}

	function getNextThemeLabel() {
		const next = getNextTheme(theme);
		switch (next) {
			case 'dark':
				return t('navbar.switch-to-dark');
			case 'light':
				return t('navbar.switch-to-light');
			case 'system':
				return t('navbar.switch-to-system');
			default:
				return t('navbar.switch-to-system');
		}
	}

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
			icon: () => getThemeIcon(),
			text: () => getNextThemeLabel(),
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
				<span class="flex items-center">
					{#if action.icon}
						<svelte:component this={action.icon()} size={18} class="mr-2" />
					{/if}
					{#if action.emoji}
						<span class="mr-2">{action.emoji()}</span>
					{/if}
					<span class="whitespace-normal">{action.text()}</span>
				</span>
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
