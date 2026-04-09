<script lang="ts">
	import api_axios from '$lib/axios';
	import { t } from '$lib/i18n';
	import InvestmentCalculatorHeader from '$components/InvestmentCalculatorHeader.svelte';
	import InvestmentCalculatorForm from '$components/InvestmentCalculatorForm.svelte';
	import InvestmentCalculatorResults from '$components/InvestmentCalculatorResults.svelte';
	import type { InvestmentCalculatorInput, InvestmentCalculatorResponse } from '$lib/types';

	import { toastStore } from '$lib/stores/toast';

	let loading = false;

	// Results
	let results: InvestmentCalculatorResponse | null = null;
	let isLoading = false;

	async function handleCalculate(event: CustomEvent<InvestmentCalculatorInput>) {
		const inputData = event.detail;

		isLoading = true;
		results = null;

		try {
			const payload = {
				initial_investment: inputData.initial_investment,
				monthly_contribution: inputData.monthly_contribution,
				annual_return_rate: inputData.annual_return_rate / 100, // Convert percentage to decimal
				investment_duration_years: inputData.investment_duration_years
			};

			const response = await api_axios.post('investment-calculator', payload);

			if (response.status !== 200) {
				throw new Error(`${$t('errors.server-error')}: ${response.status}`);
			}

			results = response.data;
		} catch (err: unknown) {
			console.error('Error calculating investment:', err);
			if (
				typeof err === 'object' &&
				err !== null &&
				'response' in err &&
				typeof (err as { response?: { data?: { error?: string } } }).response?.data?.error ===
					'string'
			) {
				toastStore.error(
					(err as { response?: { data?: { error?: string } } }).response!.data!.error!
				);
			} else {
				toastStore.error($t('investment-calculator.errors.calculation-failed'));
			}
		} finally {
			isLoading = false;
		}
	}

	function handleReset() {
		results = null;
	}
</script>

<div class="container mx-auto p-4">
	<InvestmentCalculatorHeader />

	<!-- Responsive Layout: Vertical on small/medium, horizontal on large -->
	<div class="flex flex-col lg:flex-row lg:items-start lg:gap-6">
		<!-- Left Column: Form (full width on small/medium, fixed width on large) -->
		<div class="w-full lg:w-96 lg:flex-shrink-0">
			<InvestmentCalculatorForm {isLoading} on:calculate={handleCalculate} on:reset={handleReset} />
		</div>

		<!-- Right Column: Results (full width on small/medium, remaining space on large) -->
		<div class="mt-6 w-full overflow-hidden lg:mt-0 lg:min-w-0 lg:flex-1">
			<InvestmentCalculatorResults {results} />
		</div>
	</div>
</div>
