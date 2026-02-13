<script lang="ts">
	import { postDetailCtx } from '$lib/features/post/context';
	import DetailCommentSection from '$lib/ui/detail/DetailCommentSection.svelte';

	const commentAreaIdStore = postDetailCtx.selectModelData((data) => data?.commentAreaId ?? null);
	const commentsCountStore = postDetailCtx.selectModelData((data) => data?.metrics?.comments ?? 0);
</script>

<DetailCommentSection
	commentAreaId={$commentAreaIdStore}
	commentsCount={$commentsCountStore}
	fallbackText="评论区在赶来的路上..."
	fallbackSize="w-8 h-8"
/>
