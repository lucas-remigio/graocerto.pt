/* eslint-disable @typescript-eslint/no-explicit-any */
import { get, writable } from 'svelte/store';
import { browser } from '$app/environment';

// Connection config
const SOCKETS_URL = import.meta.env.VITE_SOCKETS_URL || 'ws://localhost';
const SOCKETS_PORT = import.meta.env.VITE_SOCKETS_PORT || '8090';
const isProd = import.meta.env.VITE_IS_PRODUCTION === 'true';

const protocol = isProd ? 'wss:' : 'ws:';
const WS_URL = isProd
	? `${protocol}//${location.host}/ws`
	: `${protocol}//${SOCKETS_URL}:${SOCKETS_PORT}/ws`;

console.log('WebSocket URL:', WS_URL);

// Simple socket store
export const socket = writable<WebSocket | null>(null);
export const connected = writable(false);
export const messages = writable<any[]>([]);

// Track connection status to avoid duplicate connections
let isConnecting = false;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

// Connect to websocket
export function connect() {
	if (!browser) return;

	// Prevent multiple simultaneous connection attempts
	if (isConnecting || get(socket) !== null) return;

	isConnecting = true;

	console.log('Attempting to connect to WebSocket...');
	const ws = new WebSocket(WS_URL);

	ws.onopen = () => {
		console.log('WebSocket connected');
		connected.set(true);
		socket.set(ws);
		isConnecting = false;
	};

	ws.onmessage = (event) => {
		try {
			const data = JSON.parse(event.data);
			console.log('WebSocket message received:', data);
			messages.update((msgs) => [...msgs, data]);
		} catch (error) {
			console.error('Failed to parse message', error);
		}
	};

	ws.onclose = (event) => {
		console.log('WebSocket disconnected:', event.code, event.reason);
		connected.set(false);
		socket.set(null);
		isConnecting = false;

		// Reconnect after a delay, but only if not closed cleanly
		if (!event.wasClean) {
			console.log('Scheduling reconnection...');
			// Clear any existing reconnect timer
			if (reconnectTimer) clearTimeout(reconnectTimer);
			reconnectTimer = setTimeout(connect, 3000);
		}
	};

	ws.onerror = (error) => {
		console.error('WebSocket error', error);
		connected.set(false);
		isConnecting = false;
	};
}

// Send a message
export function sendMessage(message: any) {
	const ws = get(socket);
	if (ws && ws.readyState === WebSocket.OPEN) {
		ws.send(JSON.stringify(message));
		return true;
	} else {
		console.error('WebSocket not connected');
		return false;
	}
}

// Disconnect the WebSocket
export function disconnect() {
	const ws = get(socket);
	if (ws) {
		ws.close(1000, 'User navigated away');
		socket.set(null);
		connected.set(false);
	}

	// Clear any reconnection timers
	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
		reconnectTimer = null;
	}
}

// Connect only once when this module is loaded in browser
if (browser) {
	connect();
}

// Cleanup function for page unloads
if (browser) {
	window.addEventListener('beforeunload', disconnect);
}
