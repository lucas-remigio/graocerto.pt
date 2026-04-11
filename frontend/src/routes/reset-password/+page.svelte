<script lang="ts">
	import { goto } from '$app/navigation';
	import axios from '$lib/axios';
	import { t } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast';
	import type { AxiosError } from 'axios';

	interface APIErrorResponse {
		error?: string;
	}

	let token = '';
	let password = '';
	let confirmPassword = '';
	let isLoading = false;

	onMount(() => {
		token = new URLSearchParams(window.location.search).get('token') || '';
		window.history.replaceState({}, '', window.location.pathname);
	});

	const handleSubmit = async () => {
		if (!token) {
			toastStore.error($t('auth.missing-reset-token'));
			return;
		}

		if (!password || password !== confirmPassword) {
			toastStore.error($t('auth.passwords-no-match'));
			return;
		}

		try {
			isLoading = true;
			await axios.post('auth/reset-password', { token, password });
			toastStore.success($t('auth.password-updated'));
			goto('/login');
		} catch (error) {
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			toastStore.error(apiResponse?.error || $t('auth.unable-reset-password'));
		} finally {
			isLoading = false;
		}
	};
</script>

<main class="flex min-h-screen items-center justify-center bg-base-200 p-4">
	<div class="w-full max-w-md rounded-xl bg-base-100 p-8 shadow-lg">
		<h1 class="text-3xl font-bold text-primary">{$t('auth.create-new-password')}</h1>
		<p class="mt-2 text-sm text-base-content/70">{$t('auth.use-link-sent-email')}</p>

		<form class="mt-6 space-y-4" on:submit|preventDefault={handleSubmit}>
			<div class="form-control">
				<label for="password" class="label"><span class="label-text">{$t('auth.password')}</span></label>
				<input id="password" type="password" bind:value={password} class="input input-bordered w-full" />
			</div>
			<div class="form-control">
				<label for="confirmPassword" class="label"><span class="label-text">{$t('auth.confirm-password')}</span></label>
				<input
					id="confirmPassword"
					type="password"
					bind:value={confirmPassword}
					class="input input-bordered w-full"
				/>
			</div>
			<button class="btn btn-primary w-full text-base-100" disabled={isLoading}>
				{#if isLoading}
					<span class="loading loading-spinner"></span>
				{/if}
				<span>{$t('auth.confirm-reset-password-button')}</span>
			</button>
		</form>
	</div>
</main>
