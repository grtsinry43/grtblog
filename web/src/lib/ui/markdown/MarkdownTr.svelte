<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<tr
	class={`border-b border-ink-100/80 dark:border-ink-800/60 ${className}`.trim()}
	{...attrs}
>
	<slot />
</tr>
