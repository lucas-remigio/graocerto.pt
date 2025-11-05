<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Chart, registerables } from 'chart.js';
	import { PieChart } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import type { CategoryStatistic } from '$lib/types';
	import { themeService } from '$lib/services/themeService';

	export let data: CategoryStatistic[] = [];
	export let isCredit: boolean = false;

	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;
	let unsubscribeTheme: (() => void) | null = null;

	Chart.register(...registerables);

	// Flatten all categories and subcategories into one list
	$: flattenedData = data.flatMap((cat) => {
		const items = [];

		// If category has subcategories, add them individually
		if (cat.subcategories && cat.subcategories.length > 0) {
			// Add parent direct spending if any
			const directTotal = cat.total - cat.subcategories.reduce((sum, sub) => sum + sub.total, 0);
			const directCount = cat.count - cat.subcategories.reduce((sum, sub) => sum + sub.count, 0);

			if (directCount > 0) {
				items.push({
					name: `${cat.name}`,
					total: directTotal,
					color: cat.color,
					count: directCount,
					parent: null
				});
			}

			// Add all subcategories
			cat.subcategories.forEach((sub) => {
				items.push({
					name: sub.name,
					total: sub.total,
					color: sub.color,
					count: sub.count,
					parent: cat.name
				});
			});
		} else {
			// No subcategories, add category as-is
			items.push({
				name: cat.name,
				total: cat.total,
				color: cat.color,
				count: cat.count,
				parent: null
			});
		}

		return items;
	});

	// Sort by total descending
	$: sortedData = [...flattenedData].sort((a, b) => b.total - a.total);

	function createChart() {
		if (!canvas || sortedData.length === 0) return;

		if (chart) {
			chart.destroy();
		}

		const themeColors = themeService.getThemeColors();
		const { legendColor, tooltipBg, tooltipTitleColor, tooltipBodyColor, tooltipBorderColor } =
			themeColors;

		chart = new Chart(canvas, {
			type: 'pie',
			data: {
				labels: sortedData.map((item) => item.name),
				datasets: [
					{
						data: sortedData.map((item) => item.total),
						backgroundColor: sortedData.map((item) => item.color),
						borderColor: '#ffffff',
						borderWidth: 2,
						hoverBorderWidth: 3,
						hoverOffset: 4
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: {
						display: false
					},
					tooltip: {
						backgroundColor: tooltipBg,
						titleColor: tooltipTitleColor,
						bodyColor: tooltipBodyColor,
						borderColor: tooltipBorderColor,
						borderWidth: 1,
						callbacks: {
							label: function (context) {
								const item = sortedData[context.dataIndex];
								const total = sortedData.reduce((sum, i) => sum + i.total, 0);
								const percentage = ((item.total / total) * 100).toFixed(1);
								return [
									`${item.name}`,
									...(item.parent ? [`- ${item.parent}`] : []),
									`${item.count} ${item.count === 1 ? $t('common.transaction') : $t('common.transactions')}`,
									`€${item.total.toFixed(2)} (${percentage}%)`
								];
							}
						}
					}
				},
				elements: {
					arc: {
						borderRadius: 4
					}
				},
				animation: {
					animateRotate: true,
					animateScale: true,
					duration: 800
				}
			}
		});
	}

	$: if (canvas && sortedData) {
		createChart();
	}

	onMount(() => {
		createChart();
		unsubscribeTheme = themeService.subscribe(() => {
			if (chart) createChart();
		});
	});

	onDestroy(() => {
		if (chart) chart.destroy();
		if (unsubscribeTheme) unsubscribeTheme();
	});
</script>

<div class="relative">
	{#if sortedData.length > 0}
		<!-- Chart Title -->
		<div class="mb-3 text-center">
			<h4 class="text-sm font-medium opacity-70">
				{$t('statistics.detailed-breakdown', { default: 'Detailed Breakdown' })}
			</h4>
		</div>

		<!-- Chart -->
		<div class="h-80 w-full">
			<canvas bind:this={canvas} class="max-h-full max-w-full"></canvas>
		</div>

		<!-- Legend -->
		<div class="mt-4 max-h-96 space-y-2 overflow-y-auto">
			{#each sortedData as item}
				<div class="bg-base-200 flex items-center justify-between rounded p-2">
					<div class="flex items-center gap-2">
						<span class="h-3 w-3 flex-shrink-0 rounded-full" style="background-color: {item.color};"
						></span>
						<div class="min-w-0">
							<span class="text-sm font-medium">{item.name}</span>
							{#if item.parent}
								<span class="ml-1 text-xs opacity-60">({item.parent})</span>
							{/if}
						</div>
					</div>
					<div class="text-right">
						<div class="text-sm font-semibold">€{item.total.toFixed(2)}</div>
						<div class="text-xs opacity-60">{item.count} tx</div>
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<!-- Empty state -->
		<div class="flex h-80 items-center justify-center">
			<div class="text-center">
				<PieChart size={48} class="text-base-content/30 mx-auto mb-3" />
				<p class="text-base-content/60 text-base font-medium">{$t('statistics.no-data')}</p>
			</div>
		</div>
	{/if}
</div>
