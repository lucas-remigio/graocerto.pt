<!-- src/components/PieChart.svelte -->
<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Chart, registerables } from 'chart.js';
	import { PieChart } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import type { CategoryStatistic } from '$lib/types';

	export let data: CategoryStatistic[] = [];
	export let isCredit: boolean = false;

	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	// Register Chart.js components
	Chart.register(...registerables);

	function createChart() {
		if (!canvas || data.length === 0) return;

		// Destroy existing chart
		if (chart) {
			chart.destroy();
		}

		// Use category colors from data, with fallback to generated colors
		const colors = data.map((item) => item.color || '#6b7280');

		// Detect dark mode and set appropriate text color
		const isDarkMode =
			document.documentElement.getAttribute('data-theme')?.includes('dark') ||
			window.matchMedia('(prefers-color-scheme: dark)').matches;
		const legendColor = isDarkMode ? '#e5e7eb' : '#374151'; // light gray for dark mode, dark gray for light mode

		chart = new Chart(canvas, {
			type: 'pie',
			data: {
				labels: data.map((item) => item.name),
				datasets: [
					{
						data: data.map((item) => item.percentage),
						backgroundColor: colors,
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
						position: 'bottom',
						labels: {
							color: legendColor,
							font: {
								size: 12
							},
							padding: 15,
							usePointStyle: true,
							pointStyle: 'circle'
						}
					},
					tooltip: {
						backgroundColor: isDarkMode ? '#1f2937' : '#ffffff',
						titleColor: legendColor,
						bodyColor: legendColor,
						borderColor: isDarkMode ? '#374151' : '#e5e7eb',
						borderWidth: 1,
						callbacks: {
							label: function (context) {
								const category = data[context.dataIndex];
								return [
									`${category.name}: ${category.percentage.toFixed(1)}%`,
									`${category.count} transactions`,
									`€${category.total.toFixed(2)}`
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
					duration: 1000
				}
			}
		});
	}

	// Reactive updates when data changes
	$: if (canvas && data) {
		createChart();
	}

	onMount(() => {
		createChart();
	});

	onDestroy(() => {
		if (chart) {
			chart.destroy();
		}
	});
</script>

<div class="relative">
	{#if data.length > 0}
		<div class="h-80 w-full">
			<canvas bind:this={canvas} class="max-h-full max-w-full"></canvas>
		</div>

		<!-- Statistics below chart -->
		<div class="mt-4 space-y-2">
			{#each data as category, index}
				<div class="flex items-center justify-between text-sm">
					<div class="flex items-center gap-2">
						<div
							class="h-3 w-3 rounded-full"
							style="background-color: {category.color || '#6b7280'}"
						></div>
						<span class="font-medium">{category.name}</span>
					</div>
					<div class="text-right">
						<div class="font-semibold">{category.percentage.toFixed(1)}%</div>
						<div class="text-xs opacity-70">
							{category.count} • €{category.total.toFixed(2)}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<!-- Enhanced placeholder for empty data -->
		<div class="flex h-80 items-center justify-center">
			<div class="text-center">
				<PieChart size={48} class="text-base-content/30 mx-auto mb-3" />
				<p class="text-base-content/60 text-base font-medium">{$t('statistics.no-data')}</p>
				<p class="text-base-content/50 text-sm">
					{isCredit ? $t('statistics.no-credit-categories') : $t('statistics.no-debit-categories')}
				</p>
			</div>
		</div>
	{/if}
</div>
