<!-- src/components/TransactionsTable.svelte -->
<script lang="ts">
	import type { Account, CategoryDto, TransactionDto } from '$lib/types';
	import { CircleDollarSign, Pencil, Plus, Trash } from 'lucide-svelte';
	import CreateTransaction from './CreateTransaction.svelte';
	import { createEventDispatcher } from 'svelte';
	import EditTransaction from './EditTransaction.svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import TransactionFilters from './TransactionFilters.svelte';

	// Export props for transactions array and the account name.
	export let transactions: TransactionDto[] = [];
	export let account: Account;
	export let categories: CategoryDto[] = [];

	let showCreateTransactionModal = false;
	let showEditTransactionModal = false;
	let showDeleteTransactionModal = false;

	let selectedTransaction: TransactionDto | null = null;
	let activeFilter: any = null;
	let filteredTransactions: TransactionDto[] = transactions;

	function formatCurrency(amount: number): string {
		// make the currency have a , every 3 digits
		return amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
	}

	function getTransactionDetails(transaction: TransactionDto): string {
		return `${transaction.description} (${formatCurrency(transaction.amount)}€) with category ${transaction.category.category_name} at ${formatDate(transaction.date)}`;
	}

	function formatDate(date: string): string {
		// the format should be just month and year, in extense portuguese, without the "de" between
		const formattedDate = new Date(date).toLocaleDateString('pt-PT', {
			month: 'long',
			year: 'numeric'
		});

		let month = formattedDate.split(' ')[0];
		month = month.charAt(0).toUpperCase() + month.slice(1);
		const year = formattedDate.split(' ')[2];

		return `${month} ${year}`;
	}

	$: {
		// Update filtered transactions whenever transactions change or when filter is applied
		filteredTransactions = activeFilter
			? transactions.filter((transaction) => {
					// Date range filter
					if (
						activeFilter.dateFrom &&
						new Date(transaction.date) < new Date(activeFilter.dateFrom)
					) {
						return false;
					}
					if (activeFilter.dateTo && new Date(transaction.date) > new Date(activeFilter.dateTo)) {
						return false;
					}

					// Transaction type filter
					if (
						activeFilter.transactionTypes.length > 0 &&
						!activeFilter.transactionTypes.includes(transaction.category.transaction_type.type_slug)
					) {
						return false;
					}

					// Categories filter
					if (
						activeFilter.categories.length > 0 &&
						!activeFilter.categories.includes(transaction.category.id)
					) {
						return false;
					}

					// Amount range filter
					if (activeFilter.amountFrom !== null && transaction.amount < activeFilter.amountFrom) {
						return false;
					}
					if (activeFilter.amountTo !== null && transaction.amount > activeFilter.amountTo) {
						return false;
					}

					return true;
				})
			: transactions;
	}

	function handleFilter(
		event: CustomEvent<{
			dateFrom: string;
			dateTo: string;
			transactionTypes: string[];
			categories: number[];
			amountFrom: number | null;
			amountTo: number | null;
		}>
	) {
		activeFilter = event.detail;
	}

	function openCreateTransactionModal() {
		showCreateTransactionModal = true;
	}

	function closeCreateTransactionModal() {
		showCreateTransactionModal = false;
	}

	function handleEditTransaction(transaction: TransactionDto) {
		showEditTransactionModal = true;
		selectedTransaction = transaction;
	}

	function closeEditTransactionModal() {
		showEditTransactionModal = false;
		selectedTransaction = null;
	}

	function handleDeleteTransaction(transaction: TransactionDto) {
		showDeleteTransactionModal = true;
		selectedTransaction = transaction;
	}

	function closeDeleteTransactionModal() {
		showDeleteTransactionModal = false;
	}

	function handleDeleteTransactionCancel() {
		closeDeleteTransactionModal();
	}

	function handleDeleteTransactionConfirm() {
		closeDeleteTransactionModal();
		dispatch('deleteTransaction', { transaction: selectedTransaction! });
	}

	const dispatch = createEventDispatcher();
	function handleNewTransaction() {
		closeCreateTransactionModal();
		dispatch('newTransaction');
	}

	function handleUpdateTransaction() {
		closeEditTransactionModal();
		dispatch('updateTransaction');
	}

	// Add this function to group transactions by month
	function groupTransactionsByMonth(transactions: TransactionDto[]) {
		const groups = new Map<string, TransactionDto[]>();

		transactions.forEach((tx) => {
			const date = new Date(tx.date);
			const key = `${date.getFullYear()}-${date.getMonth()}`;
			if (!groups.has(key)) {
				groups.set(key, []);
			}
			groups.get(key)!.push(tx);
		});

		// Convert to array and sort by date (most recent first)
		return Array.from(groups.entries())
			.sort((a, b) => b[0].localeCompare(a[0]))
			.map(([key, transactions]) => ({
				month: formatDate(transactions[0].date),
				transactions
			}));
	}

	$: groupedTransactions = groupTransactionsByMonth(filteredTransactions);
