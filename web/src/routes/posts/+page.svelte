<script lang="ts">
	import PostList from '$lib/features/post/components/PostList.svelte';
	import { postListCtx } from '$lib/features/post/context';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';

	let { data } = $props();

	postListCtx.mountModelData(() => data);
</script>

<div class="max-w-5xl mx-auto py-10">
	<PageHeader
		title="文章归档"
		tag="Archive"
		subtitle="文字是思考的锚点"
		description="按时间顺序排布的思考、笔记与技术沉淀。在这里，你可以找到所有历史文章的快照。"
	/>
	<PostList />
</div>
