import { Calculator, Home, Repeat, Tag } from 'lucide-svelte';
import { locale, setLocale } from '$lib/i18n';
import { theme, updateTheme, type ThemeOption } from '$stores/uiPreferences';
import { get } from 'svelte/store';

export const navLinks = [
	{ href: '/home', icon: Home, labelKey: 'navbar.home' },
	{ href: '/categories', icon: Tag, labelKey: 'navbar.categories' },
	{ href: '/recurring-payments', icon: Repeat, labelKey: 'navbar.recurring-payments' },
	{ href: '/investment-calculator', icon: Calculator, labelKey: 'navbar.calculator' }
] as const;

const themeCycle: ThemeOption[] = ['system', 'dark', 'light'];

export function toggleTheme() {
	const currentIdx = themeCycle.indexOf(get(theme));
	updateTheme(themeCycle[(currentIdx + 1) % themeCycle.length]);
}

export function toggleLanguage() {
	const newLang = get(locale) === 'en' ? 'pt' : 'en';
	locale.set(newLang);
	setLocale(newLang);
}
