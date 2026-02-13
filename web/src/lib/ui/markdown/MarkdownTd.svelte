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

<td class={`px-3 py-2 text-sm ${className}`.trim()} {...attrs}>
	{@render children?.()}
</td>
