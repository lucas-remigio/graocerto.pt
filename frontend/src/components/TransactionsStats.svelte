<script lang="ts">
	import type { RecurringRule, TransactionDto } from '$lib/types';
	import { t } from '$lib/i18n';
	import { TransactionTypeId } from '$lib/transaction_types_types';
	import { formatCurrency } from '$lib/utils/currency';

	let {
		transactions = [],
		recurringRules = [],
		summaryItems = []
	}: {
		transactions?: TransactionDto[];
		recurringRules?: RecurringRule[];
		summaryItems?: { amount: number; typeId: number }[];
	} = $props();

	let normalizedSummaryItems = $derived(
		summaryItems.length > 0
			? summaryItems
			: recurringRules.length > 0
				? recurringRules.map((rule) => ({
						amount: rule.amount,
						typeId: rule.transaction_type_id
					}))
				: transactions.map((transaction) => ({
						amount: transaction.amount,
						typeId: transaction.transaction_type.id
					}))
	);

	let totals = $derived(() => {
		const result = {
			debit: 0,
			credit: 0,
			difference: 0
		};

		normalizedSummaryItems.forEach((item) => {
			switch (item.typeId) {
				case TransactionTypeId.Credit:
					result.credit += item.amount;
					result.difference += item.amount;
					break;
				case TransactionTypeId.Debit:
					result.debit += item.amount;
					result.difference -= item.amount;
					break;
			}
		});

		return result;
	});
</script>

<div class="flex divide-x divide-base-300 rounded-xl border border-base-300 bg-base-100 shadow-sm">
	{@render stat($t('transactions.total-credit'), totals().credit, 'text-success', '+')}
	{@render stat($t('transactions.total-debit'), totals().debit, 'text-error', '−')}
	{@render stat(
		$t('transactions.net-balance'),
		Math.abs(totals().difference),
		totals().difference >= 0 ? 'text-success' : 'text-error',
		totals().difference >= 0 ? '+' : '−'
	)}
</div>

<!-- One stat cell: micro label over a monospace, signed value. -->
{#snippet stat(label: string, value: number, colorClass: string, sign: string)}
	<div class="px-4 py-2 text-right">
		<div class="text-[0.65rem] font-medium uppercase tracking-wide text-base-content/55">
			{label}
		</div>
		<div class="font-mono text-sm font-semibold tabular-nums {colorClass}">
			{sign}{formatCurrency(value)}
		</div>
	</div>
{/snippet}
