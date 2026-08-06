<script lang="ts">
	import { onMount } from 'svelte';
	import { Send, Copy, Check, Unlink } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { dataService } from '$lib/services/dataService';

	// Set to the handle you registered with BotFather.
	const botUsername = import.meta.env.VITE_TELEGRAM_BOT_USERNAME || 'GraoCertoBot';

	let linked = false;
	let loading = true;
	let working = false;
	let error = '';
	let code = '';
	let expiresAt: Date | null = null;
	let copied = false;

	$: command = code ? `/link ${code}` : '';

	onMount(loadStatus);

	async function loadStatus() {
		loading = true;
		error = '';
		try {
			linked = await dataService.fetchTelegramStatus();
		} catch (err) {
			console.error('Failed to load telegram status:', err);
			error = $t('telegram.error', { default: 'Something went wrong. Please try again.' });
		} finally {
			loading = false;
		}
	}

	async function generateCode() {
		working = true;
		error = '';
		copied = false;
		try {
			const response = await dataService.createTelegramLinkToken();
			code = response.token;
			expiresAt = new Date(response.expires_at);
		} catch (err) {
			console.error('Failed to create telegram link code:', err);
			error = $t('telegram.error', { default: 'Something went wrong. Please try again.' });
		} finally {
			working = false;
		}
	}

	async function copyCommand() {
		try {
			await navigator.clipboard.writeText(command);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch (err) {
			console.error('Failed to copy the link command:', err);
		}
	}

	async function unlink() {
		working = true;
		error = '';
		try {
			await dataService.unlinkTelegram();
			linked = false;
			code = '';
			expiresAt = null;
		} catch (err) {
			console.error('Failed to unlink telegram:', err);
			error = $t('telegram.error', { default: 'Something went wrong. Please try again.' });
		} finally {
			working = false;
		}
	}

	function formatExpiry(date: Date): string {
		return date.toLocaleTimeString(
			localStorage.getItem('preferred-language') === 'en' ? 'en-GB' : 'pt-PT',
			{ hour: '2-digit', minute: '2-digit' }
		);
	}
</script>

<div class="rounded-lg bg-base-200 p-4">
	<div class="mb-2 flex items-center gap-2">
		<Send class="h-4 w-4 text-primary" />
		<span class="text-sm font-medium">{$t('telegram.title', { default: 'Telegram' })}</span>
		{#if !loading && linked}
			<span class="badge badge-success badge-sm">
				{$t('telegram.linked', { default: 'Linked' })}
			</span>
		{/if}
	</div>

	{#if loading}
		<span class="loading loading-spinner loading-sm"></span>
	{:else if linked}
		<p class="mb-3 text-sm text-base-content/80">
			{$t('telegram.linked-description', {
				default: 'Send a message like "3.19 cookies" to the bot to add a transaction.'
			})}
		</p>
		<button class="btn btn-outline btn-sm" on:click={unlink} disabled={working}>
			{#if working}
				<span class="loading loading-spinner loading-xs"></span>
			{:else}
				<Unlink class="mr-1 h-4 w-4" />
			{/if}
			{$t('telegram.unlink', { default: 'Unlink' })}
		</button>
	{:else if code}
		<p class="mb-2 text-sm text-base-content/80">
			{$t('telegram.instructions', { default: 'Send this message to' })}
			<span class="font-medium">@{botUsername}</span>:
		</p>
		<div class="mb-2 flex items-center gap-2">
			<code class="flex-1 rounded bg-base-300 px-3 py-2 font-mono text-sm">{command}</code>
			<button class="btn btn-square btn-ghost btn-sm" on:click={copyCommand}>
				{#if copied}
					<Check class="h-4 w-4 text-success" />
				{:else}
					<Copy class="h-4 w-4" />
				{/if}
			</button>
		</div>
		{#if expiresAt}
			<p class="mb-3 text-xs text-base-content/60">
				{$t('telegram.expires', { default: 'Valid until' })}
				{formatExpiry(expiresAt)}
			</p>
		{/if}
		<button class="btn btn-ghost btn-sm" on:click={loadStatus} disabled={working}>
			{$t('telegram.check-status', { default: "I've sent it" })}
		</button>
	{:else}
		<p class="mb-3 text-sm text-base-content/80">
			{$t('telegram.description', {
				default: 'Add transactions by texting a bot, in your own words.'
			})}
		</p>
		<button class="btn btn-outline btn-sm" on:click={generateCode} disabled={working}>
			{#if working}
				<span class="loading loading-spinner loading-xs"></span>
			{/if}
			{$t('telegram.connect', { default: 'Connect Telegram' })}
		</button>
	{/if}

	{#if error}
		<p class="mt-2 text-sm text-error">{error}</p>
	{/if}
</div>
