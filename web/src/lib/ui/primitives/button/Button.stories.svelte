<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import Button from './Button.svelte';

	const { Story } = defineMeta({
		component: Button,
		args: {
			children: 'Button',
			variant: 'primary',
			size: 'md',
			fullWidth: false,
			loading: false,
			disabled: false
		},
		argTypes: {
			children: { control: 'text' },
			variant: { control: 'select', options: ['primary', 'secondary', 'ghost', 'icon'] },
			size: { control: 'select', options: ['sm', 'md', 'lg'] },
			fullWidth: { control: 'boolean' },
			loading: { control: 'boolean' },
			disabled: { control: 'boolean' }
		}
	});
</script>

<Story name="Primary">
	{#snippet template(args)}
		<Button {...args}>{args.children}</Button>
	{/snippet}
</Story>
