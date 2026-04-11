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

	onMount(async () => {
		const token = new URLSearchParams(window.location.search).get('token') || '';
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
	</div>
</main>
