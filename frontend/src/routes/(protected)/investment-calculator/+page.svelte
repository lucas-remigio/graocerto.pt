<script lang="ts">
	import api_axios from '$lib/axios';
	import { t } from '$lib/i18n';
	import InvestmentCalculatorHeader from '$components/InvestmentCalculatorHeader.svelte';
	import InvestmentCalculatorForm from '$components/InvestmentCalculatorForm.svelte';
	import InvestmentCalculatorResults from '$components/InvestmentCalculatorResults.svelte';
	import type { InvestmentCalculatorInput, InvestmentCalculatorResponse } from '$lib/types';

	// Form data - using input object
	let input: InvestmentCalculatorInput = {
		initial_investment: 0,
		monthly_contribution: 0,
		annual_return_rate: 10, // Default 10%
		investment_duration_years: 0
	};

	// Loading and error states
	let isLoading = false;
	let error = '';

	// Results
	let results: InvestmentCalculatorResponse | null = null;

	async function handleCalculate(event: CustomEvent<InvestmentCalculatorInput>) {
		const inputData = event.detail;

		isLoading = true;
		error = '';
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
		} catch (err: any) {
			console.error('Error calculating investment:', err);
			error = err.response?.data?.error || $t('investment-calculator.errors.calculation-failed');
		} finally {
			isLoading = false;
		}
	}

	function handleReset() {
		results = null;
		error = '';
	}
</script>

<div class="container mx-auto max-w-4xl p-4">
	<InvestmentCalculatorHeader />

	<InvestmentCalculatorForm
		bind:input
		{isLoading}
		{error}
		on:calculate={handleCalculate}
		on:reset={handleReset}
	/>

	{#if results}
		<InvestmentCalculatorResults {results} />
	{/if}
</div>
