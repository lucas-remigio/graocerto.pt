<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Calendar } from 'lucide-svelte';
	import type { MonthYear } from '$lib/types';
	import { locale } from '$lib/i18n';

	export let availableMonths: MonthYear[] = [];
	export let selectedMonth: number | null = null;
	export let selectedYear: number | null = null;

	const dispatch = createEventDispatcher<{
		monthSelect: { month: number | null; year: number | null };
	}>();

	const currentMonth = new Date().getMonth() + 1; // 1-12 (1 = January)
	const currentYear = new Date().getFullYear();

	function handleMonthSelect(month: number | null, year: number | null) {
		dispatch('monthSelect', { month, year });
	}

	function formatDate(month: number, year: number): string {
		const date = new Date(year, month - 1); // month is 0-indexed in JavaScript
		return date.toLocaleString(currentLocale, { month: 'long', year: 'numeric' });
	}

	function isCurrentMonth(monthData: MonthYear): boolean {
		return monthData.month === currentMonth && monthData.year === currentYear;
	}

	$: currentLocale = $locale || 'en-US';
</script>

<div class="mb-4">
	<div class="flex items-center gap-2 overflow-x-auto pb-2">
		<button
			class="btn btn-sm btn-circle {selectedMonth === null && selectedYear === null
				? 'btn-primary'
				: 'btn-ghost'} flex-shrink-0"
			on:click={() => handleMonthSelect(null, null)}
			title="Show all transactions"
		>
			<Calendar size={20} />
		</button>

		{#each availableMonths as monthData}
			<button
				class="btn btn-sm {selectedMonth === monthData.month && selectedYear === monthData.year
					? 'btn-primary'
					: isCurrentMonth(monthData)
						? 'btn-outline btn-primary'
						: 'btn-ghost'} 
                    flex-shrink-0 whitespace-nowrap"
				on:click={() => handleMonthSelect(monthData.month, monthData.year)}
			>
				{formatDate(monthData.month, monthData.year)}
			</button>
		{/each}
	</div>
</div>
