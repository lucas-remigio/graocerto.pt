<script lang="ts">
	import type { CategoryDto } from '$lib/types';

	export let categories: CategoryDto[] = [];
	export let categoryType: 'debit' | 'credit' = 'debit';

	const borderClasses: Record<string, string> = {
		credit: 'border-green-500 dark:border-green-400',
		debit: 'border-red-500 dark:border-red-400',
		transfer: 'border-blue-500 dark:border-blue-400'
	};
	let modalBorderClass = categoryType ? borderClasses[categoryType] : 'border-gray-50';
</script>

{#if categories.length === 0}
	<p class="text-gray-500">No categories available.</p>
{:else}
	<table class="table-zebra table w-full border-2 {modalBorderClass}">
		<thead>
			<tr>
				<th>Category Name</th>
				<th>Color</th>
			</tr>
		</thead>
		<tbody>
			{#each categories as category (category.id)}
				<tr>
					<td>{category.category_name}</td>
					<td>
						<div class="flex items-center space-x-2">
							<span
								class="inline-block h-4 w-4 rounded-full"
								style="background-color: {category.color};"
							></span>
							<span>{category.color}</span>
						</div>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}
