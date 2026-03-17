<!-- src/components/Accounts.svelte -->
<script lang="ts">
	import type { Account, AccountChangeResponse } from '$lib/types';
	import { createEventDispatcher } from 'svelte';
	import { Plus, Wallet, EyeOff, Eye } from 'lucide-svelte';
	import AccountModal from './AccountModal.svelte';
	import ConfirmAction from './ConfirmAction.svelte';

	import AccountCard from './AccountCard.svelte';
	import { t } from '$lib/i18n';
	import api_axios from '$lib/axios';
	import { flip } from 'svelte/animate';
	import {
		showNonFavorites,
		updateShowNonFavorites,
		hideBalances
	} from '$lib/stores/uiPreferences';
	import { formatCurrency } from '$lib/utils/currency';

	// Export a prop to receive the accounts array.
	export let accounts: Account[] = [];
	export let selectedAccount: Account | null = null;
	export let isVertical: boolean = false;
	export let loading: boolean = false;

	let openEditAccountModal: boolean = false;
	let openDeleteAccountModal: boolean = false;
	let showCreateAccountModal: boolean = false;

	$: favoriteAccounts = accounts.filter((acc) => acc.is_favorite);
	$: nonFavoriteAccounts = accounts.filter((acc) => !acc.is_favorite);
	$: totalNetWorth = accounts.reduce((sum, acc) => sum + acc.balance, 0);

	function toggleShowNonFavorites(value: boolean) {
		updateShowNonFavorites(value);
	}

	const dispatch = createEventDispatcher<{
		select: { account: Account };
		deleteAccount: { account: Account };
		updatedAccount: AccountChangeResponse;
		newAccount: AccountChangeResponse;
	}>();

	function handleCardSelect(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		dispatch('select', { account: event.detail.account });
	}

	function handleCardEdit(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		openEditAccountModal = true;
	}

	function handleCardDelete(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		openDeleteAccountModal = true;
	}

	function handleCloseEditAccountModal() {
		openEditAccountModal = false;
	}

	function handleCloseDeleteAccountModal() {
		openDeleteAccountModal = false;
	}

	function createAccount() {
		showCreateAccountModal = true;
	}

	function closeAccountModal() {
		showCreateAccountModal = false;
	}

	function handleDeleteAccount() {
		openDeleteAccountModal = false;
		dispatch('deleteAccount', { account: selectedAccount! });
	}

	function handleUpdatedAccount(event: CustomEvent<AccountChangeResponse>) {
		handleCloseEditAccountModal();
		dispatch('updatedAccount', event.detail);
	}

	function handleNewAccount(event: CustomEvent<AccountChangeResponse>) {
		closeAccountModal();
		dispatch('newAccount', event.detail);
	}

	function handleToggleFavorite(event: CustomEvent<{ account: Account }>) {
		favoriteAccountRequest(event.detail.account);
	}

	// ── Drag-and-drop ──────────────────────────────────────────────────
	let draggedToken: string | null = null;
	let dragOverToken: string | null = null;

	function isFavorite(token: string) {
		return favoriteAccounts.some((a) => a.token === token);
	}

	function handleDragStart(token: string, event: DragEvent) {
		draggedToken = token;
		if (event.dataTransfer) {
			event.dataTransfer.effectAllowed = 'move';
			event.dataTransfer.setData('text/plain', token);
		}
	}

	function handleDragOver(token: string, event: DragEvent) {
		event.preventDefault();
		if (draggedToken && draggedToken !== token && isFavorite(draggedToken) === isFavorite(token)) {
			dragOverToken = token;
			if (event.dataTransfer) event.dataTransfer.dropEffect = 'move';
		}
	}

	function handleDragLeave(event: DragEvent) {
		const related = event.relatedTarget as HTMLElement | null;
		if (!related || !(event.currentTarget as HTMLElement).contains(related)) {
			dragOverToken = null;
		}
	}

	function handleDrop(targetToken: string, event: DragEvent) {
		event.preventDefault();
		if (
			!draggedToken ||
			draggedToken === targetToken ||
			isFavorite(draggedToken) !== isFavorite(targetToken)
		) {
			draggedToken = null;
			dragOverToken = null;
			return;
		}

		const favs = isFavorite(draggedToken);
		const group = favs ? [...favoriteAccounts] : [...nonFavoriteAccounts];
		const other = favs ? [...nonFavoriteAccounts] : [...favoriteAccounts];

		const fromIdx = group.findIndex((a) => a.token === draggedToken);
		const toIdx = group.findIndex((a) => a.token === targetToken);
		const [dragged] = group.splice(fromIdx, 1);
		group.splice(toIdx, 0, dragged);

		accounts = favs ? [...group, ...other] : [...other, ...group];
		sendReorderRequest();
		draggedToken = null;
		dragOverToken = null;
	}

	function handleDragEnd() {
		draggedToken = null;
		dragOverToken = null;
	}

	// ── Mobile up/down reorder ───────────────────────────────────────────────
	function moveAccount(token: string, direction: 'up' | 'down') {
		const favs = isFavorite(token);
		const group = favs ? [...favoriteAccounts] : [...nonFavoriteAccounts];
		const other = favs ? [...nonFavoriteAccounts] : [...favoriteAccounts];
		const idx = group.findIndex((a) => a.token === token);
		const targetIdx = direction === 'up' ? idx - 1 : idx + 1;
		if (targetIdx < 0 || targetIdx >= group.length) return;
		const [moved] = group.splice(idx, 1);
		group.splice(targetIdx, 0, moved);
		accounts = favs ? [...group, ...other] : [...other, ...group];
		sendReorderRequest();
	}

	function canMoveUp(token: string): boolean {
		const group = isFavorite(token) ? favoriteAccounts : nonFavoriteAccounts;
		return group.findIndex((a) => a.token === token) > 0;
	}

	function canMoveDown(token: string): boolean {
		const group = isFavorite(token) ? favoriteAccounts : nonFavoriteAccounts;
		const idx = group.findIndex((a) => a.token === token);
		return idx >= 0 && idx < group.length - 1;
	}

	async function sendReorderRequest() {
		const payload = {
			accounts: accounts.map((acc, idx) => ({
				token: acc.token,
				order_index: idx + 1
			}))
		};
		// Replace with your API endpoint and auth logic
		try {
			await api_axios.post('/accounts/reorder', payload);
		} catch (error) {
			console.error('Error reordering accounts:', error);
		}
	}
	async function favoriteAccountRequest(account: Account) {
		try {
			await api_axios.patch(`/accounts/${account.token}/favorite`, {
				is_favorite: !account.is_favorite
			});
			// Update the local accounts array
			accounts = accounts.map((acc) =>
				acc.token === account.token ? { ...acc, is_favorite: !acc.is_favorite } : acc
			);
		} catch (error) {
			console.error('Error toggling favorite:', error);
		}
	}
