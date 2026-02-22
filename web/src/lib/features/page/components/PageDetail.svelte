<script lang="ts">
	import type { PageDetail as PageDetailModel } from '$lib/features/page/types';
	import StickyHeader from '$lib/ui/common/StickyHeader.svelte';
	import PageDetailHeader from './page-detail/PageDetailHeader.svelte';
	import PageDetailMain from './page-detail/PageDetailMain.svelte';

	interface Props {
		page: PageDetailModel;
	}

	let { page }: Props = $props();
</script>

<StickyHeader title={page.title} />

<article class="article-enter space-y-10">
	<PageDetailHeader {page} />
	<PageDetailMain {page} />
</article>
