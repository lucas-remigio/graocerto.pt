<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { connected, socket, messages, sendMessage } from '$lib/ws'; // Import WebSocket utilities
	import type {
		Account,
		TransactionDto,
		TransactionGroup,
		CategoryDto,
		MonthYear,
		TransactionsTotals,
		TransactionStatistics,
		TransactionChangeResponse
	} from '$lib/types';
	import { dataService } from '$lib/services/dataService';
	import Accounts from '$components/Accounts.svelte';
	import TransactionsTable from '$components/TransactionsTable.svelte';
	import TransactionStatisticsComponent from '$components/TransactionStatistics.svelte';
	import MonthSelector from '$components/MonthSelector.svelte';
	import ViewToggle from '$components/ViewToggle.svelte';
	import { userEmail } from '$lib/stores/auth';
	import { t, locale } from '$lib/i18n';
	import { selectedView } from '$lib/stores/uiPreferences';
	import { TransactionTypeId } from '$lib/transaction_types_types';

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
	let categories: CategoryDto[] = $state([]);
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

	async function deleteAccount(account: Account) {
		try {
			await dataService.deleteAccount(account.token);
			await fetchAccounts(false);
		} catch (err) {
			console.error('Error deleting account:', err);
			error = $t('errors.failed-create-account');
		}
	}

	async function deleteTransaction(transaction: TransactionDto) {
		try {
			await dataService.deleteTransaction(transaction);
			await fetchAccounts(false);
		} catch (err) {
			console.error('Error deleting transaction:', err);
			error = $t('errors.failed-create-account');
		}
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
		$selectedView = 'transactions'; // Reset to transactions view when switching accounts
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

	function handleNewTransaction(event: CustomEvent<TransactionChangeResponse>) {
		// add the transaction to the correesponding group
		const transaction = event.detail.transaction;
		console.log('event', event.detail);

		selectedAccount!.balance = event.detail.account_balance;
		availableMonths = event.detail.months;

		transactions.push(transaction);
		// sort from newest to oldest
		transactions.sort((a, b) => {
			const dateA = new Date(a.date).getTime();
			const dateB = new Date(b.date).getTime();
			if (dateA !== dateB) {
				return dateB - dateA; // Newest date first
			}
			return b.id - a.id; // For same date, highest id first
		});
	}

	function handleUpdateTransaction(event: CustomEvent<TransactionChangeResponse>) {
		console.log('Update transaction event:', event.detail);
		const transaction = event.detail.transaction;
		selectedAccount!.balance = event.detail.account_balance;
		availableMonths = event.detail.months;

		const index = transactions.findIndex((t) => t.id === transaction.id);
		if (index !== -1) {
			transactions[index] = transaction;
			// Sort transactions after update to maintain order
			transactions.sort((a, b) => {
				const dateA = new Date(a.date).getTime();
				const dateB = new Date(b.date).getTime();
				if (dateA !== dateB) {
					return dateB - dateA; // Newest date first
				}
				return b.id - a.id; // For same date, highest id first
			});
		} else {
			console.warn('Transaction not found for update:', transaction.id);
		}
	}

	function handleDeleteAccount(account: Account) {
		// No need to clear all caches - the service will handle targeted cache clearing
		deleteAccount(account);

		wsUpdateScreen();
	}

	function handleNewAccount() {
		refreshAccountsAndNotify();
	}

	function handleUpdateAccount() {
		refreshAccountsAndNotify();
	}

	function refreshAccountsAndNotify() {
		dataService.clearCaches();
		fetchAccounts(false);
		wsUpdateScreen();
	}

	function handleDeleteTransaction(transaction: TransactionDto) {
		// No need to clear all caches - the service will handle targeted cache clearing
		deleteTransaction(transaction);

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
		<!-- Responsive Layout: Vertical on large screens, horizontal on small/medium -->
		<div class="flex flex-col lg:flex-row">
			<!-- Left Column: Accounts (full width on small/medium, fixed width on large) -->
			<div class="lg:w-80 lg:flex-shrink-0 lg:pr-6">
				<Accounts
					{accounts}
					{selectedAccount}
					isVertical={isLargeScreen}
					loading={accountsLoading}
					on:select={handleSelectAccount}
					on:updatedAccount={handleUpdateAccount}
					on:deleteAccount={({ detail: { account } }) => handleDeleteAccount(account)}
					on:newAccount={handleNewAccount}
				/>
			</div>

			<!-- Vertical Divider - only visible on large screens -->
			<div class="lg:bg-base-300 hidden lg:block lg:w-px"></div>

			<!-- Right Column: Transactions (full width on small/medium, remaining space on large) -->
			{#if accounts.length > 0}
				<div class="flex-1 lg:flex lg:max-h-screen lg:flex-col lg:overflow-hidden lg:pl-6">
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
						<ViewToggle bind:currentView={$selectedView} />
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
							/>
						{:else}
							<TransactionStatisticsComponent
								{selectedMonth}
								{selectedYear}
								{statistics}
								loading={statisticsLoading}
								error={statisticsError}
							/>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
