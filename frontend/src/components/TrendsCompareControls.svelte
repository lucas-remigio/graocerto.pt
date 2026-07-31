<script lang="ts">
	// Two month pickers — [baseline] → [current] — that drive every month-over-month
	// surface (the movers list + the detail table). Pure UI: binds the two indices
	// back to the page, which derives the movers. Hidden with fewer than 2 months.
	import { t } from '$lib/i18n';
	import { locale } from 'svelte-i18n';
	import { ArrowRight } from 'lucide-svelte';
	import type { TrendMonth } from '$lib/types';

	let {
		months,
		baseIdx = $bindable(),
		curIdx = $bindable(),
		onCompareChange
	}: {
		months: TrendMonth[];
		baseIdx: number;
		curIdx: number;
		onCompareChange: () => void;
	} = $props();

	let currentLocale = $derived($locale || 'pt');
	const label = (m: TrendMonth) =>
		new Date(m.year, m.month - 1, 1).toLocaleDateString(currentLocale, {
			month: 'short',
			year: '2-digit'
		});
</script>

{#if months.length >= 2}
	<div class="flex flex-wrap items-center gap-1.5">
		<span class="text-xs uppercase tracking-wide text-base-content/60">
			{$t('trends.compare-label')}
		</span>
		<select
			class="select select-bordered h-9 min-h-0 text-sm"
			bind:value={baseIdx}
			onchange={onCompareChange}
		>
			{#each months as m, i (i)}
				<option value={i}>{label(m)}</option>
			{/each}
		</select>
		<ArrowRight size={14} class="text-base-content/50" />
		<select
			class="select select-bordered h-9 min-h-0 text-sm"
			bind:value={curIdx}
			onchange={onCompareChange}
		>
			{#each months as m, i (i)}
				<option value={i}>{label(m)}</option>
			{/each}
		</select>
	</div>
{/if}
