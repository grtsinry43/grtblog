<script lang="ts">
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import Loading from '$lib/ui/common/Loading.svelte';

	interface Props {
		commentAreaId?: number | null;
		commentsCount?: number;
		containerClass?: string;
		fallbackText?: string;
		fallbackSize?: string;
		fallbackContainerClass?: string;
	}

	let {
		commentAreaId = null,
		commentsCount = 0,
		containerClass = '',
		fallbackText = '评论区在赶来的路上...',
		fallbackSize = 'w-8 h-8',
		fallbackContainerClass = 'flex justify-center py-40'
	}: Props = $props();
</script>

{#if commentAreaId}
	<div class={containerClass}>
		{#snippet commentFallback()}
			<div class={fallbackContainerClass}>
				<Loading size={fallbackSize} duration={1000} text={fallbackText} />
			</div>
		{/snippet}
		<QueryRoot
			loader={() => import('$lib/features/comment/components/CommentAreaClient.svelte')}
			loaderProps={{
				areaId: commentAreaId,
				commentsCount
			}}
			fallback={commentFallback}
		/>
	</div>
{/if}
