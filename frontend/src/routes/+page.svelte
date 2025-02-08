<script lang="ts">
	import { onMount } from 'svelte';
	import api_axios from '$lib/axios';
	import type {
		Account,
		AccountsResponse,
		TransactionDto,
		TransactionsResponseDto
	} from '$lib/types';

	// Local component state
	let accounts: Account[] = [];
	let transactions: TransactionDto[] = [];
	let error: string = '';

	// Function to fetch accounts and then fetch transactions for the first account
	async function fetchAccounts() {
		try {
			const res = await api_axios('accounts');
			console.log('Response from accounts api:', res);
			if (res.status !== 200) {
				console.error('Non-200 response status:', res.status);
				error = `Error: ${res.status}`;
				return;
			}
			const data: AccountsResponse = res.data;
			accounts = data.accounts;

			// If we have at least one account, fetch its transactions
			if (accounts && accounts.length > 0) {
				await getAccountTransactions(accounts[0].token);
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

	// Trigger the fetching when the component mounts
	onMount(() => {
		fetchAccounts();
	});
</script>

<h1 class="mb-6 text-3xl font-bold">My Accounts</h1>

{#if error}
	<div class="alert alert-error">
		<p>{error}</p>
	</div>
{:else}
	<!-- Display Accounts as cards -->
	{#if accounts.length > 0}
		<div class="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
			{#each accounts as account}
				<div class="card bg-base-100 shadow-xl">
					<div class="card-body">
						<h2 class="card-title">{account.account_name}</h2>
						<p><span class="font-bold">Token:</span> {account.token}</p>
						<p><span class="font-bold">Balance:</span> {account.balance}</p>
						<p>
							<span class="font-bold">Created at:</span>
							{new Date(account.created_at).toLocaleString()}
						</p>
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<p class="text-gray-500">No accounts found.</p>
	{/if}

	<!-- Optionally, display transactions for the first account -->
	{#if transactions && transactions.length > 0}
		<h2 class="mb-4 text-2xl font-semibold">Transactions for {accounts[0].account_name}</h2>
		<div class="overflow-x-auto">
			<table class="table w-full">
				<thead>
					<tr>
						<th>Id</th>
						<th>Description</th>
						<th>Amount</th>
						<th>Date</th>
					</tr>
				</thead>
				<tbody>
					{#each transactions as tx}
						<tr>
							<td>{tx.id}</td>
							<td>{tx.description}</td>
							<td>{tx.amount}</td>
							<td>{new Date(tx.created_at).toLocaleString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{:else if accounts.length > 0}
		<p class="text-gray-500">No transactions found for {accounts[0].account_name}.</p>
	{/if}
{/if}
