<script lang="ts">
	import { postDetailCtx } from '$lib/features/post/context';
	import { sameToc, sameMetrics } from './selector-equals';
	import PostDetailAiSummary from './PostDetailAiSummary.svelte';
	import PostDetailComments from './PostDetailComments.svelte';
	import PostDetailMarkdown from './PostDetailMarkdown.svelte';
	import PostDetailTocSidebar from './PostDetailTocSidebar.svelte';
	import PostDetailLeadIn from './PostDetailLeadIn.svelte';
	import DetailActionBar from '$lib/ui/detail/DetailActionBar.svelte';
	import SelectionToolbar from '$lib/features/reading/components/SelectionToolbar.svelte';
	import { readingPosition } from '$lib/features/reading/reading-position';
	import { toast } from 'svelte-sonner';

	const aiSummaryStore = postDetailCtx.selectModelData((data) => data?.aiSummary ?? '');
	const contentStore = postDetailCtx.selectModelData((data) => data?.content ?? '');
	const tocStore = postDetailCtx.selectModelData((data) => data?.toc ?? [], { equals: sameToc });
	const postIdStore = postDetailCtx.selectModelData((data) => data?.id ?? 0);
	const postTitleStore = postDetailCtx.selectModelData((data) => data?.title ?? '');
	const postSummaryStore = postDetailCtx.selectModelData((data) => data?.summary ?? '');
	const postCoverStore = postDetailCtx.selectModelData((data) => data?.cover ?? '');
	const metricsStore = postDetailCtx.selectModelData((data) => data?.metrics ?? null, {
		equals: sameMetrics
	});
	const commentAreaIdStore = postDetailCtx.selectModelData((data) => data?.commentAreaId ?? null);

	let contentRoot: HTMLElement | null = $state(null);
	let activeAnchor: string | null = $state(null);

	const handleContentRootChange = (node: HTMLElement | null) => {
		contentRoot = node;
	};

	const handleActiveAnchorChange = (anchor: string | null) => {
		activeAnchor = anchor;
	};

	$effect(() => {
		if (!contentRoot || !$postIdStore) return;
		const handle = readingPosition(contentRoot, {
			key: `article:${$postIdStore}`,
			onResumeAvailable: (ratio, resume) => {
				toast(`上次读到 ${Math.round(ratio * 100)}%`, {
					description: '要从上次的位置继续吗？',
					duration: 8000,
					action: { label: '继续阅读', onClick: resume }
				});
			}
		});
		return () => handle.destroy();
	});
</script>

<SelectionToolbar {contentRoot} title={$postTitleStore} allowQuote={!!$commentAreaIdStore} />

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
			shareTitle={$postTitleStore}
			shareDescription={$postSummaryStore}
			shareImageUrl={$postCoverStore}
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
