<script lang="ts">
	import { heatmapDisplayMode, updateHeatmapDisplayMode, type HeatmapDisplayMode } from '$lib/stores/uiPreferences';

	// Import types and Svelte utilities
	import type { DailyTotals } from '$lib/types';
	import { BarChart3, TrendingUp, TrendingDown, ChevronDown } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { locale, t } from 'svelte-i18n';

	// Props received from parent component
	export let dailyTransactions: DailyTotals[] = [];
	export let startDate: string;
	export let endDate: string;
	export let largestDebit: number = 0;
	export let largestCredit: number = 0;

	// Today's date in YYYY-MM-DD format
	let today: string = new Date().toISOString().split('T')[0];
	// Array of all days to display in the heatmap
	let days: string[] = [];

	// Current locale for date formatting
	$: currentLocale = $locale || 'pt';

	// Format a day string to a localized date string
	function formatDay(day: string): string {
		// input is a day in the format YYYY-MM-DD
		if (!day) return '';
		const date = new Date(day);
		// should return the day in the format dd-mm if portuguese or mm-dd if english, so depending on the locale
		return date.toLocaleDateString(currentLocale, {
			month: 'short',
			day: 'numeric'
		});
	}

	// Check if a given day is today
	function isToday(day: string): boolean {
		return day === today;
	}

	// Generate an array of day strings from start to end date (inclusive)
	function generateDays(start: string, end: string) {
		const startDate = new Date(start);
		const endDate = new Date(end);
		const days = [];
		for (let d = new Date(startDate); d <= endDate; d.setDate(d.getDate() + 1)) {
			days.push(new Date(d).toISOString().split('T')[0]);
		}
		return days;
	}

	// Map of date string to transaction total for quick lookup
	let transactionMap: { [key: string]: { credit: number; debit: number; difference: number } } = {};
	$: transactionMap = dailyTransactions.reduce(
		(map: { [key: string]: { credit: number; debit: number; difference: number } }, tx) => {
			map[tx.date] = {
				credit: tx.credit || 0,
				debit: tx.debit || 0,
				difference: tx.difference || 0
			};
			return map;
		},
		{} as { [key: string]: { credit: number; debit: number; difference: number } }
	);

	// Get the value based on selected mode
	function getValue(day: string): number {
		const data = transactionMap[day];
		if (!data) return 0;

		switch ($heatmapDisplayMode) {
			case 'credit':
				return data.credit;
			case 'debit':
				return data.debit;
			case 'difference':
				return data.difference;
			default:
				return 0;
		}
	}

	// Get maximum value for scaling based on mode
	function getMaxValue(): number {
		switch ($heatmapDisplayMode) {
			case 'credit':
				return largestCredit;
			case 'debit':
				return largestDebit;
			case 'difference':
				return Math.max(largestCredit, largestDebit);
			default:
				return 1;
		}
	}

	// Generate the days array when the component mounts
	onMount(() => {
		days = generateDays(startDate, endDate);
	});

	// Group days into weeks for table display
	let weeks: string[][] = [];
	$: {
		weeks = [];
		if (days.length) {
			let week: string[] = [];
			for (let i = 0; i < days.length; i++) {
				const dayOfWeek = new Date(days[i]).getDay(); // 0 = Sunday
				if (week.length === 0 && dayOfWeek !== 0) {
					// Fill start of first week with empty days
					for (let j = 0; j < dayOfWeek; j++) week.push('');
				}
				week.push(days[i]);
				if (week.length === 7) {
					weeks.push(week);
					week = [];
				}
			}
			if (week.length) {
				// Fill end of last week with empty days
				while (week.length < 7) week.push('');
				weeks.push(week);
			}
		}
	}

	// Get the month key (YYYY-MM) from a date string
	function getMonthKey(day: string): string {
		return day.slice(0, 7); // "YYYY-MM"
	}

	// Group days by month for multi-month display
	let months: Record<string, string[]> = {};
	$: {
		months = {};
		if (days.length) {
			for (const day of days) {
				if (!day) continue;
				const monthKey = getMonthKey(day);
				if (!months[monthKey]) months[monthKey] = [];
				months[monthKey].push(day);
			}
		}
	}

	// Helper to group days into weeks for each month
	function groupWeeks(days: string[]) {
		const weeks: string[][] = [];
		let week: string[] = [];
		for (let i = 0; i < days.length; i++) {
			const dayOfWeek = new Date(days[i]).getDay();
			if (week.length === 0 && dayOfWeek !== 0) {
				for (let j = 0; j < dayOfWeek; j++) week.push('');
			}
			week.push(days[i]);
			if (week.length === 7) {
				weeks.push(week);
				week = [];
			}
		}
		if (week.length) {
			while (week.length < 7) week.push('');
			weeks.push(week);
		}
		return weeks;
	}

	// Format a month key (YYYY-MM) to a localized month/year string
	function formatMonthYear(yearMonth: string): string {
		const [year, month] = yearMonth.split('-').map(Number);
		const date = new Date(year, month - 1, 1);
		return date.toLocaleDateString(currentLocale, {
			month: 'long',
			year: 'numeric'
		});
	}

	// Updated color function based on mode
	function getColor(day: string): string {
		const value = getValue(day);
		const data = transactionMap[day];

		if (!data) return 'heatmap-neutral';

		const maxValue = getMaxValue();
		if (maxValue === 0) return 'heatmap-neutral';

		switch ($heatmapDisplayMode) {
			case 'credit':
				return value > 0 ? 'heatmap-green' : 'heatmap-neutral';
			case 'debit':
				return value > 0 ? 'heatmap-red' : 'heatmap-neutral';
			case 'difference':
				if (value > 0) return 'heatmap-green';
				if (value < 0) return 'heatmap-red';
				return 'heatmap-neutral';
			default:
				return 'heatmap-neutral';
		}
	}

	function getIntensity(day: string): number {
		const value = getValue(day);
		const maxValue = getMaxValue();
		if (maxValue === 0) return 0;

		const intensity = Math.min(1, Math.abs(value) / maxValue);
		// Ensure minimum visibility (0.1) and scale to 0.1-1.0 range
		return 0.1 + intensity * 0.9;
	}

	// Format tooltip based on mode
	function getTooltipText(day: string): string {
		const data = transactionMap[day];
		if (!data) return `${formatDay(day)} - ${$t('common.no-data', { default: 'No data' })}`;

		switch ($heatmapDisplayMode) {
			case 'credit':
				return `${formatDay(day)} 💰 +${data.credit.toFixed(2)} €`;
			case 'debit':
				return `${formatDay(day)} 💸 -${data.debit.toFixed(2)} €`;
			case 'difference':
				const emoji = data.difference >= 0 ? '️💰' : '️💸';
				return `${formatDay(day)} ${emoji} ${data.difference >= 0 ? '+' : ''}${data.difference.toFixed(2)} €`;
			default:
				return `${formatDay(day)}`;
		}
	}

	function getDisplayModeInfo(mode: HeatmapDisplayMode) {
		switch (mode) {
			case 'difference':
				return { icon: BarChart3, label: $t('statistics.heatmap.difference', { default: 'Net' }) };
			case 'credit':
				return { icon: TrendingUp, label: $t('statistics.heatmap.credits', { default: 'Income' }) };
			case 'debit':
				return {
					icon: TrendingDown,
					label: $t('statistics.heatmap.debits', { default: 'Expenses' })
				};
			default:
				return { icon: BarChart3, label: 'Net' };
		}
	}

	$: currentModeInfo = getDisplayModeInfo($heatmapDisplayMode);
