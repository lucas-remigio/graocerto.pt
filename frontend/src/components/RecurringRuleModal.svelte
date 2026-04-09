<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { X } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { dataService } from '$lib/services/dataService';
	import type {
		Account,
		CategoryDto,
		CreateRecurringRulePayload,
		RecurringRule,
		UpdateRecurringRulePayload
	} from '$lib/types';
	import { TransactionTypeId, TransactionTypeSlug, TransactionTypes } from '$lib/transaction_types_types';
	import { buildCategoryGroups } from '$lib/utils/categoryUtils';

	const dispatch = createEventDispatcher();
	let {
		account,
		recurringRule = null
	}: { account: Account; recurringRule?: RecurringRule | null } = $props();

	let error = $state('');
	let isLoading = $state(true);
	let categories: CategoryDto[] = $state([]);
	let transaction_type_id: number = $state(TransactionTypeId.Debit);
	let filteredCategories = $derived(
		categories.filter((c) => c.transaction_type.id === transaction_type_id)
	);
	let selectedCategory = $derived(
		filteredCategories.find((category) => category.id === Number(form.category_id))
	);
	let categoryBorderColor = $derived(selectedCategory?.color || '#ccc');
	let groupedCategories = $derived(buildCategoryGroups(filteredCategories));
	const borderClasses: Record<string, string> = {
		credit: 'border-green-500 dark:border-green-400',
		debit: 'border-red-500 dark:border-red-400'
	};
	let selectedTypeSlug = $derived(
		transaction_type_id === TransactionTypeId.Credit ? 'credit' : 'debit'
	);
	let modalBorderClass = $derived(
		borderClasses[selectedTypeSlug] || 'border-blue-500 dark:border-blue-400'
	);
	$effect(() => {
		form.transaction_type_id = transaction_type_id;
		if (!filteredCategories.find((c) => c.id === form.category_id) && filteredCategories.length > 0) {
			form.category_id = filteredCategories[0].id;
		}
		if (form.frequency === 'weekly' && form.execution_weekday === undefined) {
			form.execution_weekday = new Date().getUTCDay();
		}
	});

	let form: CreateRecurringRulePayload = $state({
		account_token: '',
		category_id: 0,
		transaction_type_id: TransactionTypeId.Debit,
		amount: 0,
		description: '',
		frequency: 'monthly',
		interval_value: 1,
		active: true
	});
	const isEditMode = $derived(recurringRule !== null);

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function isFormValid(): boolean {
		if (!form.category_id) {
			error = $t('transactions.category-required');
			return false;
		}
		if (!form.amount || form.amount <= 0) {
			error = $t('transactions.amount-greater-zero');
			return false;
		}
		if (!form.interval_value || form.interval_value < 1) {
			error = $t('common.invalid');
			return false;
		}
		if (form.frequency === 'monthly' && (!form.execution_day || form.execution_day < 1 || form.execution_day > 31)) {
			error = $t('recurring.execution-day-help');
			return false;
		}
		if (
			form.frequency === 'weekly' &&
			(form.execution_weekday === undefined ||
				form.execution_weekday < 0 ||
				form.execution_weekday > 6)
		) {
			error = $t('recurring.execution-weekday-help');
			return false;
		}
		return true;
	}

	async function handleSubmit() {
		error = '';
		if (!isFormValid()) return;

		try {
			form.transaction_type_id = transaction_type_id;
			if (recurringRule) {
				const payload: UpdateRecurringRulePayload = {
					account_token: form.account_token,
					category_id: form.category_id,
					transaction_type_id: form.transaction_type_id,
					amount: form.amount,
					description: form.description,
					frequency: form.frequency,
					interval_value: form.interval_value,
					active: !!form.active
				};

				if (form.frequency === 'monthly' && form.execution_day) {
					payload.execution_day = form.execution_day;
				}
				if (form.frequency === 'weekly' && form.execution_weekday !== undefined) {
					payload.execution_weekday = form.execution_weekday;
				}

				const updatedRule = await dataService.updateRecurringRule(recurringRule.id, payload);
				dispatch('updateRecurringRule', updatedRule as RecurringRule);
				return;
			}

			const createdRule = await dataService.createRecurringRule(form);
			dispatch('newRecurringRule', createdRule as RecurringRule);
		} catch (err) {
			console.error('Error saving recurring rule:', err);
			error = $t('errors.server-error');
		}
	}

	async function fetchData() {
		isLoading = true;
		error = '';
		try {
			const categoriesData = await dataService.fetchCategories();
			categories = categoriesData;
			form.account_token = account.token;
			if (recurringRule) {
				const monthlyExecutionDay =
					recurringRule.frequency === 'monthly'
						? new Date(recurringRule.next_run_date).getUTCDate()
						: undefined;
				const weeklyExecutionWeekday =
					recurringRule.frequency === 'weekly'
						? new Date(recurringRule.next_run_date).getUTCDay()
						: undefined;
				transaction_type_id = recurringRule.transaction_type_id;
				form = {
					account_token: recurringRule.account_token,
					category_id: recurringRule.category_id,
					transaction_type_id: recurringRule.transaction_type_id,
					amount: recurringRule.amount,
					description: recurringRule.description,
					frequency: recurringRule.frequency,
					interval_value: recurringRule.interval_value,
					execution_day: monthlyExecutionDay,
					execution_weekday: weeklyExecutionWeekday,
					active: recurringRule.active
				};
			}
			if (!form.category_id && filteredCategories.length > 0) form.category_id = filteredCategories[0].id;
		} catch (err) {
			console.error('Error loading recurring modal data:', err);
			error = $t('errors.failed-load-data');
		} finally {
			isLoading = false;
		}
	}

	onMount(fetchData);
