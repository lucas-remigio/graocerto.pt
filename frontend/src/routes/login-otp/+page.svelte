<script lang="ts">
	import { goto } from '$app/navigation';
	import axios from '$lib/axios';
	import { login, token } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import { onMount } from 'svelte';
	import { toastStore } from '$lib/stores/toast';
	import type { AxiosError } from 'axios';
	import type { LoginOtpState, MfaMethod } from '$lib/types';

	interface APIErrorResponse {
		error?: string;
	}

	let email = '';
	let challengeId = '';
	let otpCode = '';
	let method: MfaMethod = 'email_otp';
	let useRecoveryCode = false;
	let rememberDevice = false;
	let isLoading = false;
	let isReady = false;
	let status = '';
	const loginOtpStateKey = 'wallet-tracker-login-otp-state';

	$: isApp = method === 'totp';
	$: codeLabel = useRecoveryCode
		? $t('auth.recovery-code')
		: isApp
			? $t('auth.app-code')
			: $t('auth.email-code');

	onMount(() => {
		const storedState = window.sessionStorage.getItem(loginOtpStateKey);
		if (!storedState) {
			status = $t('auth.missing-login-challenge');
			isReady = false;
			return;
		}

		try {
			const parsedState = JSON.parse(storedState) as Partial<LoginOtpState>;
			email = parsedState.email || '';
			challengeId = parsedState.challengeId || '';
			method = parsedState.method === 'totp' ? 'totp' : 'email_otp';
		} catch {
			window.sessionStorage.removeItem(loginOtpStateKey);
			status = $t('auth.missing-login-challenge');
			isReady = false;
			return;
		}

		isReady = Boolean(email && challengeId);
		if (!isReady) {
			status = $t('auth.missing-login-challenge');
		} else {
			// Compare `method` directly: the reactive `isApp` hasn't re-run yet inside onMount.
			status =
				method === 'totp'
					? $t('auth.app-code-prompt')
					: $t('auth.login-code-sent', { values: { email } });
		}
	});

	const handleSubmit = async () => {
		if (!challengeId) {
			toastStore.error($t('auth.missing-login-challenge'));
			return;
		}

		if (!otpCode) {
			toastStore.error(isApp ? $t('auth.enter-app-code') : $t('auth.enter-otp-code'));
			return;
		}

		try {
			isLoading = true;
			const response = await axios.post('auth/verify-login-otp', {
				challenge_id: challengeId,
				code: otpCode,
				remember_device: rememberDevice
			});

			const { token: authToken, created_at } = response.data;
			if (authToken) {
				window.sessionStorage.removeItem(loginOtpStateKey);
				login(authToken, email, created_at ?? null);
				goto('/home');
			}
		} catch (error) {
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			toastStore.error(apiResponse?.error || $t('auth.error-occurred'));
		} finally {
			isLoading = false;
		}
	};
</script>

<main class="flex min-h-screen items-center justify-center bg-base-200 p-4">
	<div class="w-full max-w-md rounded-xl bg-base-100 p-8 shadow-lg">
		<div class="mb-6 text-center">
			<div class="mb-3 flex justify-center">
				<img src="/logo.png" alt="Grão Certo Logo" class="h-14 w-auto object-contain" />
			</div>
			<h1 class="text-3xl font-bold text-primary">{$t('auth.login-otp-title')}</h1>
			<p class="mt-2 text-sm text-base-content/70">{status}</p>
		</div>

		{#if isReady}
			<form class="space-y-4" on:submit|preventDefault={handleSubmit}>
				<div class="form-control">
					<label for="otpCode" class="label">
						<span class="label-text font-medium">{codeLabel}</span>
					</label>
					<input
						id="otpCode"
						type="text"
						inputmode={useRecoveryCode ? 'text' : 'numeric'}
						autocomplete="one-time-code"
						maxlength={useRecoveryCode ? 11 : 6}
						bind:value={otpCode}
						class="input input-bordered w-full focus:input-primary"
						placeholder={useRecoveryCode ? 'XXXXX-XXXXX' : $t('auth.otp-placeholder')}
					/>
					{#if isApp}
						<button
							type="button"
							class="link link-primary mt-2 self-start text-sm"
							on:click={() => {
								useRecoveryCode = !useRecoveryCode;
								otpCode = '';
							}}
						>
							{useRecoveryCode ? $t('auth.use-app-code') : $t('auth.use-recovery-code')}
						</button>
					{/if}
				</div>

				<label class="label cursor-pointer justify-start gap-3">
					<input type="checkbox" class="checkbox-primary checkbox" bind:checked={rememberDevice} />
					<span class="label-text">{$t('auth.remember-device')}</span>
				</label>

				<button class="btn btn-secondary w-full text-base-100" disabled={isLoading}>
					{#if isLoading}
						<span class="loading loading-spinner"></span>
					{/if}
					<span>{$t('auth.verify-code')}</span>
				</button>
			</form>
		{/if}

		<div class="mt-4 text-center">
			<a href="/login" class="link link-primary">{$t('auth.back-to-login')}</a>
		</div>
	</div>
</main>
