<!-- src/components/PieChart.svelte -->
<script lang="ts">
	import type { CategoryStatistic } from '$lib/types';
	import { locale } from 'svelte-i18n';

	// Accept a single category stat
	export let category: CategoryStatistic;
	export let isCredit: boolean;

	$: currentLocale = $locale || 'pt-PT';

	// Safe helpers
	const fmt = (v: number | null) => {
		if (v === null) return '—';
		return new Intl.NumberFormat(currentLocale, { maximumFractionDigits: 0 }).format(v) + '\u00A0€';
	};

	$: spent = Math.abs(category?.total ?? 0);
	$: budget = category?.budget ?? null;

	// percentage relative to budget (can be > 100)
	$: rawPct = budget && budget > 0 ? (spent / budget) * 100 : 0;
	$: pct = Math.round(rawPct);
	// fill width clamp to 100 for bar
	$: fillPct = Math.min(Math.max(Math.round(rawPct), 0), 100);

	// colors
	// when over budget use bright red, otherwise use category color or green
	$: overBudget = budget != null && spent > budget;
	$: fillColor = overBudget ? '#ff1744' : (category?.color ?? '#22c55e');
</script>

<article class="card bg-base-100 border p-4 shadow-sm">
	<header class="mb-2 flex items-start justify-between gap-4">
		<div>
			<h4 class="text-sm font-semibold leading-tight">{category?.name ?? 'Unnamed'}</h4>
			<p class="text-base-content/60 mt-1 text-xs">
				{#if budget != null}
					Budget: <strong>{fmt(budget)}</strong>
				{:else}
					<span class="text-base-content/50">No budget set</span>
				{/if}
			</p>
		</div>
		<div class="text-right">
			<p class="text-sm font-medium">
				{fmt(spent)}
			</p>
			<p class="text-base-content/60 text-xs">
				{#if budget != null}
					{pct}%
				{:else}
					—
				{/if}
			</p>
		</div>
	</header>

	<!-- Progress bar -->
	<div aria-hidden="false" aria-label="Budget progress" class="mt-3">
		<div class="bg-base-200 border-base-300 h-4 w-full overflow-hidden rounded-lg border">
			<div class="h-full transition-all" style="width: {fillPct}%; background: {fillColor};"></div>
		</div>

		<!-- Over budget indicator -->
		{#if overBudget}
			<div class="mt-2 flex items-center gap-2 text-sm text-red-600">
				<!-- small red pill -->
				<span class="inline-block h-2 w-2 rounded-full bg-[#ff1744]"></span>
				<strong>Over budget</strong>
				<span class="text-base-content/60">({fmt(spent)} / {fmt(budget)} — {pct}%)</span>
			</div>
		{:else if budget != null}
			<div class="text-base-content/60 mt-2 text-sm">
				{fmt(spent)} / {fmt(budget)} ({pct}%)
			</div>
		{/if}
	</div>
</article>

<style>
</style>
