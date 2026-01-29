<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Divider from './Divider.svelte';

	const { Story } = defineMeta({
		component: Divider,
		args: {
			label: 'Section',
			vertical: false,
			class: ''
		},
		argTypes: {
			label: { control: 'text' },
			vertical: { control: 'boolean' },
			class: { control: 'text' }
		}
	});
</script>

<Story name="Horizontal">
	{#snippet template(args)}
		<Divider label={args.label} vertical={false} class={args.class} />
	{/snippet}
</Story>

<Story name="Vertical">
	{#snippet template(args)}
		<div class="flex h-24 items-center">
			<Divider label={args.label} vertical={true} class={args.class} />
		</div>
	{/snippet}
</Story>
