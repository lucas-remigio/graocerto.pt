<!-- src/components/TransactionsTable.svelte -->
<script lang="ts">
	import type {
		Account,
		TransactionDto,
		TransactionGroup,
		TransactionChangeResponse
	} from '$lib/types';
	import {
		ArrowRightLeft,
		Check,
		CircleDollarSign,
		Filter,
		Pencil,
		Plus,
		Trash,
		X
	} from 'lucide-svelte';
	import TransactionModal from './TransactionModal.svelte';
	import { createEventDispatcher } from 'svelte';

	import ConfirmAction from './ConfirmAction.svelte';
	import TransactionsStats from './TransactionsStats.svelte';
	import { t } from '$lib/i18n';
	import { locale } from 'svelte-i18n';
	import {
		setDraftTransaction,
		setDraftTransactionAccountToken
	} from '$lib/services/draftTransactionService';
	import { appliedTheme } from '$lib/stores/uiPreferences';
	import { fade, fly } from 'svelte/transition';
	import { formatCurrency } from '$lib/utils/currency';
	import {
		emptyTransactionFilters,
		matchesTransactionFilters
	} from '$lib/utils/transactionFilters';
	import TransferModal from './TransferModal.svelte';
	import TransactionFilters from './TransactionFilters.svelte';

	// Export props for transactions array and the account name.
	export let transactions: TransactionDto[] = [];
	export let account: Account;
	export let isAll: boolean = false; // Flag to indicate if all transactions are shown
	export let loading: boolean = false;

	// clean the draft transaction when the account changes
	$: if (account) {
		setDraftTransactionAccountToken(account.token);
	}

	// Group transactions by month/year when isAll is true, otherwise create a single group
	// Use filtered transactions for grouping
	$: transactionsGroups =
		filteredTransactions && filteredTransactions.length > 0
			? isAll
				? groupTransactionsByMonth(filteredTransactions)
				: [
						{
							month: new Date().getMonth() + 1,
							year: new Date().getFullYear(),
							transactions: filteredTransactions
						}
					]
			: [];

	function groupTransactionsByMonth(transactions: TransactionDto[]): TransactionGroup[] {
		const groups = new Map<string, TransactionGroup>();

		transactions.forEach((tx) => {
			const date = new Date(tx.date);
			const month = date.getMonth() + 1; // 1-12
			const year = date.getFullYear();
			const key = `${year}-${month}`;

			if (!groups.has(key)) {
				groups.set(key, {
					month,
					year,
					transactions: []
				});
			}

			groups.get(key)!.transactions.push(tx);
		});

		// Sort groups by year/month (newest first)
		return Array.from(groups.values()).sort((a, b) => {
			if (a.year !== b.year) return b.year - a.year;
			return b.month - a.month;
		});
	}

	let showCreateTransactionModal = false;
	let showEditTransactionModal = false;
	let showDeleteTransactionModal = false;
	let showTransferModal = false;
	let showFilters = false; // Add this state

	let selectedTransaction: TransactionDto | null = null;

	$: currentLocale = $locale || 'pt';

	function getTransactionDetails(transaction: TransactionDto): string {
		return `${transaction.description} (${formatCurrency(transaction.amount)}) ${$t('modals.with-category')} ${transaction.category.category_name} ${$t('common.at')} ${formatDate(transaction.date)}`;
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

	function getRowClass(tx: TransactionDto): string {
		if (tx.is_pending) {
			return $appliedTheme === 'dark' ? 'bg-base-300/80' : 'bg-base-200';
		}

		const type = tx.transaction_type.type_slug;
		if ($appliedTheme === 'dark') {
			if (type === 'debit') return 'bg-red-900 bg-opacity-40';
			if (type === 'credit') return 'bg-green-900 bg-opacity-100';
			return 'bg-base-300';
		} else {
			if (type === 'debit') return 'bg-red-100';
			if (type === 'credit') return 'bg-green-100';
			return '';
		}
	}

	// Add filter state
	let filters = emptyTransactionFilters();

	// Filter transactions
	$: filteredTransactions = transactions.filter((tx) =>
		matchesTransactionFilters(filters, {
			description: tx.description,
			categoryId: tx.category.id,
			typeSlug: tx.transaction_type.type_slug,
			date: tx.date,
			amount: tx.amount
		})
	);

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

	function handleApprovePendingTransaction(transaction: TransactionDto) {
		dispatch('approvePendingTransaction', { transaction });
	}

	function handleRejectPendingTransaction(transaction: TransactionDto) {
		dispatch('rejectPendingTransaction', { transaction });
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
	function handleNewTransaction(event: CustomEvent<TransactionChangeResponse>) {
		setDraftTransaction(null);
		closeCreateTransactionModal();
		dispatch('newTransaction', event.detail);
	}

	function handleUpdateTransaction(event: CustomEvent<TransactionChangeResponse>) {
		closeEditTransactionModal();
		dispatch('updateTransaction', event.detail);
	}

	function openTransferModal() {
		showTransferModal = true;
	}

	function closeTransferModal() {
		showTransferModal = false;
	}

	function handleNewTransfer(event: CustomEvent) {
		closeTransferModal();
		dispatch('newTransfer', event.detail);
	}

	function handleFilter(event: CustomEvent) {
		filters = event.detail;
	}

	function toggleFilters() {
		showFilters = !showFilters;
	}
</script>

{#if loading}
	<!-- Loading State -->
	<div class="py-12 text-center">
		<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
		<p class="text-base-content/70">{$t('common.loading')}</p>
	</div>
{:else if transactions && transactions.length > 0}
	<div class="my-2 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
		<!-- Buttons: above stats on mobile, right on md+ -->
		<div
			class="order-1 flex items-center justify-center gap-4 md:order-2 md:ml-auto md:justify-end"
		>
			<!-- Filter Button -->
			<button
				class="btn btn-outline gap-2"
				class:btn-active={showFilters}
				aria-label="Toggle Filters"
				on:click={toggleFilters}
			>
				<Filter size={20} />
				{#if Object.values(filters).some(Boolean)}
					<span class="badge badge-primary badge-sm">
						{Object.values(filters).filter(Boolean).length}
					</span>
				{/if}
			</button>

			<!-- Transfer Button -->
			<button
				class="btn btn-secondary shadow-lg"
				aria-label="Create Transfer"
				on:click={openTransferModal}
			>
				<ArrowRightLeft size={20} class="text-base-100" />
			</button>

			<!-- Transaction Button -->
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
			<TransactionsStats {transactions} />
		</div>
	</div>

	<!-- Filters Component (shown only when toggled) -->
	<TransactionFilters
		show={showFilters}
		filteredCount={filteredTransactions.length}
		totalCount={transactions.length}
		categories={[...new Map(transactions.map((tx) => [tx.category.id, tx.category])).values()]}
		transactionTypes={[
			...new Map(
				transactions.map((tx) => [tx.transaction_type.type_slug, tx.transaction_type])
			).values()
		]}
		on:filter={handleFilter}
	/>

	<div class="overflow-x-auto">
		{#if transactions.length === 0}
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
						{#each group.transactions as tx (tx.id)}
							<tr
								class={getRowClass(tx)}
								in:fly={{ y: 20, duration: 200 }}
								out:fade={{ duration: 150 }}
							>
								<td class="text-base-content">
									{formatDate(tx.date)}
								</td>
								<td class="text-base-content">
									<div class="flex flex-col items-center gap-0.5">
										{#if tx.category.parent_category_id}
											<span class="text-xs opacity-90">
												{tx.category.parent_category?.category_name || 'Parent'}
											</span>
										{/if}
										<span
											class="inline-flex items-center justify-center rounded px-3 py-1 {getTextColor(
												tx.category.color
											)}"
											style="background-color: {tx.category.color}; min-width: 4rem;"
										>
											{tx.category.category_name}
										</span>
									</div>
								</td>
								<td class="text-base-content">
									<div class="flex items-center justify-center gap-2">
										{#if tx.transfer_group_id}
											<span class="tooltip" data-tip={$t('transactions.transfer')}>
												<ArrowRightLeft size={14} class="text-info" />
											</span>
										{/if}
										{#if tx.is_pending}
											<span class="badge badge-warning badge-sm">{$t('transactions.pending')}</span>
										{/if}
										<span>{formatCurrency(tx.amount)}</span>
									</div>
								</td>
								<td class="text-base-content"> {tx.description || 'N/A'} </td>
								<td class="text-base-content">
									<div class="flex items-center justify-center gap-x-2">
										{#if tx.is_pending}
											<button
												class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-success backdrop-blur-sm"
												aria-label="Approve Transaction"
												on:click={() => handleApprovePendingTransaction(tx)}
											>
												<Check size={20} />
											</button>
											<button
												class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-error backdrop-blur-sm"
												aria-label="Reject Transaction"
												on:click={() => handleRejectPendingTransaction(tx)}
											>
												<X size={20} />
											</button>
										{:else}
											<button
												class="btn btn-circle btn-ghost btn-sm bg-base-100/80 backdrop-blur-sm"
												aria-label="Edit Transaction"
												on:click={() => handleEditTransaction(tx)}
											>
												<Pencil size={20} />
											</button>
											<button
												class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-error backdrop-blur-sm hover:bg-error/20"
												aria-label="Delete Transaction"
												on:click={() => handleDeleteTransaction(tx)}
											>
												<Trash size={20} />
											</button>
										{/if}
									</div>
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
			{$t('transactions.no-transactions-for')}
			<strong>{account?.account_name || 'Unknown Account'}</strong>.
		</p>

		<!-- Button to add a new transaction -->
		<button
			class="btn btn-primary mt-4 flex items-center gap-2 shadow-lg"
			on:click={openCreateTransactionModal}
			aria-label="Add New Transaction"
		>
			<CircleDollarSign size={20} class="h-5 w-5 text-base-100" />
			<span class="text-base-100">{$t('transactions.create-first')}</span>
		</button>
	</div>
{/if}

{#if showCreateTransactionModal}
	<TransactionModal
		{account}
		transaction={null}
		on:closeModal={closeCreateTransactionModal}
		on:newTransaction={handleNewTransaction}
	/>
{/if}

{#if showEditTransactionModal}
	<TransactionModal
		{account}
		transaction={selectedTransaction!}
		on:closeModal={closeEditTransactionModal}
		on:updateTransaction={handleUpdateTransaction}
	/>
{/if}

{#if showDeleteTransactionModal}
	<ConfirmAction
		title={`${$t('modals.delete-transaction')}`}
		message={`${$t('modals.delete-transaction-confirm')} ${getTransactionDetails(selectedTransaction!)}? ${$t('modals.cannot-be-undone')}`}
		type="danger"
		onConfirm={handleDeleteTransactionConfirm}
		onCancel={handleDeleteTransactionCancel}
	></ConfirmAction>
{/if}

{#if showTransferModal}
	<TransferModal {account} on:closeModal={closeTransferModal} on:newTransfer={handleNewTransfer} />
{/if}
