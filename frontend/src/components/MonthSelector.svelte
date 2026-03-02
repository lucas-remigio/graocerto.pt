<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Calendar, ChevronRight } from 'lucide-svelte';
	import type { MonthYear } from '$lib/types';
	import { locale, t } from '$lib/i18n';

	export let availableMonths: MonthYear[] = [];
	export let selectedMonth: number | null = null;
	export let selectedYear: number | null = null;

	const dispatch = createEventDispatcher<{
		monthSelect: { month: number | null; year: number | null };
	}>();

	const currentMonth = new Date().getMonth() + 1; // 1-12 (1 = January)
	const currentYear = new Date().getFullYear();

	let expandedYears = new Set<number>();

	function handleMonthSelect(month: number | null, year: number | null) {
		dispatch('monthSelect', { month, year });
	}

	function toggleYear(year: number) {
		if (expandedYears.has(year)) {
			expandedYears.delete(year);
		} else {
			expandedYears.add(year);
		}
		expandedYears = expandedYears; // trigger reactivity
	}

	function formatDate(month: number, year: number): string {
		const date = new Date(year, month - 1);
		return date.toLocaleString(currentLocale, { month: 'long', year: 'numeric' });
	}

	function isCurrentMonth(monthData: MonthYear): boolean {
		return monthData.month === currentMonth && monthData.year === currentYear;
	}

	$: monthsByYear = availableMonths.reduce(
		(acc, m) => {
			if (!acc[m.year]) acc[m.year] = [];
			acc[m.year].push(m);
			return acc;
		},
		{} as Record<number, MonthYear[]>
	);

	$: sortedYears = Object.keys(monthsByYear)
		.map(Number)
		.sort((a, b) => b - a);

	// Auto-expand the year of the selected month if it's a past year
	$: if (selectedYear !== null && selectedYear < currentYear) {
		expandedYears = new Set([...expandedYears, selectedYear]);
	}

	$: currentLocale = $locale || 'pt';
</script>

<div class="mb-2">
	<div class="flex items-center gap-2 overflow-x-auto pb-2">
		<button
			class="btn btn-circle btn-sm {selectedMonth === null && selectedYear === null
				? 'btn-primary text-base-100'
				: 'btn-ghost'} flex-shrink-0"
			on:click={() => handleMonthSelect(null, null)}
			title={$t('months.show-all-transactions')}
		>
			<Calendar size={20} />
		</button>

		{#each sortedYears as year}
			{#if year < currentYear}
				<!-- Past year: collapsible -->
				<button
					class="btn btn-sm {selectedYear === year
						? 'btn-primary text-base-100'
						: 'btn-ghost'} flex-shrink-0 gap-1 whitespace-nowrap"
					on:click={() => toggleYear(year)}
					title={String(year)}
				>
					{year}
					<ChevronRight
						size={14}
						class="transition-transform duration-200 {expandedYears.has(year) ? 'rotate-90' : ''}"
					/>
				</button>
				{#if expandedYears.has(year)}
					{#each monthsByYear[year] as monthData}
						<button
							class="btn btn-sm {selectedMonth === monthData.month &&
							selectedYear === monthData.year
								? 'btn-primary text-base-100'
								: 'btn-ghost'} flex-shrink-0 whitespace-nowrap"
							aria-label={formatDate(monthData.month, monthData.year)}
							on:click={() => handleMonthSelect(monthData.month, monthData.year)}
						>
							{formatDate(monthData.month, monthData.year)}
						</button>
					{/each}
				{/if}
			{:else}
				<!-- Current year: show months directly -->
				{#each monthsByYear[year] as monthData}
					<button
						class="btn btn-sm {selectedMonth === monthData.month && selectedYear === monthData.year
							? 'btn-primary text-base-100'
							: isCurrentMonth(monthData)
								? 'btn-outline btn-primary'
								: 'btn-ghost'} flex-shrink-0 whitespace-nowrap"
						aria-label={formatDate(monthData.month, monthData.year)}
						on:click={() => handleMonthSelect(monthData.month, monthData.year)}
					>
						{formatDate(monthData.month, monthData.year)}
					</button>
				{/each}
			{/if}
		{/each}
	</div>
</div>
