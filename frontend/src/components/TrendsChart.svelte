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
		type ChartDataset,
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
	// When on, every drawn line (total + each series) gets a dashed 3-month
	// trailing-average companion so the noisy monthly line shows its real direction.
	export let smooth: boolean = false;

	let canvasElement: HTMLCanvasElement;
	let chart: ChartJS | null = null;
	let unsubscribeTheme: (() => void) | null = null;

	const TOTAL_COLOR = '#6366f1'; // indigo-500
	const SMOOTH_WINDOW = 3;

	$: currentLocale = $locale || 'pt';

	function monthLabels(): string[] {
		return months.map((m) =>
			new Date(m.year, m.month - 1, 1).toLocaleDateString(currentLocale, {
				month: 'short',
				year: '2-digit'
			})
		);
	}

	// Trailing average: point i = mean of the up-to-`window` values ending at i.
	function movingAverage(data: number[], window: number): number[] {
		return data.map((_, i) => {
			const slice = data.slice(Math.max(0, i - (window - 1)), i + 1);
			return slice.reduce((a, b) => a + b, 0) / slice.length;
		});
	}

	function smoothedDataset(label: string, data: number[], color: string): ChartDataset<'line'> {
		return {
			label,
			data: movingAverage(data, SMOOTH_WINDOW),
			borderColor: color,
			backgroundColor: 'transparent',
			fill: false,
			tension: 0.4,
			borderWidth: 2,
			borderDash: [6, 4],
			pointRadius: 0,
			pointHoverRadius: 0
		};
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

		const smoothLabel = $t('trends.smoothed');
		// The set of labels whose tooltip should show a "% of that month" share.
		// Smoothed lines and the grand total are excluded (a share of itself is 100%).
		const shareLabels = new Set(series.map((s) => s.name));

		// Chart.js defines internal properties on the arrays/objects it receives, which
		// throws on Svelte $state proxies — hand it plain (cloned) arrays instead.
		const datasets: ChartDataset<'line'>[] = [];

		if (showTotal) {
			datasets.push({
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
			});
			if (smooth) datasets.push(smoothedDataset(smoothLabel, total, TOTAL_COLOR));
		}

		for (const s of series) {
			datasets.push({
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
			});
			if (smooth) datasets.push(smoothedDataset(smoothLabel, s.totals, s.color));
		}

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
						labels: {
							padding: 15,
							color: legendColor,
							font: { size: 12 },
							usePointStyle: true,
							// The dashed averages are visual guides, not their own series.
							filter: (item) => item.text !== smoothLabel
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
							label: (context: TooltipItem<'line'>) => {
								const value = context.parsed.y ?? 0;
								let line = `${context.dataset.label || ''}: ${formatCurrency(value)}`;
								// For a category line, append its share of that month's grand total.
								if (shareLabels.has(context.dataset.label || '')) {
									const monthTotal = total[context.dataIndex] ?? 0;
									line += monthTotal > 0 ? ` (${Math.round((value / monthTotal) * 100)}%)` : ' (—)';
								}
								return line;
							}
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
		void smooth;
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
