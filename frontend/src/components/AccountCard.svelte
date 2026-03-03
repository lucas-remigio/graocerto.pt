<!-- src/components/AccountCard.svelte -->
<script lang="ts">
	import type { Account } from '$lib/types';
	import { createEventDispatcher } from 'svelte';
	import { GripVertical, Pencil, Star, Trash, ArrowUp, ArrowDown } from 'lucide-svelte';
	import { hideBalances } from '$lib/stores/uiPreferences';
	import { formatCurrency } from '$lib/utils/currency';

	export let account: Account;
	export let selectedAccount: Account | null = null;
	export let canMoveUp: boolean = false;
	export let canMoveDown: boolean = false;

	const dispatch = createEventDispatcher<{
		select: { account: Account };
		edit: { account: Account };
		delete: { account: Account };
		toggleFavorite: { account: Account };
		moveUp: void;
		moveDown: void;
	}>();

	function handleCardClick() {
		dispatch('select', { account });
	}

	function handleEditAccount() {
		dispatch('edit', { account });
	}

	function handleDeleteAccount() {
		dispatch('delete', { account });
	}

	function handleFavorite(): void {
		dispatch('toggleFavorite', { account });
	}

	$: isSelected = selectedAccount?.token === account.token;
</script>

<div class="group relative">
	<!-- Desktop (lg+): drag grip -->
	<span
		class="absolute left-1 top-1/2 z-10 hidden -translate-y-1/2 cursor-grab p-2 text-base-content/20 opacity-0 transition-opacity duration-200 active:cursor-grabbing group-hover:opacity-100 lg:inline"
	>
		<GripVertical size={20} />
	</span>
	<!-- Mobile/tablet (< lg): up/down reorder buttons -->
	<div
		class="absolute left-1 top-1/2 z-10 flex -translate-y-1/2 flex-col opacity-0 transition-opacity duration-200 group-hover:opacity-100 lg:hidden"
	>
		<button
			class="btn btn-circle btn-ghost btn-xs"
			disabled={!canMoveUp}
			on:click|stopPropagation={() => dispatch('moveUp')}
			aria-label="Move up"
		>
			<ArrowUp size={12} />
		</button>
		<button
			class="btn btn-circle btn-ghost btn-xs"
			disabled={!canMoveDown}
			on:click|stopPropagation={() => dispatch('moveDown')}
			aria-label="Move down"
		>
			<ArrowDown size={12} />
		</button>
	</div>
	<button
		type="button"
		class="card w-full cursor-pointer bg-base-100 p-0 outline-none transition-all duration-200 hover:scale-[1.02] hover:shadow-2xl
		{isSelected ? 'ring-2 ring-primary ' : 'border border-base-200 shadow-lg hover:border-primary/20'}"
		on:click={handleCardClick}
	>
		<div class="card-body items-start px-6 py-4">
			<h2 class="mb-1 truncate text-base font-semibold text-base-content/80">
				{account.account_name}
			</h2>
			{#if $hideBalances}
				<p class="select-none text-3xl font-bold tracking-widest text-base-content/60">••••••</p>
			{:else}
				<p class="text-3xl font-bold text-base-content">{formatCurrency(account.balance)}</p>
			{/if}
		</div>
	</button>

	<!-- Action buttons container - only visible on hover -->
	{#if isSelected}
		<div
			class="absolute right-2 top-2 flex gap-1 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
		>
			<!-- Add a star button -->
			<button
				class="btn btn-circle btn-ghost btn-sm"
				on:click|stopPropagation={handleFavorite}
				title={account.is_favorite ? 'Unfavorite' : 'Favorite'}
			>
				{#if account.is_favorite}
					<Star fill="currentColor" size={16} class="text-amber-400" />
				{:else}
					<Star size={16} />
				{/if}
			</button>
			<button
				class="btn btn-circle btn-ghost btn-sm bg-base-100/80 backdrop-blur-sm"
				on:click|stopPropagation={handleEditAccount}
				title="Edit account"
			>
				<Pencil size={16} />
			</button>
			<button
				class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-error backdrop-blur-sm hover:bg-error/20"
				on:click|stopPropagation={handleDeleteAccount}
				title="Delete account"
			>
				<Trash size={16} />
			</button>
		</div>
	{/if}
</div>
