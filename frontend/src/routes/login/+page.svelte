<script>
	import { goto } from '$app/navigation';

	let email = '';
	let password = '';
	let errorMessage = '';

	const handleLogin = async () => {
		errorMessage = ''; // Clear previous errors

		// Send the login request to the backend
		const response = await fetch('http://localhost:8080/api/v1/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email, password })
		});

		if (response.ok) {
			const data = await response.json();
			// Save the token to localStorage (or cookie)
			localStorage.setItem('authToken', data.token);

			// Redirect to the home page or dashboard
			goto('/');
		} else {
			// Handle login error
			const errorData = await response.json();
			errorMessage = errorData.message || 'Login failed. Please try again.';
		}
	};
</script>

<main>
	<h1>Login</h1>

	<form on:submit|preventDefault={handleLogin}>
		<label for="email">Email:</label>
		<input id="email" type="email" bind:value={email} required />

		<label for="password">Password:</label>
		<input id="password" type="password" bind:value={password} required />

		{#if errorMessage}
			<p class="error">{errorMessage}</p>
		{/if}

		<button type="submit">Login</button>
	</form>
</main>

<style>
	.error {
		color: red;
		margin-top: 0.5rem;
	}
</style>
