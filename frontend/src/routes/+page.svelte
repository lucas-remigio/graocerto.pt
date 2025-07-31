<script lang="ts">
	import { t } from '$lib/i18n';
	import {
		CreditCard,
		BarChart3,
		Lock,
		Sparkles,
		Star,
		Mail,
		Linkedin,
		Github,
		CornerUpLeft,
		ArrowUpCircle,
		Heart
	} from 'lucide-svelte';
	import { onMount } from 'svelte';
	import { themeService } from '$lib/services/themeService';

	let zoomedImg: string | null = null;
	function openZoom(src: string) {
		zoomedImg = src;
	}
	function closeZoom() {
		zoomedImg = null;
	}

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
	<div class="hero bg-base-100 mb-12 min-h-[60vh] rounded-xl px-4 shadow md:px-8">
		<div class="hero-content mx-auto w-full max-w-6xl flex-col gap-12 lg:flex-row-reverse">
			<div class="relative flex flex-1 justify-end">
				<button
					type="button"
					class="transition-transform duration-200 hover:scale-105"
					on:click={() => openZoom(isDarkMode ? '/graphs_dark.png' : '/graphs_light.png')}
					aria-label="Open app screenshot"
				>
					<img
						src={isDarkMode ? '/graphs_dark.png' : '/graphs_light.png'}
						alt="App screenshot"
						class="max-w-md cursor-pointer rounded-xl shadow-xl"
						draggable="false"
					/>
					<!-- Click Me indicator with arrow -->
					<div
						class="pointer-events-none absolute -bottom-8 right-4 z-10 flex animate-bounce flex-col items-center"
					>
						<CornerUpLeft class="text-primary h-6 w-6 animate-bounce" aria-hidden="true" />
						<span
							class="bg-primary text-base-100 mb-1 rounded-full px-3 py-1 text-sm font-semibold shadow-lg"
							>Click Me!</span
						>
					</div>
				</button>
			</div>
			<div class="flex flex-1 flex-col items-center text-center lg:items-start lg:text-left">
				<img src="/favicon.svg" alt="Grão Certo" class="mb-6 h-14 w-auto" />
				<h1 class="hero-title text-primary mb-4 text-5xl font-extrabold">
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
			</div>
		</div>
	</div>

	<!-- FEATURES SECTION -->
	<div class="mb-16 grid w-full max-w-5xl grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
		<div class="card bg-base-200 flex flex-col items-center p-6 shadow">
			<CreditCard class="text-secondary mb-2 h-8 w-8" />
			<h2 class="card-title text-primary mb-1 text-center text-xl font-bold">
				{$t('landing.accounts-title')}
			</h2>
			<p class="text-base-content/70 text-center">{$t('landing.accounts-desc')}</p>
		</div>
		<div class="card bg-base-200 flex flex-col items-center p-6 shadow">
			<BarChart3 class="text-secondary mb-2 h-8 w-8" />
			<h2 class="card-title text-primary mb-1 text-center text-xl font-bold">
				{$t('landing.analytics-title')}
			</h2>
			<p class="text-base-content/70 text-center">{$t('landing.analytics-desc')}</p>
		</div>
		<div class="card bg-base-200 flex flex-col items-center p-6 shadow">
			<Lock class="text-secondary mb-2 h-8 w-8" />
			<h2 class="card-title text-primary mb-1 text-center text-xl font-bold">
				{$t('landing.privacy-title')}
			</h2>
			<p class="text-base-content/70 text-center">{$t('landing.privacy-desc')}</p>
		</div>
		<div class="card bg-base-200 flex flex-col items-center p-6 shadow">
			<Sparkles class="text-secondary mb-2 h-8 w-8" />
			<h2 class="card-title text-primary mb-1 text-center text-xl font-bold">
				{$t('landing.intuitive-title')}
			</h2>
			<p class="text-base-content/70 text-center">{$t('landing.intuitive-desc')}</p>
		</div>
	</div>

	<!-- TESTIMONIALS SECTION -->
	<div class="mb-16 w-full max-w-3xl">
		<h3 class="text-primary mb-6 text-center text-lg font-semibold">
			{$t('landing.testimonials-title')}
		</h3>
		<div class="card bg-base-200 flex flex-col items-center p-4 shadow">
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

	<!-- SECONDARY CTA -->
	<div class="card bg-base-200 mb-12 flex w-full max-w-2xl flex-col items-center p-8 shadow">
		<h4 class="text-primary mb-2 text-2xl font-bold">{$t('landing.cta-title')}</h4>
		<p class="text-base-content/80 mb-4 text-center">{$t('landing.cta-desc')}</p>
		<a
			href="/register"
			class="btn btn-primary btn-lg w-full max-w-xs text-lg font-bold transition hover:scale-105"
		>
			<span class="text-base-100">{$t('landing.cta-button')}</span>
		</a>
		<p class="text-base-content/60 mt-2 text-xs">{$t('landing.no-credit-card')}</p>
	</div>

	<!-- ABOUT ME SECTION -->
	<div
		class="card bg-base-200 mb-12 flex w-full max-w-3xl flex-col items-center p-8 text-center opacity-80 shadow"
	>
		<button
			type="button"
			class="avatar mb-3 h-32 w-32 rounded-full transition-transform duration-200 hover:scale-105"
			aria-label="Zoom image of Lucas Remigio"
			on:click={() => openZoom('/the_dev.jpeg')}
		>
			<img
				src="/the_dev.jpeg"
				alt="Lucas Remigio"
				class="h-full w-full rounded-full object-cover"
				draggable="false"
			/>
		</button>
		<h5 class="text-primary mb-1 text-lg font-semibold">{$t('landing.dev-about')}</h5>
		<p class="text-base-content/70 mb-2">
			{$t('landing.dev-presentation')}
		</p>
		<h6 class="text-primary mb-2 mt-4 text-base font-semibold">Contact me!</h6>
		<div class="mt-2 flex justify-center gap-4">
			<a
				href="mailto:remigio@graocerto.pt"
				class="btn btn-ghost btn-circle transition-transform duration-200 hover:scale-105"
				aria-label="Email"
			>
				<Mail class="text-primary h-6 w-6" />
			</a>
			<a
				href="https://linkedin.com/in/lucas-remigio"
				target="_blank"
				rel="noopener"
				class="btn btn-ghost btn-circle transition-transform duration-200 hover:scale-105"
				aria-label="LinkedIn"
			>
				<Linkedin class="text-primary h-6 w-6" />
			</a>
			<a
				href="https://github.com/lucas-remigio"
				target="_blank"
				rel="noopener"
				class="btn btn-ghost btn-circle transition-transform duration-200 hover:scale-105"
				aria-label="GitHub"
			>
				<Github class="text-primary h-6 w-6" />
			</a>
		</div>
	</div>

	<!-- FOOTER -->
	<footer
		class="footer footer-center bg-base-200 text-base-content mt-auto w-full max-w-5xl rounded-xl p-6 shadow"
	>
		<small>
			Grão Certo &mdash; {$t('landing.made-with')}

			<span class="text-error"><Heart fill="currentColor" /></span>
			{$t('common.by')} Lucas Remígio
		</small>
	</footer>
</div>

{#if zoomedImg}
	<button
		type="button"
		class="fixed inset-0 z-[100] flex items-center justify-center bg-black/70 backdrop-blur-sm"
		aria-label="Close zoomed image"
		on:click={closeZoom}
		on:keydown={(e) => {
			if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') closeZoom();
		}}
	>
		<img
			src={zoomedImg}
			alt="Zoomed"
			class="animate-zoom-in max-h-[80vh] max-w-[90vw] rounded-xl shadow-lg"
			draggable="false"
		/>
	</button>
{/if}

<style>
	@keyframes zoom-in {
		from {
			transform: scale(0.7);
			opacity: 0;
		}
		to {
			transform: scale(1);
			opacity: 1;
		}
	}
	.animate-zoom-in {
		animation: zoom-in 0.25s cubic-bezier(0.4, 2, 0.3, 1) both;
	}
</style>
