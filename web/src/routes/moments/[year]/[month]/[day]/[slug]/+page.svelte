<script lang="ts">
	import ContentViewTracker from '$lib/features/analytics/components/ContentViewTracker.svelte';
	import MomentDetail from '$lib/features/moment/components/MomentDetail.svelte';
	import { momentDetailCtx } from '$lib/features/moment/context';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
	const momentDetailStore = momentDetailCtx.mountModelData(data.moment ?? null);

	$effect(() => {
		momentDetailCtx.syncModelData(momentDetailStore, data.moment ?? null);
	});
</script>

<div class="w-full min-h-screen pt-2 md:pt-4 pb-12">
	<MomentDetail moment={data.moment} />
</div>
<ContentViewTracker contentType="moment" contentId={data.moment.id} />
