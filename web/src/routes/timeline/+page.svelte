<script lang="ts">
	import TimelineList from '$lib/features/timeline/components/TimelineList.svelte';
	import UpgradeNotification from '$lib/features/timeline/components/UpgradeNotification.svelte';

	let { data } = $props();
</script>

<TimelineList {data} />
<UpgradeNotification />
