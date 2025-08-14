<script lang="ts">
	import type { CategoryDto } from '$lib/types';
	import { Pencil, Trash } from 'lucide-svelte';
	import CategoryModal from './CategoryModal.svelte';
	import { createEventDispatcher } from 'svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import { t } from '$lib/i18n';

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

	function handleEditCategory(
		event: CustomEvent<{
			categoryId: number;
			categoryData: { category_name: string; color: string };
		}>
	) {
		closeEditCategoryModal();
		dispatch('editCategory', event.detail);
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
	<p class="text-base-content/70 py-8 text-center">{$t('categories.no-categories')}</p>
{:else}
	<div class="overflow-hidden rounded-xl border-2 {modalBorderClass}">
		<table class="table-zebra table w-full">
			<thead class="text-center">
				<tr>
					<th>{$t('categories.category-name')}</th>
					<th>{$t('categories.color')}</th>
					<th>{$t('categories.actions')}</th>
				</tr>
			</thead>
			<tbody class="text-center">
				{#each categories as category (category.id)}
					<tr>
						<td>{category.category_name}</td>
						<td>
							<div class="flex items-center justify-center space-x-2">
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
	</div>
{/if}

{#if editCategoryModalOpen}
	<CategoryModal
		category={selectedCategory!}
		transactionType={null}
		on:closeModal={closeEditCategoryModal}
		on:editCategory={handleEditCategory}
	/>
{/if}

{#if promptDeleteCategoryModalOpen}
	<ConfirmAction
		title={$t('categories.delete-category')}
		message={`${$t('categories.delete-confirm')} ${selectedCategory!.category_name}? ${$t('categories.delete-warning')}`}
		type="danger"
		onConfirm={() => handleConfirmDeleteCategory(selectedCategory!.id)}
		onCancel={closePromptDeleteCategoryModal}
	/>
{/if}
