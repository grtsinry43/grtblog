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

<li
	class={`break-words leading-relaxed text-ink-800 dark:text-ink-200 ${className}`.trim()}
	{...attrs}
>
	{@render children?.()}
</li>
