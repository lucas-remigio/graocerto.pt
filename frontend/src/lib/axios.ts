import axios from 'axios';
import { get } from 'svelte/store';
import { token } from '$lib/stores/auth';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || 'localhost';
const BACKEND_PORT = import.meta.env.VITE_BACKEND_PORT || '3001';
const isProd = import.meta.env.VITE_IS_PRODUCTION === 'true';

// Check both NODE_ENV and Vite's PROD flag
const API_URL = isProd
	? `/api/v1` // Path-based URL for production (through Nginx proxy)
	: `http://${BACKEND_URL}:${BACKEND_PORT}/api/v1`; // Full URL for development

console.log('Backend URL:', API_URL);

const api_axios = axios.create({
	baseURL: API_URL,
	withCredentials: true,
	headers: { 'Content-Type': 'application/json' }
});

// Request Interceptor: Attach token to headers
api_axios.interceptors.request.use((config) => {
	const authToken = get(token);
	if (authToken) {
		config.headers.Authorization = `Bearer ${authToken}`;
	}
	return config;
});

// Response Interceptor: Handle errors and token refresh
api_axios.interceptors.response.use(
	(response) => response, // Pass through successful responses
	async (error) => {
		if (error.response?.status === 401) {
			token.set(null);
			window.location.href = '/login';
		}
		return Promise.reject(error);
	}
);

export default api_axios;
