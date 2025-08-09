<!-- src/components/ViewToggle.svelte -->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { BarChart3, List } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { selectedView, updateSelectedView } from '$lib/stores/uiPreferences';

	function handleViewChange(view: 'transactions' | 'statistics') {
		updateSelectedView(view);
	}
</script>

<div class="mb-2 flex justify-center">
	<div class="btn-group">
		<button
			class="btn btn-sm {$selectedView === 'transactions'
				? 'btn-primary text-base-100'
				: 'btn-ghost'}"
			aria-label="View Transactions"
			on:click={() => handleViewChange('transactions')}
		>
			<List size={16} class="mr-1" />
			<span>{$t('views.transactions')}</span>
		</button>
		<button
			class="btn btn-sm {$selectedView === 'statistics'
				? 'btn-primary text-base-100'
				: 'btn-ghost'}"
			aria-label="View Statistics"
			on:click={() => handleViewChange('statistics')}
		>
			<BarChart3 size={16} class="mr-1" />
			<span>{$t('views.statistics')}</span>
		</button>
	</div>
</div>
