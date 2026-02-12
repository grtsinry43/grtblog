<script lang="ts">
	import type { PageData } from './$types';
	import PageDetail from '$lib/features/page/components/PageDetail.svelte';

	let { data }: { data: PageData } = $props();
</script>

<PageDetail page={data.page} />
