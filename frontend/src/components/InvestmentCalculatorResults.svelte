<script lang="ts">
	import { t } from '$lib/i18n';
	import InvestmentChart from './InvestmentChart.svelte';
	import type { InvestmentCalculatorResponse } from '$lib/types';

	export let results: InvestmentCalculatorResponse;

	function formatCurrency(value: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
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
					{formatCurrency(results.total_investment)}
				</div>
			</div>
			<div class="stat bg-success/10 rounded-lg">
				<div class="stat-title text-success">
					{$t('investment-calculator.results.total-return')}
				</div>
				<div class="stat-value text-success text-lg">
					{formatCurrency(results.total_return)}
				</div>
			</div>
			<div class="stat bg-accent/10 rounded-lg">
				<div class="stat-title text-accent">
					{$t('investment-calculator.results.total-value')}
				</div>
				<div class="stat-value text-accent text-lg">
					{formatCurrency(results.total_value)}
				</div>
			</div>
		</div>

		<!-- Investment Growth Chart -->
		{#if results.yearly_breakdown && results.yearly_breakdown.length > 0}
			<div>
				<InvestmentChart yearlyBreakdown={results.yearly_breakdown} />
			</div>
		{/if}
	</div>
</div>
