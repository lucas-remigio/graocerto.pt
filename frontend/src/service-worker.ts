/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />

/**
 * Self-destroying service worker.
 *
 * The previous service worker intercepted every GET request through a
 * `networkFirst` caching proxy, which caused intermittent multi-second hangs in
 * production (cold-started worker + no fetch timeout + caching cross-origin API
 * responses). This version has NO `fetch` listener, so requests are never
 * intercepted, and it unregisters itself + clears old caches on activation so it
 * is removed from clients that already installed the old worker.
 *
 * Push notifications are intentionally dropped here. If we bring them back,
 * re-add a `push` + `notificationclick` listener only — never a `fetch` handler
 * that touches API / cross-origin requests.
 */

const sw = globalThis as unknown as ServiceWorkerGlobalScope;

sw.addEventListener('install', () => {
	// Activate immediately instead of waiting for existing tabs to close.
	sw.skipWaiting();
});

sw.addEventListener('activate', (event) => {
	event.waitUntil(
		(async () => {
			// Drop every cache the old worker created.
			const keys = await caches.keys();
			await Promise.all(keys.map((key) => caches.delete(key)));

			// Remove this worker entirely.
			await sw.registration.unregister();

			// Reload open tabs so they detach from the (now gone) worker and go
			// straight to the network from here on.
			const clients = await sw.clients.matchAll({ type: 'window' });
			for (const client of clients) {
				(client as WindowClient).navigate((client as WindowClient).url);
			}
		})()
	);
});
