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
		TransactionsTotals
	} from '$lib/types';
	import Accounts from '$components/Accounts.svelte';
	import TransactionsTable from '$components/TransactionsTable.svelte';
	import MonthSelector from '$components/MonthSelector.svelte';
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

	async function fetchTransactions(
		accountToken: string,
		month: number | null,
		year: number | null
	) {
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
		fetchAccountTransactions(selectedAccount.token, selectedMonth, selectedYear);
	}

	function handleMonthSelect(month: number | null, year: number | null) {
		selectedMonth = month;
		selectedYear = year;

		fetchTransactions(selectedAccount.token, month, year);
	}

	function handleNewTransaction() {
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleNewAccount() {
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleUpdateTransaction() {
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleUpdateAccount() {
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleDeleteAccount(account: Account) {
		deleteAccount(account);

		wsUpdateScreen();
	}

	function handleDeleteTransaction(transaction: TransactionDto) {
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
		<div class="flex flex-col lg:h-[calc(100vh-120px)] lg:flex-row lg:gap-6">
			<!-- Left Column: Accounts (full width on small/medium, fixed width on large) -->
			<div class="lg:w-80 lg:flex-shrink-0">
				<Accounts
					{accounts}
					{selectedAccount}
					vertical={isLargeScreen}
					on:select={handleSelectAccount}
					on:updatedAccount={handleUpdateAccount}
					on:deleteAccount={({ detail: { account } }) => handleDeleteAccount(account)}
					on:newAccount={handleNewAccount}
				/>
			</div>

			<!-- Right Column: Transactions (full width on small/medium, remaining space on large) -->
			{#if accounts.length > 0}
				<div class="flex-1 lg:flex lg:min-h-0 lg:flex-col lg:overflow-hidden">
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

					<!-- Transactions Table with scroll container -->
					<div class="lg:min-h-0 lg:flex-1 lg:overflow-auto">
						<TransactionsTable
							{transactions}
							{categories}
							{transactionsTotals}
							account={selectedAccount}
							isAll={selectedMonth === null && selectedYear === null}
							on:newTransaction={handleNewTransaction}
							on:updatedTransaction={handleUpdateTransaction}
							on:deleteTransaction={({ detail: { transaction } }) =>
								handleDeleteTransaction(transaction)}
						/>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
