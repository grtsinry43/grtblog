<script lang="ts">
	import { postDetailCtx } from '$lib/features/post/context';
	import { sameToc } from './selector-equals';
	import PostDetailAiSummary from './PostDetailAiSummary.svelte';
	import PostDetailComments from './PostDetailComments.svelte';
	import PostDetailFooter from './PostDetailFooter.svelte';
	import PostDetailMarkdown from './PostDetailMarkdown.svelte';
	import PostDetailTocSidebar from './PostDetailTocSidebar.svelte';
	import PostDetailLeadIn from './PostDetailLeadIn.svelte';

	const aiSummaryStore = postDetailCtx.selectModelData((data) => data?.aiSummary ?? '');
	const contentStore = postDetailCtx.selectModelData((data) => data?.content ?? '');
	const tocStore = postDetailCtx.selectModelData((data) => data?.toc ?? [], { equals: sameToc });

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

		<PostDetailFooter />
		<PostDetailComments />
	</main>

	<PostDetailTocSidebar
		toc={$tocStore}
		{contentRoot}
		{activeAnchor}
		onAnchorChange={handleActiveAnchorChange}
	/>
</div>
