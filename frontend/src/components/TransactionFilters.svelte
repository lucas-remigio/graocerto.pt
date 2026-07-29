<script lang="ts">
	import type { CategoryDto, TransactionType } from '$lib/types';
	import { Search, X, Info } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { createEventDispatcher } from 'svelte';
	import { buildCategoryGroups } from '$lib/utils/categoryUtils';

	export let categories: CategoryDto[] = [];
	export let transactionTypes: TransactionType[] = [];
	export let show: boolean = false;
	export let filteredCount: number = 0;
	export let totalCount: number = 0;

	$: groupedCategories = buildCategoryGroups(categories);

	let searchTerm = '';
	let selectedCategory: number | null = null;
	let selectedType: string | null = null;
	let minAmount: number | null = null;
	let maxAmount: number | null = null;
	let startDate: string = '';
	let endDate: string = '';

	const dispatch = createEventDispatcher();

	$: activeFiltersCount = [
		searchTerm,
		selectedCategory,
		selectedType,
		minAmount,
		maxAmount,
		startDate,
		endDate
	].filter(Boolean).length;

	$: isFiltering = activeFiltersCount > 0;

	function applyFilters() {
		dispatch('filter', {
			searchTerm,
			categoryId: selectedCategory,
			typeSlug: selectedType,
			minAmount,
			maxAmount,
			startDate,
			endDate
		});
	}

	function clearFilters() {
		searchTerm = '';
		selectedCategory = null;
		selectedType = null;
		minAmount = null;
		maxAmount = null;
		startDate = '';
		endDate = '';
		applyFilters();
	}

	// Auto-apply on changes
	$: if (searchTerm !== undefined) applyFilters();
	$: if (selectedCategory !== undefined) applyFilters();
	$: if (selectedType !== undefined) applyFilters();
</script>

{#if show}
	<div class="mb-4 overflow-hidden rounded-xl border border-base-300 bg-base-100 shadow-md">
		<!-- Header -->
		<div class="bg-gradient-to-r from-primary/10 via-secondary/10 to-accent/10 px-4 py-3">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<Search size={18} class="text-primary" />
					<h3 class="font-semibold">{$t('transactions.filters')}</h3>
					{#if isFiltering}
						<span class="badge badge-primary badge-sm">{activeFiltersCount}</span>
					{/if}
				</div>
				{#if isFiltering}
					<button class="btn btn-ghost btn-sm gap-2" on:click={clearFilters}>
						<X size={16} />
						{$t('common.clear-all')}
					</button>
				{/if}
			</div>
		</div>

		<!-- Results Bar -->
		{#if isFiltering}
			<div class="border-b border-info/20 bg-info/10 px-4 py-2">
				<div class="flex items-center gap-2 text-sm">
					<Info size={16} class="text-info" />
					<span>
						{$t('transactions.showing')}
						<strong class="text-info">{filteredCount}</strong>
						{$t('transactions.of')}
						<strong>{totalCount}</strong>
						{$t('transactions.transactions-lowercase')}
					</span>
				</div>
			</div>
		{/if}

		<!-- Filters -->
		<div class="space-y-4 p-4">
			<!-- Search -->
			<div class="relative">
				<Search size={18} class="absolute left-3 top-1/2 -translate-y-1/2 text-base-content/40" />
				<input
					type="text"
					placeholder={$t('transactions.search-placeholder')}
					class="input input-bordered w-full rounded-lg pl-10 pr-10"
					bind:value={searchTerm}
				/>
				{#if searchTerm}
					<button
						class="btn btn-circle btn-ghost btn-sm absolute right-1 top-1/2 -translate-y-1/2"
						on:click={() => (searchTerm = '')}
					>
						<X size={16} />
					</button>
				{/if}
			</div>

			<!-- Grid -->
			<div class="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4">
				<!-- Category -->
				<select class="select select-bordered" bind:value={selectedCategory}>
					<option value={null}>{$t('transactions.category')}</option>
					{#each groupedCategories as group}
						{#if group.parent}
							{#if group.children.length > 0}
								<!-- Parent with children -->
								<option value={group.parent.id} class="font-semibold">
									{group.parent.category_name}
								</option>
								{#each group.children as child}
									<option value={child.id}>
										&nbsp;&nbsp;&nbsp;&nbsp;{child.category_name}
									</option>
								{/each}
							{:else}
								<!-- Parent without children -->
								<option value={group.parent.id}>
									{group.parent.category_name}
								</option>
							{/if}
						{:else}
							<!-- Orphaned children -->
							{#each group.children as child}
								<option value={child.id}>
									{child.category_name}
								</option>
							{/each}
						{/if}
					{/each}
				</select>

				<!-- Type -->
				<select class="select select-bordered" bind:value={selectedType}>
					<option value={null}>{$t('transactions.type')}</option>
					{#each transactionTypes as type}
						<option value={type.type_slug}>
							{$t(`categories.${type.type_slug}`)}
						</option>
					{/each}
				</select>

				<!-- Amount Range -->
				<input
					type="number"
					placeholder="€ Mínimo"
					class="input input-bordered"
					bind:value={minAmount}
					on:input={applyFilters}
				/>
				<input
					type="number"
					placeholder="€ Máximo"
					class="input input-bordered"
					bind:value={maxAmount}
					on:input={applyFilters}
				/>
			</div>

			<!-- Date Range -->
			<div class="grid grid-cols-2 gap-3">
				<input
					type="date"
					class="date-input input input-bordered"
					bind:value={startDate}
					on:change={applyFilters}
				/>
				<input
					type="date"
					class="date-input input input-bordered"
					bind:value={endDate}
					on:change={applyFilters}
				/>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(.dark) input[type='date']::-webkit-calendar-picker-indicator {
		filter: invert(1);
		cursor: pointer;
	}
	:global(.dark) input[type='date']::-moz-calendar-picker-indicator {
		filter: invert(1);
		cursor: pointer;
	}

	/* Light mode */
	input[type='date']::-webkit-calendar-picker-indicator {
		cursor: pointer;
	}
	input[type='date']::-moz-calendar-picker-indicator {
		cursor: pointer;
	}

	/* Remove cursor pointer from the input itself */
	.date-input {
		cursor: text;
	}
</style>
