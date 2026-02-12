<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Textarea from './Textarea.svelte';

	const { Story } = defineMeta({
		component: Textarea,
		args: {
			value: '',
			placeholder: 'Write something...',
			rows: 4,
			variant: 'default',
			resize: 'vertical',
			class: ''
		},
		argTypes: {
			value: { control: 'text' },
			placeholder: { control: 'text' },
			rows: { control: 'number' },
			variant: { control: 'select', options: ['default', 'underline'] },
			resize: { control: 'select', options: ['none', 'vertical', 'horizontal', 'both'] },
			class: { control: 'text' }
		}
	});
</script>

<Story name="Default">
	{#snippet template(args)}
		<Textarea
			value={args.value}
			placeholder={args.placeholder}
			rows={args.rows}
			variant={args.variant}
			resize={args.resize}
			class={args.class}
		/>
	{/snippet}
</Story>
