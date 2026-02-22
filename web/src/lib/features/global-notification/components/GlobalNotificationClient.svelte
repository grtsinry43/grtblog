<script lang="ts">
	import { onMount } from 'svelte';
	import { globalNotificationStore } from '$lib/features/global-notification/store.svelte';
	import { realtimeActivityStore } from '$lib/features/realtime-activity/store.svelte';

	onMount(() => {
		globalNotificationStore.start();
		realtimeActivityStore.start();
		return () => {
			globalNotificationStore.stop();
			realtimeActivityStore.stop();
		};
	});
</script>

<div aria-hidden="true"></div>
