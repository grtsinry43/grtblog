<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Tag from './Tag.svelte';

	const { Story } = defineMeta({
		component: Tag,
		args: {
			children: 'Tag',
			variant: 'ink',
			class: ''
		},
		argTypes: {
			children: { control: 'text' },
			variant: { control: 'select', options: ['jade', 'ink', 'amber', 'cinnabar', 'outline'] },
			class: { control: 'text' }
		}
	});
</script>

<Story name="Default">
	{#snippet template(args)}
		<Tag variant={args.variant} class={args.class}>{args.children}</Tag>
	{/snippet}
</Story>
