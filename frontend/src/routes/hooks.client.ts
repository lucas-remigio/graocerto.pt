import { token } from '$lib/stores/auth';

import type { RequestEvent } from '@sveltejs/kit';

export async function handle({
	event,
	resolve
}: {
	event: RequestEvent;
	resolve: (event: RequestEvent) => Promise<Response>;
}) {
	const authToken = localStorage.getItem('token'); 

	if (authToken) {
		token.set(authToken);
	}

	return resolve(event);
}
