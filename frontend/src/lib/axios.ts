import axios from 'axios';
import { get } from 'svelte/store';
import { token } from '$lib/stores/auth';

const BACKEND_URL = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8080';

const API_URL = `${BACKEND_URL}/api/v1`;
console.log('API URL:', API_URL);
console.log('Backend URL:', import.meta.env.VITE_BACKEND_URL);

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
