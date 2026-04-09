<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { dataService } from '$lib/services/dataService';
	import type { Account, CategoryDto, RecurringRule } from '$lib/types';
	import { t } from '$lib/i18n';
	import { ArrowRightLeft, Plus, Repeat } from 'lucide-svelte';
	import RecurringRuleModal from '$components/RecurringRuleModal.svelte';
	import RecurringTransferModal from '$components/RecurringTransferModal.svelte';
	import RecurringRulesTable from '$components/RecurringRulesTable.svelte';
	import AccountsSplitLayout from '$components/AccountsSplitLayout.svelte';
	import ConfirmAction from '$components/ConfirmAction.svelte';

	let loading = $state(true);
	let error = $state('');
	let recurringRules: RecurringRule[] = $state([]);
	let showCreateRecurringModal = $state(false);
	let showCreateRecurringTransferModal = $state(false);
	let showEditRecurringModal = $state(false);
	let showDeleteRecurringModal = $state(false);
	let selectedRule: RecurringRule | null = $state(null);
	let pendingDeleteRuleId: number | null = $state(null);
	let accounts: Account[] = $state([]);
	let categories: CategoryDto[] = $state([]);
	let accountsLoading = $state(false);
	let selectedAccount: Account | null = $state(null);
	let isLargeScreen: boolean = $state(false);

	let filteredRecurringRules = $derived(
		selectedAccount
			? recurringRules.filter((rule) => rule.account_token === selectedAccount!.token)
			: []
	);

	function updateScreenSize() {
		isLargeScreen = window.innerWidth >= 1024;
	}

	function getSelectedAccount() {
		if (selectedAccount || accounts.length === 0) return;
		const storedAccountToken = localStorage.getItem('selectedAccount');
		selectedAccount =
			accounts.find((account) => account.token === storedAccountToken) || accounts[0];
	}

	async function loadData() {
		loading = true;
		error = '';
		try {
			const rules = await dataService.fetchRecurringRules();
			recurringRules = rules;
		} catch (err) {
			console.error(err);
			error = $t('errors.failed-load-data');
		} finally {
			loading = false;
		}
	}

	async function fetchAccounts(showLoading: boolean) {
		accountsLoading = showLoading;
		try {
			accounts = await dataService.fetchAccounts();
			getSelectedAccount();
		} catch (err) {
			console.error('Error in fetchAccounts:', err);
			error = $t('errors.failed-load-accounts');
		} finally {
			accountsLoading = false;
		}
	}

	async function fetchCategories() {
		try {
			categories = await dataService.fetchCategories();
		} catch (err) {
			console.error('Error in fetchCategories:', err);
			error = $t('errors.failed-load-data');
		}
	}

	async function deleteRule(ruleId: number) {
		try {
			await dataService.deleteRecurringRule(ruleId);
			recurringRules = recurringRules.filter((r) => r.id !== ruleId);
		} catch (err) {
			console.error(err);
			error = $t('errors.server-error');
		}
	}

	function handleRequestDeleteRule(ruleId: number) {
		pendingDeleteRuleId = ruleId;
		showDeleteRecurringModal = true;
	}

	function handleDeleteRecurringCancel() {
		showDeleteRecurringModal = false;
		pendingDeleteRuleId = null;
	}

	async function handleDeleteRecurringConfirm() {
		if (!pendingDeleteRuleId) return;
		await deleteRule(pendingDeleteRuleId);
		showDeleteRecurringModal = false;
		pendingDeleteRuleId = null;
	}

	async function handleNewRecurringRule(event: CustomEvent<RecurringRule>) {
		showCreateRecurringModal = false;
		const created = event.detail;
		recurringRules = [created, ...recurringRules];
		await loadData();
	}

	async function handleNewRecurringTransfer(event: CustomEvent<{ rules: RecurringRule[] }>) {
		showCreateRecurringTransferModal = false;
		const createdRules = event.detail.rules;
		recurringRules = [...createdRules, ...recurringRules];
		await loadData();
	}

	async function handleUpdateRecurringRule(event: CustomEvent<RecurringRule>) {
		showEditRecurringModal = false;
		selectedRule = null;
		const updated = event.detail;
		recurringRules = recurringRules.map((rule) => (rule.id === updated.id ? updated : rule));
		await loadData();
	}

	function handleEditRule(rule: RecurringRule) {
		selectedRule = rule;
		showEditRecurringModal = true;
	}

	function handleSelectAccount(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		localStorage.setItem('selectedAccount', selectedAccount.token);
	}

	onMount(async () => {
		await Promise.all([fetchAccounts(true), fetchCategories(), loadData()]);
		updateScreenSize();
		window.addEventListener('resize', updateScreenSize);
	});

	onDestroy(() => {
		window.removeEventListener('resize', updateScreenSize);
	});
