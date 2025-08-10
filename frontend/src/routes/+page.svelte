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
		BookOpen
	} from 'lucide-svelte';
	import { t } from '$lib/i18n';

	let zoomedImg: string | null = null;

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

	function openImg(src: string) {
		zoomedImg = src;
	}
	function closeImg() {
		zoomedImg = null;
	}
	function scrollToId(id: string) {
		document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}
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

					<!-- Highlighted financial literacy callout -->
					<div class="border-primary/20 bg-primary/5 mt-5 rounded-xl border p-4 md:p-5">
						<div class="flex items-start gap-3">
							<BookOpen size={20} class="text-primary mt-0.5 flex-shrink-0" />
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
						<div class="bg-base-100 flex items-center justify-center p-4">
							<img
								src="graphs_dark.png"
								alt={$t('landing.images.hero-alt', { default: 'Wallet Tracker preview' })}
								class="rounded-lg shadow-2xl"
								loading="eager"
								decoding="async"
							/>
						</div>
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
				{$t('landing.features.title', { default: 'Why you’ll love it' })}
			</h2>
			<p class="text-base-content/70 mx-auto mt-2 max-w-2xl text-center">
				{$t('landing.features.subtitle', {
					default: 'Designed for clarity, speed, and real insight into your spending and income.'
				})}
			</p>

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
</div>

{#if zoomedImg}
	<!-- Image Modal -->
	<div class="fixed inset-0 z-50 grid place-items-center bg-black/60 p-4" on:click={closeImg}>
		<div class="max-h-[90vh] max-w-5xl">
			<img
				src={zoomedImg}
				alt={$t('landing.gallery.zoom-alt', { default: 'Screenshot' })}
				class="max-h-[90vh] w-full rounded-xl shadow-2xl"
			/>
		</div>
	</div>
{/if}

<style>
	:global(.mockup-window) {
		overflow: hidden;
	}
</style>
