<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<hr class={`my-10 border-ink-100 dark:border-ink-800/30 ${className}`.trim()} {...attrs} />
