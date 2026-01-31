<script lang="ts">
	import CommentItem from './CommentItem.svelte';
	import { Loader2 } from 'lucide-svelte';
	import { commentAreaCtx } from '$lib/features/comment/context';

	const commentsStore = commentAreaCtx.selectModelData((data) => data?.comments ?? []);
	const isLoadingStore = commentAreaCtx.selectModelData((data) => data?.isLoading ?? false);
	const isErrorStore = commentAreaCtx.selectModelData((data) => data?.isError ?? false);
</script>

<div class="space-y-8 mt-12 mb-20">
	{#if $isLoadingStore}
		<div class="flex justify-center py-10">
			<Loader2 class="animate-spin text-ink-300" />
		</div>
	{:else if $isErrorStore}
		<div class="text-center py-10 text-sm text-red-500">加载评论失败</div>
	{:else if $commentsStore && $commentsStore.length > 0}
		<div class="space-y-6">
			{#each $commentsStore as comment (comment.id)}
				<CommentItem {comment} />
			{/each}
		</div>
	{:else}
		<div class="text-center py-16 text-ink-300 dark:text-ink-600 font-serif italic text-sm">
			暂无回响，来留下第一条评论吧...
		</div>
	{/if}
</div>
