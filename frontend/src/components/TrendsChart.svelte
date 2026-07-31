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
		type ScriptableContext,
		type Plugin,
		type TooltipItem
	} from 'chart.js';
	import { t } from '$lib/i18n';
	import { locale } from 'svelte-i18n';
	import { themeService } from '$lib/services/themeService';
	import { formatCurrency } from '$lib/utils/currency';
	import type { TrendMonth } from '$lib/types';

	type ChartSeries = { name: string; color: string; totals: number[] };
	type Pt = { x: number; y: number };

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

	const SMOOTH_WINDOW = 3;
	// Ledger identity: figures and axis ticks are monospace.
	const MONO =
		"ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace";

	$: currentLocale = $locale || 'pt';

	function hexToRgba(hex: string, alpha: number): string {
		const h = hex.replace('#', '');
		const r = parseInt(h.slice(0, 2), 16);
		const g = parseInt(h.slice(2, 4), 16);
		const b = parseInt(h.slice(4, 6), 16);
		return `rgba(${r}, ${g}, ${b}, ${alpha})`;
	}

	function prefersReducedMotion(): boolean {
		return (
			typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
		);
	}

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
			isDarkMode,
			seriesTotal: totalColor,
			legendColor,
			axisTextColor,
			gridColor,
			tooltipBg,
			tooltipTitleColor,
			tooltipBodyColor,
			tooltipBorderColor
		} = themeService.getThemeColors();

		// Chart surface, for the ring around end/peak markers so they stay legible.
		const surface = isDarkMode ? '#1e1e1e' : '#ffffff';
		const totalLabel = $t('trends.total');
		const peakLabel = $t('trends.peak-month');
		const smoothLabel = $t('trends.smoothed');
		// The set of labels whose tooltip should show a "% of that month" share.
		const shareLabels = new Set(series.map((s) => s.name));

		// Chart.js defines internal properties on the arrays/objects it receives, which
		// throws on Svelte $state proxies — hand it plain (cloned) arrays instead.
		const datasets: ChartDataset<'line'>[] = [];

		if (showTotal) {
			datasets.push({
				label: totalLabel,
				data: total.slice(),
				borderColor: totalColor,
				// Gradient wash under the line, built once the plot area is known.
				backgroundColor: (context: ScriptableContext<'line'>) => {
					const { ctx, chartArea } = context.chart;
					if (!chartArea) return 'transparent';
					const g = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
					g.addColorStop(0, hexToRgba(totalColor, 0.22));
					g.addColorStop(1, hexToRgba(totalColor, 0));
					return g;
				},
				fill: true,
				tension: 0.25,
				borderWidth: 2.5,
				pointRadius: 0, // resting line is clean; endpoint is direct-labelled by the plugin
				pointHoverRadius: 6,
				pointBackgroundColor: totalColor,
				pointBorderColor: surface,
				pointBorderWidth: 2
			});
			if (smooth) datasets.push(smoothedDataset(smoothLabel, total, totalColor));
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
				pointRadius: 0,
				pointHoverRadius: 5,
				pointBackgroundColor: s.color,
				pointBorderColor: surface,
				pointBorderWidth: 2
			});
			if (smooth) datasets.push(smoothedDataset(smoothLabel, s.totals, s.color));
		}

		// Direct-label the total line's endpoint (current reading) and mark its peak —
		// the "instrument readout" that carries the page's identity.
		const readoutPlugin: Plugin<'line'> = {
			id: 'trendsReadout',
			afterDatasetsDraw(c) {
				if (!showTotal || total.length === 0) return;
				const dsIndex = c.data.datasets.findIndex((d) => d.label === totalLabel);
				if (dsIndex < 0) return;
				const meta = c.getDatasetMeta(dsIndex);
				if (!meta?.data?.length) return;

				let peakI = 0;
				for (let i = 1; i < total.length; i++) if (total[i] > total[peakI]) peakI = i;
				const lastI = total.length - 1;
				const peakEl = meta.data[peakI] as unknown as Pt;
				const lastEl = meta.data[lastI] as unknown as Pt;

				const ctx = c.ctx;
				ctx.save();

				// Peak marker (skip when it coincides with the endpoint).
				if (peakI !== lastI && peakEl) {
					ctx.beginPath();
					ctx.arc(peakEl.x, peakEl.y, 4, 0, Math.PI * 2);
					ctx.fillStyle = surface;
					ctx.fill();
					ctx.lineWidth = 2;
					ctx.strokeStyle = totalColor;
					ctx.stroke();
					ctx.font = `600 10px ${MONO}`;
					ctx.fillStyle = axisTextColor;
					ctx.textBaseline = 'bottom';
					ctx.textAlign = peakEl.x < 60 ? 'left' : 'center';
					ctx.fillText(`${peakLabel} · ${formatCurrency(total[peakI], 0)}`, peakEl.x, peakEl.y - 9);
				}

				// Endpoint: filled dot + the current value, right-anchored so it never clips.
				if (lastEl) {
					ctx.beginPath();
					ctx.arc(lastEl.x, lastEl.y, 4, 0, Math.PI * 2);
					ctx.fillStyle = totalColor;
					ctx.fill();
					ctx.lineWidth = 2;
					ctx.strokeStyle = surface;
					ctx.stroke();
					ctx.font = `600 12px ${MONO}`;
					ctx.fillStyle = tooltipTitleColor;
					ctx.textBaseline = 'bottom';
					ctx.textAlign = 'right';
					ctx.fillText(formatCurrency(total[lastI], 0), lastEl.x - 6, lastEl.y - 6);
				}

				ctx.restore();
			}
		};

		// Vertical crosshair following the shared-index tooltip.
		const crosshairPlugin: Plugin<'line'> = {
			id: 'trendsCrosshair',
			afterDraw(c) {
				const active = c.tooltip?.getActiveElements?.() ?? [];
				if (active.length === 0) return;
				const x = (active[0].element as unknown as Pt).x;
				const { top, bottom } = c.chartArea;
				const ctx = c.ctx;
				ctx.save();
				ctx.beginPath();
				ctx.moveTo(x, top);
				ctx.lineTo(x, bottom);
				ctx.lineWidth = 1;
				ctx.strokeStyle = totalColor;
				ctx.globalAlpha = 0.35;
				ctx.setLineDash([3, 3]);
				ctx.stroke();
				ctx.restore();
			}
		};

		const config: ChartConfiguration<'line'> = {
			type: 'line',
			data: { labels: monthLabels(), datasets },
			plugins: [readoutPlugin, crosshairPlugin],
			options: {
				responsive: true,
				maintainAspectRatio: false,
				animation: prefersReducedMotion() ? false : { duration: 900, easing: 'easeOutQuart' },
				interaction: { mode: 'index', intersect: false },
				layout: { padding: { top: 26, right: 12 } },
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
						padding: 10,
						bodyFont: { family: MONO },
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
						grid: { display: false },
						border: { display: false },
						ticks: { color: axisTextColor, font: { size: 11, family: MONO } }
					},
					y: {
						grid: { color: gridColor, lineWidth: 1, drawTicks: false },
						border: { display: false },
						ticks: {
							color: axisTextColor,
							font: { size: 11, family: MONO },
							padding: 8,
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
