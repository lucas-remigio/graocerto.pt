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
		TransactionStatistics
	} from '$lib/types';
	import { dataService } from '$lib/services/dataService';
	import Accounts from '$components/Accounts.svelte';
	import TransactionsTable from '$components/TransactionsTable.svelte';
	import TransactionStatisticsComponent from '$components/TransactionStatistics.svelte';
	import MonthSelector from '$components/MonthSelector.svelte';
	import ViewToggle from '$components/ViewToggle.svelte';
	import { userEmail } from '$lib/stores/auth';
	import { t, locale } from '$lib/i18n';

	// Track WebSocket connection status
	let hasJoinedRoom = false;
	let wsConnected = false;

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
	let accounts: Account[] = [];
	let accountsLoading = false;
	let transactionGroups: TransactionGroup[] = [];
	let transactionsLoading = false;
	let transactionsTotals: TransactionsTotals = {
		debit: 0,
		credit: 0,
		difference: 0
	};
	let statistics: TransactionStatistics | null = null;
	let statisticsLoading = false;
	let statisticsError: string = '';
	let categories: CategoryDto[] = [];
	let error: string = '';

	let selectedAccount: Account;

	// Month selector state
	let availableMonths: MonthYear[] = [];
	const currentMonth = new Date().getMonth() + 1; // 1-12 (1 = January)
	const currentYear = new Date().getFullYear();
	let selectedMonth: number | null = currentMonth;
	let selectedYear: number | null = currentYear;

	// Track screen size for responsive layout
	let isLargeScreen: boolean = false;

	// View toggle state
	let currentView: 'transactions' | 'statistics' = 'transactions';

	$: currentLocale = $locale || 'en-US'; // Default to 'en-US' if locale is not set

	// Update screen size tracking
	function updateScreenSize() {
		isLargeScreen = window.innerWidth >= 1024; // lg breakpoint in Tailwind
	}

	function getSelectedAccount() {
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

	function getSelectedView() {
		const storedView = localStorage.getItem('selectedView');
		if (storedView === 'statistics' || storedView === 'transactions') {
			currentView = storedView;
		}
		// If no stored view or invalid value, keep default 'transactions'
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
				getSelectedView();
				await fetchAccountTransactions(
					selectedAccount.token,
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
			if (currentView === 'statistics') {
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
			transactionGroups = result.transactionGroups;
			transactionsTotals = result.totals;
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
		currentView = 'transactions'; // Reset to transactions view when switching accounts
		localStorage.setItem('selectedView', currentView);
		fetchAccountTransactions(selectedAccount.token, selectedMonth, selectedYear, true);
	}

	function handleMonthSelect(month: number | null, year: number | null) {
		selectedMonth = month;
		selectedYear = year;

		// Fetch data based on current view
		if (currentView === 'statistics') {
			// Only fetch statistics when in statistics view
			fetchStatistics(selectedAccount.token, month, year, true);
		} else {
			// Only fetch transactions when in transactions view
			fetchTransactions(selectedAccount.token, month, year, true);
		}
	}

	$: selectedFormatedDate = (() => {
		const year = selectedYear ?? currentYear;
		const month = selectedMonth ?? currentMonth;
		const date = new Date(year, month - 1); // month is 0-indexed in JavaScript
		return date.toLocaleString(currentLocale, { month: 'long', year: 'numeric' });
	})();

	function handleViewToggle(event: CustomEvent<{ view: 'transactions' | 'statistics' }>) {
		currentView = event.detail.view;
		localStorage.setItem('selectedView', currentView);

		// Fetch appropriate data for current view if not cached
		if (selectedAccount) {
			if (currentView === 'statistics') {
				// Switching to statistics view - fetch statistics
				fetchStatistics(selectedAccount.token, selectedMonth, selectedYear, true);
			} else {
				// Switching to transactions view - fetch transactions
				fetchTransactions(selectedAccount.token, selectedMonth, selectedYear, true);
			}
		}
	}

	function handleNewTransaction() {
		dataService.clearCaches(); // Clear caches since data changed
		fetchAccounts(false);

		wsUpdateScreen();
	}

	function handleNewAccount() {
		dataService.clearCaches(); // Clear caches since data changed
		fetchAccounts(false);

		wsUpdateScreen();
	}

	function handleUpdateTransaction() {
		dataService.clearCaches(); // Clear caches since data changed
		fetchAccounts(false);

		wsUpdateScreen();
	}

	function handleUpdateAccount() {
		dataService.clearCaches(); // Clear caches since data changed
		fetchAccounts(false);

		wsUpdateScreen();
	}

	function handleDeleteAccount(account: Account) {
		// No need to clear all caches - the service will handle targeted cache clearing
		deleteAccount(account);

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

<div class="container mx-auto p-4">
	{#if error}
		<div class="alert alert-error">
			<p>{error}</p>
		</div>
	{:else}
		<!-- Responsive Layout: Vertical on large screens, horizontal on small/medium -->
		<div class="flex flex-col lg:h-[calc(100vh-100px)] lg:flex-row">
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
				<div class="flex-1 lg:flex lg:min-h-0 lg:flex-col lg:overflow-hidden lg:pl-6">
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
						<ViewToggle {currentView} on:viewChange={handleViewToggle} />
					</div>

					<div class="divider my-0"></div>

					<!-- Content Container with scroll -->
					<div class="lg:min-h-0 lg:flex-1 lg:overflow-auto">
						{#if currentView === 'transactions'}
							<TransactionsTable
								transactionsGroups={transactionGroups}
								{transactionsTotals}
								account={selectedAccount}
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
