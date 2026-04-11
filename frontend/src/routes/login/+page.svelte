<script lang="ts">
	import { goto } from '$app/navigation';
	import axios from '$lib/axios';
	import { login, token } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import type { AxiosError } from 'axios';
	import { validateEmail, isPasswordLengthValid } from '$lib/authValidation';
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast';

	// Vite will replace this at build time, failing gracefully if undefined
	const googleClientId = import.meta.env.VITE_PUBLIC_GOOGLE_CLIENT_ID || '';

	let email = '';
	let password = '';
	let otpCode = '';
	let challengeId = '';
	let isLoading = false;
	let isOtpLoading = false;
	let isGoogleLoading = false;
	let isOtpStep = false;

	interface APIErrorResponse {
		token?: string; // The main error message
		error?: string; // Optional error code or additional details
		challenge_id?: string;
		message?: string;
	}

	const initGoogle = () => {
		if (!googleClientId) {
			console.error('VITE_PUBLIC_GOOGLE_CLIENT_ID is blank/undefined in this environment.');
			return;
		}

		if (!window.google) return;

		window.google.accounts.id.initialize({
			client_id: googleClientId,
			callback: handleGoogleLogin
		});

		window.google.accounts.id.renderButton(document.getElementById('googleSignInDiv'), {
			theme: 'outline',
			size: 'large'
		});
	};

	onMount(() => {
		// Just in case the script loaded incredibly fast from browser cache
		if (window.google) {
			initGoogle();
		}
	});

	const handleGoogleLogin = async (response: any) => {
		try {
			isGoogleLoading = true;
			const res = await axios.post('auth/google', { token: response.credential });
			const { token: authToken, email, created_at } = res.data;

			if (authToken) {
				login(authToken, email, created_at ?? null);
				goto('/home');
			}
		} catch (error) {
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			toastStore.error(apiResponse?.error || $t('auth.error-occurred'));
			localStorage.removeItem('token');
			token.set(null);
		} finally {
			isGoogleLoading = false;
		}
	};

	const validateForm = (): boolean => {
		if (!email || !validateEmail(email)) {
			toastStore.error($t('auth.email') + ' ' + $t('common.invalid'));
			return false;
		}

		if (!password || !isPasswordLengthValid(password)) {
			toastStore.error($t('auth.password-length-invalid'));
			return false;
		}

		return true;
	};

	const handleLogin = async () => {
		const isValid = validateForm();
		if (!isValid) {
			return;
		}

		try {
			isLoading = true;
			const response = await axios.post('login', { email, password });
			const data = response.data;
			if (response.status === 202 && data.challenge_id) {
				challengeId = data.challenge_id;
				isOtpStep = true;
				otpCode = '';
				toastStore.success(
					data.message || $t('auth.login-code-sent', { values: { email } })
				);
				return;
			}

			if (data.token) {
				login(data.token, email, data.created_at ?? null);
				goto('/home');
			}
		} catch (error) {
			// Type the error as AxiosError
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			if (axiosError.response?.status === 404) {
				toastStore.error($t('auth.user-not-found'));
			} else {
				toastStore.error(apiResponse?.error || $t('auth.error-occurred'));
			}

			// Clear any existing tokens on error
			localStorage.removeItem('token');
			token.set(null);
		} finally {
			isLoading = false;
		}
	};

	const handleOtpVerify = async () => {
		if (!challengeId || !otpCode) {
			toastStore.error('Enter the 6-digit code sent to your email.');
			return;
		}

		try {
			isOtpLoading = true;
			const response = await axios.post('auth/verify-login-otp', {
				challenge_id: challengeId,
				code: otpCode
			});

			const { token: authToken, created_at } = response.data;
			if (authToken) {
				login(authToken, email, created_at ?? null);
				goto('/home');
			}
		} catch (error) {
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			toastStore.error(apiResponse?.error || $t('auth.error-occurred'));
		} finally {
			isOtpLoading = false;
		}
	};
</script>

<svelte:head>
	{#if googleClientId}
		<script src="https://accounts.google.com/gsi/client" async defer on:load={initGoogle}></script>
	{/if}
</svelte:head>

<main class="flex min-h-screen items-center justify-center overflow-auto bg-base-200 p-4">
	<div class="w-full max-w-md rounded-xl bg-base-100 p-8 shadow-lg">
		<!-- Logo and Brand -->
		<div class="mb-8 text-center">
			<div class="mb-4 flex justify-center">
				<img src="/logo.png" alt="Grão Certo Logo" class="h-16 w-auto object-contain" />
			</div>
			<h1 class="text-3xl font-bold text-primary">Grão Certo</h1>
			<p class="mt-2 text-sm text-base-content/70">{$t('auth.welcome-back')}</p>
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
					class="input input-bordered w-full focus:input-primary"
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
					class="input input-bordered w-full focus:input-primary"
					placeholder={$t('auth.enter-password')}
				/>
			</div>

			<div class="flex items-center justify-between text-sm">
				<a href="/forgot-password" class="link link-primary">{$t('auth.forgot-password')}</a>
				{#if isOtpStep}
					<button
						type="button"
						class="link link-primary"
						on:click={() => {
							isOtpStep = false;
							challengeId = '';
							otpCode = '';
						}}
					>
						{$t('auth.use-different-email')}
					</button>
				{/if}
			</div>

			<div class="form-control mt-8">
				<button
					type="submit"
					class="btn btn-primary w-full text-lg font-semibold"
					disabled={isLoading || isGoogleLoading || isOtpLoading}
				>
					{#if isLoading}
						<span class="loading loading-spinner"></span>
					{/if}
					<span class="text-base-100">{$t('auth.login')}</span>
				</button>
			</div>
		</form>

		{#if isOtpStep}
			<div class="mt-4 rounded-lg border border-primary/30 bg-primary/5 p-4">
				<p class="mb-3 text-sm font-medium">
					{$t('auth.login-code-sent', { values: { email } })}
				</p>
				<div class="form-control">
					<label for="otpCode" class="label">
						<span class="label-text font-medium">{$t('auth.email-code')}</span>
					</label>
					<input
						id="otpCode"
						type="text"
						inputmode="numeric"
						maxlength="6"
						bind:value={otpCode}
						class="input input-bordered w-full focus:input-primary"
						placeholder="123456"
					/>
				</div>
				<div class="mt-4">
					<button class="btn btn-secondary w-full" disabled={isOtpLoading} on:click={handleOtpVerify}>
						{#if isOtpLoading}
							<span class="loading loading-spinner"></span>
						{/if}
						<span>{$t('auth.verify-code')}</span>
					</button>
				</div>
			</div>
		{/if}

		<div class="divider text-base-content/50">{$t('auth.or')}</div>

		<div class="text-center">
			<p class="text-sm text-base-content/70">
				{$t('auth.no-account')}
			</p>
			<a
				href="/register"
				class="btn btn-outline btn-primary mt-2 w-full"
				class:btn-disabled={isLoading || isGoogleLoading}
				aria-disabled={isLoading || isGoogleLoading}
			>
				{$t('auth.create-account')}
			</a>
		</div>

		{#if googleClientId}
			<div class="mt-4 w-full">
				<div id="googleSignInDiv" class="flex justify-center" class:hidden={isGoogleLoading}></div>

				{#if isGoogleLoading}
					<button class="btn btn-disabled btn-outline w-full !text-base-content" disabled>
						<span class="loading loading-spinner"></span>
						{$t('common.loading')}...
					</button>
				{/if}
			</div>
		{/if}
	</div>
</main>

<style>
</style>
