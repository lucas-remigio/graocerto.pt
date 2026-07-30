<script lang="ts">
	// "What's changing" — category momentum. Direction/pct/sparkline data are all
	// computed server-side (CategoryMover); this only renders them.
	import { t } from '$lib/i18n';
	import { TrendingUp, TrendingDown } from 'lucide-svelte';
	import CategorySparkline from '$components/CategorySparkline.svelte';
	import type { CategoryMover } from '$lib/types';

	let { movers }: { movers: CategoryMover[] } = $props();
</script>

{#if movers.length > 0}
	<div class="mt-4 rounded-xl border border-base-300 bg-base-100 p-4 shadow-sm">
		<div class="mb-3">
			<p class="text-sm font-semibold">{$t('trends.whats-changing')}</p>
			<p class="text-xs text-base-content/50">{$t('trends.momentum-caption')}</p>
		</div>
		<div class="grid gap-x-6 gap-y-3 sm:grid-cols-2">
			{#each movers as m (m.id)}
				<div class="flex items-center gap-3">
					<span class="h-2.5 w-2.5 shrink-0 rounded-full" style="background-color: {m.color};"></span>
					<span class="min-w-0 flex-1 truncate text-sm">{m.name}</span>
					<CategorySparkline data={m.totals} color={m.color} />
					<span
						class="flex items-center gap-1 rounded-full bg-base-200 px-2 py-0.5 text-xs font-semibold text-base-content/80"
					>
						{#if m.direction === 'new'}
							{$t('trends.momentum-new')}
						{:else if m.direction === 'up'}
							<TrendingUp size={13} />
							+{m.pct}%
						{:else}
							<TrendingDown size={13} />
							{m.pct}%
						{/if}
					</span>
				</div>
			{/each}
		</div>
	</div>
{/if}
