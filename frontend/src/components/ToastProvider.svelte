<script lang="ts">
	import { toastStore } from '$lib/stores/toast';
	import { CheckCircle2, AlertCircle, AlertTriangle, Info, X } from 'lucide-svelte';
	import { slide } from 'svelte/transition';

	function getAlertClass(type: string) {
		switch (type) {
			case 'success':
				return 'alert-success text-base-100';
			case 'error':
				return 'alert-error text-base-100';
			case 'warning':
				return 'alert-warning';
			case 'info':
				return 'alert-info text-base-100';
			default:
				return 'bg-base-200 text-base-content';
		}
	}

	function getIcon(type: string) {
		switch (type) {
			case 'success':
				return CheckCircle2;
			case 'error':
				return AlertCircle;
			case 'warning':
				return AlertTriangle;
			case 'info':
				return Info;
			default:
				return Info;
		}
	}
</script>

<div
	class="toast toast-center pointer-events-none z-[100] mb-4 sm:toast-end sm:toast-bottom sm:mb-8"
>
	{#each $toastStore as toast (toast.id)}
		<div
			class="alert {getAlertClass(
				toast.type
			)} pointer-events-auto flex cursor-pointer items-center justify-between gap-3 rounded-xl font-medium shadow-lg"
			transition:slide={{ duration: 250, axis: 'y' }}
			onclick={() => toastStore.remove(toast.id)}
			onkeydown={(e) => e.key === 'Enter' && toastStore.remove(toast.id)}
			role="button"
			tabindex="0"
		>
			<div class="flex items-center gap-2">
				<svelte:component this={getIcon(toast.type)} size={18} class="shrink-0" />
				<span class="text-sm">{toast.message}</span>
			</div>
			<button
				class="btn btn-circle btn-ghost btn-xs ml-2 opacity-70 hover:opacity-100"
				onclick={(e) => {
					e.stopPropagation();
					toastStore.remove(toast.id);
				}}
				aria-label="Close"
			>
				<X size={14} />
			</button>
		</div>
	{/each}
</div>
