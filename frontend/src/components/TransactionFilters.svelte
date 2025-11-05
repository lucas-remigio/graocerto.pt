<script lang="ts">
	import type { Category, CategoryDto, TransactionType } from '$lib/types';
	import { Search, X, Calendar, DollarSign, Tag, FileText, Info } from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { createEventDispatcher } from 'svelte';

	export let categories: CategoryDto[] = [];
	export let transactionTypes: TransactionType[] = [];
	export let show: boolean = false;
	export let filteredCount: number = 0;
	export let totalCount: number = 0;

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
	<div class="bg-base-100 border-base-300 mb-6 overflow-hidden rounded-2xl border shadow-lg">
		<!-- Header with Gradient -->
		<div class="from-primary/10 via-secondary/10 to-accent/10 bg-gradient-to-r px-6 py-4">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-3">
					<div class="bg-base-100 rounded-lg p-2 shadow-sm">
						<Search size={20} class="text-primary" />
					</div>
					<div>
						<h3 class="text-lg font-bold">{$t('transactions.filters')}</h3>
						{#if isFiltering}
							<p class="text-xs opacity-70">
								{activeFiltersCount}
								{activeFiltersCount === 1 ? 'filtro ativo' : 'filtros ativos'}
							</p>
						{/if}
					</div>
				</div>
				{#if isFiltering}
					<button class="btn btn-ghost btn-sm gap-2 rounded-lg" on:click={clearFilters}>
						<X size={16} />
						{$t('common.clear-all', { default: 'Limpar Tudo' })}
					</button>
				{/if}
			</div>
		</div>

		<!-- Results Summary Bar (when filtering) -->
		{#if isFiltering}
			<div class="bg-info/10 border-info/20 border-b px-6 py-3">
				<div class="flex items-center gap-2 text-sm">
					<Info size={16} class="text-info" />
					<span class="font-medium">
						{$t('transactions.showing')}
						<strong class="text-info">{filteredCount}</strong>
						{$t('transactions.of')}
						<strong>{totalCount}</strong>
						{$t('transactions.transactions-lowercase')}
					</span>
				</div>
			</div>
		{/if}

		<!-- Filters Grid -->
		<div class="space-y-6 p-6">
			<!-- Search - Full Width -->
			<div class="form-control">
				<label class="label" for="search-filter">
					<span class="label-text flex items-center gap-2 font-semibold">
						<FileText size={16} class="text-primary" />
						{$t('transactions.search-description', { default: 'Pesquisar descrição' })}
					</span>
				</label>
				<div class="relative">
					<Search size={18} class="text-base-content/40 absolute left-4 top-1/2 -translate-y-1/2" />
					<input
						id="search-filter"
						type="text"
						placeholder={$t('transactions.search-placeholder', {
							default: 'Digite para pesquisar...'
						})}
						class="input input-bordered w-full rounded-xl pl-12 pr-12 shadow-sm transition-all focus:shadow-md"
						bind:value={searchTerm}
					/>
					{#if searchTerm}
						<button
							class="btn btn-ghost btn-sm btn-circle absolute right-2 top-1/2 -translate-y-1/2"
							on:click={() => (searchTerm = '')}
						>
							<X size={16} />
						</button>
					{/if}
				</div>
			</div>

			<!-- Divider -->
			<div class="divider my-2 opacity-50">Filtros Avançados</div>

			<!-- Category & Type -->
			<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
				<!-- Category Filter -->
				<div class="form-control">
					<label class="label" for="category-filter">
						<span class="label-text flex items-center gap-2 font-semibold">
							<Tag size={16} class="text-secondary" />
							{$t('transactions.category')}
						</span>
					</label>
					<div class="relative">
						<Tag size={16} class="text-base-content/40 absolute left-4 top-1/2 -translate-y-1/2" />
						<select
							id="category-filter"
							class="select select-bordered w-full rounded-xl pl-11 shadow-sm transition-all focus:shadow-md"
							bind:value={selectedCategory}
						>
							<option value={null}>{$t('common.all')}</option>
							{#each categories as category}
								<option value={category.id}>
									{category.parent_category_id ? '  ↳ ' : ''}{category.category_name}
								</option>
							{/each}
						</select>
					</div>
				</div>

				<!-- Transaction Type Filter -->
				<div class="form-control">
					<label class="label" for="type-filter">
						<span class="label-text flex items-center gap-2 font-semibold">
							<FileText size={16} class="text-accent" />
							{$t('transactions.type', { default: 'Tipo' })}
						</span>
					</label>
					<div class="relative">
						<FileText
							size={16}
							class="text-base-content/40 absolute left-4 top-1/2 -translate-y-1/2"
						/>
						<select
							id="type-filter"
							class="select select-bordered w-full rounded-xl pl-11 shadow-sm transition-all focus:shadow-md"
							bind:value={selectedType}
						>
							<option value={null}>{$t('common.all')}</option>
							{#each transactionTypes as type}
								<option value={type.type_slug}>{type.type_name}</option>
							{/each}
						</select>
					</div>
				</div>
			</div>

			<!-- Amount Range -->
			<div class="form-control">
				<label class="label">
					<span class="label-text flex items-center gap-2 font-semibold">
						<DollarSign size={16} class="text-success" />
						{$t('transactions.amount-range', { default: 'Intervalo de Valor' })}
					</span>
				</label>
				<div class="grid grid-cols-2 gap-4">
					<div class="relative">
						<span class="text-base-content/40 absolute left-4 top-1/2 -translate-y-1/2 text-sm"
							>€</span
						>
						<input
							type="number"
							placeholder="Mínimo"
							class="input input-bordered w-full rounded-xl pl-9 shadow-sm transition-all focus:shadow-md"
							bind:value={minAmount}
							on:input={applyFilters}
						/>
					</div>
					<div class="relative">
						<span class="text-base-content/40 absolute left-4 top-1/2 -translate-y-1/2 text-sm"
							>€</span
						>
						<input
							type="number"
							placeholder="Máximo"
							class="input input-bordered w-full rounded-xl pl-9 shadow-sm transition-all focus:shadow-md"
							bind:value={maxAmount}
							on:input={applyFilters}
						/>
					</div>
				</div>
			</div>

			<!-- Date Range -->
			<div class="form-control">
				<label class="label">
					<span class="label-text flex items-center gap-2 font-semibold">
						<Calendar size={16} class="text-info" />
						{$t('transactions.date-range', { default: 'Período' })}
					</span>
				</label>
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
					<div class="form-control">
						<label class="label" for="start-date-input">
							<span class="label-text text-xs"
								>{$t('transactions.start-date', { default: 'Data Inicial' })}</span
							>
						</label>
						<input
							type="date"
							id="start-date-input"
							class="input input-bordered hover:border-info date-input w-full transition-all hover:shadow-md"
							bind:value={startDate}
						/>
					</div>
					<div class="form-control">
						<label class="label" for="end-date-input">
							<span class="label-text text-xs"
								>{$t('transactions.end-date', { default: 'Data Final' })}</span
							>
						</label>
						<input
							type="date"
							id="end-date-input"
							class="input input-bordered hover:border-info date-input w-full transition-all hover:shadow-md"
							bind:value={endDate}
						/>
					</div>
				</div>
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
