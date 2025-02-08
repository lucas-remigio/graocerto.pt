<script lang="ts">
	import api_axios from '$lib/axios';
	import type {
		Account,
		CategoriesResponse,
		Category,
		TransactionType,
		TransactionTypesResponse
	} from '$lib/types';
	import { createEventDispatcher, onMount } from 'svelte';

	// Input account
	export let account: Account;

	let error: string = '';
	let transactionTypes: TransactionType[] = [];
	let categories: Category[] = [];

	// Filtered categories based on transaction type
	let transaction_type_id: number | string = '';
	let filteredCategories: Category[] = [];
	// Reactive block to filter categories whenever transaction_type_id changes
	$: filteredCategories = categories.filter(
		(cat) => cat.transaction_type_id === Number(transaction_type_id)
	);

	// Form field variables
	let account_token = account.token;
	let category_id: number | string = '';
	let amount: number = 0;
	let description = '';
	let date = ''; // expects format "YYYY-MM-DD" from the date input

	// Create event dispatcher (to emit events to the parent)
	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		// Build the transaction object in the format your API expects
		const transaction = {
			account_token,
			category_id: Number(category_id),
			amount,
			description,
			date // already in YYYY-MM-DD format
		};

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
			console.log('Transaction types:', transactionTypes);
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
			console.log('Categories:', categories);
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
	<div class="modal-box relative">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}
			>✕</button
		>
		<h3 class="mb-4 text-lg font-bold">New Transaction</h3>
		<form on:submit|preventDefault={handleSubmit}>
			<!-- Transaction Type (Debit / Credit) Field -->
			<div class="form-control mt-4">
				<label class="label">
					<span class="label-text">Transaction Type</span>
				</label>
				<select class="select select-bordered" bind:value={transaction_type_id} required>
					<option value="" disabled selected>Select Transaction Type</option>
					<!-- cycle through the transaction types-->
					{#each transactionTypes as type}
						<option value={type.id}>{type.type_name}</option>
					{/each}
					<!-- Add more options as needed -->
				</select>
			</div>
			<!-- Category Field -->
			<div class="form-control mt-4">
				<label class="label">
					<span class="label-text">Category</span>
				</label>
				<select class="select select-bordered" bind:value={category_id} required>
					<option value="" disabled selected>Select category</option>
					{#each filteredCategories as cat}
						<option value={cat.id}>{cat.category_name}</option>
					{/each}
					<!-- Add more options as needed -->
				</select>
			</div>

			<!-- Amount Field -->
			<div class="form-control mt-4">
				<label class="label">
					<span class="label-text">Amount</span>
				</label>
				<input
					type="number"
					placeholder="Enter amount"
					class="input input-bordered"
					bind:value={amount}
					required
				/>
			</div>

			<!-- Description Field -->
			<div class="form-control mt-4">
				<label class="label">
					<span class="label-text">Description</span>
				</label>
				<input
					type="text"
					placeholder="Transaction description"
					class="input input-bordered"
					bind:value={description}
					required
				/>
			</div>

			<!-- Date Field -->
			<div class="form-control mt-4">
				<label class="label">
					<span class="label-text">Date</span>
				</label>
				<input type="date" class="input input-bordered" bind:value={date} required />
			</div>

			<!-- Form Actions -->
			<div class="modal-action mt-6">
				<button type="button" class="btn" on:click={handleCloseModal}>Cancel</button>
				<button type="submit" class="btn btn-primary">Create Transaction</button>
			</div>
		</form>
	</div>
</div>
