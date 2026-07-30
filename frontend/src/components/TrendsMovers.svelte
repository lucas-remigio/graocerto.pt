<script lang="ts">
	// "What's changing" — context chips (best/worst month, vs last month) plus
	// category momentum. Momentum data is computed server-side (CategoryMover); the
	// chips are thin Math.max/min derivations of the totals[] aggregate.
	import { t } from '$lib/i18n';
	import { locale } from 'svelte-i18n';
	import { TrendingUp, TrendingDown } from 'lucide-svelte';
	import CategorySparkline from '$components/CategorySparkline.svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import type { CategoryMover, TrendMonth, TrendsType } from '$lib/types';

	let {
		movers,
		totals,
		months,
		type
	}: {
		movers: CategoryMover[];
		totals: number[];
		months: TrendMonth[];
		type: TrendsType;
	} = $props();

	let currentLocale = $derived($locale || 'pt');
	const monthLabel = (m: TrendMonth) =>
		new Date(m.year, m.month - 1, 1).toLocaleDateString(currentLocale, {
			month: 'short',
			year: '2-digit'
		});

	// Best/worst month + month-over-month, ignoring leading-zero months (before any
	// activity). null when the window is all-zero; mom is null with <2 active months.
	let chips = $derived.by(() => {
		const n = totals.length;
		const start = totals.findIndex((v) => v > 0);
		if (start === -1) return null;

		let hiIdx = start;
		let loIdx = start;
		for (let i = start; i < n; i++) {
			if (totals[i] > totals[hiIdx]) hiIdx = i;
			if (totals[i] < totals[loIdx]) loIdx = i;
		}

		let mom: number | null = null;
		if (n - start >= 2 && totals[n - 2] > 0) {
			mom = ((totals[n - 1] - totals[n - 2]) / totals[n - 2]) * 100;
		}
		return { hiIdx, loIdx, mom };
	});

	let accentClass = $derived(type === 'debit' ? 'text-error' : 'text-success');
</script>

{#if chips || movers.length > 0}
	<div class="mt-4 rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		<div class="mb-3">
			<p class="text-sm font-semibold">{$t('trends.whats-changing')}</p>
			<p class="text-xs text-base-content/50">{$t('trends.momentum-caption')}</p>
		</div>

		{#if chips}
			<div class="mb-4 flex flex-wrap gap-2">
				<span class="rounded-full bg-base-200 px-2.5 py-1 text-xs">
					<span class="opacity-60">{$t('trends.highest-month')}</span>
					<span class="font-semibold">{monthLabel(months[chips.hiIdx])}</span>
					<span class="opacity-50">·</span>
					<span class="font-semibold {accentClass}">{formatCurrency(totals[chips.hiIdx])}</span>
				</span>
				<span class="rounded-full bg-base-200 px-2.5 py-1 text-xs">
					<span class="opacity-60">{$t('trends.lowest-month')}</span>
					<span class="font-semibold">{monthLabel(months[chips.loIdx])}</span>
					<span class="opacity-50">·</span>
					<span class="font-semibold">{formatCurrency(totals[chips.loIdx])}</span>
				</span>
				{#if chips.mom !== null}
					<span class="flex items-center gap-1 rounded-full bg-base-200 px-2.5 py-1 text-xs">
						<span class="opacity-60">{$t('trends.vs-last-month')}</span>
						{#if chips.mom >= 0}
							<TrendingUp size={13} />
							<span class="font-semibold">+{Math.round(chips.mom)}%</span>
						{:else}
							<TrendingDown size={13} />
							<span class="font-semibold">{Math.round(chips.mom)}%</span>
						{/if}
					</span>
				{/if}
			</div>
		{/if}

		{#if movers.length > 0}
			<div class="grid gap-x-6 gap-y-3 sm:grid-cols-2">
				{#each movers as m (m.id)}
					<div class="flex items-center gap-3">
						<span class="h-2.5 w-2.5 shrink-0 rounded-full" style="background-color: {m.color};"
						></span>
						<span class="min-w-0 flex-1 truncate text-sm">{m.name}</span>
						<CategorySparkline data={m.totals} color={m.color} />
						<span
							class="tooltip tooltip-left flex cursor-help items-center gap-1 rounded-full bg-base-200 px-2 py-0.5 text-xs font-semibold text-base-content/80 before:z-10 before:max-w-[12rem] before:whitespace-normal before:text-[0.7rem] before:content-[attr(data-tip)]"
							data-tip={$t('trends.momentum-tip')}
						>
							{#if m.direction === 'new'}
								{$t('trends.momentum-new')}
							{:else if m.direction === 'up'}
								<TrendingUp size={13} />
								+{m.pct}%
							{:else}
								<TrendingDown size={13} />
								{m.pct}%
							{/if}
						</span>
					</div>
				{/each}
			</div>
		{/if}
	</div>
{/if}
