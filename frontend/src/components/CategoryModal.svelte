<script lang="ts">
	import { dataService } from '$lib/services/dataService';
	import type { CategoryChangeResponse, CategoryDto, TransactionType } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { t } from '$lib/i18n';

	// Unified inputs: provide either `category` (edit) or `transactionType` (create)
	export let category: CategoryDto | null = null; // edit mode if provided
	export let transactionType: TransactionType | null = null; // create mode if provided

	// Mode
	$: isEditMode = !!category;

	// Local state
	let error: string = '';
	let category_name: string = category ? category.category_name : '';
	let color: string = category ? category.color : '#ffffff';

	// Events
	const dispatch = createEventDispatcher();

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function validateForm(): boolean {
		if (!category_name) {
			error = $t('categories.category-name-required');
			return false;
		}

		category_name = category_name.trim();

		if (category_name.length > 50) {
			error = $t('categories.category-name-too-long');
			return false;
		}

		if (category_name.length < 3) {
			error = $t('categories.category-name-too-short');
			return false;
		}

		if (!color) {
			error = $t('categories.color-required');
			return false;
		}

		color = color.trim();

		if (color[0] !== '#') {
			error = $t('categories.color-invalid');
			return false;
		}

		if (color.length !== 7) {
			error = $t('categories.color-invalid');
			return false;
		}

		return true;
	}

	const borderClasses: Record<string, string> = {
		credit: 'border-green-500 dark:border-green-400',
		debit: 'border-red-500 dark:border-red-400',
		transfer: 'border-blue-500 dark:border-blue-400'
	};

	$: typeSlug = isEditMode ? category!.transaction_type.type_slug : transactionType?.type_slug;
	$: modalBorderClass = typeSlug ? borderClasses[typeSlug] : 'bg-gray-50';

	async function handleSubmit() {
		error = '';
		if (!validateForm()) return;

		if (isEditMode) {
			// Delegate API call to parent (keeps existing contract)
			const editCategoryData = {
				category_name,
				color
			};
			dispatch('editCategory', { categoryId: category!.id, categoryData: editCategoryData });
			return;
		}

		// Create mode: call API here (keeps existing contract)
		if (!transactionType) {
			error = $t('errors.failed-create-category');
			return;
		}

		const categoryData = {
			transaction_type_id: transactionType.id,
			category_name,
			color
		};

		try {
			const response: CategoryChangeResponse = await dataService.createCategory(categoryData);
			dispatch('newCategory', response);
		} catch (err: any) {
			console.error('Error in handleSubmit:', err);
			error = err?.message || $t('errors.failed-create-category');
		}
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative border-4 {modalBorderClass}">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}>
			<X />
		</button>

		<h3 class="mb-4 text-lg font-bold">
			{#if isEditMode}
				{$t('categories.edit-category-title')} -
				{$t('transaction-types.' + category!.transaction_type.type_slug)} - {category!
					.category_name}
			{:else}
				{$t('categories.new-category-for')} -
				{$t('transaction-types.' + (transactionType ? transactionType.type_slug : ''))}
			{/if}
		</h3>

		{#if error}
			<div class="alert alert-error">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}

		<form on:submit|preventDefault={handleSubmit}>
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
					<!-- Color preview -->
					<div
						class="border-base-300 h-14 w-14 rounded-full border-2 shadow"
						style="background-color: {color};"
					></div>
					<!-- Hex value -->
					<input
						type="text"
						class="input input-bordered input-sm w-24 text-center"
						value={color}
						readonly
						tabindex="-1"
					/>
					<!-- Full-size invisible color input -->
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
