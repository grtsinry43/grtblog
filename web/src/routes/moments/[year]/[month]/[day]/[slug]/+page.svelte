<script lang="ts">
	import { browser } from '$app/environment';
	import { onMount } from 'svelte';
	import ContentViewTracker from '$lib/features/analytics/components/ContentViewTracker.svelte';
	import { fetchContentMetrics } from '$lib/features/analytics/api';
	import { createMomentLiveUpdate } from '$lib/features/moment/live-update';
	import MomentDetail from '$lib/features/moment/components/MomentDetail.svelte';
	import { momentDetailCtx } from '$lib/features/moment/context';
	import { get } from 'svelte/store';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
	momentDetailCtx.mountModelData(() => data.moment ?? null);
	const { updateModelData } = momentDetailCtx.useModelActions();
	const momentStore = momentDetailCtx.selectModelData((d) => d);
	const momentIdStore = momentDetailCtx.selectModelData((d) => d?.id ?? null);
	const contentHashStore = momentDetailCtx.selectModelData((d) => d?.contentHash ?? null);

	$effect(() => {
		if (!browser) return;
		const momentId = get(momentIdStore);
		if (!momentId) return;

		const liveUpdate = createMomentLiveUpdate({
			getId: () => get(momentIdStore),
			getContentHash: () => get(contentHashStore),
			updateMoment: updateModelData
		});
		liveUpdate.start(momentId);
		return () => liveUpdate.destroy();
	});

	onMount(async () => {
		const momentId = get(momentIdStore);
		if (!momentId) return;
		const m = await fetchContentMetrics('moment', momentId);
		if (m) updateModelData((prev) => (prev ? { ...prev, metrics: m } : prev));
	});
</script>

<div class="w-full min-h-screen pt-2 md:pt-4 pb-12">
	<MomentDetail moment={$momentStore ?? data.moment} />
</div>
<ContentViewTracker contentType="moment" contentId={$momentIdStore ?? data.moment.id} />
