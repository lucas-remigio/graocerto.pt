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
</script>

<div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-3">
	<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		<p class="text-xs uppercase tracking-wide opacity-60">{$t('trends.kpi-total')}</p>
		<p class="text-2xl font-bold {accentClass}">{formatCurrency(windowTotal)}</p>
	</div>
	<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		<p class="text-xs uppercase tracking-wide opacity-60">{$t('trends.kpi-average')}</p>
		<p class="text-2xl font-bold">{formatCurrency(monthlyAverage)}</p>
	</div>
	<div class="col-span-2 rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm sm:col-span-1">
		<p class="text-xs uppercase tracking-wide opacity-60">{$t('trends.kpi-categories')}</p>
		<p class="text-2xl font-bold">{categoriesCount}</p>
	</div>
</div>
