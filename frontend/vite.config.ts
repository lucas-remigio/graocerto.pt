import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	build: {
		// SPA and HTTP/2 optimizations
		rollupOptions: {
			output: {
				// Better chunking for HTTP/2 multiplexing
				manualChunks: {
					vendor: ['svelte']
				},
				// Ensure proper cache busting with consistent naming
				entryFileNames: 'assets/[name]-[hash].js',
				chunkFileNames: 'assets/[name]-[hash].js',
				assetFileNames: 'assets/[name]-[hash].[ext]'
			}
		},
		// Disable source maps for production
		sourcemap: false,
		// Optimize chunk size for HTTP/2
		chunkSizeWarningLimit: 1000,
		// Ensure proper cache invalidation
		cssCodeSplit: true
	}
});
