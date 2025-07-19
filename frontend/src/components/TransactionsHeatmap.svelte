<script lang="ts">
	import type { DailyTotals } from '$lib/types';
	import { onMount } from 'svelte';

	export let dailyTransactions: DailyTotals[] = [];
	export let startDate: string;
	export let endDate: string;
	export let largestDebit: number = 0;
	export let largestCredit: number = 0;

	let days: string[] = [];

	function generateDays(start: string, end: string) {
		const startDate = new Date(start);
		const endDate = new Date(end);
		const days = [];
		for (let d = new Date(startDate); d <= endDate; d.setDate(d.getDate() + 1)) {
			days.push(new Date(d).toISOString().split('T')[0]);
		}
		return days;
	}

	let transactionMap: { [key: string]: number } = {};
	$: transactionMap = dailyTransactions.reduce((map: { [key: string]: number }, tx) => {
		map[tx.date] = tx.total;
		return map;
	}, {});

	onMount(() => {
		days = generateDays(startDate, endDate);
	});

	// Group days by week (array of arrays)
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

	const greenClasses = [
		'bg-green-100',
		'bg-green-200',
		'bg-green-300',
		'bg-green-400',
		'bg-green-500',
		'bg-green-600',
		'bg-green-700',
		'bg-green-800',
		'bg-green-900'
	];
	const redClasses = [
		'bg-red-100',
		'bg-red-200',
		'bg-red-300',
		'bg-red-400',
		'bg-red-500',
		'bg-red-600',
		'bg-red-700',
		'bg-red-800',
		'bg-red-900'
	];

	function getColor(total: number | undefined): string {
		if (total === undefined) return 'bg-gray-200';
		if (total > 0) {
			const intensity = Math.ceil((total / largestCredit) * (greenClasses.length - 1));
			return greenClasses[Math.min(greenClasses.length - 1, Math.max(0, intensity))];
		}
		if (total < 0) {
			const intensity = Math.ceil((Math.abs(total) / largestDebit) * (redClasses.length - 1));
			return redClasses[Math.min(redClasses.length - 1, Math.max(0, intensity))];
		}
		return 'bg-gray-200';
	}
</script>

<div class="inline-flex w-full flex-col gap-2">
	{#each weeks as week}
		<div class="flex gap-2">
			{#each week as day}
				<div class="tooltip" data-tip={day ? `${transactionMap[day] ?? 0}€` : ''}>
					<div
						class="heatmap-square aspect-square w-[32px] rounded {getColor(transactionMap[day])}"
					></div>
				</div>
			{/each}
		</div>
	{/each}
</div>

<style>
	.heatmap-square {
		transition: transform 0.2s ease-in-out;
	}
	.heatmap-square:hover {
		transform: scale(1.1);
		z-index: 1;
	}
</style>
