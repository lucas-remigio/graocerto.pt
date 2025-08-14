<script lang="ts">
	import { onMount } from 'svelte';
	import { dataService } from '$lib/services/dataService';
	import type { Category, CategoryChangeResponse, CategoryDto, TransactionType } from '$lib/types';
	import CategoriesTable from '$components/CategoriesTable.svelte';
	import { Plus, Tag } from 'lucide-svelte';
	import {
		TransactionTypeId,
		TransactionTypes,
		TransactionTypeSlug
	} from '$lib/transaction_types_types';
	import CategoryModal from '$components/CategoryModal.svelte';
	import { t } from '$lib/i18n';

	let showCreateCategoryModal = $state(false);
	let selectedTransactionType: TransactionType | undefined = $state(undefined);

	let categories: CategoryDto[] = $state([]);
	let error: string = $state('');
	let deleteError: string = $state('');
	let loading: boolean = $state(true);

	// Derived lists for credit and debit categories
	let creditCategories = $derived(() =>
		categories.filter((c) => c.transaction_type.type_slug === TransactionTypeSlug.Credit)
	);
	let debitCategories = $derived(() =>
		categories.filter((c) => c.transaction_type.type_slug === TransactionTypeSlug.Debit)
	);

	async function fetchCategories(showLoading: boolean) {
		loading = showLoading;
		try {
			const fetched = await dataService.fetchCategories();
			categories = fetched;
		} catch (err) {
			console.error('Error in fetchCategories:', err);
			error = $t('errors.failed-load-categories');
		} finally {
			loading = false;
		}
	}

	async function deleteCategory(categoryId: number) {
		try {
			await dataService.deleteCategory(categoryId);
			categories = categories.filter((c) => c.id !== categoryId);
		} catch (err: any) {
			console.error('Error in deleteCategory:', err);
			const error = err.message || 'Unknown error';
			showErrorMessage(error);
		}
	}

	async function editCategory(
		categoryId: number,
		categoryData: { category_name: string; color: string }
	) {
		try {
			const response = await dataService.editCategory(categoryId, categoryData);
			updateCategory(response.category);
		} catch (err: any) {
			console.error('Error in editCategory:', err);
			const error = err.message || 'Unknown error';
			showErrorMessage(error);
		}
	}

	function updateCategory(category: CategoryDto): void {
		const idx = categories.findIndex((c) => c.id === category.id);
		if (idx !== -1) {
			categories[idx] = category;
		}
	}

	function showErrorMessage(error: string) {
		deleteError = `Failed to process category: ${error}`;
		setTimeout(() => {
			deleteError = '';
		}, 5000);
	}

	function openCreateCategoryModal(transactionType: TransactionTypeId) {
		const matchingType = TransactionTypes.find((t) => t.id === transactionType);
		selectedTransactionType = matchingType;

		if (!selectedTransactionType) {
			console.error('Could not find transaction type id for:', transactionType);
			return;
		}

		showCreateCategoryModal = true;
	}

	function closeCreateCategoryModal() {
		showCreateCategoryModal = false;
	}

	function handleCreateCategorySuccess(event: CustomEvent<CategoryChangeResponse>) {
		// Clear category caches and refetch
		closeCreateCategoryModal();
		dataService.clearCategoryCaches();
		categories.push(event.detail.category);
		sortCategories();
	}

	function sortCategories() {
		// newest to oldest
		categories.sort((a, b) => b.id - a.id);
	}

	function handleEditCategorySuccess(
		event: CustomEvent<{
			categoryId: number;
			categoryData: { category_name: string; color: string };
		}>
	) {
		const { categoryId, categoryData } = event.detail;
		editCategory(categoryId, categoryData);
	}

	function handleDeleteCategory(categoryId: number) {
		deleteCategory(categoryId);
	}

	onMount(async () => {
		fetchCategories(true);
	});
</script>

{#if error}
	<p class="text-red-500">{error}</p>
{:else if loading}
	<div class="container mx-auto p-4">
		<h1 class="mb-6 text-3xl font-bold">{$t('navbar.categories')}</h1>
		<div class="flex min-h-64 items-center justify-center">
			<div class="loading loading-spinner loading-lg text-primary"></div>
		</div>
	</div>
{:else}
	<div class="container mx-auto p-4">
		{#if deleteError}
			<p class="text-red-500">{deleteError}</p>
		{/if}
		<h1 class="mb-6 text-3xl font-bold">{$t('navbar.categories')}</h1>
		<div class="flex flex-col md:flex-row md:space-x-4">
			<!-- Credit Categories Table (Left) -->
			<div class="flex-1">
				<div class="mb-2 flex items-center justify-between">
					<h2 class="text-xl font-bold">{$t('categories.credit')}</h2>
					<button
						class="btn btn-primary flex items-center gap-2"
						onclick={() => openCreateCategoryModal(TransactionTypeId.Credit)}
						aria-label={$t('categories.create-category')}
					>
						<Plus size={20} class="text-base-100" />
						<Tag size={20} class="text-base-100" />
					</button>
				</div>
				<CategoriesTable
					categories={creditCategories()}
					categoryType={TransactionTypeSlug.Credit}
					on:editCategory={handleEditCategorySuccess}
					on:deleteCategory={({ detail: { categoryId } }) => {
						handleDeleteCategory(categoryId);
					}}
				/>
			</div>

			<!-- Debit Categories Table (Right) -->
			<div class="mt-4 flex-1 md:mt-0">
				<div class="mb-2 flex items-center justify-between">
					<h2 class="text-xl font-bold">{$t('categories.debit')}</h2>
					<button
						class="btn btn-primary flex items-center gap-2"
						onclick={() => openCreateCategoryModal(TransactionTypeId.Debit)}
						aria-label={$t('categories.create-category')}
					>
						<Plus size={20} class="text-base-100" />
						<Tag size={20} class="text-base-100" />
					</button>
				</div>
				<CategoriesTable
					categories={debitCategories()}
					categoryType={TransactionTypeSlug.Debit}
					on:editCategory={handleEditCategorySuccess}
					on:deleteCategory={({ detail: { categoryId } }) => {
						handleDeleteCategory(categoryId);
					}}
				/>
			</div>
		</div>
	</div>
{/if}

{#if showCreateCategoryModal && selectedTransactionType}
	<CategoryModal
		category={null}
		transactionType={selectedTransactionType}
		on:closeModal={closeCreateCategoryModal}
		on:newCategory={handleCreateCategorySuccess}
	/>
{/if}
