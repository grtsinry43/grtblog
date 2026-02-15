import type { RequestEvent } from '@sveltejs/kit';

export const ISR_DEPS_HEADER = 'x-grt-deps';

export const trackISRDeps = (
	event: RequestEvent,
	...deps: Array<string | null | undefined | false>
): void => {
	if (!event.locals.isrDeps) {
		event.locals.isrDeps = new Set<string>();
	}

	for (const dep of deps) {
		if (typeof dep !== 'string') continue;
		const normalized = dep.trim();
		if (!normalized) continue;
		event.locals.isrDeps.add(normalized);
	}
};
