<script lang="ts">
	import api_axios from '$lib/axios';
	import { toastStore } from '$lib/stores/toast';
	import type { Account, AiFeedbackResponse } from '$lib/types';
	import {
		X,
		Lightbulb,
		BarChart,
		PieChart,
		CalendarClock,
		ArrowUpRight,
		Loader2,
		AlertCircle
	} from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { locale, t } from 'svelte-i18n';
	import { fade, fly } from 'svelte/transition';

	export let account: Account;
	export let closeModal: () => void;
	export let month: number;
	export let year: number;

	let isLoading: boolean = true;
	let feedbackMessage: string = '';
	let inDepthAnalysis: string = '';

	async function getTransactionsAiFeedback() {
		isLoading = true;
		const language = $locale || 'pt';
		try {
			const res = await api_axios('accounts/' + account.token + '/feedback-month', {
				params: {
					month,
					year,
					language
				}
			});

			if (res.status !== 200) {
				console.error('Non-200 response status for transactions:', res.status);
				toastStore.error(`Error: ${res.status}`);
				return;
			}

			const data: AiFeedbackResponse = res.data;
			feedbackMessage = data.feedback_message;
			inDepthAnalysis = data.in_depth_analysis;
		} catch (err) {
			console.error('Error in getAccountTransactions:', err);
			toastStore.error('Failed to load transactions');
		} finally {
			isLoading = false;
		}
	}

	function formattedDate(): string {
		const date = new Date(year, month - 1); // month is 0-indexed in JS
		return date.toLocaleDateString(currentLocale, {
			month: 'long',
			year: 'numeric'
		});
	}

	function handleCloseModal() {
		closeModal();
	}

	$: currentLocale = $locale || 'pt';

	onMount(() => {
		getTransactionsAiFeedback();
	});
</script>

<div class="modal modal-open">
	<div
		class="modal-box relative flex max-h-[90vh] max-w-3xl flex-col overflow-hidden p-0"
		transition:fly={{ y: 20, duration: 300 }}
	>
		<!-- Gradient header (fixed) -->
		<div class="sticky top-0 z-10 bg-gradient-to-r from-primary to-secondary px-6 py-5 shadow-lg">
			<div class="flex items-center gap-2">
				<BarChart class="h-5 w-5 text-base-100" />
				<h3 class="text-xl font-bold text-base-100">
					{$t('ai-feedback.title') + ' ' + account.account_name}
				</h3>
			</div>
			<!-- Close button -->
			<button
				class="btn btn-circle btn-sm absolute right-2 top-2 border-none bg-base-100/20 hover:bg-base-100/30"
				on:click={handleCloseModal}
			>
				<X class="text-base-100" />
			</button>
			<div class="mt-1 flex items-center gap-2 text-sm text-base-100/80">
				<CalendarClock class="h-4 w-4 text-base-100" />
				<span>{$t('ai-feedback.monthly-analysis-for') + ' ' + formattedDate()}</span>
			</div>
		</div>

		<!-- Message body (scrollable) -->
		<div class="overflow-y-auto p-6">
			{#if isLoading}
				<div class="flex flex-col items-center justify-center py-12" in:fade>
					<Loader2 class="mb-4 h-12 w-12 animate-spin text-primary" />
					<p class="text-base-content/70">{$t('ai-feedback.analyzing')}</p>
				</div>
			{:else}
				<div in:fade={{ duration: 300, delay: 100 }}>
					<!-- Summary section -->
					<div
						class="mb-6 rounded-lg border border-primary/20 bg-gradient-to-r from-primary/5 to-secondary/5 p-5 shadow-sm"
					>
						<div class="flex items-start gap-3">
							<div class="rounded-full bg-primary/10 p-2">
								<Lightbulb class="h-5 w-5 text-primary" />
							</div>
							<div>
								<h3 class="mb-2 text-lg font-medium text-primary">
									{$t('ai-feedback.key-insights')}
								</h3>
								<p class="text-base-content/80">{feedbackMessage}</p>
							</div>
						</div>
					</div>

					<!-- Detailed analysis section -->
					<div class="mt-6">
						<div class="mb-4 flex items-center gap-2">
							<PieChart class="h-5 w-5 text-secondary" />
							<h3 class="text-lg font-semibold text-base-content">
								{$t('ai-feedback.detailed-analysis')}
							</h3>
						</div>

						<div
							class="prose prose-sm max-w-none whitespace-pre-line leading-relaxed text-base-content/80"
						>
							{inDepthAnalysis}
						</div>

						<!-- End card with upward trend icon -->
						<div class="mt-8 flex justify-end">
							<div class="inline-flex items-center gap-1.5 text-sm text-success">
								<ArrowUpRight class="h-4 w-4" />
								<span>{$t('ai-feedback.based-on-history')}</span>
							</div>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	/* Animation for loader */
	:global(.animate-spin) {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}
</style>
