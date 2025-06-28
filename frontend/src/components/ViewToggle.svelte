<!-- src/components/ViewToggle.svelte -->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { BarChart3, List } from 'lucide-svelte';
	import { t } from '$lib/i18n';

	export let currentView: 'transactions' | 'statistics' = 'transactions';

	const dispatch = createEventDispatcher<{
		viewChange: { view: 'transactions' | 'statistics' };
	}>();

	function handleViewChange(view: 'transactions' | 'statistics') {
		if (view !== currentView) {
			currentView = view;
			dispatch('viewChange', { view });
		}
	}
</script>

<div class="mb-2 flex justify-center">
	<div class="btn-group">
		<button
			class="btn btn-sm {currentView === 'transactions' ? 'btn-primary' : 'btn-outline'}"
			aria-label="View Transactions"
			on:click={() => handleViewChange('transactions')}
		>
			<List size={16} class="mr-1" />
			<span>{$t('views.transactions')}</span>
		</button>
		<button
			class="btn btn-sm {currentView === 'statistics' ? 'btn-primary' : 'btn-outline'}"
			aria-label="View Statistics"
			on:click={() => handleViewChange('statistics')}
		>
			<BarChart3 size={16} class="mr-1" />
			<span>{$t('views.statistics')}</span>
		</button>
	</div>
</div>
