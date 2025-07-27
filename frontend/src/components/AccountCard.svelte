<!-- src/components/AccountCard.svelte -->
<script lang="ts">
	import type { Account } from '$lib/types';
	import { createEventDispatcher } from 'svelte';
	import { Pencil, Trash } from 'lucide-svelte';
	import { hideBalances } from '$lib/stores/uiPreferences';

	export let account: Account;
	export let selectedAccount: Account | null = null;

	const dispatch = createEventDispatcher<{
		select: { account: Account };
		edit: { account: Account };
		delete: { account: Account };
	}>();

	function formatCurrency(amount: number): string {
		// make the currency have a , every 3 digits
		return amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
	}

	function handleCardClick() {
		dispatch('select', { account });
	}

	function handleEditAccount() {
		dispatch('edit', { account });
	}

	function handleDeleteAccount() {
		dispatch('delete', { account });
	}

	$: isSelected = selectedAccount?.token === account.token;
</script>

<div class="group relative">
	<button
		type="button"
		class="card bg-base-100 w-full cursor-pointer p-0 outline-none transition-all duration-200 hover:scale-[1.02] hover:shadow-2xl
		{isSelected ? 'ring-primary ring-2 ' : 'border-base-200 hover:border-primary/20 border shadow-lg'}"
		on:click={handleCardClick}
	>
		<div class="card-body items-start px-6 py-4">
			<h2 class="text-base-content/80 mb-1 truncate text-base font-semibold">
				{account.account_name}
			</h2>
			{#if $hideBalances}
				<p class="text-base-content/60 select-none text-3xl font-bold tracking-widest">••••••</p>
			{:else}
				<p class="text-base-content text-3xl font-bold">{formatCurrency(account.balance)}€</p>
			{/if}
		</div>
	</button>

	<!-- Action buttons container - only visible on hover -->
	{#if isSelected}
		<div
			class="absolute right-2 top-2 flex gap-1 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
		>
			<button
				class="btn btn-ghost btn-sm btn-circle bg-base-100/80 backdrop-blur-sm"
				on:click|stopPropagation={handleEditAccount}
				title="Edit account"
			>
				<Pencil size={16} />
			</button>
			<button
				class="btn btn-ghost btn-sm btn-circle bg-base-100/80 text-error hover:bg-error/20 backdrop-blur-sm"
				on:click|stopPropagation={handleDeleteAccount}
				title="Delete account"
			>
				<Trash size={16} />
			</button>
		</div>
	{/if}
</div>
