<script lang="ts">
	import { imageLazy } from '$lib/shared/actions/image-lazy';

	const { src = '', alt = '', title = '' } = $props<{
		src?: string;
		alt?: string;
		title?: string;
	}>();
</script>

<figure class="md-figure">
	<img class="md-img" {src} {alt} loading="lazy" decoding="async" use:imageLazy={{ src }} />
	{#if title}
		<figcaption class="md-caption">{title}</figcaption>
	{/if}
</figure>
