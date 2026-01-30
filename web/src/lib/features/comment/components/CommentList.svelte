<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { getCommentTree } from '$lib/features/comment/api';
	import CommentItem from './CommentItem.svelte';
	import { Loader2 } from 'lucide-svelte';

	let { areaId } = $props<{ areaId: number }>();

	const query = createQuery(() => ({
		queryKey: ['comments', areaId],
		queryFn: () => getCommentTree(undefined, areaId)
	}));
</script>

<div class="space-y-8 mt-12 mb-20">
	{#if query.isLoading}
		<div class="flex justify-center py-10">
			<Loader2 class="animate-spin text-ink-300" />
		</div>
	{:else if query.isError}
		<div class="text-center py-10 text-sm text-red-500">加载评论失败</div>
	{:else if query.data && query.data.length > 0}
		<div class="space-y-6">
			{#each query.data as comment (comment.id)}
				<CommentItem {comment} />
			{/each}
		</div>
	{:else}
		<div class="text-center py-16 text-ink-300 dark:text-ink-600 font-serif italic text-sm">
			暂无回响，来留下第一条评论吧...
		</div>
	{/if}
</div>
