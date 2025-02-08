<!-- src/components/Accounts.svelte -->
<script lang="ts">
	import type { Account } from '$lib/types';

	// Export a prop to receive the accounts array.
	export let accounts: Account[] = [];
</script>

{#if accounts.length > 0}
	<div class="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
		{#each accounts as account}
			<div class="card bg-base-100 shadow-xl">
				<div class="card-body">
					<h2 class="card-title">{account.account_name}</h2>
					<p><span class="font-bold">Balance:</span> {account.balance}</p>
				</div>
			</div>
		{/each}
	</div>
{:else}
	<p class="text-gray-500">No accounts found.</p>
{/if}
