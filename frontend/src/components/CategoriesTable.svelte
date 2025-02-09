<script lang="ts">
	import type { CategoryDto } from '$lib/types';

	export let categories: CategoryDto[] = [];
	export let categoryType: 'debit' | 'credit' = 'debit';
</script>

{#if categories.length === 0}
	<p class="text-gray-500">No categories available.</p>
{:else}
	<table
		class="table-zebra table w-full"
		class:bg-green-50={categoryType === 'credit'}
		class:bg-red-50={categoryType === 'debit'}
	>
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
