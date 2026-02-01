<script lang="ts">
	import MomentDetail from '$lib/features/moment/components/MomentDetail.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<div class="w-full min-h-screen pt-8 md:pt-16 pb-12">
	<MomentDetail moment={data.moment} />
</div>
