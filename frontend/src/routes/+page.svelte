<script lang="ts">
	import { t } from '$lib/i18n';
	import { CreditCard, BarChart3, Lock, Sparkles, Star } from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { themeService } from '$lib/services/themeService';

	let isDarkMode = false;

	function updateTheme() {
		isDarkMode = themeService.isDarkMode();
	}

	onMount(() => {
		updateTheme();
		const unsubscribe = themeService.subscribe(updateTheme);
		return unsubscribe;
	});
</script>

<div class="bg-base-100 flex min-h-screen flex-col items-center px-4 py-8">
	<!-- HERO SECTION -->
	<section
		class="flex w-full max-w-5xl flex-col-reverse items-center justify-between gap-10 py-12 lg:flex-row"
	>
		<!-- Hero Text -->
		<div class="flex flex-1 flex-col items-center text-center lg:items-start lg:text-left">
			<!-- Grão Certo Logo -->
			<img src="/favicon.svg" alt="Grão Certo" class="mb-6 h-14 w-auto" />
			<h1 class="text-primary mb-4 text-5xl font-extrabold">
				{$t('landing.hero-title')}
			</h1>
			<p class="text-base-content/80 mb-6 max-w-xl text-lg">
				{$t('landing.hero-desc')}
			</p>
			<a
				href="/register"
				class="btn btn-primary btn-lg px-8 py-3 text-lg font-bold shadow-lg transition hover:scale-105"
			>
				<span class="text-base-100">{$t('landing.cta-button')}</span>
			</a>
			<p class="text-base-content/60 mt-2 text-xs">{$t('landing.no-credit-card')}</p>
			<!-- Social proof -->
			<!-- To be completed-->
		</div>
		<!-- Hero Illustration -->
		<div class="flex flex-1 justify-center">
			<img
				src={isDarkMode ? '/graphs_dark.png' : '/graphs_light.png'}
				alt="App screenshot"
				class="w-full max-w-md rounded-xl shadow-xl"
			/>
		</div>
	</section>

	<!-- FEATURES SECTION -->
	<section class="mt-8 grid w-full max-w-5xl grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
		<div class="bg-base-200 flex flex-col items-center rounded-xl p-6 shadow">
			<CreditCard class="text-secondary mb-2 h-8 w-8" />
			<h2 class="text-primary mb-1 text-center text-xl font-bold">
				{$t('landing.accounts-title')}
			</h2>
			<p class="text-base-content/70 text-center">{$t('landing.accounts-desc')}</p>
		</div>
		<div class="bg-base-200 flex flex-col items-center rounded-xl p-6 shadow">
			<BarChart3 class="text-secondary mb-2 h-8 w-8" />
			<h2 class="text-primary mb-1 text-center text-xl font-bold">
				{$t('landing.analytics-title')}
			</h2>
			<p class="text-base-content/70 text-center">{$t('landing.analytics-desc')}</p>
		</div>
		<div class="bg-base-200 flex flex-col items-center rounded-xl p-6 shadow">
			<Lock class="text-secondary mb-2 h-8 w-8" />
			<h2 class="text-primary mb-1 text-center text-xl font-bold">{$t('landing.privacy-title')}</h2>
			<p class="text-base-content/70 text-center">{$t('landing.privacy-desc')}</p>
		</div>
		<div class="bg-base-200 flex flex-col items-center rounded-xl p-6 shadow">
			<Sparkles class="text-secondary mb-2 h-8 w-8" />
			<h2 class="text-primary mb-1 text-center text-xl font-bold">
				{$t('landing.intuitive-title')}
			</h2>
			<p class="text-base-content/70 text-center">{$t('landing.intuitive-desc')}</p>
		</div>
	</section>

	<!-- SOCIAL PROOF / TESTIMONIALS -->
	<section class="mt-16 w-full max-w-3xl">
		<h3 class="text-primary mb-6 text-center text-lg font-semibold">
			{$t('landing.testimonials-title')}
		</h3>
		<div class="grid grid-cols-1 gap-6 md:grid-cols-1">
			<div class="bg-base-200 flex flex-col items-center rounded-xl p-4 shadow">
				<div class="mb-2 flex flex-row gap-1">
					<Star fill="currentColor" class="text-warning h-5 w-5" />
					<Star fill="currentColor" class="text-warning h-5 w-5" />
					<Star fill="currentColor" class="text-warning h-5 w-5" />
					<Star fill="currentColor" class="text-warning h-5 w-5" />
					<Star fill="currentColor" class="text-warning h-5 w-5" />
				</div>
				<p class="text-base-content/80 mb-2 text-center">{$t('landing.testimonial-1')}</p>
				<span class="text-base-content/60 text-xs">— {$t('landing.testimonial-1-author')}</span>
			</div>
		</div>
	</section>

	<!-- SECONDARY CTA -->
	<section class="mt-16 flex w-full max-w-2xl flex-col items-center">
		<h4 class="text-primary mb-2 text-2xl font-bold">{$t('landing.cta-title')}</h4>
		<p class="text-base-content/80 mb-4 text-center">{$t('landing.cta-desc')}</p>
		<a href="/register" class="btn btn-primary btn-lg w-full max-w-xs text-lg font-bold">
			<span class="text-base-100">{$t('landing.cta-button')}</span>
		</a>
		<p class="text-base-content/60 mt-2 text-xs">{$t('landing.no-credit-card')}</p>
	</section>
</div>
