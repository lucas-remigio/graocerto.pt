<script lang="ts">
	import api_axios from '$lib/axios';
	import type { Account, AccountChangeResponse } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/i18n';
	import { validateAccountForm } from '$lib/accountValidation';

	// Props
	export let account: Account | null = null; // if provided => edit mode

	// Mode
	let isEditMode: boolean = !!account;
	$: isEditMode = !!account;

	// State
	let error: string = '';

	// Form fields
	let account_name: string = isEditMode ? account!.account_name : '';
	let balance: number | string = isEditMode ? account!.balance : 0;

	const dispatch = createEventDispatcher();

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function isFormValid(): boolean {
		const result = validateAccountForm(balance, account_name, $t);
		if (result.error) {
			error = result.error;
			return false;
		}
		balance = result.balance;
		return true;
	}

	async function handleSubmit() {
		error = '';
		if (!isFormValid()) return;

		try {
			if (isEditMode) {
				const payload = { balance, account_name };
				const response = await api_axios.put(`accounts/${account!.id}`, payload);
				if (response.status !== 200) {
					console.error('Non-200 response status:', response.status);
					error = `Error: ${response.status}`;
					return;
				}
				dispatch('updatedAccount', response.data as AccountChangeResponse);
			} else {
				const payload = { balance, account_name };
				const response = await api_axios.post('accounts', payload);
				if (response.status !== 200) {
					console.error('Non-200 response status:', response.status);
					error = `Error: ${response.status}`;
					return;
				}
				dispatch('newAccount', response.data as AccountChangeResponse);
			}
		} catch (err) {
			console.error('Error in handleSubmit:', err);
			error = isEditMode ? $t('errors.failed-update-account') : $t('errors.failed-create-account');
		}
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}
			><X /></button
		>
		<h3 class="mb-4 text-lg font-bold">
			{#if isEditMode}
				{$t('accounts.edit-account-title')}
			{:else}
				{$t('accounts.create-account')}
			{/if}
		</h3>

		{#if error}
			<div class="alert alert-error">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}

		<form on:submit|preventDefault={handleSubmit}>
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

			<div class="modal-action mt-6">
				<button type="button" class="btn" on:click={handleCloseModal}>{$t('common.cancel')}</button>
				{#if isEditMode}
					<button type="submit" class="btn btn-primary text-base-100">
						{$t('accounts.update-account')}
					</button>
				{:else}
					<button type="submit" class="btn btn-primary text-base-100">
						{$t('accounts.create-account')}
					</button>
				{/if}
			</div>
		</form>
	</div>
</div>
