<!-- src/components/TransactionStatistics.svelte -->
<script lang="ts">
	import type { Account, TransactionStatistics } from '$lib/types';
	import { BarChart3, TrendingUp, TrendingDown, DollarSign, PieChart } from 'lucide-svelte';
	import { t } from '$lib/i18n';

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

	function getProgressColor(percentage: number): string {
		if (percentage >= 30) return 'bg-error';
		if (percentage >= 15) return 'bg-warning';
		return 'bg-success';
	}
</script>

<div class="space-y-6">
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
			<p class="text-base-content/70">{$t('statistics.no-data-description')}</p>
		</div>
	{:else}
		<!-- Overview Cards -->
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
			<!-- Total Balance Change -->
			<div class="card bg-base-100 shadow-lg">
				<div class="card-body">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm opacity-70">{$t('statistics.balance-change')}</p>
							<p
								class="text-2xl font-bold {statistics.totals.difference >= 0
									? 'text-success'
									: 'text-error'}"
							>
								{statistics.totals.difference >= 0 ? '+' : ''}{formatCurrency(
									statistics.totals.difference
								)}€
							</p>
						</div>
						{#if statistics.totals.difference >= 0}
							<TrendingUp size={24} class="text-success" />
						{:else}
							<TrendingDown size={24} class="text-error" />
						{/if}
					</div>
				</div>
			</div>

			<!-- Total Transactions -->
			<div class="card bg-base-100 shadow-lg">
				<div class="card-body">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm opacity-70">{$t('statistics.total-transactions')}</p>
							<p class="text-2xl font-bold">{statistics.total_transactions}</p>
						</div>
						<DollarSign size={24} class="text-primary" />
					</div>
				</div>
			</div>

			<!-- Average Transaction -->
			<div class="card bg-base-100 shadow-lg">
				<div class="card-body">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm opacity-70">{$t('statistics.average-transaction')}</p>
							<p class="text-2xl font-bold">{formatCurrency(statistics.average_transaction)}€</p>
						</div>
						<BarChart3 size={24} class="text-info" />
					</div>
				</div>
			</div>

			<!-- Daily Average -->
			<div class="card bg-base-100 shadow-lg">
				<div class="card-body">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm opacity-70">{$t('statistics.daily-average')}</p>
							<p class="text-2xl font-bold">{formatCurrency(statistics.daily_average)}€</p>
						</div>
						<TrendingUp size={24} class="text-secondary" />
					</div>
				</div>
			</div>
		</div>

		<!-- Totals Summary -->
		<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
			<div class="card bg-base-100 shadow-lg">
				<div class="card-body text-center">
					<h3 class="text-success text-lg font-semibold">{$t('statistics.total-credits')}</h3>
					<p class="text-success text-3xl font-bold">
						+{formatCurrency(statistics.totals.credit)}€
					</p>
				</div>
			</div>

			<div class="card bg-base-100 shadow-lg">
				<div class="card-body text-center">
					<h3 class="text-error text-lg font-semibold">{$t('statistics.total-debits')}</h3>
					<p class="text-error text-3xl font-bold">-{formatCurrency(statistics.totals.debit)}€</p>
				</div>
			</div>

			<div class="card bg-base-100 shadow-lg">
				<div class="card-body text-center">
					<h3 class="text-lg font-semibold">{$t('statistics.largest-amounts')}</h3>
					<div class="space-y-1">
						<p class="text-sm">
							<span class="font-medium">{$t('statistics.largest-credit')}:</span>
							<span class="text-success">+{formatCurrency(statistics.largest_credit)}€</span>
						</p>
						<p class="text-sm">
							<span class="font-medium">{$t('statistics.largest-debit')}:</span>
							<span class="text-error">-{formatCurrency(statistics.largest_debit)}€</span>
						</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Category Breakdowns with Pie Charts -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<!-- Credit Categories -->
			{#if statistics.credit_category_breakdown && statistics.credit_category_breakdown.length > 0}
				<div class="card bg-base-100 shadow-lg">
					<div class="card-body">
						<h3 class="card-title text-success mb-4">
							{$t('statistics.credit-categories')}
						</h3>
						<div class="space-y-3">
							{#each statistics.credit_category_breakdown as category}
								<div class="flex items-center justify-between">
									<div class="flex-1">
										<div class="mb-1 flex items-center justify-between">
											<span class="font-medium">{category.name}</span>
											<span class="text-sm opacity-70">
												{category.count}
												{$t('statistics.transactions')} • +{formatCurrency(category.total)}€
											</span>
										</div>
										<div class="bg-base-200 h-2 w-full rounded-full">
											<div
												class="bg-success h-2 rounded-full"
												style="width: {Math.min(category.percentage, 100)}%"
											></div>
										</div>
										<span class="text-xs opacity-50">{category.percentage.toFixed(1)}%</span>
									</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			{/if}

			<!-- Debit Categories -->
			{#if statistics.debit_category_breakdown && statistics.debit_category_breakdown.length > 0}
				<div class="card bg-base-100 shadow-lg">
					<div class="card-body">
						<h3 class="card-title text-error mb-4">
							{$t('statistics.debit-categories')}
						</h3>
						<div class="space-y-3">
							{#each statistics.debit_category_breakdown as category}
								<div class="flex items-center justify-between">
									<div class="flex-1">
										<div class="mb-1 flex items-center justify-between">
											<span class="font-medium">{category.name}</span>
											<span class="text-sm opacity-70">
												{category.count}
												{$t('statistics.transactions')} • -{formatCurrency(category.total)}€
											</span>
										</div>
										<div class="bg-base-200 h-2 w-full rounded-full">
											<div
												class="bg-error h-2 rounded-full"
												style="width: {Math.min(category.percentage, 100)}%"
											></div>
										</div>
										<span class="text-xs opacity-50">{category.percentage.toFixed(1)}%</span>
									</div>
								</div>
							{/each}
						</div>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
