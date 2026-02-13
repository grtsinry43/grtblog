<script lang="ts">
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import type { TrackViewContentType } from '$lib/features/analytics/types';

	interface Props {
		contentType: TrackViewContentType;
		contentId: number;
	}

	let { contentType, contentId }: Props = $props();
</script>

{#snippet fallback()}
	<span class="hidden" aria-hidden="true"></span>
{/snippet}

<QueryRoot
	loader={() => import('$lib/features/analytics/components/ContentViewTrackerClient.svelte')}
	loaderProps={{ contentType, contentId }}
	{fallback}
/>
