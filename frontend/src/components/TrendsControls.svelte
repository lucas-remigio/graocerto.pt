<script lang="ts">
	// Spending/income + time-range toggles for the trends page. Pure UI: it only
	// binds the two selections back to the page, which owns fetching.
	import { t } from '$lib/i18n';
	import { TrendingUp, TrendingDown } from 'lucide-svelte';
	import type { TrendsRangeMonths, TrendsType } from '$lib/types';

	let {
		type = $bindable(),
		months = $bindable()
	}: {
		type: TrendsType;
		months: TrendsRangeMonths;
	} = $props();

	const ranges: TrendsRangeMonths[] = [6, 12, 24];
</script>

<div class="mb-4 flex flex-wrap items-center gap-3">
	<div class="btn-group">
		<button
			class="btn btn-sm gap-1 {type === 'debit' ? 'btn-primary text-base-100' : 'btn-ghost'}"
			onclick={() => (type = 'debit')}
		>
			<TrendingDown size={15} />
			{$t('trends.spending')}
		</button>
		<button
			class="btn btn-sm gap-1 {type === 'credit' ? 'btn-primary text-base-100' : 'btn-ghost'}"
			onclick={() => (type = 'credit')}
		>
			<TrendingUp size={15} />
			{$t('trends.income')}
		</button>
	</div>

	<div class="btn-group ml-auto">
		{#each ranges as r (r)}
			<button
				class="btn btn-sm {months === r ? 'btn-primary text-base-100' : 'btn-ghost'}"
				onclick={() => (months = r)}
			>
				{r}M
			</button>
		{/each}
	</div>
</div>
