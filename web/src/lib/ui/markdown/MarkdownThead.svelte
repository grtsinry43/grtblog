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

<thead class={`bg-ink-50 dark:bg-ink-900/60 ${className}`.trim()} {...attrs}>
	{@render children?.()}
</thead>
