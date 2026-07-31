<script lang="ts">
	// The three headline KPI tiles. All values are computed server-side / derived on
	// the page; this component only formats and lays them out. Figures are monospace
	// (the ledger identity) and the two time-series tiles carry a subtle trend spark.
	import { t } from '$lib/i18n';
	import { formatCurrency } from '$lib/utils/currency';
	import CategorySparkline from '$components/CategorySparkline.svelte';
	import type { TrendsType } from '$lib/types';

	let {
		windowTotal,
		monthlyAverage,
		topShare,
		topN,
		totals,
		type
	}: {
		windowTotal: number;
		monthlyAverage: number;
		topShare: number | null; // top-N categories as a share of the window (0..1)
		topN: number;
		totals: number[];
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
	let concLabel = $derived($t('trends.kpi-concentration', { values: { n: topN } }));
	let concValue = $derived(topShare === null ? '—' : `${Math.round(topShare * 100)}%`);
</script>

<div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-3">
	<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		{@render kpiLabel(totalLabel, totalTip)}
		<p class="mt-1 font-mono text-2xl font-semibold tabular-nums tracking-tight {accentClass}">
			{formatCurrency(windowTotal)}
		</p>
		<div class="mt-2 text-base-content/25">
			<CategorySparkline data={totals} color="currentColor" width={132} height={26} />
		</div>
	</div>

	<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		{@render kpiLabel($t('trends.kpi-average'), averageTip)}
		<p class="mt-1 font-mono text-2xl font-semibold tabular-nums tracking-tight">
			{formatCurrency(monthlyAverage)}
		</p>
		<div class="mt-2 text-base-content/25">
			<CategorySparkline data={totals} color="currentColor" width={132} height={26} />
		</div>
	</div>

	<div class="col-span-2 rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm sm:col-span-1">
		{@render kpiLabel(concLabel, $t('trends.kpi-concentration-tip'))}
		<p class="mt-1 font-mono text-2xl font-semibold tabular-nums tracking-tight">{concValue}</p>
	</div>
</div>

<!-- Tile label with an on-hover DaisyUI tooltip explaining the stat. -->
{#snippet kpiLabel(label: string, tip: string)}
	<span
		class="tooltip tooltip-bottom block w-fit cursor-help text-xs font-medium uppercase tracking-wide text-base-content/60 before:z-10 before:max-w-[12rem] before:whitespace-normal before:text-[0.7rem] before:normal-case before:shadow-lg before:content-[attr(data-tip)]"
		data-tip={tip}
	>
		{label}
	</span>
{/snippet}
