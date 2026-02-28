<script lang="ts">
	import type { PageDetail } from '$lib/features/page/types';
	import DetailAiSummary from '$lib/ui/detail/DetailAiSummary.svelte';
	import DetailCommentSection from '$lib/ui/detail/DetailCommentSection.svelte';
	import DetailMarkdownContent from '$lib/ui/detail/DetailMarkdownContent.svelte';
	import PageDetailTocSidebar from './PageDetailTocSidebar.svelte';
	import LeadIn from '$lib/ui/detail/LeadIn.svelte';

	interface Props {
		page: PageDetail;
	}

	let { page }: Props = $props();

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
		{#if page.aiSummary}
			<DetailAiSummary summary={page.aiSummary} />
		{/if}

		<LeadIn content={page.description || ''} label="引入" />

		<DetailMarkdownContent
			content={page.content}
			toc={page.toc}
			className="markdown-body max-w-none text-[15px] leading-[1.8] font-normal text-ink-800 md:text-base dark:text-ink-200"
			onContentRootChange={handleContentRootChange}
			onActiveAnchorChange={handleActiveAnchorChange}
		/>

		<DetailCommentSection
			commentAreaId={page.commentAreaId}
			commentsCount={page.metrics?.comments ?? 0}
			fallbackText="评论区在赶来的路上..."
			fallbackSize="w-8 h-8"
		/>
	</main>

	{#if page.toc?.length}
		<PageDetailTocSidebar
			toc={page.toc}
			{contentRoot}
			{activeAnchor}
			onAnchorChange={handleActiveAnchorChange}
		/>
	{/if}
</div>
