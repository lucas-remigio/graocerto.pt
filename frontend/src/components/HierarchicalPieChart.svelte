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
	let selectedParent: CategoryStatistic | null = null;

	Chart.register(...registerables);

	// Parent-level data (only root categories)
	$: parentData = data.map((cat) => ({
		name: cat.name,
		total: cat.total,
		percentage: cat.percentage,
		color: cat.color,
		count: cat.count,
		hasChildren: cat.subcategories && cat.subcategories.length > 0
	}));

	// Drill-down data when a parent is selected
	$: drillDownData = selectedParent
		? [
				// Parent direct spending (if any)
				...(selectedParent.count >
				(selectedParent.subcategories?.reduce((sum, sub) => sum + sub.count, 0) || 0)
					? [
							{
								name: `${selectedParent.name} (Direct)`,
								total:
									selectedParent.total -
									(selectedParent.subcategories?.reduce((sum, sub) => sum + sub.total, 0) || 0),
								percentage: 0, // Will recalculate
								color: selectedParent.color,
								count:
									selectedParent.count -
									(selectedParent.subcategories?.reduce((sum, sub) => sum + sub.count, 0) || 0)
							}
						]
					: []),
				// Subcategories
				...(selectedParent.subcategories?.map((sub) => ({
					name: sub.name,
					total: sub.total,
					percentage: 0, // Will recalculate
					color: sub.color,
					count: sub.count
				})) || [])
			]
		: null;

	function createChart() {
		if (!canvas || data.length === 0) return;

		if (chart) {
			chart.destroy();
		}

		const themeColors = themeService.getThemeColors();
		const { legendColor, tooltipBg, tooltipTitleColor, tooltipBodyColor, tooltipBorderColor } =
			themeColors;

		const displayData = drillDownData || parentData;
		const labels = displayData.map((item) => item.name);
		const values = displayData.map((item) => item.total);
		const colors = displayData.map((item) => item.color);

		chart = new Chart(canvas, {
			type: 'pie',
			data: {
				labels,
				datasets: [
					{
						data: values,
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
				onClick: (event, elements) => {
					if (drillDownData) {
						// In drill-down mode, clicking goes back
						selectedParent = null;
					} else if (elements.length > 0) {
						// In parent mode, clicking drills down if has children
						const index = elements[0].index;
						const clickedCategory = data[index];
						if (clickedCategory.subcategories && clickedCategory.subcategories.length > 0) {
							selectedParent = clickedCategory;
						}
					}
				},
				plugins: {
					legend: {
						display: false // We'll create custom legend
					},
					tooltip: {
						backgroundColor: tooltipBg,
						titleColor: tooltipTitleColor,
						bodyColor: tooltipBodyColor,
						borderColor: tooltipBorderColor,
						borderWidth: 1,
						callbacks: {
							label: function (context) {
								const item = displayData[context.dataIndex];
								const total = values.reduce((a, b) => a + b, 0);
								const percentage = ((item.total / total) * 100).toFixed(1);
								return [
									`${item.name}`,
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

	$: if (canvas && (parentData || drillDownData)) {
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
	{#if data.length > 0}
		<!-- Chart Title -->
		<div class="mb-3 text-center">
			<h4 class="text-sm font-medium opacity-70">
				{#if selectedParent}
					{selectedParent.name}
					<button
						class="btn btn-ghost btn-xs ml-2"
						on:click={() => (selectedParent = null)}
						aria-label="Back to overview"
					>
						← {$t('common.back', { default: 'Back' })}
					</button>
				{:else}
					{$t('statistics.main-categories', { default: 'Main Categories' })}
				{/if}
			</h4>
		</div>

		<!-- Chart -->
		<div class="h-64 w-full">
			<canvas bind:this={canvas} class="max-h-full max-w-full"></canvas>
		</div>

		<!-- Legend -->
		<div class="mt-4 space-y-2">
			{#if drillDownData}
				<!-- Drill-down legend -->
				{#each drillDownData as item}
					<div class="bg-base-200 flex items-center justify-between rounded p-2">
						<div class="flex items-center gap-2">
							<span class="h-3 w-3 rounded-full" style="background-color: {item.color};"></span>
							<span class="text-sm">{item.name}</span>
							<span class="text-xs opacity-60">({item.count})</span>
						</div>
						<div class="text-right">
							<span class="text-sm font-semibold">€{item.total.toFixed(2)}</span>
						</div>
					</div>
				{/each}
			{:else}
				<!-- Parent-level legend -->
				{#each parentData as item, index}
					<button
						class="bg-base-200 hover:bg-base-300 flex w-full items-center justify-between rounded p-2 transition-colors"
						class:cursor-pointer={item.hasChildren}
						on:click={() => {
							if (item.hasChildren) {
								selectedParent = data[index];
							}
						}}
						disabled={!item.hasChildren}
					>
						<div class="flex items-center gap-2">
							<span class="h-3 w-3 rounded-full" style="background-color: {item.color};"></span>
							<span class="text-sm">{item.name}</span>
							{#if item.hasChildren}
								<span class="badge badge-xs">
									{data[index].subcategories?.length || 0}
								</span>
							{/if}
						</div>
						<div class="text-right">
							<span class="text-sm font-semibold">€{item.total.toFixed(2)}</span>
							<span class="ml-2 text-xs opacity-60">{item.percentage.toFixed(1)}%</span>
						</div>
					</button>
				{/each}
			{/if}
		</div>
	{:else}
		<!-- Empty state -->
		<div class="flex h-64 items-center justify-center">
			<div class="text-center">
				<PieChart size={48} class="text-base-content/30 mx-auto mb-3" />
				<p class="text-base-content/60 text-base font-medium">{$t('statistics.no-data')}</p>
			</div>
		</div>
	{/if}
</div>
