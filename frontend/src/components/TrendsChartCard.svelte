<script lang="ts">
	// The chart plus its toggleable legend. Owns the client-only interaction state
	// (which series are drawn, which root is drilled open) so the page stays a thin
	// orchestrator. All numbers come pre-computed from the backend.
	import { t } from '$lib/i18n';
	import TrendsChart from '$components/TrendsChart.svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import { getContrastTextClass } from '$lib/utils/categoryUtils';
	import { ChevronDown, ChevronRight } from 'lucide-svelte';
	import type { CategoryTrend, CategoryTrendsResponse, TrendsType } from '$lib/types';

	let { trends, type }: { trends: CategoryTrendsResponse; type: TrendsType } = $props();

	const TOTAL_COLOR = '#6366f1'; // must match TrendsChart's total line

	let selectedCategoryIds = $state<number[]>([]);
	let showTotal = $state(true);
	let expandedRootId = $state<number | null>(null);
	let showSmoothed = $state(false);

	// Flat lookup of every drawable series (roots + subcategories) by id.
	let seriesById = $derived.by(() => {
		const map = new Map<number, CategoryTrend>();
		for (const root of trends.categories) {
			map.set(root.id, root);
			for (const sub of root.subcategories ?? []) map.set(sub.id, sub);
		}
		return map;
	});

	// The picked categories, resolved to their full series. Stale ids (e.g. after
	// switching account) resolve to nothing and simply stop being drawn.
	let selectedCats = $derived(
		selectedCategoryIds.map((id) => seriesById.get(id)).filter((c): c is CategoryTrend => !!c)
	);
	// Shape the chart expects (raw lines); it derives its own smoothed companions.
	let selectedSeries = $derived(
		selectedCats.map((c) => ({ name: c.name, color: c.color, totals: c.totals }))
	);

	let hasSeries = $derived(showTotal || selectedSeries.length > 0);

	let monthsCount = $derived(trends.months.length || 1);
	let lastIdx = $derived(trends.months.length - 1);

	// --- Per-category stats. Denominators guarded so a zero window shows "—". ---

	// Share of income over the window (debit) or of total credits (credit mode). For
	// a spending category this reads as its % of income; for a savings category it
	// is the savings rate.
	function windowShare(cat: CategoryTrend): number | null {
		const denom = type === 'debit' ? trends.window_income : trends.window_total;
		return denom > 0 ? cat.total / denom : null;
	}

	// The category's share of the latest month's grand total.
	function monthShare(cat: CategoryTrend): number | null {
		if (lastIdx < 0) return null;
		const monthTotal = trends.totals[lastIdx] ?? 0;
		return monthTotal > 0 ? (cat.totals[lastIdx] ?? 0) / monthTotal : null;
	}

	const perMonth = (cat: CategoryTrend) => cat.total / monthsCount;

	// Latest month vs the previous one. 'new' = spend appeared with no prior month.
	function momOf(cat: CategoryTrend): {
		dir: 'up' | 'down' | 'flat' | 'new' | 'none';
		pct: number;
	} {
		const n = cat.totals.length;
		if (n < 2) return { dir: 'none', pct: 0 };
		const prev = cat.totals[n - 2];
		const cur = cat.totals[n - 1];
		if (prev === 0) return { dir: cur > 0 ? 'new' : 'none', pct: 0 };
		const p = ((cur - prev) / prev) * 100;
		return { dir: p > 0.5 ? 'up' : p < -0.5 ? 'down' : 'flat', pct: Math.round(p) };
	}

	// Concentration: top-N roots as a share of the window. Answers "where does it
	// go" before any click. Categories arrive already sorted biggest-first.
	let topN = $derived(Math.min(3, trends.categories.length));
	let concentration = $derived.by(() => {
		if (topN === 0 || trends.window_total <= 0) return null;
		const top = trends.categories.slice(0, topN).reduce((sum, c) => sum + c.total, 0);
		return top / trends.window_total;
	});

	const pct = (v: number | null) => (v === null ? '—' : `${Math.round(v * 100)}%`);

	let expandedRoot = $derived(
		expandedRootId !== null
			? (trends.categories.find((c) => c.id === expandedRootId) ?? null)
			: null
	);

	function toggleCategory(id: number) {
		selectedCategoryIds = selectedCategoryIds.includes(id)
			? selectedCategoryIds.filter((x) => x !== id)
			: [...selectedCategoryIds, id];
	}

	function toggleExpand(id: number) {
		expandedRootId = expandedRootId === id ? null : id;
	}
</script>

<!-- One toggleable series pill (used for roots and subcategories) -->
{#snippet seriesPill(id: number, name: string, color: string, catTotal: number)}
	{@const active = selectedCategoryIds.includes(id)}
	<button
		type="button"
		class="flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm font-medium transition-colors {active
			? `border-transparent shadow-sm ${getContrastTextClass(color)}`
			: 'border-base-300 text-base-content/60 hover:bg-base-200'}"
		style={active ? `background-color: ${color};` : ''}
		onclick={() => toggleCategory(id)}
	>
		<span class="h-2.5 w-2.5 rounded-full" style="background-color: {color};"></span>
		{name}
		<span class="text-xs {active ? 'opacity-80' : 'opacity-50'}">{formatCurrency(catTotal)}</span>
	</button>
{/snippet}

