import type { Handle } from '@sveltejs/kit';

const publicRoutes = ['/login', '/register'];

export const handle: Handle = async ({ event, resolve }) => {
	// Skip route guard during the build
	if (!event.route.id) {
		return resolve(event);
	}

	const authToken = event.cookies.get('authToken'); // Retrieve the auth token from cookies

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
