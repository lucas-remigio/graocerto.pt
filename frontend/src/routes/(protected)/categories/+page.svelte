<script lang="ts">
	import { onMount } from 'svelte';
	import api_axios from '$lib/axios';
	import type {
		Category,
		CategoriesResponse,
		CategoryDto,
		CategoriesDtoResponse
	} from '$lib/types';
	import CategoriesTable from '$components/CategoriesTable.svelte';
	import { Tag } from 'lucide-svelte';
	import { TransactionTypes, TransactionTypeSlug } from '$lib/transaction_types_types';

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

	function openCreateCategoryModal() {
		console.log('Open create category modal');
	}

	onMount(async () => {
		fetchCategories();
	});
</script>

{#if error}
	<p class="text-red-500">{error}</p>
{:else}
	<div class="container mx-auto p-4">
		<!-- Button to create a new category -->
		<div class="flex justify-between">
			<h1 class="mb-6 text-3xl font-bold">My Categories</h1>
			<!-- button to create new account -->
			<button class="btn btn-primary" on:click={openCreateCategoryModal}><Tag /></button>
		</div>
		<div class="flex flex-col md:flex-row md:space-x-4">
			<!-- Credit Categories Table (Left) -->
			<div class="flex-1">
				<h2 class="mb-2 text-xl font-bold">Credit</h2>
				<CategoriesTable categories={creditCategories} categoryType={TransactionTypeSlug.Credit} />
			</div>

			<!-- Debit Categories Table (Right) -->
			<div class="mt-4 flex-1 md:mt-0">
				<h2 class="mb-2 text-xl font-bold">Debit</h2>
				<CategoriesTable categories={debitCategories} categoryType={TransactionTypeSlug.Debit} />
			</div>
		</div>
	</div>
{/if}
