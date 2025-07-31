<script lang="ts">
	import { t } from '$lib/i18n';
	import { Github, Heart, Linkedin, Mail } from 'lucide-svelte';
</script>

<footer class="footer footer-center bg-base-200 text-base-content w-full rounded-xl p-6 shadow-lg">
	<div class="flex flex-col items-center gap-1">
		<small class="flex items-center gap-2 text-lg font-medium">
			<!-- Small icon left of Grão Certo -->
			<img src="/favicon.svg" alt="Grão Certo" class="mr-1 inline-block h-6 w-6 align-middle" />
			Grão Certo &mdash; {$t('landing.made-with')}
			<span class="text-error mx-1 inline-block align-middle">
				<Heart fill="currentColor" class="h-5 w-5" />
			</span>
			{$t('common.by')} <span class="text-primary font-semibold">Lucas Remígio</span>
		</small>
		<h6 class="text-primary mt-2 text-base font-semibold tracking-wide">
			{$t('landing.dev-contact-me')}
		</h6>
		<div class="flex justify-center gap-4">
			<a
				href="mailto:remigio@graocerto.pt"
				class="btn btn-ghost btn-circle transition-transform duration-200 hover:scale-110"
				aria-label="Email"
			>
				<Mail class="text-primary h-6 w-6" />
			</a>
			<a
				href="https://linkedin.com/in/lucas-remigio"
				target="_blank"
				rel="noopener"
				class="btn btn-ghost btn-circle transition-transform duration-200 hover:scale-110"
				aria-label="LinkedIn"
			>
				<Linkedin class="text-primary h-6 w-6" />
			</a>
			<a
				href="https://github.com/lucas-remigio"
				target="_blank"
				rel="noopener"
				class="btn btn-ghost btn-circle transition-transform duration-200 hover:scale-110"
				aria-label="GitHub"
			>
				<Github class="text-primary h-6 w-6" />
			</a>
		</div>
	</div>
</footer>
