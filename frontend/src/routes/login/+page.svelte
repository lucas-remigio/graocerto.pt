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

<main class="bg-base-200 flex h-[calc(100vh-64px)] items-center justify-center p-4">
	<div class="bg-base-100 w-full max-w-md rounded-xl p-8 shadow-lg">
		<!-- Logo and Brand -->
		<div class="mb-8 text-center">
			<div class="mb-4 flex justify-center">
				<img src="/logo.png" alt="Grão Certo Logo" class="h-16 w-auto object-contain" />
			</div>
			<h1 class="text-primary text-3xl font-bold">Grão Certo</h1>
			<p class="text-base-content/70 mt-2 text-sm">{$t('auth.welcome-back')}</p>
		</div>

		<form class="space-y-6" on:submit|preventDefault={handleLogin}>
			<div class="form-control">
				<label for="email" class="label">
					<span class="label-text font-medium">{$t('auth.email')}</span>
				</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="input input-bordered focus:input-primary w-full"
					placeholder={$t('auth.enter-email')}
				/>
			</div>

			<div class="form-control">
				<label for="password" class="label">
					<span class="label-text font-medium">{$t('auth.password')}</span>
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="input input-bordered focus:input-primary w-full"
					placeholder={$t('auth.enter-password')}
				/>
			</div>

			{#if errorMessage}
				<div class="alert alert-error shadow-sm">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-6 w-6 shrink-0 stroke-current"
						fill="none"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<span class="text-sm">{errorMessage}</span>
				</div>
			{/if}

			<div class="form-control mt-8">
				<button type="submit" class="btn btn-primary w-full text-lg font-semibold">
					<span class="text-base-100">{$t('auth.login')}</span>
				</button>
			</div>
		</form>

		<div class="divider text-base-content/50">{$t('auth.or')}</div>

		<div class="text-center">
			<p class="text-base-content/70 text-sm">
				{$t('auth.no-account')}
			</p>
			<a href="/register" class="btn btn-outline btn-primary mt-2 w-full">
				{$t('auth.create-account')}
			</a>
		</div>
	</div>
</main>

<style>
</style>
