<script lang="ts">
	import { Shield, Mail, Database, Eye, Download, Trash2, Calendar } from 'lucide-svelte';

	let lastUpdated = '07-08-2025';
	const email = 'remigio@graocerto.pt';
</script>

<svelte:head>
	<title>Privacy Policy - Grão Certo</title>
	<meta
		name="description"
		content="Privacy Policy for Grão Certo - How we collect, use, and protect your personal data."
	/>
</svelte:head>

<div class="bg-base-100 min-h-screen">
	<div class="container mx-auto max-w-4xl px-4 py-8">
		<!-- Header -->
		<div class="mb-12 text-center">
			<div class="mb-4 flex justify-center">
				<Shield class="text-primary h-16 w-16" />
			</div>
			<h1 class="text-base-content mb-4 text-4xl font-bold">Privacy Policy</h1>
			<p class="text-base-content/70 text-lg">
				How we collect, use, and protect your personal data
			</p>
			<p class="text-base-content/60 mt-4 text-sm">
				<Calendar class="mr-1 inline h-4 w-4" />
				Last updated: {lastUpdated}
			</p>
		</div>

		<!-- Content -->
		<div class="space-y-8">
			<!-- 1. Who We Are -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Mail class="text-primary h-6 w-6" />
						Who We Are
					</h2>
					<div class="bg-base-100 rounded-lg p-4">
						<p><strong>Grão Certo</strong> is operated by <strong>Lucas Remígio</strong></p>
						<p class="mt-2">
							<Mail class="text-primary mr-1 inline h-4 w-4" />
							Contact: <a href="mailto:{email}" class="link link-primary">{email}</a>
						</p>
					</div>
				</div>
			</section>

			<!-- 2. What Data We Collect -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Database class="text-primary h-6 w-6" />
						What Data We Collect
					</h2>
					<div class="grid gap-4 md:grid-cols-3">
						<div class="bg-base-100 rounded-lg p-4">
							<h3 class="text-primary mb-2 font-semibold">Account Info</h3>
							<ul class="space-y-1 text-sm">
								<li>• Email address</li>
								<li>• Encrypted password</li>
							</ul>
						</div>
						<div class="bg-base-100 rounded-lg p-4">
							<h3 class="text-primary mb-2 font-semibold">Financial Data</h3>
							<ul class="space-y-1 text-sm">
								<li>• Account names & balances</li>
								<li>• Transaction details</li>
							</ul>
						</div>
						<div class="bg-base-100 rounded-lg p-4">
							<h3 class="text-primary mb-2 font-semibold">Usage Data</h3>
							<ul class="space-y-1 text-sm">
								<li>• Analytics (with consent)</li>
								<li>• Essential cookies only</li>
							</ul>
						</div>
					</div>
				</div>
			</section>

			<!-- 3. Why We Use It -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Eye class="text-primary h-6 w-6" />
						Why We Use Your Data
					</h2>
					<div class="bg-primary/10 rounded-lg p-6">
						<p class="mb-4 text-center text-lg font-medium">
							🎯 <strong>Simple Purpose:</strong> To provide the wallet tracking service you signed up
							for
						</p>
						<div class="grid gap-4 text-sm md:grid-cols-2">
							<div class="flex items-center gap-2">
								<span class="text-success">✅</span>
								<span>Store your financial data securely</span>
							</div>
							<div class="flex items-center gap-2">
								<span class="text-success">✅</span>
								<span>Authenticate your account</span>
							</div>
							<div class="flex items-center gap-2">
								<span class="text-error">❌</span>
								<span>We don't sell your data</span>
							</div>
							<div class="flex items-center gap-2">
								<span class="text-error">❌</span>
								<span>We don't share with third parties</span>
							</div>
						</div>
					</div>
				</div>
			</section>

			<!-- 4. Where It's Stored -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Shield class="text-primary h-6 w-6" />
						Where Your Data Is Stored
					</h2>
					<div class="bg-base-100 rounded-lg p-4">
						<div class="mb-3 flex items-center gap-3">
							<Shield class="text-success h-8 w-8" />
							<div>
								<h3 class="font-semibold">Secure EU Servers</h3>
								<p class="text-base-content/70 text-sm">Hosted by Supabase in European Union</p>
							</div>
						</div>
						<div class="grid gap-4 text-sm md:grid-cols-2">
							<div class="flex items-center gap-2">
								<span class="text-success">🔒</span>
								<span>Encrypted at rest and in transit</span>
							</div>
							<div class="flex items-center gap-2">
								<span class="text-success">🛡️</span>
								<span>GDPR compliant infrastructure</span>
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
						Your Rights
					</h2>
					<p class="text-base-content/80 mb-4">You have full control over your data:</p>

					<div class="mb-6 grid gap-4 md:grid-cols-3">
						<div class="bg-base-100 rounded-lg p-4 text-center">
							<Trash2 class="text-error mx-auto mb-2 h-8 w-8" />
							<h4 class="font-semibold">Delete Account</h4>
							<p class="text-base-content/70 text-sm">Remove all your data permanently</p>
						</div>
						<div class="bg-base-100 rounded-lg p-4 text-center">
							<Download class="text-primary mx-auto mb-2 h-8 w-8" />
							<h4 class="font-semibold">Export Data</h4>
							<p class="text-base-content/70 text-sm">Get a copy of all your information</p>
						</div>
						<div class="bg-base-100 rounded-lg p-4 text-center">
							<Mail class="text-success mx-auto mb-2 h-8 w-8" />
							<h4 class="font-semibold">Contact Us</h4>
							<p class="text-base-content/70 text-sm">For any data-related requests</p>
						</div>
					</div>

					<div class="bg-base-100 rounded-lg p-4 text-center">
						<p class="font-medium">Need help with your data?</p>
						<a href="mailto:{email}" class="btn btn-primary btn-sm mt-2">
							<Mail class="mr-1 h-4 w-4" />
							Contact {email}
						</a>
					</div>
				</div>
			</section>

			<!-- 6. Cookies -->
			<section class="card bg-base-200/50">
				<div class="card-body">
					<h2 class="card-title mb-4 text-2xl">
						<Database class="text-primary h-6 w-6" />
						Cookies
					</h2>
					<div class="bg-base-100 rounded-lg p-4">
						<div class="mb-3 flex items-center gap-3">
							<span class="text-2xl">🍪</span>
							<div>
								<h3 class="font-semibold">Minimal Cookie Usage</h3>
								<p class="text-base-content/70 text-sm">
									We only use essential cookies for login functionality
								</p>
							</div>
						</div>
						<p class="text-sm">
							You can manage your cookie preferences through our cookie banner that appears on your
							first visit.
						</p>
					</div>
				</div>
			</section>
		</div>

		<!-- Footer -->
		<div class="mt-12 text-center">
			<div class="bg-base-200 mb-6 rounded-lg p-6">
				<p class="text-base-content/70 mb-4">
					<strong>Questions about this policy?</strong><br />
					We're here to help and will respond within 48 hours.
				</p>
				<a href="mailto:{email}" class="btn btn-outline btn-sm mr-4">
					<Mail class="mr-1 h-4 w-4" />
					Email Us
				</a>
			</div>

			<a href="/" class="btn btn-primary btn-lg">
				<span class="mr-2">←</span>
				Back to Home
			</a>
		</div>
	</div>
</div>
