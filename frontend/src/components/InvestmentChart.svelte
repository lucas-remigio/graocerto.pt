<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Chart as ChartJS,
		Title,
		Tooltip,
		Legend,
		LineElement,
		LinearScale,
		PointElement,
		CategoryScale,
		LineController,
		Filler,
		type ChartConfiguration
	} from 'chart.js';
	import { t } from '$lib/i18n';

	// Register Chart.js components
	ChartJS.register(
		Title,
		Tooltip,
		Legend,
		LineElement,
		LinearScale,
		PointElement,
		CategoryScale,
		LineController,
		Filler
	);

	export let yearlyBreakdown: Array<{
		year: number;
		total_investment: number;
		total_return: number;
		total_value: number;
	}> = [];

	let canvasElement: HTMLCanvasElement;
	let chart: ChartJS | null = null;

	function formatCurrency(value: number): string {
		return new Intl.NumberFormat('pt-PT', {
			style: 'currency',
			currency: 'EUR',
			minimumFractionDigits: 0,
			maximumFractionDigits: 0
		}).format(value);
	}

	function createChart() {
		if (!canvasElement || !yearlyBreakdown.length) return;

		// Destroy existing chart
		if (chart) {
			chart.destroy();
			chart = null;
		}

		// Detect dark mode and set appropriate colors (same pattern as CategoriesPieChart)
		const isDarkMode =
			document.documentElement.getAttribute('data-theme')?.includes('dark') ||
			window.matchMedia('(prefers-color-scheme: dark)').matches;
		const legendColor = isDarkMode ? '#e5e7eb' : '#374151'; // light gray for dark mode, dark gray for light mode
		const axisTextColor = isDarkMode ? '#e5e7eb' : '#111827'; // light gray for dark mode, dark text for light mode
		const gridColor = isDarkMode ? '#374151' : '#f3f4f6'; // darker grid for dark mode, light grid for light mode
		const tooltipBg = isDarkMode ? '#1f2937' : '#ffffff';
		const tooltipTitleColor = isDarkMode ? '#e5e7eb' : '#111827';
		const tooltipBodyColor = isDarkMode ? '#e5e7eb' : '#374151';
		const tooltipBorderColor = isDarkMode ? '#374151' : '#d1d5db';

		const config: ChartConfiguration = {
			type: 'line',
			data: {
				labels: yearlyBreakdown.map((y) => `${$t('investment-calculator.results.year')} ${y.year}`),
				datasets: [
					{
						label: $t('investment-calculator.results.total-investment'),
						data: yearlyBreakdown.map((y) => y.total_investment),
						borderColor: '#6366f1', // indigo-500
						backgroundColor: 'rgba(99, 102, 241, 0.1)',
						fill: false,
						tension: 0.2,
						borderWidth: 3,
						pointRadius: 4,
						pointHoverRadius: 6,
						pointBackgroundColor: '#6366f1',
						pointBorderColor: '#ffffff',
						pointBorderWidth: 2
					},
					{
						label: $t('investment-calculator.results.total-value'),
						data: yearlyBreakdown.map((y) => y.total_value),
						borderColor: '#10b981', // emerald-500
						backgroundColor: 'rgba(16, 185, 129, 0.1)',
						fill: false,
						tension: 0.2,
						borderWidth: 3,
						pointRadius: 4,
						pointHoverRadius: 6,
						pointBackgroundColor: '#10b981',
						pointBorderColor: '#ffffff',
						pointBorderWidth: 2
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					title: {
						display: false
					},
					legend: {
						display: true,
						position: 'bottom',
						labels: {
							padding: 15,
							color: 'fafafa', // Dynamic color based on theme
							font: {
								size: 12,
								weight: 'normal'
							}
						}
					},
					tooltip: {
						backgroundColor: tooltipBg,
						titleColor: tooltipTitleColor,
						bodyColor: tooltipBodyColor,
						borderColor: tooltipBorderColor,
						borderWidth: 1,
						cornerRadius: 8,
						callbacks: {
							label: function (context: any) {
								const label = context.dataset.label || '';
								const value = formatCurrency(context.parsed.y);
								return `${label}: ${value}`;
							}
						}
					}
				},
				scales: {
					x: {
						display: true,
						grid: {
							color: gridColor, // Dynamic grid color
							lineWidth: 1
						},
						ticks: {
							color: axisTextColor, // Dynamic axis text color
							font: {
								size: 11
							}
						}
					},
					y: {
						display: true,
						grid: {
							color: gridColor, // Dynamic grid color
							lineWidth: 1
						},
						ticks: {
							color: axisTextColor, // Dynamic axis text color
							font: {
								size: 11
							},
							callback: function (value: any) {
								return formatCurrency(Number(value));
							}
						}
					}
				}
			}
		};

		chart = new ChartJS(canvasElement, config);
	}

	// Reactive updates when data changes
	$: if (canvasElement && yearlyBreakdown) {
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

{#if yearlyBreakdown && yearlyBreakdown.length > 0}
	<div class="w-full">
		<div class="h-96">
			<canvas bind:this={canvasElement}></canvas>
		</div>
	</div>
{/if}
