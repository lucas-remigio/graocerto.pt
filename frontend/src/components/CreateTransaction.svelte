<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	let description = '';
	let amount: number = 0;

	function handleSubmit() {
		// Process the form data to create a new transaction
		// For example, call an API or update a store
		console.log('Creating transaction with:', { description, amount });

		// After creating the transaction, you can reset fields and close the modal
		description = '';
		amount = 0;
	}

	const dispatch = createEventDispatcher<any>();
	function handleCloseModal() {
		dispatch('closeModal');
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}>✕</button>
		<h3 class="text-lg font-bold">New Transaction</h3>
		<!-- Add your transaction form here -->
		<form on:submit|preventDefault={handleSubmit}>
			<div class="form-control">
				<label class="label">
					<span class="label-text">Description</span>
				</label>
				<input
					type="text"
					placeholder="Transaction description"
					class="input input-bordered"
					bind:value={description}
					required
				/>
			</div>
			<div class="form-control mt-4">
				<label class="label">
					<span class="label-text">Amount</span>
				</label>
				<input
					type="number"
					placeholder="Amount"
					class="input input-bordered"
					bind:value={amount}
					required
				/>
			</div>
			<!-- Add more fields as needed -->
			<div class="modal-action">
				<button type="button" class="btn" on:click={handleCloseModal}>Cancel</button>
				<button type="submit" class="btn btn-primary">Create Transaction</button>
			</div>
		</form>
	</div>
</div>
