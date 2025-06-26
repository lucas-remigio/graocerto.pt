<script lang="ts">
	import { onMount } from 'svelte';
	import api_axios from '$lib/axios';
	import { dataService } from '$lib/services/dataService';
	import type {
		Category,
		CategoriesResponse,
		CategoryDto,
		CategoriesDtoResponse,
		TransactionType
	} from '$lib/types';
	import CategoriesTable from '$components/CategoriesTable.svelte';
	import { Plus, Tag } from 'lucide-svelte';
	import { TransactionTypes, TransactionTypeSlug } from '$lib/transaction_types_types';
	import CreateCategory from '$components/CreateCategory.svelte';
	import { t } from '$lib/i18n';

	let showCreateCategoryModal = false;
	let selectedTransactionType: TransactionType | undefined;

	let categories: CategoryDto[] = [];
	let debitCategories: CategoryDto[] = [];
	let creditCategories: CategoryDto[] = [];
	let error: string = '';
	let deleteError: string = '';

	async function fetchCategories() {
		try {
			const groupedCategories = await dataService.fetchCategoriesGrouped();

			// Convert grouped categories to flat arrays for each type
			categories = Object.values(groupedCategories).flat();

			// Filter categories by transaction type slug instead of hardcoded IDs
			debitCategories = categories.filter(
				(category) => category.transaction_type.type_slug === 'debit'
			);
			creditCategories = categories.filter(
				(category) => category.transaction_type.type_slug === 'credit'
			);
		} catch (err) {
			console.error('Error in fetchCategories:', err);
			error = $t('errors.failed-load-categories');
		}
	}

	async function deleteCategory(categoryId: number) {
		try {
			// We'll still use direct API call for delete since it's not a frequently cached operation
			const res = await api_axios.delete(`categories/${categoryId}`);

			if (res.status !== 200) {
				console.error('Non-200 response status:', res.status);
				showErrorMessage(res.data.error);
				return;
			}

			// Clear category caches and refetch
			dataService.clearCategoryCaches();
			fetchCategories();
		} catch (err: any) {
			console.error('Error in deleteCategory:', err);
			const error = err.response.data.error;
			showErrorMessage(error);
		}
	}

	function showErrorMessage(error: string) {
		deleteError = `Failed to delete category: ${error}`;
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

	function handleDeleteCategory(categoryId: number) {
		deleteCategory(categoryId);
	}

	onMount(async () => {
		fetchCategories();
	});
</script>

{#if error}
	<p class="text-red-500">{error}</p>
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
						<Plus size={20} class="text-base-content" />
						<Tag size={20} class="text-base-content" />
					</button>
				</div>
				<CategoriesTable
					categories={creditCategories}
					categoryType={TransactionTypeSlug.Credit}
					on:editCategory={fetchCategories}
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
						<Plus size={20} class="text-base-content" />
						<Tag size={20} class="text-base-content" />
					</button>
				</div>
				<CategoriesTable
					categories={debitCategories}
					categoryType={TransactionTypeSlug.Debit}
					on:editCategory={fetchCategories}
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
