<script lang="ts">
	import type { CategoryStatistic } from '$lib/types';
	import { t } from '$lib/i18n';
	import { formatCurrency } from '$lib/utils/currency';

	export let stat: CategoryStatistic;
	export let isCredit: boolean;
	export let size: 'sm' | 'md' = 'md';
	export let showStatus: boolean = true;

	const fmt = (v: number | null) => (v === null ? '—' : formatCurrency(v));

	$: pct = Math.round(stat.budget_percentage);
	$: fillPct = Math.min(Math.max(pct, 0), 100);
	$: overBudget = stat.budget != null && stat.total > stat.budget;
	$: fillClass = overBudget ? (isCredit ? 'bg-emerald-600' : 'bg-rose-500') : '';
	$: barStyle = !overBudget && stat.color ? `background:${stat.color};` : '';
	$: pctInside = fillPct >= 18;
	$: barHeight = size === 'sm' ? 'h-3' : 'h-4';
	$: textSize = size === 'sm' ? 'text-xs' : 'text-xs';
	$: statusSize = size === 'sm' ? 'text-xs' : 'text-sm';
</script>

<div class="w-full">
	<!-- Progress Bar -->
	<div
		class="border-base-300 bg-base-200 {barHeight} relative w-full overflow-hidden rounded-lg border"
	>
		<div class="h-full transition-all {fillClass}" style="width: {fillPct}%; {barStyle}"></div>
		<div class="absolute right-2 top-1/2 -translate-y-1/2 {textSize} font-semibold">
			<span class={pctInside ? 'text-white' : 'text-base-content/80'}>
				{pct}%
			</span>
		</div>
	</div>

	<!-- Status Row -->
	{#if showStatus}
		<div class="mt-2 flex items-center {statusSize}">
			{#if overBudget}
				<div class="flex items-center gap-2">
					<span
						class="inline-block h-2 w-2 rounded-full {isCredit ? 'bg-emerald-600' : 'bg-rose-500'}"
					></span>
					<strong class={isCredit ? 'text-emerald-600' : 'text-rose-600'}>
						{isCredit ? $t('statistics.over-target') : $t('statistics.over-budget')}
					</strong>
				</div>
			{:else}
				<span class="text-base-content/80">{$t('statistics.within-budget')}</span>
			{/if}
			<div class="ml-auto whitespace-nowrap font-medium text-base-content/60">
				{fmt(stat.total)}
			</div>
		</div>
	{/if}
</div>
