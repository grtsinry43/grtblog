<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<blockquote
	class={`my-8 border-l border-jade-500/40 py-0.5 pl-5 text-[0.95em] italic opacity-90 text-ink-600 dark:text-ink-400 ${className}`.trim()}
	{...attrs}
>
	<slot />
</blockquote>
