<script lang="ts">
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import type { TrackLikeContentType } from '$lib/features/analytics/types';

	interface Props {
		contentType: TrackLikeContentType;
		contentId: number;
		likes?: number;
		className?: string;
	}

	let { contentType, contentId, likes = 0, className = '' }: Props = $props();
</script>

{#snippet fallback()}
	<span class={className}>喜欢 {likes}</span>
{/snippet}

<QueryRoot
	loader={() => import('$lib/features/analytics/components/ContentLikeButtonClient.svelte')}
	loaderProps={{
		contentType,
		contentId,
		initialLikes: likes,
		className
	}}
	{fallback}
/>
