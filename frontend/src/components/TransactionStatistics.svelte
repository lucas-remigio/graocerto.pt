<!-- src/components/TransactionStatistics.svelte -->
<script lang="ts">
	import type { Account, TransactionStatistics } from '$lib/types';
	import {
		BarChart3,
		TrendingUp,
		TrendingDown,
		DollarSign,
		PieChart,
		Bot,
		BarChart
	} from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { formatCurrency } from '$lib/utils/currency';
	import TransactionsHeatmap from './TransactionsHeatmap.svelte';
	import AiFeedback from './AiFeedback.svelte';
	import HierarchicalPieChart from './HierarchicalPieChart.svelte';
	import DetailedCategoriesChart from './DetailedCategoriesChart.svelte';
	import HierarchicalBudgetCard from './HierarchicalBudgetCard.svelte';

	export let selectedMonth: number | null;
	export let selectedYear: number | null;
	export let statistics: TransactionStatistics | null = null;
	export let account: Account;
	export let loading: boolean = false;
	export let error: string = '';

	let statsView: 'transactions' | 'categories' | 'overview' = 'transactions';

	$: isAll = selectedMonth === null && selectedYear === null;

	let showAiFeedbackModal = false;

	$: month = selectedMonth !== null ? selectedMonth : new Date().getMonth() + 1;
	$: year = selectedYear !== null ? selectedYear : new Date().getFullYear();

	function openAiFeedbackModal() {
		if (!statistics || statistics.total_transactions === 0) {
			error = $t('transactions.no-transactions-ai');
			return;
		}
		error = '';
		showAiFeedbackModal = true;
	}

	function closeAiFeedbackModal() {
		showAiFeedbackModal = false;
	}
