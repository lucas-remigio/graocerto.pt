<script lang="ts">
	import { goto } from '$app/navigation';
	import api from '$lib/axios';
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
			const response = await api.post('login', { email, password });

			const data = response.data;
			// Save the token to localStorage (or cookie)
			localStorage.setItem('authToken', data.token);
			document.cookie = `authToken=${data.token}; Path=/; SameSite=Strict; Secure`;

			goto('/');
		} catch (error) {
			// Type the error as AxiosError
			const axiosError = error as AxiosError;
			const apiResponse = axiosError.response?.data as APIErrorResponse;
			errorMessage = apiResponse?.error || 'An error occurred';
		}
	};
</script>

<main class="bg-base-200 flex min-h-screen items-center justify-center">
	<div class="bg-base-100 w-full max-w-md rounded-lg p-6 shadow-md">
		<h1 class="mb-6 text-center text-2xl font-bold">Login</h1>

		<form class="space-y-4" on:submit|preventDefault={handleLogin}>
			<div class="form-control">
				<label for="email" class="label">
					<span class="label-text">Email:</span>
				</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="input input-bordered w-full"
					placeholder="Enter your email"
				/>
			</div>

			<div class="form-control">
				<label for="password" class="label">
					<span class="label-text">Password:</span>
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="input input-bordered w-full"
					placeholder="Enter your password"
				/>
			</div>

			{#if errorMessage}
				<p class="text-sm text-red-500">{errorMessage}</p>
			{/if}

			<div class="form-control mt-4">
				<button type="submit" class="btn btn-primary w-full">Login</button>
			</div>
		</form>

		<p class="mt-4 text-center text-sm">
			Don't have an account? <a href="/register" class="link link-primary">Register</a>
		</p>
	</div>
</main>

<style>
</style>
