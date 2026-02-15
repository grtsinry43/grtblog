import type { Handle } from '@sveltejs/kit';
import { ISR_DEPS_HEADER } from '$lib/server/isr-deps';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.isrDeps = new Set<string>();

	const response = await resolve(event);
	if (event.locals.isrDeps.size === 0) {
		return response;
	}

	const headers = new Headers(response.headers);
	headers.set(ISR_DEPS_HEADER, JSON.stringify(Array.from(event.locals.isrDeps)));
	return new Response(response.body, {
		status: response.status,
		statusText: response.statusText,
		headers
	});
};
