<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { t } from '$lib/i18n';
	import { ShieldCheck } from 'lucide-svelte';
	import QRCode from 'qrcode';
	import type { AxiosError } from 'axios';
	import api_axios from './axios';
	import { toastStore } from '$lib/stores/toast';
	import type { TOTPEnableResponse, TOTPSetupResponse, TOTPStatusResponse } from '$lib/types';

	type View = 'loading' | 'email' | 'app' | 'setup' | 'recovery' | 'disable';

	// Enabling or disabling forgets trusted devices server-side; the parent reloads that list.
	const dispatch = createEventDispatcher<{ change: void }>();

	let view: View = 'loading';
	let setup: TOTPSetupResponse | null = null;
	let qrDataUrl = '';
	let recoveryCodes: string[] = [];
	let code = '';
	let isBusy = false;

	onMount(async () => {
		try {
			const response = await api_axios.get<TOTPStatusResponse>('auth/totp');
			view = response.data.enabled ? 'app' : 'email';
		} catch {
			toastStore.error($t('auth.error-occurred'));
			view = 'email';
		}
	});

	function showError(error: unknown) {
		const status = (error as AxiosError).response?.status;
		if (status === 400) {
			toastStore.error($t('profile.two-factor-invalid-code'));
		} else if (status === 409) {
			toastStore.error($t('auth.code-already-used'));
		} else if (status === 503) {
			toastStore.error($t('profile.two-factor-unavailable'));
		} else {
			toastStore.error($t('auth.error-occurred'));
		}
	}

	async function run(action: () => Promise<void>) {
		isBusy = true;
		try {
			await action();
		} catch (error) {
			showError(error);
		} finally {
			isBusy = false;
		}
	}

	const startSetup = () =>
		run(async () => {
			const response = await api_axios
				.post<TOTPSetupResponse>('auth/totp/setup')
				.catch((error: AxiosError) => {
					// 409 here means it was already enabled (e.g. in another tab): just show that.
					if (error.response?.status === 409) return null;
					throw error;
				});
			if (!response) {
				view = 'app';
				return;
			}
			setup = response.data;
			qrDataUrl = await QRCode.toDataURL(setup.uri, { margin: 1, width: 200 });
			code = '';
			view = 'setup';
		});

	const confirmEnable = () =>
		run(async () => {
			const response = await api_axios.post<TOTPEnableResponse>('auth/totp/enable', { code });
			recoveryCodes = response.data.recovery_codes;
			setup = null;
			view = 'recovery';
			dispatch('change');
		});

	const confirmDisable = () =>
		run(async () => {
			await api_axios.post('auth/totp/disable', { code });
			view = 'email';
			dispatch('change');
		});

	function openView(next: View) {
		code = '';
		view = next;
	}

	async function copyRecoveryCodes() {
		try {
			await navigator.clipboard.writeText(recoveryCodes.join('\n'));
			toastStore.success($t('profile.two-factor-copied'));
		} catch {
			toastStore.error($t('auth.error-occurred'));
		}
	}

	// Groups of 4 are much easier to type into an app by hand.
	$: groupedSecret = setup?.secret.match(/.{1,4}/g)?.join(' ') ?? '';
</script>

<div class="rounded-lg bg-base-200 p-4">
	<div class="mb-1 flex items-center gap-2">
		<ShieldCheck class="h-4 w-4 text-primary" />
		<span class="text-sm font-medium">{$t('profile.two-factor-title')}</span>
	</div>

	{#if view === 'loading'}
		<span class="loading loading-spinner loading-sm"></span>
	{:else if view === 'email'}
		<p class="mb-3 text-sm text-base-content/70">{$t('profile.two-factor-email')}</p>
		<button
			class="btn btn-outline btn-primary btn-sm w-full"
			disabled={isBusy}
			on:click={startSetup}
		>
			{#if isBusy}<span class="loading loading-spinner loading-xs"></span>{/if}
			{$t('profile.two-factor-enable')}
		</button>
	{:else if view === 'app'}
		<p class="mb-3 text-sm text-base-content/70">{$t('profile.two-factor-app')}</p>
		<button class="btn btn-ghost btn-sm w-full" on:click={() => openView('disable')}>
			{$t('profile.two-factor-disable')}
		</button>
	{:else if view === 'setup' && setup}
		<p class="mb-3 text-sm text-base-content/70">{$t('profile.two-factor-scan')}</p>
		<div class="mb-3 flex justify-center rounded-md bg-white p-2">
			<img src={qrDataUrl} alt="QR code" width="200" height="200" />
		</div>
		<p class="text-xs text-base-content/60">{$t('profile.two-factor-manual')}</p>
		<p class="mb-1 select-all break-all font-mono text-sm">{groupedSecret}</p>
		<a href={setup.uri} class="link link-primary mb-3 inline-block text-xs">
			{$t('profile.two-factor-open-app')}
		</a>
		<form class="flex gap-2" on:submit|preventDefault={confirmEnable}>
			<input
				type="text"
				inputmode="numeric"
				autocomplete="one-time-code"
				maxlength="6"
				bind:value={code}
				placeholder="123456"
				class="input input-sm input-bordered w-full"
				aria-label={$t('auth.app-code')}
			/>
			<button class="btn btn-primary btn-sm" disabled={isBusy || code.length !== 6}>
				{$t('profile.two-factor-confirm')}
			</button>
		</form>
		<button class="btn btn-ghost btn-xs mt-2 w-full" on:click={() => openView('email')}>
			{$t('common.cancel', { default: 'Cancel' })}
		</button>
	{:else if view === 'recovery'}
		<p class="mb-1 text-sm font-medium">{$t('profile.two-factor-recovery-title')}</p>
		<p class="mb-3 text-xs text-base-content/60">{$t('profile.two-factor-recovery-hint')}</p>
		<ul class="mb-3 grid grid-cols-2 gap-1 rounded-md bg-base-100 p-3 font-mono text-sm">
			{#each recoveryCodes as recoveryCode}
				<li>{recoveryCode}</li>
			{/each}
		</ul>
		<div class="flex gap-2">
			<button class="btn btn-outline btn-sm flex-1" on:click={copyRecoveryCodes}>
				{$t('profile.two-factor-copy')}
			</button>
			<button class="btn btn-primary btn-sm flex-1" on:click={() => openView('app')}>
				{$t('profile.two-factor-done')}
			</button>
		</div>
	{:else if view === 'disable'}
		<p class="mb-3 text-sm text-base-content/70">{$t('profile.two-factor-disable-hint')}</p>
		<form class="flex gap-2" on:submit|preventDefault={confirmDisable}>
			<input
				type="text"
				autocomplete="one-time-code"
				maxlength="11"
				bind:value={code}
				placeholder="123456"
				class="input input-sm input-bordered w-full"
				aria-label={$t('auth.app-code')}
			/>
			<button class="btn btn-error btn-sm" disabled={isBusy || code.length < 6}>
				{$t('profile.two-factor-confirm')}
			</button>
		</form>
		<button class="btn btn-ghost btn-xs mt-2 w-full" on:click={() => openView('app')}>
			{$t('common.cancel', { default: 'Cancel' })}
		</button>
	{/if}
</div>
