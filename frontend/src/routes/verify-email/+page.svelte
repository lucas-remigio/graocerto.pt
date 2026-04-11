<script lang="ts">
	import { goto } from '$app/navigation';
	import axios from '$lib/axios';
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast';
	import type { AxiosError } from 'axios';

	interface APIErrorResponse {
		error?: string;
	}

	let status = 'Verifying your email...';
	let isLoading = true;

	onMount(async () => {
		const token = new URLSearchParams(window.location.search).get('token') || '';

		if (!token) {
			status = 'Missing verification token.';
			isLoading = false;
			return;
		}

		try {
			await axios.post('auth/verify-email', { token });
			status = 'Email verified. You can log in now.';
			toastStore.success(status);
			setTimeout(() => goto('/login'), 1200);
		} catch (error) {
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			status = apiResponse?.error || 'Unable to verify email.';
			toastStore.error(status);
		} finally {
			isLoading = false;
		}
	});
</script>

<main class="flex min-h-screen items-center justify-center bg-base-200 p-4">
	<div class="w-full max-w-md rounded-xl bg-base-100 p-8 shadow-lg text-center">
		<h1 class="text-3xl font-bold text-primary">Verify email</h1>
		<p class="mt-4 text-sm text-base-content/70">{status}</p>
		{#if isLoading}
			<div class="mt-6">
				<span class="loading loading-spinner loading-lg"></span>
			</div>
		{/if}
	</div>
</main>
