<script lang="ts">
	import FriendTimelineList from '$lib/features/friend-timeline/components/FriendTimelineList.svelte';
	import { friendTimelineListCtx } from '$lib/features/friend-timeline/context';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';

	let { data } = $props();

	friendTimelineListCtx.mountModelData(() => ({
		items: data.items,
		total: data.pagination.total,
		page: data.pagination.page,
		size: data.pagination.size
	}));
</script>

<svelte:head>
	<title>朋友圈 | GrtBlog</title>
</svelte:head>

<div class="max-w-5xl mx-auto py-10">
	<PageHeader
		title="朋友圈"
		tag="Friend Timeline"
		subtitle="见字如面，相逢在字里行间"
		description="聚合了友情链接中朋友们的最新文章与动态，感受网络邻居们的思考与生活。"
	/>
	<FriendTimelineList />
</div>
