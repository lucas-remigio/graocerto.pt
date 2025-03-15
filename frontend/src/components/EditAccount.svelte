<script lang="ts">
	import api_axios from '$lib/axios';
	import type { Account } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';

	export let account: Account;

	let error: string = '';
	// Form field variables
	let balance: number = account.balance;
	let account_name: string = account.account_name;

	// Create event dispatcher (to emit events to the parent)
	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		error = '';
		if (!validateForm()) {
			return;
		}

		const updatedAccount = {
			balance,
			account_name
		};

		try {
			const response = await api_axios.put(`accounts/${account.id}`, updatedAccount);

			if (response.status !== 200) {
				console.error('Non-200 response status:', response.status);
				error = `Error: ${response.status}`;
				return;
			}
			handleUpdatedAccount();
		} catch (err) {
			console.error('Error in handleSubmit:', err);
			error = 'Failed to update account';
		}
	}

	function validateForm(): boolean {
		// round the balance
		balance = Math.round(balance * 100) / 100;

		// validations
		if (balance < 0) {
			error = 'Balance must be greater than 0';
			return false;
		}

		if (balance > 999999999) {
			error = 'Balance must be less than 999999999';
			return false;
		}

		if (account_name.length < 3) {
			error = 'Account name must be at least 3 characters';
			return false;
		}

		if (account_name.length > 50) {
			error = 'Account name must be less than 50 characters';
			return false;
		}

		return true;
	}

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function handleUpdatedAccount() {
		dispatch('updatedAccount');
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}
			><X /></button
		>
		<h3 class="mb-4 text-lg font-bold">Edit Account</h3>
		<!--Error message-->
		{#if error}
			<div class="alert alert-error">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}
		<form on:submit|preventDefault={handleSubmit}>
			<!-- Description Field -->
			<div class="form-control mt-4">
				<label class="label" for="account_name">
					<span class="label-text">Account name</span>
				</label>
				<input
					id="account_name"
					type="text"
					placeholder="Account name"
					class="input input-bordered"
					bind:value={account_name}
					required
				/>
			</div>

			<!-- Amount Field -->
			<div class="form-control mt-4">
				<label class="label" for="balance">
					<span class="label-text">Balance</span>
				</label>
				<input
					id="balance"
					type="number"
					placeholder="Enter amount"
					class="input input-bordered"
					min="0"
					step="0.01"
					max="999999999"
					bind:value={balance}
					required
				/>
			</div>

			<!-- Form Actions -->
			<div class="modal-action mt-6">
				<button type="button" class="btn" on:click={handleCloseModal}>Cancel</button>
				<button type="submit" class="btn btn-primary">Update Account</button>
			</div>
		</form>
	</div>
</div>
