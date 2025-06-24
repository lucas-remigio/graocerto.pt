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
		MonthYear
	} from '$lib/types';
	import { Plus, Wallet, Calendar } from 'lucide-svelte';
	import Accounts from '$components/Accounts.svelte';
	import TransactionsTable from '$components/TransactionsTable.svelte';
	import CreateAccount from '$components/CreateAccount.svelte';
	import { userEmail } from '$lib/stores/auth';
	import { t } from '$lib/i18n';
	import { format } from 'svelte-i18n';

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
	let categories: CategoryDto[] = [];
	let error: string = '';
	let showCreateAccountModal = false;

	let selectedAccount: Account;

	// Month selector state
	let availableMonths: MonthYear[] = [];
	let selectedMonth = new Date().getMonth() + 1; // 1-12 (1 = January)
	let selectedYear = new Date().getFullYear();

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
			error = 'Failed to create account';
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
			error = 'Failed to create account';
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
			error = 'Failed to load accounts';
		}
	}

	// Function to fetch transactions for a given account token
	async function fetchAccountTransactions(accountToken: string, month?: number, year?: number) {
		try {
			await Promise.all([
				fetchTransactions(accountToken, month, year),
				fetchAvailableMonths(accountToken)
			]);
		} catch (err) {
			console.error('Error in fetchAccountTransactions:', err);
			error = 'Failed to load transactions';
		}
	}

	async function fetchTransactions(accountToken: string, month?: number, year?: number) {
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
		} catch (err) {
			console.error('Error in fetchAccountTransactions:', err);
			error = 'Failed to load transactions';
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

			console.log('Available months response:', res.data);
			availableMonths = res.data.months as MonthYear[];
		} catch (err) {
			console.error('Error in fetchAvailableMonths:', err);
			error = 'Failed to load available months';
		}
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
			error = 'Failed to load categories';
		}
	}

	function handleSelectAccount(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		localStorage.setItem('selectedAccount', selectedAccount.token);
		fetchAccountTransactions(selectedAccount.token);
	}

	function createAccount() {
		showCreateAccountModal = true;
	}

	function closeAccountModal() {
		showCreateAccountModal = false;
	}

	function handleMonthSelect(month: number, year: number) {
		selectedMonth = month;
		selectedYear = year;

		fetchAccountTransactions(selectedAccount.token, month, year);
	}

	function handleNewTransaction() {
		fetchAccounts();

		wsUpdateScreen();
	}

	function handleNewAccount() {
		closeAccountModal();
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

	function formatDate(month: number, year: number): string {
		const date = new Date(year, month - 1); // month is 0-indexed in JavaScript
		return date.toLocaleString('default', { month: 'long', year: 'numeric' });
	}

	function isCurrentMonth(monthData: MonthYear): boolean {
		const currentDate = new Date();
		return (
			monthData.month === currentDate.getMonth() + 1 && monthData.year === currentDate.getFullYear()
		);
	}

	onMount(async () => {
		await Promise.all([fetchAccounts(), fetchCategories()]);
	});

	// Clean up subscription when component is destroyed
	onDestroy(() => {
		unsubConnected();
		unsubMessages();
	});
</script>

<div class="container mx-auto p-4">
	<div class="flex justify-between">
		<h1 class="mb-6 text-3xl font-bold">{$t('page.my-accounts')}</h1>
		<!-- button to create new account -->
		<button class="btn btn-primary" on:click={createAccount}>
			<Plus size={20} />
			<Wallet size={20} /></button
		>
	</div>

	{#if error}
		<div class="alert alert-error">
			<p>{error}</p>
		</div>
	{:else}
		<!-- Render the Accounts component -->
		<Accounts
			{accounts}
			{selectedAccount}
			on:select={handleSelectAccount}
			on:updatedAccount={handleUpdateAccount}
			on:deleteAccount={({ detail: { account } }) => handleDeleteAccount(account)}
		/>

		<!-- Month Selector and Transactions Layout -->
		{#if accounts.length > 0}
			<!-- Month Selector - Simple horizontal layout for all screen sizes -->
			<div class="mb-6">
				<div class="mb-3 flex items-center gap-2">
					<Calendar size={16} class="text-primary" />
					<span class="text-sm font-medium">Select Month:</span>
				</div>
				<div class="flex gap-2 overflow-x-auto pb-2">
					{#each availableMonths as monthData}
						<button
							class="btn btn-sm {selectedMonth === monthData.month &&
							selectedYear === monthData.year
								? 'btn-primary'
								: isCurrentMonth(monthData)
									? 'btn-outline btn-primary'
									: 'btn-ghost'} 
								flex-shrink-0 whitespace-nowrap"
							on:click={() => handleMonthSelect(monthData.month, monthData.year)}
						>
							{formatDate(monthData.month, monthData.year)}
						</button>
					{/each}
				</div>
			</div>

			<!-- Transactions Table - Simple single layout -->
			<TransactionsTable
				{transactions}
				{categories}
				account={selectedAccount}
				on:newTransaction={handleNewTransaction}
				on:updatedTransaction={handleUpdateTransaction}
				on:deleteTransaction={({ detail: { transaction } }) => handleDeleteTransaction(transaction)}
			/>
		{/if}

		<!-- Modal: only rendered when showModal is true -->
		{#if showCreateAccountModal}
			<CreateAccount on:closeModal={closeAccountModal} on:newAccount={handleNewAccount} />
		{/if}
	{/if}
</div>
