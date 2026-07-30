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
	import { locale } from 'svelte-i18n';
	import { themeService } from '$lib/services/themeService';
	import { formatCurrency } from '$lib/utils/currency';
	import type { TrendMonth } from '$lib/types';

	type ChartSeries = { name: string; color: string; totals: number[] };

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

	export let months: TrendMonth[] = [];
	export let total: number[] = [];
	export let series: ChartSeries[] = [];
	export let showTotal: boolean = true;

	let canvasElement: HTMLCanvasElement;
	let chart: ChartJS | null = null;
	let unsubscribeTheme: (() => void) | null = null;

	const TOTAL_COLOR = '#6366f1'; // indigo-500

	$: currentLocale = $locale || 'pt';

	function monthLabels(): string[] {
		return months.map((m) =>
			new Date(m.year, m.month - 1, 1).toLocaleDateString(currentLocale, {
				month: 'short',
				year: '2-digit'
			})
		);
	}

	function createChart() {
		if (!canvasElement || months.length === 0) return;

		if (chart) {
			chart.destroy();
			chart = null;
		}

		const {
			legendColor,
			axisTextColor,
			gridColor,
			tooltipBg,
			tooltipTitleColor,
			tooltipBodyColor,
			tooltipBorderColor
		} = themeService.getThemeColors();

		// Chart.js defines internal properties on the arrays/objects it receives, which
		// throws on Svelte $state proxies — hand it plain (cloned) arrays instead.
		const totalDataset = showTotal
			? [
					{
						label: $t('trends.total'),
						data: total.slice(),
						borderColor: TOTAL_COLOR,
						backgroundColor: 'rgba(99, 102, 241, 0.10)',
						fill: true,
						tension: 0.25,
						borderWidth: 3,
						pointRadius: 3,
						pointHoverRadius: 6,
						pointBackgroundColor: TOTAL_COLOR
					}
				]
			: [];

		const datasets = [
			...totalDataset,
			...series.map((s) => ({
				label: s.name,
				data: s.totals.slice(),
				borderColor: s.color,
				backgroundColor: s.color,
				fill: false,
				tension: 0.25,
				borderWidth: 2,
				pointRadius: 2,
				pointHoverRadius: 5,
				pointBackgroundColor: s.color
			}))
		];

		const config: ChartConfiguration = {
			type: 'line',
			data: { labels: monthLabels(), datasets },
			options: {
				responsive: true,
				maintainAspectRatio: false,
				interaction: { mode: 'index', intersect: false },
				plugins: {
					title: { display: false },
					legend: {
						display: true,
						position: 'bottom',
						labels: { padding: 15, color: legendColor, font: { size: 12 }, usePointStyle: true }
					},
					tooltip: {
						backgroundColor: tooltipBg,
						titleColor: tooltipTitleColor,
						bodyColor: tooltipBodyColor,
						borderColor: tooltipBorderColor,
						borderWidth: 1,
						cornerRadius: 8,
						callbacks: {
							label: (context: TooltipItem<'line'>) =>
								`${context.dataset.label || ''}: ${formatCurrency(context.parsed.y ?? 0)}`
						}
					}
				},
				scales: {
					x: {
						grid: { color: gridColor, lineWidth: 1 },
						ticks: { color: axisTextColor, font: { size: 11 } }
					},
					y: {
						grid: { color: gridColor, lineWidth: 1 },
						ticks: {
							color: axisTextColor,
							font: { size: 11 },
							callback: (value: string | number) => formatCurrency(Number(value), 0)
						}
					}
				}
			}
		};

		chart = new ChartJS(canvasElement, config);
	}

	// Rebuild on data, selection, total-toggle, or locale change.
	$: if (canvasElement && months.length > 0) {
		void total;
		void series;
		void showTotal;
		void currentLocale;
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

<div class="h-80 w-full sm:h-96">
	<canvas bind:this={canvasElement}></canvas>
</div>
