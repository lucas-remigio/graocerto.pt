<script lang="ts">
	import { t } from '$lib/i18n';
</script>

<div
	class="bg-base-100 flex flex-col items-center justify-center px-4 py-12 lg:h-[calc(100vh-64px)]"
>
	<div
		class="flex w-full max-w-6xl flex-col items-center gap-8 lg:flex-row lg:items-start lg:justify-between"
	>
		<div class="flex flex-col items-center lg:w-1/2 lg:items-start">
			<img src="/logo.png" alt="Grão Certo Logo" class="mb-6 h-20 w-auto" />
			<h1 class="text-primary mb-4 text-center text-5xl font-extrabold lg:text-left">Grão Certo</h1>
			<p class="text-base-content/80 text-center text-lg lg:text-left">
				<strong>Grão Certo</strong>
				{$t('landing.intro')}
			</p>

			<!-- CTA only on large screens -->
			<div class="hidden lg:block">
				<div class="mb-10 mt-10 text-left">
					<p class="text-base-content/90 mb-2 text-lg font-semibold">
						{$t('landing.cta-title')}
					</p>
					<p class="text-base-content/70">{$t('landing.cta-desc')}</p>
				</div>
				<a href="/login" class="btn btn-primary btn-lg max-w-xs text-lg font-bold">
					<span class="text-base-100">{$t('landing.cta-button')}</span>
				</a>
			</div>
		</div>

		<div class="grid w-full grid-cols-1 gap-6 md:grid-cols-2 lg:w-1/2">
			<!-- Card 1: Accounts -->
			<div class="bg-base-200 rounded-xl shadow">
				<!-- Desktop: Always show full content -->
				<div class="hidden flex-col items-center p-6 lg:flex">
					<h2 class="text-primary mb-2 text-center text-xl font-bold">
						{$t('landing.accounts-title')}
					</h2>
					<p class="text-base-content/70 text-center">{$t('landing.accounts-desc')}</p>
				</div>

				<!-- Mobile: DaisyUI Accordion -->
				<div tabindex="-1" class="collapse-arrow bg-base-200 collapse lg:hidden">
					<div class="collapse-title text-primary font-bold">
						{$t('landing.accounts-title')}
					</div>
					<div class="collapse-content">
						<p class="text-base-content/70">{$t('landing.accounts-desc')}</p>
					</div>
				</div>
			</div>

			<!-- Card 2: Analytics -->
			<div class="bg-base-200 rounded-xl shadow">
				<!-- Desktop: Always show full content -->
				<div class="hidden flex-col items-center p-6 lg:flex">
					<h2 class="text-primary mb-2 text-center text-xl font-bold">
						{$t('landing.analytics-title')}
					</h2>
					<p class="text-base-content/70 text-center">{$t('landing.analytics-desc')}</p>
				</div>

				<!-- Mobile: DaisyUI Accordion -->
				<div tabindex="-2" class="collapse-arrow bg-base-200 collapse lg:hidden">
					<div class="collapse-title text-primary font-bold">
						{$t('landing.analytics-title')}
					</div>
					<div class="collapse-content">
						<p class="text-base-content/70">{$t('landing.analytics-desc')}</p>
					</div>
				</div>
			</div>

			<!-- Card 3: Privacy -->
			<div class="bg-base-200 rounded-xl shadow">
				<!-- Desktop: Always show full content -->
				<div class="hidden flex-col items-center p-6 lg:flex">
					<h2 class="text-primary mb-2 text-center text-xl font-bold">
						{$t('landing.privacy-title')}
					</h2>
					<p class="text-base-content/70 text-center">{$t('landing.privacy-desc')}</p>
				</div>

				<!-- Mobile: DaisyUI Accordion -->
				<div tabindex="-3" class="collapse-arrow bg-base-200 collapse lg:hidden">
					<div class="collapse-title text-primary font-bold">
						{$t('landing.privacy-title')}
					</div>
					<div class="collapse-content">
						<p class="text-base-content/70">{$t('landing.privacy-desc')}</p>
					</div>
				</div>
			</div>

			<!-- Card 4: Intuitive -->
			<div class="bg-base-200 rounded-xl shadow">
				<!-- Desktop: Always show full content -->
				<div class="hidden flex-col items-center p-6 lg:flex">
					<h2 class="text-primary mb-2 text-center text-xl font-bold">
						{$t('landing.intuitive-title')}
					</h2>
					<p class="text-base-content/70 text-center">{$t('landing.intuitive-desc')}</p>
				</div>

				<!-- Mobile: DaisyUI Accordion -->
				<div tabindex="-4" class="collapse-arrow bg-base-200 collapse lg:hidden">
					<input type="radio" name="landing-accordion" />
					<div class="collapse-title text-primary font-bold">
						{$t('landing.intuitive-title')}
					</div>
					<div class="collapse-content">
						<p class="text-base-content/70">{$t('landing.intuitive-desc')}</p>
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- CTA for mobile/tablet at the bottom -->
	<div class="mt-8 block w-full max-w-2xl lg:hidden">
		<div class="mb-4 text-center">
			<p class="text-base-content/90 mb-2 text-lg font-semibold">
				{$t('landing.cta-title')}
			</p>
			<p class="text-base-content/70">{$t('landing.cta-desc')}</p>
		</div>
		<div class="flex w-full justify-center">
			<a href="/login" class="btn btn-primary btn-lg w-full max-w-xs text-lg font-bold">
				<span class="text-base-100">{$t('landing.cta-button')}</span>
			</a>
		</div>
	</div>
</div>
