<script lang="ts">
	import { X } from 'lucide-svelte';
	import { t } from '$lib/i18n';

	export let title: string;
	export let message: string;
	export let type: 'danger' | 'warning' | 'info' | 'success';

	export let onConfirm: () => void;
	export let onCancel: () => void;

	const typeStyles = {
		danger: {
			header: 'bg-red-500 text-white',
			button: 'btn-error text-white'
		},
		warning: {
			header: 'bg-yellow-500 text-white',
			button: 'btn-warning text-white'
		},
		info: {
			header: 'bg-blue-500 text-white',
			button: 'btn-info text-white'
		},
		success: {
			header: 'bg-green-500 text-white',
			button: 'btn-success text-white'
		}
	};

	function handleConfirm() {
		onConfirm();
	}

	function handleCloseModal() {
		onCancel();
	}
</script>

<div class="modal modal-open">
	<div class="modal-box relative p-0">
		<!-- Header with dynamic background -->
		<div class="px-6 py-4 {typeStyles[type].header}">
			<h3 class="text-lg font-bold">{title}</h3>
			<!-- Close button -->
			<button
				class="btn btn-sm btn-circle absolute right-2 top-2 border-none bg-transparent hover:bg-white/20"
				on:click={handleCloseModal}
			>
				<X class="text-white" />
			</button>
		</div>

		<!-- Message body -->
		<div class="p-6">
			<p class="py-4">{message}</p>

			<!-- Action buttons -->
			<div class="modal-action">
				<button class="btn border border-gray-900" on:click={handleCloseModal}
					>{$t('common.cancel')}</button
				>
				<button class="btn {typeStyles[type].button} border" on:click={handleConfirm}
					>{$t('modals.confirm')}</button
				>
			</div>
		</div>
	</div>
</div>
