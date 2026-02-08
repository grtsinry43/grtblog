<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<th
	class={`px-3 py-2 text-left text-[11px] uppercase tracking-[0.18em] text-ink-500 dark:text-ink-400 ${className}`.trim()}
	{...attrs}
>
	<slot />
</th>
