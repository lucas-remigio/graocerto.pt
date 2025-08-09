<!-- src/components/Accounts.svelte -->
<script lang="ts">
	import type { Account, AccountChangeResponse } from '$lib/types';
	import { createEventDispatcher, onMount } from 'svelte';
	import { Pencil, Trash, Plus, Wallet, EyeOff, Eye } from 'lucide-svelte';
	import EditAccount from './EditAccount.svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import CreateAccount from './CreateAccount.svelte';
	import AccountCard from './AccountCard.svelte';
	import { t } from '$lib/i18n';
	import api_axios from '$lib/axios';
	import { flip } from 'svelte/animate';
	import { showNonFavorites, updateShowNonFavorites } from '$lib/stores/uiPreferences';

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

	function toggleShowNonFavorites(value: boolean) {
		updateShowNonFavorites(value);
	}

	const dispatch = createEventDispatcher<any>();

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

	function handleConfirmAccountDeletion() {
		openDeleteAccountModal = true;
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

	function handleMoveUp(event: CustomEvent<{ account: Account }>) {
		moveAccount(event.detail.account, 'up');
	}

	function handleMoveDown(event: CustomEvent<{ account: Account }>) {
		moveAccount(event.detail.account, 'down');
	}

	function handleToggleFavorite(event: CustomEvent<{ account: Account }>) {
		favoriteAccountRequest(event.detail.account);
	}

	function moveAccount(account: Account, direction: 'up' | 'down') {
		// get the current index and calculate the target index
		const idx = accounts.findIndex((acc) => acc.token === account.token);
		const targetIdx = direction === 'up' ? idx - 1 : idx + 1;
		// Ensure target index is within bounds
		if (targetIdx < 0 || targetIdx >= accounts.length) {
			return;
		}

		const newAccounts = [...accounts];
		// swap them outtttt
		[newAccounts[idx], newAccounts[targetIdx]] = [newAccounts[targetIdx], newAccounts[idx]];
		accounts = newAccounts;
		sendReorderRequest();
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
		{#each favoriteAccounts as account, i (account.token)}
			<div animate:flip={{ duration: 500 }}>
				<AccountCard
					{account}
					{selectedAccount}
					canMoveUp={i > 0}
					canMoveDown={i < favoriteAccounts.length - 1}
					on:select={handleCardSelect}
					on:edit={handleCardEdit}
					on:delete={handleCardDelete}
					on:moveUp={handleMoveUp}
					on:moveDown={handleMoveDown}
					on:toggleFavorite={handleToggleFavorite}
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
				{#each nonFavoriteAccounts as account, i (account.token)}
					<div animate:flip={{ duration: 500 }}>
						<AccountCard
							{account}
							{selectedAccount}
							canMoveUp={i > 0}
							canMoveDown={i < nonFavoriteAccounts.length - 1}
							on:select={handleCardSelect}
							on:edit={handleCardEdit}
							on:delete={handleCardDelete}
							on:moveUp={handleMoveUp}
							on:moveDown={handleMoveDown}
							on:toggleFavorite={handleToggleFavorite}
						/>
					</div>
				{/each}
			</div>
			<div class="flex justify-center">
				<button
					class="btn btn-sm btn-ghost mt-2 flex items-center gap-1"
					on:click={() => toggleShowNonFavorites(false)}
				>
					<EyeOff size={16} />
					{$t('page.hide-non-favorite')}
				</button>
			</div>
		{:else}
			<div class="flex justify-center">
				<button
					class="btn btn-sm btn-ghost mt-2 flex items-center gap-1"
					on:click={() => toggleShowNonFavorites(true)}
				>
					<Eye size={16} />
					{$t('page.show-all-accounts')}
				</button>
			</div>
		{/if}
	{/if}
{:else}
	<p class="text-gray-500">{$t('page.no-accounts')}</p>
{/if}

<!-- Create Account Modal -->
{#if showCreateAccountModal}
	<CreateAccount on:closeModal={closeAccountModal} on:newAccount={handleNewAccount} />
{/if}

{#if openEditAccountModal}
	<EditAccount
		account={selectedAccount!}
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
