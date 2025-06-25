<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { connected, socket, messages, sendMessage } from '$lib/ws'; // Import WebSocket utilities
	import api_axios from '$lib/axios';
	import type {
		Account,
		AccountsResponse,
		TransactionDto,
		TransactionsResponseDto,
		CategoryDto,
		MonthYear,
		TransactionsTotals,
		TransactionStatistics
	} from '$lib/types';
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
				fetchAccounts(); // Refresh data
			}
		}
	});

	// Local component state
	let accounts: Account[] = [];
	let transactions: TransactionDto[] = []; // Store all transactions
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

	// Caches to avoid unnecessary API calls
	let statisticsCache = new Map<string, TransactionStatistics>();
	let transactionsCache = new Map<
		string,
		{ transactions: TransactionDto[]; totals: TransactionsTotals }
	>();

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

	async function deleteAccount(account: Account) {
		try {
			const response = await api_axios.delete(`accounts/${account.token}`);

			if (response.status !== 200) {
				console.error('Non-200 response status:', response.status);
				error = `Error: ${response.status}`;
				return;
			}

			fetchAccounts();
		} catch (err) {
			console.error('Error in handleSubmit:', err);
			error = $t('errors.failed-create-account');
		}
	}

	async function deleteTransaction(transaction: TransactionDto) {
		try {
			const response = await api_axios.delete(`transactions/${transaction.id}`);

			if (response.status !== 200) {
				console.error('Non-200 response status:', response.status);
				error = `Error: ${response.status}`;
				return;
			}

			fetchAccounts();
		} catch (err) {
			console.error('Error in handleSubmit:', err);
			error = $t('errors.failed-create-account');
		}
	}

	// Function to fetch accounts and then fetch transactions for the first account
	async function fetchAccounts() {
		try {
			const res = await api_axios('accounts');

			if (res.status !== 200) {
				console.error('Non-200 response status:', res.status);
				error = `Error: ${res.status}`;
				return;
			}
			const data: AccountsResponse = res.data;
			accounts = data.accounts;

			// If we have at least one account, fetch its transactions
			if (accounts && accounts.length > 0) {
				getSelectedAccount();
				await fetchAccountTransactions(selectedAccount.token, selectedMonth, selectedYear);
			}
		} catch (err) {
			console.error('Error in fetchAccounts:', err);
			error = $t('errors.failed-load-accounts');
		}
	}

	// Function to fetch transactions for a given account token
	async function fetchAccountTransactions(
		accountToken: string,
		month: number | null,
		year: number | null
	) {
		try {
			await Promise.all([
				fetchTransactions(accountToken, month, year),
				fetchAvailableMonths(accountToken)
			]);
		} catch (err) {
			console.error('Error in fetchAccountTransactions:', err);
			error = $t('errors.failed-load-transactions');
		}
	}

	// Function to generate cache key for transactions and statistics
	function getCacheKey(accountToken: string, month: number | null, year: number | null): string {
		return `${accountToken}-${month ?? 'null'}-${year ?? 'null'}`;
	}

	async function fetchTransactions(
		accountToken: string,
		month: number | null,
		year: number | null
	) {
		const cacheKey = getCacheKey(accountToken, month, year);

		// Check cache first
		if (transactionsCache.has(cacheKey)) {
			const cached = transactionsCache.get(cacheKey)!;
			transactions = cached.transactions;
			transactionsTotals = cached.totals;
			return;
		}

		try {
			const res = await api_axios('transactions/dto/' + accountToken, {
				params: {
					month,
					year
				}
			});

			if (res.status !== 200) {
				console.error('Non-200 response status for transactions:', res.status);
				error = `Error: ${res.status}`;
				return;
			}

			const data: TransactionsResponseDto = res.data;
			transactions = data.transactions;
			transactionsTotals = data.totals;

			// Cache the result
			transactionsCache.set(cacheKey, {
				transactions: data.transactions,
				totals: data.totals
			});
		} catch (err) {
			console.error('Error in fetchAccountTransactions:', err);
			error = $t('errors.failed-load-transactions');
		}
	}

	async function fetchAvailableMonths(accountToken: string) {
		try {
			const res = await api_axios('transactions/months/' + accountToken);

			if (res.status !== 200) {
				console.error('Non-200 response status for months:', res.status);
				error = `Error: ${res.status}`;
				return;
			}

			availableMonths = res.data.months as MonthYear[];
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
	async function fetchStatistics(accountToken: string, month: number | null, year: number | null) {
		const cacheKey = getCacheKey(accountToken, month, year);

		// Check cache first
		if (statisticsCache.has(cacheKey)) {
			statistics = statisticsCache.get(cacheKey)!;
			return;
		}

		statisticsLoading = true;
		statisticsError = '';

		try {
			const params: any = {};
			if (month !== null) params.month = month;
			if (year !== null) params.year = year;

			const response = await api_axios.get(`transactions/statistics/${accountToken}`, { params });

			if (response.status === 200) {
				statistics = response.data;
				// Cache the result (only if not null)
				if (statistics) {
					statisticsCache.set(cacheKey, statistics);
				}
			} else {
				statisticsError = `Failed to load statistics: ${response.status}`;
			}
		} catch (err) {
			console.error('Error fetching statistics:', err);
			statisticsError = $t('errors.failed-load-transactions');
		} finally {
			statisticsLoading = false;
		}
	}

	// Function to clear caches when data changes
	function clearCaches() {
		statisticsCache.clear();
		transactionsCache.clear();
		statistics = null;
	}

	function addCurrentMonth() {
		const currentMonthYear: MonthYear = {
			month: currentMonth,
			year: currentYear,
			count: 0
		};

		availableMonths.unshift(currentMonthYear);
	}

	// Function to fetch categories
	async function fetchCategories() {
		try {
			const res = await api_axios('categories/dto');
			if (res.status === 200) {
				categories = res.data.categories;
			}
		} catch (err) {
			console.error('Error fetching categories:', err);
			error = $t('errors.failed-load-categories');
		}
	}

	function handleSelectAccount(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		localStorage.setItem('selectedAccount', selectedAccount.token);
		selectedMonth = currentMonth;
		selectedYear = currentYear;
		clearCaches(); // Clear caches when switching accounts
		currentView = 'transactions'; // Reset to transactions view when switching accounts
		fetchAccountTransactions(selectedAccount.token, selectedMonth, selectedYear);
	}

	function handleMonthSelect(month: number | null, year: number | null) {
		selectedMonth = month;
		selectedYear = year;

		// Fetch data based on current view
		if (currentView === 'statistics') {
			// Only fetch statistics when in statistics view
			fetchStatistics(selectedAccount.token, month, year);
		} else {
			// Only fetch transactions when in transactions view
			fetchTransactions(selectedAccount.token, month, year);
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

		// Fetch appropriate data for current view if not cached
		if (selectedAccount) {
			if (currentView === 'statistics') {
				// Switching to statistics view - fetch statistics
				fetchStatistics(selectedAccount.token, selectedMonth, selectedYear);
			} else {
				// Switching to transactions view - fetch transactions
				fetchTransactions(selectedAccount.token, selectedMonth, selectedYear);
			}
		}
	}

	function handleNewTransaction() {
		clearCaches(); // Clear caches since data changed
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleNewAccount() {
		clearCaches(); // Clear caches since data changed
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleUpdateTransaction() {
		clearCaches(); // Clear caches since data changed
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleUpdateAccount() {
		clearCaches(); // Clear caches since data changed
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleDeleteAccount(account: Account) {
		clearCaches(); // Clear caches since data changed
		deleteAccount(account);

		wsUpdateScreen();
	}

	function handleDeleteTransaction(transaction: TransactionDto) {
		clearCaches(); // Clear caches since data changed
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
		await Promise.all([fetchAccounts(), fetchCategories()]);

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
		<div class="flex flex-col lg:h-[calc(100vh-120px)] lg:flex-row">
			<!-- Left Column: Accounts (full width on small/medium, fixed width on large) -->
			<div class="lg:w-80 lg:flex-shrink-0 lg:pr-6">
				<Accounts
					{accounts}
					{selectedAccount}
					isVertical={isLargeScreen}
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

					<!-- Content Container with scroll -->
					<div class="lg:min-h-0 lg:flex-1 lg:overflow-auto">
						{#if currentView === 'transactions'}
							<TransactionsTable
								{transactions}
								{categories}
								{transactionsTotals}
								account={selectedAccount}
								formatedDate={selectedFormatedDate}
								isAll={selectedMonth === null && selectedYear === null}
								on:newTransaction={handleNewTransaction}
								on:updatedTransaction={handleUpdateTransaction}
								on:deleteTransaction={({ detail: { transaction } }) =>
									handleDeleteTransaction(transaction)}
							/>
						{:else}
							<TransactionStatisticsComponent
								account={selectedAccount}
								{selectedMonth}
								{selectedYear}
								formatedDate={selectedFormatedDate}
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