<div class="rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
	<!-- Concentration + smoothing: context before any click. -->
	<div class="mb-3 flex flex-wrap items-center justify-between gap-2">
		{#if concentration !== null}
			<p class="text-xs text-base-content/60">
				{$t(type === 'debit' ? 'trends.concentration-spending' : 'trends.concentration-income', {
					values: { n: topN, pct: pct(concentration) }
				})}
			</p>
		{:else}
			<span></span>
		{/if}
		<label class="flex cursor-pointer items-center gap-2 text-xs text-base-content/60">
			<input type="checkbox" class="toggle toggle-xs" bind:checked={showSmoothed} />
			{$t('trends.smoothing')}
		</label>
	</div>

	{#if hasSeries}
		<TrendsChart
			months={trends.months}
			total={trends.totals}
			series={selectedSeries}
			{showTotal}
			smooth={showSmoothed}
		/>
	{:else}
		<div class="flex h-80 items-center justify-center">
			<p class="text-base-content/50">{$t('trends.pick-series')}</p>
		</div>
	{/if}

	<!-- Per-category detail for the selection. Hover the chart for every month's %. -->
	{#if selectedCats.length > 0}
		<div class="mt-3 overflow-x-auto rounded-lg bg-base-200/50">
			<div class="min-w-[24rem] px-3 py-2">
				<div
					class="grid grid-cols-[minmax(0,1fr)_4rem_4rem_4rem_4.5rem] items-center gap-2 pb-1 text-[0.65rem] uppercase tracking-wide opacity-50"
				>
					<span></span>
					<span class="text-right">{$t('trends.col-income')}</span>
					<span class="text-right">{$t('trends.col-this-month')}</span>
					<span class="text-right">{$t('trends.col-mom')}</span>
					<span class="text-right">{$t('trends.col-permonth')}</span>
				</div>
				{#each selectedCats as cat (cat.id)}
					{@const m = momOf(cat)}
					<div
						class="grid grid-cols-[minmax(0,1fr)_4rem_4rem_4rem_4.5rem] items-center gap-2 py-0.5 text-sm"
					>
						<span class="flex min-w-0 items-center gap-1.5">
							<span class="h-2.5 w-2.5 shrink-0 rounded-full" style="background-color: {cat.color};"
							></span>
							<span class="truncate">{cat.name}</span>
						</span>
						<span class="text-right font-semibold">{pct(windowShare(cat))}</span>
						<span class="text-right">{pct(monthShare(cat))}</span>
						<span class="text-right tabular-nums">
							{#if m.dir === 'new'}
								{$t('trends.momentum-new')}
							{:else if m.dir === 'none'}
								—
							{:else}
								{m.pct > 0 ? '+' : ''}{m.pct}%
							{/if}
						</span>
						<span class="text-right tabular-nums">{formatCurrency(perMonth(cat), 0)}</span>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Toggleable legend: Total + roots (with drill-down into subcategories) -->
	<div class="mt-4 flex flex-col gap-3">
		<div class="flex flex-wrap items-center gap-2">
			<button
				type="button"
				class="flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm font-medium transition-colors {showTotal
					? 'border-transparent text-white shadow-sm'
					: 'border-base-300 text-base-content/60 hover:bg-base-200'}"
				style={showTotal ? `background-color: ${TOTAL_COLOR};` : ''}
				onclick={() => (showTotal = !showTotal)}
			>
				<span class="h-2.5 w-2.5 rounded-full" style="background-color: {TOTAL_COLOR};"></span>
				{$t('trends.total')}
			</button>

			{#each trends.categories as root (root.id)}
				<div class="flex items-center">
					{@render seriesPill(root.id, root.name, root.color, root.total)}
					{#if root.subcategories && root.subcategories.length > 0}
						<button
							type="button"
							class="ml-0.5 flex h-7 w-7 items-center justify-center rounded-full text-base-content/50 transition-colors hover:bg-base-200 {expandedRootId ===
							root.id
								? 'text-base-content'
								: ''}"
							onclick={() => toggleExpand(root.id)}
							aria-label={$t('trends.subcategories')}
							aria-expanded={expandedRootId === root.id}
						>
							{#if expandedRootId === root.id}
								<ChevronDown size={16} />
							{:else}
								<ChevronRight size={16} />
							{/if}
						</button>
					{/if}
				</div>
			{/each}
		</div>

		<!-- Drill-down: the expanded root's subcategories -->
		{#if expandedRoot && expandedRoot.subcategories}
			<div class="rounded-lg border border-base-300 bg-base-200/40 p-3">
				<p class="mb-2 flex items-center gap-1.5 text-xs font-semibold opacity-70">
					<span class="h-2.5 w-2.5 rounded-full" style="background-color: {expandedRoot.color};"
					></span>
					{expandedRoot.name} · {$t('trends.subcategories')}
				</p>
				<div class="flex flex-wrap gap-2">
					{#each expandedRoot.subcategories as sub (sub.id)}
						{@render seriesPill(sub.id, sub.name, sub.color, sub.total)}
					{/each}
				</div>
			</div>
		{/if}
	</div>
</div>
