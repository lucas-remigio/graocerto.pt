<script lang="ts">
	import type { CategoryDto } from '$lib/types';
	import { ChevronDown, ChevronRight, Pencil, Plus, Trash } from 'lucide-svelte';
	import CategoryModal from './CategoryModal.svelte';
	import { createEventDispatcher } from 'svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import { locale, t } from '$lib/i18n';

	export let categories: CategoryDto[] = [];
	export let categoryType: 'debit' | 'credit' = 'debit';

	let editCategoryModalOpen = false;
	let createSubcategoryModalOpen = false;
	let selectedCategory: CategoryDto | null = null;
	let parentForNewSubcategory: CategoryDto | null = null;

	let promptDeleteCategoryModalOpen = false;

	// Track expanded parent categories
	let expandedCategories = new Set<number>();

	$: currentLocale = $locale || 'pt';

	// Build hierarchy from flat list
	$: categoryHierarchy = buildHierarchy(categories);

	$: {
		categoryHierarchy = buildHierarchy(categories);
		// Auto-expand all parents with children
		expandedCategories = new Set(
			categoryHierarchy.filter((node) => node.children.length > 0).map((node) => node.category.id)
		);
	}

	interface CategoryNode {
		category: CategoryDto;
		children: CategoryDto[];
	}

	function buildHierarchy(flatCategories: CategoryDto[]): CategoryNode[] {
		const parents = flatCategories.filter((c) => !c.parent_category_id);
		const children = flatCategories.filter((c) => c.parent_category_id);

		return parents.map((parent) => ({
			category: parent,
			children: children.filter((child) => child.parent_category_id === parent.id)
		}));
	}

	function toggleExpand(categoryId: number) {
		if (expandedCategories.has(categoryId)) {
			expandedCategories.delete(categoryId);
		} else {
			expandedCategories.add(categoryId);
		}
		expandedCategories = expandedCategories; // trigger reactivity
	}

	function formatCurrency(v: number | null) {
		if (v == null) return '';
		return new Intl.NumberFormat(currentLocale, { maximumFractionDigits: 0 }).format(v) + '€';
	}

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
		selectedCategory = null;
	}

	function openCreateSubcategoryModal(parent: CategoryDto) {
		parentForNewSubcategory = parent;
		createSubcategoryModalOpen = true;
	}

	function closeCreateSubcategoryModal() {
		createSubcategoryModalOpen = false;
		parentForNewSubcategory = null;
	}

	const dispatch = createEventDispatcher();

	function handleEditCategory(
		event: CustomEvent<{
			categoryId: number;
			categoryData: {
				parent_category_id?: number | null;
				category_name: string;
				color: string;
				budget?: number | null;
			};
		}>
	) {
		closeEditCategoryModal();
		dispatch('editCategory', event.detail);
	}

	function handleNewSubcategory(event: CustomEvent) {
		closeCreateSubcategoryModal();
		dispatch('newCategory', event.detail);
	}

	function handlePromptDeleteCategory(category: CategoryDto) {
		selectedCategory = category;
		promptDeleteCategoryModalOpen = true;
	}

	function closePromptDeleteCategoryModal() {
		promptDeleteCategoryModalOpen = false;
		selectedCategory = null;
	}

	function handleConfirmDeleteCategory(categoryId: number) {
		closePromptDeleteCategoryModal();
		dispatch('deleteCategory', { categoryId });
	}
</script>

