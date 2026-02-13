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

<div class="my-6 w-full overflow-x-auto">
	<table
		class={`w-full border-collapse text-sm text-ink-800 dark:text-ink-200 ${className}`.trim()}
		{...attrs}
	>
		{@render children?.()}
	</table>
</div>
