<script lang="ts">
	import type { Snippet } from 'svelte';

	const {
		children,
		class: className = '',
		...attrs
	} = $props<{
		class?: string;
		children?: Snippet;
	}>();
</script>

<tbody class={className} {...attrs}>
	{@render children?.()}
</tbody>