{#if categories.length === 0}
	<p class="text-base-content/70 py-8 text-center">{$t('categories.no-categories')}</p>
{:else}
	<div class="overflow-x-auto rounded-xl border-2 {modalBorderClass}">
		<table class="table-zebra table w-full">
			<thead class="text-center">
				<tr>
					<th class="w-12"></th>
					<th>{$t('categories.category-name')}</th>
					<th>{$t('categories.color')}</th>
					<th>{$t('categories.budget')}</th>
					<th>{$t('categories.actions')}</th>
				</tr>
			</thead>
			<tbody class="text-center">
				{#each categoryHierarchy as node (node.category.id)}
					<!-- Parent Category Row -->
					{@const { category } = node}
					<tr class="font-medium">
						<td>
							{#if node.children.length > 0}
								<button
									class="btn btn-ghost btn-xs btn-circle"
									on:click={() => toggleExpand(category.id)}
									aria-label={expandedCategories.has(category.id) ? 'Collapse' : 'Expand'}
								>
									{#if expandedCategories.has(category.id)}
										<ChevronDown size={16} />
									{:else}
										<ChevronRight size={16} />
									{/if}
								</button>
							{/if}
						</td>
						<td class="text-left">{category.category_name}</td>
						<td>
							<div class="flex items-center justify-center space-x-2">
								<span
									class="inline-block h-4 w-4 rounded-full"
									style="background-color: {category.color};"
								></span>
								<span class="text-sm">{category.color}</span>
							</div>
						</td>
						<td>{category.budget ? formatCurrency(category.budget) : '—'}</td>
						<td>
							<div class="flex items-center justify-center gap-1">
								<button
									class="btn btn-ghost btn-sm btn-circle bg-base-100/80 text-success hover:bg-success/20 backdrop-blur-sm"
									on:click={() => openCreateSubcategoryModal(category)}
									title={$t('categories.add-subcategory')}
								>
									<Plus size={20} />
								</button>
								<button
									class="btn btn-ghost btn-sm btn-circle bg-base-100/80 backdrop-blur-sm"
									on:click={() => openEditCategoryModal(category)}
									title={$t('common.edit')}
								>
									<Pencil size={20} />
								</button>
								<button
									class="btn btn-ghost btn-sm btn-circle bg-base-100/80 text-error hover:bg-error/20 backdrop-blur-sm"
									on:click={() => handlePromptDeleteCategory(category)}
									title={$t('common.delete')}
								>
									<Trash size={20} />
								</button>
							</div>
						</td>
					</tr>

					<!-- Subcategories -->
					{#if expandedCategories.has(category.id) && node.children.length > 0}
						{#each node.children as subcategory (subcategory.id)}
							<tr class="bg-base-200/50">
								<td></td>
								<td class="text-left">
									<span class="ml-8 text-sm opacity-90">{subcategory.category_name}</span>
								</td>
								<td>
									<div class="flex items-center justify-center space-x-2">
										<span
											class="inline-block h-4 w-4 rounded-full"
											style="background-color: {subcategory.color};"
										></span>
										<span class="text-sm">{subcategory.color}</span>
									</div>
								</td>
								<td class="text-sm">
									{subcategory.budget ? formatCurrency(subcategory.budget) : '—'}
								</td>
								<td>
									<div class="flex items-center justify-center gap-1">
										<button
											class="btn btn-ghost btn-sm btn-circle bg-base-100/80 backdrop-blur-sm"
											on:click={() => openEditCategoryModal(subcategory)}
											title={$t('common.edit')}
										>
											<Pencil size={18} />
										</button>
										<button
											class="btn btn-ghost btn-sm btn-circle bg-base-100/80 text-error hover:bg-error/20 backdrop-blur-sm"
											on:click={() => handlePromptDeleteCategory(subcategory)}
											title={$t('common.delete')}
										>
											<Trash size={18} />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					{/if}
				{/each}
			</tbody>
		</table>
	</div>
{/if}
<!-- Edit Category Modal -->
{#if editCategoryModalOpen && selectedCategory}
	<CategoryModal
		category={selectedCategory}
		transactionType={null}
		parentCategory={null}
		on:closeModal={closeEditCategoryModal}
		on:editCategory={handleEditCategory}
	/>
{/if}

<!-- Create Subcategory Modal -->
{#if createSubcategoryModalOpen && parentForNewSubcategory}
	<CategoryModal
		category={null}
		transactionType={null}
		parentCategory={parentForNewSubcategory}
		on:closeModal={closeCreateSubcategoryModal}
		on:newCategory={handleNewSubcategory}
	/>
{/if}

<!-- Delete Confirmation Modal -->
{#if promptDeleteCategoryModalOpen && selectedCategory}
	<ConfirmAction
		title={$t('categories.delete-category')}
		message={`${$t('categories.delete-confirm')} ${selectedCategory.category_name}? ${$t('categories.delete-warning')}`}
		type="danger"
		onConfirm={() => handleConfirmDeleteCategory(selectedCategory!.id)}
		onCancel={closePromptDeleteCategoryModal}
	/>
{/if}
