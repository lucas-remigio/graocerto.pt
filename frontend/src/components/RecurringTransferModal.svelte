<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { ArrowRight, X } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { dataService } from '$lib/services/dataService';
	import type {
		Account,
		CategoryDto,
		CreateRecurringTransferPayload,
		RecurringRule,
		UpdateRecurringTransferPayload
	} from '$lib/types';
	import { TransactionTypeId } from '$lib/transaction_types_types';
	import { buildCategoryGroups } from '$lib/utils/categoryUtils';
	import { toastStore } from '$lib/stores/toast';

	let {
		account,
		transferGroupId = null,
		initialRules = []
	}: {
		account: Account;
		transferGroupId?: string | null;
		initialRules?: RecurringRule[];
	} = $props();

	const dispatch = createEventDispatcher<{
		closeModal: void;
		newRecurringTransfer: { rules: RecurringRule[] };
		updateRecurringTransfer: { rules: RecurringRule[] };
	}>();

	let isLoading = $state(true);
	let accounts: Account[] = $state([]);
	let debitCategories: CategoryDto[] = $state([]);
	let creditCategories: CategoryDto[] = $state([]);

	let sourceAccountToken = $state('');
	let destinationAccountToken = $state('');
	let debitCategoryId = $state('');
	let creditCategoryId = $state('');
	let amount = $state(0);
	let description = $state('');
	let frequency = $state<CreateRecurringTransferPayload['frequency']>('monthly');
	let intervalValue = $state(1);
	let executionDay = $state<number | undefined>(undefined);
	let executionWeekday = $state<number | undefined>(undefined);
	let active = $state(true);

	let selectedDebitCategory = $derived(
		debitCategories.find((category) => category.id === Number(debitCategoryId))
	);
	let selectedCreditCategory = $derived(
		creditCategories.find((category) => category.id === Number(creditCategoryId))
	);
	let debitBorderColor = $derived(selectedDebitCategory?.color || '#ef4444');
	let creditBorderColor = $derived(selectedCreditCategory?.color || '#22c55e');
	let groupedDebitCategories = $derived(buildCategoryGroups(debitCategories));
	let groupedCreditCategories = $derived(buildCategoryGroups(creditCategories));
	let isEditMode = $derived(!!transferGroupId);

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function isFormValid(): boolean {
		if (!destinationAccountToken) {
			toastStore.error($t('transfers.select-destination-account'));
			return false;
		}
		if (sourceAccountToken === destinationAccountToken) {
			toastStore.error($t('transfers.same-account-error'));
			return false;
		}
		if (!debitCategoryId) {
			toastStore.error($t('transfers.select-debit-category'));
			return false;
		}
		if (!creditCategoryId) {
			toastStore.error($t('transfers.select-credit-category'));
			return false;
		}
		if (amount <= 0) {
			toastStore.error($t('transfers.amount-must-be-positive'));
			return false;
		}
		if (intervalValue < 1) {
			toastStore.error($t('common.invalid'));
			return false;
		}
		if (frequency === 'monthly' && (!executionDay || executionDay < 1 || executionDay > 31)) {
			toastStore.error($t('recurring.execution-day-help'));
			return false;
		}
		if (
			frequency === 'weekly' &&
			(executionWeekday === undefined || executionWeekday < 0 || executionWeekday > 6)
		) {
			toastStore.error($t('recurring.execution-weekday-help'));
			return false;
		}
		return true;
	}

	async function handleSubmit() {
		if (!isFormValid()) return;

		try {
			const payloadBase: CreateRecurringTransferPayload = {
				source_account_token: sourceAccountToken,
				destination_account_token: destinationAccountToken,
				debit_category_id: Number(debitCategoryId),
				credit_category_id: Number(creditCategoryId),
				amount,
				description,
				frequency,
				interval_value: intervalValue,
				active
			};
			if (frequency === 'monthly' && executionDay) {
				payloadBase.execution_day = executionDay;
			}
			if (frequency === 'weekly' && executionWeekday !== undefined) {
				payloadBase.execution_weekday = executionWeekday;
			}

			if (transferGroupId) {
				const updatePayload: UpdateRecurringTransferPayload = {
					...payloadBase,
					active
				};
				const rules = await dataService.updateRecurringTransfer(transferGroupId, updatePayload);
				dispatch('updateRecurringTransfer', { rules });
				return;
			}

			const rules = await dataService.createRecurringTransfer(payloadBase);
			toastStore.success(
				isEditMode
					? $t('recurring.recurring-transfer-updated')
					: $t('recurring.recurring-transfer-created')
			);
			dispatch('newRecurringTransfer', { rules });
		} catch (err) {
			console.error('Error creating recurring transfer:', err);
			toastStore.error($t('errors.server-error'));
		}
	}

	async function fetchData() {
		isLoading = true;
		try {
			accounts = await dataService.fetchAccounts();
			const allCategories = await dataService.fetchCategories();
			debitCategories = allCategories.filter(
				(category) => category.transaction_type.type_slug === 'debit'
			);
			creditCategories = allCategories.filter(
				(category) => category.transaction_type.type_slug === 'credit'
			);

			if (debitCategories.length > 0 && !debitCategoryId) {
				debitCategoryId = String(debitCategories[0].id);
			}
			if (creditCategories.length > 0 && !creditCategoryId) {
				creditCategoryId = String(creditCategories[0].id);
			}

			if (initialRules.length > 0) {
				const debitRule = initialRules.find(
					(rule) => rule.transaction_type_id === TransactionTypeId.Debit
				);
				const creditRule = initialRules.find(
					(rule) => rule.transaction_type_id === TransactionTypeId.Credit
				);

				if (debitRule) {
					sourceAccountToken = debitRule.account_token;
					debitCategoryId = String(debitRule.category_id);
					amount = debitRule.amount;
					description = debitRule.description;
					frequency = debitRule.frequency;
					intervalValue = debitRule.interval_value;
					active = debitRule.active;
					if (debitRule.frequency === 'monthly') {
						executionDay = new Date(debitRule.next_run_date).getUTCDate();
					}
					if (debitRule.frequency === 'weekly') {
						executionWeekday = new Date(debitRule.next_run_date).getUTCDay();
					}
				}

				if (creditRule) {
					destinationAccountToken = creditRule.account_token;
					creditCategoryId = String(creditRule.category_id);
				}
			}
		} catch (err) {
			console.error('Error loading recurring transfer data:', err);
			toastStore.error($t('errors.failed-load-data'));
		} finally {
			isLoading = false;
		}
	}

	onMount(fetchData);

	$effect(() => {
		if (!sourceAccountToken) {
			sourceAccountToken = account.token;
		}
		if (frequency === 'weekly' && executionWeekday === undefined) {
			executionWeekday = new Date().getUTCDay();
		}
	});
