<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { connected, messages, sendMessage } from '$lib/ws'; // Import WebSocket utilities
	import type {
		Account,
		TransactionDto,
		MonthYear,
		TransactionStatistics,
		TransactionChangeResponse,
		TransferResponse
	} from '$lib/types';
	import { dataService } from '$lib/services/dataService';
	import AccountsSplitLayout from '$components/AccountsSplitLayout.svelte';
	import TransactionsTable from '$components/TransactionsTable.svelte';
	import TransactionStatisticsComponent from '$components/TransactionStatistics.svelte';
	import MonthSelector from '$components/MonthSelector.svelte';
	import ViewToggle from '$components/ViewToggle.svelte';
	import { userEmail } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import { selectedView, updateSelectedView } from '$lib/stores/uiPreferences';

	// WebSocket state
	let hasJoinedRoom = $state(false);
	let wsConnected = $state(false);

	// Subscribe to connection status and room logic
	const unsubConnected = connected.subscribe((value) => {
		wsConnected = value;

		// Join room when connected and not already joined
		if (value && $userEmail && !hasJoinedRoom) {
			// Add a slight delay to ensure WebSocket is fully established
			setTimeout(() => {
				sendMessage({
					type: 'join_room',
					email: $userEmail
				});
				hasJoinedRoom = true;
			}, 500);
		}
	});

	// Subscribe to messages
	const unsubMessages = messages.subscribe((msgs) => {
		if (msgs.length > 0) {
			// Process only the latest message to avoid duplicates
			const latestMsg = msgs[msgs.length - 1];

			// Check if it's an update notification
			if (latestMsg.type === 'account_update') {
				fetchAccounts(false); // Refresh data
			}
		}
	});

	// Local component state
	let accounts: Account[] = $state([]);
	let accountsLoading = $state(false);
	let transactions: TransactionDto[] = $state([]);
	let transactionsLoading = $state(false);
	let statistics: TransactionStatistics | null = $state(null);
	let statisticsLoading = $state(false);
	let statisticsError: string = $state('');
	let error: string = $state('');

	let selectedAccount: Account | null = $state(null);

	// Month selector state
	let availableMonths: MonthYear[] = $state([]);
	const currentMonth = new Date().getMonth() + 1;
	const currentYear = new Date().getFullYear();
	let selectedMonth: number | null = $state(currentMonth);
	let selectedYear: number | null = $state(currentYear);

	// Track screen size for responsive layout
	let isLargeScreen: boolean = $state(false);
	let initialDataLoaded = $state(false);
	// Update screen size tracking
	function updateScreenSize() {
		isLargeScreen = window.innerWidth >= 1024; // lg breakpoint in Tailwind
	}

	function getSelectedAccount() {
		// if there is already a selected account, use it
		if (selectedAccount) {
			return;
		}

		if (accounts.length === 0) {
			return;
		}

		const storedAccountToken = localStorage.getItem('selectedAccount');
		if (!storedAccountToken) {
			selectedAccount = accounts[0];
			return;
		}

		const foundAccount = accounts.find((account) => account.token === storedAccountToken);
		if (!foundAccount) {
			selectedAccount = accounts[0];
			return;
		}

		selectedAccount = foundAccount;
	}

	// Function to fetch accounts and then fetch transactions for the first account
	async function fetchAccounts(showLoading: boolean) {
		accountsLoading = showLoading;
		try {
			accounts = await dataService.fetchAccounts();

			// If we have at least one account, fetch its transactions
			if (accounts && accounts.length > 0) {
				getSelectedAccount();
				await fetchAccountTransactions(
					selectedAccount!.token,
					selectedMonth,
					selectedYear,
					showLoading
				);
			}
		} catch (err) {
			console.error('Error in fetchAccounts:', err);
			error = $t('errors.failed-load-accounts');
		} finally {
			accountsLoading = false;
		}
	}

	// Function to fetch transactions for a given account token
	async function fetchAccountTransactions(
		accountToken: string,
		month: number | null,
		year: number | null,
		showLoading: boolean
	) {
		try {
			const promises = [fetchAvailableMonths(accountToken)];

			// If current view is statistics, also fetch statistics
			if ($selectedView === 'statistics') {
				promises.push(fetchStatistics(accountToken, month, year, showLoading));
			} else {
				promises.push(fetchTransactions(accountToken, month, year, showLoading));
			}

			await Promise.all(promises);
		} catch (err) {
			console.error('Error in fetchAccountTransactions:', err);
			error = $t('errors.failed-load-transactions');
		}
	}

	async function fetchTransactions(
		accountToken: string,
		month: number | null,
		year: number | null,
		showLoading: boolean
	) {
		transactionsLoading = showLoading;
		try {
			const result = await dataService.fetchTransactions(accountToken, month, year);
			transactions = result.transactions;
		} catch (err) {
			console.error('Error fetching transactions:', err);
			error = $t('errors.failed-load-transactions');
		} finally {
			transactionsLoading = false;
		}
	}

	async function fetchAvailableMonths(accountToken: string) {
		try {
			availableMonths = await dataService.fetchAvailableMonths(accountToken);

			// check if there is this current month in the available months. if not, add it
			if (
				!availableMonths.some(
					(monthData) => monthData.month === currentMonth && monthData.year === currentYear
				)
			) {
				addCurrentMonth();
			}
		} catch (err) {
			console.error('Error in fetchAvailableMonths:', err);
			error = $t('errors.failed-load-months');
		}
	}

	// Function to fetch statistics for a given account token and month/year
	async function fetchStatistics(
		accountToken: string,
		month: number | null,
		year: number | null,
		showLoading: boolean
	) {
		statisticsLoading = showLoading;
		statisticsError = '';

		try {
			statistics = await dataService.fetchStatistics(accountToken, month, year);
		} catch (err) {
			console.error('Error fetching statistics:', err);
			statisticsError = $t('errors.failed-load-transactions');
		} finally {
			statisticsLoading = false;
		}
	}

	function addCurrentMonth() {
		const currentMonthYear: MonthYear = {
			month: currentMonth,
			year: currentYear,
			count: 0
		};

		availableMonths.unshift(currentMonthYear);
	}

	function handleSelectAccount(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		localStorage.setItem('selectedAccount', selectedAccount.token);
		selectedMonth = currentMonth;
		selectedYear = currentYear;
		// by triggering the selected view, we ensure that the transactions are fetched
		// so no need to manually call the fetch account transaction
		updateSelectedView('transactions'); // Reset to transactions view when switching accounts
	}

	function handleMonthSelect(month: number | null, year: number | null) {
		selectedMonth = month;
		selectedYear = year;

		// by changing the selected month, we ensure that the transactions are fetched
		// by the reactive statement
	}

	$effect(() => {
		if (selectedAccount && $selectedView && initialDataLoaded) {
			fetchAccountTransactions(selectedAccount.token, selectedMonth, selectedYear, true);
		}
	});

	/* =========================================================
	 * Transaction Logic
	 * ========================================================
	 */

	function isTransactionInCurrentMonthAndYear(transaction: TransactionDto): boolean {
		const transactionDate = new Date(transaction.date);
		const transactionMonth = transactionDate.getMonth() + 1; // getMonth() is zero-based
		const transactionYear = transactionDate.getFullYear();

		if (selectedMonth === null || selectedYear === null) {
			// If no month/year is selected, show all transactions
			return selectedAccount?.token === transaction.account_token;
		}

		// this makes is so that when we are in august and add a transaction to september,
		// we do not add it to the current months list
		return (
			transactionMonth === selectedMonth &&
			transactionYear === selectedYear &&
			selectedAccount?.token === transaction.account_token
		);
	}

	function updateAccountAndMonths(response: TransactionChangeResponse) {
		selectedAccount!.balance = response.account_balance;
		if (response.account_pending_balance !== undefined) {
			selectedAccount!.pending_balance = response.account_pending_balance;
		}
		availableMonths = response.months;
	}

	function sortTransactions() {
		transactions.sort((a, b) => {
			const dateA = new Date(a.date).getTime();
			const dateB = new Date(b.date).getTime();
			if (dateA !== dateB) return dateB - dateA;
			return b.id - a.id;
		});
	}

	function upsertTransaction(transaction: TransactionDto) {
		// first we need to check if the transaction is in the current month and year
		if (!isTransactionInCurrentMonthAndYear(transaction)) {
			return;
		}

		const idx = transactions.findIndex((t) => t.id === transaction.id);
		if (idx !== -1) {
			transactions[idx] = transaction;
		} else {
			transactions.push(transaction);
		}
		sortTransactions();
	}

	function handleNewTransaction(event: CustomEvent<TransactionChangeResponse>) {
		const { transaction } = event.detail;
		updateAccountAndMonths(event.detail);
		upsertTransaction(transaction);
		refreshCachesAndNotify();
	}

	function handleUpdateTransaction(event: CustomEvent<TransactionChangeResponse>) {
		const { transaction } = event.detail;
		updateAccountAndMonths(event.detail);
		upsertTransaction(transaction);
		refreshCachesAndNotify();
	}

	async function handleDeleteTransaction(transaction: TransactionDto) {
		await deleteTransaction(transaction);
		wsUpdateScreen();
	}

	async function handleApprovePendingTransaction(transaction: TransactionDto) {
		try {
			const response = await dataService.approvePendingTransaction(transaction);
			updateAccountAndMonths(response);
			upsertTransaction(response.transaction);
			refreshCachesAndNotify();
		} catch (err) {
			console.error('Error approving pending transaction:', err);
			error = $t('errors.failed-update-transaction');
		}
	}

	async function deleteTransaction(transaction: TransactionDto) {
		try {
			const response = await dataService.deleteTransaction(transaction);

			// Always remove the deleted transaction from current view
			transactions = transactions.filter((t) => t.id !== transaction.id);

			// Update current account
			updateAccountAndMonths(response);

			// Handle transfer deletion
			if (response.is_transfer) {
				handleTransferDeletion(transaction, response);
			}
		} catch (err) {
			console.error('Error deleting transaction:', err);
			error = $t('errors.failed-delete-transaction');
		}
	}

	function handleTransferDeletion(
		deletedTransaction: TransactionDto,
		response: TransactionChangeResponse
	) {
		if (!response.paired_account_token) return;

		// Find and update the paired account
		const pairedAccount = accounts.find((acc) => acc.token === response.paired_account_token);
		if (pairedAccount && response.paired_account_balance !== undefined) {
			pairedAccount.balance = response.paired_account_balance;
		}
		if (pairedAccount && response.paired_account_pending_balance !== undefined) {
			pairedAccount.pending_balance = response.paired_account_pending_balance;
		}

		// If we're currently viewing the paired account, update its view
		if (isViewingPairedAccount(response.paired_account_token)) {
			updatePairedAccountView(deletedTransaction, response);
		}
	}

	function isViewingPairedAccount(pairedAccountToken: string): boolean {
		return selectedAccount?.token === pairedAccountToken;
	}

	function updatePairedAccountView(
		deletedTransaction: TransactionDto,
		response: TransactionChangeResponse
	) {
		// Update balance
		if (response.paired_account_balance !== undefined) {
			selectedAccount!.balance = response.paired_account_balance;
		}
		if (response.paired_account_pending_balance !== undefined) {
			selectedAccount!.pending_balance = response.paired_account_pending_balance;
		}

		// Update available months
		if (response.paired_account_months) {
			availableMonths = response.paired_account_months;
		}

		// Remove the paired transaction using transfer_group_id
		if (deletedTransaction.transfer_group_id) {
			transactions = transactions.filter(
				(t) => t.transfer_group_id !== deletedTransaction.transfer_group_id
			);
		}
	}
	/* ========================================================
	 * Transfer Logic
	 * ========================================================
	 */

	function handleNewTransfer(event: CustomEvent<TransferResponse>) {
		const response = event.detail;

		// Update source account balance if it's the selected account
		if (selectedAccount?.token === response.debit_transaction.account_token) {
			selectedAccount.balance = response.source_account_balance;
			availableMonths = response.source_account_months;
			upsertTransaction(response.debit_transaction);
		}

		// Update destination account balance
		const destAccount = accounts.find(
			(acc) => acc.token === response.credit_transaction.account_token
		);
		if (destAccount) {
			destAccount.balance = response.destination_account_balance;
		}

		// If destination is selected account, update months and add credit transaction
		if (selectedAccount?.token === response.credit_transaction.account_token) {
			selectedAccount.balance = response.destination_account_balance;
			availableMonths = response.destination_account_months;
			upsertTransaction(response.credit_transaction);
		}

		refreshCachesAndNotify();
	}

	/* ========================================================
	 * Account Logic
	 * ========================================================
	 */

	function upsertAccount(account: Account) {
		const idx = accounts.findIndex((acc) => acc.token === account.token);
		if (idx !== -1) {
			accounts[idx] = account; // update existing account
		} else {
			accounts.push(account); // add new account
		}
		accounts.sort((a, b) => a.order_index - b.order_index);
	}

	function handleNewAccount(account: Account) {
		upsertAccount(account);
		selectedAccount = null; // Clear selected account
		getSelectedAccount(); // Update selected account if needed
		wsUpdateScreen();
	}

	function handleUpdateAccount(account: Account) {
		upsertAccount(account);
		wsUpdateScreen();
	}

	function handleDeleteAccount(account: Account) {
		// No need to clear all caches - the service will handle targeted cache clearing
		deleteAccount(account);

		wsUpdateScreen();
	}

	async function deleteAccount(account: Account) {
		try {
			await dataService.deleteAccount(account.token);
			accounts = accounts.filter((acc) => acc.token !== account.token);
			selectedAccount = null; // Clear selected account
			getSelectedAccount(); // Update selected account if needed
		} catch (err) {
			console.error('Error deleting account:', err);
			error = $t('errors.failed-create-account');
		}
	}

	/* ========================================================
	 * UI Logic
	 * ========================================================
	 */

	function refreshCachesAndNotify() {
		dataService.clearTransactionCaches();
		wsUpdateScreen();
	}

	function wsUpdateScreen() {
		// this function is called on every deletion, edition or creation of both an account and a transaction
		// Notify other users of the change
		if (wsConnected) {
			sendMessage({
				type: 'account_update',
				action: 'update',
				email: $userEmail
			});
		}
	}

	onMount(async () => {
		await fetchAccounts(true);
		initialDataLoaded = true;

		// Set up screen size tracking
		updateScreenSize();
		window.addEventListener('resize', updateScreenSize);
	});

	// Clean up subscription when component is destroyed
	onDestroy(() => {
		unsubConnected();
		unsubMessages();
		window.removeEventListener('resize', updateScreenSize);
	});
