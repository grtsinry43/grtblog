<script lang="ts">
	import StickyHeader from '$lib/ui/common/StickyHeader.svelte';
	import { postDetailCtx } from '$lib/features/post/context';
	import { buildImageExtInfoState, imageExtInfoCtx } from '$lib/shared/markdown/image-ext-info';
	import PostDetailHeader from './post-detail/PostDetailHeader.svelte';
	import PostDetailMain from './post-detail/PostDetailMain.svelte';

	const hasPostStore = postDetailCtx.selectModelData((data) => Boolean(data));
	const postTitleStore = postDetailCtx.selectModelData((data) => data?.title ?? '');
	const postExtInfoStore = postDetailCtx.selectModelData((data) => data?.extInfo ?? null, {
		equals: (a, b) => a === b
	});
	const imageExtInfoStore = imageExtInfoCtx.mountModelData(
		buildImageExtInfoState($postExtInfoStore)
	);

	$effect(() => {
		imageExtInfoCtx.syncModelData(imageExtInfoStore, buildImageExtInfoState($postExtInfoStore));
	});
</script>

{#if $hasPostStore}
	<StickyHeader title={$postTitleStore} />

	<article class="article-enter space-y-10">
		<PostDetailHeader />
		<PostDetailMain />
	</article>
{:else}
	<div class="py-24 text-center font-serif text-sm text-ink-400 italic">
		<p>请求的内容未能呈现。</p>
	</div>
{/if}
