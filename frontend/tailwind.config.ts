import containerQueries from '@tailwindcss/container-queries';
import forms from '@tailwindcss/forms';
import typography from '@tailwindcss/typography';
import type { Config } from 'tailwindcss';
import daisyui from 'daisyui';

export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],

	theme: {
		extend: {}
	},

	// Add DaisyUI theme customization
	daisyui: {
		themes: [
			{
				light: {
					primary: '#006FF9', // Primary color
					secondary: '#767DEA', // Secondary color
					accent: '#2DD4BF', // Accent color
					neutral: '#2E3440', // Neutral color
					'base-100': '#F9FAFB', // Base background
					'base-200': '#ECEEF2', // Secondary background
					'base-300': '#DCE0E8', // Tertiary background
					'base-content': '#1F2937', // Base text color
					info: '#0EA5E9', // Info
					success: '#10B981', // Success
					warning: '#F59E0B', // Warning
					error: '#EF4444' // Error
				},
				dark: {
					primary: '#4F99FF', // Primary color (lighter in dark mode)
					secondary: '#8388F0', // Secondary color (lighter in dark mode)
					accent: '#3DE0CB', // Accent color (lighter in dark mode)
					neutral: '#3B4252', // Neutral color
					'base-100': '#121212', // Base background (darker)
					'base-200': '#1E1E1E', // Secondary background (darker)
					'base-300': '#292929', // Tertiary background (darker)
					'base-content': '#E0E0E0', // Base text color (lighter for contrast)
					info: '#38BDF8', // Info
					success: '#34D399', // Success
					warning: '#FBBF24', // Warning
					error: '#F87171' // Error
				}
			}
		],
		darkTheme: 'dark' // Name of the dark theme
	},

	plugins: [typography, forms, containerQueries, daisyui]
} satisfies Config;
