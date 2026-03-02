<script lang="ts">
	import { t } from '$lib/i18n';
	import InvestmentChart from './InvestmentChart.svelte';
	import type { InvestmentCalculatorResponse } from '$lib/types';
	import { TrendingUp } from 'lucide-svelte';

	import { formatCurrency } from '$lib/utils/currency';

	export let results: InvestmentCalculatorResponse | null;
</script>

<div class="bg-base-100">
	<div class="p-6">
		<h3 class="mb-4 text-center text-lg font-semibold text-success">
			{$t('investment-calculator.results.title')}
		</h3>

		<!-- Summary Stats -->
		<div class="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
			<div class="stat rounded-lg bg-primary/10">
				<div class="stat-title text-primary">
					{$t('investment-calculator.results.total-investment')}
				</div>
				<div class="stat-value text-lg text-primary">
					{formatCurrency(results?.total_investment || 0)}
				</div>
			</div>
			<div class="stat rounded-lg bg-success/10">
				<div class="stat-title text-success">
					{$t('investment-calculator.results.total-return')}
				</div>
				<div class="stat-value text-lg text-success">
					{formatCurrency(results?.total_return || 0)}
				</div>
			</div>
			<div class="stat rounded-lg bg-accent/10">
				<div class="stat-title text-accent">
					{$t('investment-calculator.results.total-value')}
				</div>
				<div class="stat-value text-lg text-accent">
					{formatCurrency(results?.total_value || 0)}
				</div>
			</div>
		</div>

		<!-- Investment Growth Chart -->
		{#if results?.yearly_breakdown && results.yearly_breakdown.length > 0}
			<div>
				<InvestmentChart yearlyBreakdown={results.yearly_breakdown} />
			</div>
		{:else}
			<!-- Placeholder chart area -->
			<div
				class="bg-base-50 flex h-64 items-center justify-center rounded-lg border-2 border-dashed border-base-300"
			>
				<div class="text-center text-base-content/60">
					<TrendingUp size={48} class="mx-auto mb-2 text-base-content/40" />
					<p class="text-sm">{$t('investment-calculator.chart.placeholder')}</p>
				</div>
			</div>
		{/if}
	</div>
</div>
