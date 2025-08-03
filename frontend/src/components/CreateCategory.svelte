<script lang="ts">
	import { dataService } from '$lib/services/dataService';
	import type { Category, CategoryChangeResponse, CategoryDto, TransactionType } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';
	import { t } from '$lib/i18n';

	export let transactionType: TransactionType;

	let error: string = '';
	// Form field variables
	let category_name: string = '';
	let color: string = '#ffffff';

	// Create event dispatcher (to emit events to the parent)
	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		error = '';
		if (!validateForm()) {
			return;
		}

		const categoryData = {
			transaction_type_id: transactionType.id,
			category_name: category_name,
			color: color
		};

		try {
			const response = await dataService.createCategory(categoryData);
			handleNewCategory(response);
		} catch (err: any) {
			console.error('Error in handleSubmit:', err);
			error = err.message || $t('errors.failed-create-category');
		}
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

	let modalBorderClass = borderClasses[transactionType.type_slug] || 'bg-gray-50';

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function handleNewCategory(response: CategoryChangeResponse) {
		dispatch('newCategory', response);
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative border-4 {modalBorderClass}">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}>
			<X />
		</button>
		<h3 class="mb-4 text-lg font-bold">
			{$t('categories.new-category-for')} -
			{$t('transaction-types.' + transactionType.type_slug)}
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
				<button type="submit" class="btn btn-primary text-base-100"
					>{$t('categories.create-category')}</button
				>
			</div>
		</form>
	</div>
</div>
