<script lang="ts">
	import { onMount } from 'svelte';
	import api_axios from '$lib/axios';
	import type {
		Account,
		AccountsResponse,
		TransactionDto,
		TransactionsResponseDto
	} from '$lib/types';
	import Accounts from '../components/Accounts.svelte';
	import TransactionsTable from '../components/TransactionsTable.svelte';
	import { Wallet } from 'lucide-svelte';
	import CreateAccount from '../components/CreateAccount.svelte';

	// Local component state
	let accounts: Account[] = [];
	let transactions: TransactionDto[] = [];
	let selectedAccount: Account;
	let error: string = '';
	let showCreateAccountModal = false;

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
				selectedAccount = accounts[0];
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

	function handleSelect(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		getAccountTransactions(selectedAccount.token);
	}

	function createAccount() {
		// Instead of just logging, we set showModal to true
		showCreateAccountModal = true;
	}

	function closeAccountModal() {
		showCreateAccountModal = false;
	}

	function handleNewTransaction() {
		getAccountTransactions(selectedAccount.token);
	}

	function handleNewAccount() {
		closeAccountModal();
		fetchAccounts();
	}

	// Trigger the fetching when the component mounts
	onMount(() => {
		fetchAccounts();
	});
</script>

<div class="container mx-auto p-6">
	<div class="flex justify-between">
		<h1 class="mb-6 text-3xl font-bold">My Accounts</h1>
		<!-- button to create new account -->
		<button class="btn btn-primary" on:click={createAccount}><Wallet /></button>
	</div>

	{#if error}
		<div class="alert alert-error">
			<p>{error}</p>
		</div>
	{:else}
		<!-- Render the Accounts component -->
		<Accounts {accounts} on:select={handleSelect} />

		<!-- Render the TransactionsTable component only if accounts exist -->
		{#if accounts.length > 0}
			<TransactionsTable
				{transactions}
				account={selectedAccount}
				on:newTransaction={handleNewTransaction}
			/>
		{/if}
		<!-- Modal: only rendered when showModal is true -->
		{#if showCreateAccountModal}
			<CreateAccount on:closeModal={closeAccountModal} on:newAccount={handleNewAccount} />
		{/if}
	{/if}
</div>
