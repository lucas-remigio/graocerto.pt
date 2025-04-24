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
	import { fade, fly } from 'svelte/transition';

	export let account: Account;
	export let closeModal: () => void;

	let error: string = '';
	let isLoading: boolean = true;
	let feedbackMessage: string = '';
	let inDepthAnalysis: string = '';

	async function getTransactionsAiFeedback() {
		isLoading = true;
		try {
			const res = await api_axios('accounts/' + account.token + '/feedback-month');

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

	function handleCloseModal() {
		closeModal();
	}

	onMount(() => {
		getTransactionsAiFeedback();
	});
</script>

<div class="modal modal-open">
	<div
		class="modal-box relative max-w-3xl overflow-hidden p-0"
		transition:fly={{ y: 20, duration: 300 }}
	>
		<!-- Gradient header -->
		<div class="bg-gradient-to-r from-purple-600 to-blue-500 px-6 py-5 text-white shadow-lg">
			<div class="flex items-center gap-2">
				<BarChart class="h-5 w-5" />
				<h3 class="text-xl font-bold">Financial Insights: {account.account_name}</h3>
			</div>
			<!-- Close button -->
			<button
				class="btn btn-sm btn-circle absolute right-2 top-2 border-none bg-white/20 hover:bg-white/30"
				on:click={handleCloseModal}
			>
				<X class="text-white" />
			</button>
			<div class="mt-1 flex items-center gap-2 text-sm text-white/80">
				<CalendarClock class="h-4 w-4" />
				<span
					>Monthly analysis for {new Date().toLocaleString('default', {
						month: 'long',
						year: 'numeric'
					})}</span
				>
			</div>
		</div>

		<!-- Message body -->
		<div class="p-6">
			{#if isLoading}
				<div class="flex flex-col items-center justify-center py-12" in:fade>
					<Loader2 class="mb-4 h-12 w-12 animate-spin text-blue-500" />
					<p class="text-gray-600 dark:text-gray-400">Analyzing your financial data...</p>
				</div>
			{:else if error}
				<div class="rounded border-l-4 border-red-500 bg-red-50 p-4 dark:bg-red-900/20" in:fade>
					<p class="flex items-center gap-2 text-red-500">
						<AlertCircle class="h-5 w-5 text-red-500" />
						{error}
					</p>
				</div>
			{:else}
				<div in:fade={{ duration: 300, delay: 100 }}>
					<!-- Summary section -->
					<div
						class="mb-6 rounded-lg border border-blue-100 bg-gradient-to-r from-blue-50 to-purple-50 p-5 shadow-sm dark:border-blue-900/20 dark:from-blue-900/10 dark:to-purple-900/10"
					>
						<div class="flex items-start gap-3">
							<div class="rounded-full bg-blue-100 p-2 dark:bg-blue-800">
								<Lightbulb class="h-5 w-5 text-blue-600 dark:text-blue-200" />
							</div>
							<div>
								<h3 class="mb-2 text-lg font-medium text-blue-800 dark:text-blue-200">
									Key Insights
								</h3>
								<p class="text-gray-700 dark:text-gray-300">{feedbackMessage}</p>
							</div>
						</div>
					</div>

					<!-- Detailed analysis section -->
					<div class="mt-6">
						<div class="mb-4 flex items-center gap-2">
							<PieChart class="h-5 w-5 text-purple-600 dark:text-purple-400" />
							<h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200">
								Detailed Analysis
							</h3>
						</div>

						<div
							class="prose prose-sm max-w-none whitespace-pre-line leading-relaxed text-gray-700 dark:text-gray-300"
						>
							{inDepthAnalysis}
						</div>

						<!-- End card with upward trend icon -->
						<div class="mt-8 flex justify-end">
							<div
								class="inline-flex items-center gap-1.5 text-sm text-green-600 dark:text-green-400"
							>
								<ArrowUpRight class="h-4 w-4" />
								<span>Based on your transaction history</span>
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
