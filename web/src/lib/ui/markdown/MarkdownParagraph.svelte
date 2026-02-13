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

<p
	class={`mb-6 last:mb-0 wrap-break-word leading-[1.8] text-ink-800 dark:text-ink-200 ${className}`.trim()}
	{...attrs}
>
	{@render children?.()}
</p>
