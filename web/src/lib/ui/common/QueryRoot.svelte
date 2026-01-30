<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import ClientOnly from '$lib/ui/common/ClientOnly.svelte';
	import type { QueryClientConfig } from '@tanstack/svelte-query';

	import type { Component } from 'svelte';

	type LoaderProps = Record<string, unknown>;
	type LoaderComponent = Component<any>;
	let { children, options, fallback, loader, loaderProps } = $props<{
		children?: Snippet;
		options?: QueryClientConfig;
		fallback?: Snippet;
		loader?: () => Promise<{ default: LoaderComponent }>;
		loaderProps?: LoaderProps;
	}>();
	let Provider = $state<null | typeof import('@tanstack/svelte-query').QueryClientProvider>(null);
	let client = $state<null | import('@tanstack/svelte-query').QueryClient>(null);
	let Loaded = $state<null | LoaderComponent>(null);
	let ready = $state(false);

	onMount(async () => {
		// Dynamically import for minimal bundle size
		const mod = await import('@tanstack/svelte-query');
		const { QueryClient, QueryClientProvider } = mod;
		client = new QueryClient(options);
		Provider = QueryClientProvider;
		if (loader) {
			const loaded = await loader();
			Loaded = loaded.default;
		}
		ready = true;
	});
</script>

<ClientOnly fallback={fallback}>
	{#if ready && Provider && client}
		<Provider client={client}>
			{#if Loaded}
				<Loaded {...loaderProps} />
			{:else}
				{@render children?.()}
			{/if}
		</Provider>
	{/if}
</ClientOnly>
