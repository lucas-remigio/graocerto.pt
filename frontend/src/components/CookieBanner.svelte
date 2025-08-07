<script lang="ts">
	import { cookieConsent } from '$lib/stores/cookieConsent';
	import { Cookie, Settings, X } from 'lucide-svelte';
	import { t } from '$lib/i18n';

	let showBanner = false;
	let showPreferences = false;
	let tempPreferences = {
		analytics: false,
		marketing: false
	};

	// Show banner if user hasn't made a choice
	$: showBanner = !$cookieConsent.hasChosenPreferences;

	function acceptAll() {
		cookieConsent.acceptAll();
		showBanner = false;
		showPreferences = false;
	}

	function rejectAll() {
		cookieConsent.rejectAll();
		showBanner = false;
		showPreferences = false;
	}

	function openPreferences() {
		tempPreferences.analytics = $cookieConsent.analytics;
		tempPreferences.marketing = $cookieConsent.marketing;
		showPreferences = true;
	}

	function savePreferences() {
		cookieConsent.setPreferences(tempPreferences);
		showBanner = false;
		showPreferences = false;

	}

</script>

<!-- Cookie Banner -->
{#if showBanner && !showPreferences}
	<div
		class="bg-base-300/95 border-base-content/10 fixed bottom-0 left-0 right-0 z-50 border-t shadow-xl backdrop-blur-sm"
	>
		<div class="container mx-auto max-w-6xl px-4 py-6">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
				<div class="flex flex-1 items-start gap-3">
					<Cookie class="text-primary mt-1 h-6 w-6 flex-shrink-0" />
					<div class="flex-1">
						<h3 class="text-base-content mb-2 font-semibold">
							{$t('cookies.banner.title', { default: 'We value your privacy' })}
						</h3>
						<p class="text-base-content/80 text-sm leading-relaxed">
							{$t('cookies.banner.description', {
								default:
									'We use cookies to enhance your experience, analyze site usage, and improve our services. Choose your preferences below or accept all to continue.'
							})}
							<a href="/privacy-policy" class="link link-primary ml-1">
								{$t('cookies.banner.learn-more', { default: 'Learn more' })}
							</a>
						</p>
					</div>
				</div>

				<div class="flex flex-col gap-3 sm:flex-row lg:flex-shrink-0">
					<button class="btn btn-ghost btn-sm gap-2" on:click={openPreferences}>
						<Settings class="h-4 w-4" />
						{$t('cookies.banner.customize', { default: 'Customize' })}
					</button>
					<button class="btn btn-outline btn-sm" on:click={rejectAll}>
						{$t('cookies.banner.reject', { default: 'Reject All' })}
					</button>
					<button class="btn btn-primary btn-sm" on:click={acceptAll}>
						{$t('cookies.banner.accept', { default: 'Accept All' })}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- Cookie Preferences Modal -->
{#if showPreferences}
	<div
		class="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
	>
		<div class="card bg-base-100 w-full max-w-md shadow-xl">
			<div class="card-body">
				<div class="mb-4 flex items-center justify-between">
					<h3 class="card-title text-lg">
						<Cookie class="text-primary h-5 w-5" />
						{$t('cookies.preferences.title', { default: 'Cookie Preferences' })}
					</h3>
					<button
						class="btn btn-ghost btn-sm btn-circle"
						on:click={() => (showPreferences = false)}
						aria-label="Close"
					>
						<X class="h-4 w-4" />
					</button>
				</div>

				<div class="space-y-4">
					<!-- Necessary Cookies -->
					<div class="form-control">
						<label class="label cursor-pointer justify-start gap-3">
							<input type="checkbox" checked={true} disabled class="checkbox checkbox-primary" />
							<div class="flex-1">
								<span class="label-text font-medium">
									{$t('cookies.preferences.necessary', { default: 'Necessary' })}
								</span>
								<p class="text-base-content/60 mt-1 text-xs">
									{$t('cookies.preferences.necessary-desc', {
										default: 'Required for the website to function properly. Cannot be disabled.'
									})}
								</p>
							</div>
						</label>
					</div>

					<!-- Analytics Cookies -->
					<div class="form-control">
						<label class="label cursor-pointer justify-start gap-3">
							<input
								type="checkbox"
								bind:checked={tempPreferences.analytics}
								class="checkbox checkbox-primary"
							/>
							<div class="flex-1">
								<span class="label-text font-medium">
									{$t('cookies.preferences.analytics', { default: 'Analytics' })}
								</span>
								<p class="text-base-content/60 mt-1 text-xs">
									{$t('cookies.preferences.analytics-desc', {
										default: 'Help us understand how visitors interact with our website.'
									})}
								</p>
							</div>
						</label>
					</div>

					<!-- Marketing Cookies -->
					<div class="form-control">
						<label class="label cursor-pointer justify-start gap-3">
							<input
								type="checkbox"
								bind:checked={tempPreferences.marketing}
								class="checkbox checkbox-primary"
							/>
							<div class="flex-1">
								<span class="label-text font-medium">
									{$t('cookies.preferences.marketing', { default: 'Marketing' })}
								</span>
								<p class="text-base-content/60 mt-1 text-xs">
									{$t('cookies.preferences.marketing-desc', {
										default: 'Used to track visitors across websites for marketing purposes.'
									})}
								</p>
							</div>
						</label>
					</div>
				</div>

				<div class="card-actions mt-6 justify-end gap-2">
					<button class="btn btn-ghost btn-sm" on:click={() => (showPreferences = false)}>
						{$t('cookies.preferences.cancel', { default: 'Cancel' })}
					</button>
					<button class="btn btn-primary btn-sm" on:click={savePreferences}>
						{$t('cookies.preferences.save', { default: 'Save Preferences' })}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
