<script lang="ts">
	import DetailCommentSection from '$lib/ui/detail/DetailCommentSection.svelte';

	type Props = {
		areaId?: number | null;
		commentsCount?: number;
	};

	let { areaId = null, commentsCount = 0 }: Props = $props();
</script>

{#if areaId}
	<DetailCommentSection
		commentAreaId={areaId}
		commentsCount={commentsCount}
		containerClass="thinking-comments-window"
		fallbackText="评论区正在加载中..."
		fallbackSize="w-6 h-6"
		fallbackContainerClass="flex justify-center py-20"
	/>
{:else}
	<div
		class="rounded-default border border-ink-100/70 dark:border-ink-800/60 bg-ink-50/50 dark:bg-ink-900/35 px-4 py-8 text-center text-sm font-serif text-ink-500 dark:text-ink-400"
	>
		该思考暂未开放评论区
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";

	:global(.thinking-comments-window > #comments) {
		margin-top: 0;
		padding-top: 0;
		border-top: 0;
	}
</style>
