import type { QueryClient, QueryClientConfig } from '@tanstack/svelte-query';

let singleton: QueryClient | null = null;

const defaultOptions: QueryClientConfig = {
	defaultOptions: {
		queries: {
			staleTime: 60_000,
			refetchOnWindowFocus: false
		}
	}
};

/**
 * Get or lazily create a singleton QueryClient.
 * The import of `@tanstack/svelte-query` is deferred so it stays out of the SSR bundle.
 */
export async function getOrCreateQueryClient(
	options?: QueryClientConfig
): Promise<QueryClient> {
	if (singleton) return singleton;
	const { QueryClient } = await import('@tanstack/svelte-query');
	singleton = new QueryClient(options ?? defaultOptions);
	return singleton;
}
