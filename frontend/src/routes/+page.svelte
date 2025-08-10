<script lang="ts">
	import {
		Wallet,
		BarChart3,
		Lock,
		Shield,
		Sparkles,
		Calendar,
		PieChart,
		Cloud,
		Zap,
		Github,
		MonitorSmartphone,
		BookOpen,
		Quote
	} from 'lucide-svelte';
	import { t } from '$lib/i18n';
	import { themeService } from '$lib/services/themeService';
	import { onMount } from 'svelte';

	let zoomedImg: string | null = null;
	let isDarkMode = false;

	// Use translation keys + defaults; render text via $t in the markup
	const features = [
		{
			icon: Wallet,
			titleKey: 'landing.features.track.title',
			defaultTitle: 'Track Transactions',
			descKey: 'landing.features.track.desc',
			defaultDesc: 'Log credits and debits effortlessly with clean categorization.'
		},
		{
			icon: BarChart3,
			titleKey: 'landing.features.insights.title',
			defaultTitle: 'Clear Insights',
			descKey: 'landing.features.insights.desc',
			defaultDesc: 'Understand your finances with statistics, charts, and heatmaps.'
		},
		{
			icon: Calendar,
			titleKey: 'landing.features.monthly.title',
			defaultTitle: 'Monthly View',
			descKey: 'landing.features.monthly.desc',
			defaultDesc: 'Navigate history by month and year with quick filters.'
		},
		{
			icon: PieChart,
			titleKey: 'landing.features.categories.title',
			defaultTitle: 'Category Breakdown',
			descKey: 'landing.features.categories.desc',
			defaultDesc: 'Spot trends by category to optimize your spending.'
		},
		{
			icon: MonitorSmartphone,
			titleKey: 'landing.features.responsive.title',
			defaultTitle: 'Responsive Design',
			descKey: 'landing.features.responsive.desc',
			defaultDesc: 'Optimized for laptops and phones — use it anywhere.'
		},
		{
			icon: Shield,
			titleKey: 'landing.features.security.title',
			defaultTitle: 'Secure by Design',
			descKey: 'landing.features.security.desc',
			defaultDesc: 'HTTPS, DB SSL verification, and best-practice security.'
		}
	];

	const testimonials = [
		{
			quoteKey: 'landing.testimonials.1.quote',
			defaultQuote:
				'Understanding my spending has given me space to make better choices. What to save, what to spend, what to invest. A simple, complete tool that brings lightness to everyday life.',
			authorKey: 'landing.testimonials.1.author',
			defaultAuthor: 'Lara G.'
		}
	];

	const faqs = [
		{
			qKey: 'landing.faq.1.q',
			defaultQ: 'Is my data private?',
			aKey: 'landing.faq.1.a',
			defaultA: 'Yes. We use HTTPS and database SSL verification. No ads, no data selling.'
		},
		{
			qKey: 'landing.faq.2.q',
			defaultQ: 'Is it free and open source?',
			aKey: 'landing.faq.2.a',
			defaultA: 'Yes. The project is open source on GitHub and free to use.'
		},
		{
			qKey: 'landing.faq.3.q',
			defaultQ: 'How do I start?',
			aKey: 'landing.faq.3.a',
			defaultA:
				'Create an account, set a monthly budget, and import or add your first transactions.'
		}
	];

	const highlights = [
		{ labelKey: 'landing.highlights.fast', defaultLabel: 'Fast', icon: Zap },
		{ labelKey: 'landing.highlights.privacy', defaultLabel: 'Privacy-first', icon: Lock },
		{ labelKey: 'landing.highlights.opensource', defaultLabel: 'Open source', icon: Github }
	];

	const screenshots = [
		{
			src: '/graphs_light.png',
			altKey: 'landing.screenshots.dashboard',
			defaultAlt: 'Dashboard overview'
		},
		{
			src: '/transactions_light.png',
			altKey: 'landing.screenshots.transactions',
			defaultAlt: 'Transactions list'
		}
	];

	function updateTheme() {
		isDarkMode = themeService.isDarkMode();
		console.log('Theme updated:', isDarkMode ? 'Dark' : 'Light');
	}

	// Pick themed asset: supports `{theme}` placeholder or `_light/_dark` suffix
    function themedSrc(src: string, dark: boolean) {
        const theme = dark ? 'dark' : 'light';
        if (src.includes('{theme}')) return src.replace('{theme}', theme);
        return src.replace(/_(light|dark)(\.\w+)$/, `_${theme}$2`);
    }

    // Reactive hero image based on theme (explicit dependency on isDarkMode)
    let baseHeroImg = '/graphs_{theme}.png';
    $: heroImg = themedSrc(baseHeroImg, isDarkMode);

	function openImg(src: string) {
		zoomedImg = src;
	}
	function closeImg() {
		zoomedImg = null;
	}
	function scrollToId(id: string) {
		document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	onMount(() => {
		updateTheme();
		const unsubscribe = themeService.subscribe(updateTheme);
		return unsubscribe;
	});
</script>

<div class="bg-base-100 flex min-h-screen flex-col">
	<!-- Hero -->
	<section class="relative">
		<div
			class="from-primary/5 pointer-events-none absolute inset-0 bg-gradient-to-b via-transparent to-transparent"
		></div>
		<div class="container mx-auto max-w-6xl px-4 pb-12 pt-16 md:pt-24">
			<div class="grid items-center gap-10 md:grid-cols-2">
				<div>
					<div class="badge badge-primary text-base-100 mb-4">Grão Certo</div>
					<h1 class="text-3xl font-extrabold tracking-tight md:text-5xl">
						{$t('landing.hero.title', { default: 'Take control of your finances with clarity' })}
					</h1>
					<p class="text-base-content/70 mt-4 max-w-prose text-base md:text-lg">
						{$t('landing.hero.subtitle', {
							default:
								'Simple transaction tracking, insightful statistics, and a clean heatmap to visualize your daily activity.'
						})}
					</p>
					<div class="mt-6 flex flex-wrap items-center gap-3">
						<a
							href="/login"
							class="btn btn-primary shadow-lg"
							aria-label={$t('landing.cta.get-started', { default: 'Get Started' })}
						>
							<Sparkles size={18} class="text-base-100" />
							<span class="text-base-100"
								>{$t('landing.cta.get-started', { default: 'Get Started' })}</span
							>
						</a>
						<button
							class="btn btn-ghost shadow-lg"
							on:click={() => scrollToId('features')}
							aria-label={$t('landing.cta.see-features', { default: 'See Features' })}
						>
							<BarChart3 size={18} />
							<span>{$t('landing.cta.see-features', { default: 'See Features' })}</span>
						</button>
						<a
							class="btn btn-outline shadow-lg"
							href="https://github.com/lucas-remigio/graocerto.pt"
							target="_blank"
							rel="noreferrer"
							aria-label={$t('landing.cta.view-github', { default: 'View on GitHub' })}
						>
							<Github size={18} />
							<span>{$t('landing.cta.view-github', { default: 'View on GitHub' })}</span>
						</a>
					</div>

					<div class="mt-6 flex flex-wrap gap-2">
						{#each highlights as h}
							<span class="badge badge-ghost gap-1">
								<svelte:component this={h.icon} size={14} />
								{$t(h.labelKey, { default: h.defaultLabel })}
							</span>
						{/each}
					</div>
				</div>

				<!-- Hero mockup -->
				<div class="hidden md:block">
					<div class="mockup-window border-base-300 bg-base-200 border shadow-xl">
						<button
							type="button"
							class="bg-base-100 flex h-full w-full items-center justify-center p-4"
							on:click={() => openImg('graphs_dark.png')}
							on:keydown={(e) => {
								if (e.key === 'Enter' || e.key === ' ') openImg('graphs_dark.png');
							}}
							aria-label={$t('landing.actions.open-screenshot', { default: 'Open screenshot' })}
							tabindex="0"
							style="border: none; background: none; padding: 0;"
						>
							<img
								src={heroImg}
								alt={$t('landing.images.hero-alt', { default: 'Wallet Tracker preview' })}
								class="rounded-lg shadow-2xl"
								loading="eager"
								decoding="async"
							/>
						</button>
					</div>
					<p class="text-base-content/50 mt-3 text-center text-xs">
						{$t('landing.images.hero-caption', { default: 'Preview of dashboard and statistics' })}
					</p>
				</div>
			</div>
		</div>
	</section>

	<!-- Features -->
	<section id="features" class="bg-base-100">
		<div class="container mx-auto max-w-6xl px-4 py-12 md:py-16">
			<h2 class="text-center text-2xl font-bold md:text-3xl">
				{$t('landing.features.title', { default: "Why you'll love it" })}
			</h2>
			<p class="text-base-content/70 mx-auto mt-2 max-w-2xl text-center">
				{$t('landing.features.subtitle', {
					default: 'Designed for clarity, speed, and real insight into your spending and income.'
				})}
			</p>

			<!-- Integrated financial literacy callout -->
			<div
				class="border-primary/20 from-primary/10 via-primary/5 mt-6 rounded-xl border bg-gradient-to-r to-transparent p-4 md:p-5"
			>
				<div class="flex flex-col gap-4 sm:flex-row sm:items-center">
					<div class="flex flex-1 items-start gap-3">
						<div class="text-primary bg-primary/10 rounded-lg p-2">
							<BookOpen size={20} />
						</div>
						<div>
							<h3 class="font-semibold">
								{$t('landing.finlit.callout.title', {
									default: 'Start with a budget. Understand your spending.'
								})}
							</h3>
							<p class="text-base-content/70 text-sm">
								{$t('landing.finlit.callout.desc', {
									default:
										'Rooted in Portuguese financial literacy: begin each month with a plan and track where and how your money is spent.'
								})}
							</p>
						</div>
					</div>
				</div>
			</div>

			<div class="mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
				{#each features as f}
					<div
						class="card bg-base-100 border-base-200 border shadow-sm transition-shadow hover:shadow-md"
					>
						<div class="card-body">
							<div class="flex items-center gap-3">
								<div class="text-primary">
									<svelte:component this={f.icon} size={22} />
								</div>
								<h3 class="card-title text-lg">{$t(f.titleKey, { default: f.defaultTitle })}</h3>
							</div>
							<p class="text-base-content/70">{$t(f.descKey, { default: f.defaultDesc })}</p>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</section>

	<!-- Screenshots -->
	<section class="bg-base-200/40">
		<div class="container mx-auto max-w-6xl px-4 py-12 md:py-16">
			<h2 class="text-center text-2xl font-bold md:text-3xl">
				{$t('landing.gallery.title', { default: 'A look inside' })}
			</h2>
			<p class="text-base-content/70 mx-auto mt-2 max-w-2xl text-center">
				{$t('landing.gallery.subtitle', {
					default: 'Clean tables, intuitive statistics, and a daily activity heatmap.'
				})}
			</p>

			<div class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-2">
				{#each screenshots as s}
					<button
						class="border-base-300 bg-base-100 group rounded-xl border p-2 shadow transition hover:shadow-lg"
						on:click={() => openImg(s.src)}
						aria-label={`${$t('landing.actions.open-screenshot', { default: 'Open screenshot' })}: ${$t(s.altKey, { default: s.defaultAlt })}`}
					>
						<img
							src={s.src}
							alt={$t(s.altKey, { default: s.defaultAlt })}
							class="aspect-video w-full rounded-lg object-cover transition group-hover:brightness-105"
							loading="lazy"
							decoding="async"
						/>
					</button>
				{/each}
			</div>

			<div class="mt-8 flex justify-center">
				<a
					href="/login"
					class="btn btn-primary shadow-lg"
					aria-label={$t('landing.cta.start-tracking', { default: 'Start Tracking' })}
				>
					<Wallet size={18} class="text-base-100" />
					<span class="text-base-100"
						>{$t('landing.cta.start-tracking', { default: 'Start Tracking' })}</span
					>
				</a>
			</div>
		</div>
	</section>

	<!-- Testimonials -->
	<section class="bg-base-100">
		<div class="container mx-auto max-w-6xl px-4 py-12 md:py-16">
			<h2 class="text-center text-2xl font-bold md:text-3xl">
				{$t('landing.testimonials.title', { default: 'What people say' })}
			</h2>
			<p class="text-base-content/70 mx-auto mt-2 max-w-2xl text-center">
				{$t('landing.testimonials.subtitle', {
					default: 'Real stories from users improving their finances.'
				})}
			</p>

			<div class="mt-8 grid justify-items-center gap-6">
				{#each testimonials as tItem}
					<figure class="card bg-base-100 border-base-200 w-full max-w-2xl border shadow-sm">
						<div class="card-body">
							<Quote size={18} class="text-primary opacity-80" />
							<blockquote class="mt-2 text-base">
								“{$t(tItem.quoteKey, { default: tItem.defaultQuote })}”
							</blockquote>
						</div>
						<figcaption class="text-base-content/60 px-6 pb-5 pt-0 text-sm">
							— {$t(tItem.authorKey, { default: tItem.defaultAuthor })}
						</figcaption>
					</figure>
				{/each}
			</div>
		</div>
	</section>

	<!-- FAQ -->
	<section class="bg-base-200/40">
		<div class="container mx-auto max-w-6xl px-4 py-12 md:py-16">
			<h2 class="text-center text-2xl font-bold md:text-3xl">
				{$t('landing.faq.title', { default: 'Frequently asked questions' })}
			</h2>
			<div class="mx-auto mt-6 max-w-3xl space-y-3">
				{#each faqs as f}
					<details class="border-base-300 bg-base-100 rounded-lg border p-4">
						<summary class="cursor-pointer list-none font-medium">
							{$t(f.qKey, { default: f.defaultQ })}
						</summary>
						<p class="text-base-content/70 mt-2 text-sm">
							{$t(f.aKey, { default: f.defaultA })}
						</p>
					</details>
				{/each}
			</div>
		</div>
	</section>
	<!-- Final CTA -->
	<section class="bg-base-200/40">
		<div class="container mx-auto max-w-6xl px-4 py-10 text-center">
			<a
				href="/login"
				class="btn btn-primary shadow-lg"
				aria-label={$t('landing.cta.get-started', { default: 'Get Started' })}
			>
				<Sparkles size={18} class="text-base-100" />
				<span class="text-base-100"
					>{$t('landing.cta.get-started', { default: 'Get Started' })}</span
				>
			</a>
		</div>
	</section>
</div>

{#if zoomedImg}
	<div class="fixed inset-0 z-50 grid place-items-center bg-black/60 p-4">
		<!-- Backdrop as a real button for a11y -->
		<button
			type="button"
			class="absolute inset-0 h-full w-full cursor-zoom-out"
			aria-label={$t('landing.actions.close-modal', { default: 'Close image modal' })}
			on:click={closeImg}
		></button>
		<!-- Dialog content -->
		<div role="dialog" aria-modal="true" class="relative z-10 max-h-[90vh] max-w-5xl outline-none">
			<img
				src={zoomedImg}
				alt={$t('landing.gallery.zoom-alt', { default: 'Screenshot' })}
				class="max-h-[90vh] w-full rounded-xl shadow-2xl"
			/>
		</div>
	</div>
{/if}

<!-- Close on Escape globally -->
<svelte:window
	on:keydown={(e) => {
		if (e.key === 'Escape' && zoomedImg) closeImg();
	}}
/>

<style>
	:global(.mockup-window) {
		overflow: hidden;
	}
</style>
