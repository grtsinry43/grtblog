<script lang="ts">
	const { class: className = '', ...attrs } = $props<{ class?: string }>();
</script>

<tbody class={className} {...attrs}>
	<slot />
</tbody>
