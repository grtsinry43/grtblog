<script lang="ts">
	import { browser } from '$app/environment';
	import { createPageLiveUpdate } from '$lib/features/page/live-update';
	import type { PageDetail as PageDetailModel } from '$lib/features/page/types';
	import type { PageData } from './$types';
	import PageDetail from '$lib/features/page/components/PageDetail.svelte';
	import ContentViewTracker from '$lib/features/analytics/components/ContentViewTracker.svelte';

	let { data }: { data: PageData } = $props();
	let pageModel = $state<PageDetailModel | null>(null);

	$effect(() => {
		pageModel = data.page ?? null;
	});

	$effect(() => {
		if (!browser || !pageModel?.id) return;

		const liveUpdate = createPageLiveUpdate({
			getId: () => pageModel?.id ?? null,
			getContentHash: () => pageModel?.contentHash ?? null,
			updatePage: (updater) => {
				pageModel = updater(pageModel);
			}
		});
		liveUpdate.start(pageModel.id);
		return () => liveUpdate.destroy();
	});
</script>

<PageDetail page={pageModel} />
<ContentViewTracker contentType="page" contentId={pageModel?.id ?? 0} />
