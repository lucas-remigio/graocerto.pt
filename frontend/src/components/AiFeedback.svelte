<script lang="ts">
	import api_axios from '$lib/axios';
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

	let error: string = '';
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
				error = `Error: ${res.status}`;
				return;
			}

			const data: AiFeedbackResponse = res.data;
			feedbackMessage = data.feedback_message;
			inDepthAnalysis = data.in_depth_analysis;
		} catch (err) {
			console.error('Error in getAccountTransactions:', err);
			error = 'Failed to load transactions';
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
		<div class="from-primary to-secondary sticky top-0 z-10 bg-gradient-to-r px-6 py-5 shadow-lg">
			<div class="flex items-center gap-2">
				<BarChart class="text-base-100 h-5 w-5" />
				<h3 class="text-base-100 text-xl font-bold">
					{$t('ai-feedback.title') + ' ' + account.account_name}
				</h3>
			</div>
			<!-- Close button -->
			<button
				class="btn btn-sm btn-circle bg-base-100/20 hover:bg-base-100/30 absolute right-2 top-2 border-none"
				on:click={handleCloseModal}
			>
				<X class="text-base-100" />
			</button>
			<div class="text-base-100/80 mt-1 flex items-center gap-2 text-sm">
				<CalendarClock class="text-base-100 h-4 w-4" />
				<span>{$t('ai-feedback.monthly-analysis-for') + ' ' + formattedDate()}</span>
			</div>
		</div>

		<!-- Message body (scrollable) -->
		<div class="overflow-y-auto p-6">
			{#if isLoading}
				<div class="flex flex-col items-center justify-center py-12" in:fade>
					<Loader2 class="text-primary mb-4 h-12 w-12 animate-spin" />
					<p class="text-base-content/70">{$t('ai-feedback.analyzing')}</p>
				</div>
			{:else if error}
				<div class="border-error bg-error/10 rounded border-l-4 p-4" in:fade>
					<p class="text-error flex items-center gap-2">
						<AlertCircle class="text-error h-5 w-5" />
						{$t('ai-feedback.error') + ' ' + error}
					</p>
				</div>
			{:else}
				<div in:fade={{ duration: 300, delay: 100 }}>
					<!-- Summary section -->
					<div
						class="border-primary/20 from-primary/5 to-secondary/5 mb-6 rounded-lg border bg-gradient-to-r p-5 shadow-sm"
					>
						<div class="flex items-start gap-3">
							<div class="bg-primary/10 rounded-full p-2">
								<Lightbulb class="text-primary h-5 w-5" />
							</div>
							<div>
								<h3 class="text-primary mb-2 text-lg font-medium">
									{$t('ai-feedback.key-insights')}
								</h3>
								<p class="text-base-content/80">{feedbackMessage}</p>
							</div>
						</div>
					</div>

					<!-- Detailed analysis section -->
					<div class="mt-6">
						<div class="mb-4 flex items-center gap-2">
							<PieChart class="text-secondary h-5 w-5" />
							<h3 class="text-base-content text-lg font-semibold">
								{$t('ai-feedback.detailed-analysis')}
							</h3>
						</div>

						<div
							class="prose prose-sm text-base-content/80 max-w-none whitespace-pre-line leading-relaxed"
						>
							{inDepthAnalysis}
						</div>

						<!-- End card with upward trend icon -->
						<div class="mt-8 flex justify-end">
							<div class="text-success inline-flex items-center gap-1.5 text-sm">
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
