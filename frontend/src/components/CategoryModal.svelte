<script lang="ts">
	import { dataService } from '$lib/services/dataService';
	import type { CategoryChangeResponse, CategoryDto, TransactionType } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/i18n';
	import { formatCurrency } from '$lib/utils/currency';
	import { toastStore } from '$lib/stores/toast';

	// Unified inputs
	export let category: CategoryDto | null = null; // edit mode if provided
	export let transactionType: TransactionType | null = null; // create mode if provided
	export let parentCategory: CategoryDto | null = null; // for creating subcategories from table

	// Mode
	$: isEditMode = !!category;
	$: isSubcategoryMode = !!parentCategory;

	// Local state
	let category_name: string = category ? category.category_name : '';
	let color: string = category ? category.color : parentCategory?.color || '#ffffff';
	let budget: number | null = category ? (category.budget ?? null) : null;
	let budgetInput: string = category && category.budget != null ? String(category.budget) : '';
	let selectedParentId: number | null = category?.parent_category_id ?? parentCategory?.id ?? null;

	// Fetch available parent categories based on transaction type
	let availableParents: CategoryDto[] = [];
	let loadingParents = false;

	$: currentTransactionTypeId = isEditMode
		? category!.transaction_type.id
		: isSubcategoryMode
			? parentCategory!.transaction_type.id
			: transactionType?.id;

	// Fetch parent categories when transaction type changes
	$: if (currentTransactionTypeId) {
		fetchAvailableParents(currentTransactionTypeId);
	}

	async function fetchAvailableParents(transactionTypeId: number) {
		loadingParents = true;
		try {
			const allCategories = await dataService.fetchCategories();
			// Filter: same transaction type, no parent (root only), exclude current category if editing
			availableParents = allCategories.filter(
				(c) =>
					c.transaction_type.id === transactionTypeId &&
					!c.parent_category_id &&
					(!isEditMode || c.id !== category!.id) // Can't be parent of itself
			);
		} catch (err) {
			console.error('Failed to fetch parent categories:', err);
			availableParents = [];
		} finally {
			loadingParents = false;
		}
	}

	function onBudgetInput(e: Event) {
		const raw = (e.target as HTMLInputElement).value;
		const digits = raw.replace(/[^\d]/g, '');
		budgetInput = digits;
		if (digits === '') {
			budget = null;
		} else {
			const n = parseInt(digits, 10);
			if (Number.isNaN(n)) {
				budget = null;
			} else {
				budget = Math.min(99999, Math.max(1, n));
			}
		}
	}

	const dispatch = createEventDispatcher();

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function validateForm(): boolean {
		if (!category_name) {
			toastStore.error($t('categories.category-name-required'));
			return false;
		}

		category_name = category_name.trim();

		if (category_name.length > 50) {
			toastStore.error($t('categories.category-name-too-long'));
			return false;
		}

		if (category_name.length < 3) {
			toastStore.error($t('categories.category-name-too-short'));
			return false;
		}

		if (!color) {
			toastStore.error($t('categories.color-required'));
			return false;
		}

		color = color.trim();

		if (color[0] !== '#') {
			toastStore.error($t('categories.color-invalid'));
			return false;
		}

		if (color.length !== 7) {
			toastStore.error($t('categories.color-invalid'));
			return false;
		}

		if (budget != null && (budget < 1 || isNaN(budget) || budget > 99999)) {
			toastStore.error($t('categories.budget-invalid'));
			return false;
		}

		return true;
	}

	const borderClasses: Record<string, string> = {
		credit: 'border-green-500 dark:border-green-400',
		debit: 'border-red-500 dark:border-red-400',
		transfer: 'border-blue-500 dark:border-blue-400'
	};

	$: typeSlug = isEditMode
		? category!.transaction_type.type_slug
		: isSubcategoryMode
			? parentCategory!.transaction_type.type_slug
			: transactionType?.type_slug;
	$: modalBorderClass = typeSlug ? borderClasses[typeSlug] : 'bg-gray-50';

	async function handleSubmit() {
		if (!validateForm()) return;

		if (isEditMode) {
			const editCategoryData = {
				parent_category_id: selectedParentId,
				category_name,
				color,
				budget
			};
			dispatch('editCategory', { categoryId: category!.id, categoryData: editCategoryData });
			return;
		}

		// Create mode
		let transaction_type_id: number;
		let parent_category_id: number | null = selectedParentId;

		if (isSubcategoryMode) {
			transaction_type_id = parentCategory!.transaction_type.id;
			parent_category_id = parentCategory!.id; // Override with explicit parent
		} else if (transactionType) {
			transaction_type_id = transactionType.id;
		} else {
			toastStore.error($t('errors.failed-create-category'));
			return;
		}

		const categoryData = {
			transaction_type_id,
			parent_category_id,
			category_name,
			color,
			budget
		};

		try {
			const response: CategoryChangeResponse = await dataService.createCategory(categoryData);
			toastStore.success($t('common.success'));
			dispatch('newCategory', response);
		} catch (err: unknown) {
			console.error('Error in handleSubmit:', err);
			toastStore.error($t('errors.failed-create-category'));
		}
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative border-4 {modalBorderClass}">
		<button class="btn btn-circle btn-sm absolute right-2 top-2" on:click={handleCloseModal}>
			<X />
		</button>

		<h3 class="mb-4 text-lg font-bold">
			{#if isEditMode}
				{$t('categories.edit-category-title')} -
				{$t('transaction-types.' + category!.transaction_type.type_slug)} - {category!
					.category_name}
			{:else if isSubcategoryMode}
				{$t('categories.new-subcategory-for')} - {parentCategory!.category_name}
			{:else}
				{$t('categories.new-category-for')} -
				{$t('transaction-types.' + (transactionType ? transactionType.type_slug : ''))}
			{/if}
		</h3>

		<form on:submit|preventDefault={handleSubmit}>
			<!-- Parent Category Selector (only if NOT in subcategory mode from table) -->
			{#if !isSubcategoryMode && availableParents.length > 0}
				<div class="form-control mt-4">
					<label class="label" for="parent_category">
						<span class="label-text"
							>{$t('categories.parent-category')} ({$t('common.optional')})</span
						>
					</label>
					{#if loadingParents}
						<div class="flex items-center justify-center py-2">
							<span class="loading loading-spinner loading-sm"></span>
						</div>
					{:else}
						<select
							id="parent_category"
							class="select select-bordered"
							bind:value={selectedParentId}
						>
							<option value={null}>{$t('categories.no-parent')}</option>
							{#each availableParents as parent}
								<option value={parent.id}>
									{parent.category_name}
								</option>
							{/each}
						</select>
						<p class="mt-1 text-xs text-base-content/70">
							{$t('categories.parent-category-help')}
						</p>
					{/if}
				</div>
			{/if}

			<!-- Category Name Field -->
			<div class="form-control mt-4">
				<label class="label" for="category_name">
					<span class="label-text">{$t('categories.category-name')}</span>
				</label>
				<input
					id="category_name"
					type="text"
					placeholder={$t('categories.category-name-placeholder')}
					class="input input-bordered"
					bind:value={category_name}
					required
				/>
			</div>

			<!-- Color Field -->
			<div class="form-control mt-4">
				<label class="label" for="color">
					<span class="label-text">{$t('categories.color')}</span>
				</label>
				<div class="relative flex h-14 w-full items-center gap-4">
					<div
						class="h-14 w-14 rounded-full border-2 border-base-300 shadow"
						style="background-color: {color};"
					></div>
					<input
						type="text"
						class="input input-sm input-bordered w-24 text-center"
						value={color}
						readonly
						tabindex="-1"
					/>
					<input
						id="color"
						type="color"
						class="absolute left-0 top-0 h-full w-full cursor-pointer opacity-0"
						bind:value={color}
						required
						aria-label={$t('categories.color')}
					/>
				</div>
			</div>

			<!-- Budget Field -->
			<div class="form-control mt-4">
				<label class="label" for="budget">
					<span class="label-text">{$t('categories.budget')} ({$t('common.optional')})</span>
				</label>
				<div class="flex items-center gap-3">
					<input
						id="budget"
						type="text"
						inputmode="numeric"
						pattern="\d*"
						placeholder={$t('categories.budget-placeholder')}
						class="input input-bordered w-full"
						value={budgetInput}
						on:input={onBudgetInput}
						aria-describedby="budget-help"
					/>
					<div class="rounded border px-3 py-2 text-sm text-base-content/60">
						{#if budget != null}
							{formatCurrency(budget, 0)}
						{:else}
							<span class="text-base-content/40">—</span>
						{/if}
					</div>
				</div>
				<p id="budget-help" class="mt-2 text-xs text-base-content/70">
					{$t('categories.budget-help')}
				</p>
			</div>

			<!-- Form Actions -->
			<div class="modal-action mt-6">
				<button type="button" class="btn" on:click={handleCloseModal}>{$t('common.cancel')}</button>
				{#if isEditMode}
					<button type="submit" class="btn btn-primary text-base-100">
						{$t('categories.edit-category')}
					</button>
				{:else}
					<button type="submit" class="btn btn-primary text-base-100">
						{$t('categories.create-category')}
					</button>
				{/if}
			</div>
		</form>
	</div>
</div>
