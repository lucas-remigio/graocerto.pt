<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { dataService } from '$lib/services/dataService';
	import type {
		Account,
		CategoryTrend,
		CategoryTrendsResponse,
		TrendsRangeMonths,
		TrendsType
	} from '$lib/types';
	import { t } from '$lib/i18n';
	import { toastStore } from '$lib/stores/toast';
	import AccountsSplitLayout from '$components/AccountsSplitLayout.svelte';
	import TrendsChart from '$components/TrendsChart.svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import { getContrastTextClass } from '$lib/utils/categoryUtils';
	import { TrendingUp, TrendingDown, ChevronDown, ChevronRight } from 'lucide-svelte';

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
	let expandedRootId = $state<number | null>(null);
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
			// Keep only picks (roots or subcategories) that still exist in this dataset.
			const ids = new Set<number>();
			for (const root of result.categories) {
				ids.add(root.id);
				for (const sub of root.subcategories ?? []) ids.add(sub.id);
			}
			selectedCategoryIds = selectedCategoryIds.filter((id) => ids.has(id));
			if (expandedRootId !== null && !result.categories.some((c) => c.id === expandedRootId)) {
				expandedRootId = null;
			}
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
		expandedRootId = null;
	}

	function toggleCategory(id: number) {
		selectedCategoryIds = selectedCategoryIds.includes(id)
			? selectedCategoryIds.filter((x) => x !== id)
			: [...selectedCategoryIds, id];
	}

	function toggleExpand(id: number) {
		expandedRootId = expandedRootId === id ? null : id;
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

	// Flat lookup of every drawable series (roots + subcategories) by id.
	let seriesById = $derived.by(() => {
		const map = new Map<number, CategoryTrend>();
		if (trends) {
			for (const root of trends.categories) {
				map.set(root.id, root);
				for (const sub of root.subcategories ?? []) map.set(sub.id, sub);
			}
		}
		return map;
	});

	let selectedSeries = $derived(
		selectedCategoryIds
			.map((id) => seriesById.get(id))
			.filter((s): s is CategoryTrend => !!s)
			.map((s) => ({ name: s.name, color: s.color, totals: s.totals }))
	);

	let expandedRoot = $derived(
		trends && expandedRootId !== null
			? (trends.categories.find((c) => c.id === expandedRootId) ?? null)
			: null
	);
</script>

<!-- One toggleable series pill (used for roots and subcategories) -->
{#snippet seriesPill(id: number, name: string, color: string, catTotal: number)}
	{@const active = selectedCategoryIds.includes(id)}
	<button
		type="button"
		class="flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm font-medium transition-colors {active
			? `border-transparent shadow-sm ${getContrastTextClass(color)}`
			: 'border-base-300 text-base-content/60 hover:bg-base-200'}"
		style={active ? `background-color: ${color};` : ''}
		onclick={() => toggleCategory(id)}
	>
		<span class="h-2.5 w-2.5 rounded-full" style="background-color: {color};"></span>
		{name}
		<span class="text-xs {active ? 'opacity-80' : 'opacity-50'}">{formatCurrency(catTotal)}</span>
	</button>
{/snippet}

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
					<TrendsChart
						months={trends.months}
						total={trends.totals}
						series={selectedSeries}
						{showTotal}
					/>
				{:else}
					<div class="flex h-80 items-center justify-center">
						<p class="text-base-content/50">{$t('trends.pick-series')}</p>
					</div>
				{/if}

				<!-- Toggleable legend: Total + roots (with drill-down into subcategories) -->
				<div class="mt-4 flex flex-col gap-3">
					<div class="flex flex-wrap items-center gap-2">
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

						{#each trends.categories as root (root.id)}
							<div class="flex items-center">
								{@render seriesPill(root.id, root.name, root.color, root.total)}
								{#if root.subcategories && root.subcategories.length > 0}
									<button
										type="button"
										class="ml-0.5 flex h-7 w-7 items-center justify-center rounded-full text-base-content/50 transition-colors hover:bg-base-200 {expandedRootId ===
										root.id
											? 'text-base-content'
											: ''}"
										onclick={() => toggleExpand(root.id)}
										aria-label={$t('trends.subcategories')}
										aria-expanded={expandedRootId === root.id}
									>
										{#if expandedRootId === root.id}
											<ChevronDown size={16} />
										{:else}
											<ChevronRight size={16} />
										{/if}
									</button>
								{/if}
							</div>
						{/each}
					</div>

					<!-- Drill-down: the expanded root's subcategories -->
					{#if expandedRoot && expandedRoot.subcategories}
						<div class="rounded-lg border border-base-300 bg-base-200/40 p-3">
							<p class="mb-2 flex items-center gap-1.5 text-xs font-semibold opacity-70">
								<span
									class="h-2.5 w-2.5 rounded-full"
									style="background-color: {expandedRoot.color};"
								></span>
								{expandedRoot.name} · {$t('trends.subcategories')}
							</p>
							<div class="flex flex-wrap gap-2">
								{#each expandedRoot.subcategories as sub (sub.id)}
									{@render seriesPill(sub.id, sub.name, sub.color, sub.total)}
								{/each}
							</div>
						</div>
					{/if}
				</div>
			</div>
		{:else}
			<div class="flex h-96 items-center justify-center">
				<p class="text-base-content/60">{$t('trends.no-data')}</p>
			</div>
		{/if}
	</AccountsSplitLayout>
</div>
