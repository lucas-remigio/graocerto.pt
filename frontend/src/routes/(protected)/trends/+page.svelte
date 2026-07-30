<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { dataService } from '$lib/services/dataService';
	import type { Account, CategoryTrendsResponse, TrendsRangeMonths, TrendsType } from '$lib/types';
	import { t } from '$lib/i18n';
	import { toastStore } from '$lib/stores/toast';
	import AccountsSplitLayout from '$components/AccountsSplitLayout.svelte';
	import TrendsChart from '$components/TrendsChart.svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import { getContrastTextClass } from '$lib/utils/categoryUtils';
	import { TrendingUp, TrendingDown } from 'lucide-svelte';

	const TOTAL_COLOR = '#6366f1'; // must match TrendsChart's total line

	let accounts = $state<Account[]>([]);
	let selectedAccount = $state<Account | null>(null);
	let accountsLoading = $state(false);
	let isLargeScreen = $state(false);

	let months = $state<TrendsRangeMonths>(12);
	let type = $state<TrendsType>('debit');
	let trends = $state<CategoryTrendsResponse | null>(null);
	let selectedCategoryIds = $state<number[]>([]);
	let showTotal = $state(true);
	let loading = $state(false);
	let loadKey = '';

	const ranges: TrendsRangeMonths[] = [6, 12, 24];

	function updateScreenSize() {
		isLargeScreen = window.innerWidth >= 1024;
	}

	function pickInitialAccount() {
		if (selectedAccount || accounts.length === 0) return;
		const stored = localStorage.getItem('selectedAccount');
		selectedAccount = accounts.find((a) => a.token === stored) ?? accounts[0];
	}

	onMount(async () => {
		updateScreenSize();
		window.addEventListener('resize', updateScreenSize);
		accountsLoading = true;
		try {
			accounts = await dataService.fetchAccounts();
			pickInitialAccount();
		} catch (err) {
			console.error('Error loading accounts:', err);
			toastStore.error($t('errors.failed-load-accounts'));
		} finally {
			accountsLoading = false;
		}
	});

	onDestroy(() => {
		if (typeof window !== 'undefined') window.removeEventListener('resize', updateScreenSize);
	});

	async function loadTrends(token: string | undefined, range: TrendsRangeMonths, kind: TrendsType) {
		if (!token) return;
		const key = `${token}-${range}-${kind}`;
		if (loadKey === key) return;
		loadKey = key;
		loading = true;
		try {
			const result = await dataService.fetchCategoryTrends(token, range, kind);
			trends = result;
			// Drop any picked categories that no longer exist in this dataset.
			const ids = new Set(result.categories.map((c) => c.id));
			selectedCategoryIds = selectedCategoryIds.filter((id) => ids.has(id));
		} catch (err) {
			console.error('Error loading trends:', err);
			toastStore.error($t('errors.failed-load-transactions'));
			trends = null;
		} finally {
			loading = false;
		}
	}

	// Refetch whenever account / range / type changes (deps read as arguments).
	$effect(() => {
		void loadTrends(selectedAccount?.token, months, type);
	});

	function setType(next: TrendsType) {
		if (next === type) return;
		type = next;
		selectedCategoryIds = []; // spending and income have different category sets
	}

	function toggleCategory(id: number) {
		selectedCategoryIds = selectedCategoryIds.includes(id)
			? selectedCategoryIds.filter((x) => x !== id)
			: [...selectedCategoryIds, id];
	}

	/* ---- Account cards (mirror Home/Recurring behaviour) ---- */
	function upsertAccount(account: Account) {
		const idx = accounts.findIndex((a) => a.token === account.token);
		if (idx !== -1) accounts[idx] = account;
		else accounts.push(account);
		accounts.sort((a, b) => a.order_index - b.order_index);
	}

	function handleSelectAccount(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		localStorage.setItem('selectedAccount', selectedAccount.token);
	}

	function handleNewAccount(account: Account) {
		upsertAccount(account);
		selectedAccount = null;
		pickInitialAccount();
	}

	function handleUpdateAccount(account: Account) {
		upsertAccount(account);
		if (selectedAccount?.token === account.token) selectedAccount = account;
	}

	async function handleDeleteAccount(account: Account) {
		try {
			await dataService.deleteAccount(account.token);
			accounts = accounts.filter((a) => a.token !== account.token);
			if (selectedAccount?.token === account.token) {
				selectedAccount = null;
				pickInitialAccount();
			}
			toastStore.success($t('common.success'));
		} catch (err) {
			console.error('Error deleting account:', err);
			toastStore.error($t('errors.failed-create-account'));
		}
	}

	/* ---- Derived ---- */
	let hasData = $derived(!!trends && trends.totals.some((v) => v > 0));
	let hasSeries = $derived(showTotal || selectedCategoryIds.length > 0);
	let windowTotal = $derived(trends ? trends.totals.reduce((sum, v) => sum + v, 0) : 0);
	let monthlyAverage = $derived(
		trends && trends.months.length > 0 ? windowTotal / trends.months.length : 0
	);
	let accentClass = $derived(type === 'debit' ? 'text-error' : 'text-success');
