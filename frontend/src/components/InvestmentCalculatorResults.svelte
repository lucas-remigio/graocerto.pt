<script lang="ts">
	import { t } from '$lib/i18n';
	import InvestmentChart from './InvestmentChart.svelte';
	import type { InvestmentCalculatorResponse } from '$lib/types';
	import { TrendingUp } from 'lucide-svelte';

	export let results: InvestmentCalculatorResponse | null;

	function formatCurrency(value: number): string {
		return new Intl.NumberFormat('pt-PT', {
			style: 'currency',
			currency: 'EUR'
		}).format(value);
	}
</script>

<div class="card bg-base-100 shadow-lg">
	<div class="card-body">
		<h3 class="card-title text-success mb-4 justify-center">
			{$t('investment-calculator.results.title')}
		</h3>

		<!-- Summary Stats -->
		<div class="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
			<div class="stat bg-primary/10 rounded-lg">
				<div class="stat-title text-primary">
					{$t('investment-calculator.results.total-investment')}
				</div>
				<div class="stat-value text-primary text-lg">
					{formatCurrency(results?.total_investment || 0)}
				</div>
			</div>
			<div class="stat bg-success/10 rounded-lg">
				<div class="stat-title text-success">
					{$t('investment-calculator.results.total-return')}
				</div>
				<div class="stat-value text-success text-lg">
					{formatCurrency(results?.total_return || 0)}
				</div>
			</div>
			<div class="stat bg-accent/10 rounded-lg">
				<div class="stat-title text-accent">
					{$t('investment-calculator.results.total-value')}
				</div>
				<div class="stat-value text-accent text-lg">
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
				class="border-base-300 bg-base-50 flex h-64 items-center justify-center rounded-lg border-2 border-dashed"
			>
				<div class="text-base-content/60 text-center">
					<TrendingUp size={48} class="text-base-content/40 mx-auto mb-2" />
					<p class="text-sm">{$t('investment-calculator.chart.placeholder')}</p>
				</div>
			</div>
		{/if}
	</div>
</div>
