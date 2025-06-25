<!-- src/components/Accounts.svelte -->
<script lang="ts">
	import type { Account } from '$lib/types';
	import { createEventDispatcher, onMount } from 'svelte';
	import { Pencil, Trash, Plus, Wallet } from 'lucide-svelte';
	import EditAccount from './EditAccount.svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import CreateAccount from './CreateAccount.svelte';
	import AccountCard from './AccountCard.svelte';
	import { t } from '$lib/i18n';

	// Export a prop to receive the accounts array.
	export let accounts: Account[] = [];
	export let selectedAccount: Account | null = null;
	export let isVertical: boolean = false;
	export let loading: boolean = false;

	let openEditAccountModal: boolean = false;
	let openDeleteAccountModal: boolean = false;
	let showCreateAccountModal: boolean = false;

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

	function handleDeleteAccount() {
		openDeleteAccountModal = false;
		dispatch('deleteAccount', { account: selectedAccount! });
	}

	function handleUpdatedAccount() {
		handleCloseEditAccountModal();
		dispatch('updatedAccount');
	}

	function createAccount() {
		showCreateAccountModal = true;
	}

	function closeAccountModal() {
		showCreateAccountModal = false;
	}

	function handleNewAccount() {
		closeAccountModal();
		dispatch('newAccount');
	}
</script>

<!-- Header with title and create button -->
<div class="mb-6 flex justify-between">
	<h1 class="text-3xl font-bold">{$t('page.my-accounts')}</h1>
	<button class="btn btn-primary" on:click={createAccount}>
		<Plus size={20} class="text-base-content" />
		<Wallet size={20} class="text-base-content" />
	</button>
</div>

{#if loading}
	<!-- Loading State -->
	<div class="py-12 text-center">
		<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
		<p class="text-base-content/70">{$t('common.loading')}</p>
	</div>
{:else if accounts.length > 0}
	<div
		class="p-1 {isVertical
			? 'flex max-h-[calc(100vh-200px)] flex-col gap-4 overflow-y-auto pr-2'
			: 'grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4'}"
	>
		{#each accounts as account}
			<AccountCard
				{account}
				{selectedAccount}
				on:select={handleCardSelect}
				on:edit={handleCardEdit}
				on:delete={handleCardDelete}
			/>
		{/each}
	</div>
{:else}
	<p class="text-gray-500">No accounts found.</p>
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

<!-- Create Account Modal -->
{#if showCreateAccountModal}
	<CreateAccount on:closeModal={closeAccountModal} on:newAccount={handleNewAccount} />
{/if}
