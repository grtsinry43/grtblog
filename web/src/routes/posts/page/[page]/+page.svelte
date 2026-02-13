<script lang="ts">
	import PostList from '$lib/features/post/components/PostList.svelte';
	import { postListCtx } from '$lib/features/post/context';

	let { data } = $props();

	postListCtx.mountModelData(() => data);
</script>

<PostList />
