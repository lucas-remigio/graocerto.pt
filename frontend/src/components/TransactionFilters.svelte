<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { CategoryDto } from '$lib/types';
	import { Filter } from 'lucide-svelte';

	export let categories: CategoryDto[] = [];

	const dispatch = createEventDispatcher<{
		filter: {
			dateFrom: string;
			dateTo: string;
			transactionTypes: string[];
			categories: number[];
			amountFrom: number | null;
			amountTo: number | null;
		};
	}>();

	let showFilters = false;
	let dateFrom = '';
	let dateTo = '';
	let selectedTypes: string[] = [];
	let selectedCategories: number[] = [];
	let amountFrom: string = '';
	let amountTo: string = '';

	const transactionTypes = [
		{ slug: 'credit', name: 'Credit' },
		{ slug: 'debit', name: 'Debit' }
	];

	function handleFilter() {
		dispatch('filter', {
			dateFrom,
			dateTo,
			transactionTypes: selectedTypes,
			categories: selectedCategories,
			amountFrom: amountFrom ? parseFloat(amountFrom) : null,
			amountTo: amountTo ? parseFloat(amountTo) : null
		});
	}

	function clearFilters() {
		dateFrom = '';
		dateTo = '';
		selectedTypes = [];
		selectedCategories = [];
		amountFrom = '';
		amountTo = '';
		handleFilter();
	}
</script>

<div class="mb-4">
	<button
		class="btn btn-ghost gap-2"
		on:click={() => (showFilters = !showFilters)}
		class:btn-active={showFilters}
	>
		<Filter size={20} />
		Filters
	</button>

	{#if showFilters}
		<div class="card bg-base-100 mt-2 p-4 shadow-lg">
			<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
				<!-- Date Range -->
				<div class="form-control">
					<label class="label" for="dateFrom">
						<span class="label-text">From Date</span>
					</label>
					<input type="date" id="dateFrom" class="input input-bordered" bind:value={dateFrom} />
				</div>

				<div class="form-control">
					<label class="label" for="dateTo">
						<span class="label-text">To Date</span>
					</label>
					<input type="date" id="dateTo" class="input input-bordered" bind:value={dateTo} />
				</div>

				<!-- Transaction Types -->
				<div class="form-control">
					<label class="label">
						<span class="label-text">Transaction Types</span>
					</label>
					<div class="flex flex-wrap gap-2">
						{#each transactionTypes as type}
							<label class="label cursor-pointer gap-2">
								<input
									type="checkbox"
									class="checkbox"
									bind:group={selectedTypes}
									value={type.slug}
								/>
								<span class="label-text">{type.name}</span>
							</label>
						{/each}
					</div>
				</div>

				<!-- Categories -->
				<div class="form-control">
					<label class="label">
						<span class="label-text">Categories</span>
					</label>
					<div class="flex flex-wrap gap-2">
						{#each categories as category}
							<label class="label cursor-pointer gap-2">
								<input
									type="checkbox"
									class="checkbox"
									bind:group={selectedCategories}
									value={category.id}
								/>
								<span class="label-text">{category.category_name}</span>
							</label>
						{/each}
					</div>
				</div>

				<!-- Amount Range -->
				<div class="form-control">
					<label class="label" for="amountFrom">
						<span class="label-text">Min Amount</span>
					</label>
					<input
						type="number"
						id="amountFrom"
						class="input input-bordered"
						bind:value={amountFrom}
						step="0.01"
						min="0"
					/>
				</div>

				<div class="form-control">
					<label class="label" for="amountTo">
						<span class="label-text">Max Amount</span>
					</label>
					<input
						type="number"
						id="amountTo"
						class="input input-bordered"
						bind:value={amountTo}
						step="0.01"
						min="0"
					/>
				</div>
			</div>

			<!-- Action Buttons -->
			<div class="mt-4 flex justify-end gap-2">
				<button class="btn btn-ghost" on:click={clearFilters}>Clear</button>
				<button class="btn btn-primary" on:click={handleFilter}>Apply Filters</button>
			</div>
		</div>
	{/if}
</div>