</script>

<div class="container mx-auto flex flex-col p-4">
	{#if error}
		<div class="alert alert-error">
			<p>{error}</p>
		</div>
	{:else}
		<AccountsSplitLayout
			{accounts}
			{selectedAccount}
			{isLargeScreen}
			accountsLoading={accountsLoading}
			showRightPanel={accounts.length > 0}
			on:select={handleSelectAccount}
			on:updatedAccount={({ detail: { account } }) => handleUpdateAccount(account)}
			on:deleteAccount={({ detail: { account } }) => handleDeleteAccount(account)}
			on:newAccount={({ detail: { account } }) => handleNewAccount(account)}
		>
			<!-- Horizontal Divider - only visible on small/medium screens -->
			<div class="divider lg:hidden"></div>

			<!-- Month Selector Component -->
			<div class="lg:flex-shrink-0">
				<MonthSelector
					{availableMonths}
					{selectedMonth}
					{selectedYear}
					on:monthSelect={({ detail }) => handleMonthSelect(detail.month, detail.year)}
				/>
			</div>

			<!-- View Toggle Component -->
			<div class="lg:flex-shrink-0">
				<ViewToggle />
			</div>

			<div class="divider my-0"></div>

			<!-- Content Container with scroll -->
			<div class="min-h-0 flex-1 overflow-y-auto">
				{#if $selectedView === 'transactions'}
					<TransactionsTable
						{transactions}
						account={selectedAccount!}
						isAll={selectedMonth === null && selectedYear === null}
						loading={transactionsLoading}
						on:newTransaction={handleNewTransaction}
						on:updateTransaction={handleUpdateTransaction}
						on:deleteTransaction={({ detail: { transaction } }) =>
							handleDeleteTransaction(transaction)}
						on:newTransfer={handleNewTransfer}
						on:approvePendingTransaction={({ detail: { transaction } }) =>
							handleApprovePendingTransaction(transaction)}
						on:rejectPendingTransaction={({ detail: { transaction } }) =>
							handleDeleteTransaction(transaction)}
					/>
				{:else}
					<TransactionStatisticsComponent
						{selectedMonth}
						{selectedYear}
						{statistics}
						account={selectedAccount!}
						loading={statisticsLoading}
						error={statisticsError}
					/>
				{/if}
			</div>
		</AccountsSplitLayout>
	{/if}
</div>
