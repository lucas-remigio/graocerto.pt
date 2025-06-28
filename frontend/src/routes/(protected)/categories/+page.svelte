<script lang="ts">
	import { onMount } from 'svelte';
	import { dataService } from '$lib/services/dataService';
	import type { Category, CategoryDto, TransactionType } from '$lib/types';
	import CategoriesTable from '$components/CategoriesTable.svelte';
	import { Plus, Tag } from 'lucide-svelte';
	import { TransactionTypes, TransactionTypeSlug } from '$lib/transaction_types_types';
	import CreateCategory from '$components/CreateCategory.svelte';
	import { t } from '$lib/i18n';

	let showCreateCategoryModal = false;
	let selectedTransactionType: TransactionType | undefined;

	let debitCategories: CategoryDto[] = [];
	let creditCategories: CategoryDto[] = [];
	let error: string = '';
	let deleteError: string = '';
	let loading: boolean = true;

	async function fetchCategories() {
		loading = true;
		try {
			creditCategories = [];
			debitCategories = [];
			const categories = await dataService.fetchCategories();

			// Use the grouped categories directly instead of filtering
			categories.forEach((c) => {
				switch (c.transaction_type.type_slug) {
					case TransactionTypeSlug.Credit:
						creditCategories.push(c);
						break;
					case TransactionTypeSlug.Debit:
						debitCategories.push(c);
						break;
					default:
						console.warn(`Unknown transaction type slug: ${c.transaction_type.type_slug}`);
				}
			});
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
			fetchCategories();
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
			await dataService.editCategory(categoryId, categoryData);
			fetchCategories();
		} catch (err: any) {
			console.error('Error in editCategory:', err);
			const error = err.message || 'Unknown error';
			showErrorMessage(error);
		}
	}

	function showErrorMessage(error: string) {
		deleteError = `Failed to process category: ${error}`;
		setTimeout(() => {
			deleteError = '';
		}, 5000);
	}

	function openCreateCategoryModal(transactionType: TransactionTypeSlug) {
		const matchingType = TransactionTypes.find((t) => t.type_slug === transactionType);
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

	function handleCreateCategorySuccess() {
		// Clear category caches and refetch
		dataService.clearCategoryCaches();
		fetchCategories();
		closeCreateCategoryModal();
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
		fetchCategories();
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
						on:click={() => openCreateCategoryModal(TransactionTypeSlug.Credit)}
						aria-label={$t('categories.create-category')}
					>
						<Plus size={20} class="text-base-100" />
						<Tag size={20} class="text-base-100" />
					</button>
				</div>
				<CategoriesTable
					categories={creditCategories}
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
						on:click={() => openCreateCategoryModal(TransactionTypeSlug.Debit)}
						aria-label={$t('categories.create-category')}
					>
						<Plus size={20} class="text-base-100" />
						<Tag size={20} class="text-base-100" />
					</button>
				</div>
				<CategoriesTable
					categories={debitCategories}
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
	<CreateCategory
		transactionType={selectedTransactionType}
		on:closeModal={closeCreateCategoryModal}
		on:newCategory={handleCreateCategorySuccess}
	/>
{/if}
