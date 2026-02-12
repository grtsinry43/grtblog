<script lang="ts">
	import Input from '$lib/ui/primitives/input/Input.svelte';

	let {
		label,
		name,
		type = 'text',
		autocomplete,
		required = false
	} = $props<{
		label: string;
		name: string;
		type?: string;
		autocomplete?: string;
		required?: boolean;
	}>();
</script>

<div class="space-y-2">
	<label class="text-xs font-mono text-ink-500">{label}</label>
	<Input
		{name}
		{type}
		{autocomplete}
		{required}
		variant="underline"
		inputClass="rounded-default border-ink-200 bg-white/70 text-ink-900 placeholder:text-ink-400 focus:border-jade-500 focus:ring-2 dark:border-ink-700 dark:bg-ink-900/40 dark:text-ink-100 dark:placeholder:text-ink-500"
	/>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
