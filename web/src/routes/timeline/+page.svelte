<script lang="ts">
	import TimelineList from '$lib/features/timeline/components/TimelineList.svelte';

	let { data } = $props();
</script>

<svelte:head>
	<title>Timeline — grtsinry43</title>
	<meta name="description" content="A visual timeline of my digital footprint." />
</svelte:head>

<TimelineList items={data.items} />
