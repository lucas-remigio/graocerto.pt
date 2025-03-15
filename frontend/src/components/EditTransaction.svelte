<script lang="ts">
	import api_axios from '$lib/axios';
	import type {
		Account,
		CategoriesResponse,
		Category,
		TransactionDto,
		TransactionType,
		TransactionTypesResponse
	} from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';

	// Input account
	export let account: Account;
	export let transaction: TransactionDto;

	let error: string = '';
	let transactionTypes: TransactionType[] = [];
	let transaction_type_id: number = transaction.category.transaction_type.id;
	let categories: Category[] = [];
	let categoriesMappedById: Map<number, Category> = new Map();

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
		(cat) => cat.transaction_type_id === transaction_type_id
	);

	$: selectedCategory = categoriesMappedById.get(Number(category_id));
	$: borderColor = selectedCategory ? selectedCategory.color : '#ccc';

	// Form field variables
	let account_token = account.token;
	let category_id: number | string = transaction.category.id;
	let amount: number = transaction.amount;
	let description = transaction.description;
	let date = transaction.date; // expects format "YYYY-MM-DD" from the date input

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
			error = 'Failed to update transaction';
		}
	}

	function isFormValid(): boolean {
		// round the amount
		amount = Math.round(amount * 100) / 100;

		// category must be from transaction type
		const category: Category | undefined = categories.find((cat) => cat.id === Number(category_id));
		if (!category) {
			error = 'Category is required';
			return false;
		}

		if (category.transaction_type_id !== Number(transaction_type_id)) {
			error = 'Category must be from the selected transaction type';
			return false;
		}

		// validations
		if (amount <= 0) {
			error = 'Amount must be greater than 0';
			return false;
		}

		if (amount > 999999999) {
			error = 'Amount must be less than 999999999';
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
			const res = await api_axios('transaction-types');

			if (res.status !== 200) {
				console.error('Non-200 response status:', res.status);
				error = `Error: ${res.status}`;
				return;
			}

			const data: TransactionTypesResponse = res.data;
			transactionTypes = data.transaction_types;
			// pop the slug with transfer
			transactionTypes = transactionTypes.filter((type) => type.type_slug !== 'transfer');
		} catch (err) {
			console.error('Error in fetchTransactionTypes:', err);
			error = 'Failed to load transaction types';
		}
	}

	async function fetchCategories() {
		try {
			const res = await api_axios('categories');

			if (res.status !== 200) {
				console.error('Non-200 response status:', res.status);
				error = `Error: ${res.status}`;
				return;
			}

			const data: CategoriesResponse = res.data;
			categories = data.categories;

			categoriesMappedById = new Map(categories.map((cat) => [cat.id, cat]));
		} catch (err) {
			console.error('Error in fetchCategories:', err);
			error = 'Failed to load categories';
		}
	}

	onMount(() => {
		fetchTransactionTypes();
		fetchCategories();
	});
</script>

<div class="modal modal-open">
	<div class="modal-box relative border-4 {modalBorderClass}">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}
			><X /></button
		>
		<h3 class="mb-4 text-lg font-bold">
			Edit Transaction for <strong>{account.account_name}</strong>
		</h3>
		<!--Error message-->
		{#if error}
			<div class="alert alert-error">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}
		<form on:submit|preventDefault={handleSubmit}>
			{#if filteredCategories.length > 0}
				<!-- Display both Transaction Type and Category side by side -->
				<div class="mt-4 flex flex-col gap-4 md:flex-row">
					<!-- Transaction Type Field -->
					<div class="form-control flex-1">
						<label class="label" for="transaction-type">
							<span class="label-text">Transaction Type</span>
						</label>
						<select
							id="transaction-type"
							class="select select-bordered w-full"
							bind:value={transaction_type_id}
							required
						>
							<option value="" disabled selected>Select Transaction Type</option>
							{#each transactionTypes as type}
								<option value={type.id}>{type.type_name}</option>
							{/each}
						</select>
					</div>

					<!-- Category Field -->
					<div class="form-control flex-1">
						<label class="label" for="category">
							<span class="label-text">Category</span>
						</label>
						<select
							id="category"
							class="select select-bordered w-full border-2"
							bind:value={category_id}
							required
							style="border-color: {borderColor} !important;"
						>
							<option value="" disabled selected>Select category</option>
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
						<span class="label-text">Transaction Type</span>
					</label>
					<select
						id="transaction-type"
						class="select select-bordered w-full"
						bind:value={transaction_type_id}
						required
					>
						<option value="" disabled selected>Select Transaction Type</option>
						{#each transactionTypes as type}
							<option value={type.id}>{type.type_name}</option>
						{/each}
					</select>
				</div>
				<div class="form-control mt-4">
					<p class="text-gray-500">
						No categories available for the selected transaction type.
						<a href="/categories" class="link">Click here to create one!</a>
					</p>
				</div>
			{/if}

			<!-- Description Field -->
			<div class="form-control mt-4">
				<label class="label" for="description">
					<span class="label-text">Description</span>
				</label>
				<input
					id="description"
					type="text"
					placeholder="Transaction description"
					class="input input-bordered"
					bind:value={description}
				/>
			</div>

			<div class="mt-4 flex gap-4">
				<!-- Amount Field -->
				<div class="form-control flex-1">
					<label class="label" for="amount">
						<span class="label-text">Amount</span>
					</label>
					<input
						id="amount"
						type="number"
						placeholder="Enter amount"
						class="input input-bordered w-full"
						bind:value={amount}
						min="0"
						step="0.01"
						max="999999999"
						required
					/>
				</div>

				<!-- Date Field -->
				<div class="form-control flex-1">
					<label class="label" for="date">
						<span class="label-text">Date</span>
					</label>
					<input id="date" type="date" class="input input-bordered w-full" bind:value={date} />
				</div>
			</div>
			<!-- Form Actions -->
			<div class="modal-action mt-6">
				<button type="button" class="btn" on:click={handleCloseModal}>Cancel</button>
				<button type="submit" class="btn btn-primary">Update Transaction</button>
			</div>
		</form>
	</div>
</div>
