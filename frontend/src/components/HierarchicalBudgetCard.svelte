<script lang="ts">
	import type { CategoryStatistic } from '$lib/types';
	import { ChevronDown, ChevronRight } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import BudgetProgressBar from './BudgetProgressBar.svelte';
	import { formatCurrency } from '$lib/utils/currency';

	export let category: CategoryStatistic;
	export let isCredit: boolean;

	let expanded = false;
	const fmt = (v: number | null) => (v === null ? '—' : formatCurrency(v));

	$: hasSubcategories = !!(category.subcategories && category.subcategories.length > 0);

	// Calculate direct spending
	$: directTotal =
		hasSubcategories && category.subcategories
			? category.total - category.subcategories.reduce((sum, sub) => sum + sub.total, 0)
			: 0;

	$: directCount =
		hasSubcategories && category.subcategories
			? category.count - category.subcategories.reduce((sum, sub) => sum + sub.count, 0)
			: 0;
</script>

<article class="card border bg-base-100 shadow-sm">
	<!-- Parent Header -->
	<header class="flex items-center gap-3 p-4">
		<!-- Expand button -->
		{#if hasSubcategories}
			<button
				class="btn btn-circle btn-ghost btn-xs flex-shrink-0"
				on:click={() => (expanded = !expanded)}
				aria-label={expanded ? 'Collapse' : 'Expand'}
			>
				{#if expanded}
					<ChevronDown size={16} />
				{:else}
					<ChevronRight size={16} />
				{/if}
			</button>
		{:else}
			<div class="w-8 flex-shrink-0"></div>
		{/if}

		<!-- Name & Budget -->
		<div class="flex min-w-0 flex-1 items-center justify-between gap-4">
			<h4 class="truncate text-sm font-semibold">{category.name}</h4>
			<div class="text-right text-sm">
				{#if category.budget != null}
					<strong class="whitespace-nowrap">{fmt(category.budget)}</strong>
				{:else}
					<span class="text-base-content/50">{$t('statistics.no-budget')}</span>
				{/if}
			</div>
		</div>
	</header>

	<!-- Parent Progress Bar -->
	{#if category.budget}
		<div class="px-4 pb-4">
			<BudgetProgressBar stat={category} {isCredit} />
		</div>
	{:else}
		<div class="px-4 pb-4 text-sm text-base-content/60">
			{fmt(category.total)}
		</div>
	{/if}

	<!-- Subcategories (Collapsible) -->
	{#if hasSubcategories && expanded && category.subcategories}
		<div class="space-y-3 border-t border-base-300 bg-base-200/30 px-4 py-4">
			<!-- Direct Spending -->
			{#if directCount > 0}
				<div class="rounded-lg bg-base-100 p-3">
					<div class="flex items-center justify-between">
						<span class="text-xs opacity-70">{category.name} (Direto)</span>
						<span class="text-xs font-semibold {isCredit ? 'text-success' : 'text-error'}">
							{fmt(directTotal)}
						</span>
					</div>
				</div>
			{/if}

			<!-- Subcategories -->
			{#each category.subcategories as sub}
				<div class="rounded-lg bg-base-100 p-3">
					<!-- Sub Header -->
					<div class="mb-2 flex items-center justify-between gap-4">
						<span class="text-xs font-medium">{sub.name}</span>
						<div class="text-right text-xs">
							{#if sub.budget != null}
								<strong>{fmt(sub.budget)}</strong>
							{:else}
								<span class="opacity-50">{$t('statistics.no-budget')}</span>
							{/if}
						</div>
					</div>

					<!-- Sub Progress Bar -->
					{#if sub.budget}
						<BudgetProgressBar stat={sub} {isCredit} size="sm" showStatus={false} />
						<div class="mt-1.5 flex items-center justify-between text-xs">
							<span class="opacity-70">
								{sub.total > sub.budget
									? isCredit
										? $t('statistics.over-target')
										: $t('statistics.over-budget')
									: $t('statistics.within-budget')}
							</span>
							<span class="font-medium text-base-content/60">{fmt(sub.total)}</span>
						</div>
					{:else}
						<div class="text-xs opacity-60">{fmt(sub.total)}</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</article>
