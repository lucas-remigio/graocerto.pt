<script lang="ts">
	import type { CategoryDto, RecurringForecastItem } from '$lib/types';
	import { t } from '$lib/i18n';
	import { locale } from 'svelte-i18n';
	import { ArrowRightLeft } from 'lucide-svelte';
	import { formatCurrency } from '$lib/utils/currency';
	import { getContrastTextClass } from '$lib/utils/categoryUtils';
	import { TransactionTypeId } from '$lib/transaction_types_types';
	import { appliedTheme } from '$lib/stores/uiPreferences';

	let {
		items = [],
		categories = []
	}: { items?: RecurringForecastItem[]; categories?: CategoryDto[] } = $props();

	function getCategoryById(categoryId: number): CategoryDto | undefined {
		return categories.find((category) => category.id === categoryId);
	}

	function getParentCategoryName(category: CategoryDto | undefined): string | null {
		if (!category?.parent_category_id) return null;
		if (category.parent_category?.category_name) return category.parent_category.category_name;
		return getCategoryById(category.parent_category_id)?.category_name || null;
	}

	function getRowClass(item: RecurringForecastItem): string {
		const isDebit = item.transaction_type_id === TransactionTypeId.Debit;
		if ($appliedTheme === 'dark') {
			return isDebit ? 'bg-red-900/40' : 'bg-green-900/40';
		}
		return isDebit ? 'bg-red-100' : 'bg-green-100';
	}

	let currentLocale = $derived($locale || 'pt');

	function formatForecastDate(date: string): string {
		return new Date(date).toLocaleDateString(currentLocale, {
			day: 'numeric',
			month: 'long'
		});
	}
</script>

<div class="overflow-x-auto">
	<table class="table w-full">
		<thead class="sticky top-0 text-center">
			<tr>
				<th style="width: 15%">{$t('transactions.date')}</th>
				<th style="width: 20%">{$t('transactions.category')}</th>
				<th style="width: 15%">{$t('transactions.amount')}</th>
				<th style="width: 40%">{$t('transactions.description')}</th>
			</tr>
		</thead>
		<tbody class="text-center">
			{#each items as item (`${item.recurring_rule_id}-${item.date}`)}
				<tr class={getRowClass(item)}>
					<td class="text-base-content">{formatForecastDate(item.date)}</td>
					<td class="text-base-content">
						<div class="flex flex-col items-center gap-0.5">
							{#if getParentCategoryName(getCategoryById(item.category_id))}
								<span class="text-xs opacity-90">
									{getParentCategoryName(getCategoryById(item.category_id))}
								</span>
							{/if}
							<span
								class="inline-flex items-center justify-center rounded px-3 py-1 {getCategoryById(
									item.category_id
								)
									? getContrastTextClass(getCategoryById(item.category_id)!.color)
									: 'text-gray-900'}"
								style="background-color: {getCategoryById(item.category_id)?.color ||
									'#d1d5db'}; min-width: 4rem;"
							>
								{getCategoryById(item.category_id)?.category_name || 'N/A'}
							</span>
						</div>
					</td>
					<td class="text-base-content">
						<div class="flex items-center justify-center gap-2">
							{#if item.recurring_transfer_group_id}
								<span class="tooltip" data-tip={$t('transactions.transfer')}>
									<ArrowRightLeft size={14} class="text-info" />
								</span>
							{/if}
							<span>{formatCurrency(item.amount)}</span>
						</div>
					</td>
					<td class="text-base-content">{item.description || 'N/A'}</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
