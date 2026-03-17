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
		ArrowLeft
	} from 'lucide-svelte';

	let lastUpdated = '07-08-2025';
	const email = 'remigio@graocerto.pt';
</script>

<svelte:head>
	<title>{$t('privacy.page.title')} - Grão Certo</title>
	<meta name="description" content={$t('privacy.page.description')} />
</svelte:head>

<div class="min-h-screen bg-base-100">
	<div class="container mx-auto max-w-4xl px-4 py-8">
		<!-- Header -->
		<div class="mb-12 text-center">
			<div class="mb-4 flex justify-center">
				<Shield class="h-16 w-16 text-primary" />
			</div>
			<h1 class="mb-4 text-4xl font-bold text-base-content">{$t('privacy.title')}</h1>
			<p class="text-lg text-base-content/70">
				{$t('privacy.subtitle')}
			</p>
			<p class="mt-4 text-sm text-base-content/60">
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
						<Mail class="h-6 w-6 text-primary" />
						{$t('privacy.who-we-are.title')}
					</h2>
					<div class="rounded-lg bg-base-100 p-4">
						<p>
							<strong>Grão Certo</strong>
							{$t('privacy.who-we-are.operated-by')} <strong>Lucas Remígio</strong>
						</p>
						<p class="mt-2">
							<Mail class="mr-1 inline h-4 w-4 text-primary" />
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
						<Database class="h-6 w-6 text-primary" />
						{$t('privacy.data-collected.title')}
					</h2>
					<div class="grid gap-4 md:grid-cols-3">
						<div class="rounded-lg bg-base-100 p-4">
							<h3 class="mb-2 flex items-center gap-2 font-semibold text-primary">
								<Mail class="h-4 w-4" />
								{$t('privacy.data-collected.account.title')}
							</h3>
							<ul class="space-y-1 text-sm">
								<li class="flex items-center gap-2">
									<div class="h-1 w-1 rounded-full bg-primary"></div>
									{$t('privacy.data-collected.account.email')}
								</li>
								<li class="flex items-center gap-2">
									<div class="h-1 w-1 rounded-full bg-primary"></div>
									{$t('privacy.data-collected.account.password')}
								</li>
							</ul>
						</div>
						<div class="rounded-lg bg-base-100 p-4">
							<h3 class="mb-2 flex items-center gap-2 font-semibold text-primary">
								<Database class="h-4 w-4" />
								{$t('privacy.data-collected.financial.title')}
							</h3>
							<ul class="space-y-1 text-sm">
								<li class="flex items-center gap-2">
									<div class="h-1 w-1 rounded-full bg-primary"></div>
									{$t('privacy.data-collected.financial.accounts')}
								</li>
								<li class="flex items-center gap-2">
									<div class="h-1 w-1 rounded-full bg-primary"></div>
									{$t('privacy.data-collected.financial.transactions')}
								</li>
							</ul>
						</div>
						<div class="rounded-lg bg-base-100 p-4">
							<h3 class="mb-2 flex items-center gap-2 font-semibold text-primary">
								<Eye class="h-4 w-4" />
								{$t('privacy.data-collected.usage.title')}
							</h3>
							<ul class="space-y-1 text-sm">
								<li class="flex items-center gap-2">
									<div class="h-1 w-1 rounded-full bg-primary"></div>
									{$t('privacy.data-collected.usage.analytics')}
								</li>
								<li class="flex items-center gap-2">
									<div class="h-1 w-1 rounded-full bg-primary"></div>
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
						<Target class="h-6 w-6 text-primary" />
						{$t('privacy.usage.title')}
					</h2>
					<div class="rounded-lg bg-primary/10 p-6">
						<p class="mb-4 flex items-center justify-center gap-2 text-center text-lg font-medium">
							<strong>{$t('privacy.usage.purpose')}</strong>
						</p>
						<div class="grid gap-4 text-sm md:grid-cols-2">
							<div class="flex items-center gap-2">
								<Check class="h-4 w-4 text-success" />
								<span>{$t('privacy.usage.we-do.store')}</span>
							</div>
							<div class="flex items-center gap-2">
								<Check class="h-4 w-4 text-success" />
								<span>{$t('privacy.usage.we-do.authenticate')}</span>
							</div>
							<div class="flex items-center gap-2">
								<X class="h-4 w-4 text-error" />
								<span>{$t('privacy.usage.we-dont.sell')}</span>
							</div>
							<div class="flex items-center gap-2">
								<X class="h-4 w-4 text-error" />
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
						<Server class="h-6 w-6 text-primary" />
						{$t('privacy.storage.title')}
					</h2>
					<div class="rounded-lg bg-base-100 p-4">
						<div class="mb-3 flex items-center gap-3">
							<Shield class="h-8 w-8 text-success" />
							<div>
								<h3 class="font-semibold">{$t('privacy.storage.location.title')}</h3>
								<p class="text-sm text-base-content/70">
									{$t('privacy.storage.location.description')}
								</p>
							</div>
						</div>
						<div class="grid gap-4 text-sm md:grid-cols-2">
							<div class="flex items-center gap-2">
								<Lock class="h-4 w-4 text-success" />
								<span>{$t('privacy.storage.security.encrypted')}</span>
							</div>
							<div class="flex items-center gap-2">
								<Shield class="h-4 w-4 text-success" />
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
						<Shield class="h-6 w-6 text-primary" />
						{$t('privacy.rights.title')}
					</h2>
					<p class="mb-4 text-base-content/80">{$t('privacy.rights.description')}</p>

					<div class="mb-6 grid gap-4 md:grid-cols-3">
						<div class="rounded-lg bg-base-100 p-4 text-center">
							<Trash2 class="mx-auto mb-2 h-8 w-8 text-error" />
							<h4 class="font-semibold">{$t('privacy.rights.delete.title')}</h4>
							<p class="text-sm text-base-content/70">{$t('privacy.rights.delete.description')}</p>
						</div>
						<div class="rounded-lg bg-base-100 p-4 text-center">
							<Download class="mx-auto mb-2 h-8 w-8 text-primary" />
							<h4 class="font-semibold">{$t('privacy.rights.export.title')}</h4>
							<p class="text-sm text-base-content/70">{$t('privacy.rights.export.description')}</p>
						</div>
						<div class="rounded-lg bg-base-100 p-4 text-center">
							<Mail class="mx-auto mb-2 h-8 w-8 text-success" />
							<h4 class="font-semibold">{$t('privacy.rights.contact.title')}</h4>
							<p class="text-sm text-base-content/70">{$t('privacy.rights.contact.description')}</p>
						</div>
					</div>

					<div class="rounded-lg bg-base-100 p-4 text-center">
						<p class="font-medium">{$t('privacy.rights.help.question')}</p>
						<a href="mailto:{email}" class="btn btn-primary btn-sm mt-2 text-base-100">
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
						<Cookie class="h-6 w-6 text-primary" />
						{$t('privacy.cookies.title')}
					</h2>
					<div class="rounded-lg bg-base-100 p-4">
						<div class="mb-3 flex items-center gap-3">
							<Cookie class="h-8 w-8 text-primary" />
							<div>
								<h3 class="font-semibold">{$t('privacy.cookies.usage.title')}</h3>
								<p class="text-sm text-base-content/70">
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
			<div class="mb-6 rounded-lg bg-base-200 p-6">
				<p class="mb-4 text-base-content/70">
					<strong>{$t('privacy.footer.questions')}</strong><br />
					{$t('privacy.footer.response-time')}
				</p>
				<a href="mailto:{email}" class="btn btn-outline btn-sm mr-4">
					<Mail class="mr-1 h-4 w-4" />
					{$t('privacy.footer.email-us')}
				</a>
			</div>

			<a href="/" class="btn btn-primary btn-lg text-base-100">
				<ArrowLeft class="mr-2 h-4 w-4" />
				{$t('privacy.footer.back-home')}
			</a>
		</div>
	</div>
</div>