</script>

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
		<PieChart size={64} class="mx-auto mb-4 text-base-content/50" />
		<h3 class="mb-2 text-lg font-semibold">{$t('statistics.no-data')}</h3>
	</div>
{:else}
	<!-- View toggle for transaction vs category statistics -->
	<div class="mb-4 flex justify-center">
		<div role="tablist" aria-label="Statistics view" class="btn-group">
			<button
				role="tab"
				aria-selected={statsView === 'transactions'}
				class="btn btn-sm {statsView === 'transactions'
					? 'btn-primary text-base-100'
					: 'btn-ghost'}"
				on:click={() => (statsView = 'transactions')}
			>
				<BarChart size={16} class="md:mr-1" />
				<span class="hidden md:inline"
					>{$t('statistics.transactions', { default: 'Transactions' })}</span
				>
			</button>
			<button
				role="tab"
				aria-selected={statsView === 'overview'}
				class="btn btn-sm {statsView === 'overview' ? 'btn-primary text-base-100' : 'btn-ghost'}"
				on:click={() => (statsView = 'overview')}
			>
				<PieChart size={16} class="md:mr-1" />
				<span class="hidden md:inline">{$t('statistics.overview', { default: 'Overview' })}</span>
			</button>
			<button
				role="tab"
				aria-selected={statsView === 'categories'}
				class="btn btn-sm {statsView === 'categories' ? 'btn-primary text-base-100' : 'btn-ghost'}"
				on:click={() => (statsView = 'categories')}
			>
				<BarChart3 size={16} class="md:mr-1" />
				<span class="hidden md:inline"
					>{$t('statistics.categories', { default: 'Categories' })}</span
				>
			</button>
		</div>
	</div>

	{#if statsView === 'transactions'}
		<!-- Compact Statistics Summary -->
		<div class="bg-base-100">
			<div class="p-6 pt-2">
				<!-- AI Feedback Button -->
				{#if !isAll}
					<div class="mb-4 flex justify-center">
						<button
							class="btn btn-primary shadow-lg"
							on:click={openAiFeedbackModal}
							aria-label="Get AI Feedback"
						>
							<Bot size={20} class="text-base-100" />
						</button>
					</div>
				{/if}
				<!-- Main Statistics Row -->
				<div class="grid grid-cols-2 gap-6 md:grid-cols-4">
					<!-- Total Transactions -->
					<div class="text-center">
						<p class="text-xs uppercase tracking-wide opacity-60">
							{$t('statistics.total-transactions')}
						</p>
						<p class="text-xl font-bold text-primary">{statistics.total_transactions}</p>
					</div>

					<!-- Total Credits -->
					<div class="text-center">
						<p class="text-xs uppercase tracking-wide opacity-60">
							{$t('statistics.total-credits')}
						</p>
						<p class="text-xl font-bold text-success">
							+{formatCurrency(statistics.totals.credit)}
						</p>
					</div>

					<!-- Total Debits -->
					<div class="text-center">
						<p class="text-xs uppercase tracking-wide opacity-60">
							{$t('statistics.total-debits')}
						</p>
						<p class="text-xl font-bold text-error">-{formatCurrency(statistics.totals.debit)}</p>
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
				<!-- Gap between sections -->
				<div class="mt-2"></div>

				<!-- Largest Transactions Row -->
				<div class="grid grid-cols-2 gap-4">
					<!-- Largest Credit -->
					<div class="flex items-center justify-between rounded-lg bg-success/10 p-3">
						<div>
							<p class="text-xs uppercase tracking-wide opacity-60">
								{$t('statistics.largest-credit')}
							</p>
							<p class="text-lg font-bold text-success">
								+{formatCurrency(statistics.largest_credit)}
							</p>
						</div>
						<TrendingUp size={20} class="text-success" />
					</div>

					<!-- Largest Debit -->
					<div class="flex items-center justify-between rounded-lg bg-error/10 p-3">
						<div>
							<p class="text-xs uppercase tracking-wide opacity-60">
								{$t('statistics.largest-debit')}
							</p>
							<p class="text-lg font-bold text-error">
								-{formatCurrency(statistics.largest_debit)}
							</p>
						</div>
						<TrendingDown size={20} class="text-error" />
					</div>
				</div>
			</div>
		</div>
		<div class="mx-4 my-6">
			<TransactionsHeatmap
				dailyTransactions={statistics.daily_totals}
				startDate={statistics.start_date}
				endDate={statistics.end_date}
			/>
		</div>
	{/if}

	{#if statsView === 'overview'}
		<!-- Overview: Hierarchical Pie Charts at Top, Detailed Below -->
		<div class="space-y-8">
			<!-- Main Categories (Clickable for drill-down) -->
			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				<div class="rounded-lg bg-base-100 p-6">
					<h3 class="mb-4 text-center text-lg font-semibold text-success">
						{$t('statistics.credit-categories')}
					</h3>
					<HierarchicalPieChart data={statistics.credit_category_breakdown} />
				</div>

				<div class="rounded-lg bg-base-100 p-6">
					<h3 class="mb-4 text-center text-lg font-semibold text-error">
						{$t('statistics.debit-categories')}
					</h3>
					<HierarchicalPieChart data={statistics.debit_category_breakdown} />
				</div>
			</div>

			<!-- Detailed Breakdown (All categories split) -->
			<div class="divider">{$t('statistics.detailed-view', { default: 'Detailed View' })}</div>

			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				<div class="rounded-lg bg-base-100 p-6">
					<h3 class="mb-4 text-center text-lg font-semibold text-success">
						{$t('statistics.all-credit-categories', { default: 'All Credit Categories' })}
					</h3>
					<DetailedCategoriesChart data={statistics.credit_category_breakdown} />
				</div>

				<div class="rounded-lg bg-base-100 p-6">
					<h3 class="mb-4 text-center text-lg font-semibold text-error">
						{$t('statistics.all-debit-categories', { default: 'All Debit Categories' })}
					</h3>
					<DetailedCategoriesChart data={statistics.debit_category_breakdown} />
				</div>
			</div>
		</div>
	{/if}

	{#if statsView === 'categories'}
		<!-- Category Breakdowns with Hierarchical Cards -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<!-- Credit Column -->
			<div class="bg-base-100">
				<div class="px-6 py-4">
					<h3 class="mb-3 text-lg font-semibold text-success">
						{$t('statistics.credit-categories')}
					</h3>
					<div class="space-y-3">
						{#each statistics.credit_category_breakdown ?? [] as cat}
							<HierarchicalBudgetCard category={cat} isCredit={true} />
						{:else}
							<div class="flex h-32 items-center justify-center">
								<div class="text-center">
									<PieChart size={36} class="mx-auto mb-2 text-base-content/30" />
									<p class="text-sm text-base-content/60">{$t('statistics.no-data')}</p>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

			<!-- Debit Column -->
			<div
				class="mt-4 border-t border-base-300 bg-base-100 pt-4 lg:mt-0 lg:border-l lg:border-t-0 lg:pl-6 lg:pt-0"
			>
				<div class="px-6 py-4">
					<h3 class="mb-3 text-lg font-semibold text-error">
						{$t('statistics.debit-categories')}
					</h3>
					<div class="space-y-3">
						{#each statistics.debit_category_breakdown ?? [] as cat}
							<HierarchicalBudgetCard category={cat} isCredit={false} />
						{:else}
							<div class="flex h-32 items-center justify-center">
								<div class="text-center">
									<PieChart size={36} class="mx-auto mb-2 text-base-content/30" />
									<p class="text-sm text-base-content/60">{$t('statistics.no-data')}</p>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{/if}
{/if}

{#if showAiFeedbackModal}
	<AiFeedback {account} {month} {year} closeModal={closeAiFeedbackModal} />
{/if}
