<script lang="ts">
	import { t } from '$lib/i18n';
	import {
		Shield,
		Mail,
		Database,
		Eye,
		Download,
		Trash2,
		Calendar,
		Target,
		Check,
		X,
		Lock,
		Server,
		Cookie,
		CookieIcon
	} from 'lucide-svelte';

	let lastUpdated = '07-08-2025';
	const email = 'remigio@graocerto.pt';
</script>

<svelte:head>
	<title>{$t('privacy.page.title')} - Grão Certo</title>
	<meta name="description" content={$t('privacy.page.description')} />
</svelte:head>

<div class="bg-base-100 min-h-screen">
	<div class="container mx-auto max-w-4xl px-4 py-8">
		<!-- Header -->
		<div class="mb-12 text-center">
			<div class="mb-4 flex justify-center">
				<Shield class="text-primary h-16 w-16" />
			</div>
			<h1 class="text-base-content mb-4 text-4xl font-bold">{$t('privacy.title')}</h1>
			<p class="text-base-content/70 text-lg">
				{$t('privacy.subtitle')}
			</p>
			<p class="text-base-content/60 mt-4 text-sm">
				<Calendar class="mr-1 inline h-4 w-4" />
				{$t('privacy.last-updated')}: {lastUpdated}
			</p>
		</div>

		<!-- Content -->
		<div class="space-y-8">
			<!-- 1. Who We Are -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Mail class="text-primary h-6 w-6" />
						{$t('privacy.who-we-are.title')}
					</h2>
					<div class="bg-base-100 rounded-lg p-4">
						<p>
							<strong>Grão Certo</strong>
							{$t('privacy.who-we-are.operated-by')} <strong>Lucas Remígio</strong>
						</p>
						<p class="mt-2">
							<Mail class="text-primary mr-1 inline h-4 w-4" />
							{$t('privacy.who-we-are.contact')}:
							<a href="mailto:{email}" class="link link-primary">{email}</a>
						</p>
					</div>
				</div>
			</section>

			<!-- 2. What Data We Collect -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Database class="text-primary h-6 w-6" />
						{$t('privacy.data-collected.title')}
					</h2>
					<div class="grid gap-4 md:grid-cols-3">
						<div class="bg-base-100 rounded-lg p-4">
							<h3 class="text-primary mb-2 flex items-center gap-2 font-semibold">
								<Mail class="h-4 w-4" />
								{$t('privacy.data-collected.account.title')}
							</h3>
							<ul class="space-y-1 text-sm">
								<li class="flex items-center gap-2">
									<div class="bg-primary h-1 w-1 rounded-full"></div>
									{$t('privacy.data-collected.account.email')}
								</li>
								<li class="flex items-center gap-2">
									<div class="bg-primary h-1 w-1 rounded-full"></div>
									{$t('privacy.data-collected.account.password')}
								</li>
							</ul>
						</div>
						<div class="bg-base-100 rounded-lg p-4">
							<h3 class="text-primary mb-2 flex items-center gap-2 font-semibold">
								<Database class="h-4 w-4" />
								{$t('privacy.data-collected.financial.title')}
							</h3>
							<ul class="space-y-1 text-sm">
								<li class="flex items-center gap-2">
									<div class="bg-primary h-1 w-1 rounded-full"></div>
									{$t('privacy.data-collected.financial.accounts')}
								</li>
								<li class="flex items-center gap-2">
									<div class="bg-primary h-1 w-1 rounded-full"></div>
									{$t('privacy.data-collected.financial.transactions')}
								</li>
							</ul>
						</div>
						<div class="bg-base-100 rounded-lg p-4">
							<h3 class="text-primary mb-2 flex items-center gap-2 font-semibold">
								<Eye class="h-4 w-4" />
								{$t('privacy.data-collected.usage.title')}
							</h3>
							<ul class="space-y-1 text-sm">
								<li class="flex items-center gap-2">
									<div class="bg-primary h-1 w-1 rounded-full"></div>
									{$t('privacy.data-collected.usage.analytics')}
								</li>
								<li class="flex items-center gap-2">
									<div class="bg-primary h-1 w-1 rounded-full"></div>
									{$t('privacy.data-collected.usage.cookies')}
								</li>
							</ul>
						</div>
					</div>
				</div>
			</section>

			<!-- 3. Why We Use It -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Target class="text-primary h-6 w-6" />
						{$t('privacy.usage.title')}
					</h2>
					<div class="bg-primary/10 rounded-lg p-6">
						<p class="mb-4 flex items-center justify-center gap-2 text-center text-lg font-medium">
							<strong>{$t('privacy.usage.purpose')}</strong>
						</p>
						<div class="grid gap-4 text-sm md:grid-cols-2">
							<div class="flex items-center gap-2">
								<Check class="text-success h-4 w-4" />
								<span>{$t('privacy.usage.we-do.store')}</span>
							</div>
							<div class="flex items-center gap-2">
								<Check class="text-success h-4 w-4" />
								<span>{$t('privacy.usage.we-do.authenticate')}</span>
							</div>
							<div class="flex items-center gap-2">
								<X class="text-error h-4 w-4" />
								<span>{$t('privacy.usage.we-dont.sell')}</span>
							</div>
							<div class="flex items-center gap-2">
								<X class="text-error h-4 w-4" />
								<span>{$t('privacy.usage.we-dont.share')}</span>
							</div>
						</div>
					</div>
				</div>
			</section>

			<!-- 4. Where It's Stored -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Server class="text-primary h-6 w-6" />
						{$t('privacy.storage.title')}
					</h2>
					<div class="bg-base-100 rounded-lg p-4">
						<div class="mb-3 flex items-center gap-3">
							<Shield class="text-success h-8 w-8" />
							<div>
								<h3 class="font-semibold">{$t('privacy.storage.location.title')}</h3>
								<p class="text-base-content/70 text-sm">
									{$t('privacy.storage.location.description')}
								</p>
							</div>
						</div>
						<div class="grid gap-4 text-sm md:grid-cols-2">
							<div class="flex items-center gap-2">
								<Lock class="text-success h-4 w-4" />
								<span>{$t('privacy.storage.security.encrypted')}</span>
							</div>
							<div class="flex items-center gap-2">
								<Shield class="text-success h-4 w-4" />
								<span>{$t('privacy.storage.security.gdpr')}</span>
							</div>
						</div>
					</div>
				</div>
			</section>

			<!-- 5. Your Rights -->
			<section class="card bg-primary/10">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Shield class="text-primary h-6 w-6" />
						{$t('privacy.rights.title')}
					</h2>
					<p class="text-base-content/80 mb-4">{$t('privacy.rights.description')}</p>

					<div class="mb-6 grid gap-4 md:grid-cols-3">
						<div class="bg-base-100 rounded-lg p-4 text-center">
							<Trash2 class="text-error mx-auto mb-2 h-8 w-8" />
							<h4 class="font-semibold">{$t('privacy.rights.delete.title')}</h4>
							<p class="text-base-content/70 text-sm">{$t('privacy.rights.delete.description')}</p>
						</div>
						<div class="bg-base-100 rounded-lg p-4 text-center">
							<Download class="text-primary mx-auto mb-2 h-8 w-8" />
							<h4 class="font-semibold">{$t('privacy.rights.export.title')}</h4>
							<p class="text-base-content/70 text-sm">{$t('privacy.rights.export.description')}</p>
						</div>
						<div class="bg-base-100 rounded-lg p-4 text-center">
							<Mail class="text-success mx-auto mb-2 h-8 w-8" />
							<h4 class="font-semibold">{$t('privacy.rights.contact.title')}</h4>
							<p class="text-base-content/70 text-sm">{$t('privacy.rights.contact.description')}</p>
						</div>
					</div>

					<div class="bg-base-100 rounded-lg p-4 text-center">
						<p class="font-medium">{$t('privacy.rights.help.question')}</p>
						<a href="mailto:{email}" class="btn btn-primary btn-sm mt-2">
							<Mail class="mr-1 h-4 w-4" />
							{$t('privacy.rights.help.contact')}
							{email}
						</a>
					</div>
				</div>
			</section>

			<!-- 6. Cookies -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Cookie class="text-primary h-6 w-6" />
						{$t('privacy.cookies.title')}
					</h2>
					<div class="bg-base-100 rounded-lg p-4">
						<div class="mb-3 flex items-center gap-3">
							<Cookie class="text-primary h-8 w-8" />
							<div>
								<h3 class="font-semibold">{$t('privacy.cookies.usage.title')}</h3>
								<p class="text-base-content/70 text-sm">
									{$t('privacy.cookies.usage.description')}
								</p>
							</div>
						</div>
						<p class="text-sm">
							{$t('privacy.cookies.management')}
						</p>
					</div>
				</div>
			</section>
		</div>

		<!-- Footer -->
		<div class="mt-12 text-center">
			<div class="bg-base-200 mb-6 rounded-lg p-6">
				<p class="text-base-content/70 mb-4">
					<strong>{$t('privacy.footer.questions')}</strong><br />
					{$t('privacy.footer.response-time')}
				</p>
				<a href="mailto:{email}" class="btn btn-outline btn-sm mr-4">
					<Mail class="mr-1 h-4 w-4" />
					{$t('privacy.footer.email-us')}
				</a>
			</div>

			<a href="/" class="btn btn-primary btn-lg">
				<Shield class="mr-2 h-4 w-4" />
				{$t('privacy.footer.back-home')}
			</a>
		</div>
	</div>
</div>
