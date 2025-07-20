<script lang="ts">
	import type { Category, CategoryDto, TransactionType } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import { t } from '$lib/i18n';

	export let category: CategoryDto;

	let error: string = '';
	// Form field variables
	let category_name: string = category.category_name;
	let color: string = category.color;

	// Create event dispatcher (to emit events to the parent)
	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		error = '';
		if (!validateForm()) {
			return;
		}

		const editCategoryData = {
			category_name: category_name,
			color: color
		};

		// Dispatch the edit data to the parent component
		dispatch('editCategory', { categoryId: category.id, categoryData: editCategoryData });
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

	let modalBorderClass = borderClasses[category.transaction_type.type_slug] || 'bg-gray-50';

	function handleCloseModal() {
		dispatch('closeModal');
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative border-4 {modalBorderClass}">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}>
			<X />
		</button>
		<h3 class="mb-4 text-lg font-bold">
			{$t('categories.edit-category-title')} -
			{$t('transaction-types.' + category.transaction_type.type_slug)} - {category.category_name}
		</h3>
		<!-- Error message -->
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
				<div class="flex items-center gap-4">
					<!-- Native color input, visually hidden but accessible -->
					<label class="cursor-pointer">
						<input id="color" type="color" class="sr-only" bind:value={color} required />
						<span
							class="border-base-300 bg-base-100 inline-block h-10 w-10 rounded-full border-2 transition hover:scale-105"
							style="background-color: {color};"
						></span>
					</label>
					<!-- Hex value input -->
					<span class="text-sm font-medium">{color}</span>
				</div>
			</div>

			<!-- Form Actions -->
			<div class="modal-action mt-6">
				<button type="button" class="btn" on:click={handleCloseModal}>{$t('common.cancel')}</button>
				<button type="submit" class="btn btn-primary text-base-100"
					>{$t('categories.edit-category')}</button
				>
			</div>
		</form>
	</div>
</div>