</script>

<!-- Mode Selection Dropdown with highlight -->
<div class="mb-4 flex justify-center">
	<div class="dropdown">
		<div tabindex="0" role="button" class="btn btn-sm btn-primary text-base-100 shadow-lg">
			<svelte:component this={currentModeInfo.icon} size={16} class="text-base-100 mr-1" />
			<span class="text-base-100">{currentModeInfo.label}</span>
			<ChevronDown size={16} class="text-base-100 ml-1" />
		</div>
		<ul
			tabindex="-1"
			class="dropdown-content menu bg-base-100 rounded-box border-base-300 z-[1] w-48 border p-2 shadow-xl"
		>
			<li>
				<button
					class="flex items-center gap-2 {$heatmapDisplayMode === 'difference'
						? 'bg-primary text-base-100'
						: ''}"
					on:click={() => (updateHeatmapDisplayMode('difference'))}
				>
					<BarChart3 size={16} />
					<span>{$t('statistics.heatmap.difference', { default: 'Net' })}</span>
				</button>
			</li>
			<li>
				<button
					class="flex items-center gap-2 {$heatmapDisplayMode === 'credit'
						? 'bg-primary text-base-100'
						: ''}"
					on:click={() => (updateHeatmapDisplayMode('credit'))}
				>
					<TrendingUp size={16} />
					<span>{$t('statistics.heatmap.credits', { default: 'Income' })}</span>
				</button>
			</li>
			<li>
				<button
					class="flex items-center gap-2 {$heatmapDisplayMode === 'debit'
						? 'bg-primary text-base-100'
						: ''}"
					on:click={() => (updateHeatmapDisplayMode('debit'))}
				>
					<TrendingDown size={16} />
					<span>{$t('statistics.heatmap.debits', { default: 'Expenses' })}</span>
				</button>
			</li>
		</ul>
	</div>
</div>

<!-- Render each month that has at least one transaction -->
{#each Object.entries(months) as [monthKey, monthDays]}
	{#if monthDays.some((day) => transactionMap[day] !== undefined)}
		<!-- Month label -->
		<h3 class="mb-2 mt-4 font-semibold">{formatMonthYear(monthKey)}</h3>
		<!-- Heatmap table for the month -->
		<table class="w-full border-separate border-spacing-2 px-10">
			<tbody>
				<!-- Render each week as a table row -->
				{#each groupWeeks(monthDays) as week}
					<tr>
						<!-- Render each day as a table cell -->
						{#each week as day}
							{#if day}
								<!-- Cell with colored square and tooltip -->
								<td class="text-center align-middle">
									<div
										class="tooltip heatmap-square mx-auto my-auto flex aspect-square h-6 w-6 items-center justify-center rounded {getColor(
											day
										)} {isToday(day) ? 'border-primary border-2' : ''}"
										style="--intensity: {getIntensity(day)}"
										data-tip={getTooltipText(day)}
									></div>
								</td>
							{:else}
								<!-- Empty cell for days not in the current month -->
								<td></td>
							{/if}
						{/each}
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
{/each}

<style>
	.heatmap-green {
		background-color: rgba(34, 197, 94, var(--intensity));
	}
	.heatmap-red {
		background-color: rgba(239, 68, 68, var(--intensity));
	}
	.heatmap-neutral {
		background-color: #e5e7eb; /* gray-200 */
	}
	/* Ensures squares are always square and sized */
	.aspect-square {
		width: 1.5rem;
		height: 1.5rem;
		display: inline-block;
	}
	/* Transition effect for heatmap squares on hover */
	.heatmap-square {
		transition: transform 0.2s ease-in-out;
	}
	.heatmap-square:hover {
		transform: scale(1.1);
		z-index: 1;
	}
</style>
