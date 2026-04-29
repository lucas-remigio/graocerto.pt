<!-- src/components/PieChart.svelte -->
<script lang="ts">
	import type { CategoryStatistic } from '$lib/types';
	import { t } from 'svelte-i18n';
	import { formatCurrency } from '$lib/utils/currency';

	// Accept a single category stat
	export let category: CategoryStatistic;
	export let isCredit: boolean;

	// Safe helpers
	const fmt = (v: number | null) => (v === null ? '—' : formatCurrency(v));

	$: spent = Math.abs(category?.total ?? 0);
	$: budget = category?.budget ?? null;

	// Use backend-calculated percentage
	$: pct = Math.round(category?.budget_percentage ?? 0);
	// fill width clamp to 100 for bar
	$: fillPct = Math.min(Math.max(pct, 0), 100);

	$: overBudget = budget != null && spent > budget;
	$: remaining = budget != null ? budget - spent : null;

	// Tailwind classes for bar when over / under
	$: fillClass = overBudget ? (isCredit ? 'bg-emerald-600' : 'bg-rose-500') : ''; // when not over, we use inline category color via barStyle

	// inline style when not overBudget and category has hex color
	$: barStyle = !overBudget && category?.color ? `background:${category.color};` : '';

	// decide if percentage text should be inside the filled area (enough fill)
	$: pctInside = fillPct >= 18;
</script>

<article class="card border bg-base-100 p-4 shadow-sm">
	<header class="flex items-center justify-between gap-4">
		<h4 class="truncate text-sm font-semibold leading-tight">{category?.name ?? 'Unnamed'}</h4>
		<div class="text-right text-sm">
			{#if budget != null}
				<strong class="whitespace-nowrap">{fmt(budget)}</strong>
			{:else}
				<span class="text-base-content/50">{$t('statistics.no-budget')}</span>
			{/if}
		</div>
	</header>

	<!-- Progress bar -->
	<div aria-hidden="false" aria-label="Budget progress" class="mt-3">
		<div class="relative h-4 w-full overflow-hidden rounded-lg border border-base-300 bg-base-200">
			<!-- filled portion -->
			<div class="h-full transition-all {fillClass}" style="width: {fillPct}%; {barStyle}"></div>

			<!-- percentage badge positioned at the right end of the bar -->
			<div
				class="absolute right-2 top-1/2 -translate-y-1/2 rounded px-2 py-0.5 text-xs font-semibold"
				style="pointer-events: none;"
			>
				<span class={pctInside ? 'text-white' : 'text-base-content/80'}>
					{pct}%
				</span>
			</div>
		</div>

		<!-- simple status row: left = label/dot, right = spent (aligned) -->
		<div class="mt-2 flex items-center text-sm">
			{#if overBudget}
				<div class="flex items-center gap-2">
					<span
						class={'inline-block h-2 w-2 rounded-full ' +
							(isCredit ? 'bg-emerald-600' : 'bg-rose-500')}
					></span>
					<strong class={isCredit ? 'text-emerald-600' : 'text-rose-600'}>
						{isCredit ? $t('statistics.over-target') : $t('statistics.over-budget')}
					</strong>
				</div>
			{:else}
				<!-- within budget-->
				<div class="flex items-center gap-2">
					<span class="text-base-content/80">
						{$t('statistics.within-budget')}
					</span>
				</div>
			{/if}

			<!-- spent aligned to the right -->
			<div class="ml-auto flex flex-col items-end whitespace-nowrap text-right text-sm font-medium">
				<div class="text-base-content/60">
					{fmt(spent)}
				</div>
				{#if remaining !== null && !overBudget}
					<div class="text-[10px] uppercase tracking-wider text-base-content/40">
						{$t('statistics.remaining')}: {fmt(remaining)}
					</div>
				{/if}
			</div>
		</div>
	</div>
</article>
