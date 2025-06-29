<script lang="ts">
	import { goto } from '$app/navigation';
	import axios from '$lib/axios';
	import { t } from '$lib/i18n';
	import type { AxiosError } from 'axios';

	let first_name = '';
	let last_name = '';
	let email = '';
	let password = '';
	let confirmPassword = '';
	let errorMessage = '';
	let successMessage = '';

	interface APIErrorResponse {
		message: string; // Main error message
		error?: string; // Optional error code or additional details
	}

	const handleRegister = async () => {
		errorMessage = '';
		successMessage = '';

		// Validate passwords match
		if (password !== confirmPassword) {
			errorMessage = $t('auth.passwords-no-match');
			return;
		}

		try {
			// Send the register request to the backend
			const response = await axios.post('register', {
				first_name,
				last_name,
				email,
				password
			});

			successMessage = $t('auth.registration-successful');
			setTimeout(() => {
				goto('/login');
			}, 2000);
		} catch (error) {
			// Handle errors and display error messages
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			errorMessage = apiResponse?.error || $t('auth.registration-failed');
		}
	};
</script>

<main
	class="bg-base-200 flex items-center justify-center overflow-auto p-4 md:h-[calc(100vh-64px)]"
>
	<div class="bg-base-100 w-full max-w-lg rounded-xl p-6 shadow-lg">
		<!-- Logo and Brand -->
		<div class="mb-6 text-center">
			<div class="mb-2 flex justify-center">
				<img src="/logo.png" alt="Grão Certo Logo" class="h-16 w-auto object-contain" />
			</div>
			<h1 class="text-primary text-3xl font-bold">Grão Certo</h1>
			<p class="text-base-content/70 mt-2 text-sm">{$t('auth.create-account-desc')}</p>
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
						class="input input-bordered focus:input-primary w-full"
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
						class="input input-bordered focus:input-primary w-full"
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
					class="input input-bordered focus:input-primary w-full"
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
						class="input input-bordered focus:input-primary w-full"
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
						class="input input-bordered focus:input-primary w-full"
						placeholder={$t('auth.reenter-password')}
					/>
				</div>
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

			{#if successMessage}
				<div class="alert alert-success shadow-sm">
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
							d="M5 13l4 4L19 7"
						/>
					</svg>
					<span class="text-sm">{successMessage}</span>
				</div>
			{/if}

			<div class="form-control mt-4">
				<button type="submit" class="btn btn-primary w-full text-lg font-semibold">
					<span class="text-base-100">{$t('auth.register')}</span>
				</button>
			</div>
		</form>

		<div class="divider text-base-content/50">{$t('auth.or')}</div>

		<div class="text-center">
			<p class="text-base-content/70 text-sm">
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
