<script lang="ts">
	import { postDetailCtx } from '$lib/features/post/context';
	import { sameToc, sameMetrics } from './selector-equals';
	import PostDetailAiSummary from './PostDetailAiSummary.svelte';
	import PostDetailComments from './PostDetailComments.svelte';
	import PostDetailMarkdown from './PostDetailMarkdown.svelte';
	import PostDetailTocSidebar from './PostDetailTocSidebar.svelte';
	import PostDetailLeadIn from './PostDetailLeadIn.svelte';
	import DetailActionBar from '$lib/ui/detail/DetailActionBar.svelte';

	const aiSummaryStore = postDetailCtx.selectModelData((data) => data?.aiSummary ?? '');
	const contentStore = postDetailCtx.selectModelData((data) => data?.content ?? '');
	const tocStore = postDetailCtx.selectModelData((data) => data?.toc ?? [], { equals: sameToc });
	const postIdStore = postDetailCtx.selectModelData((data) => data?.id ?? 0);
	const metricsStore = postDetailCtx.selectModelData((data) => data?.metrics ?? null, {
		equals: sameMetrics
	});

	let contentRoot: HTMLElement | null = $state(null);
	let activeAnchor: string | null = $state(null);

	const handleContentRootChange = (node: HTMLElement | null) => {
		contentRoot = node;
	};

	const handleActiveAnchorChange = (anchor: string | null) => {
		activeAnchor = anchor;
	};
</script>

<div class="grid gap-10 lg:grid-cols-[1fr_220px] lg:gap-16">
	<main class="min-w-0">
		{#if $aiSummaryStore}
			<PostDetailAiSummary summary={$aiSummaryStore} />
		{/if}

		<PostDetailLeadIn />

		<PostDetailMarkdown
			content={$contentStore}
			toc={$tocStore}
			onContentRootChange={handleContentRootChange}
			onActiveAnchorChange={handleActiveAnchorChange}
		/>

		<DetailActionBar
			contentType="article"
			contentId={$postIdStore}
			likes={$metricsStore?.likes ?? 0}
			comments={$metricsStore?.comments ?? 0}
			tone="jade"
		/>
		<PostDetailComments />
	</main>

	<PostDetailTocSidebar
		toc={$tocStore}
		{contentRoot}
		{activeAnchor}
		onAnchorChange={handleActiveAnchorChange}
	/>
</div>
