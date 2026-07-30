<script lang="ts">
	// The three headline KPI tiles. All values are computed server-side; this
	// component only formats and lays them out.
	import { t } from '$lib/i18n';
	import { formatCurrency } from '$lib/utils/currency';
	import type { TrendsType } from '$lib/types';

	let {
		windowTotal,
		monthlyAverage,
		categoriesCount,
		type
	}: {
		windowTotal: number;
		monthlyAverage: number;
		categoriesCount: number;
		type: TrendsType;
	} = $props();

	let accentClass = $derived(type === 'debit' ? 'text-error' : 'text-success');

	// Mode-aware so "Total" is never ambiguous about spending vs income.
	let totalLabel = $derived(
		type === 'debit' ? $t('trends.kpi-total-spending') : $t('trends.kpi-total-income')
	);
	let totalTip = $derived(
		type === 'debit' ? $t('trends.kpi-total-tip-spending') : $t('trends.kpi-total-tip-income')
	);
	let averageTip = $derived(
		type === 'debit' ? $t('trends.kpi-average-tip-spending') : $t('trends.kpi-average-tip-income')
	);
</script>

<div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-3">
	<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		{@render kpiLabel(totalLabel, totalTip)}
		<p class="text-2xl font-bold {accentClass}">{formatCurrency(windowTotal)}</p>
	</div>
	<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		{@render kpiLabel($t('trends.kpi-average'), averageTip)}
		<p class="text-2xl font-bold">{formatCurrency(monthlyAverage)}</p>
	</div>
	<div class="col-span-2 rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm sm:col-span-1">
		{@render kpiLabel($t('trends.kpi-categories'), $t('trends.kpi-categories-tip'))}
		<p class="text-2xl font-bold">{categoriesCount}</p>
	</div>
</div>

<!-- Tile label with an on-hover DaisyUI tooltip explaining the stat. -->
{#snippet kpiLabel(label: string, tip: string)}
	<span
		class="tooltip tooltip-bottom block w-fit cursor-help text-xs uppercase tracking-wide text-base-content/60 underline decoration-dotted underline-offset-2 before:z-10 before:max-w-[12rem] before:whitespace-normal before:text-[0.7rem] before:normal-case before:shadow-lg before:content-[attr(data-tip)]"
		data-tip={tip}
	>
		{label}
	</span>
{/snippet}
