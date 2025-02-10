<!-- src/components/Accounts.svelte -->
<script lang="ts">
	import type { Account } from '$lib/types';
	import { createEventDispatcher } from 'svelte';
	import Page from '../routes/+page.svelte';

	// Export a prop to receive the accounts array.
	export let accounts: Account[] = [];

	function formatCurrency(amount: number): string {
		// make the currency have a , every 3 digits
		return amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
	}

	const dispatch = createEventDispatcher<any>();

	function handleCardClick(account: Account) {
		dispatch('select', { account });
	}
</script>

{#if accounts.length > 0}
	<div class="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
		{#each accounts as account}
			<button
				type="button"
				class="card bg-base-100 cursor-pointer border-none p-0 shadow-xl outline-none"
				on:click={() => handleCardClick(account)}
			>
				<div class="card-body">
					<h2 class="card-title">{account.account_name}</h2>
					<p class="text-3xl font-bold">{formatCurrency(account.balance)}€</p>
				</div>
			</button>
		{/each}
	</div>
{:else}
	<p class="text-gray-500">No accounts found.</p>
{/if}
