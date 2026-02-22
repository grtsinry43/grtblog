<script lang="ts">
	import TagDirectory from '$lib/features/tag/components/TagDirectory.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<div class="mx-auto w-full max-w-6xl px-4 py-8 sm:px-6 sm:py-12">
	<TagDirectory tags={data.tags} />
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
