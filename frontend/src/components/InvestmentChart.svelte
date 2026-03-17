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
		type ChartConfiguration,
		type TooltipItem
	} from 'chart.js';
	import { t } from '$lib/i18n';
	import { themeService } from '$lib/services/themeService';

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
	import { formatCurrency } from '$lib/utils/currency';

	let chart: ChartJS | null = null;
	let unsubscribeTheme: (() => void) | null = null;

	function createChart() {
		if (!canvasElement || !yearlyBreakdown.length) return;

		// Destroy existing chart
		if (chart) {
			chart.destroy();
			chart = null;
		}

		// Get theme colors from theme service
		const themeColors = themeService.getThemeColors();

		const {
			legendColor,
			axisTextColor,
			gridColor,
			tooltipBg,
			tooltipTitleColor,
			tooltipBodyColor,
			tooltipBorderColor
		} = themeColors;

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
							color: legendColor, // Dynamic color based on theme
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
							label: function (context: TooltipItem<'line'>) {
								const label = context.dataset.label || '';
								const value = formatCurrency(context.parsed.y, 0);
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
							callback: function (value: string | number) {
								return formatCurrency(Number(value), 0);
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

		// Subscribe to theme changes
		unsubscribeTheme = themeService.subscribe(() => {
			if (chart) {
				createChart();
			}
		});
	});

	onDestroy(() => {
		if (chart) {
			chart.destroy();
		}
		if (unsubscribeTheme) {
			unsubscribeTheme();
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
