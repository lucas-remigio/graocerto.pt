/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

const CACHE = `cache-${version}`;
const ASSETS = [...build, ...files];

const self = globalThis as unknown as ServiceWorkerGlobalScope;

self.addEventListener('install', (event) => {
	async function addAssetsToCache() {
		const cache = await caches.open(CACHE);
		await cache.addAll(ASSETS);
	}

	event.waitUntil(addAssetsToCache());
});

self.addEventListener('activate', (event) => {
	async function deleteOldCaches() {
		for (const key of await caches.keys()) {
			if (key !== CACHE) await caches.delete(key);
		}
	}

	event.waitUntil(deleteOldCaches());
});

self.addEventListener('fetch', (event) => {
	if (event.request.method !== 'GET') return;

	async function respond() {
		const url = new URL(event.request.url);
		const cache = await caches.open(CACHE);

		if (ASSETS.includes(url.pathname)) {
			const response = await cache.match(url.pathname);
			if (response) return response;
		}

		try {
			const response = await fetch(event.request);

			// Only cache http/https requests to avoid errors with browser extensions
			if (response.status === 200 && url.protocol.startsWith('http')) {
				cache.put(event.request, response.clone());
			}

			return response;
		} catch {
			return cache.match(event.request);
		}
	}

	event.respondWith(respond() as Promise<Response>);
});

// Push notification handling
self.addEventListener('push', (event) => {
	if (!event.data) return;

	try {
		const payload = event.data.json();
		const title = payload.title || 'Grão Certo';
		const options = {
			body: payload.body,
			icon: payload.icon || '/logo.png',
			badge: '/favicon-96x96.png',
			data: payload.data
		};

		event.waitUntil(self.registration.showNotification(title, options));
	} catch (err) {
		console.error('Error handling push event:', err);
	}
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();

	const urlToOpen = event.notification.data?.url || '/';

	event.waitUntil(
		self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
			for (let i = 0; i < windowClients.length; i++) {
				const client = windowClients[i];
				if (client.url.includes(urlToOpen) && 'focus' in client) {
					return client.focus();
				}
			}
			if (self.clients.openWindow) {
				return self.clients.openWindow(urlToOpen);
			}
		})
	);
});
