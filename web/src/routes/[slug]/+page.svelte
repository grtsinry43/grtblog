<script lang="ts">
	import type { PageData } from './$types';
	import PageDetail from '$lib/features/page/components/PageDetail.svelte';
	import ContentViewTracker from '$lib/features/analytics/components/ContentViewTracker.svelte';

	let { data }: { data: PageData } = $props();
</script>

<PageDetail page={data.page} />
<ContentViewTracker contentType="page" contentId={data.page.id} />
