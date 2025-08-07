<script lang="ts">
	import { t } from '$lib/i18n';
	import { Github, Linkedin, Mail, Cookie } from 'lucide-svelte';
	import { cookieConsent } from './stores/cookieConsent';

	function reopenCookieBanner() {
		cookieConsent.reopenBanner();
	}
</script>

<footer class="footer sm:footer-horizontal bg-base-300 text-base-content p-10">
	<!-- Logo and Brand Section -->
	<aside class="flex flex-col items-center sm:items-start">
		<div class="mb-2 flex items-center gap-2">
			<img src="/favicon.svg" alt="Grão Certo" class="h-8 w-8" />
			<span class="text-primary text-xl font-bold">Grão Certo</span>
		</div>
		<p class="text-base-content/70 max-w-xs text-center text-sm sm:text-left">
			{$t('footer.tagline', { default: 'Simple and secure wallet tracking for everyone.' })}
		</p>
		<p class="text-base-content/60 mt-2 text-xs">
			© {new Date().getFullYear()} Lucas Remígio
		</p>
	</aside>

	<!-- App Section -->
	<nav>
		<h6 class="footer-title">{$t('footer.app', { default: 'App' })}</h6>
		<a href="/register" class="link link-hover"
			>{$t('footer.get-started', { default: 'Get Started' })}</a
		>
		<a href="/login" class="link link-hover">{$t('footer.login', { default: 'Login' })}</a>
		<a href="/#features" class="link link-hover">{$t('footer.features', { default: 'Features' })}</a
		>
		<a href="/#faq" class="link link-hover">{$t('footer.faq', { default: 'FAQ' })}</a>
	</nav>

	<!-- Legal Section -->
	<nav>
		<h6 class="footer-title">{$t('footer.legal', { default: 'Legal' })}</h6>
		<a href="/privacy-policy" class="link link-hover">
			{$t('footer.privacy-policy', { default: 'Privacy Policy' })}
		</a>
		<button class="link link-hover text-left" on:click={reopenCookieBanner}>
			{$t('footer.cookie-settings', { default: 'Cookie Settings' })}
		</button>
	</nav>

	<!-- Contact & Social Section -->
	<nav>
		<h6 class="footer-title">{$t('footer.contact', { default: 'Contact' })}</h6>
		<div class="flex gap-4">
			<a
				href="mailto:remigio@graocerto.pt"
				class="btn btn-ghost btn-circle btn-sm transition-transform duration-200 hover:scale-110"
				aria-label="Email"
			>
				<Mail class="text-primary h-5 w-5" />
			</a>
			<a
				href="https://linkedin.com/in/lucas-remigio"
				target="_blank"
				rel="noopener"
				class="btn btn-ghost btn-circle btn-sm transition-transform duration-200 hover:scale-110"
				aria-label="LinkedIn"
			>
				<Linkedin class="text-primary h-5 w-5" />
			</a>
			<a
				href="https://github.com/lucas-remigio"
				target="_blank"
				rel="noopener"
				class="btn btn-ghost btn-circle btn-sm transition-transform duration-200 hover:scale-110"
				aria-label="GitHub"
			>
				<Github class="text-primary h-5 w-5" />
			</a>
		</div>
		<p class="text-base-content/60 mt-2 text-xs">
			{$t('footer.response-time', { default: 'Response within 48h' })}
		</p>
	</nav>
</footer>
