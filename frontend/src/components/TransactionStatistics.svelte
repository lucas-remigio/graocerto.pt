<!-- src/components/TransactionStatistics.svelte -->
<script lang="ts">
	import type { Account, TransactionStatistics } from '$lib/types';
	import { BarChart3, TrendingUp, TrendingDown, DollarSign, PieChart } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import PieChartComponent from './CategoriesPieChart.svelte';

	export let account: Account;
	export let selectedMonth: number | null;
	export let selectedYear: number | null;
	export let formatedDate: string = '';
	export let statistics: TransactionStatistics | null = null;
	export let loading: boolean = false;
	export let error: string = '';

	$: isAll = selectedMonth === null && selectedYear === null;

	function formatCurrency(amount: number): string {
		return amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
	}
</script>

<!-- Header -->
<div class="flex items-center gap-3">
	<BarChart3 size={24} class="text-primary" />
	<h2 class="text-2xl font-bold">
		{$t('statistics.title')}
		{account.account_name}
		{#if !isAll}
			- {formatedDate}
		{/if}
	</h2>
</div>
{#if loading}
	<!-- Loading State -->
	<div class="py-12 text-center">
		<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
		<p class="text-base-content/70">{$t('common.loading')}</p>
	</div>
{:else if error}
	<!-- Error State -->
	<div class="alert alert-error">
		<p>{error}</p>
	</div>
{:else if !statistics || statistics.total_transactions === 0}
	<!-- Empty State -->
	<div class="py-12 text-center">
		<PieChart size={64} class="text-base-content/50 mx-auto mb-4" />
		<h3 class="mb-2 text-lg font-semibold">{$t('statistics.no-data')}</h3>
	</div>
{:else}
	<!-- Compact Statistics Summary -->
	<div class="card bg-base-100 shadow-lg">
		<div class="card-body p-6">
			<!-- Main Statistics Row -->
			<div class="grid grid-cols-2 gap-6 md:grid-cols-4">
				<!-- Total Transactions -->
				<div class="text-center">
					<p class="text-xs uppercase tracking-wide opacity-60">
						{$t('statistics.total-transactions')}
					</p>
					<p class="text-primary text-xl font-bold">{statistics.total_transactions}</p>
				</div>

				<!-- Total Credits -->
				<div class="text-center">
					<p class="text-xs uppercase tracking-wide opacity-60">
						{$t('statistics.total-credits')}
					</p>
					<p class="text-success text-xl font-bold">
						+{formatCurrency(statistics.totals.credit)}€
					</p>
				</div>

				<!-- Total Debits -->
				<div class="text-center">
					<p class="text-xs uppercase tracking-wide opacity-60">
						{$t('statistics.total-debits')}
					</p>
					<p class="text-error text-xl font-bold">-{formatCurrency(statistics.totals.debit)}€</p>
				</div>

				<!-- Net Balance -->
				<div class="text-center">
					<p class="text-xs uppercase tracking-wide opacity-60">
						{$t('transactions.net-balance')}
					</p>
					<div class="flex items-center justify-center gap-1">
						<p
							class="text-xl font-bold {statistics.totals.difference >= 0
								? 'text-success'
								: 'text-error'}"
						>
							{statistics.totals.difference >= 0 ? '+' : ''}{formatCurrency(
								statistics.totals.difference
							)}€
						</p>
						{#if statistics.totals.difference >= 0}
							<TrendingUp size={16} class="text-success" />
						{:else}
							<TrendingDown size={16} class="text-error" />
						{/if}
					</div>
				</div>
			</div>

			<div class="divider my-1"></div>

			<!-- Largest Transactions Row -->
			<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
				<!-- Largest Credit -->
				<div class="bg-success/10 flex items-center justify-between rounded-lg p-3">
					<div>
						<p class="text-xs uppercase tracking-wide opacity-60">
							{$t('statistics.largest-credit')}
						</p>
						<p class="text-success text-lg font-bold">
							+{formatCurrency(statistics.largest_credit)}€
						</p>
					</div>
					<TrendingUp size={20} class="text-success" />
				</div>

				<!-- Largest Debit -->
				<div class="bg-error/10 flex items-center justify-between rounded-lg p-3">
					<div>
						<p class="text-xs uppercase tracking-wide opacity-60">
							{$t('statistics.largest-debit')}
						</p>
						<p class="text-error text-lg font-bold">
							-{formatCurrency(statistics.largest_debit)}€
						</p>
					</div>
					<TrendingDown size={20} class="text-error" />
				</div>
			</div>
		</div>
	</div>

	<!-- Category Breakdowns with Pie Charts -->
	<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
		<!-- Credit Categories -->
		{#if statistics.credit_category_breakdown && statistics.credit_category_breakdown.length > 0}
			<div class="card bg-base-100 shadow-lg">
				<div class="card-body py-0 px-6">
					<h3 class="card-title text-success mb-4">
						{$t('statistics.credit-categories')}
					</h3>
					<PieChartComponent
						data={statistics.credit_category_breakdown}
						title={$t('statistics.credit-categories')}
					/>
				</div>
			</div>
		{/if}

		<!-- Debit Categories -->
		{#if statistics.debit_category_breakdown && statistics.debit_category_breakdown.length > 0}
			<div class="card bg-base-100 shadow-lg">
				<div class="card-body py-0 px-6">
					<h3 class="card-title text-error mb-4">
						{$t('statistics.debit-categories')}
					</h3>
					<PieChartComponent
						data={statistics.debit_category_breakdown}
						title={$t('statistics.debit-categories')}
					/>
				</div>
			</div>
		{/if}
	</div>
{/if}
