<!-- src/components/TransactionsTable.svelte -->
<script lang="ts">
	import type { Account, CategoryDto, TransactionDto } from '$lib/types';
	import { Bot, CircleDollarSign, Plus, Trash } from 'lucide-svelte';
	import CreateTransaction from './CreateTransaction.svelte';
	import { createEventDispatcher } from 'svelte';
	import EditTransaction from './EditTransaction.svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import TransactionFilters from './TransactionFilters.svelte';
	import AiFeedback from './AiFeedback.svelte';
	import { t } from '$lib/i18n';
	import { format, locale } from 'svelte-i18n';

	// Export props for transactions array and the account name.
	export let transactions: TransactionDto[] = [];
	export let account: Account;
	export let categories: CategoryDto[] = [];
	export let isAll: boolean = false; // Flag to indicate if all transactions are shown

	let showCreateTransactionModal = false;
	let showEditTransactionModal = false;
	let showDeleteTransactionModal = false;
	let showAiFeedbackModal = false;
	let error: string = '';

	let month: number = new Date().getMonth() + 1; // Current month (1-12)
	let year: number = new Date().getFullYear(); // Current year

	let selectedTransaction: TransactionDto | null = null;
	let activeFilter: any = null;
	let filteredTransactions: TransactionDto[] = transactions;

	$: currentLocale = $locale || 'en';

	function formatCurrency(amount: number): string {
		// make the currency have a , every 3 digits
		return amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
	}

	function getTransactionDetails(transaction: TransactionDto): string {
		return `${transaction.description} (${formatCurrency(transaction.amount)}€) with category ${transaction.category.category_name} at ${formatDate(transaction.date)}`;
	}

	function formatDate(date: string): string {
		// the format should be just month and year, in extense portuguese, without the "de" between
		const formattedDate = new Date(date).toLocaleDateString(currentLocale, {
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		});

		return `${formattedDate}`;
	}

	function transactionsFormattedDate(): string {
		// get the first transactions from the list
		if (transactions.length === 0) {
			return '';
		}
		const firstTransaction = transactions[0];
		return new Date(firstTransaction.date).toLocaleDateString(currentLocale, {
			month: 'long',
			year: 'numeric'
		});
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

	function closeAiFeedbackModal() {
		showAiFeedbackModal = false;
	}

	function openAiFeedbackModal() {
		// this should get the first transactions's month and year
		if (filteredTransactions.length === 0) {
			error = 'No transactions available for AI feedback.';
			return;
		}
		const firstTransaction = filteredTransactions[0];
		month = new Date(firstTransaction.date).getMonth() + 1; // Get month (1-12)
		year = new Date(firstTransaction.date).getFullYear();
		error = '';
		showAiFeedbackModal = true;
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
			// Use YYYY-MM format for consistent sorting
			const sortKey = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;

			if (!groups.has(sortKey)) {
				groups.set(sortKey, []);
			}
			groups.get(sortKey)!.push(tx);
		});

		// Convert to array and sort by key (most recent first)
		return Array.from(groups.entries())
			.sort((a, b) => b[0].localeCompare(a[0]))
			.map(([key, transactions]) => ({
				month: new Date(transactions[0].date).toLocaleDateString('pt-PT', {
					month: 'long',
					year: 'numeric'
				}),
				transactions: transactions.sort(
					(a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()
				)
			}));
	}

	$: groupedTransactions = isAll
		? groupTransactionsByMonth(filteredTransactions)
		: [
				{
					month: '',
					transactions: filteredTransactions
				}
			];
</script>

{#if transactions && transactions.length > 0}
	<div class="mb-2 flex justify-between">
		<h2 class="mb-4 text-2xl font-semibold">
			{$t('page.transactions-for')}
			{account.account_name}
			{#if !isAll}
				- {transactionsFormattedDate()}
			{/if}
		</h2>
		<div class="flex items-center gap-4">
			<!-- Button to get feedback -->
			{#if !isAll}
				<button
					class="btn btn-primary shadow-lg"
					on:click={openAiFeedbackModal}
					aria-label="Get AI Feedback"
				>
					<div class="flex items-center gap-1">
						<Bot size={20} class="text-base-content" />
					</div>
				</button>
			{/if}
			<!-- Button to add a new transaction-->
			<button class="btn btn-primary shadow-lg" on:click={openCreateTransactionModal}>
				<Plus size={20} class="text-base-content" />
				<CircleDollarSign size={20} class="text-base-content" />
			</button>
		</div>
	</div>

	<div class="overflow-x-auto">
		<TransactionFilters {categories} on:filter={handleFilter} />

		{#if filteredTransactions.length === 0}
			<p class="text-center text-gray-500">No transactions found.</p>
		{:else}
			<table class="table w-full">
				<thead class="sticky top-0 text-center">
					<tr>
						<th style="width: 15%">Date</th>
						<th style="width: 20%">Category</th>
						<th style="width: 15%">Amount</th>
						<th style="width: 40%">Description</th>
						<th style="width: 10%">Actions</th>
					</tr>
				</thead>
				<tbody class="text-center">
					{#each groupedTransactions as group}
						<!-- Only show month header if isAll is true and month is not empty -->
						{#if isAll && group.month}
							<tr class="bg-base-200">
								<td colspan="5" class="px-4 py-2 text-left font-bold">
									{group.month}
								</td>
							</tr>
						{/if}
						{#each group.transactions as tx}
							<tr
								class={tx.category.transaction_type.type_slug === 'debit'
									? 'bg-red-100'
									: tx.category.transaction_type.type_slug === 'credit'
										? 'bg-green-100'
										: ''}
							>
								<td class="text-gray-900">
									{formatDate(tx.date)}
								</td>
								<td class="text-gray-900">
									<span
										class="rounded px-2 py-1 text-white"
										style="background-color: {tx.category.color};"
									>
										{tx.category.category_name}
									</span>
								</td>
								<td class="text-gray-900">{formatCurrency(tx.amount)}€</td>
								<td class="text-gray-900">{tx.description || 'N/A'}</td>
								<td class="text-gray-900">
									<button
										class="btn btn-ghost btn-sm btn-circle bg-base-100/80 text-error hover:bg-error/20 backdrop-blur-sm"
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

{#if showAiFeedbackModal}
	<AiFeedback {account} {month} {year} closeModal={closeAiFeedbackModal}></AiFeedback>
{/if}