</script>

<div class="container mx-auto flex flex-col p-4">
	<AccountsSplitLayout
		{accounts}
		{selectedAccount}
		{isLargeScreen}
		{accountsLoading}
		showRightPanel={accounts.length > 0}
		on:select={handleSelectAccount}
		on:updatedAccount={({ detail: { account } }) => handleUpdateAccount(account)}
		on:deleteAccount={({ detail: { account } }) => handleDeleteAccount(account)}
		on:newAccount={({ detail: { account } }) => handleNewAccount(account)}
	>
		<div class="divider lg:hidden"></div>

		<!-- Header -->
		<div class="mb-4">
			<h1 class="text-2xl font-bold">{$t('trends.title')}</h1>
			<p class="text-sm text-base-content/60">{$t('trends.subtitle')}</p>
		</div>

		<!-- Controls -->
		<div class="mb-4 flex flex-wrap items-center gap-3">
			<div class="btn-group">
				<button
					class="btn btn-sm gap-1 {type === 'debit' ? 'btn-primary text-base-100' : 'btn-ghost'}"
					onclick={() => setType('debit')}
				>
					<TrendingDown size={15} />
					{$t('trends.spending')}
				</button>
				<button
					class="btn btn-sm gap-1 {type === 'credit' ? 'btn-primary text-base-100' : 'btn-ghost'}"
					onclick={() => setType('credit')}
				>
					<TrendingUp size={15} />
					{$t('trends.income')}
				</button>
			</div>

			<div class="btn-group ml-auto">
				{#each ranges as r (r)}
					<button
						class="btn btn-sm {months === r ? 'btn-primary text-base-100' : 'btn-ghost'}"
						onclick={() => (months = r)}
					>
						{r}M
					</button>
				{/each}
			</div>
		</div>

		{#if loading && !trends}
			<div class="flex h-96 items-center justify-center">
				<span class="loading loading-spinner loading-lg"></span>
			</div>
		{:else if hasData && trends}
			<!-- KPI tiles -->
			<div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-3">
				<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
					<p class="text-xs uppercase tracking-wide opacity-60">{$t('trends.kpi-total')}</p>
					<p class="text-2xl font-bold {accentClass}">{formatCurrency(windowTotal)}</p>
				</div>
				<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
					<p class="text-xs uppercase tracking-wide opacity-60">{$t('trends.kpi-average')}</p>
					<p class="text-2xl font-bold">{formatCurrency(monthlyAverage)}</p>
				</div>
				<div
					class="col-span-2 rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm sm:col-span-1"
				>
					<p class="text-xs uppercase tracking-wide opacity-60">{$t('trends.kpi-categories')}</p>
					<p class="text-2xl font-bold">{trends.categories.length}</p>
				</div>
			</div>

			<!-- Chart card -->
			<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
				{#if hasSeries}
					<TrendsChart {trends} {selectedCategoryIds} {showTotal} />
				{:else}
					<div class="flex h-80 items-center justify-center">
						<p class="text-base-content/50">{$t('trends.pick-series')}</p>
					</div>
				{/if}

				<!-- Toggleable legend: Total + one pill per category -->
				<div class="mt-4 flex flex-wrap gap-2">
					<button
						type="button"
						class="flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm font-medium transition-colors {showTotal
							? 'border-transparent text-white shadow-sm'
							: 'border-base-300 text-base-content/60 hover:bg-base-200'}"
						style={showTotal ? `background-color: ${TOTAL_COLOR};` : ''}
						onclick={() => (showTotal = !showTotal)}
					>
						<span class="h-2.5 w-2.5 rounded-full" style="background-color: {TOTAL_COLOR};"></span>
						{$t('trends.total')}
					</button>

					{#each trends.categories as cat (cat.id)}
						{@const active = selectedCategoryIds.includes(cat.id)}
						<button
							type="button"
							class="flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm font-medium transition-colors {active
								? `border-transparent shadow-sm ${getContrastTextClass(cat.color)}`
								: 'border-base-300 text-base-content/60 hover:bg-base-200'}"
							style={active ? `background-color: ${cat.color};` : ''}
							onclick={() => toggleCategory(cat.id)}
						>
							<span class="h-2.5 w-2.5 rounded-full" style="background-color: {cat.color};"></span>
							{cat.name}
							<span class="text-xs {active ? 'opacity-80' : 'opacity-50'}"
								>{formatCurrency(cat.total)}</span
							>
						</button>
					{/each}
				</div>
			</div>
		{:else}
			<div class="flex h-96 items-center justify-center">
				<p class="text-base-content/60">{$t('trends.no-data')}</p>
			</div>
		{/if}
	</AccountsSplitLayout>
</div>
