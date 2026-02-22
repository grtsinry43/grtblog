<script lang="ts">
	import { browser } from '$app/environment';
	import { createPageLiveUpdate } from '$lib/features/page/live-update';
	import type { PageDetail as PageDetailModel } from '$lib/features/page/types';
	import type { PageData } from './$types';
	import PageDetail from '$lib/features/page/components/PageDetail.svelte';
	import ContentViewTracker from '$lib/features/analytics/components/ContentViewTracker.svelte';

	let { data }: { data: PageData } = $props();
	let pageModel = $state<PageDetailModel>(data.page);

	$effect(() => {
		if (!browser) return;

		const liveUpdate = createPageLiveUpdate({
			getId: () => pageModel.id,
			getContentHash: () => pageModel.contentHash,
			updatePage: (updater) => {
				const next = updater(pageModel);
				if (next) {
					pageModel = next;
				}
			}
		});
		liveUpdate.start(pageModel.id);
		return () => liveUpdate.destroy();
	});
</script>

<PageDetail page={pageModel} />
<ContentViewTracker contentType="page" contentId={pageModel.id} />
