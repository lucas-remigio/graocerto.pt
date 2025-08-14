<script lang="ts">
	import api_axios from '$lib/axios';
	import { dataService } from '$lib/services/dataService';
	import type {
		Account,
		CategoryDto,
		Transaction,
		TransactionChangeResponse,
		TransactionDto,
		TransactionType
	} from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import { t } from '$lib/i18n';
	import { getDraftTransaction } from '$lib/services/draftTransactionService';
	import { TransactionTypeId } from '$lib/transaction_types_types';
	import { validateTransactionForm } from '$lib/transactionValidation';
	import CategoryModal from './CategoryModal.svelte';

	// Inputs
	export let account: Account;
	export let transaction: TransactionDto | null = null; // if provided => edit mode

	// Mode
	let isEditMode: boolean = !!transaction;
	$: isEditMode = !!transaction;

	// State
	let error: string = '';
	let transactionTypes: TransactionType[] = [];
	let categories: CategoryDto[] = [];
	let categoriesMappedById: Map<number, CategoryDto> = new Map();
	let isLoading: boolean = true;

	// Last draft (create mode only)
	const draftTransaction: Transaction | null = isEditMode ? null : getDraftTransaction();

	// Form fields
	let account_token = account.token;
	let transaction_type_id: number = isEditMode
		? transaction!.category.transaction_type.id
		: draftTransaction?.transaction_type_id || TransactionTypeId.Debit;

	let category_id: string = isEditMode
		? String(transaction!.category.id)
		: draftTransaction?.category_id != null
			? String(draftTransaction.category_id)
			: '';

	let amount: number = isEditMode ? transaction!.amount : draftTransaction?.amount || 0;
	let description: string = isEditMode
		? transaction!.description
		: draftTransaction?.description || '';
	let date: string = isEditMode
		? transaction!.date
			? transaction!.date.split('T')[0]
			: new Date().toISOString().split('T')[0]
		: draftTransaction?.date || new Date().toISOString().split('T')[0];

	// Derived
	$: selectedTransactionType = transactionTypes.find((t) => t.id === transaction_type_id);
	$: filteredCategories = categories.filter(
		(cat) => cat.transaction_type.id === transaction_type_id
	);
	$: selectedCategory = categoriesMappedById.get(Number(category_id));
	$: borderColor = selectedCategory ? selectedCategory.color : '#ccc';

	// Reactive guard: if current category is invalid for this type or is sentinel, reset to placeholder
	$: if (
		category_id &&
		(category_id === '__create__' || !filteredCategories.some((c) => String(c.id) === category_id))
	) {
		category_id = '';
	}

	// Fallback for transaction type id when types load
	$: if (transactionTypes.length > 0 && !transaction_type_id) {
		transaction_type_id = transactionTypes[0].id;
	}

	const borderClasses: Record<string, string> = {
		credit: 'border-green-500 dark:border-green-400',
		debit: 'border-red-500 dark:border-red-400',
		transfer: 'border-blue-500 dark:border-blue-400'
	};
	$: modalBorderClass = selectedTransactionType
		? borderClasses[selectedTransactionType.type_slug]
		: 'bg-gray-50';

	// Events
	const dispatch = createEventDispatcher();

	function handleCloseModal() {
		if (!isEditMode) {
			// Preserve create modal behavior: return draft on close
			const tx: Transaction = {
				account_token,
				category_id: Number(category_id),
				amount,
				description,
				transaction_type_id,
				date: date || new Date().toISOString().split('T')[0]
			};
			dispatch('closeModal', { transaction: tx });
		} else {
			dispatch('closeModal');
		}
	}

	function isFormValid(): boolean {
		const result = validateTransactionForm(
			amount,
			category_id,
			transaction_type_id,
			categories,
			date,
			$t
		);

		if (result.error) {
			error = result.error;
			return false;
		}

		amount = result.amount;
		date = result.date;
		return true;
	}

	async function handleSubmit() {
		error = '';
		if (!isFormValid()) return;

		try {
			if (isEditMode) {
				const updatedTransaction = {
					id: transaction!.id,
					category_id: Number(category_id),
					amount,
					description,
					date
				};
				const response = await api_axios(`transactions/${transaction!.id}`, {
					method: 'PUT',
					data: updatedTransaction
				});
				if (response.status !== 200) {
					console.error('Non-200 response status:', response.status);
					error = `Error: ${response.status}`;
					return;
				}
				dispatch('updateTransaction', response.data as TransactionChangeResponse);
			} else {
				const newTx = {
					account_token,
					category_id: Number(category_id),
					amount,
					description,
					date
				};
				const response = await api_axios('transactions', { method: 'POST', data: newTx });
				if (response.status !== 200) {
					console.error('Non-200 response status:', response.status);
					error = `Error: ${response.status}`;
					return;
				}
				dispatch('newTransaction', response.data as TransactionChangeResponse);
			}
		} catch (err) {
			console.error('Error in handleSubmit:', err);
			error = isEditMode
				? $t('errors.failed-update-transaction')
				: $t('errors.failed-create-transaction');
		}
	}

	async function fetchTransactionTypes() {
		try {
			transactionTypes = await dataService.fetchTransactionTypes();
			transactionTypes = transactionTypes.filter((type) => type.type_slug !== 'transfer');
		} catch (err) {
			console.error('Error in fetchTransactionTypes:', err);
			error = $t('transactions.failed-load-types');
		}
	}

	async function fetchCategories() {
		try {
			categories = await dataService.fetchCategories();
			categoriesMappedById = new Map(categories.map((cat) => [cat.id, cat]));
		} catch (err) {
			console.error('Error in fetchCategories:', err);
			error = $t('errors.failed-load-categories');
		}
	}

	async function fetchData() {
		isLoading = true;
		error = '';
		try {
			await Promise.all([fetchTransactionTypes(), fetchCategories()]);
		} catch (err) {
			console.error('Error in fetchData:', err);
			error = $t('errors.failed-load-data');
		} finally {
			isLoading = false;
		}
	}

	let showCreateCategoryModal = false;

	function handleCategorySelect(e: Event) {
		const value = (e.target as HTMLSelectElement).value;
		if (value === '__create__') {
			showCreateCategoryModal = true;
			return;
		}
		category_id = value;
	}

	function handleCreatedCategory(e: CustomEvent<any>) {
		const payload = e.detail;
		const newCat = payload?.category ?? payload;
		if (!newCat?.id) {
			showCreateCategoryModal = false;
			return;
		}
		categories = [...categories, newCat];
		categoriesMappedById.set(newCat.id, newCat);
		if (newCat.transaction_type?.id === transaction_type_id) {
			category_id = String(newCat.id);
		} else {
			category_id = '';
		}
		showCreateCategoryModal = false;
	}

	function handleCloseCreateCategory() {
		if (category_id === '__create__') category_id = '';
		showCreateCategoryModal = false;
	}

	onMount(() => {
		fetchData();
	});
