<script lang="ts">
	import PostList from '$lib/features/post/components/PostList.svelte';
	import { postListCtx } from '$lib/features/post/context';
	import { SlideIn, FadeIn } from '$lib/ui/animation';

	let { data } = $props();

	const postStore = postListCtx.mountModelData(data);

	$effect(() => {
		postListCtx.syncModelData(postStore, data);
	});
</script>

<PostList />
