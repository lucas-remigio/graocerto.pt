<script lang="ts">
	import { goto } from '$app/navigation';
	import axios from '$lib/axios';
	import { login, token } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import type { AxiosError } from 'axios';

	let email = '';
	let password = '';
	let errorMessage = '';

	interface APIErrorResponse {
		token?: string; // The main error message
		error?: string; // Optional error code or additional details
	}

	const handleLogin = async () => {
		errorMessage = '';

		// Send the login request to the backend
		try {
			const response = await axios.post('login', { email, password });
			const data = response.data;
			const authToken = data.token;

			if (authToken) {
				// Store in both localStorage and Svelte store
				login(authToken, email);
			}

			goto('/');
		} catch (error) {
			// Type the error as AxiosError
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			errorMessage = apiResponse?.error || $t('auth.error-occurred');

			// Clear any existing tokens on error
			localStorage.removeItem('token');
			token.set(null);
		}
	};
</script>

<main
	class="bg-base-200 flex  items-center justify-center h-[calc(100vh-80px)] min-h-0"
>
	<div class="bg-base-100 w-full max-w-md rounded-lg p-6 shadow-md">
		<h1 class="mb-6 text-center text-2xl font-bold">{$t('auth.login')}</h1>

		<form class="space-y-4" on:submit|preventDefault={handleLogin}>
			<div class="form-control">
				<label for="email" class="label">
					<span class="label-text">{$t('auth.email')}:</span>
				</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="input input-bordered w-full"
					placeholder={$t('auth.enter-email')}
				/>
			</div>

			<div class="form-control">
				<label for="password" class="label">
					<span class="label-text">{$t('auth.password')}:</span>
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="input input-bordered w-full"
					placeholder={$t('auth.enter-password')}
				/>
			</div>

			{#if errorMessage}
				<p class="text-sm text-red-500">{errorMessage}</p>
			{/if}

			<div class="form-control mt-4">
				<button type="submit" class="btn btn-primary w-full">{$t('auth.login')}</button>
			</div>
		</form>

		<p class="mt-4 text-center text-sm">
			{$t('auth.no-account')}
			<a href="/register" class="link link-primary">{$t('auth.register')}</a>
		</p>
	</div>
</main>

<style>
</style>
