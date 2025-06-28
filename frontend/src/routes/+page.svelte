<script lang="ts">
	import { t } from '$lib/i18n';
</script>

<div class="bg-base-100 flex min-h-screen flex-col items-center justify-center px-4 py-12">
	<div class="flex w-full max-w-2xl flex-col items-center">
		<img src="/logo.png" alt="Grão Certo Logo" class="mb-6 h-20 w-auto" />
		<h1 class="text-primary mb-4 text-center text-5xl font-extrabold">Grão Certo</h1>
		<p class="text-base-content/80 mb-8 text-center text-lg">
			<strong>Grão Certo</strong>
			{$t('landing.intro')}
		</p>

		<div class="mb-8 grid w-full grid-cols-1 gap-6 md:grid-cols-2">
			<div class="bg-base-200 flex flex-col items-center rounded-xl p-6 shadow">
				<h2 class="text-primary mb-2 text-xl font-bold">{$t('landing.accounts-title')}</h2>
				<p class="text-base-content/70 text-center">{$t('landing.accounts-desc')}</p>
			</div>
			<div class="bg-base-200 flex flex-col items-center rounded-xl p-6 shadow">
				<h2 class="text-primary mb-2 text-xl font-bold">{$t('landing.analytics-title')}</h2>
				<p class="text-base-content/70 text-center">{$t('landing.analytics-desc')}</p>
			</div>
			<div class="bg-base-200 flex flex-col items-center rounded-xl p-6 shadow">
				<h2 class="text-primary mb-2 text-xl font-bold">{$t('landing.privacy-title')}</h2>
				<p class="text-base-content/70 text-center">{$t('landing.privacy-desc')}</p>
			</div>
			<div class="bg-base-200 flex flex-col items-center rounded-xl p-6 shadow">
				<h2 class="text-primary mb-2 text-xl font-bold">{$t('landing.intuitive-title')}</h2>
				<p class="text-base-content/70 text-center">{$t('landing.intuitive-desc')}</p>
			</div>
		</div>

		<div class="mb-10 mt-2 text-center">
			<p class="text-base-content/90 mb-2 text-lg font-semibold">
				{$t('landing.cta-title')}
			</p>
			<p class="text-base-content/70">{$t('landing.cta-desc')}</p>
		</div>

		<a href="/login" class="btn btn-primary btn-lg w-full max-w-xs text-lg font-bold">
			<span class="text-base-100">{$t('landing.cta-button')}</span>
		</a>
	</div>
</div>
