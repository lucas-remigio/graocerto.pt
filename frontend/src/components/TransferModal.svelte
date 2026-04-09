<script lang="ts">
	import api_axios from '$lib/axios';
	import { dataService } from '$lib/services/dataService';
	import type { Account, CategoryDto, TransferResponse } from '$lib/types';
	import { ArrowRight, X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import { t } from '$lib/i18n';
	import { buildCategoryGroups } from '$lib/utils/categoryUtils';
	import { toastStore } from '$lib/stores/toast';

	export let account: Account;

	const dispatch = createEventDispatcher();

	let isLoading: boolean = true;
	let accounts: Account[] = [];
	let debitCategories: CategoryDto[] = [];
	let creditCategories: CategoryDto[] = [];

	// Form fields
	let sourceAccountToken = account.token;
	let destinationAccountToken = '';
	let debitCategoryId: string = '';
	let creditCategoryId: string = '';
	let amount: number = 0;
	let description: string = '';
	let date: string = new Date().toISOString().split('T')[0];

	$: selectedDebitCategory = debitCategories.find((c) => c.id === Number(debitCategoryId));
	$: selectedCreditCategory = creditCategories.find((c) => c.id === Number(creditCategoryId));
	$: debitBorderColor = selectedDebitCategory?.color || '#ef4444';
	$: creditBorderColor = selectedCreditCategory?.color || '#22c55e';
	$: groupedDebitCategories = buildCategoryGroups(debitCategories);
	$: groupedCreditCategories = buildCategoryGroups(creditCategories);

	function handleCloseModal() {
		dispatch('closeModal');
	}

	async function handleSubmit() {
		// Validation
		if (!destinationAccountToken) {
			toastStore.error($t('transfers.select-destination-account'));
			return;
		}

		if (sourceAccountToken === destinationAccountToken) {
			toastStore.error($t('transfers.same-account-error'));
			return;
		}

		if (!debitCategoryId) {
			toastStore.error($t('transfers.select-debit-category'));
			return;
		}

		if (!creditCategoryId) {
			toastStore.error($t('transfers.select-credit-category'));
			return;
		}

		if (amount <= 0) {
			toastStore.error($t('transfers.amount-must-be-positive'));
			return;
		}

		try {
			const transferPayload = {
				source_account_token: sourceAccountToken,
				destination_account_token: destinationAccountToken,
				debit_category_id: Number(debitCategoryId),
				credit_category_id: Number(creditCategoryId),
				amount,
				description,
				date
			};

			const response = await api_axios('transactions/transfer', {
				method: 'POST',
				data: transferPayload
			});

			if (response.status !== 200) {
				toastStore.error(`Error: ${response.status}`);
				return;
			}

			toastStore.success($t('common.success'));
			dispatch('newTransfer', response.data as TransferResponse);
		} catch (err) {
			console.error('Error creating transfer:', err);
			toastStore.error($t('errors.failed-create-transfer'));
		}
	}

	async function fetchData() {
		isLoading = true;

		try {
			// Fetch all accounts for the user
			accounts = await dataService.fetchAccounts();

			// Fetch categories and filter by type
			const allCategories = await dataService.fetchCategories();
			debitCategories = allCategories.filter((cat) => cat.transaction_type.type_slug === 'debit');
			creditCategories = allCategories.filter((cat) => cat.transaction_type.type_slug === 'credit');

			// Set default categories if available
			if (debitCategories.length > 0 && !debitCategoryId) {
				debitCategoryId = String(debitCategories[0].id);
			}
			if (creditCategories.length > 0 && !creditCategoryId) {
				creditCategoryId = String(creditCategories[0].id);
			}
		} catch (err) {
			console.error('Error fetching data:', err);
			toastStore.error($t('errors.failed-load-data'));
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		fetchData();
	});
</script>

<div class="modal modal-open">
	<div class="modal-box relative max-w-3xl border-4 border-blue-500 dark:border-blue-400">
		<button class="btn btn-circle btn-sm absolute right-2 top-2" on:click={handleCloseModal}>
			<X />
		</button>

		<h3 class="mb-4 text-lg font-bold">
			{$t('transfers.create-transfer')}
		</h3>

		{#if isLoading}
			<div class="py-12 text-center">
				<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
				<p class="text-base-content/70">{$t('common.loading')}</p>
			</div>
		{:else}
			<form on:submit|preventDefault={handleSubmit}>
				<!-- From/To Section -->
				<div class="flex flex-col gap-4 md:flex-row md:items-center">
					<!-- Left Side: Source Account & Debit Category -->
					<div class="flex-1 space-y-4">
						<!-- Source Account -->
						<div class="form-control">
							<label class="label" for="source-account">
								<span class="label-text font-semibold">{$t('transfers.from-account')}</span>
							</label>
							<select
								id="source-account"
								class="select select-bordered w-full"
								bind:value={sourceAccountToken}
								required
							>
								{#each accounts as acc}
									<option value={acc.token}>{acc.account_name}</option>
								{/each}
							</select>
						</div>

						<!-- Debit Category (Money leaving source account) -->
						{#if debitCategories.length > 0}
							<div class="form-control">
								<label class="label" for="debit-category">
									<span class="label-text">{$t('transfers.debit-category')}</span>
									<span class="label-text-alt text-xs opacity-70"
										>{$t('transfers.money-leaving')}</span
									>
								</label>
								<select
									id="debit-category"
									class="select select-bordered w-full border-2"
									bind:value={debitCategoryId}
									required
									style="border-color: {debitBorderColor} !important;"
								>
									{#each groupedDebitCategories as group}
										{#if group.parent}
											{#if group.children.length > 0}
												<option value={String(group.parent.id)} class="font-semibold">
													{group.parent.category_name}
												</option>
												{#each group.children as child}
													<option value={String(child.id)}>
														&nbsp;&nbsp;&nbsp;&nbsp;{child.category_name}
													</option>
												{/each}
											{:else}
												<option value={String(group.parent.id)}>
													{group.parent.category_name}
												</option>
											{/if}
										{:else}
											{#each group.children as child}
												<option value={String(child.id)}>{child.category_name}</option>
											{/each}
										{/if}
									{/each}
								</select>
							</div>
						{:else}
							<div class="alert alert-warning">
								<div class="flex flex-col gap-1">
									<p class="text-sm">{$t('transfers.no-debit-categories')}</p>
									<a href="/categories" class="link text-xs"
										>{$t('transfers.create-category-first')}</a
									>
								</div>
							</div>
						{/if}
					</div>

					<!-- Arrow Icon (Desktop only) -->
					<div class="hidden items-center justify-center md:flex md:px-4">
						<ArrowRight size={32} class="text-primary" />
					</div>

					<!-- Right Side: Destination Account & Credit Category -->
					<div class="flex-1 space-y-4">
						<!-- Destination Account -->
						<div class="form-control">
							<label class="label" for="dest-account">
								<span class="label-text font-semibold">{$t('transfers.to-account')}</span>
							</label>
							<select
								id="dest-account"
								class="select select-bordered w-full"
								bind:value={destinationAccountToken}
								required
							>
								<option value="" disabled>{$t('transfers.select-destination')}</option>
								{#each accounts.filter((a) => a.token !== sourceAccountToken) as acc}
									<option value={acc.token}>{acc.account_name}</option>
								{/each}
							</select>
						</div>

						<!-- Credit Category (Money entering destination account) -->
						{#if creditCategories.length > 0}
							<div class="form-control">
								<label class="label" for="credit-category">
									<span class="label-text">{$t('transfers.credit-category')}</span>
									<span class="label-text-alt text-xs opacity-70"
										>{$t('transfers.money-entering')}</span
									>
								</label>
								<select
									id="credit-category"
									class="select select-bordered w-full border-2"
									bind:value={creditCategoryId}
									required
									style="border-color: {creditBorderColor} !important;"
								>
									{#each groupedCreditCategories as group}
										{#if group.parent}
											{#if group.children.length > 0}
												<option value={String(group.parent.id)} class="font-semibold">
													{group.parent.category_name}
												</option>
												{#each group.children as child}
													<option value={String(child.id)}>
														&nbsp;&nbsp;&nbsp;&nbsp;{child.category_name}
													</option>
												{/each}
											{:else}
												<option value={String(group.parent.id)}>
													{group.parent.category_name}
												</option>
											{/if}
										{:else}
											{#each group.children as child}
												<option value={String(child.id)}>{child.category_name}</option>
											{/each}
										{/if}
									{/each}
								</select>
							</div>
						{:else}
							<div class="alert alert-warning">
								<div class="flex flex-col gap-1">
									<p class="text-sm">{$t('transfers.no-credit-categories')}</p>
									<a href="/categories" class="link text-xs"
										>{$t('transfers.create-category-first')}</a
									>
								</div>
							</div>
						{/if}
					</div>
				</div>

				<!-- Description -->
				<div class="form-control mt-6">
					<label class="label" for="description">
						<span class="label-text">{$t('transactions.description')}</span>
					</label>
					<input
						id="description"
						type="text"
						placeholder={$t('transfers.transfer-description')}
						class="input input-bordered"
						bind:value={description}
					/>
				</div>

				<!-- Date and Amount -->
				<div class="mt-4 flex gap-4">
					<!-- Date -->
					<div class="form-control flex-1">
						<label class="label" for="date">
							<span class="label-text">{$t('transactions.date')}</span>
						</label>
						<input id="date" type="date" class="input input-bordered w-full" bind:value={date} />
					</div>

					<!-- Amount -->
					<div class="form-control flex-1">
						<label class="label" for="amount">
							<span class="label-text">{$t('transactions.amount')}</span>
						</label>
						<input
							id="amount"
							type="number"
							placeholder={$t('transfers.transfer-amount')}
							class="input input-bordered w-full"
							bind:value={amount}
							min="0.01"
							step="0.01"
							required
						/>
					</div>
				</div>

				<!-- Actions -->
				<div class="modal-action mt-6">
					<button type="button" class="btn" on:click={handleCloseModal}>
						{$t('common.cancel')}
					</button>
					<button
						type="submit"
						class="btn btn-primary text-base-100"
						disabled={debitCategories.length === 0 || creditCategories.length === 0}
					>
						{$t('transfers.create-transfer')}
					</button>
				</div>
			</form>
		{/if}
	</div>
</div>

<style>
	:global(.dark) input[type='date']::-webkit-calendar-picker-indicator {
		filter: invert(1);
	}
</style>
