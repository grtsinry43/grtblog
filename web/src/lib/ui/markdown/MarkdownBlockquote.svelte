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

<blockquote
	class={`my-8 border-l border-jade-500/40 py-0.5 pl-5 text-[0.95em] font-serif opacity-90 text-ink-600 dark:text-ink-400 ${className}`.trim()}
	{...attrs}
>
	{@render children?.()}
</blockquote>