</script>

<div class="modal modal-open">
	<div class="modal-box relative border-4 {modalBorderClass}">
		<button class="btn btn-circle btn-sm absolute right-2 top-2" onclick={handleCloseModal}>
			<X />
		</button>

		<h3 class="mb-4 text-lg font-bold">
			{isEditMode ? $t('recurring.edit-recurring-payment') : $t('recurring.new-recurring-payment')}
		</h3>

		{#if error}
			<div class="alert alert-error mb-4">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}

		{#if isLoading}
			<div class="py-12 text-center">
				<div class="loading loading-spinner loading-lg mx-auto mb-4"></div>
				<p class="text-base-content/70">{$t('common.loading')}</p>
			</div>
		{:else}
			<form
				onsubmit={(event) => {
					event.preventDefault();
					handleSubmit();
				}}
			>
				<div class="form-control mt-4">
					<div class="label">
						<span class="label-text">{$t('accounts.select-account')}</span>
					</div>
					<div class="input input-bordered flex w-full items-center bg-base-200/60">
						{account.account_name}
					</div>
				</div>

				<div class="mt-4 flex flex-col gap-4 md:flex-row">
					<div class="form-control flex-1">
						<label class="label" for="transaction-type">
							<span class="label-text">{$t('transactions.transaction-type')}</span>
						</label>
						<select id="transaction-type" class="select select-bordered w-full" bind:value={transaction_type_id}>
							{#each TransactionTypes.filter((tt) => tt.type_slug !== TransactionTypeSlug.Transfer) as type}
								<option value={type.id}>{$t('transaction-types.' + type.type_slug)}</option>
							{/each}
						</select>
					</div>

					<div class="form-control flex-1">
						<label class="label" for="category">
							<span class="label-text">{$t('transactions.category')}</span>
						</label>
						<select
							id="category"
							class="select select-bordered w-full border-2"
							bind:value={form.category_id}
							required
							style="border-color: {categoryBorderColor} !important;"
						>
							{#each groupedCategories as group}
								{#if group.parent}
									{#if group.children.length > 0}
										<option value={group.parent.id} class="font-semibold">
											{group.parent.category_name}
										</option>
										{#each group.children as child}
											<option value={child.id}>
												&nbsp;&nbsp;&nbsp;&nbsp;{child.category_name}
											</option>
										{/each}
									{:else}
										<option value={group.parent.id}>
											{group.parent.category_name}
										</option>
									{/if}
								{:else}
									{#each group.children as child}
										<option value={child.id}>{child.category_name}</option>
									{/each}
								{/if}
							{/each}
						</select>
					</div>
				</div>

				<div class="form-control mt-4">
					<label class="label" for="description">
						<span class="label-text">{$t('transactions.description')}</span>
					</label>
					<input
						id="description"
						type="text"
						class="input input-bordered"
						placeholder={$t('recurring.description-placeholder')}
						bind:value={form.description}
					/>
				</div>

				<div class="mt-4 flex gap-4">
					<div class="form-control flex-1">
						<label class="label" for="amount">
							<span class="label-text">{$t('transactions.amount')}</span>
						</label>
						<input
							id="amount"
							type="number"
							min="0.01"
							step="0.01"
							max="999999999"
							class="input input-bordered w-full"
							placeholder={$t('transactions.transaction-amount')}
							bind:value={form.amount}
						/>
					</div>
				</div>

				<div class="mt-4 flex gap-4">
					<div class="form-control flex-1">
						<label class="label" for="frequency">
							<span class="label-text">{$t('recurring.frequency')}</span>
						</label>
						<select id="frequency" class="select select-bordered w-full" bind:value={form.frequency}>
							<option value="daily">{$t('recurring.frequency-daily')}</option>
							<option value="weekly">{$t('recurring.frequency-weekly')}</option>
							<option value="monthly">{$t('recurring.frequency-monthly')}</option>
							<option value="every_x_days">{$t('recurring.frequency-every-x-days')}</option>
						</select>
					</div>

					<div class="form-control flex-1">
						<label class="label" for="interval">
							<span class="label-text">{$t('recurring.interval')}</span>
						</label>
						<input id="interval" type="number" min="1" step="1" class="input input-bordered w-full" bind:value={form.interval_value} />
					</div>
				</div>

				{#if form.frequency === 'monthly'}
					<div class="form-control mt-4">
						<label class="label" for="execution-day">
							<span class="label-text">{$t('recurring.execution-day')}</span>
						</label>
						<input
							id="execution-day"
							type="number"
							min="1"
							max="31"
							step="1"
							class="input input-bordered w-full"
							placeholder="1"
							bind:value={form.execution_day}
						/>
						<div class="label">
							<span class="label-text-alt">{$t('recurring.execution-day-help')}</span>
						</div>
						{#if form.execution_day && form.execution_day >= 29}
							<div class="label">
								<span class="label-text-alt text-warning">
									{$t('recurring.execution-day-end-of-month-note')}
								</span>
							</div>
						{/if}
					</div>
				{/if}

				{#if form.frequency === 'weekly'}
					<div class="form-control mt-4">
						<label class="label" for="execution-weekday">
							<span class="label-text">{$t('recurring.execution-weekday')}</span>
						</label>
						<select
							id="execution-weekday"
							class="select select-bordered w-full"
							bind:value={form.execution_weekday}
						>
							<option value={0}>{$t('recurring.weekday-sunday')}</option>
							<option value={1}>{$t('recurring.weekday-monday')}</option>
							<option value={2}>{$t('recurring.weekday-tuesday')}</option>
							<option value={3}>{$t('recurring.weekday-wednesday')}</option>
							<option value={4}>{$t('recurring.weekday-thursday')}</option>
							<option value={5}>{$t('recurring.weekday-friday')}</option>
							<option value={6}>{$t('recurring.weekday-saturday')}</option>
						</select>
						<div class="label">
							<span class="label-text-alt">{$t('recurring.execution-weekday-help')}</span>
						</div>
					</div>
				{/if}

				<div class="form-control mt-4">
					<label class="label cursor-pointer justify-start gap-2">
						<input class="toggle" type="checkbox" bind:checked={form.active} />
						<span>{$t('recurring.active')}</span>
					</label>
				</div>

				<div class="modal-action mt-6">
					<button type="button" class="btn" onclick={handleCloseModal}>{$t('common.cancel')}</button>
					<button type="submit" class="btn btn-primary text-base-100">
						{isEditMode ? $t('transactions.update-transaction') : $t('common.create')}
					</button>
				</div>
			</form>
		{/if}
	</div>
</div>