</script>

{#if transactions && transactions.length > 0}
	<div class="mb-2 flex justify-between">
		<h2 class="mb-4 text-2xl font-semibold">
			Transactions for {account.account_name}
		</h2>
		<!-- Button to add a new transaction-->
		<button class="btn btn-primary shadow-lg" on:click={openCreateTransactionModal}>
			<Plus size={20} />
			<CircleDollarSign size={20} />
		</button>
	</div>

	<div class="overflow-x-auto">
		<TransactionFilters {categories} on:filter={handleFilter} />

		{#if filteredTransactions.length === 0}
			<p class="text-center text-gray-500">No transactions found.</p>
		{:else}
			<table class="table w-full">
				<thead class="sticky top-0 text-center">
					<tr>
						<th class="text-gray-900 dark:text-gray-100">Date</th>
						<th class="text-gray-900 dark:text-gray-100">Category</th>
						<th class="text-gray-900 dark:text-gray-100">Amount</th>
						<th class="text-gray-900 dark:text-gray-100">Description</th>
						<th class="text-gray-900 dark:text-gray-100">Actions</th>
					</tr>
				</thead>
				<tbody class="text-center">
					{#each groupedTransactions as group}
						<tr class="bg-base-200">
							<td colspan="5" class="px-4 py-2 text-left font-bold">
								{group.month}
							</td>
						</tr>
						{#each group.transactions as tx}
							<tr
								class={tx.category.transaction_type.type_slug === 'debit'
									? 'bg-red-100'
									: tx.category.transaction_type.type_slug === 'credit'
										? 'bg-green-100'
										: ''}
							>
								<td class="dark:text-gray-900">
									{formatDate(tx.date)}
								</td>
								<td>
									<span
										class="rounded px-2 py-1 text-white"
										style="background-color: {tx.category.color};"
									>
										{tx.category.category_name}
									</span>
								</td>
								<td class="dark:text-gray-900">{formatCurrency(tx.amount)}€</td>
								<td class="dark:text-gray-900">{tx.description || 'N/A'}</td>
								<td class="flex w-10 gap-1">
									<button
										class="btn btn-sm btn-circle text-blue-300"
										on:click={() => handleEditTransaction(tx)}
									>
										<Pencil size={20} />
									</button>
									<button
										class="btn btn-sm btn-circle text-red-300"
										on:click={() => handleDeleteTransaction(tx)}
									>
										<Trash size={20} />
									</button>
								</td>
							</tr>
						{/each}
					{/each}
				</tbody>
			</table>
		{/if}
	</div>
{:else}
	<div class="flex h-96 flex-col items-center justify-center">
		<p class="text-gray-500">
			No transactions found for <strong>{account.account_name}</strong>.
		</p>

		<!-- Button to add a new transaction -->
		<button
			class="btn btn-primary mt-4 flex items-center gap-2 shadow-lg"
			on:click={openCreateTransactionModal}
			aria-label="Add New Transaction"
		>
			<CircleDollarSign size={20} class="h-5 w-5" />
			<span>Create First Transaction</span>
		</button>
	</div>
{/if}

{#if showCreateTransactionModal}
	<CreateTransaction
		{account}
		on:closeModal={closeCreateTransactionModal}
		on:newTransaction={handleNewTransaction}
	></CreateTransaction>
{/if}

{#if showEditTransactionModal}
	<EditTransaction
		{account}
		transaction={selectedTransaction!}
		on:closeModal={closeEditTransactionModal}
		on:updateTransaction={handleUpdateTransaction}
	></EditTransaction>
{/if}

{#if showDeleteTransactionModal}
	<ConfirmAction
		title="Delete Transaction"
		message={`Are you sure you want to delete this transaction ${getTransactionDetails(selectedTransaction!)}? This action cannot be undone.`}
		type="danger"
		onConfirm={handleDeleteTransactionConfirm}
		onCancel={handleDeleteTransactionCancel}
	></ConfirmAction>
{/if}
