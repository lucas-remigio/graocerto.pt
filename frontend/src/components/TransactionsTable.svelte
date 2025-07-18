<!-- src/components/TransactionsTable.svelte -->
<script lang="ts">
	import type {
		Account,
		CategoryDto,
		TransactionDto,
		TransactionsTotals,
		TransactionGroup,
		Transaction
	} from '$lib/types';
	import { Bot, CircleDollarSign, List, Plus, Trash } from 'lucide-svelte';
	import CreateTransaction from './CreateTransaction.svelte';
	import { createEventDispatcher } from 'svelte';
	import EditTransaction from './EditTransaction.svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import TransactionsStats from './TransactionsStats.svelte';
	import AiFeedback from './AiFeedback.svelte';
	import { t } from '$lib/i18n';
	import { format, locale } from 'svelte-i18n';
	import {
		setDraftTransaction,
		setDraftTransactionAccountToken
	} from '$lib/services/draftTransactionService';

	// Export props for transactions array and the account name.
	export let transactionsGroups: TransactionGroup[] = [];
	export let transactionsTotals: TransactionsTotals;
	export let account: Account;
	export let isAll: boolean = false; // Flag to indicate if all transactions are shown
	export let loading: boolean = false;

	// clean the draft transaction when the account changes
	$: if (account) {
		setDraftTransactionAccountToken(account.token);
	}

	let showCreateTransactionModal = false;
	let showEditTransactionModal = false;
	let showDeleteTransactionModal = false;
	let showAiFeedbackModal = false;
	let error: string = '';

	let month: number = new Date().getMonth() + 1; // Current month (1-12)
	let year: number = new Date().getFullYear(); // Current year

	let selectedTransaction: TransactionDto | null = null;

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
			month: 'long'
		});

		return `${formattedDate}`;
	}

	function formatMonthYear(month: number, year: number): string {
		// Create a date object with the given month (1-12) and year
		// Using day 1 to avoid timezone issues
		const date = new Date(year, month - 1, 1);
		return date.toLocaleDateString(currentLocale, {
			month: 'long',
			year: 'numeric'
		});
	}

	function getTextColor(backgroundColor: string): string {
		// Remove # if present
		const hex = backgroundColor.replace('#', '');

		// Convert hex to RGB
		const r = parseInt(hex.substr(0, 2), 16);
		const g = parseInt(hex.substr(2, 2), 16);
		const b = parseInt(hex.substr(4, 2), 16);

		// Calculate relative luminance (WCAG formula)
		const getLuminance = (color: number) => {
			const c = color / 255;
			return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
		};

		const luminance =
			0.2126 * getLuminance(r) + 0.7152 * getLuminance(g) + 0.0722 * getLuminance(b);

		// Use white text on dark backgrounds, dark text on light backgrounds
		return luminance > 0.5 ? 'text-gray-900' : 'text-gray-100';
	}

	function openCreateTransactionModal() {
		showCreateTransactionModal = true;
	}

	function closeCreateTransactionModal(event?: CustomEvent) {
		if (event?.detail) {
			setDraftTransaction(event.detail.transaction);
		}
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
		if (transactionsGroups.length === 0 || transactionsGroups[0].transactions.length === 0) {
			error = $t('transactions.no-transactions-ai');
			return;
		}
		const firstTransaction = transactionsGroups[0].transactions[0];
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
</script>

{#if loading}
	<!-- Loading State -->
	<div class="py-12 text-center">
		<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
		<p class="text-base-content/70">{$t('common.loading')}</p>
	</div>
{:else if error}
	<!-- Error State -->
	<div class="alert alert-error">
		<p>{error}</p>
	</div>
{:else if transactionsGroups && transactionsGroups.length > 0}
	<div class="my-2 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
		<!-- Buttons: above stats on mobile, right on md+ -->
		<div
			class="order-1 flex items-center justify-center gap-4 md:order-2 md:ml-auto md:justify-end"
		>
			<!-- Button to get feedback -->
			{#if !isAll}
				<button
					class="btn btn-primary shadow-lg"
					on:click={openAiFeedbackModal}
					aria-label="Get AI Feedback"
				>
					<div class="flex items-center gap-1">
						<Bot size={20} class="text-base-100" />
					</div>
				</button>
			{/if}
			<!-- Button to add a new transaction-->
			<button
				class="btn btn-primary shadow-lg"
				aria-label="Create New Transaction"
				on:click={openCreateTransactionModal}
			>
				<Plus size={20} class="text-base-100" />
				<CircleDollarSign size={20} class="text-base-100" />
			</button>
		</div>

		<!-- Totals Summary below buttons on mobile, left on md+ -->
		<div class="order-2 flex justify-center md:order-1 md:justify-start">
			<TransactionsStats totals={transactionsTotals} />
		</div>
	</div>

	<div class="overflow-x-auto">
		{#if transactionsGroups.length === 0}
			<p class="text-center text-gray-500">{$t('transactions.no-transactions')}</p>
		{:else}
			<table class="table w-full">
				<thead class="sticky top-0 text-center">
					<tr>
						<th style="width: 15%">{$t('transactions.date')}</th>
						<th style="width: 20%">{$t('transactions.category')}</th>
						<th style="width: 15%">{$t('transactions.amount')}</th>
						<th style="width: 40%">{$t('transactions.description')}</th>
						<th style="width: 10%">{$t('transactions.actions')}</th>
					</tr>
				</thead>
				<tbody class="text-center">
					{#each transactionsGroups as group}
						<!-- Show month header only if isAll is true -->
						{#if isAll}
							<tr class="bg-base-200">
								<td colspan="5" class="px-4 py-2 text-left font-bold">
									{formatMonthYear(group.month, group.year)}
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
										class="rounded px-2 py-1 {getTextColor(tx.category.color)}"
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
										aria-label="Delete Transaction"
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
			{$t('transactions.no-transactions-for')} <strong>{account.account_name}</strong>.
		</p>

		<!-- Button to add a new transaction -->
		<button
			class="btn btn-primary mt-4 flex items-center gap-2 shadow-lg"
			on:click={openCreateTransactionModal}
			aria-label="Add New Transaction"
		>
			<CircleDollarSign size={20} class="text-base-100 h-5 w-5" />
			<span class="text-base-100">{$t('transactions.create-first')}</span>
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
		message={`${$t('modals.delete-transaction-confirm')} ${getTransactionDetails(selectedTransaction!)}? ${$t('modals.cannot-be-undone')}`}
		type="danger"
		onConfirm={handleDeleteTransactionConfirm}
		onCancel={handleDeleteTransactionCancel}
	></ConfirmAction>
{/if}

{#if showAiFeedbackModal}
	<AiFeedback {account} {month} {year} closeModal={closeAiFeedbackModal}></AiFeedback>
{/if}
