<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<div class="my-6 w-full overflow-x-auto">
	<table
		class={`w-full border-collapse text-sm text-ink-800 dark:text-ink-200 ${className}`.trim()}
		{...attrs}
	>
		<slot />
	</table>
</div>
