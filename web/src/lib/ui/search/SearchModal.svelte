<script lang="ts">
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import { uiState } from '$lib/shared/stores/ui.svelte';
</script>

{#if uiState.isSearchOpen}
	<QueryRoot loader={() => import('./SearchModalClient.svelte')} />
{/if}
