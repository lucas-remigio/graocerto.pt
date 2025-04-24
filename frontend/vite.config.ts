import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],

	// Define environment variables with fallbacks
	define: {
		'import.meta.env.VITE_BACKEND_URL': JSON.stringify(
			process.env.VITE_BACKEND_URL || 'http://localhost:8080'
		),
		'import.meta.env.VITE_SOCKETS_URL': JSON.stringify(
			process.env.VITE_SOCKETS_URL || 'ws://localhost:8090'
		)
	}
});
