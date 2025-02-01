import axios from 'axios';
import { get } from 'svelte/store';
import { token } from '$lib/stores/auth';

const API_URL = import.meta.env.PROD
	? 'http://lucas-remigio-dev.pt:8080/api/v1'
	: 'http://localhost:8080/api/v1';

const api = axios.create({
	baseURL: API_URL,
	withCredentials: true,
	headers: { 'Content-Type': 'application/json' }
});

// Request Interceptor: Attach token to headers
api.interceptors.request.use((config) => {
	const authToken = get(token);
	if (authToken) {
		config.headers.Authorization = `Bearer ${authToken}`;
	}
	return config;
});

// Response Interceptor: Handle errors and token refresh
api.interceptors.response.use(
	(response) => response, // Pass through successful responses
	async (error) => {
		if (error.response?.status === 401) {
			token.set(null);
			window.location.href = '/login';
		}
		return Promise.reject(error);
	}
);

export default api;
