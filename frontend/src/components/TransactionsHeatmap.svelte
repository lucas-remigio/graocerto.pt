<script lang="ts">
	import { t } from '$lib/i18n';
	import {
		heatmapDisplayMode,
		updateHeatmapDisplayMode,
		type HeatmapDisplayMode
	} from '$lib/stores/uiPreferences';
	import type { DailyTotals } from '$lib/types';
	import { BarChart3, TrendingUp, TrendingDown, ChevronDown } from 'lucide-svelte';
	import { locale } from 'svelte-i18n';

	// Props
	export let dailyTransactions: DailyTotals[] = [];
	export let startDate: string;
	export let endDate: string;

	const today = new Date().toISOString().split('T')[0];
	$: currentLocale = $locale || 'pt';

	// Build map + maxima in one pass
	type DayAgg = { credit: number; debit: number; difference: number };
	let transactionMap: Record<string, DayAgg> = {};
	let max = { credit: 0, debit: 0, absDiff: 0 };

	$: {
		transactionMap = {};
		max = { credit: 0, debit: 0, absDiff: 0 };
		for (const d of dailyTransactions) {
			const entry: DayAgg = {
				credit: d.credit || 0,
				debit: d.debit || 0,
				difference: d.difference || 0
			};
			transactionMap[d.date] = entry;
			if (entry.credit > max.credit) max.credit = entry.credit;
			if (entry.debit > max.debit) max.debit = entry.debit;
			const abs = Math.abs(entry.difference);
			if (abs > max.absDiff) max.absDiff = abs;
		}
	}

	// Calendar structure: months -> weeks -> days
	interface MonthBlock {
		key: string;
		label: string;
		weeks: string[][];
		hasData: boolean;
	}
	let months: MonthBlock[] = [];

	function formatMonthYear(key: string): string {
		const [y, m] = key.split('-').map(Number);
		return new Date(y, m - 1, 1).toLocaleDateString(currentLocale, {
			month: 'long',
			year: 'numeric'
		});
	}

	function eachDay(start: string, end: string): string[] {
		const out: string[] = [];
		const s = new Date(start);
		const e = new Date(end);
		for (let d = new Date(s); d <= e; d.setDate(d.getDate() + 1)) {
			out.push(d.toISOString().split('T')[0]);
		}
		return out;
	}

	function toWeeks(days: string[]): string[][] {
		const weeks: string[][] = [];
		let w: string[] = [];
		for (const day of days) {
			const dow = new Date(day).getDay();
			if (w.length === 0 && dow !== 0) {
				for (let i = 0; i < dow; i++) w.push('');
			}
			w.push(day);
			if (w.length === 7) {
				weeks.push(w);
				w = [];
			}
		}
		if (w.length) {
			while (w.length < 7) w.push('');
			weeks.push(w);
		}
		return weeks;
	}

	// Rebuild months reactively
	$: {
		if (!startDate || !endDate) {
			months = [];
		} else {
			const allDays = eachDay(startDate, endDate);
			const byMonth: Record<string, string[]> = {};
			for (const d of allDays) {
				const key = d.slice(0, 7);
				(byMonth[key] ||= []).push(d);
			}
			months = Object.entries(byMonth).map(([key, days]) => {
				const hasData = days.some((d) => transactionMap[d]);
				return {
					key,
					label: formatMonthYear(key),
					weeks: toWeeks(days),
					hasData
				};
			});
		}
	}

	// Value + max lookup
	function getValue(day: string): number {
		const data = transactionMap[day];
		if (!data) return 0;
		switch ($heatmapDisplayMode) {
			case 'credit':
				return data.credit;
			case 'debit':
				return data.debit;
			default:
				return data.difference;
		}
	}
	function getMax(): number {
		switch ($heatmapDisplayMode) {
			case 'credit':
				return max.credit || 1;
			case 'debit':
				return max.debit || 1;
			default:
				return max.absDiff || 1;
		}
	}

	// Color class (income green / expense red / net)
	function getColor(day: string): string {
		const v = getValue(day);
		if (v === 0) return 'heatmap-neutral';
		if ($heatmapDisplayMode === 'credit') return 'heatmap-green';
		if ($heatmapDisplayMode === 'debit') return 'heatmap-red';
		// difference
		return v > 0 ? 'heatmap-green' : 'heatmap-red';
	}

	// 3 bucket intensity
	function getIntensity(day: string): number {
		const v = getValue(day);
		if (v === 0) return 0;
		const pct = (Math.abs(v) / getMax()) * 100;
		if (pct < 25) return 0.35;
		if (pct < 75) return 0.65;
		return 0.95;
	}

	function formatDay(day: string): string {
		if (!day) return '';
		return new Date(day).toLocaleDateString(currentLocale, {
			month: 'short',
			day: 'numeric'
		});
	}

	function getTooltipText(day: string): string {
		const d = transactionMap[day];
		if (!d) return `${formatDay(day)} - ${$t('common.no-data', { default: 'No data' })}`;
		if ($heatmapDisplayMode === 'credit') return `${formatDay(day)} 💰 +${d.credit.toFixed(2)} €`;
		if ($heatmapDisplayMode === 'debit') return `${formatDay(day)} 💸 -${d.debit.toFixed(2)} €`;
		const emoji = d.difference >= 0 ? '💰' : '💸';
		return `${formatDay(day)} ${emoji} ${d.difference >= 0 ? '+' : ''}${d.difference.toFixed(2)} €`;
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

<!-- Mode selector -->
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
					on:click={() => updateHeatmapDisplayMode('difference')}
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
					on:click={() => updateHeatmapDisplayMode('credit')}
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
					on:click={() => updateHeatmapDisplayMode('debit')}
				>
					<TrendingDown size={16} />
					<span>{$t('statistics.heatmap.debits', { default: 'Expenses' })}</span>
				</button>
			</li>
		</ul>
	</div>
</div>

{#each months as m}
	{#if m.hasData}
		<h3 class="mb-2 mt-4 font-semibold">{m.label}</h3>
		<table class="w-full border-separate border-spacing-2">
			<tbody>
				{#each m.weeks as week}
					<tr>
						{#each week as day}
							<td class="text-center align-middle">
								{#if day}
									<div
										class="tooltip heatmap-square mx-auto aspect-square h-6 w-6 rounded {getColor(
											day
										)} {day === today ? 'border-primary border-2' : ''}"
										style="--intensity:{getIntensity(day)}"
										data-tip={getTooltipText(day)}
									></div>
								{/if}
							</td>
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
		background-color: #e5e7eb;
	}
	.aspect-square {
		width: 1.5rem;
		height: 1.5rem;
		display: inline-block;
	}
	.heatmap-square {
		transition: transform 0.2s ease-in-out;
	}
	.heatmap-square:hover {
		transform: scale(1.1);
		z-index: 1;
	}
</style>
