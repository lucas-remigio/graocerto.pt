<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Chart, registerables } from 'chart.js';
	import { PieChart } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import type { CategoryStatistic } from '$lib/types';
	import { themeService } from '$lib/services/themeService';

	export let data: CategoryStatistic[] = [];

	let canvas: HTMLCanvasElement;
	let chart: Chart | null = null;
	let unsubscribeTheme: (() => void) | null = null;
	let selectedParent: CategoryStatistic | null = null;

	Chart.register(...registerables);

	// Sort parent data by percentage (descending)
	$: sortedData = [...(data ?? [])].sort((a, b) => b.percentage - a.percentage);

	// Parent-level data (only root categories)
	$: parentData = sortedData.map((cat) => ({
		name: cat.name,
		total: cat.total,
		percentage: cat.percentage,
		color: cat.color,
		count: cat.count,
		hasChildren: cat.subcategories && cat.subcategories.length > 0
	}));

	// Drill-down data when a parent is selected (also sorted)
	$: drillDownData = selectedParent
		? [
				// Parent direct spending (if any)
				...(selectedParent.count >
				(selectedParent.subcategories?.reduce((sum, sub) => sum + sub.count, 0) || 0)
					? [
							{
								name: `${selectedParent.name}`,
								total:
									selectedParent.total -
									(selectedParent.subcategories?.reduce((sum, sub) => sum + sub.total, 0) || 0),
								percentage: 0,
								color: selectedParent.color,
								count:
									selectedParent.count -
									(selectedParent.subcategories?.reduce((sum, sub) => sum + sub.count, 0) || 0)
							}
						]
					: []),
				// Subcategories (sorted by percentage)
				...(selectedParent.subcategories
					?.slice()
					.sort((a, b) => b.percentage - a.percentage)
					.map((sub) => ({
						name: sub.name,
						total: sub.total,
						percentage: sub.percentage,
						color: sub.color,
						count: sub.count
					})) || [])
			]
		: null;

	function createChart() {
		if (!canvas || !data?.length) return;

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
					if (!drillDownData && elements.length > 0) {
						const index = elements[0].index;
						const clickedCategory = sortedData[index];
						if (clickedCategory.subcategories && clickedCategory.subcategories.length > 0) {
							selectedParent = clickedCategory;
						}
					}
				},
				onHover: (event, elements) => {
					const target = event.native?.target as HTMLElement | null;
					if (!target) return;

					if (!drillDownData && elements.length > 0) {
						const index = elements[0].index;
						const hoveredCategory = sortedData[index];
						target.style.cursor =
							hoveredCategory.subcategories && hoveredCategory.subcategories.length > 0
								? 'pointer'
								: 'default';
					} else if (drillDownData) {
						target.style.cursor = 'pointer';
					} else {
						target.style.cursor = 'default';
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
	{#if data?.length > 0}
		<!-- Chart Title -->
		<div class="mb-3 text-center">
			<h4 class="text-sm font-medium opacity-70">
				{#if selectedParent}
					<button
						class="btn btn-ghost btn-xs mr-2"
						on:click={() => (selectedParent = null)}
						aria-label="Back to overview"
					>
						← {$t('common.back', { default: 'Back' })}
					</button>
					{selectedParent.name}
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
					<div class="flex items-center justify-between rounded bg-base-200 p-2">
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
						class="flex w-full items-center justify-between rounded bg-base-200 p-2 transition-colors hover:bg-base-300"
						class:cursor-pointer={item.hasChildren}
						class:cursor-default={!item.hasChildren}
						on:click={() => {
							if (item.hasChildren) {
								selectedParent = sortedData[index];
							}
						}}
						disabled={!item.hasChildren}
						class:tooltip={item.hasChildren}
						class:tooltip-top={item.hasChildren}
						data-tip={item.hasChildren
							? $t('statistics.click-to-view-details', { default: 'Click to view details' })
							: ''}
					>
						<div class="flex items-center gap-2">
							<span class="h-3 w-3 rounded-full" style="background-color: {item.color};"></span>
							<span class="text-sm">{item.name}</span>
							{#if item.hasChildren}
								<span class="badge badge-xs">
									{sortedData[index].subcategories?.length || 0}
									<!-- ← Use sortedData -->
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
				<PieChart size={48} class="mx-auto mb-3 text-base-content/30" />
				<p class="text-base font-medium text-base-content/60">{$t('statistics.no-data')}</p>
			</div>
		</div>
	{/if}
</div>
