<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Input from './Input.svelte';

	const { Story } = defineMeta({
		component: Input,
		args: {
			value: '',
			placeholder: 'Search...',
			type: 'text',
			variant: 'default',
			class: ''
		},
		argTypes: {
			value: { control: 'text' },
			placeholder: { control: 'text' },
			type: { control: 'text' },
			variant: { control: 'select', options: ['default', 'underline'] },
			class: { control: 'text' }
		}
	});
</script>

<Story name="Default">
	{#snippet template(args)}
		<Input
			value={args.value}
			placeholder={args.placeholder}
			type={args.type}
			variant={args.variant}
			class={args.class}
		/>
	{/snippet}
</Story>
