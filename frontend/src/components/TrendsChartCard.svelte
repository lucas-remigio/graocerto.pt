<script lang="ts">
	// The chart plus its toggleable legend. Owns the client-only interaction state
	// (which series are drawn, which root is drilled open) so the page stays a thin
	// orchestrator. All numbers come pre-computed from the backend.
	import { t } from '$lib/i18n';
	import TrendsChart from '$components/TrendsChart.svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import { getContrastTextClass } from '$lib/utils/categoryUtils';
	import { ChevronDown, ChevronRight } from 'lucide-svelte';
	import type { CategoryTrend, CategoryTrendsResponse } from '$lib/types';

	let { trends }: { trends: CategoryTrendsResponse } = $props();

	const TOTAL_COLOR = '#6366f1'; // must match TrendsChart's total line

	let selectedCategoryIds = $state<number[]>([]);
	let showTotal = $state(true);
	let expandedRootId = $state<number | null>(null);

	// Flat lookup of every drawable series (roots + subcategories) by id.
	let seriesById = $derived.by(() => {
		const map = new Map<number, CategoryTrend>();
		for (const root of trends.categories) {
			map.set(root.id, root);
			for (const sub of root.subcategories ?? []) map.set(sub.id, sub);
		}
		return map;
	});

	// Stale ids (e.g. after switching account) simply resolve to nothing here, so
	// no pruning is needed — they just stop being drawn.
	let selectedSeries = $derived(
		selectedCategoryIds
			.map((id) => seriesById.get(id))
			.filter((s): s is CategoryTrend => !!s)
			.map((s) => ({ name: s.name, color: s.color, totals: s.totals }))
	);

	let hasSeries = $derived(showTotal || selectedSeries.length > 0);

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
	{#if hasSeries}
		<TrendsChart months={trends.months} total={trends.totals} series={selectedSeries} {showTotal} />
	{:else}
		<div class="flex h-80 items-center justify-center">
			<p class="text-base-content/50">{$t('trends.pick-series')}</p>
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
					<span class="h-2.5 w-2.5 rounded-full" style="background-color: {expandedRoot.color};"></span>
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
