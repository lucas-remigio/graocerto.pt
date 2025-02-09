<script lang="ts">
	import api_axios from '$lib/axios';
	import type { Category, TransactionType } from '$lib/types';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';

	export let transactionType: TransactionType;

	let error: string = '';
	// Form field variables
	let category_name: string = '';
	let color: string = '';

	// Create event dispatcher (to emit events to the parent)
	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		error = '';
		if (!validateForm()) {
			return;
		}

		const category = {
			transaction_type_id: transactionType.id,
			category_name: category_name,
			color: color
		};

		try {
			const response = await api_axios.post('categories', category);

			if (response.status !== 200) {
				console.error('Non-200 response status:', response.status);
				error = `Error: ${response.status}`;
				return;
			}

			handleNewCategory();
		} catch (err) {
			console.error('Error in handleSubmit:', err);
			error = 'Failed to create category';
		}
	}

	function validateForm(): boolean {
		if (!category_name) {
			error = 'Category name is required';
			return false;
		}

		category_name = category_name.trim();

		if (category_name.length > 50) {
			error = 'Category name must be less than 50 characters';
			return false;
		}

		if (category_name.length < 3) {
			error = 'Category name must be at least 3 characters';
			return false;
		}

		if (!color) {
			error = 'Color is required';
			return false;
		}

		color = color.trim();

		if (color[0] !== '#') {
			error = 'Color must be a valid hex color';
			return false;
		}

		if (color.length !== 7) {
			error = 'Color must be a valid hex color';
			return false;
		}

		return true;
	}

	const backgroundClasses: Record<string, string> = {
		credit: 'bg-green-50',
		debit: 'bg-red-50',
		transfer: 'bg-blue-50'
	};

	let modalBackgroundClass = backgroundClasses[transactionType.type_slug] || 'bg-gray-50';

	function handleCloseModal() {
		dispatch('closeModal');
	}

	function handleNewCategory() {
		dispatch('newCategory');
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative {modalBackgroundClass}">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}>
			<X />
		</button>
		<h3 class="mb-4 text-lg font-bold">
			New {transactionType.type_name} Category
		</h3>
		<!-- Error message -->
		{#if error}
			<div class="alert alert-error">
				<p class="text-gray-100">{error}</p>
			</div>
		{/if}
		<form on:submit|preventDefault={handleSubmit}>
			<!-- Category Name Field -->
			<div class="form-control mt-4">
				<label class="label" for="category_name">
					<span class="label-text">Category Name</span>
				</label>
				<input
					id="category_name"
					type="text"
					placeholder="Category name"
					class="input input-bordered"
					bind:value={category_name}
					required
				/>
			</div>

			<!-- Color Field -->
			<div class="form-control mt-4">
				<label class="label" for="color">
					<span class="label-text">Color</span>
				</label>
				<div class="flex items-center space-x-4">
					<!-- The native color input -->
					<input
						id="color"
						type="color"
						class="h-10 w-80 border-0 p-0"
						bind:value={color}
						required
					/>
					<!-- A swatch preview showing the chosen color -->
					<div
						class="h-8 w-8 rounded-full border border-gray-300"
						style="background-color: {color};"
					></div>
					<!-- Display the hex value -->
					<span class="text-sm font-medium">{color}</span>
				</div>
			</div>

			<!-- Form Actions -->
			<div class="modal-action mt-6">
				<button type="button" class="btn" on:click={handleCloseModal}>Cancel</button>
				<button type="submit" class="btn btn-primary">Create Category</button>
			</div>
		</form>
	</div>
</div>
