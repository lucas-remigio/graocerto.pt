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
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
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
							color: '#374151', // gray-700 for light mode
							font: {
								size: 12,
								weight: 'normal'
							}
						}
					},
					tooltip: {
						backgroundColor: '#ffffff',
						titleColor: '#111827', // gray-900
						bodyColor: '#374151', // gray-700
						borderColor: '#d1d5db', // gray-300
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
							color: '#f3f4f6', // gray-100
							lineWidth: 1
						},
						ticks: {
							color: '#6b7280', // gray-500
							font: {
								size: 11
							}
						}
					},
					y: {
						display: true,
						grid: {
							color: '#f3f4f6', // gray-100
							lineWidth: 1
						},
						ticks: {
							color: '#6b7280', // gray-500
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

	// Recreate chart when data changes
	$: if (yearlyBreakdown && yearlyBreakdown.length > 0 && canvasElement && !chart) {
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