</script>

<div class="modal modal-open">
	<div class="modal-box relative max-w-3xl border-4 border-blue-500 dark:border-blue-400">
		<button class="btn btn-circle btn-sm absolute right-2 top-2" onclick={handleCloseModal}>
			<X />
		</button>

		<h3 class="mb-4 text-lg font-bold">
			{isEditMode
				? $t('recurring.edit-recurring-transfer')
				: $t('recurring.new-recurring-transfer')}
		</h3>

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
				<div class="flex flex-col gap-4 md:flex-row md:items-center">
					<div class="flex-1 space-y-4">
						<div class="form-control">
							<label class="label" for="source-account">
								<span class="label-text font-semibold">{$t('transfers.from-account')}</span>
							</label>
							<select
								id="source-account"
								class="select select-bordered w-full"
								bind:value={sourceAccountToken}
								required
							>
								{#each accounts as acc}
									<option value={acc.token}>{acc.account_name}</option>
								{/each}
							</select>
						</div>

						{#if debitCategories.length > 0}
							<div class="form-control">
								<label class="label" for="debit-category">
									<span class="label-text">{$t('transfers.debit-category')}</span>
									<span class="label-text-alt text-xs opacity-70"
										>{$t('transfers.money-leaving')}</span
									>
								</label>
								<select
									id="debit-category"
									class="select select-bordered w-full border-2"
									bind:value={debitCategoryId}
									required
									style="border-color: {debitBorderColor} !important;"
								>
									{#each groupedDebitCategories as group}
										{#if group.parent}
											{#if group.children.length > 0}
												<option value={String(group.parent.id)} class="font-semibold">
													{group.parent.category_name}
												</option>
												{#each group.children as child}
													<option value={String(child.id)}>
														&nbsp;&nbsp;&nbsp;&nbsp;{child.category_name}
													</option>
												{/each}
											{:else}
												<option value={String(group.parent.id)}>
													{group.parent.category_name}
												</option>
											{/if}
										{:else}
											{#each group.children as child}
												<option value={String(child.id)}>{child.category_name}</option>
											{/each}
										{/if}
									{/each}
								</select>
							</div>
						{:else}
							<div class="alert alert-warning">
								<p class="text-sm">{$t('transfers.no-debit-categories')}</p>
							</div>
						{/if}
					</div>

					<div class="hidden items-center justify-center md:flex md:px-4">
						<ArrowRight size={32} class="text-primary" />
					</div>

					<div class="flex-1 space-y-4">
						<div class="form-control">
							<label class="label" for="dest-account">
								<span class="label-text font-semibold">{$t('transfers.to-account')}</span>
							</label>
							<select
								id="dest-account"
								class="select select-bordered w-full"
								bind:value={destinationAccountToken}
								required
							>
								<option value="" disabled>{$t('transfers.select-destination')}</option>
								{#each accounts.filter((a) => a.token !== sourceAccountToken) as acc}
									<option value={acc.token}>{acc.account_name}</option>
								{/each}
							</select>
						</div>

						{#if creditCategories.length > 0}
							<div class="form-control">
								<label class="label" for="credit-category">
									<span class="label-text">{$t('transfers.credit-category')}</span>
									<span class="label-text-alt text-xs opacity-70"
										>{$t('transfers.money-entering')}</span
									>
								</label>
								<select
									id="credit-category"
									class="select select-bordered w-full border-2"
									bind:value={creditCategoryId}
									required
									style="border-color: {creditBorderColor} !important;"
								>
									{#each groupedCreditCategories as group}
										{#if group.parent}
											{#if group.children.length > 0}
												<option value={String(group.parent.id)} class="font-semibold">
													{group.parent.category_name}
												</option>
												{#each group.children as child}
													<option value={String(child.id)}>
														&nbsp;&nbsp;&nbsp;&nbsp;{child.category_name}
													</option>
												{/each}
											{:else}
												<option value={String(group.parent.id)}>
													{group.parent.category_name}
												</option>
											{/if}
										{:else}
											{#each group.children as child}
												<option value={String(child.id)}>{child.category_name}</option>
											{/each}
										{/if}
									{/each}
								</select>
							</div>
						{:else}
							<div class="alert alert-warning">
								<p class="text-sm">{$t('transfers.no-credit-categories')}</p>
							</div>
						{/if}
					</div>
				</div>

				<div class="form-control mt-6">
					<label class="label" for="description">
						<span class="label-text">{$t('transactions.description')}</span>
					</label>
					<input
						id="description"
						type="text"
						placeholder={$t('transfers.transfer-description')}
						class="input input-bordered"
						bind:value={description}
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
							placeholder={$t('transfers.transfer-amount')}
							class="input input-bordered w-full"
							bind:value={amount}
							min="0.01"
							step="0.01"
							required
						/>
					</div>
				</div>

				<div class="mt-4 flex gap-4">
					<div class="form-control flex-1">
						<label class="label" for="frequency">
							<span class="label-text">{$t('recurring.frequency')}</span>
						</label>
						<select id="frequency" class="select select-bordered w-full" bind:value={frequency}>
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
						<input
							id="interval"
							type="number"
							min="1"
							step="1"
							class="input input-bordered w-full"
							bind:value={intervalValue}
						/>
					</div>
				</div>

				{#if frequency === 'monthly'}
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
							bind:value={executionDay}
						/>
						<div class="label">
							<span class="label-text-alt">{$t('recurring.execution-day-help')}</span>
						</div>
						{#if executionDay && executionDay >= 29}
							<div class="label">
								<span class="label-text-alt text-warning">
									{$t('recurring.execution-day-end-of-month-note')}
								</span>
							</div>
						{/if}
					</div>
				{/if}

				{#if frequency === 'weekly'}
					<div class="form-control mt-4">
						<label class="label" for="execution-weekday">
							<span class="label-text">{$t('recurring.execution-weekday')}</span>
						</label>
						<select
							id="execution-weekday"
							class="select select-bordered w-full"
							bind:value={executionWeekday}
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
						<input class="toggle" type="checkbox" bind:checked={active} />
						<span>{$t('recurring.active')}</span>
					</label>
				</div>

				<div class="modal-action mt-6">
					<button type="button" class="btn" onclick={handleCloseModal}>{$t('common.cancel')}</button
					>
					<button
						type="submit"
						class="btn btn-primary text-base-100"
						disabled={debitCategories.length === 0 || creditCategories.length === 0}
					>
						{isEditMode
							? $t('recurring.update-recurring-transfer')
							: $t('recurring.create-recurring-transfer')}
					</button>
				</div>
			</form>
		{/if}
	</div>
</div>
