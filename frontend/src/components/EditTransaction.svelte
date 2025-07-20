<script lang="ts">
	import api_axios from '$lib/axios';
	import { dataService } from '$lib/services/dataService';
	import type { Account, Category, CategoryDto, TransactionDto, TransactionType } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import { t } from '$lib/i18n';

	// Input account
	export let account: Account;
	export let transaction: TransactionDto;

	let error: string = '';
	let transactionTypes: TransactionType[] = [];
	let transaction_type_id: number = transaction.category.transaction_type.id;
	let categories: CategoryDto[] = [];
	let categoriesMappedById: Map<number, CategoryDto> = new Map();
	let isLoading: boolean = true;

	$: if (transactionTypes.length > 0 && !transaction_type_id) {
		transaction_type_id = transactionTypes[0].id;
	}

	$: selectedTransactionType = transactionTypes.find((t) => t.id === transaction_type_id);

	const borderClasses: Record<string, string> = {
		credit: 'border-green-500 dark:border-green-400',
		debit: 'border-red-500 dark:border-red-400',
		transfer: 'border-blue-500 dark:border-blue-400'
	};
	$: modalBorderClass = selectedTransactionType
		? borderClasses[selectedTransactionType.type_slug]
		: 'bg-gray-50';

	// Filter categories based on the selected transaction type id.
	$: filteredCategories = categories.filter(
		(cat) => cat.transaction_type.id === transaction_type_id
	);

	$: selectedCategory = categoriesMappedById.get(Number(category_id));
	$: borderColor = selectedCategory ? selectedCategory.color : '#ccc';

	// Form field variables
	let account_token = account.token;
	let category_id: number | string = transaction.category.id;
	let amount: number = transaction.amount;
	let description = transaction.description;
	let date = transaction.date
		? transaction.date.split('T')[0]
		: new Date().toISOString().split('T')[0];

	// Create event dispatcher (to emit events to the parent)
	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		error = '';
		if (!isFormValid()) {
			return;
		}

		// Build the transaction object in the format your API expects
		const updatedTransaction = {
			id: transaction.id,
			category_id: Number(category_id),
			amount,
			description,
			date // already in YYYY-MM-DD format
		};

		try {
			const response = await api_axios(`transactions/${transaction.id}`, {
				method: 'PUT',
				data: updatedTransaction
			});

			if (response.status !== 200) {
				console.error('Non-200 response status:', response.status);
				error = `Error: ${response.status}`;
				return;
			}
			handleUpdateTransaction();
		} catch (err) {
			console.error('Error in handleSubmit:', err);
			error = $t('errors.failed-update-transaction');
		}
	}

	function isFormValid(): boolean {
		// round the amount
		amount = parseFloat(amount.toString().replace(',', '.'));
		amount = Math.round(amount * 100) / 100;

		// category must be from transaction type
		const category: CategoryDto | undefined = categories.find(
			(cat) => cat.id === Number(category_id)
		);
		if (!category) {
			error = $t('transactions.category-required');
			return false;
		}

		if (category.transaction_type.id !== Number(transaction_type_id)) {
			error = $t('transactions.category-must-match');
			return false;
		}

		// validations
		if (amount <= 0) {
			error = $t('transactions.amount-greater-zero');
			return false;
		}

		if (amount > 999999999) {
			error = $t('transactions.amount-too-large');
			return false;
		}

		if (!date) {
			// default to today
			date = new Date().toISOString().split('T')[0];
		}

		return true;
	}

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function handleUpdateTransaction() {
		dispatch('updateTransaction');
	}

	async function fetchTransactionTypes() {
		try {
			transactionTypes = await dataService.fetchTransactionTypes();
			// Filter out transfer types
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
			{$t('transactions.edit-transaction-for')} <strong>{account.account_name}</strong>
		</h3>
		<!--Error message-->
		{#if error}
			<div class="alert alert-error">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}

		{#if isLoading}
			<!-- Loading State -->
			<div class="py-12 text-center">
				<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
				<p class="text-base-content/70">{$t('common.loading')}</p>
			</div>
		{:else}
			<form on:submit|preventDefault={handleSubmit}>
				{#if filteredCategories.length > 0}
					<!-- Display both Transaction Type and Category side by side -->
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
								<option value="" disabled selected
									>{$t('transactions.select-transaction-type')}</option
								>
								{#each transactionTypes as type}
									<option value={type.id}>{$t('transaction-types.' + type.type_slug)}</option>
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
								required
								style="border-color: {borderColor} !important;"
							>
								<option value="" disabled selected>{$t('transactions.select-category')}</option>
								{#each filteredCategories as cat}
									<option value={cat.id}>{cat.category_name}</option>
								{/each}
							</select>
						</div>
					</div>
				{:else}
					<!-- If no categories, show the Transaction Type field and a message -->
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
							<option value="" disabled selected
								>{$t('transactions.select-transaction-type')}</option
							>
							{#each transactionTypes as type}
								<option value={type.id}>{type.type_name}</option>
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
				<!-- Form Actions -->
				<div class="modal-action mt-6">
					<button type="button" class="btn" on:click={handleCloseModal}>
						{$t('common.cancel')}
					</button>
					<button type="submit" class="btn btn-primary text-base-100">
						{$t('transactions.update-transaction')}
					</button>
				</div>
			</form>
		{/if}
	</div>
</div>
