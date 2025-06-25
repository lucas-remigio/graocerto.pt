<script lang="ts">
	import { t } from '$lib/i18n';
	import { TrendingUp } from 'lucide-svelte';
</script>

<div class="mb-8 text-center">
	<h1 class="text-primary mb-4 flex items-center justify-center gap-3 text-3xl font-bold">
		<TrendingUp size={32} class="text-primary" />
		{$t('investment-calculator.title')}
	</h1>
	<p class="text-base-content/70">
		{$t('investment-calculator.description')}
	</p>
</div>
