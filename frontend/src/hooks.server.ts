import type { Handle } from '@sveltejs/kit';

const publicRoutes = ['/login', '/register'];

export const handle: Handle = async ({ event, resolve }) => {
	const authToken = event.cookies.get('authToken'); // Retrieve the auth token from cookies
	console.log('hello!!' + authToken);

	// Check if the route is a public route
	if (publicRoutes.includes(event.url.pathname)) {
		return resolve(event); // Allow public routes without authentication
	}

	// If no token is found, redirect to the login page
	if (!authToken) {
		return new Response(null, {
			status: 302,
			headers: { Location: '/login' }
		});
	}

	// Allow access to private routes
	return resolve(event);
};
