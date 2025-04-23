<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { connected, socket, messages, sendMessage } from '$lib/ws'; // Import WebSocket utilities
	import api_axios from '$lib/axios';
	import type {
		Account,
		AccountsResponse,
		TransactionDto,
		TransactionsResponseDto,
		CategoryDto
	} from '$lib/types';
	import { Plus, Wallet } from 'lucide-svelte';
	import Accounts from '$components/Accounts.svelte';
	import TransactionsTable from '$components/TransactionsTable.svelte';
	import CreateAccount from '$components/CreateAccount.svelte';
	import { userEmail } from '$lib/stores/auth';

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
				console.log(`JOIN ROOM message sent for ${$userEmail}`);
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
				console.log('Received update via WebSocket:', latestMsg);
				fetchAccounts(); // Refresh data
			}
		}
	});

	// Local component state
	let accounts: Account[] = [];
	let transactions: TransactionDto[] = [];
	let categories: CategoryDto[] = [];
	let error: string = '';
	let showCreateAccountModal = false;

	let selectedAccount: Account;

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
				await getAccountTransactions(selectedAccount.token);
			}
		} catch (err) {
			console.error('Error in fetchAccounts:', err);
			error = 'Failed to load accounts';
		}
	}

	// Function to fetch transactions for a given account token
	async function getAccountTransactions(accountToken: string) {
		try {
			const res = await api_axios('transactions/dto/' + accountToken);

			if (res.status !== 200) {
				console.error('Non-200 response status for transactions:', res.status);
				error = `Error: ${res.status}`;
				return;
			}

			const data: TransactionsResponseDto = res.data;
			transactions = data.transactions;
		} catch (err) {
			console.error('Error in getAccountTransactions:', err);
			error = 'Failed to load transactions';
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

	function handleSelect(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		localStorage.setItem('selectedAccount', selectedAccount.token);
		getAccountTransactions(selectedAccount.token);
	}

	function createAccount() {
		showCreateAccountModal = true;
	}

	function closeAccountModal() {
		showCreateAccountModal = false;
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

	// Trigger the fetching when the component mounts
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
		<h1 class="mb-6 text-3xl font-bold">My Accounts</h1>
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
			on:select={handleSelect}
			on:updatedAccount={handleUpdateAccount}
			on:deleteAccount={({ detail: { account } }) => handleDeleteAccount(account)}
		/>

		<!-- Render the TransactionsTable component only if accounts exist -->
		{#if accounts.length > 0}
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
