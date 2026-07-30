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

	async function loadTrends(token: string | undefined, range: TrendsRangeMonths, kind: TrendsType) {
		if (!token) return;
		const key = `${token}-${range}-${kind}`;
		if (loadKey === key) return;
		loadKey = key;
		loading = true;
		try {
			trends = await dataService.fetchCategoryTrends(token, range, kind);
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
					categoriesCount={trends.categories.length}
					{type}
				/>

				<!-- Keyed on type so switching spending/income resets the legend selection. -->
				{#key type}
					<TrendsChartCard {trends} {type} />
				{/key}

				<TrendsMovers movers={trends.movers} totals={trends.totals} months={trends.months} {type} />
			{:else}
				<div class="flex h-96 items-center justify-center">
					<p class="text-base-content/60">{$t('trends.no-data')}</p>
				</div>
			{/if}
		</div>
	</AccountsSplitLayout>
</div>
