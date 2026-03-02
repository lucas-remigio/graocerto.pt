<script lang="ts">
	import type { CategoryDto } from '$lib/types';
	import {
		ArrowDown,
		ArrowUp,
		ChevronDown,
		ChevronRight,
		Pencil,
		Plus,
		Trash
	} from 'lucide-svelte';
	import CategoryModal from './CategoryModal.svelte';
	import { createEventDispatcher } from 'svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import { locale, t } from '$lib/i18n';
	import { dataService } from '$lib/services/dataService';
	import { flip } from 'svelte/animate';

	export let categories: CategoryDto[] = [];
	export let categoryType: 'debit' | 'credit' = 'debit';

	// Local mutable copy so we can reorder without mutating the prop
	let localCategories: CategoryDto[] = [];
	$: localCategories = [...categories];

	let editCategoryModalOpen = false;
	let createSubcategoryModalOpen = false;
	let selectedCategory: CategoryDto | null = null;
	let parentForNewSubcategory: CategoryDto | null = null;

	let promptDeleteCategoryModalOpen = false;

	// Track expanded parent categories
	let expandedCategories = new Set<number>();

	$: currentLocale = $locale || 'pt';

	// Build hierarchy from flat list
	// Build hierarchy from flatCategories, sorted by order_index
	$: categoryHierarchy = buildHierarchy(localCategories);

	$: {
		categoryHierarchy = buildHierarchy(localCategories);
		// Auto-expand all parents with children
		expandedCategories = new Set(
			categoryHierarchy.filter((node) => node.children.length > 0).map((node) => node.category.id)
		);
	}

	interface CategoryNode {
		category: CategoryDto;
		children: CategoryDto[];
	}

	type FlatRow =
		| { type: 'parent'; node: CategoryNode; nodeIdx: number }
		| { type: 'child'; subcategory: CategoryDto; parentNode: CategoryNode; childIdx: number };

	function buildFlatRows(hierarchy: CategoryNode[], expanded: Set<number>): FlatRow[] {
		const rows: FlatRow[] = [];
		hierarchy.forEach((node, nodeIdx) => {
			rows.push({ type: 'parent', node, nodeIdx });
			if (expanded.has(node.category.id)) {
				node.children.forEach((subcategory, childIdx) => {
					rows.push({ type: 'child', subcategory, parentNode: node, childIdx });
				});
			}
		});
		return rows;
	}

	$: flatRows = buildFlatRows(categoryHierarchy, expandedCategories);

	function buildHierarchy(flatCategories: CategoryDto[]): CategoryNode[] {
		const parents = flatCategories
			.filter((c) => !c.parent_category_id)
			.sort((a, b) => a.order_index - b.order_index);
		const children = flatCategories.filter((c) => c.parent_category_id);

		return parents.map((parent) => ({
			category: parent,
			children: children
				.filter((child) => child.parent_category_id === parent.id)
				.sort((a, b) => a.order_index - b.order_index)
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

	function openPromptDeleteCategoryModal(category: CategoryDto) {
		selectedCategory = category;
		promptDeleteCategoryModalOpen = true;
	}

	function moveParent(node: CategoryNode, direction: 'up' | 'down') {
		const parents = categoryHierarchy.map((n) => n.category);
		const idx = parents.findIndex((p) => p.id === node.category.id);
		const targetIdx = direction === 'up' ? idx - 1 : idx + 1;
		if (targetIdx < 0 || targetIdx >= parents.length) return;

		const aOrderIndex = parents[idx].order_index;
		const bOrderIndex = parents[targetIdx].order_index;
		localCategories = localCategories.map((c) => {
			if (c.id === parents[idx].id) return { ...c, order_index: bOrderIndex };
			if (c.id === parents[targetIdx].id) return { ...c, order_index: aOrderIndex };
			return c;
		});
		sendReorderRequest([
			{ id: parents[idx].id, order_index: bOrderIndex },
			{ id: parents[targetIdx].id, order_index: aOrderIndex }
		]);
	}

	function moveChild(child: CategoryDto, parentId: number, direction: 'up' | 'down') {
		const node = categoryHierarchy.find((n) => n.category.id === parentId);
		if (!node) return;
		const children = node.children;
		const idx = children.findIndex((c) => c.id === child.id);
		const targetIdx = direction === 'up' ? idx - 1 : idx + 1;
		if (targetIdx < 0 || targetIdx >= children.length) return;

		const aOrderIndex = children[idx].order_index;
		const bOrderIndex = children[targetIdx].order_index;
		localCategories = localCategories.map((c) => {
			if (c.id === children[idx].id) return { ...c, order_index: bOrderIndex };
			if (c.id === children[targetIdx].id) return { ...c, order_index: aOrderIndex };
			return c;
		});
		sendReorderRequest([
			{ id: children[idx].id, order_index: bOrderIndex },
			{ id: children[targetIdx].id, order_index: aOrderIndex }
		]);
	}

	async function sendReorderRequest(updates: { id: number; order_index: number }[]) {
		try {
			await dataService.reorderCategories(updates);
		} catch (error) {
			console.error('Error reordering categories:', error);
		}
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
	<p class="py-8 text-center text-base-content/70">{$t('categories.no-categories')}</p>
{:else}
	<div class="overflow-x-auto rounded-xl border-2 {modalBorderClass}">
		<table class="table table-zebra w-full">
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
				{#each flatRows as row (`${row.type}-${row.type === 'parent' ? row.node.category.id : row.subcategory.id}`)}
					<tr
						animate:flip={{ duration: 300 }}
						class={row.type === 'parent' ? 'font-medium' : 'bg-base-200/50'}
					>
						{#if row.type === 'parent'}
							<td>
								{#if row.node.children.length > 0}
									<button
										class="btn btn-circle btn-ghost btn-xs"
										on:click={() => toggleExpand(row.node.category.id)}
										aria-label={expandedCategories.has(row.node.category.id)
											? 'Collapse'
											: 'Expand'}
									>
										{#if expandedCategories.has(row.node.category.id)}
											<ChevronDown size={16} />
										{:else}
											<ChevronRight size={16} />
										{/if}
									</button>
								{/if}
							</td>
							<td class="text-left">{row.node.category.category_name}</td>
							<td>
								<div class="flex items-center justify-center space-x-2">
									<span
										class="inline-block h-4 w-4 rounded-full"
										style="background-color: {row.node.category.color};"
									></span>
									<span class="text-sm">{row.node.category.color}</span>
								</div>
							</td>
							<td>{row.node.category.budget ? formatCurrency(row.node.category.budget) : '—'}</td>
							<td>
								<div class="flex items-center justify-center gap-1">
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80"
										on:click={() => moveParent(row.node, 'up')}
										title="Move up"
										disabled={row.nodeIdx === 0}
									>
										<ArrowUp size={16} />
									</button>
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80"
										on:click={() => moveParent(row.node, 'down')}
										title="Move down"
										disabled={row.nodeIdx === categoryHierarchy.length - 1}
									>
										<ArrowDown size={16} />
									</button>
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-success backdrop-blur-sm hover:bg-success/20"
										on:click={() => openCreateSubcategoryModal(row.node.category)}
										title={$t('categories.add-subcategory')}
									>
										<Plus size={20} />
									</button>
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80 backdrop-blur-sm"
										on:click={() => openEditCategoryModal(row.node.category)}
										title={$t('common.edit')}
									>
										<Pencil size={20} />
									</button>
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-error backdrop-blur-sm hover:bg-error/20"
										on:click={() => openPromptDeleteCategoryModal(row.node.category)}
										title={$t('common.delete')}
									>
										<Trash size={20} />
									</button>
								</div>
							</td>
						{:else}
							<td></td>
							<td class="text-left">
								<span class="ml-8 text-sm opacity-90">{row.subcategory.category_name}</span>
							</td>
							<td>
								<div class="flex items-center justify-center space-x-2">
									<span
										class="inline-block h-4 w-4 rounded-full"
										style="background-color: {row.subcategory.color};"
									></span>
									<span class="text-sm">{row.subcategory.color}</span>
								</div>
							</td>
							<td class="text-sm">
								{row.subcategory.budget ? formatCurrency(row.subcategory.budget) : '—'}
							</td>
							<td>
								<div class="flex items-center justify-center gap-1">
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80"
										on:click={() => moveChild(row.subcategory, row.parentNode.category.id, 'up')}
										title="Move up"
										disabled={row.childIdx === 0}
									>
										<ArrowUp size={16} />
									</button>
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80"
										on:click={() => moveChild(row.subcategory, row.parentNode.category.id, 'down')}
										title="Move down"
										disabled={row.childIdx === row.parentNode.children.length - 1}
									>
										<ArrowDown size={16} />
									</button>
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80 backdrop-blur-sm"
										on:click={() => openEditCategoryModal(row.subcategory)}
										title={$t('common.edit')}
									>
										<Pencil size={18} />
									</button>
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-error backdrop-blur-sm hover:bg-error/20"
										on:click={() => openPromptDeleteCategoryModal(row.subcategory)}
										title={$t('common.delete')}
									>
										<Trash size={18} />
									</button>
								</div>
							</td>
						{/if}
					</tr>
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
