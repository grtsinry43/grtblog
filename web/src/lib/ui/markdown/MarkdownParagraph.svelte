<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<p
		class={`mb-6 last:mb-0 wrap-break-word leading-[1.8] text-ink-800 dark:text-ink-200 ${className}`.trim()}
		{...attrs}
>
	<slot />
</p>