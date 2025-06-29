<script lang="ts">
	import { goto } from '$app/navigation';
	import axios from '$lib/axios';
	import { login, token } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import type { AxiosError } from 'axios';
	import { validateEmail, isPasswordValid, isPasswordLengthValid } from '$lib/authValidation';
	import { XIcon } from 'lucide-svelte';

	let email = '';
	let password = '';
	let errorMessage = '';

	interface APIErrorResponse {
		token?: string; // The main error message
		error?: string; // Optional error code or additional details
	}

	const validateForm = (): boolean => {
		if (!email || !validateEmail(email)) {
			errorMessage = $t('auth.email') + ' ' + $t('common.invalid');
			return false;
		}

		if (!password || !isPasswordLengthValid(password)) {
			errorMessage = $t('auth.password-length-invalid');
			return false;
		}

		errorMessage = '';
		return true;
	};

	const handleLogin = async () => {
		errorMessage = '';

		const isValid = validateForm();
		if (!isValid) {
			return;
		}

		// Send the login request to the backend
		try {
			const response = await axios.post('login', { email, password });
			const data = response.data;
			const authToken = data.token;

			if (authToken) {
				// Store in both localStorage and Svelte store
				login(authToken, email);
			}

			goto('/home');
		} catch (error) {
			// Type the error as AxiosError
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			if (axiosError.response?.status === 404) {
				errorMessage = $t('auth.user-not-found');
			} else {
				errorMessage = apiResponse?.error || $t('auth.error-occurred');
			}

			// Clear any existing tokens on error
			localStorage.removeItem('token');
			token.set(null);
		}
	};
</script>

<main
	class="bg-base-200 flex items-center justify-center overflow-auto p-4 md:h-[calc(100vh-64px)]"
>
	<div class="bg-base-100 w-full max-w-md rounded-xl p-8 shadow-lg">
		<!-- Logo and Brand -->
		<div class="mb-8 text-center">
			<div class="mb-4 flex justify-center">
				<img src="/logo.png" alt="Grão Certo Logo" class="h-16 w-auto object-contain" />
			</div>
			<h1 class="text-primary text-3xl font-bold">Grão Certo</h1>
			<p class="text-base-content/70 mt-2 text-sm">{$t('auth.welcome-back')}</p>
		</div>

		<form class="space-y-3" on:submit|preventDefault={handleLogin}>
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
					<XIcon class="text-base-100 h-6 w-6" />
					<span class="text-base-100 text-sm">{errorMessage}</span>
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