</script>

<div class="modal modal-open">
	<div class="modal-box relative border-4 {modalBorderClass}">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}
			><X /></button
		>
		<h3 class="mb-4 text-lg font-bold">
			{#if isEditMode}
				{$t('transactions.edit-transaction-for')} <strong>{account.account_name}</strong>
			{:else}
				{$t('transactions.new-transaction-for')} <strong>{account.account_name}</strong>
			{/if}
		</h3>

		{#if error}
			<div class="alert alert-error">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}

		{#if isLoading}
			<div class="py-12 text-center">
				<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
				<p class="text-base-content/70">{$t('common.loading')}</p>
			</div>
		{:else}
			<form on:submit|preventDefault={handleSubmit}>
				{#if filteredCategories.length > 0}
					<div class="mt-4 flex flex-col gap-4 md:flex-row">
						<!-- Transaction Type Field -->
						<div class="form-control flex-1">
							<label class="label" for="transaction-type">
								<span class="label-text">{$t('transactions.transaction-type')}</span>
							</label>
							<select
								id="transaction-type"
								class="select select-bordered w-full"
								bind:value={transaction_type_id}
								required
							>
								<option value="" disabled>{$t('transactions.select-transaction-type')}</option>
								{#each transactionTypes as type}
									<option value={type.id}
										>{$t('transaction-types.' + type.type_slug, {
											default: type.type_name
										})}</option
									>
								{/each}
							</select>
						</div>

						<!-- Category Field -->
						<div class="form-control flex-1">
							<label class="label" for="category">
								<span class="label-text">{$t('transactions.category')}</span>
							</label>
							<select
								id="category"
								class="select select-bordered w-full border-2"
								bind:value={category_id}
								on:change={handleCategorySelect}
								required
								style="border-color: {borderColor} !important;"
							>
								<option value="" disabled>
									{$t('transactions.select-category')}
								</option>
								{#each filteredCategories as cat}
									<option value={String(cat.id)}>{cat.category_name}</option>
								{/each}
								<option value="__create__">
									+ {$t('categories.create-new', { default: 'Create new category' })}
								</option>
							</select>
						</div>
					</div>
				{:else}
					<div class="form-control mt-4">
						<label class="label" for="transaction-type">
							<span class="label-text">{$t('transactions.transaction-type')}</span>
						</label>
						<select
							id="transaction-type"
							class="select select-bordered w-full"
							bind:value={transaction_type_id}
							required
						>
							<option value="" disabled>{$t('transactions.select-transaction-type')}</option>
							{#each transactionTypes as type}
								<option value={type.id}
									>{$t('transaction-types.' + type.type_slug, { default: type.type_name })}</option
								>
							{/each}
						</select>
					</div>
					<div class="form-control mt-4">
						<p class="text-base-content/70">
							{$t('transactions.no-categories-available')}
							<a href="/categories" class="link">{$t('transactions.click-to-create')}</a>
						</p>
					</div>
				{/if}

				<!-- Description Field -->
				<div class="form-control mt-4">
					<label class="label" for="description">
						<span class="label-text">{$t('transactions.description')}</span>
					</label>
					<input
						id="description"
						type="text"
						placeholder={$t('transactions.transaction-description')}
						class="input input-bordered"
						bind:value={description}
					/>
				</div>

				<div class="mt-4 flex gap-4">
					<!-- Date Field -->
					<div class="form-control flex-1">
						<label class="label" for="date">
							<span class="label-text">{$t('transactions.date')}</span>
						</label>
						<input id="date" type="date" class="input input-bordered w-full" bind:value={date} />
					</div>

					<!-- Amount Field -->
					<div class="form-control flex-1">
						<label class="label" for="amount">
							<span class="label-text">{$t('transactions.amount')}</span>
						</label>
						<input
							id="amount"
							type="text"
							inputmode="decimal"
							placeholder={$t('transactions.transaction-amount')}
							class="input input-bordered w-full"
							bind:value={amount}
							min="0"
							step="0.01"
							max="999999999"
							required
						/>
					</div>
				</div>

				<div class="modal-action mt-6">
					<button type="button" class="btn" on:click={handleCloseModal}>
						{$t('common.cancel')}
					</button>
					{#if isEditMode}
						<button type="submit" class="btn btn-primary text-base-100">
							{$t('transactions.update-transaction')}
						</button>
					{:else}
						<button type="submit" class="btn btn-primary text-base-100">
							{$t('transactions.create-transaction')}
						</button>
					{/if}
				</div>
			</form>
		{/if}
	</div>
</div>

{#if showCreateCategoryModal && selectedTransactionType}
	<CategoryModal
		category={null}
		transactionType={selectedTransactionType}
		on:closeModal={handleCloseCreateCategory}
		on:newCategory={handleCreatedCategory}
	/>
{/if}

<style>
	:global(.dark) input[type='date']::-webkit-calendar-picker-indicator {
		filter: invert(1);
	}
	:global(.dark) input[type='date']::-moz-calendar-picker-indicator {
		filter: invert(1);
	}
</style>
