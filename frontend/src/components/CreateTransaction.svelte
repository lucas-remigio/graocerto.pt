<script lang="ts">
	import api_axios from '$lib/axios';
	import { TransactionTypes } from '$lib/transaction_types_types';
	import type {
		Account,
		CategoriesResponse,
		Category,
		TransactionType,
		TransactionTypesResponse
	} from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';

	// Input account
	export let account: Account;

	let error: string = '';
	let transactionTypes: TransactionType[] = [];
	let transaction_type_id: number = 0;
	let categories: Category[] = [];
	let categoriesMappedById: Map<number, Category> = new Map();

	$: if (transactionTypes.length > 0 && !transaction_type_id) {
		transaction_type_id = transactionTypes[0].id;
	}

	$: selectedTransactionType = transactionTypes.find((t) => t.id === transaction_type_id);

	const backgroundClasses: Record<string, string> = {
		credit: 'bg-green-50',
		debit: 'bg-red-50',
		transfer: 'bg-blue-50'
	};
	$: modalBackgroundClass = selectedTransactionType
		? backgroundClasses[selectedTransactionType.type_slug]
		: 'bg-gray-50';

	// Filter categories based on the selected transaction type id.
	$: filteredCategories = categories.filter(
		(cat) => cat.transaction_type_id === transaction_type_id
	);

	$: selectedCategory = categoriesMappedById.get(Number(category_id));
	$: borderColor = selectedCategory ? selectedCategory.color : '#ccc';

	// Form field variables
	let account_token = account.token;
	let category_id: number | string = '';
	let amount: number = 0;
	let description = '';
	let date = ''; // expects format "YYYY-MM-DD" from the date input

	// Create event dispatcher (to emit events to the parent)
	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		error = '';
		if (!isFormValid()) {
			return;
		}

		// Build the transaction object in the format your API expects
		const transaction = {
			account_token,
			category_id: Number(category_id),
			amount,
			description,
			date // already in YYYY-MM-DD format
		};

		try {
			const response = await api_axios('transactions', {
				method: 'POST',
				data: transaction
			});

			if (response.status !== 200) {
				console.error('Non-200 response status:', response.status);
				error = `Error: ${response.status}`;
				return;
			}
			handleNewTransaction();
		} catch (err) {
			console.error('Error in handleSubmit:', err);
			error = 'Failed to create transaction';
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
		if (amount < 0) {
			error = 'Amount must be greater than 0';
			return false;
		}

		if (amount > 999999999) {
			error = 'Amount must be less than 999999999';
			return false;
		}

		if (description.length < 3) {
			error = 'Description must be at least 3 characters';
			return false;
		}

		if (description.length > 50) {
			error = 'Description must be less than 50 characters';
			return false;
		}

		if (!date) {
			error = 'Date is required';
			return false;
		}

		return true;
	}

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function handleNewTransaction() {
		dispatch('newTransaction');
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
	<div class="modal-box relative {modalBackgroundClass}">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}
			><X /></button
		>
		<h3 class="mb-4 text-lg font-bold">New Transaction</h3>
		<!--Error message-->
		{#if error}
			<div class="alert alert-error">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}
		<form on:submit|preventDefault={handleSubmit}>
			<!-- Transaction Type (Debit / Credit) Field -->
			<div class="form-control mt-4">
				<label class="label" for="transaction-type">
					<span class="label-text">Transaction Type</span>
				</label>
				<select
					id="transaction-type"
					class="select select-bordered"
					bind:value={transaction_type_id}
					required
				>
					<option value="" disabled selected>Select Transaction Type</option>
					<!-- cycle through the transaction types-->
					{#each transactionTypes as type}
						<option value={type.id}>{type.type_name}</option>
					{/each}
					<!-- Add more options as needed -->
				</select>
			</div>
			{#if filteredCategories.length === 0}
				<div class="form-control mt-4">
					<p class="text-gray-500">
						No categories available for the selected transaction type.
						<a href="/categories" class="link"> Click me to create one! </a>
					</p>
				</div>
			{/if}
			{#if filteredCategories.length > 0}
				<!-- Category Field -->
				<div class="form-control mt-4">
					<label class="label" for="category">
						<span class="label-text">Category</span>
					</label>
					<select
						id="category"
						class="select select-bordered border-2"
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

				<!-- Amount Field -->
				<div class="form-control mt-4">
					<label class="label" for="amount">
						<span class="label-text">Amount</span>
					</label>
					<input
						id="amount"
						type="number"
						placeholder="Enter amount"
						class="input input-bordered"
						bind:value={amount}
						min="0"
						step="0.01"
						max="999999999"
						required
					/>
				</div>

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
						required
					/>
				</div>

				<!-- Date Field -->
				<div class="form-control mt-4">
					<label class="label" for="date">
						<span class="label-text">Date</span>
					</label>
					<input id="date" type="date" class="input input-bordered" bind:value={date} required />
				</div>
			{/if}
			<!-- Form Actions -->
			<div class="modal-action mt-6">
				<button type="button" class="btn" on:click={handleCloseModal}>Cancel</button>
				<button type="submit" class="btn btn-primary">Create Transaction</button>
			</div>
		</form>
	</div>
</div>
