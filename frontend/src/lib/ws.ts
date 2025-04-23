import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Connection config
const WS_URL = import.meta.env.PROD ? `wss://${window.location.host}/ws` : 'ws://localhost:8090';

// Simple socket store
export const socket = writable<WebSocket | null>(null);
export const connected = writable(false);
export const messages = writable<any[]>([]);

// Connect to websocket
export function connect() {
	if (!browser) return;

	const ws = new WebSocket(WS_URL);

	ws.onopen = () => {
		console.log('WebSocket connected');
		connected.set(true);
		socket.set(ws);
	};

	ws.onmessage = (event) => {
		try {
			const data = JSON.parse(event.data);
			messages.update((msgs) => [...msgs, data]);
		} catch (error) {
			console.error('Failed to parse message', error);
		}
	};

	ws.onclose = () => {
		console.log('WebSocket disconnected');
		connected.set(false);
		socket.set(null);
	};

	ws.onerror = (error) => {
		console.error('WebSocket error', error);
		connected.set(false);
	};
}

// Send a message
export function sendMessage(message: any) {
	socket.update((ws) => {
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify(message));
		} else {
			console.error('WebSocket not connected');
		}
		return ws;
	});
}

// Connect when browser is ready
if (browser) {
	connect();
}
