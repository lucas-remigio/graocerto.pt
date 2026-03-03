<script lang="ts">
	import type { CategoryDto } from '$lib/types';
	import { ChevronDown, ChevronRight, GripVertical, Pencil, Plus, Trash } from 'lucide-svelte';
	import CategoryModal from './CategoryModal.svelte';
	import { createEventDispatcher } from 'svelte';
	import ConfirmAction from './ConfirmAction.svelte';
	import { t } from '$lib/i18n';
	import { dataService } from '$lib/services/dataService';
	import { flip } from 'svelte/animate';
	import { formatCurrency } from '$lib/utils/currency';

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
		| { type: 'parent'; cat: CategoryDto; node: CategoryNode; nodeIdx: number }
		| { type: 'child'; cat: CategoryDto; parentNode: CategoryNode; childIdx: number };

	function buildFlatRows(hierarchy: CategoryNode[], expanded: Set<number>): FlatRow[] {
		const rows: FlatRow[] = [];
		hierarchy.forEach((node, nodeIdx) => {
			rows.push({ type: 'parent', cat: node.category, node, nodeIdx });
			if (expanded.has(node.category.id)) {
				node.children.forEach((subcategory, childIdx) => {
					rows.push({ type: 'child', cat: subcategory, parentNode: node, childIdx });
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

	// ── Drag-and-drop state ──────────────────────────────────────────────────
	let draggedRowKey: string | null = null;
	let dragOverRowKey: string | null = null;

	function rowKey(row: FlatRow) {
		return `${row.type}-${row.cat.id}`;
	}

	function isCompatibleTarget(sourceKey: string, targetKey: string): boolean {
		if (sourceKey === targetKey) return false;
		const src = flatRows.find((r) => rowKey(r) === sourceKey);
		const tgt = flatRows.find((r) => rowKey(r) === targetKey);
		if (!src || !tgt) return false;
		if (src.type === 'parent' && tgt.type === 'parent') return true;
		if (
			src.type === 'child' &&
			tgt.type === 'child' &&
			src.parentNode.category.id === tgt.parentNode.category.id
		)
			return true;
		return false;
	}

	function handleDragStart(event: DragEvent, row: FlatRow) {
		draggedRowKey = rowKey(row);
		if (event.dataTransfer) {
			event.dataTransfer.effectAllowed = 'move';
			event.dataTransfer.setData('text/plain', draggedRowKey);
		}
	}

	function handleDragOver(event: DragEvent, row: FlatRow) {
		event.preventDefault();
		if (draggedRowKey && isCompatibleTarget(draggedRowKey, rowKey(row))) {
			dragOverRowKey = rowKey(row);
			if (event.dataTransfer) event.dataTransfer.dropEffect = 'move';
		}
	}

	function handleDragLeave(event: DragEvent) {
		// Only clear when leaving the row entirely (not entering a child element)
		const related = event.relatedTarget as HTMLElement | null;
		if (!related || !(event.currentTarget as HTMLElement).contains(related)) {
			dragOverRowKey = null;
		}
	}

	function handleDrop(event: DragEvent, targetRow: FlatRow) {
		event.preventDefault();
		if (!draggedRowKey) return;
		if (!isCompatibleTarget(draggedRowKey, rowKey(targetRow))) {
			draggedRowKey = null;
			dragOverRowKey = null;
			return;
		}

		const sourceRow = flatRows.find((r) => rowKey(r) === draggedRowKey);
		if (!sourceRow) return;

		if (sourceRow.type === 'parent' && targetRow.type === 'parent') {
			reorderGroup(
				categoryHierarchy.map((n) => n.category),
				sourceRow.cat.id,
				targetRow.cat.id
			);
		} else if (sourceRow.type === 'child' && targetRow.type === 'child') {
			const node = categoryHierarchy.find(
				(n) => n.category.id === sourceRow.parentNode.category.id
			);
			if (node) reorderGroup(node.children, sourceRow.cat.id, targetRow.cat.id);
		}

		draggedRowKey = null;
		dragOverRowKey = null;
	}

	function handleDragEnd() {
		draggedRowKey = null;
		dragOverRowKey = null;
	}

	// ── Touch drag state ────────────────────────────────────────────────────
	function handleTouchStart(event: TouchEvent, row: FlatRow) {
		draggedRowKey = rowKey(row);
	}

	function handleTouchMove(event: TouchEvent) {
		if (!draggedRowKey) return;
		event.preventDefault();

		const touch = event.touches[0];
		const el = document.elementFromPoint(touch.clientX, touch.clientY);
		if (!el) return;

		const trEl = el.closest('[data-row-key]') as HTMLElement | null;
		if (!trEl) {
			dragOverRowKey = null;
			return;
		}

		const targetKey = trEl.dataset.rowKey ?? null;
		dragOverRowKey = targetKey && isCompatibleTarget(draggedRowKey, targetKey) ? targetKey : null;
	}

	function handleTouchEnd() {
		if (draggedRowKey && dragOverRowKey) {
			const sourceRow = flatRows.find((r) => rowKey(r) === draggedRowKey);
			const targetRow = flatRows.find((r) => rowKey(r) === dragOverRowKey);
			if (sourceRow && targetRow) {
				if (sourceRow.type === 'parent' && targetRow.type === 'parent') {
					reorderGroup(
						categoryHierarchy.map((n) => n.category),
						sourceRow.cat.id,
						targetRow.cat.id
					);
				} else if (sourceRow.type === 'child' && targetRow.type === 'child') {
					const node = categoryHierarchy.find(
						(n) => n.category.id === sourceRow.parentNode.category.id
					);
					if (node) reorderGroup(node.children, sourceRow.cat.id, targetRow.cat.id);
				}
			}
		}
		draggedRowKey = null;
		dragOverRowKey = null;
	}

	function handleTouchCancel() {
		draggedRowKey = null;
		dragOverRowKey = null;
	}

	/** Splice `sourceId` into the position of `targetId` within `group`, then save. */
	function reorderGroup(group: CategoryDto[], sourceId: number, targetId: number) {
		const items = [...group];
		const fromIdx = items.findIndex((c) => c.id === sourceId);
		const toIdx = items.findIndex((c) => c.id === targetId);
		if (fromIdx === -1 || toIdx === -1) return;

		const [dragged] = items.splice(fromIdx, 1);
		items.splice(toIdx, 0, dragged);

		const updates = items.map((c, i) => ({ id: c.id, order_index: i + 1 }));
		localCategories = localCategories.map((c) => {
			const u = updates.find((upd) => upd.id === c.id);
			return u ? { ...c, order_index: u.order_index } : c;
		});
		sendReorderRequest(updates);
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
				{#each flatRows as row (`${row.type}-${row.cat.id}`)}
					<tr
						data-row-key="{row.type}-{row.cat.id}"
						animate:flip={{ duration: 200 }}
						draggable="true"
						on:dragstart={(e) => handleDragStart(e, row)}
						on:dragover={(e) => handleDragOver(e, row)}
						on:dragleave={handleDragLeave}
						on:drop={(e) => handleDrop(e, row)}
						on:dragend={handleDragEnd}
						on:touchstart={(e) => handleTouchStart(e, row)}
						on:touchmove|nonpassive={handleTouchMove}
						on:touchend={handleTouchEnd}
						on:touchcancel={handleTouchCancel}
						class="{row.type === 'parent' ? 'font-medium' : 'bg-base-200/50'} transition-colors
							{dragOverRowKey === `${row.type}-${row.cat.id}`
							? 'outline outline-2 outline-offset-[-2px] outline-primary'
							: ''}
							{draggedRowKey === `${row.type}-${row.cat.id}` ? 'opacity-40' : ''}"
					>
						<!-- Drag handle + expand toggle -->
						<td class="w-16">
							<div class="flex items-center justify-center gap-1">
								<span
									class="cursor-grab p-2 text-base-content/30 hover:text-base-content/60 active:cursor-grabbing"
								>
									<GripVertical size={18} />
								</span>
								{#if row.type === 'parent' && row.node.children.length > 0}
									<button
										class="btn btn-circle btn-ghost btn-xs"
										on:click={() => toggleExpand(row.cat.id)}
										aria-label={expandedCategories.has(row.cat.id) ? 'Collapse' : 'Expand'}
									>
										{#if expandedCategories.has(row.cat.id)}
											<ChevronDown size={16} />
										{:else}
											<ChevronRight size={16} />
										{/if}
									</button>
								{/if}
							</div>
						</td>
						<!-- Name -->
						<td class="text-left">
							{#if row.type === 'parent'}
								{row.cat.category_name}
							{:else}
								<span class="ml-8 text-sm opacity-90">{row.cat.category_name}</span>
							{/if}
						</td>
						<!-- Color (shared) -->
						<td>
							<div class="flex items-center justify-center space-x-2">
								<span
									class="inline-block h-6 w-6 rounded-full"
									style="background-color: {row.cat.color};"
								></span>
							</div>
						</td>
						<!-- Budget (shared) -->
						<td class={row.type === 'parent' ? '' : 'text-sm'}>
							{row.cat.budget ? formatCurrency(row.cat.budget, 0) : '—'}
						</td>
						<!-- Actions -->
						<td>
							<div class="flex items-center justify-center gap-1">
								{#if row.type === 'parent'}
									<button
										class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-success backdrop-blur-sm hover:bg-success/20"
										on:click={() => openCreateSubcategoryModal(row.cat)}
										title={$t('categories.add-subcategory')}
									>
										<Plus size={20} />
									</button>
								{/if}
								<button
									class="btn btn-circle btn-ghost btn-sm bg-base-100/80 backdrop-blur-sm"
									on:click={() => openEditCategoryModal(row.cat)}
									title={$t('common.edit')}
								>
									<Pencil size={row.type === 'parent' ? 20 : 18} />
								</button>
								<button
									class="btn btn-circle btn-ghost btn-sm bg-base-100/80 text-error backdrop-blur-sm hover:bg-error/20"
									on:click={() => openPromptDeleteCategoryModal(row.cat)}
									title={$t('common.delete')}
								>
									<Trash size={row.type === 'parent' ? 20 : 18} />
								</button>
							</div>
						</td>
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
