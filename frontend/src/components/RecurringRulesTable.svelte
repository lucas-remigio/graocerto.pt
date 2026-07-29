<script lang="ts">
	import type { CategoryDto, RecurringRule } from '$lib/types';
	import { t } from '$lib/i18n';
	import { locale } from 'svelte-i18n';
	import { ArrowRightLeft, Pencil, Trash2, Zap } from 'lucide-svelte';
	import { appliedTheme } from '$lib/stores/uiPreferences';
	import { formatCurrency } from '$lib/utils/currency';
	import { getContrastTextClass } from '$lib/utils/categoryUtils';
	import { TransactionTypeId } from '$lib/transaction_types_types';
	import { createEventDispatcher } from 'svelte';

	let {
		recurringRules = [],
		categories = []
	}: { recurringRules?: RecurringRule[]; categories?: CategoryDto[] } = $props();

	const dispatch = createEventDispatcher<{
		editRule: { rule: RecurringRule };
		deleteRule: { ruleId: number };
		generateRule: { rule: RecurringRule };
	}>();

	function getTypeSlug(rule: RecurringRule): 'debit' | 'credit' {
		return rule.transaction_type_id === TransactionTypeId.Debit ? 'debit' : 'credit';
	}

	function getFrequencyLabel(rule: RecurringRule): string {
		switch (rule.frequency) {
			case 'daily':
				return $t('recurring.frequency-daily');
			case 'weekly':
				return $t('recurring.frequency-weekly');
			case 'monthly':
				return $t('recurring.frequency-monthly');
			case 'every_x_days':
				return $t('recurring.frequency-every-x-days');
			default:
				return rule.frequency;
		}
	}

	function getCategoryById(categoryId: number): CategoryDto | undefined {
		return categories.find((category) => category.id === categoryId);
	}

	function getParentCategoryName(category: CategoryDto | undefined): string | null {
		if (!category?.parent_category_id) return null;
		if (category.parent_category?.category_name) return category.parent_category.category_name;
		return getCategoryById(category.parent_category_id)?.category_name || null;
	}

	function getRuleCategory(rule: RecurringRule): CategoryDto | undefined {
		return getCategoryById(rule.category_id);
	}

	function getRuleRowClass(rule: RecurringRule): string {
		const type = getTypeSlug(rule);
		if ($appliedTheme === 'dark') {
			return type === 'debit' ? 'bg-red-900/40' : 'bg-green-900/40';
		}
		return type === 'debit' ? 'bg-red-100' : 'bg-green-100';
	}

	let currentLocale = $derived($locale || 'pt');

	function formatNextRunDate(date: string): string {
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
				<th style="width: 15%">{$t('recurring.frequency')}</th>
				<th style="width: 15%">{$t('recurring.next-run')}</th>
				<th style="width: 20%">{$t('transactions.category')}</th>
				<th style="width: 15%">{$t('transactions.amount')}</th>
				<th style="width: 20%">{$t('transactions.description')}</th>
				<th style="width: 10%">{$t('recurring.interval')}</th>
				<th style="width: 10%">{$t('transactions.actions')}</th>
			</tr>
		</thead>
		<tbody class="text-center">
			{#each recurringRules as rule (rule.id)}
				<tr class={getRuleRowClass(rule)}>
					<td class="text-base-content">
						<span class="badge badge-ghost">{getFrequencyLabel(rule)}</span>
					</td>
					<td class="text-base-content">{formatNextRunDate(rule.next_run_date)}</td>
					<td class="text-base-content">
						<div class="flex flex-col items-center gap-0.5">
							{#if getParentCategoryName(getRuleCategory(rule))}
								<span class="text-xs opacity-90">
									{getParentCategoryName(getRuleCategory(rule))}
								</span>
							{/if}
							<span
								class="inline-flex items-center justify-center rounded px-3 py-1 {getRuleCategory(rule)
									? getContrastTextClass(getRuleCategory(rule)!.color)
									: 'text-gray-900'}"
								style="background-color: {getRuleCategory(rule)?.color || '#d1d5db'}; min-width: 4rem;"
							>
								{getRuleCategory(rule)?.category_name || 'N/A'}
							</span>
						</div>
					</td>
					<td class="text-base-content">
						<div class="flex items-center justify-center gap-2">
							{#if rule.recurring_transfer_group_id}
								<span class="tooltip" data-tip={$t('transactions.transfer')}>
									<ArrowRightLeft size={14} class="text-info" />
								</span>
							{/if}
							<span>{formatCurrency(rule.amount)}</span>
						</div>
					</td>
					<td class="text-base-content">{rule.description || 'N/A'}</td>
					<td class="font-medium text-base-content">{rule.interval_value}</td>
					<td class="text-base-content">
						<div class="flex items-center justify-center gap-x-2">
							<button
								class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-primary backdrop-blur-sm hover:bg-primary/20"
								aria-label={$t('recurring.generate-now')}
								title={$t('recurring.generate-now')}
								onclick={() => dispatch('generateRule', { rule })}
							>
								<Zap size={18} />
							</button>
							<button
								class="btn btn-circle btn-ghost btn-sm bg-base-100/80 backdrop-blur-sm"
								onclick={() => dispatch('editRule', { rule })}
							>
								<Pencil size={18} />
							</button>
							<button
								class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-error backdrop-blur-sm hover:bg-error/20"
								onclick={() => dispatch('deleteRule', { ruleId: rule.id })}
							>
								<Trash2 size={18} />
							</button>
						</div>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
