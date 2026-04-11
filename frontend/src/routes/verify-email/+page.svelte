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

	let status = '';
	let isLoading = true;
	let resendEmail = '';
	let isResendLoading = false;
	let resendRequested = false;

	onMount(async () => {
		const token = new URLSearchParams(window.location.search).get('token') || '';
		window.history.replaceState({}, '', window.location.pathname);
		status = $t('auth.verifying-email');

		if (!token) {
			status = $t('auth.missing-verification-token');
			isLoading = false;
			return;
		}

		try {
			await axios.post('auth/verify-email', { token });
			status = $t('auth.email-verified');
			toastStore.success(status);
			setTimeout(() => goto('/login'), 1200);
		} catch (error) {
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			status = apiResponse?.error || $t('auth.unable-verify-email');
			toastStore.error(status);
		} finally {
			isLoading = false;
		}
	});

	const handleResendVerification = async () => {
		if (!resendEmail) {
			toastStore.error(`${$t('auth.email')} ${$t('common.required')}`);
			return;
		}

		try {
			isResendLoading = true;
			await axios.post('auth/resend-verification', { email: resendEmail });
			resendRequested = true;
			toastStore.success($t('auth.verification-email-requested'));
		} catch (error) {
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			toastStore.error(apiResponse?.error || $t('auth.unable-send-verification-email'));
		} finally {
			isResendLoading = false;
		}
	};
</script>

<main class="flex min-h-screen items-center justify-center bg-base-200 p-4">
	<div class="w-full max-w-md rounded-xl bg-base-100 p-8 shadow-lg text-center">
		<h1 class="text-3xl font-bold text-primary">{$t('auth.verify-email-title')}</h1>
		<p class="mt-4 text-sm text-base-content/70">{status}</p>
		{#if isLoading}
			<div class="mt-6">
				<span class="loading loading-spinner loading-lg"></span>
			</div>
		{/if}

		<div class="mt-8 text-left">
			<h2 class="text-lg font-semibold text-base-content">{$t('auth.resend-verification-title')}</h2>
			<p class="mt-1 text-sm text-base-content/70">{$t('auth.resend-verification-desc')}</p>

			<form class="mt-4 space-y-3" on:submit|preventDefault={handleResendVerification}>
				<div class="form-control">
					<label for="resendEmail" class="label">
						<span class="label-text font-medium">{$t('auth.email')}</span>
					</label>
					<input
						id="resendEmail"
						type="email"
						bind:value={resendEmail}
						class="input input-bordered w-full focus:input-primary"
						placeholder={$t('auth.enter-email')}
					/>
				</div>

				<button class="btn btn-primary w-full text-base-100" disabled={isResendLoading}>
					{#if isResendLoading}
						<span class="loading loading-spinner"></span>
					{/if}
					<span>{$t('auth.send-verification-email')}</span>
				</button>
			</form>

			{#if resendRequested}
				<div class="mt-4 rounded-lg bg-success/10 p-4 text-sm text-success">
					{$t('auth.verification-email-requested')}
				</div>
			{/if}
		</div>
	</div>
</main>
