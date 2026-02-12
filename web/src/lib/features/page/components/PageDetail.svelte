<script lang="ts">
	import type { PageDetail as PageDetailModel } from '$lib/features/page/types';
	import StickyHeader from '$lib/ui/common/StickyHeader.svelte';
	import PageDetailHeader from './page-detail/PageDetailHeader.svelte';
	import PageDetailMain from './page-detail/PageDetailMain.svelte';

	interface Props {
		page: PageDetailModel | null;
	}

	let { page }: Props = $props();
</script>

{#if page}
	<StickyHeader title={page.title} />

	<article class="article-enter space-y-10">
		<PageDetailHeader {page} />
		<PageDetailMain {page} />
	</article>
{:else}
	<div class="py-24 text-center font-serif text-sm text-ink-400 italic">
		<p>请求的页面未能呈现。</p>
	</div>
{/if}
