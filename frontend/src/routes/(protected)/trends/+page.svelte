<script lang="ts">
	import { onMount } from 'svelte';
	import { dataService } from '$lib/services/dataService';
	import type { Account, CategoryTrendsResponse, TrendsRangeMonths, TrendsType } from '$lib/types';
	import { t } from '$lib/i18n';
	import { toastStore } from '$lib/stores/toast';
	import AccountsSplitLayout from '$components/AccountsSplitLayout.svelte';
	import TrendsControls from '$components/TrendsControls.svelte';
	import TrendsKpis from '$components/TrendsKpis.svelte';
	import TrendsChartCard from '$components/TrendsChartCard.svelte';
	import TrendsMovers from '$components/TrendsMovers.svelte';
	import { pickInitialAccount, rememberSelectedAccount, upsertAccount } from '$lib/utils/accounts';
	import { useIsLargeScreen } from '$lib/utils/screenSize.svelte';

	const screen = useIsLargeScreen();

	let accounts = $state<Account[]>([]);
	let selectedAccount = $state<Account | null>(null);
	let accountsLoading = $state(false);

	let months = $state<TrendsRangeMonths>(12);
	let type = $state<TrendsType>('debit');
	let trends = $state<CategoryTrendsResponse | null>(null);
	let loading = $state(false);
	let loadKey = '';

	// The two months every month-over-month surface compares (axis indices). The
	// backend resolves the default and echoes the chosen pair back; the user can
	// then pick any two months, which refetches with those indices.
	let baseIdx = $state(-1);
	let curIdx = $state(-1);

	onMount(async () => {
		accountsLoading = true;
		try {
			accounts = await dataService.fetchAccounts();
			selectedAccount = pickInitialAccount(accounts, selectedAccount);
		} catch (err) {
			console.error('Error loading accounts:', err);
			toastStore.error($t('errors.failed-load-accounts'));
		} finally {
			accountsLoading = false;
		}
	});

	// base/cur are the comparison months to request; undefined lets the backend pick
	// its default and echo the resolved pair back (which we mirror into the pickers).
	async function loadTrends(
		token: string | undefined,
		range: TrendsRangeMonths,
		kind: TrendsType,
		base?: number,
		cur?: number
	) {
		if (!token) return;
		const key = `${token}-${range}-${kind}-${base ?? 'd'}-${cur ?? 'd'}`;
		if (loadKey === key) return;
		loadKey = key;
		loading = true;
		try {
			trends = await dataService.fetchCategoryTrends(token, range, kind, base, cur);
			baseIdx = trends.compare_base;
			curIdx = trends.compare_current;
		} catch (err) {
			console.error('Error loading trends:', err);
			toastStore.error($t('errors.failed-load-transactions'));
			trends = null;
		} finally {
			loading = false;
		}
	}

	// Structural refetch: account / range / type. Comparison is left to the backend
	// default here (does not read baseIdx/curIdx, so echoing them back won't loop).
	$effect(() => {
		void loadTrends(selectedAccount?.token, months, type);
	});

	// The user picked different comparison months — refetch so the backend re-ranks
	// the movers for that pair. Optimistic index update keeps the pickers responsive.
	function handleCompareChange() {
		void loadTrends(selectedAccount?.token, months, type, baseIdx, curIdx);
	}

	let movers = $derived(trends?.movers ?? []);

	// Category concentration for the KPI tile: the top-N roots as a share of the
	// window total. Thin client-side derivation of aggregates already returned.
	let topN = $derived(trends ? Math.min(3, trends.categories.length) : 0);
	let topShare = $derived.by(() => {
		if (!trends || topN === 0 || trends.window_total <= 0) return null;
		const top = trends.categories.slice(0, topN).reduce((sum, c) => sum + c.total, 0);
		return top / trends.window_total;
	});

	/* ---- Account cards (mirror Home/Recurring behaviour) ---- */
	function handleSelectAccount(event: CustomEvent<{ account: Account }>) {
		selectedAccount = event.detail.account;
		rememberSelectedAccount(selectedAccount.token);
	}

	function handleNewAccount(account: Account) {
		accounts = upsertAccount(accounts, account);
		selectedAccount = null;
		selectedAccount = pickInitialAccount(accounts, selectedAccount);
	}

	function handleUpdateAccount(account: Account) {
		accounts = upsertAccount(accounts, account);
		if (selectedAccount?.token === account.token) selectedAccount = account;
	}

	async function handleDeleteAccount(account: Account) {
		try {
			await dataService.deleteAccount(account.token);
			accounts = accounts.filter((a) => a.token !== account.token);
			if (selectedAccount?.token === account.token) {
				selectedAccount = null;
				selectedAccount = pickInitialAccount(accounts, selectedAccount);
			}
			toastStore.success($t('common.success'));
		} catch (err) {
			console.error('Error deleting account:', err);
			toastStore.error($t('errors.failed-create-account'));
		}
	}

	// Presentation only; all metrics come from the backend.
	let hasData = $derived(!!trends && trends.totals.some((v) => v > 0));
</script>

<div class="container mx-auto flex flex-col p-4">
	<AccountsSplitLayout
		{accounts}
		{selectedAccount}
		isLargeScreen={screen.value}
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

		<TrendsControls bind:type bind:months />

		<!-- Scrollable content; header + controls above stay fixed (mirrors Home/Recurring). -->
		<div class="min-h-0 flex-1 overflow-y-auto pb-4 lg:pr-2">
			{#if loading && !trends}
				<div class="flex h-96 items-center justify-center">
					<span class="loading loading-spinner loading-lg"></span>
				</div>
			{:else if hasData && trends}
				<TrendsKpis
					windowTotal={trends.window_total}
					monthlyAverage={trends.monthly_average}
					{topShare}
					{topN}
					totals={trends.totals}
					{type}
				/>

				<!-- Keyed on type so switching spending/income resets the legend selection. -->
				{#key type}
					<TrendsChartCard {trends} {type} {baseIdx} {curIdx} />
				{/key}

				<TrendsMovers
					{movers}
					totals={trends.totals}
					months={trends.months}
					bind:baseIdx
					bind:curIdx
					onCompareChange={handleCompareChange}
				/>
			{:else}
				<div class="flex h-96 items-center justify-center">
					<p class="text-base-content/60">{$t('trends.no-data')}</p>
				</div>
			{/if}
		</div>
	</AccountsSplitLayout>
</div>
