<script lang="ts">
	import { t } from '$lib/i18n';
	import { Github, Linkedin, Mail } from 'lucide-svelte';
	import { cookieConsent } from './stores/cookieConsent';

	function reopenCookieBanner() {
		cookieConsent.reopenBanner();
	}
</script>

<footer class="sm:footer-horizontal footer bg-base-300 p-10 text-base-content">
	<!-- Logo and Brand Section -->
	<aside class="flex flex-col items-center sm:items-start">
		<div class="mb-2 flex items-center gap-2">
			<img src="/favicon.svg" alt="Grão Certo" class="h-8 w-8" />
			<span class="text-xl font-bold text-primary">Grão Certo</span>
		</div>
		<p class="max-w-xs text-center text-sm text-base-content/70 sm:text-left">
			{$t('footer.tagline', { default: 'Simple and secure wallet tracking for everyone.' })}
		</p>
		<p class="mt-2 text-xs text-base-content/60">
			{$t('footer.made-by', { default: 'Made by Lucas Remígio' })}
		</p>
	</aside>

	<!-- App Section -->
	<nav>
		<h6 class="footer-title">{$t('footer.app', { default: 'App' })}</h6>
		<a href="/register" class="link-hover link"
			>{$t('footer.get-started', { default: 'Get Started' })}</a
		>
		<a href="/login" class="link-hover link">{$t('footer.login', { default: 'Login' })}</a>
		<a href="/#features" class="link-hover link">{$t('footer.features', { default: 'Features' })}</a
		>
		<a href="/#faq" class="link-hover link">{$t('footer.faq', { default: 'FAQ' })}</a>
	</nav>

	<!-- Legal Section -->
	<nav>
		<h6 class="footer-title">{$t('footer.legal', { default: 'Legal' })}</h6>
		<a href="/privacy-policy" class="link-hover link">
			{$t('footer.privacy-policy', { default: 'Privacy Policy' })}
		</a>
		<button class="link-hover link text-left" on:click={reopenCookieBanner}>
			{$t('footer.cookie-settings', { default: 'Cookie Settings' })}
		</button>
	</nav>

	<!-- Contact & Social Section -->
	<nav>
		<h6 class="footer-title">{$t('footer.contact', { default: 'Contact' })}</h6>
		<div class="flex gap-4">
			<a
				href="mailto:remigio@graocerto.pt"
				class="btn btn-circle btn-ghost btn-sm transition-transform duration-200 hover:scale-110"
				aria-label="Email"
			>
				<Mail class="h-5 w-5 text-primary" />
			</a>
			<a
				href="https://linkedin.com/in/lucas-remigio"
				target="_blank"
				rel="noopener"
				class="btn btn-circle btn-ghost btn-sm transition-transform duration-200 hover:scale-110"
				aria-label="LinkedIn"
			>
				<Linkedin class="h-5 w-5 text-primary" />
			</a>
			<a
				href="https://github.com/lucas-remigio"
				target="_blank"
				rel="noopener"
				class="btn btn-circle btn-ghost btn-sm transition-transform duration-200 hover:scale-110"
				aria-label="GitHub"
			>
				<Github class="h-5 w-5 text-primary" />
			</a>
		</div>
		<p class="mt-2 text-xs text-base-content/60">
			{$t('footer.response-time', { default: 'Response within 48h' })}
		</p>
	</nav>
</footer>
