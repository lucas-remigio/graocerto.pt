<script lang="ts">
	import type { TransactionDto, TransactionsTotals } from '$lib/types';
	import { t } from '$lib/i18n';
	import { TransactionTypeId } from '$lib/transaction_types_types';
	import { formatCurrency } from '$lib/utils/currency';

	let { transactions = [] }: { transactions: TransactionDto[] } = $props();

	let totals = $derived(() => {
		const result = {
			debit: 0,
			credit: 0,
			difference: 0
		};

		transactions.forEach((transaction) => {
			switch (transaction.transaction_type.id) {
				case TransactionTypeId.Credit:
					result.credit += transaction.amount;
					result.difference += transaction.amount;
					break;
				case TransactionTypeId.Debit:
					result.debit += transaction.amount;
					result.difference -= transaction.amount;
					break;
			}
		});

		return result;
	});
</script>

<div class="flex items-center justify-end gap-4 text-sm">
	<div class="stats stats-horizontal shadow-sm">
		<div class="stat px-4 py-2 text-right">
			<div class="stat-title text-right text-xs text-base-content/70">
				{$t('transactions.total-credit')}
			</div>
			<div class="stat-value text-right text-sm text-success">
				+{formatCurrency(totals().credit)}
			</div>
		</div>
		<div class="stat px-4 py-2 text-right">
			<div class="stat-title text-right text-xs text-base-content/70">
				{$t('transactions.total-debit')}
			</div>
			<div class="stat-value text-right text-sm text-error">
				-{formatCurrency(totals().debit)}
			</div>
		</div>
		<div class="stat px-4 py-2 text-right">
			<div class="stat-title text-right text-xs text-base-content/70">
				{$t('transactions.net-balance')}
			</div>
			<div
				class="stat-value text-right text-sm {totals().difference >= 0
					? 'text-success'
					: 'text-error'}"
			>
				{totals().difference >= 0 ? '+' : ''}{formatCurrency(totals().difference)}
			</div>
		</div>
	</div>
</div>
