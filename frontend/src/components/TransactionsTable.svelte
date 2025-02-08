<!-- src/components/TransactionsTable.svelte -->
<script lang="ts">
	import type { Account, TransactionDto } from '$lib/types';

	// Export props for transactions array and the account name.
	export let transactions: TransactionDto[] = [];
	export let account: Account;

	// when i receive the transactions, i want them ordered by date, but with the most recent on top
	transactions = transactions.sort((a, b) => {
		return new Date(b.date).getTime() - new Date(a.date).getTime();
	});
</script>

{#if transactions && transactions.length > 0}
	<h2 class="mb-4 text-2xl font-semibold">
		Transactions for {account.account_name}
	</h2>
	<div class="overflow-x-auto">
		<table class="table w-full">
			<thead class="text-center">
				<tr>
					<th>Date</th>
					<th>Description</th>
					<th>Category</th>
					<th>Amount</th>
					<th>Balance</th>
				</tr>
			</thead>
			<tbody class="text-center">
				{#each transactions as tx}
					<tr
						class={tx.category.transaction_type.type_slug === 'debit'
							? 'bg-red-100'
							: tx.category.transaction_type.type_slug === 'credit'
								? 'bg-green-100'
								: ''}
					>
						<td>
							{new Date(tx.created_at).toLocaleDateString('pt-PT', {
								day: '2-digit',
								month: '2-digit',
								year: 'numeric'
							})}
						</td>
						<td>{tx.description}</td>
						<td>
							<span
								class="rounded px-2 py-1 text-white"
								style="background-color: {tx.category.color};"
							>
								{tx.category.category_name}
							</span>
						</td>
						<td>{tx.amount}$</td>
						<td>{tx.balance}$</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
	<p class="text-gray-500">
		No transactions found for {account.account_name}.
	</p>
{/if}