</script>

<!-- Header with title and create button -->
<div class="mb-2 flex items-center justify-between">
	<h1 class="text-3xl font-bold">{$t('page.my-accounts')}</h1>
	<button class="btn btn-primary" on:click={createAccount} aria-label="Create new account">
		<Plus size={20} class="text-base-100" />
		<Wallet size={20} class="text-base-100" />
	</button>
</div>

{#if loading}
	<!-- Loading State -->
	<div class="py-12 text-center">
		<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
		<p class="text-base-content/70">{$t('common.loading')}</p>
	</div>
{:else if accounts.length > 0}
	<!-- Favorites -->
	<div
		class={isVertical
			? 'flex max-h-[calc(100vh-200px)] flex-col gap-4 overflow-y-auto p-2'
			: 'grid grid-cols-1 gap-4 p-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4'}
	>
		{#each favoriteAccounts as account (account.token)}
			<div
				role="listitem"
				data-token={account.token}
				animate:flip={{ duration: 300 }}
				draggable="true"
				on:dragstart={(e) => handleDragStart(account.token, e)}
				on:dragover={(e) => handleDragOver(account.token, e)}
				on:dragleave={handleDragLeave}
				on:drop={(e) => handleDrop(account.token, e)}
				on:dragend={handleDragEnd}
				class="transition-all duration-150
					{dragOverToken === account.token
					? 'scale-[1.03] rounded-2xl ring-2 ring-primary ring-offset-2'
					: ''}
					{draggedToken === account.token ? 'opacity-40' : ''}"
			>
				<AccountCard
					{account}
					{selectedAccount}
					canMoveUp={canMoveUp(account.token)}
					canMoveDown={canMoveDown(account.token)}
					on:select={handleCardSelect}
					on:edit={handleCardEdit}
					on:delete={handleCardDelete}
					on:toggleFavorite={handleToggleFavorite}
					on:moveUp={() => moveAccount(account.token, 'up')}
					on:moveDown={() => moveAccount(account.token, 'down')}
				/>
			</div>
		{/each}
	</div>

	<!-- Non-favorites toggle and section -->
	{#if nonFavoriteAccounts.length}
		{#if $showNonFavorites}
			<div
				class="mt-2 opacity-70 {isVertical
					? 'flex flex-col gap-4'
					: 'grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4'}"
			>
				{#each nonFavoriteAccounts as account (account.token)}
					<div
						role="listitem"
						data-token={account.token}
						animate:flip={{ duration: 300 }}
						draggable="true"
						on:dragstart={(e) => handleDragStart(account.token, e)}
						on:dragover={(e) => handleDragOver(account.token, e)}
						on:dragleave={handleDragLeave}
						on:drop={(e) => handleDrop(account.token, e)}
						on:dragend={handleDragEnd}
						class="transition-all duration-150
							{dragOverToken === account.token
							? 'scale-[1.03] rounded-2xl ring-2 ring-primary ring-offset-2'
							: ''}
							{draggedToken === account.token ? 'opacity-40' : ''}"
					>
						<AccountCard
							{account}
							{selectedAccount}
							canMoveUp={canMoveUp(account.token)}
							canMoveDown={canMoveDown(account.token)}
							on:select={handleCardSelect}
							on:edit={handleCardEdit}
							on:delete={handleCardDelete}
							on:toggleFavorite={handleToggleFavorite}
							on:moveUp={() => moveAccount(account.token, 'up')}
							on:moveDown={() => moveAccount(account.token, 'down')}
						/>
					</div>
				{/each}
			</div>
			<div class="flex justify-center">
				<button
					class="btn btn-ghost btn-sm mt-2 flex items-center gap-1"
					on:click={() => toggleShowNonFavorites(false)}
				>
					<EyeOff size={16} />
					{$t('page.hide-non-favorite')}
				</button>
			</div>
		{:else}
			<div class="flex justify-center">
				<button
					class="btn btn-ghost btn-sm mt-2 flex items-center gap-1"
					on:click={() => toggleShowNonFavorites(true)}
				>
					<Eye size={16} />
					{$t('page.show-all-accounts')}
				</button>
			</div>
		{/if}
	{/if}
	<!-- Net worth summary -->
	<div class="mt-4 flex items-center justify-between border-t border-base-300 pt-3">
		<span class="text-sm text-base-content/50">{$t('accounts.total-networth')}</span>
		<span class="text-md tabular-nums text-base-content/70">
			{$hideBalances ? '••••••' : formatCurrency(totalNetWorth)}
		</span>
	</div>
{:else}
	<p class="text-gray-500">{$t('page.no-accounts')}</p>
{/if}

<!-- Create Account Modal -->
{#if showCreateAccountModal}
	<AccountModal account={null} on:closeModal={closeAccountModal} on:newAccount={handleNewAccount} />
{/if}

{#if openEditAccountModal}
	<AccountModal
		account={selectedAccount}
		on:closeModal={handleCloseEditAccountModal}
		on:updatedAccount={handleUpdatedAccount}
	/>
{/if}

{#if openDeleteAccountModal}
	<ConfirmAction
		title={`Delete Account ${selectedAccount?.account_name}`}
		message={`${$t('modals.delete-account-confirm')} ${selectedAccount?.account_name}? ${$t('modals.cannot-be-undone')}`}
		type="danger"
		onConfirm={() => handleDeleteAccount()}
		onCancel={() => handleCloseDeleteAccountModal()}
	/>
{/if}
