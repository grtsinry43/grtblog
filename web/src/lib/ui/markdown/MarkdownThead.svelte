<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<thead class={`bg-ink-50 dark:bg-ink-900/60 ${className}`.trim()} {...attrs}>
	<slot />
</thead>
