<script lang="ts">
	import api_axios from '$lib/axios';
	import { X } from 'lucide-svelte';
	import { createEventDispatcher, onMount } from 'svelte';

	export let title: string;
	export let message: string;

	export let onConfirm: () => void;
	export let onCancel: () => void;

	function handleConfirm() {
		onConfirm();
	}

	function handleCloseModal() {
		onCancel();
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative">
		<!-- Close button -->
		<button class="btn btn-sm btn-circle absolute right-2 top-2" on:click={handleCloseModal}
			><X /></button
		>
		<h3 class="mb-4 text-lg font-bold">{title}</h3>
		<p>{message}</p>
		<!-- Form Actions -->
		<div class="modal-action mt-6">
			<button type="button" class="btn" on:click={handleCloseModal}>Cancel</button>
			<button type="submit" class="btn btn-primary" on:click={handleConfirm}>Confirm</button>
		</div>
	</div>
</div>
