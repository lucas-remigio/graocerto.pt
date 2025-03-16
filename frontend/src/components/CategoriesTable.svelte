<script lang="ts">
	import type { CategoryDto } from '$lib/types';
	import { Pencil, Trash } from 'lucide-svelte';
	import EditCategory from './EditCategory.svelte';
	import { createEventDispatcher } from 'svelte';
	import ConfirmAction from './ConfirmAction.svelte';

	export let categories: CategoryDto[] = [];
	export let categoryType: 'debit' | 'credit' = 'debit';

	let editCategoryModalOpen = false;
	let selectedCategory: CategoryDto | null = null;

	let promptDeleteCategoryModalOpen = false;

	const borderClasses: Record<string, string> = {
		credit: 'border-green-500 dark:border-green-400',
		debit: 'border-red-500 dark:border-red-400',
		transfer: 'border-blue-500 dark:border-blue-400'
	};
	let modalBorderClass = categoryType ? borderClasses[categoryType] : 'border-gray-50';

	function openEditCategoryModal(category: CategoryDto) {
		selectedCategory = category;
		editCategoryModalOpen = true;
	}

	function closeEditCategoryModal() {
		editCategoryModalOpen = false;
	}

	const dispatch = createEventDispatcher();

	function handleEditCategory() {
		closeEditCategoryModal();
		dispatch('editCategory');
	}

	function handlePromptDeleteCategory(category: CategoryDto) {
		selectedCategory = category;
		promptDeleteCategoryModalOpen = true;
	}

	function closePromptDeleteCategoryModal() {
		promptDeleteCategoryModalOpen = false;
	}

	function handleConfirmDeleteCategory(categoryId: number) {
		closePromptDeleteCategoryModal();
		dispatch('deleteCategory', { categoryId });
	}
</script>

{#if categories.length === 0}
	<p class="text-gray-500">No categories available.</p>
{:else}
	<table class="table-zebra table w-full border-2 {modalBorderClass}">
		<thead>
			<tr>
				<th>Category Name</th>
				<th>Color</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			{#each categories as category (category.id)}
				<tr>
					<td>{category.category_name}</td>
					<td>
						<div class="flex items-center space-x-2">
							<span
								class="inline-block h-4 w-4 rounded-full"
								style="background-color: {category.color};"
							></span>
							<span>{category.color}</span>
						</div>
					</td>
					<td>
						<button
							class="btn btn-ghost btn-sm btn-circle bg-base-100/80 backdrop-blur-sm"
							on:click={() => openEditCategoryModal(category)}
						>
							<Pencil size={20} />
						</button>
						<button
							class="btn btn-ghost btn-sm btn-circle bg-base-100/80 text-error hover:bg-error/20 backdrop-blur-sm"
							on:click={() => handlePromptDeleteCategory(category)}
						>
							<Trash size={20} />
						</button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}

{#if editCategoryModalOpen}
	<EditCategory
		category={selectedCategory!}
		on:closeModal={closeEditCategoryModal}
		on:editCategory={handleEditCategory}
	/>
{/if}

{#if promptDeleteCategoryModalOpen}
	<ConfirmAction
		title="Delete Category"
		message={`Are you sure you want to delete the category ${selectedCategory!.category_name}? This action cannot be undone.`}
		type="danger"
		onConfirm={() => handleConfirmDeleteCategory(selectedCategory!.id)}
		onCancel={closePromptDeleteCategoryModal}
	/>
{/if}
