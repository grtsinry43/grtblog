<script lang="ts">
	import { get } from 'svelte/store';
	import PostDetail from '$lib/features/post/components/PostDetail.svelte';
	import { browser } from '$app/environment';
	import { postDetailCtx } from '$lib/features/post/context';
	import { createPostLiveUpdate } from '$lib/features/post/live-update';
	import ContentViewTracker from '$lib/features/analytics/components/ContentViewTracker.svelte';

	let { data } = $props();

	const postDetailStore = postDetailCtx.mountModelData(data.post ?? null);
	const { updateModelData } = postDetailCtx.useModelActions();
	const postIdStore = postDetailCtx.selectModelData((d) => d?.id ?? null);
	const contentHashStore = postDetailCtx.selectModelData((d) => d?.contentHash ?? null);

	$effect(() => {
		postDetailCtx.syncModelData(postDetailStore, data.post ?? null);
	});

	$effect(() => {
		if (!browser) return;
		const postId = get(postIdStore);
		if (!postId) return;

		const liveUpdate = createPostLiveUpdate({
			getId: () => get(postIdStore),
			getContentHash: () => get(contentHashStore),
			updatePost: updateModelData
		});
		liveUpdate.start(postId);
		return () => liveUpdate.destroy();
	});
</script>

<PostDetail />
<ContentViewTracker contentType="article" contentId={$postIdStore ?? 0} />
