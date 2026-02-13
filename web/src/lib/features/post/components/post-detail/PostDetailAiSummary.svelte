<script lang="ts">
	import DetailAiSummary from '$lib/ui/detail/DetailAiSummary.svelte';

	interface Props {
		summary: string;
	}

	let { summary }: Props = $props();
</script>

<DetailAiSummary {summary} />
