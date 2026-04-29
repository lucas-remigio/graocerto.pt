/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

// Configuration
const CACHE_NAME = `cache-${version}`;
const STATIC_ASSETS = new Set([...build, ...files]);
const DEFAULT_ICON = '/logo.png';
const DEFAULT_BADGE = '/favicon-96x96.png';

const sw = globalThis as unknown as ServiceWorkerGlobalScope;

/**
 * Lifecycle: Install
 * Caches all static assets provided by SvelteKit.
 */
sw.addEventListener('install', (event) => {
	event.waitUntil(
		caches.open(CACHE_NAME).then((cache) => cache.addAll(Array.from(STATIC_ASSETS)))
	);
	sw.skipWaiting();
});

/**
 * Lifecycle: Activate
 * Removes outdated caches from previous versions.
 */
sw.addEventListener('activate', (event) => {
	event.waitUntil(
		caches.keys().then(async (keys) => {
			for (const key of keys) {
				if (key !== CACHE_NAME) {
					await caches.delete(key);
				}
			}
			await sw.clients.claim();
		})
	);
});

/**
 * Fetch Event
 * Implements different caching strategies based on the request type.
 */
sw.addEventListener('fetch', (event) => {
	if (event.request.method !== 'GET') return;

	const url = new URL(event.request.url);
	const isStaticAsset = STATIC_ASSETS.has(url.pathname);

	event.respondWith(
		isStaticAsset ? cacheFirst(event.request) : networkFirst(event.request)
	);
});

async function cacheFirst(request: Request): Promise<Response> {
	const cache = await caches.open(CACHE_NAME);
	const cachedResponse = await cache.match(request);
	return cachedResponse || fetch(request);
}

async function networkFirst(request: Request): Promise<Response> {
	const cache = await caches.open(CACHE_NAME);
	try {
		const response = await fetch(request);
		const url = new URL(request.url);

		// Cache successful responses from http/https (ignoring chrome-extension, etc.)
		if (response.status === 200 && url.protocol.startsWith('http')) {
			cache.put(request, response.clone());
		}
		return response;
	} catch (error) {
		const cachedResponse = await cache.match(request);
		if (cachedResponse) return cachedResponse;
		throw error;
	}
}

/**
 * Push Notifications
 * Displays incoming push notifications using the provided payload.
 */
sw.addEventListener('push', (event) => {
	if (!event.data) return;

	try {
		const payload = event.data.json();
		const title = payload.title || 'Grão Certo';
		const options: NotificationOptions = {
			body: payload.body,
			icon: payload.icon || DEFAULT_ICON,
			badge: DEFAULT_BADGE,
			data: payload.data
		};

		event.waitUntil(sw.registration.showNotification(title, options));
	} catch (err) {
		console.error('[Service Worker] Push event error:', err);
	}
});

/**
 * Notification Click
 * Handles navigation and window focus when a user clicks a notification.
 */
sw.addEventListener('notificationclick', (event) => {
	event.notification.close();

	const urlToOpen = event.notification.data?.url || '/';
	event.waitUntil(handleNotificationClick(urlToOpen));
});

async function handleNotificationClick(url: string) {
	const windowClients = await sw.clients.matchAll({
		type: 'window',
		includeUncontrolled: true
	});

	// Try to focus an existing window with the same URL
	for (const client of windowClients) {
		if (client.url.includes(url) && 'focus' in client) {
			return client.focus();
		}
	}

	// Otherwise, open a new window
	if (sw.clients.openWindow) {
		return sw.clients.openWindow(url);
	}
}
