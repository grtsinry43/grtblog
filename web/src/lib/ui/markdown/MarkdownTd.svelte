<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<td class={`px-3 py-2 text-sm ${className}`.trim()} {...attrs}>
	<slot />
</td>
