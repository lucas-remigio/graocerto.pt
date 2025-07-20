<script lang="ts">
	import api_axios from '$lib/axios';
	import type { Account } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import { t } from '$lib/i18n';
	import { validateAccountForm } from '$lib/accountValidation';

	export let account: Account;

	let error: string = '';
	// Form field variables
	let balance: number = account.balance;
	let account_name: string = account.account_name;

	// Create event dispatcher (to emit events to the parent)
	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		error = '';
		if (!isFormValid()) {
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
			error = $t('errors.failed-update-account');
		}
	}

	function isFormValid(): boolean {
		const result = validateAccountForm(balance, account_name, $t);
		if (result.error) {
			error = result.error;
			return false;
		}
		balance = result.balance; // use parsed/rounded balance
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
		<h3 class="mb-4 text-lg font-bold">{$t('accounts.edit-account-title')}</h3>
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
					<span class="label-text">{$t('accounts.account-name')}</span>
				</label>
				<input
					id="account_name"
					type="text"
					placeholder={$t('accounts.account-name-placeholder')}
					class="input input-bordered"
					bind:value={account_name}
					required
				/>
			</div>

			<!-- Amount Field -->
			<div class="form-control mt-4">
				<label class="label" for="balance">
					<span class="label-text">{$t('accounts.balance')}</span>
				</label>
				<input
					id="balance"
					type="text"
					inputmode="decimal"
					placeholder={$t('accounts.balance-placeholder')}
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
				<button type="button" class="btn" on:click={handleCloseModal}>{$t('common.cancel')}</button>
				<button type="submit" class="btn btn-primary text-base-100"
					>{$t('accounts.update-account')}</button
				>
			</div>
		</form>
	</div>
</div>
