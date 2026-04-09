<script lang="ts">
	import { goto } from '$app/navigation';
	import axios from '$lib/axios';
	import { t } from '$lib/i18n';
	import type { AxiosError } from 'axios';
	import { validateEmail, isPasswordValid } from '$lib/authValidation';
	import { toastStore } from '$lib/stores/toast';

	let first_name = '';
	let last_name = '';
	let email = '';
	let password = '';
	let confirmPassword = '';

	interface APIErrorResponse {
		message: string; // Main error message
		error?: string; // Optional error code or additional details
	}

	const validateForm = (): boolean => {
		const checks = [
			{
				valid: first_name && first_name.length <= 32,
				error: `${$t('auth.first-name')} ${first_name ? $t('common.too-long') : $t('common.required')}`
			},
			{
				valid: last_name && last_name.length <= 32,
				error: `${$t('auth.last-name')} ${last_name ? $t('common.too-long') : $t('common.required')}`
			},
			{
				valid: email && email.length <= 255 && validateEmail(email),
				error: `${$t('auth.email')} ${!email ? $t('common.required') : !validateEmail(email) ? $t('common.invalid') : $t('common.too-long')}`
			},
			{
				valid: password && password.length >= 8 && password.length <= 64,
				error: `${$t('auth.password')} ${!password ? $t('common.required') : password.length < 8 ? $t('common.too-short') : $t('common.too-long')}`
			}
		];

		for (const { valid, error } of checks) {
			if (!valid) {
				toastStore.error(error);
				return false;
			}
		}

		if (!isPasswordValid(password)) {
			toastStore.error($t('auth.password-invalid'));
			return false;
		}

		if (password !== confirmPassword) {
			toastStore.error($t('auth.passwords-no-match'));
			return false;
		}

		return true;
	};

	const handleRegister = async () => {
		const isValid = validateForm();
		if (!isValid) {
			return; // Stop if validation fails
		}

		try {
			// Send the register request to the backend
			await axios.post('register', {
				first_name,
				last_name,
				email,
				password
			});

			toastStore.success($t('auth.registration-successful'));
			setTimeout(() => {
				goto('/login');
			}, 1000);
		} catch (error) {
			// Handle errors and display error messages
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			toastStore.error(apiResponse?.error || $t('auth.registration-failed'));
		}
	};
</script>

<main class="flex min-h-screen items-center justify-center overflow-auto bg-base-200 p-4">
	<div class="w-full max-w-lg rounded-xl bg-base-100 p-6 shadow-lg">
		<!-- Logo and Brand -->
		<div class="mb-6 text-center">
			<div class="mb-2 flex justify-center">
				<img src="/logo.png" alt="Grão Certo Logo" class="h-16 w-auto object-contain" />
			</div>
			<h1 class="text-3xl font-bold text-primary">Grão Certo</h1>
			<p class="mt-2 text-sm text-base-content/70">{$t('auth.create-account-desc')}</p>
		</div>

		<form class="space-y-3" on:submit|preventDefault={handleRegister}>
			<div class="flex flex-col gap-2 md:flex-row">
				<div class="form-control w-full md:w-1/2">
					<label for="first_name" class="label">
						<span class="label-text font-medium">{$t('auth.first-name')}</span>
					</label>
					<input
						id="first_name"
						type="text"
						bind:value={first_name}
						required
						class="input input-bordered w-full focus:input-primary"
						placeholder={$t('auth.enter-first')}
					/>
				</div>
				<div class="form-control w-full md:w-1/2">
					<label for="last_name" class="label">
						<span class="label-text font-medium">{$t('auth.last-name')}</span>
					</label>
					<input
						id="last_name"
						type="text"
						bind:value={last_name}
						required
						class="input input-bordered w-full focus:input-primary"
						placeholder={$t('auth.enter-last')}
					/>
				</div>
			</div>

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

			<div class="flex flex-col gap-2 md:flex-row">
				<div class="form-control w-full md:w-1/2">
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
				<div class="form-control w-full md:w-1/2">
					<label for="confirmPassword" class="label">
						<span class="label-text font-medium">{$t('auth.confirm-password')}</span>
					</label>
					<input
						id="confirmPassword"
						type="password"
						bind:value={confirmPassword}
						required
						class="input input-bordered w-full focus:input-primary"
						placeholder={$t('auth.reenter-password')}
					/>
				</div>
			</div>

			<div class="form-control mt-4">
				<button type="submit" class="btn btn-primary w-full text-lg font-semibold">
					<span class="text-base-100">{$t('auth.register')}</span>
				</button>
			</div>
		</form>

		<div class="divider text-base-content/50">{$t('auth.or')}</div>

		<div class="text-center">
			<p class="text-sm text-base-content/70">
				{$t('auth.have-account')}
			</p>
			<a href="/login" class="btn btn-outline btn-primary mt-2 w-full">
				{$t('auth.login')}
			</a>
		</div>
	</div>
</main>

<style>
</style>
