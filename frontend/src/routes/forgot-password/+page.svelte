<script lang="ts">
	import { t } from '$lib/i18n';
	import axios from '$lib/axios';
	import { goto } from '$app/navigation';
	import { toastStore } from '$lib/stores/toast';
	import type { AxiosError } from 'axios';

	interface APIErrorResponse {
		error?: string;
	}

	let email = '';
	let isLoading = false;
	let submitted = false;

	const handleSubmit = async () => {
		if (!email) {
			toastStore.error(`${$t('auth.email')} ${$t('common.required')}`);
			return;
		}

		try {
			isLoading = true;
			await axios.post('auth/forgot-password', { email });
			submitted = true;
			toastStore.success($t('auth.reset-link-sent'));
		} catch (error) {
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			toastStore.error(apiResponse?.error || $t('auth.unable-send-reset-email'));
		} finally {
			isLoading = false;
		}
	};
</script>

<main class="flex min-h-screen items-center justify-center bg-base-200 p-4">
	<div class="w-full max-w-md rounded-xl bg-base-100 p-8 shadow-lg">
		<h1 class="text-3xl font-bold text-primary">{$t('auth.forgot-password-title')}</h1>
		<p class="mt-2 text-sm text-base-content/70">{$t('auth.forgot-password-desc')}</p>

		{#if submitted}
			<div class="mt-6 rounded-lg bg-success/10 p-4 text-success">
				{$t('auth.reset-link-sent')}
			</div>
		{/if}

		<form class="mt-6 space-y-4" on:submit|preventDefault={handleSubmit}>
			<div class="form-control">
				<label for="email" class="label"><span class="label-text">{$t('auth.email')}</span></label>
				<input id="email" type="email" bind:value={email} class="input input-bordered w-full" />
			</div>
			<button class="btn btn-primary w-full text-base-100" disabled={isLoading}>
				{#if isLoading}
					<span class="loading loading-spinner"></span>
				{/if}
				<span>{$t('auth.send-reset-email')}</span>
			</button>
		</form>

		<button class="btn btn-ghost mt-4 w-full" on:click={() => goto('/login')}>
			{$t('auth.back-to-login')}
		</button>
	</div>
</main>
