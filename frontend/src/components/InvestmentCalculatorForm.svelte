<script lang="ts">
	import { t } from '$lib/i18n';
	import type { InvestmentCalculatorInput } from '$lib/types';
	import { Calculator } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher<{
		calculate: InvestmentCalculatorInput;
		reset: void;
	}>();

	// Form data - using the input object directly
	export let input: InvestmentCalculatorInput = {
		initial_investment: 0,
		monthly_contribution: 0,
		annual_return_rate: 10, // Default 10%
		investment_duration_years: 0
	};

	// Loading and error states
	export let isLoading = false;
	export let error = '';

	// Validation errors
	let initialInvestmentError = '';
	let monthlyContributionError = '';
	let annualReturnRateError = '';
	let investmentDurationYearsError = '';

	function validateForm(): boolean {
		// Reset validation errors
		initialInvestmentError = '';
		monthlyContributionError = '';
		annualReturnRateError = '';
		investmentDurationYearsError = '';

		let isValid = true;

		// Validate initial investment (must be non-negative, max 100,000)
		if (input.initial_investment < 0) {
			initialInvestmentError = $t('investment-calculator.errors.initial-investment-negative');
			isValid = false;
		} else if (input.initial_investment > 100000) {
			initialInvestmentError = $t('investment-calculator.errors.initial-investment-max');
			isValid = false;
		}

		// Validate monthly contribution (greater than 0, max 10,000)
		if (input.monthly_contribution <= 0) {
			monthlyContributionError = $t('investment-calculator.errors.monthly-contribution-positive');
			isValid = false;
		} else if (input.monthly_contribution > 10000) {
			monthlyContributionError = $t('investment-calculator.errors.monthly-contribution-max');
			isValid = false;
		}

		// Validate annual return rate (between 0 and 100 percent)
		if (input.annual_return_rate < 0 || input.annual_return_rate > 100) {
			annualReturnRateError = $t('investment-calculator.errors.annual-return-rate-range');
			isValid = false;
		}

		// Validate investment duration years (greater than 0, max 100)
		if (input.investment_duration_years <= 0) {
			investmentDurationYearsError = $t(
				'investment-calculator.errors.investment-duration-years-positive'
			);
			isValid = false;
		} else if (input.investment_duration_years > 100) {
			investmentDurationYearsError = $t(
				'investment-calculator.errors.investment-duration-years-max'
			);
			isValid = false;
		}

		return isValid;
	}

	function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		dispatch('calculate', input);
	}

	function handleReset() {
		input = {
			initial_investment: 0,
			monthly_contribution: 0,
			annual_return_rate: 10,
			investment_duration_years: 0
		};
		initialInvestmentError = '';
		monthlyContributionError = '';
		annualReturnRateError = '';
		investmentDurationYearsError = '';
		dispatch('reset');
	}
</script>

{#if error}
	<div class="alert alert-error mb-6">
		<span>{error}</span>
	</div>
{/if}

<!-- Input Form -->
<div class="card bg-base-100 mb-2 shadow-lg">
	<div class="card-body">
		<h2 class="card-title text-primary mb-6 justify-center">
			<Calculator size={24} class="text-primary" />
			{$t('investment-calculator.form.title')}
		</h2>

		<form on:submit|preventDefault={handleSubmit} class="space-y-6">
			<!-- Input Fields Grid -->
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-1">
				<!-- Initial Investment -->
				<div class="form-control">
					<label for="initial-investment" class="label">
						<span class="label-text w-full text-center font-semibold">
							{$t('investment-calculator.form.initial-investment')}
						</span>
					</label>
					<input
						id="initial-investment"
						type="number"
						step="0.01"
						min="0"
						max="100000"
						bind:value={input.initial_investment}
						class="input input-bordered {initialInvestmentError ? 'input-error' : ''}"
						placeholder="5000"
					/>
					{#if initialInvestmentError}
						<div class="label">
							<span class="label-text-alt text-error">{initialInvestmentError}</span>
						</div>
					{/if}
				</div>

				<!-- Monthly Contribution -->
				<div class="form-control">
					<label for="monthly-contribution" class="label">
						<span class="label-text w-full text-center font-semibold">
							{$t('investment-calculator.form.monthly-contribution')}
						</span>
					</label>
					<input
						id="monthly-contribution"
						type="number"
						step="0.01"
						min="0"
						max="10000"
						bind:value={input.monthly_contribution}
						class="input input-bordered {monthlyContributionError ? 'input-error' : ''}"
						placeholder="1000"
						required
					/>
					{#if monthlyContributionError}
						<div class="label">
							<span class="label-text-alt text-error">{monthlyContributionError}</span>
						</div>
					{/if}
				</div>

				<!-- Annual Return Rate -->
				<div class="form-control">
					<label for="annual-return-rate" class="label">
						<span class="label-text w-full text-center font-semibold">
							{$t('investment-calculator.form.annual-return-rate')}
						</span>
						<span class="label-text-alt text-base-content/70 w-full text-center">
							{$t('investment-calculator.form.percentage-per-year')}
						</span>
					</label>
					<input
						id="annual-return-rate"
						type="number"
						step="0.1"
						min="0"
						max="100"
						bind:value={input.annual_return_rate}
						class="input input-bordered {annualReturnRateError ? 'input-error' : ''}"
						placeholder="10"
						required
					/>
					{#if annualReturnRateError}
						<div class="label">
							<span class="label-text-alt text-error">{annualReturnRateError}</span>
						</div>
					{/if}
				</div>

				<!-- Investment Duration Years -->
				<div class="form-control">
					<label for="investment-duration-years" class="label">
						<span class="label-text w-full text-center font-semibold">
							{$t('investment-calculator.form.investment-duration-years')}
						</span>
					</label>
					<input
						id="investment-duration-years"
						type="number"
						min="1"
						max="100"
						bind:value={input.investment_duration_years}
						class="input input-bordered {investmentDurationYearsError ? 'input-error' : ''}"
						placeholder="10"
						required
					/>
					{#if investmentDurationYearsError}
						<div class="label">
							<span class="label-text-alt text-error">{investmentDurationYearsError}</span>
						</div>
					{/if}
				</div>
			</div>

			<!-- Action Buttons -->
			<div class="card-actions justify-center gap-4">
				<button type="button" class="btn btn-ghost" on:click={handleReset}>
					{$t('common.reset')}
				</button>
				<button type="submit" class="btn btn-primary" disabled={isLoading}>
					{#if isLoading}
						<span class="loading loading-spinner loading-sm"></span>
						{$t('common.calculating')}
					{:else}
                        <span class="text-base-100">
                            {$t('investment-calculator.form.calculate')}
                        </span>
					{/if}
				</button>
			</div>
		</form>
	</div>
</div>