</script>

<div class="container mx-auto p-4">
	{#if error}
		<div class="alert alert-error mb-4">{error}</div>
	{/if}

	<AccountsSplitLayout
		{accounts}
		{selectedAccount}
		{isLargeScreen}
		accountsLoading={accountsLoading}
		showRightPanel={!!selectedAccount}
		on:select={handleSelectAccount}
	>
		<div class="my-2 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
			<div
				class="order-1 flex items-center justify-center gap-4 md:order-2 md:ml-auto md:justify-end"
			>
				<button
					class="btn btn-secondary shadow-lg"
					aria-label={$t('recurring.new-recurring-transfer')}
					onclick={() => (showCreateRecurringTransferModal = true)}
				>
					<ArrowRightLeft size={20} class="text-base-100" />
				</button>
				<button class="btn btn-primary shadow-lg" onclick={() => (showCreateRecurringModal = true)}>
					<Plus size={20} class="text-base-100" />
					<Repeat size={20} class="text-base-100" />
				</button>
			</div>
			<div class="order-2 flex justify-center md:order-1 md:justify-start">
				<span class="badge badge-outline badge-lg">
					{$t('recurring.templates-active', {
						values: { count: filteredRecurringRules.length }
					})}
				</span>
			</div>
		</div>

		{#if loading}
			<div class="loading loading-spinner loading-lg"></div>
		{:else if filteredRecurringRules.length === 0}
			<div class="flex h-64 flex-col items-center justify-center">
				<p class="text-base-content/70">{$t('recurring.no-templates-yet')}</p>
				<button
					class="btn btn-primary mt-4 flex items-center gap-2 shadow-lg"
					onclick={() => (showCreateRecurringModal = true)}
				>
					<Plus size={20} class="text-base-100" />
					<Repeat size={20} class="text-base-100" />
					<span class="text-base-100">{$t('recurring.create-first-recurring')}</span>
				</button>
			</div>
		{:else}
			<RecurringRulesTable
				recurringRules={filteredRecurringRules}
				{categories}
				on:editRule={({ detail: { rule } }) => handleEditRule(rule)}
				on:deleteRule={({ detail: { ruleId } }) => handleRequestDeleteRule(ruleId)}
			/>
		{/if}
	</AccountsSplitLayout>
</div>

{#if showCreateRecurringModal}
	<RecurringRuleModal
		account={selectedAccount!}
		recurringRule={null}
		on:closeModal={() => (showCreateRecurringModal = false)}
		on:newRecurringRule={handleNewRecurringRule}
	/>
{/if}

{#if showCreateRecurringTransferModal}
	<RecurringTransferModal
		account={selectedAccount!}
		on:closeModal={() => (showCreateRecurringTransferModal = false)}
		on:newRecurringTransfer={handleNewRecurringTransfer}
	/>
{/if}

{#if showEditRecurringModal && selectedRule}
	<RecurringRuleModal
		account={selectedAccount!}
		recurringRule={selectedRule}
		on:closeModal={() => {
			showEditRecurringModal = false;
			selectedRule = null;
		}}
		on:updateRecurringRule={handleUpdateRecurringRule}
	/>
{/if}

{#if showDeleteRecurringModal}
	<ConfirmAction
		title={$t('modals.confirm')}
		message={`${$t('recurring.delete-recurring-confirm')} ${$t('modals.cannot-be-undone')}`}
		type="danger"
		onConfirm={handleDeleteRecurringConfirm}
		onCancel={handleDeleteRecurringCancel}
	></ConfirmAction>
{/if}
