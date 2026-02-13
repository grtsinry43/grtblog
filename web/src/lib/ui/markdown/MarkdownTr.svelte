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

<tr class={`border-b border-ink-100/80 dark:border-ink-800/60 ${className}`.trim()} {...attrs}>
	{@render children?.()}
</tr>
