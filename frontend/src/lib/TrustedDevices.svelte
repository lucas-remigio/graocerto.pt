<script lang="ts">
	import { onMount } from 'svelte';
	import { t, locale } from '$lib/i18n';
	import { MonitorSmartphone } from 'lucide-svelte';
	import api_axios from './axios';
	import { toastStore } from '$lib/stores/toast';
	import type { TrustedDevice } from '$lib/types';

	let devices: TrustedDevice[] = [];
	let isLoading = true;
	let busyId: number | 'all' | null = null;

	onMount(loadDevices);

	async function loadDevices() {
		try {
			const response = await api_axios.get<TrustedDevice[]>('auth/devices');
			devices = response.data;
		} catch {
			toastStore.error($t('auth.error-occurred'));
		} finally {
			isLoading = false;
		}
	}

	async function forget(target: number | 'all') {
		busyId = target;
		try {
			await api_axios.delete(target === 'all' ? 'auth/devices' : `auth/devices/${target}`);
			devices = target === 'all' ? [] : devices.filter((device) => device.id !== target);
		} catch {
			toastStore.error($t('auth.error-occurred'));
		} finally {
			busyId = null;
		}
	}

	function formatDate(value: string): string {
		return new Date(value).toLocaleDateString($locale === 'en' ? 'en-GB' : 'pt-PT', {
			day: '2-digit',
			month: 'short',
			year: 'numeric'
		});
	}
</script>

<div class="rounded-lg bg-base-200 p-4">
	<div class="mb-1 flex items-center gap-2">
		<MonitorSmartphone class="h-4 w-4 text-primary" />
		<span class="text-sm font-medium">{$t('profile.trusted-devices')}</span>
	</div>
	<p class="mb-3 text-xs text-base-content/60">{$t('profile.trusted-devices-hint')}</p>

	{#if isLoading}
		<span class="loading loading-spinner loading-sm"></span>
	{:else if devices.length === 0}
		<p class="text-sm text-base-content/70">{$t('profile.trusted-devices-empty')}</p>
	{:else}
		<ul class="space-y-2">
			{#each devices as device (device.id)}
				<li class="flex items-center justify-between gap-2 rounded-md bg-base-100 px-3 py-2">
					<div class="min-w-0">
						<p class="truncate text-sm font-medium">
							{device.label}
							{#if device.is_current}
								<span class="badge badge-primary badge-sm ml-1">{$t('profile.this-device')}</span>
							{/if}
						</p>
						<p class="text-xs text-base-content/60">
							{$t('profile.last-used', { values: { date: formatDate(device.last_used_at) } })}
						</p>
					</div>
					<button
						class="btn btn-ghost btn-xs text-error"
						disabled={busyId !== null}
						on:click={() => forget(device.id)}
					>
						{#if busyId === device.id}
							<span class="loading loading-spinner loading-xs"></span>
						{:else}
							{$t('profile.forget-device')}
						{/if}
					</button>
				</li>
			{/each}
		</ul>

		{#if devices.length > 1}
			<button
				class="btn btn-outline btn-error btn-sm mt-3 w-full"
				disabled={busyId !== null}
				on:click={() => forget('all')}
			>
				{$t('profile.forget-all-devices')}
			</button>
		{/if}
	{/if}
</div>
