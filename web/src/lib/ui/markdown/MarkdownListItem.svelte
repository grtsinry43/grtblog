<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<li class={`break-words leading-relaxed text-ink-800 dark:text-ink-200 ${className}`.trim()} {...attrs}>
	<slot />
</li>
