<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { dataService } from '$lib/services/dataService';
	import type {
		Account,
		CategoryDto,
		RecurringForecastItem,
		RecurringForecastRangeDays,
		RecurringRule
	} from '$lib/types';
	import { t } from '$lib/i18n';
	import { ArrowRightLeft, BarChart3, List, Plus, Repeat } from 'lucide-svelte';
	import RecurringRuleModal from '$components/RecurringRuleModal.svelte';
	import RecurringTransferModal from '$components/RecurringTransferModal.svelte';
	import RecurringForecastTable from '$components/RecurringForecastTable.svelte';
	import RecurringRulesTable from '$components/RecurringRulesTable.svelte';
	import TransactionsStats from '$components/TransactionsStats.svelte';
	import AccountsSplitLayout from '$components/AccountsSplitLayout.svelte';
	import ConfirmAction from '$components/ConfirmAction.svelte';

	let loading = $state(true);
	let error = $state('');
	let recurringRules: RecurringRule[] = $state([]);
	let showCreateRecurringModal = $state(false);
	let showCreateRecurringTransferModal = $state(false);
	let showEditRecurringModal = $state(false);
	let showEditRecurringTransferModal = $state(false);
	let showDeleteRecurringModal = $state(false);
	let selectedRule: RecurringRule | null = $state(null);
	let selectedTransferGroupId: string | null = $state(null);
	let selectedTransferRules: RecurringRule[] = $state([]);
	let pendingDeleteRuleId: number | null = $state(null);
	let accounts: Account[] = $state([]);
	let categories: CategoryDto[] = $state([]);
	let accountsLoading = $state(false);
	let selectedAccount: Account | null = $state(null);
	let isLargeScreen: boolean = $state(false);
	let recurringViewMode = $state<'rules' | 'forecast'>('rules');
	let forecastDays: RecurringForecastRangeDays = $state(30);
	let forecastItems: RecurringForecastItem[] = $state([]);
	let forecastLoading = $state(false);

	let filteredRecurringRules = $derived(
		selectedAccount
			? recurringRules.filter((rule) => rule.account_token === selectedAccount!.token)
			: []
	);

	let forecastSummaryItems = $derived(
		forecastItems.map((item) => ({
			amount: item.amount,
			typeId: item.transaction_type_id
		}))
	);

	let selectedSummaryItems = $derived(
		recurringViewMode === 'forecast'
			? forecastSummaryItems
			: filteredRecurringRules.map((rule) => ({
					amount: rule.amount,
					typeId: rule.transaction_type_id
				}))
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

	async function loadForecast() {
		if (!selectedAccount) {
			forecastItems = [];
			return;
		}
		forecastLoading = true;
		try {
			const forecast = await dataService.fetchRecurringForecast(
				selectedAccount.token,
				forecastDays
			);
			forecastItems = forecast.items;
		} catch (err) {
			console.error(err);
			error = $t('errors.failed-load-data');
		} finally {
			forecastLoading = false;
		}
	}

	async function syncForecastIfNeeded() {
		if (recurringViewMode === 'forecast') {
			await loadForecast();
		}
	}

	async function refreshRulesAndForecast() {
		await loadData();
		await syncForecastIfNeeded();
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
		await syncForecastIfNeeded();
	}

	async function handleNewRecurringRule(event: CustomEvent<RecurringRule>) {
		showCreateRecurringModal = false;
		const created = event.detail;
		recurringRules = [created, ...recurringRules];
		await refreshRulesAndForecast();
	}

	async function handleNewRecurringTransfer(event: CustomEvent<{ rules: RecurringRule[] }>) {
		showCreateRecurringTransferModal = false;
		const createdRules = event.detail.rules;
		recurringRules = [...createdRules, ...recurringRules];
		await refreshRulesAndForecast();
	}

	async function handleUpdateRecurringTransfer(event: CustomEvent<{ rules: RecurringRule[] }>) {
		showEditRecurringTransferModal = false;
		selectedTransferGroupId = null;
		selectedTransferRules = [];

		const updatedRules = event.detail.rules;
		const updatedIds = new Set(updatedRules.map((rule) => rule.id));
		const withoutUpdated = recurringRules.filter((rule) => !updatedIds.has(rule.id));
		recurringRules = [...updatedRules, ...withoutUpdated];
		await refreshRulesAndForecast();
	}

	async function handleUpdateRecurringRule(event: CustomEvent<RecurringRule>) {
		showEditRecurringModal = false;
		selectedRule = null;
		const updated = event.detail;
		recurringRules = recurringRules.map((rule) => (rule.id === updated.id ? updated : rule));
		await refreshRulesAndForecast();
	}

	function handleEditRule(rule: RecurringRule) {
		if (rule.recurring_transfer_group_id) {
			selectedTransferGroupId = rule.recurring_transfer_group_id;
			selectedTransferRules = recurringRules.filter(
				(recurringRule) =>
					recurringRule.recurring_transfer_group_id === rule.recurring_transfer_group_id
			);
			showEditRecurringTransferModal = true;
			return;
		}

		selectedRule = rule;
		showEditRecurringModal = true;
	}

	function handleSelectAccount(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		localStorage.setItem('selectedAccount', selectedAccount.token);
		syncForecastIfNeeded();
	}

	function setRecurringViewMode(mode: 'rules' | 'forecast') {
		recurringViewMode = mode;
		if (mode === 'forecast') {
			loadForecast();
		}
	}

	function setForecastDays(days: RecurringForecastRangeDays) {
		forecastDays = days;
		if (recurringViewMode === 'forecast') {
			loadForecast();
		}
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
		{accountsLoading}
		showRightPanel={!!selectedAccount}
		on:select={handleSelectAccount}
	>
		<div class="mb-2 flex justify-center">
			<div class="btn-group">
				<button
					class="btn btn-sm {recurringViewMode === 'rules' ? 'btn-primary text-base-100' : 'btn-ghost'}"
					onclick={() => setRecurringViewMode('rules')}
				>
					<List size={16} class="mr-1" />
					<span>{$t('recurring.view-rules')}</span>
				</button>
				<button
					class="btn btn-sm {recurringViewMode === 'forecast'
						? 'btn-primary text-base-100'
						: 'btn-ghost'}"
					onclick={() => setRecurringViewMode('forecast')}
				>
					<BarChart3 size={16} class="mr-1" />
					<span>{$t('recurring.view-forecast')}</span>
				</button>
			</div>
		</div>

		<div class="divider my-0"></div>

		<div class="my-2 flex flex-col items-center gap-3 md:flex-row md:items-center">
			<div class="order-1 flex w-full justify-center md:flex-1 md:justify-start">
				<TransactionsStats summaryItems={selectedSummaryItems} />
			</div>

			<div class="order-2 flex justify-center md:px-4">
				<div class="inline-flex items-center gap-2 rounded-full bg-base-200/60 px-3 py-1 shadow-sm">
					<span class="text-xs font-medium uppercase tracking-wide text-base-content/60">
						{$t('recurring.templates-active-label')}
					</span>
					<span class="badge badge-primary badge-md font-bold text-base-100">
						{filteredRecurringRules.length}
					</span>
				</div>
			</div>

			<div class="order-3 flex w-full items-center justify-center gap-4 md:flex-1 md:justify-end">
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
		</div>


		{#if loading}
			<div class="loading loading-spinner loading-lg"></div>
		{:else if recurringViewMode === 'forecast'}
			<div class="mb-3 mt-1 flex justify-center">
				<div class="btn-group">
					<button
						class="btn btn-sm {forecastDays === 30 ? 'btn-primary text-base-100' : 'btn-ghost'}"
						onclick={() => setForecastDays(30)}
					>
						30d
					</button>
					<button
						class="btn btn-sm {forecastDays === 60 ? 'btn-primary text-base-100' : 'btn-ghost'}"
						onclick={() => setForecastDays(60)}
					>
						60d
					</button>
					<button
						class="btn btn-sm {forecastDays === 90 ? 'btn-primary text-base-100' : 'btn-ghost'}"
						onclick={() => setForecastDays(90)}
					>
						90d
					</button>
				</div>
			</div>
			{#if forecastLoading}
				<div class="loading loading-spinner loading-lg"></div>
			{:else if forecastItems.length === 0}
				<div class="flex h-40 flex-col items-center justify-center">
					<p class="text-base-content/70">{$t('recurring.no-forecast-items')}</p>
				</div>
			{:else}
				<RecurringForecastTable items={forecastItems} {categories} />
			{/if}
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
		transferGroupId={null}
		initialRules={[]}
		on:closeModal={() => (showCreateRecurringTransferModal = false)}
		on:newRecurringTransfer={handleNewRecurringTransfer}
	/>
{/if}

{#if showEditRecurringTransferModal && selectedTransferGroupId}
	<RecurringTransferModal
		account={selectedAccount!}
		transferGroupId={selectedTransferGroupId}
		initialRules={selectedTransferRules}
		on:closeModal={() => {
			showEditRecurringTransferModal = false;
			selectedTransferGroupId = null;
			selectedTransferRules = [];
		}}
		on:updateRecurringTransfer={handleUpdateRecurringTransfer}
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
