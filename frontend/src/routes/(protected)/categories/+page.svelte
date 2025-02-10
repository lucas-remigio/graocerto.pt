<script lang="ts">
	import { onMount } from 'svelte';
	import api_axios from '$lib/axios';
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

	let showCreateCategoryModal = false;
	let selectedTransactionType: TransactionType | undefined;

	let categories: CategoryDto[] = [];
	let debitCategories: CategoryDto[] = [];
	let creditCategories: CategoryDto[] = [];
	let error: string = '';

	async function fetchCategories() {
		try {
			const res = await api_axios.get('categories/dto');

			if (res.status !== 200) {
				console.error('Non-200 response status:', res.status);
				error = `Error: ${res.status}`;
				return;
			}

			const data: CategoriesDtoResponse = res.data;
			categories = data.categories;

			debitCategories = categories.filter(
				(category) => category.transaction_type.type_slug === 'debit'
			);
			creditCategories = categories.filter(
				(category) => category.transaction_type.type_slug === 'credit'
			);
		} catch (err) {
			console.error('Error in fetchCategories:', err);
			error = 'Failed to load categories';
		}
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
		fetchCategories();
		closeCreateCategoryModal();
	}

	onMount(async () => {
		fetchCategories();
	});
</script>

{#if error}
	<p class="text-red-500">{error}</p>
{:else}
	<div class="container mx-auto p-4">
		<h1 class="mb-6 text-3xl font-bold">My Categories</h1>
		<div class="flex flex-col md:flex-row md:space-x-4">
			<!-- Credit Categories Table (Left) -->
			<div class="flex-1">
				<div class="mb-2 flex items-center justify-between">
					<h2 class="text-xl font-bold">Credit</h2>
					<button
						class="btn btn-primary flex items-center gap-2"
						on:click={() => openCreateCategoryModal(TransactionTypeSlug.Credit)}
						aria-label="Create New Credit Category"
					>
						<Plus size={20} />
						<Tag size={20} />
					</button>
				</div>
				<CategoriesTable categories={creditCategories} categoryType={TransactionTypeSlug.Credit} />
			</div>

			<!-- Debit Categories Table (Right) -->
			<div class="mt-4 flex-1 md:mt-0">
				<div class="mb-2 flex items-center justify-between">
					<h2 class="text-xl font-bold">Debit</h2>
					<button
						class="btn btn-primary flex items-center gap-2"
						on:click={() => openCreateCategoryModal(TransactionTypeSlug.Debit)}
						aria-label="Create New Debit Category"
					>
						<Plus size={20} />
						<Tag size={20} />
					</button>
				</div>
				<CategoriesTable categories={debitCategories} categoryType={TransactionTypeSlug.Debit} />
			</div>
		</div>
	</div>
	{#if showCreateCategoryModal && selectedTransactionType}
		<CreateCategory
			transactionType={selectedTransactionType}
			on:closeModal={closeCreateCategoryModal}
			on:newCategory={handleCreateCategorySuccess}
		/>
	{/if}
{/if}
