<script lang="ts">
	// "What changed this month" — one consistent metric (month-over-month). A context
	// line (total vs last month + peak month) sits above the categories that rose or
	// fell the most. Movers are computed server-side (CategoryMover, MoM); the context
	// line is a thin derivation of the totals[] aggregate.
	import { t } from '$lib/i18n';
	import { locale } from 'svelte-i18n';
	import { TrendingUp, TrendingDown } from 'lucide-svelte';
	import CategorySparkline from '$components/CategorySparkline.svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import type { CategoryMover, TrendMonth } from '$lib/types';

	let {
		movers,
		totals,
		months
	}: {
		movers: CategoryMover[];
		totals: number[];
		months: TrendMonth[];
	} = $props();

	let currentLocale = $derived($locale || 'pt');
	const monthLabel = (m: TrendMonth) =>
		new Date(m.year, m.month - 1, 1).toLocaleDateString(currentLocale, {
			month: 'short',
			year: '2-digit'
		});

	// Peak month + total month-over-month, ignoring leading-zero months (before any
	// activity). null when the window is all-zero; mom is null with <2 active months.
	let context = $derived.by(() => {
		const n = totals.length;
		const start = totals.findIndex((v) => v > 0);
		if (start === -1) return null;

		let hiIdx = start;
		for (let i = start; i < n; i++) {
			if (totals[i] > totals[hiIdx]) hiIdx = i;
		}

		let mom: number | null = null;
		if (n - start >= 2 && totals[n - 2] > 0) {
			mom = Math.round(((totals[n - 1] - totals[n - 2]) / totals[n - 2]) * 100);
		}
		return { hiIdx, mom };
	});

	// The backend already splits by direction; group for display. "new" reads as a rise.
	let risers = $derived(movers.filter((m) => m.direction === 'up' || m.direction === 'new'));
	let fallers = $derived(movers.filter((m) => m.direction === 'down'));

	// Euro change from last month to the latest month (the mover's ranking key).
	const deltaOf = (m: CategoryMover) =>
		m.totals[m.totals.length - 1] - m.totals[m.totals.length - 2];
</script>

<!-- One category row: name, sparkline, euro change + % (or "New"). -->
{#snippet moverRow(m: CategoryMover)}
	{@const d = deltaOf(m)}
	<div class="flex items-center gap-2">
		<span class="h-2.5 w-2.5 shrink-0 rounded-full" style="background-color: {m.color};"></span>
		<span class="min-w-0 flex-1 truncate text-sm">{m.name}</span>
		<CategorySparkline data={m.totals} color={m.color} />
		<span class="flex flex-col items-end leading-tight">
			<span class="text-sm font-semibold tabular-nums">
				{d > 0 ? '+' : ''}{formatCurrency(d, 0)}
			</span>
			<span class="text-[0.65rem] uppercase tracking-wide opacity-60">
				{#if m.direction === 'new'}
					{$t('trends.momentum-new')}
				{:else}
					{(m.pct ?? 0) > 0 ? '+' : ''}{m.pct}%
				{/if}
			</span>
		</span>
	</div>
{/snippet}

{#if context}
	<div class="mt-4 rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		<p class="text-sm font-semibold">{$t('trends.changed-title')}</p>
		<p class="mt-0.5 text-xs text-base-content/50">
			{#if context.mom !== null}
				{$t('trends.changed-total')}
				<span class="font-semibold">{context.mom > 0 ? '+' : ''}{context.mom}%</span>
				{$t('trends.vs-last-month')} ·
			{/if}
			{$t('trends.peak-month')}
			<span class="font-semibold">{monthLabel(months[context.hiIdx])}</span>
			<span class="opacity-60">·</span>
			<span class="font-semibold">{formatCurrency(totals[context.hiIdx])}</span>
		</p>

		{#if risers.length > 0 || fallers.length > 0}
			<div class="mt-3 grid gap-x-6 gap-y-4 sm:grid-cols-2">
				{#if risers.length > 0}
					<div>
						<p
							class="mb-1.5 flex items-center gap-1 text-xs font-semibold uppercase tracking-wide opacity-60"
						>
							<TrendingUp size={13} />
							{$t('trends.biggest-rises')}
						</p>
						<div class="flex flex-col gap-2">
							{#each risers as m (m.id)}
								{@render moverRow(m)}
							{/each}
						</div>
					</div>
				{/if}
				{#if fallers.length > 0}
					<div>
						<p
							class="mb-1.5 flex items-center gap-1 text-xs font-semibold uppercase tracking-wide opacity-60"
						>
							<TrendingDown size={13} />
							{$t('trends.biggest-drops')}
						</p>
						<div class="flex flex-col gap-2">
							{#each fallers as m (m.id)}
								{@render moverRow(m)}
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<p class="mt-3 text-sm text-base-content/50">{$t('trends.no-changes')}</p>
		{/if}
	</div>
{/if}
