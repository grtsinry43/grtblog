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

<th
	class={`px-3 py-2 text-left text-[11px] uppercase tracking-[0.18em] text-ink-500 dark:text-ink-400 ${className}`.trim()}
	{...attrs}
>
	{@render children?.()}
</th>
